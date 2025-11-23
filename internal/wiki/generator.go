// Package wiki generates HTML from markdown for a given wiki.
package wiki

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/stalexan/gomarkwiki/internal/util"
)

// markdownExts specifies markdown file exptensions.
var markdownExts = [...]string{".md", ".mdwn", ".markdown"}

// isPathMarkdown returns true if path has a markdown extension.
func isPathMarkdown(path string) bool {
	ext := filepath.Ext(path)
	for _, markdownExt := range markdownExts {
		if ext == markdownExt {
			return true
		}
	}
	return false
}

// removeFileExtention removes the file extention from path; e.g. Foo/Bar.md becomes Foo/Bar
func removeFileExtension(path string) string {
	extension := filepath.Ext(path)
	return path[:len(path)-len(extension)]
}

var gitHubDirective []byte = []byte("#[style(github)]")

// checkForStyleDirective looks for the GitHub style directive on the first
// line of `data`. Returns true if found and removes directive. Otherwise,
// returns false.
func checkForStyleDirective(data []byte) (bool, []byte) {
	// Check for directive
	hasDirective := false
	if bytes.HasPrefix(data, gitHubDirective) {
		hasDirective = true
		// Trim off the directive and any whitespace.
		data = data[len(gitHubDirective):]
		data = bytes.TrimSpace(data)
	}
	return hasDirective, data
}

// generateHtmlFromMarkdown generates an HTML file from a markdown file.
func (wiki Wiki) generateHtmlFromMarkdown(mdInfo fs.FileInfo, mdPath, mdRelPath string, regen bool, version string) (string, error) {
	// Determine the output path for the HTML file. For example, if the markdown
	// relative path (mdRelPath) is Foo/Bar.mdwn and the destination directory (destDir)
	// is /wiki-html, the output path (outPath) is /wiki-html/Foo/Bar.html.
	relPathNoExt := removeFileExtension(mdRelPath)
	relOutPath := relPathNoExt + ".html"
	outPath := filepath.Join(wiki.DestDir, relOutPath)
	outDir := filepath.Dir(outPath)

	// Skip generating the HTML if markdown is older than current HTML.
	if !regen && destIsOlder(mdInfo, outPath) {
		return relOutPath, nil
	}
	util.PrintVerbose("Generating '%s'", outPath)

	// Read markdown file.
	var data []byte
	var err error
	if data, err = os.ReadFile(mdPath); err != nil {
		if os.IsNotExist(err) {
			util.PrintVerbose("markdown '%s' no longer exists and so no HTML will be generated for it", mdPath)
			return "", nil
		} else {
			return "", fmt.Errorf("failed to read markdown file '%s': %v", mdPath, err)
		}
	}

	// Check for style directive.
	useGitHubStyle, data := checkForStyleDirective(data)

	// Make substitutions.
	data = wiki.makeSubstitutions(data)

	// Determine relative path from the file being generated to the dest dir. For
	// example if the file being generated is /wiki-html/Foo/Bar.html and the
	// dest dir is /wiki-html, the relative path is ../
	relPathJustDir := filepath.Dir(mdRelPath)
	dirCount := 0
	if relPathJustDir != "." {
		dirCount = strings.Count(relPathJustDir, "/") + 1
	}
	rootRelPath := strings.Repeat("../", dirCount)

	// Generate the start of the HTML file using the template htmlHeaderTemplate.
	html := &strings.Builder{}
	title := filepath.Base(relPathNoExt) // Markdown file name without file extension
	if useGitHubStyle {
		if err = githubHtmlHeaderTemplate.Execute(html, templateData{title, version, rootRelPath}); err != nil {
			return "", fmt.Errorf("failed to create GitHub HTML header for '%s': %v", outPath, err)
		}
	} else {
		if err = defaultHtmlHeaderTemplate.Execute(html, templateData{title, version, rootRelPath}); err != nil {
			return "", fmt.Errorf("failed to create default HTML header for '%s': %v", outPath, err)
		}
	}

	// Generate the body of the HTML from markdown.
	if err = markdown.Convert(data, html); err != nil {
		return "", fmt.Errorf("failed to generate HTML body for '%s': %v", outPath, err)
	}

	// Generate end of HTML file.
	if useGitHubStyle {
		html.WriteString("</article>")
	} else {
		html.WriteString("</body>\n</html>")
	}

	// Create output directory if necessary.
	if err = os.MkdirAll(outDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory '%s': %v", outDir, err)
	}

	// Write out the HTML file.
	if err := os.WriteFile(outPath, []byte(html.String()), 0644); err != nil {
		return "", fmt.Errorf("failed to write HTML file '%s': %v", outPath, err)
	}

	return relOutPath, nil
}

// generateFromContent generates the part of the wiki that comes from the source content.
func (wiki Wiki) generateFromContent(regen bool, version string) (map[string]bool, error) {
	// Walk the source directory and generate the wiki from the files found.
	util.PrintDebug("Generating wiki '%s' from '%s'", wiki.DestDir, wiki.SourceDir)
	relDestPaths := map[string]bool{}
	err := filepath.Walk(wiki.ContentDir, func(contentPath string, info fs.FileInfo, err error) error {
		// Was there an error looking up this file?
		if err != nil {
			util.PrintError(err, "failed to lookup info on '%s'", contentPath)
			return nil
		}

		// Is this file regular and readable?
		if !isReadableFile(info, contentPath) {
			return nil
		}

		// Ignore this file?
		if wiki.ignoreFile(contentPath) {
			util.PrintVerbose("Ignoring '%s'", contentPath)
			return nil
		}

		// What's the relative path to this file with respect to the content dir?
		var relContentPath string
		relContentPath, err = filepath.Rel(wiki.ContentDir, contentPath)
		if err != nil {
			util.PrintError(err, "failed to find relative path of '%s' given '%s'", contentPath, wiki.ContentDir)
			return nil
		}

		// Create the dest version of this file.
		var relDestPath string
		if isPathMarkdown(contentPath) {
			// Generate HTML from markdown.
			relDestPath, err = wiki.generateHtmlFromMarkdown(info, contentPath, relContentPath, regen, version)
			if err != nil {
				util.PrintError(err, "failed to find generate HTML for '%s'", contentPath)
				return nil
			}
		} else {
			// This is not a markdown file. Just copy it.
			if err := wiki.copyFileToDest(info, contentPath, relContentPath, regen); err != nil {
				util.PrintError(err, "failed to copy '%s' to dest", contentPath)
				return nil
			}
			relDestPath = relContentPath
		}

		// Record that this file corresponds to a file from the source dir.
		if relDestPath != "" {
			relDestPaths[relDestPath] = true
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("generate destination content failed: %v", err)
	}

	return relDestPaths, nil
}
