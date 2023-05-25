// Package generator generates HTML from markdown.
package generator

import (
	"bytes"
	"embed"
	"encoding/csv"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"

	"github.com/stalexan/gomarkwiki/internal/util"
)

//go:embed static/style.css
var embeddedFileSystem embed.FS

var markdown goldmark.Markdown
var htmlHeaderTemplate *template.Template

func init() {
	// Create markdown converter.
	markdown = goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
			parser.WithAttribute(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)

	// Create HTML header template.
	htmlHeaderTemplate = template.Must(template.New("html").Parse(htmlHeaderTemplateText))
}

// WikiDirs stores the source files are located and where to store the generated wiki.
type WikiDirs struct {
	sourceDir  string
	contentDir string
	destDir    string
}

// NewWikiDirs constructs a new instance of WikiDirs.
func NewWikiDirs(sourceDir, destDir string) WikiDirs {
	return WikiDirs{
		sourceDir:  sourceDir,
		contentDir: filepath.Join(sourceDir, "content"),
		destDir:    destDir,
	}
}

// CheckDirs checks that the dirs in Wikidirs exist.
func (dirs WikiDirs) CheckDirs() error {
	for _, dir := range []string{dirs.sourceDir, dirs.contentDir, dirs.destDir} {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return fmt.Errorf("directory '%s' not found", dir)
		}
	}
	return nil
}

// GenerateWiki generates the wiki.
func GenerateWiki(dirs WikiDirs, regen bool, clean bool, version string) error {
	// Do directories exist?
	var err error
	if err = dirs.CheckDirs(); err != nil {
		return err
	}

	// Generate the part of the wiki that comes from the source content.
	var relDestPaths map[string]bool
	if relDestPaths, err = generateFromContent(dirs, regen, version); err != nil {
		return err
	}

	// Copy styles.css to destDir.
	if err = copyStyleCss(dirs.destDir); err != nil {
		return err
	}
	relDestPaths["style.css"] = true

	// Clean dest dir.
	if clean {
		if err = cleanDestDir(dirs.destDir, relDestPaths); err != nil {
			return fmt.Errorf("failed to clean dest dir %s: %v", dirs.destDir, err)
		}
	}

	return nil
}

