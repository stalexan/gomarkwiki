// Package wiki generates HTML from markdown for a given wiki.
package wiki

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/stalexan/gomarkwiki/internal/util"
)

// isReadableFile checks to see whether path is a regular file and readable.
func isReadableFile(info fs.FileInfo, path string) bool {
	// Is this a dir?
	if info.IsDir() {
		return false
	}

	// Is the file regular and readable?
	mode := info.Mode()
	if mode.IsRegular() || (mode&os.ModeSymlink != 0) {
		if mode.Perm()&(1<<2) == 0 {
			util.PrintWarning("Skipping not readable file '%s'", path)
			return false
		}
	} else {
		util.PrintWarning("Skipping not regular file '%s'", path)
		return false
	}

	// This is readable file.
	return true
}

// destIsOlder returns true if dest is older than source.
func destIsOlder(sourceInfo fs.FileInfo, destPath string) bool {
	destInfo, err := os.Stat(destPath)
	if err == nil && sourceInfo.ModTime().Before(destInfo.ModTime()) {
		return true
	}
	return false
}

// copyToFile copies source to the file at destPath, overwriting destPath if it exists.
func copyToFile(destPath string, source io.Reader) error {
	// Create and open dest file. Truncate it if it exists.
	var destFile *os.File
	var err error
	if destFile, err = os.Create(destPath); err != nil {
		return fmt.Errorf("failed to open file '%s': %v", destPath, err)
	}
	defer destFile.Close()

	// Copy source to destFile.
	if _, err = io.Copy(destFile, source); err != nil {
		return fmt.Errorf("failed to write to '%s': %v", destPath, err)
	}

	return nil
}

// copyFileToDest copies a file from the source dir to the dest dir.
func (wiki Wiki) copyFileToDest(sourceInfo fs.FileInfo, sourcePath, sourceRelPath string, regen bool) error {
	// Skip copying if source is older than dest.
	destPath := filepath.Join(wiki.DestDir, sourceRelPath)
	if !regen && destIsOlder(sourceInfo, destPath) {
		return nil
	}

	// Create dest dir if it doesn't exist.
	destDirPath := filepath.Dir(destPath)
	var err error
	if _, err = os.Stat(destDirPath); os.IsNotExist(err) {
		if err = os.MkdirAll(destDirPath, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create dest dir '%s': %v", destDirPath, err)
		}
	}

	// Copy file.
	util.PrintVerbose("Copying '%s'", sourceRelPath)
	var source *os.File
	if source, err = os.Open(sourcePath); err != nil {
		if os.IsNotExist(err) {
			util.PrintVerbose("'%s' was not copied to dest because it no longer exists", sourcePath)
		} else {
			util.PrintError(err, "could not open '%s' for copy to dest", sourcePath)
		}
	}
	defer source.Close()
	if err := copyToFile(destPath, source); err != nil {
		return err
	}

	return nil
}

// copyCssFile copies the embedded css `file` to dest dir.
func (wiki *Wiki) copyCssFile(file string) error {
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
	if err := copyToFile(destPath, bytes.NewReader(css)); err != nil {
		return err
	}

	return nil
}

// copyCssFiles copies CSS files to dest dir.
func (wiki *Wiki) copyCssFiles(relDestPaths map[string]bool) error {
	// Don't delete css files even though they don't have a corresponding
	// file in the source dir.
	cssFiles := []string{"style.css", "github-style.css"}
	for _, cssFile := range cssFiles {
		relDestPaths[cssFile] = true
	}

	// Is copy neeeded?
	if !wiki.styleCssCopyNeeded {
		return nil
	}

	// Copy CSS files.
	for _, cssFile := range cssFiles {
		if err := wiki.copyCssFile(cssFile); err != nil {
			return err
		}
	}

	// CSS files only need to be copied once per run.
	wiki.styleCssCopyNeeded = false

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
	entries, err := listDirectoryContents(path)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		entryPath := filepath.Join(path, entry.Name())

		if entry.IsDir() {
			// Recursively delete empty directories in subdirectories.
			err := deleteEmptyDirectories(entryPath)
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
				util.PrintVerbose("Deleting empty directory '%s'", entryPath)
				err := os.Remove(entryPath)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// cleanDestDir cleans the dest dir by any deleting files that don't have
// a corresponding source file, and by deleting any empty directories.
func (wiki Wiki) cleanDestDir(relDestPaths map[string]bool) error {
	// Delete dest files that don't have a corresponding source file.
	err := filepath.Walk(wiki.DestDir, func(destPath string, info fs.FileInfo, err error) error {
		// Was there an error looking up this file?
		if err != nil {
			return err
		}

		// Is this file regular and readable?
		if !isReadableFile(info, destPath) {
			return nil
		}

		// What's the relative path to this file with respect to the dest dir?
		var relDestPath string
		relDestPath, err = filepath.Rel(wiki.DestDir, destPath)
		if err != nil {
			return fmt.Errorf("failed to find relative path of '%s' given '%s': %v", destPath, wiki.DestDir, err)
		}

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

	// Delete empty directories.
	deleteEmptyDirectories(wiki.DestDir)

	return nil
}
