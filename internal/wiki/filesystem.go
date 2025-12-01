// Package wiki generates HTML from markdown for a given wiki.
package wiki

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/stalexan/gomarkwiki/internal/util"
)

// warnIfSymlinkToDir checks if path is a symlink to a directory and prints a warning if so.
// Returns true if it is a symlink to a directory, false otherwise.
//
// Security note: Skipping symlinks to directories is important for two reasons:
// 1. Prevents symlink loops that could cause infinite recursion
// 2. Prevents bypassing the MaxRecursionDepth limit via symlinks to deep directory trees
// filepath.Walk uses os.Lstat internally, so it never follows directory symlinks automatically.
func warnIfSymlinkToDir(info fs.FileInfo, path string) bool {
	// Check if this is a symlink
	// Note: filepath.Walk uses os.Lstat internally, so symlinks will have os.ModeSymlink set.
	if info.Mode()&os.ModeSymlink == 0 {
		return false
	}

	// Check if the symlink target is a directory
	targetInfo, err := os.Stat(path)
	if err != nil {
		// Can't determine target, not a symlink to a directory
		return false
	}

	if targetInfo.IsDir() {
		util.PrintWarning("Skipping symlink to directory '%s' (symlinks to directories are not followed)", path)
		return true
	}

	return false
}

// isReadableFile checks to see whether path is a regular file and readable.
func isReadableFile(info fs.FileInfo, path string) bool {
	// Is this a dir?
	if info.IsDir() {
		return false
	}

	// Is the file regular or a symlink?
	// Note: filepath.Walk uses os.Lstat internally, so symlinks will have os.ModeSymlink set.
	mode := info.Mode()
	isSymlink := mode&os.ModeSymlink != 0
	if !mode.IsRegular() && !isSymlink {
		util.PrintWarning("Skipping not regular file '%s'", path)
		return false
	}

	// If it's a symlink, verify the target is a regular file.
	// os.Stat follows symlinks, so this checks the target's type.
	if isSymlink {
		targetInfo, err := os.Stat(path)
		if err != nil {
			util.PrintWarning("Skipping symlink with unresolvable target '%s': %v", path, err)
			return false
		}
		if !targetInfo.Mode().IsRegular() {
			util.PrintWarning("Skipping symlink to non-regular file '%s'", path)
			return false
		}
	}

	// Check readability by attempting to open the file.
	file, err := os.Open(path)
	if err != nil {
		util.PrintWarning("Skipping not readable file '%s': %v", path, err)
		return false
	}
	file.Close()

	// This is readable file.
	return true
}

// sourceIsOlder returns true if source is older than dest (i.e., dest is newer).
func sourceIsOlder(sourceInfo fs.FileInfo, destPath string) bool {
	destInfo, err := os.Stat(destPath)
	if err == nil && sourceInfo.ModTime().Before(destInfo.ModTime()) {
		return true
	}
	return false
}

// copyToFile copies source to the file at destPath, overwriting destPath if it exists.
// For large files, the copy operation respects context cancellation.
// This function uses atomic write semantics: it writes to a temporary file first,
// then renames it to the destination. This ensures that on cancellation or error,
// the destination file is never left in a partially written state.
func copyToFile(ctx context.Context, destPath string, source io.Reader) (err error) {
	// Create temp file in the same directory as the destination.
	// This ensures the rename will be atomic (same filesystem).
	destDir := filepath.Dir(destPath)
	tempFile, err := os.CreateTemp(destDir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file in '%s': %v", destDir, err)
	}
	tempPath := tempFile.Name()

	// Clean up temp file on error or cancellation.
	defer func() {
		if err != nil {
			os.Remove(tempPath)
		}
	}()

	defer func() {
		if closeErr := tempFile.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close temp file '%s': %v", tempPath, closeErr)
		}
	}()

	// Copy source to temp file with context cancellation support.
	// We check context periodically during the copy to allow cancellation
	// of large file operations.
	const bufSize = 32 * 1024 // 32KB chunks
	buf := make([]byte, bufSize)

	for {
		// Check for cancellation before each chunk
		select {
		case <-ctx.Done():
			return fmt.Errorf("copy cancelled: %w", ctx.Err())
		default:
		}

		// Read chunk
		nr, readErr := source.Read(buf)
		if nr > 0 {
			// Write chunk
			nw, writeErr := tempFile.Write(buf[:nr])
			if writeErr != nil {
				return fmt.Errorf("failed to write to temp file '%s': %v", tempPath, writeErr)
			}
			if nw != nr {
				return fmt.Errorf("failed to write to temp file '%s': short write", tempPath)
			}
		}

		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return fmt.Errorf("failed to read source: %v", readErr)
		}
	}

	// Atomically rename temp file to destination.
	// This is atomic on POSIX systems, ensuring no partial writes are visible.
	if err := os.Rename(tempPath, destPath); err != nil {
		return fmt.Errorf("failed to rename temp file '%s' to '%s': %v", tempPath, destPath, err)
	}

	return nil
}

