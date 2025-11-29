// Package wiki generates HTML from markdown for a given wiki.
package wiki

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/stalexan/gomarkwiki/internal/util"
)

// markdownExts specifies markdown file extensions.
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

// removeFileExtension removes the file extension from path; e.g. Foo/Bar.md becomes Foo/Bar
func removeFileExtension(path string) string {
	extension := filepath.Ext(path)
	return path[:len(path)-len(extension)]
}

var gitHubDirective []byte = []byte("#[style(github)]")

// checkForStyleDirective looks for the GitHub style directive on the first
// line of `data`. Returns true if found and removes directive. Otherwise,
// returns false.
func checkForStyleDirective(data []byte) (bool, []byte) {
	// Strip UTF-8 BOM if present
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		data = data[3:]
	}

	// Trim leading whitespace only for checking the directive
	trimmed := bytes.TrimLeft(data, " \t\n\r")

	if bytes.HasPrefix(trimmed, gitHubDirective) {
		// Directive found: remove it and trim trailing whitespace
		trimmed = trimmed[len(gitHubDirective):]
		return true, bytes.TrimSpace(trimmed)
	}

	// No directive found: return data (BOM stripped if present, otherwise unchanged)
	return false, data
}

// generateHtmlFromMarkdown generates an HTML file from a markdown file.
// relDestPath is the relative destination path (e.g., "Foo/Bar.html") that was
// already computed for collision detection in the caller.
func (wiki Wiki) generateHtmlFromMarkdown(mdInfo fs.FileInfo, mdPath, mdRelPath, relDestPath string, regen bool, version string) (string, error) {
	// Compute the full output path. For example, if relDestPath is Foo/Bar.html
	// and the destination directory (destDir) is /wiki-html, the output path is /wiki-html/Foo/Bar.html.
	outPath := filepath.Join(wiki.DestDir, relDestPath)
	outDir := filepath.Dir(outPath)

	// Skip generating the HTML if markdown is older than current HTML.
	// Note: We use mdInfo here, but if the file is deleted before reading, we'll handle that below.
	if !regen && sourceIsOlder(mdInfo, outPath) {
		return relDestPath, nil
	}
	util.PrintVerbose("Generating '%s'", outPath)

	// Re-stat the file right before reading to get current info and prevent TOCTOU issues.
	currentInfo, err := os.Stat(mdPath)
	if err != nil {
		if os.IsNotExist(err) {
			util.PrintVerbose("markdown '%s' no longer exists and so no HTML will be generated for it", mdPath)
			return "", nil
		} else {
			return "", fmt.Errorf("failed to stat markdown file '%s': %v", mdPath, err)
		}
	}

	// Check file size before reading to prevent resource exhaustion
	if currentInfo.Size() > MaxMarkdownFileSize {
		return "", fmt.Errorf("markdown file '%s' is too large (%d bytes, max %d bytes)", mdPath, currentInfo.Size(), MaxMarkdownFileSize)
	}

	// Read markdown file.
	var data []byte
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
		dirCount = strings.Count(relPathJustDir, string(filepath.Separator)) + 1
	}
	rootRelPath := strings.Repeat("../", dirCount)

	// Generate the start of the HTML file using the template htmlHeaderTemplate.
	html := &strings.Builder{}
	// Extract title from file path. The html/template package automatically escapes
	// all template variables (including this title) to prevent XSS attacks, so
	// special characters in file paths are safely handled.
	relPathNoExt := removeFileExtension(relDestPath) // Remove .html extension for title
	title := filepath.Base(relPathNoExt)             // Markdown file name without file extension
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

	return relDestPath, nil
}

