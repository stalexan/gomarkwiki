// Package wiki generates HTML from markdown for a given wiki.
package wiki

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/stalexan/gomarkwiki/internal/util"
)

// Wiki stores data about a single wiki.
type Wiki struct {
	// Directories
	SourceDir  string // Wiki source directory
	ContentDir string // Content directory within source directory
	DestDir    string // Dest directory where wiki will be generated

	subStrings [][2]string // Substitution strings. Each pair is the string to look for and what to replace it with.
	subsPath   string      // Path to substitution strings file.

	ignore     []*regexp.Regexp // Which files to ingore
	ignorePath string           // Path to ignore.txt file.
}

// NewWiki constructs a new instance of Wiki.
func NewWiki(sourceDir, destDir string) (*Wiki, error) {
	// Resolve absolute paths for comparison
	absSourceDir, err := filepath.Abs(sourceDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path of source directory '%s': %v", sourceDir, err)
	}
	absDestDir, err := filepath.Abs(destDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path of destination directory '%s': %v", destDir, err)
	}

	// Use EvalSymlinks to handle cases where one path might be a symlink to the other
	evalSourceDir, err := filepath.EvalSymlinks(absSourceDir)
	if err != nil {
		// If EvalSymlinks fails, fall back to absSourceDir
		evalSourceDir = absSourceDir
	}
	evalDestDir, err := filepath.EvalSymlinks(absDestDir)
	if err != nil {
		// If EvalSymlinks fails, fall back to absDestDir
		evalDestDir = absDestDir
	}

	// Check that source and dest directories are not the same
	if evalSourceDir == evalDestDir {
		return nil, fmt.Errorf("source directory '%s' and destination directory '%s' are the same", sourceDir, destDir)
	}

	// Check that dest directory is not under source directory
	relPath, err := filepath.Rel(evalSourceDir, evalDestDir)
	if err != nil {
		// If Rel fails, paths might be on different volumes (Windows), so they're not nested
		// Continue without error in this case
	} else {
		// If relative path doesn't start with "..", dest is under source
		if relPath != "." && !strings.HasPrefix(relPath, "..") {
			return nil, fmt.Errorf("destination directory '%s' cannot be under source directory '%s'", destDir, sourceDir)
		}
	}

	// Check that source directory is not under dest directory
	relPath2, err2 := filepath.Rel(evalDestDir, evalSourceDir)
	if err2 == nil {
		if relPath2 != "." && !strings.HasPrefix(relPath2, "..") {
			return nil, fmt.Errorf("source directory '%s' cannot be under destination directory '%s'", sourceDir, destDir)
		}
	}

	wiki := Wiki{
		SourceDir:  sourceDir,
		ContentDir: filepath.Join(sourceDir, "content"),
		DestDir:    destDir,
		subStrings: nil,
		subsPath:   "",
		ignore:     nil,
		ignorePath: "",
	}

	// Check that the dirs in Wiki exist.
	for _, dir := range []string{wiki.SourceDir, wiki.ContentDir, wiki.DestDir} {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return nil, fmt.Errorf("directory '%s' not found", dir)
		}
	}

	// Load substitution strings.
	if err := wiki.loadSubstitutionStrings(); err != nil {
		return nil, err
	}

	// Load ignore expressions.
	if err := wiki.loadIgnoreExpressions(); err != nil {
		return nil, err
	}

	return &wiki, nil
}

// Generate generates a wiki and then optionally watches for changes in the
// wiki to regenerate files on the fly.
func (wiki *Wiki) Generate(ctx context.Context, regen, clean, watch bool, version string) error {
	util.PrintVerbose("Generating wiki '%s' from '%s'", wiki.DestDir, wiki.SourceDir)

	// Check for cancellation before starting
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Generate wiki.
	if err := wiki.generate(ctx, regen, clean, version); err != nil {
		return fmt.Errorf("failed to generate wiki '%s': %v", wiki.SourceDir, err)
	}

	// Watch for changes and regenerate files on the fly.
	if watch {
		if err := wiki.watch(ctx, clean, version); err != nil {
			// Don't wrap context.Canceled errors
			if err == context.Canceled {
				return err
			}
			return fmt.Errorf("failed to watch '%s': %v", wiki.ContentDir, err)
		}
	}

	return nil
}

// generate generates the wiki.
func (wiki *Wiki) generate(ctx context.Context, regen, clean bool, version string) error {
	// Check for cancellation before starting
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Generate the part of the wiki that comes from content found in the source dir.
	var relDestPaths map[string]bool
	var processingErr error // Store error but don't return immediately
	if relDestPaths, processingErr = wiki.generateFromContent(ctx, regen, version); processingErr != nil {
		// Log but continue - we still want CSS and cleanup for successfully processed files
		util.PrintError(processingErr, "some files failed to process")
	}

	// Check for cancellation before copying CSS files
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Copy css files to destDir (even with partial results).
	if err := wiki.copyCssFiles(ctx, relDestPaths); err != nil {
		return err
	}

	// Check for cancellation before cleaning
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Clean dest dir.
	// Only clean if we have a valid list of generated files.
	// If relDestPaths is nil (critical failure), skipping clean prevents wiping the directory.
	if clean && relDestPaths != nil {
		if err := wiki.cleanDestDir(ctx, relDestPaths); err != nil {
			return fmt.Errorf("failed to clean dest dir '%s': %v", wiki.DestDir, err)
		}
	}

	// Return the processing error at the end (if any)
	return processingErr
}