// copyFileToDest copies a file from the source dir to the dest dir.
func (wiki Wiki) copyFileToDest(ctx context.Context, sourceInfo fs.FileInfo, sourcePath, sourceRelPath string, regen bool) error {
	// Check for cancellation before starting
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Re-stat the file to get current info and prevent TOCTOU issues.
	// This must happen BEFORE the skip decision to avoid a race condition where the file
	// changes between the Walk and the copy decision.
	currentInfo, err := os.Stat(sourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			util.PrintVerbose("'%s' was not copied to dest because it no longer exists", sourcePath)
			return nil
		} else {
			return fmt.Errorf("failed to stat source file '%s': %v", sourcePath, err)
		}
	}

	// Skip copying if source is older than dest.
	// Use currentInfo (not sourceInfo) to ensure we have the latest modification time.
	destPath := filepath.Join(wiki.DestDir, sourceRelPath)
	if !regen && sourceIsOlder(currentInfo, destPath) {
		return nil
	}

	// Create dest dir if it doesn't exist.
	// os.MkdirAll is idempotent, so no need to check existence first (avoids TOCTOU).
	destDirPath := filepath.Dir(destPath)
	if err := os.MkdirAll(destDirPath, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create dest dir '%s': %v", destDirPath, err)
	}

	// Copy file.
	util.PrintVerbose("Copying '%s'", sourceRelPath)
	source, err := os.Open(sourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			util.PrintVerbose("'%s' was not copied to dest because it no longer exists", sourcePath)
			return nil
		} else {
			util.PrintError(err, "could not open '%s' for copy to dest", sourcePath)
			return nil
		}
	}
	defer source.Close()
	if err := copyToFile(ctx, destPath, source); err != nil {
		return err
	}

	return nil
}

// copyCssFile copies the embedded css `file` to dest dir.
func (wiki *Wiki) copyCssFile(ctx context.Context, file string) error {
	// Read file
	var css []byte
	var err error
	sourcePath := fmt.Sprintf("static/%s", file)
	if css, err = embeddedFileSystem.ReadFile(sourcePath); err != nil {
		return fmt.Errorf("failed to read embedded file '%s': %v", sourcePath, err)
	}

	// Copy file
	destPath := fmt.Sprintf("%s/%s", wiki.DestDir, file)
	util.PrintVerbose("Copying '%s' to '%s'", sourcePath, destPath)
	if err := copyToFile(ctx, destPath, bytes.NewReader(css)); err != nil {
		return err
	}

	return nil
}

// copyCssFiles copies CSS files to dest dir.
func (wiki *Wiki) copyCssFiles(ctx context.Context, relDestPaths map[string]bool) error {
	// Don't delete css files even though they don't have a corresponding
	// file in the source dir.
	cssFiles := []string{"style.css", "github-style.css"}
	if relDestPaths != nil {
		for _, cssFile := range cssFiles {
			relDestPaths[cssFile] = true
		}
	}

	// Always copy CSS files to ensure users get the latest styles after upgrades.
	for _, cssFile := range cssFiles {
		// Check for cancellation before each file
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err := wiki.copyCssFile(ctx, cssFile); err != nil {
			return err
		}
	}

	return nil
}

// listDirectoryContents lists the contents of a directory.
func listDirectoryContents(path string) ([]os.FileInfo, error) {
	dir, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer dir.Close()

	entries, err := dir.Readdir(0)
	if err != nil {
		return nil, err
	}

	return entries, nil
}