// generateFromContent generates the part of the wiki that comes from the source content.
//
// Collision detection: If multiple markdown files would generate the same HTML path
// (e.g., "foo.md" and "foo.markdown" both generate "foo.html"), the first file
// encountered during the walk wins, and subsequent files are skipped with a warning.
// The ordering is deterministic because filepath.Walk processes files in lexicographic
// order (as returned by the filesystem), ensuring consistent collision resolution
// across regeneration cycles in watch mode.
func (wiki Wiki) generateFromContent(ctx context.Context, regen bool, version string) (map[string]bool, error) {
	// Walk the source directory and generate the wiki from the files found.
	util.PrintDebug("Generating wiki '%s' from '%s'", wiki.DestDir, wiki.SourceDir)
	relDestPaths := map[string]bool{}
	sourceFileMap := map[string]string{} // Track which source file generated each HTML path
	fileCount := 0
	baseDepth := strings.Count(wiki.ContentDir, string(filepath.Separator))
	var processingErrors []error
	err := filepath.Walk(wiki.ContentDir, func(contentPath string, info fs.FileInfo, err error) error {
		// Check for cancellation periodically during walk
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Check recursion depth
		currentDepth := strings.Count(contentPath, string(filepath.Separator)) - baseDepth
		if currentDepth > MaxRecursionDepth {
			return fmt.Errorf("directory recursion depth exceeded at '%s' (depth %d, max %d)", contentPath, currentDepth, MaxRecursionDepth)
		}

		// Was there an error looking up this file?
		if err != nil {
			util.PrintError(err, "failed to lookup info on '%s'", contentPath)
			processingErrors = append(processingErrors, fmt.Errorf("failed to lookup info on '%s': %w", contentPath, err))
			return nil
		}

		// Is this file regular and readable?
		if !isReadableFile(info, contentPath) {
			// Warn if this is a symlink to a directory
			warnIfSymlinkToDir(info, contentPath)
			return nil
		}

		// Check file count limit
		fileCount++
		if fileCount > MaxFilesProcessed {
			return fmt.Errorf("maximum number of files processed exceeded (%d files, max %d files)", fileCount, MaxFilesProcessed)
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
			processingErrors = append(processingErrors, fmt.Errorf("failed to find relative path of '%s': %w", contentPath, err))
			return nil
		}

		// Create the dest version of this file.
		var relDestPath string
		if isPathMarkdown(contentPath) {
			// Determine the output path for the HTML file.
			relPathNoExt := removeFileExtension(relContentPath)
			relDestPath = relPathNoExt + ".html"

			// Check for collision with previously processed markdown files.
			if existingSource, collision := sourceFileMap[relDestPath]; collision {
				util.PrintWarning("Skipping '%s': would generate '%s' which is already generated by '%s'", relContentPath, relDestPath, existingSource)
				return nil
			}

			// Generate HTML from markdown.
			relDestPath, err = wiki.generateHtmlFromMarkdown(info, contentPath, relContentPath, relDestPath, regen, version)
			if err != nil {
				util.PrintError(err, "failed to generate HTML for '%s'", contentPath)
				processingErrors = append(processingErrors, fmt.Errorf("failed to generate HTML for '%s': %w", contentPath, err))
				return nil
			}

			// Record the source file that generated this HTML path.
			if relDestPath != "" {
				sourceFileMap[relDestPath] = relContentPath
			}
		} else {
			// This is not a markdown file. Just copy it.
			if err := wiki.copyFileToDest(ctx, info, contentPath, relContentPath, regen); err != nil {
				util.PrintError(err, "failed to copy '%s' to dest", contentPath)
				processingErrors = append(processingErrors, fmt.Errorf("failed to copy '%s': %w", contentPath, err))
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

	// Return collected processing errors if any occurred
	if len(processingErrors) > 0 {
		errMsg := fmt.Sprintf("failed to process %d file(s)", len(processingErrors))
		for i, e := range processingErrors {
			errMsg += fmt.Sprintf("\n  %d. %v", i+1, e)
		}
		return relDestPaths, fmt.Errorf("%s", errMsg)
	}

	return relDestPaths, nil
}