// generateFromContent generates the part of the wiki that comes from the source content.
func generateFromContent(dirs WikiDirs, regen bool, version string) (map[string]bool, error) {
	// Load the substition strings.
	if err := loadSubstitionStrings(dirs.sourceDir); err != nil {
		return nil, err
	}

	// Iterate recursively over the source directory and generate the wiki from the files found.
	util.PrintVerboseMessage("Looking for markdown in %s", dirs.contentDir)
	util.PrintVerboseMessage("Writing HTML to %s", dirs.destDir)
	relDestPaths := map[string]bool{}
	err := filepath.Walk(dirs.contentDir, func(contentPath string, info fs.FileInfo, err error) error {
		// Was there an error looking up this file?
		if err != nil {
			return err
		}

		// Is this file regular and readable?
		if !isReadableFile(info, contentPath) {
			return nil
		}

		// What's the relative path to this file with respect to the content dir?
		var relContentPath string
		relContentPath, err = filepath.Rel(dirs.contentDir, contentPath)
		if err != nil {
			return fmt.Errorf("failed to find relative path of %s given %s: %v", contentPath, dirs.contentDir, err)
		}

		// Create the dest version of this file.
		var relDestPath string
		if isPathMarkdown(contentPath) {
			// Generate HTML from markdown.
			relDestPath, err = generateHtmlFromMarkdown(info, contentPath, relContentPath, dirs.destDir, regen, version)
			if err != nil {
				return err
			}
		} else {
			// This is not a markdown file. Just copy it.
			if err := copyFileToDest(info, contentPath, relContentPath, dirs.destDir, regen); err != nil {
				return err
			}
			relDestPath = relContentPath
		}

		// Record that this file corresponds to a file from the source dir.
		relDestPaths[relDestPath] = true

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("generate destination content failed: %v", err)
	}

	return relDestPaths, nil
}

// cleanDestDir cleans the dest dir by any deleting files that don't have any
// a corresponding source file, and by deleting any empty directories.
func cleanDestDir(destDir string, relDestPaths map[string]bool) error {
	// Delete dest files that don't have a corresponding source file.
	err := filepath.Walk(destDir, func(destPath string, info fs.FileInfo, err error) error {
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
		relDestPath, err = filepath.Rel(destDir, destPath)
		if err != nil {
			return fmt.Errorf("failed to find relative path of %s given %s: %v", destPath, destDir, err)
		}

		// Delete this file if it doesn't have a corresponding file in the source dir.
		if !relDestPaths[relDestPath] {
			util.PrintVerboseMessage("Deleting %s", destPath)
			if err = os.Remove(destPath); err != nil {
				util.PrintWarning("Failed to delete %s: %v", destPath, err)
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("cleaning destination failed: %v", err)
	}

	// Delete empty directories.
	deleteEmptyDirectories(destDir)

	return nil
}

// Recursively deletes empty directories using depth-first search

// deleteEmptyDirectories deletes any empty directories within path,
// including directories that have just empty directories.
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
				util.PrintVerboseMessage("Deleting empty directory %s", entryPath)
				err := os.Remove(entryPath)
				if err != nil {
					return err
				}
			}
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

// subStrings holds the substitution strings. Each pair is the string to look
// for and what to replace it with.
var subStrings [][2]string

// loadSubstitionStrings loads substition strings from substition-strings.csv
func loadSubstitionStrings(sourceDir string) error {
	// Is there a substition strings file?
	const subsFileName = "substition-strings.csv"
	subsPath := filepath.Join(sourceDir, subsFileName)
	var err error
	if _, err = os.Stat(subsPath); err != nil {
		// There are no substitions to make.
		return nil
	}

	// Open substition strings file.
	util.PrintVerboseMessage("Loading substition strings from %v", subsPath)
	var file *os.File
	if file, err = os.Open(subsPath); err != nil {
		return fmt.Errorf("unable to open %s: %v", subsFileName, err)
	}
	defer file.Close()

	// Read contents of substition strings file.
	reader := csv.NewReader(file)
	reader.Comma = ','
	reader.FieldsPerRecord = 2
	var records [][]string
	if records, err = reader.ReadAll(); err != nil {
		return fmt.Errorf("unable to read %s: %v", subsFileName, err)
	}

	// Save substitions.
	for _, record := range records {
		placeholder := record[0]
		if len(placeholder) == 0 {
			continue
		}
		placeholder = fmt.Sprintf("{{%s}}", placeholder)
		substitution := record[1]
		subStrings = append(subStrings, [2]string{placeholder, substitution})
	}

	return nil
}

// makeSubstitions makes substitions in data.
func makeSubstitions(data []byte) []byte {
	for _, pair := range subStrings {
		data = bytes.ReplaceAll(data, []byte(pair[0]), []byte(pair[1]))
	}
	return data
}

// markdownExts holds markdown file extensions.
var markdownExts = [...]string{".md", ".mdwn", ".markdown"}

// isPathMarkdown determines whether the file name ends with a markdown extension.
func isPathMarkdown(path string) bool {
	ext := filepath.Ext(path)
	for _, markdownExt := range markdownExts {
		if ext == markdownExt {
			return true
		}
	}
	return false
}

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
			util.PrintWarning("Skipping not readable file %s", path)
			return false
		}
	} else {
		util.PrintWarning("Skipping not regular file %s", path)
		return false
	}

	// This is readable file.
	return true
}