// isDirectoryEmpty checks whether a directory is empty.
func isDirectoryEmpty(path string) (bool, error) {
	entries, err := listDirectoryContents(path)
	if err != nil {
		return false, err
	}
	return len(entries) == 0, nil
}

// deleteEmptyDirectories deletes any empty directories within path, including
// directories that have just empty directories.
func deleteEmptyDirectories(path string) error {
	return deleteEmptyDirectoriesWithDepth(path, 0)
}

// deleteEmptyDirectoriesWithDepth is the internal recursive implementation that tracks depth.
func deleteEmptyDirectoriesWithDepth(path string, depth int) error {
	// Check recursion depth limit
	if depth > MaxRecursionDepth {
		return fmt.Errorf("directory recursion depth exceeded at '%s' (depth %d, max %d)", path, depth, MaxRecursionDepth)
	}

	entries, err := listDirectoryContents(path)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		entryPath := filepath.Join(path, entry.Name())

		if entry.IsDir() {
			// Recursively delete empty directories in subdirectories.
			err := deleteEmptyDirectoriesWithDepth(entryPath, depth+1)
			if err != nil {
				return err
			}

			// Check wehther the directory is empty.
			isEmpty, err := isDirectoryEmpty(entryPath)
			if err != nil {
				return err
			}

			if isEmpty {
				// Delete the empty directory.
				// Ignore any error - if removal fails (e.g., due to TOCTOU race), it's harmless.
				util.PrintVerbose("Deleting empty directory '%s'", entryPath)
				os.Remove(entryPath)
			}
		}
	}

	return nil
}

// cleanDestDir cleans the dest dir by any deleting files that don't have
// a corresponding source file, and by deleting any empty directories.
func (wiki Wiki) cleanDestDir(ctx context.Context, relDestPaths map[string]bool) error {
	// Delete dest files that don't have a corresponding source file.
	baseDepth := strings.Count(wiki.DestDir, string(filepath.Separator))
	err := filepath.Walk(wiki.DestDir, func(destPath string, info fs.FileInfo, err error) error {
		// Check for cancellation periodically during walk
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Check recursion depth
		currentDepth := strings.Count(destPath, string(filepath.Separator)) - baseDepth
		if currentDepth > MaxRecursionDepth {
			return fmt.Errorf("directory recursion depth exceeded at '%s' (depth %d, max %d)", destPath, currentDepth, MaxRecursionDepth)
		}

		// Was there an error looking up this file?
		if err != nil {
			return err
		}

		// Is this file regular and readable?
		if !isReadableFile(info, destPath) {
			// Warn if this is a symlink to a directory
			warnIfSymlinkToDir(info, destPath)
			return nil
		}

		// What's the relative path to this file with respect to the dest dir?
		var relDestPath string
		relDestPath, err = filepath.Rel(wiki.DestDir, destPath)
		if err != nil {
			return fmt.Errorf("failed to find relative path of '%s' given '%s': %v", destPath, wiki.DestDir, err)
		}

		// Defensive check: ensure the relative path doesn't escape the dest dir.
		// This should never happen with filepath.Walk, but adds a safety layer.
		if strings.Contains(relDestPath, "..") {
			util.PrintWarning("Skipping suspicious path with '..': '%s' (from '%s')", relDestPath, destPath)
			return nil
		}

		// Clean the path to normalize it (removes redundant separators, etc.)
		relDestPath = filepath.Clean(relDestPath)

		// Delete this file if it doesn't have a corresponding file in the source dir.
		if !relDestPaths[relDestPath] {
			util.PrintVerbose("Deleting '%s'", destPath)
			if err = os.Remove(destPath); err != nil {
				util.PrintWarning("Failed to delete '%s': %v", destPath, err)
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("cleaning destination failed: %v", err)
	}

	// Check for cancellation before deleting empty directories
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Delete empty directories.
	if err := deleteEmptyDirectories(wiki.DestDir); err != nil {
		return fmt.Errorf("failed to delete empty directories in '%s': %v", wiki.DestDir, err)
	}

	return nil
}
