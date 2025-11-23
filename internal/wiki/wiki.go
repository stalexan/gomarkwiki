// Package wiki generates HTML from markdown for a given wiki.
package wiki

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/stalexan/gomarkwiki/internal/util"
)

// Wiki stores data about a single wiki.
type Wiki struct {
	// Directories
	SourceDir  string // Wiki source directory
	ContentDir string // Content directory within source directory
	DestDir    string // Dest directory where wiki will be generated

	styleCssCopyNeeded bool // Whether CSS files needs to be copied to dest

	subStrings [][2]string // Substitution strings. Each pair is the string to look for and what to replace it with.
	subsPath   string      // Path to substitution strings file.

	ignore []*regexp.Regexp // Which files to ingore
}

// NewWiki constructs a new instance of Wiki.
func NewWiki(sourceDir, destDir string) (*Wiki, error) {
	wiki := Wiki{
		SourceDir:          sourceDir,
		ContentDir:         filepath.Join(sourceDir, "content"),
		DestDir:            destDir,
		styleCssCopyNeeded: true,
		subStrings:         nil,
		subsPath:           "",
		ignore:             nil,
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
func (wiki *Wiki) Generate(regen, clean, watch bool, version string) error {
	util.PrintVerbose("Generating wiki '%s' from '%s'", wiki.DestDir, wiki.SourceDir)

	// Generate wiki.
	if err := wiki.generate(regen, clean, version); err != nil {
		return fmt.Errorf("failed to generate wiki '%s': %v", wiki.SourceDir, err)
	}

	// Watch for changes and regenerate files on the fly.
	if watch {
		if err := wiki.watch(clean, version); err != nil {
			return fmt.Errorf("failed to watch '%s': %v", wiki.ContentDir, err)
		}
	}

	return nil
}

// generate generates the wiki.
func (wiki *Wiki) generate(regen, clean bool, version string) error {
	// Generate the part of the wiki that comes from content found in the source dir.
	var relDestPaths map[string]bool
	var err error
	if relDestPaths, err = wiki.generateFromContent(regen, version); err != nil {
		return err
	}

	// Copy css files to destDir.
	if err = wiki.copyCssFiles(relDestPaths); err != nil {
		return err
	}

	// Clean dest dir.
	if clean {
		if err = wiki.cleanDestDir(relDestPaths); err != nil {
			return fmt.Errorf("failed to clean dest dir '%s': %v", wiki.DestDir, err)
		}
	}

	return nil
}