// htmlHeaderTemplateText is the text used to create the HTML template that
// generates the start of each HTML file.
const htmlHeaderTemplateText = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8" />
<meta name="viewport" content="width=device-width, initial-scale=1" />
<meta name=generator content="gomarkwiki {{.Version}}">
<title>{{.Title}}</title>
<link rel="icon" href="{{.RootRelPath}}favicon.ico" type="image/x-icon" />
<link href="{{.RootRelPath}}style.css" rel="stylesheet" />
<link href="{{.RootRelPath}}local.css" rel="stylesheet" />
</head>
<body>
`

// templateData holds the values used to instantiate HTML from the HTML header template.
type templateData struct {
	Title       string
	Version     string
	RootRelPath string
}

// generateHtml generates an HTML file from a markdown file.
func generateHtmlFromMarkdown(mdInfo fs.FileInfo, mdPath, mdRelPath, destDir string, regen bool, version string) (string, error) {
	// Determine the output path for the HTML file. For example, if the markdown
	// relative path (mdRelPath) is Foo/Bar.mdwn and the destination directory (destDir)
	// is /wiki-html, the output path (outPath) is /wiki-html/Foo/Bar.html.
	relPathNoExt := removeFileExtension(mdRelPath)
	relOutPath := relPathNoExt + ".html"
	outPath := filepath.Join(destDir, relOutPath)
	outDir := filepath.Dir(outPath)

	// Skip generating the HTML if markdown is older than current HTML.
	if !regen && destIsOlder(mdInfo, outPath) {
		return relOutPath, nil
	}
	util.PrintVerboseMessage("Generating %s", relOutPath)

	// Read markdown file.
	var data []byte
	var err error
	if data, err = os.ReadFile(mdPath); err != nil {
		return "", fmt.Errorf("failed to read markdown file %s: %v", mdPath, err)
	}

	// Make substituions.
	data = makeSubstitions(data)

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
	if err = htmlHeaderTemplate.Execute(html, templateData{title, version, rootRelPath}); err != nil {
		return "", fmt.Errorf("failed to create HTML header for %s: %v", outPath, err)
	}

	// Generate the body of the HTML from markdown.
	if err = markdown.Convert(data, html); err != nil {
		return "", fmt.Errorf("failed to generate HTML for %s: %v", outPath, err)
	}

	// Generate end of HTML file.
	html.WriteString("</body>\n</html>")

	// Create output directory if necessary.
	if err = os.MkdirAll(outDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %v", outDir, err)
	}

	// Write out the HTML file.
	if err := os.WriteFile(outPath, []byte(html.String()), 0644); err != nil {
		return "", fmt.Errorf("failed to write HTML file %s: %v", outPath, err)
	}

	return relOutPath, nil
}

// destIsOlder returns true if dest is older than source.
func destIsOlder(sourceInfo fs.FileInfo, destPath string) bool {
	destInfo, err := os.Stat(destPath)
	if err == nil && sourceInfo.ModTime().Before(destInfo.ModTime()) {
		return true
	}
	return false
}

// removeFileExtention removes the file extention from path; e.g. Foo/Bar.md becomes Foo/Bar
func removeFileExtension(path string) string {
	extension := filepath.Ext(path)
	return path[:len(path)-len(extension)]
}

// copyFileToDest copies a file from the source dir to the dest dir.
func copyFileToDest(sourceInfo fs.FileInfo, sourcePath, sourceRelPath, destDir string, regen bool) error {
	// Skip copying if source is older than dest.
	destPath := filepath.Join(destDir, sourceRelPath)
	if !regen && destIsOlder(sourceInfo, destPath) {
		return nil
	}

	// Copy file.
	util.PrintVerboseMessage("Copying %s", sourceRelPath)
	var source *os.File
	var err error
	if source, err = os.Open(sourcePath); err != nil {
		return fmt.Errorf("failed to open %s: %v", sourcePath, err)
	}
	defer source.Close()
	if err := copyToFile(destPath, source); err != nil {
		return err
	}

	return nil
}

// copyStyleCss copies styles.css to destDir.
func copyStyleCss(destDir string) error {
	// Copy style.css.
	var styleCss []byte
	var err error
	util.PrintVerboseMessage("Copying style.css")
	if styleCss, err = embeddedFileSystem.ReadFile("static/style.css"); err != nil {
		return fmt.Errorf("failed to read style.css: %v", err)
	}
	if err := copyToFile(fmt.Sprintf("%s/style.css", destDir), bytes.NewReader(styleCss)); err != nil {
		return err
	}

	return nil
}

// copyFile copies source to the file at destPath.
func copyToFile(destPath string, source io.Reader) error {
	// Create dest file.
	var destFile *os.File
	var err error
	if destFile, err = os.Create(destPath); err != nil {
		return fmt.Errorf("failed to open file %s: %v", destPath, err)
	}
	defer destFile.Close()

	// Copy source to destFile.
	if _, err = io.Copy(destFile, source); err != nil {
		return fmt.Errorf("failed to write to %s: %v", destPath, err)
	}

	return nil
}
