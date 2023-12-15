// Package wiki generates HTML from markdown for a given wiki.
package wiki

import (
	"bufio"
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"

	"github.com/stalexan/gomarkwiki/internal/util"
)

var markdown goldmark.Markdown
var defaultHtmlHeaderTemplate *template.Template
var githubHtmlHeaderTemplate *template.Template

//go:embed static/style.css static/github-style.css
var embeddedFileSystem embed.FS

// Times to wait while waiting for changes to finish.
const CHANGE_WAIT = 100      // milliseconds
const MAX_CHANGE_WAIT = 5000 // milliseconds

// Max time before regenerating wiki while watching for changes.
const MAX_REGEN_INTERVAL = 10 // minutes

// defaultHtmlHeaderTemplateText is the text used to create the HTML template that
// generates the start of each HTML file that uses default styles.
const defaultHtmlHeaderTemplateText = `<!doctype html>
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

// githubHtmlHeaderTemplateText is the text used to create the HTML template that
// generates the start of each HTML file that uses GitHub styles.
const githubHtmlHeaderTemplateText = `<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name=generator content="gomarkwiki {{.Version}}">
<title>{{.Title}}</title>
<link rel="icon" href="{{.RootRelPath}}favicon.ico" type="image/x-icon" />
<link href="{{.RootRelPath}}github-style.css" rel="stylesheet" />
<link href="{{.RootRelPath}}github-local.css" rel="stylesheet" />
<style>
	.markdown-body {
		box-sizing: border-box;
		min-width: 200px;
		max-width: 980px;
		margin: 0 auto;
		padding: 45px;
	}

	@media (max-width: 767px) {
		.markdown-body {
			padding: 15px;
		}
	}
</style>
<article class="markdown-body">
`

// templateData holds the values used to instantiate HTML from the HTML header template.
type templateData struct {
	Title       string
	Version     string
	RootRelPath string
}

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

	// Create HTML header templates.
	defaultHtmlHeaderTemplate = template.Must(template.New("defaultHtml").Parse(defaultHtmlHeaderTemplateText))
	githubHtmlHeaderTemplate = template.Must(template.New("githubHtml").Parse(githubHtmlHeaderTemplateText))
}

// Wiki stores data about a single wiki.
type Wiki struct {
	// Directories
	SourceDir  string // Wiki source directory
	ContentDir string // Content directory within source directory
	DestDir    string // Dest directory where wiki will be generated

	styleCssCopyNeeded bool // Whether CSS files needs to be copied to dest

	subStrings [][2]string // Substitution strings. Each pair is the string to look for and what to replace it with.

	ignore []*regexp.Regexp // Which files to ingore
}

// NewWikiDirs constructs a new instance of WikiDirs.
func NewWiki(sourceDir, destDir string) (*Wiki, error) {
	wiki := Wiki{
		SourceDir:          sourceDir,
		ContentDir:         filepath.Join(sourceDir, "content"),
		DestDir:            destDir,
		styleCssCopyNeeded: true,
		subStrings:         nil,
		ignore:             nil,
	}

	// Check that the dirs in Wiki exist.
	for _, dir := range []string{wiki.SourceDir, wiki.ContentDir, wiki.DestDir} {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return nil, fmt.Errorf("directory '%s' not found", dir)
		}
	}

	// Load substition strings.
	if err := wiki.loadSubstitionStrings(); err != nil {
		return nil, err
	}

	// Load ignore expressions.
	if err := wiki.loadIgnoreExpressions(); err != nil {
		return nil, err
	}

	return &wiki, nil
}

// loadSubstitionStrings loads substition strings for a wiki, from its substition-strings.csv
func (wiki *Wiki) loadSubstitionStrings() error {
	// Is there a substition strings file?
	const subsFileName = "substition-strings.csv"
	subsPath := filepath.Join(wiki.SourceDir, subsFileName)
	var pairs [][2]string
	var err error
	if pairs, err = util.LoadStringPairs(subsPath); err != nil {
		return fmt.Errorf("failed to load substition strings from '%s': %v", subsPath, err)
	}
	if len(pairs) == 0 {
		// There's either no substituion strings file or the file is empty.
		return nil
	}

	// Save substitions.
	for _, pair := range pairs {
		placeholder := pair[0]
		if len(placeholder) == 0 {
			continue
		}
		placeholder = fmt.Sprintf("{{%s}}", placeholder)
		substitution := pair[1]
		wiki.subStrings = append(wiki.subStrings, [2]string{placeholder, substitution})
	}

	return nil
}

// makeSubstitions makes string substitions in data.
func (wiki Wiki) makeSubstitions(data []byte) []byte {
	for _, pair := range wiki.subStrings {
		data = bytes.ReplaceAll(data, []byte(pair[0]), []byte(pair[1]))
	}
	return data
}

// loadIngoreExpressions loads regular expressions that define which files to ingore.
func (wiki *Wiki) loadIgnoreExpressions() error {
	// Open ingore file, if there is one.
	const ignoreFileName = "ignore.txt"
	ignorePath := filepath.Join(wiki.SourceDir, ignoreFileName)
	var file *os.File
	var err error
	if file, err = os.Open(ignorePath); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("unable to open '%s': %v", ignorePath, err)
		} else {
			// There is no ignore file.
			return nil
		}
	}
	defer file.Close()

	// Read expressions.
	scanner := bufio.NewScanner(file)
	lineCount := 0
	for scanner.Scan() {
		lineCount++
		line := scanner.Text()
		expression, err := regexp.Compile(line)
		if err != nil {
			return fmt.Errorf("error compiling regular expression '%s' on line %d: %v", line, lineCount, err)
		}
		wiki.ignore = append(wiki.ignore, expression)
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading '%s': %v", ignorePath, err)
	}

	return nil
}

// ignoreFile returns true if the file at path should be ignored.
func (wiki Wiki) ignoreFile(path string) bool {
	for _, expr := range wiki.ignore {
		if expr.MatchString(path) {
			return true
		}
	}
	return false
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

// removeFileExtention removes the file extention from path; e.g. Foo/Bar.md becomes Foo/Bar
func removeFileExtension(path string) string {
	extension := filepath.Ext(path)
	return path[:len(path)-len(extension)]
}

// destIsOlder returns true if dest is older than source.
func destIsOlder(sourceInfo fs.FileInfo, destPath string) bool {
	destInfo, err := os.Stat(destPath)
	if err == nil && sourceInfo.ModTime().Before(destInfo.ModTime()) {
		return true
	}
	return false
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

// generateHtml generates an HTML file from a markdown file.
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

	// Make substituions.
	data = wiki.makeSubstitions(data)

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

// copyFile copies source to the file at destPath, overwriting destPath if it exists.
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

	// Copy file.
	util.PrintVerbose("Copying '%s'", sourceRelPath)
	var source *os.File
	var err error
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

// watch watches for changes in the wiki content directory and regenerates files on the fly.
func (wiki *Wiki) watch(clean bool, version string) error {
	util.PrintVerbose("Watching for changes in '%s'", wiki.ContentDir)

	var snapshot []fileSnapshot // Latest snapshot
	for {
		// Wait for when wiki needs to be updated.
		var err error
		if snapshot, err = wiki.waitForWhenGenerateNeeded(clean, version, snapshot); err != nil {
			return fmt.Errorf("failed waiting to update %s wiki: %v", wiki.SourceDir, err)
		}

		// Update wiki.
		if err = wiki.generate(false, clean, version); err != nil {
			return fmt.Errorf("failed to update %s wiki: %v", wiki.SourceDir, err)
		}
	}
}

func (wiki *Wiki) waitForWhenGenerateNeeded(clean bool, version string, snapshot []fileSnapshot) ([]fileSnapshot, error) {
	// Create timeout context so that a generate is done at least every MAX_REGEN_INTERVAL minutes.
	ctx := context.Background()
	ctx, cancelCtx := context.WithTimeout(ctx, MAX_REGEN_INTERVAL*time.Minute)
	defer cancelCtx()

	// Create channel to propagate errors from within goroutine.
	errorChan := make(chan error, 1)
	defer func() {
		close(errorChan)
	}()

	// Create a chan for the goroutine to say it's done.
	doneChan := make(chan struct{}, 1)
	defer func() {
		close(doneChan)
	}()

	// Create a chan to return snapshot.
	snapshotChan := make(chan []fileSnapshot, 1)
	defer func() {
		close(snapshotChan)
	}()

	go func() {
		// Say when this goroutine is done.
		defer func() {
			doneChan <- struct{}{}
		}()

		// Watch for changes.
		var err error
		if err = watchForChangeEvent(ctx, wiki.ContentDir, clean, version, snapshot); err != nil {
			errorChan <- fmt.Errorf("watch for change event in %s failed: %v", wiki.SourceDir, err)
			return
		}

		// Continue?
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Wait for changes to finish.
		var newSnapshot []fileSnapshot
		if newSnapshot, err = waitForChangesToFinish(ctx, wiki.ContentDir); err != nil {
			errorChan <- fmt.Errorf("wait for changes to finish for %s failed: %v", wiki.SourceDir, err)
			return
		}

		// We're exiting normally.
		snapshotChan <- newSnapshot
	}()

	// Wait on results.
	var err error
	select {
	case snapshot = <-snapshotChan:
	case err = <-errorChan:
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			util.PrintDebug("Regen timer expired for %s", wiki.SourceDir)
		}
	}

	// Wait for goroutine to finish, so that the chans it writes to aren't closed before any final writes.
	<-doneChan

	return snapshot, err
}

// watchDirRecursive sets up watches on the specified directory and all subdirectories recursively.
func watchDirRecursive(path string, watcher *fsnotify.Watcher) error {
	err := watcher.Add(path)
	if err != nil {
		return fmt.Errorf("failed to watch directory '%s': '%s'", path, err)
	}

	err = filepath.Walk(path, func(subPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			err = watcher.Add(subPath)
			if err != nil {
				return fmt.Errorf("failed to watch subdirectory '%s': '%s'", subPath, err)
			}
		}
		return nil
	})

	return err
}

// fileSnapshot records the name and modification time for a given file or directory.
type fileSnapshot struct {
	name      string
	timestamp int64
	isDir     bool
}

// TakeFilesSnapshot records the names and  modification times of all files and directories in dir recursively.
func takeFilesSnapshot(dir string) ([]fileSnapshot, error) {
	var snapshots []fileSnapshot

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		snapshot := fileSnapshot{
			name:      path,
			timestamp: info.ModTime().Unix(),
			isDir:     info.IsDir(),
		}
		snapshots = append(snapshots, snapshot)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return snapshots, nil
}

// FilesSnapshotsAreEqual compares two snapshots and returns true if they exist and are equal.
func filesSnapshotsAreEqual(snapshot1, snapshot2 []fileSnapshot) bool {
	if snapshot1 == nil || snapshot2 == nil {
		return false
	}
	return reflect.DeepEqual(snapshot1, snapshot2)
}

// waitForChangesToFinish waits for changes in dir to finish.
func waitForChangesToFinish(ctx context.Context, dir string) ([]fileSnapshot, error) {
	util.PrintDebug("Waiting for changes to finish in %s", dir)
	// Create channels.
	snapshotsMatchChan := make(chan []fileSnapshot, 1) // Signals that changes are complete.
	errorChan := make(chan error, 1)                   // Signals that an error happened while waiting.
	doneChan := make(chan struct{}, 1)                 // Signals that goroutine is done.
	defer func() {
		close(snapshotsMatchChan)
		close(errorChan)
		close(doneChan)
	}()

	// Wait for changes to complete.
	go func() {
		// Say when this goroutine is done.
		defer func() {
			doneChan <- struct{}{}
		}()

		// Initial wait.
		time.Sleep(CHANGE_WAIT * time.Millisecond)

		var snapshot1, snapshot2 []fileSnapshot
		var err error
		for waitPass := 1; !filesSnapshotsAreEqual(snapshot1, snapshot2); waitPass++ {
			select {
			case <-ctx.Done():
				// The context has ended and so end this goroutine too.
				return
			default:
				// Print wait status.
				message := fmt.Sprintf("Wait for change pass %d for %s", waitPass, dir)
				if waitPass > 1 {
					util.PrintVerbose(message)
				} else {
					util.PrintDebug(message)
				}

				// Take a before snapshot.
				if snapshot2 != nil {
					snapshot1 = snapshot2
				} else {
					snapshot1, err = takeFilesSnapshot(dir)
					if err != nil {
						errorChan <- fmt.Errorf("failed to take files snapshot for %s: %v", dir, err)
					}
				}

				// Wait
				waitTime := waitPass * waitPass * CHANGE_WAIT
				if waitTime > MAX_CHANGE_WAIT {
					waitTime = MAX_CHANGE_WAIT
				}
				util.PrintDebug("Waiting %d ms for %s", waitTime, dir)
				time.Sleep(time.Duration(waitTime) * time.Millisecond)

				// Take an after snapshot.
				if snapshot2, err = takeFilesSnapshot(dir); err != nil {
					errorChan <- fmt.Errorf("failed to take files snapshot for %s: %v", dir, err)
				}
			}
		}

		// Snapshots match. Remember last snapshot.
		snapshotsMatchChan <- snapshot2
	}()

	// Wait for results.
	var snapshot []fileSnapshot
	var err error
	select {
	case snapshot = <-snapshotsMatchChan:
		util.PrintDebug("Snapshots match for %s", dir)
		break
	case err = <-errorChan:
		break
	case <-ctx.Done():
		break
	}

	// Wait for goroutine to finish, so that the chans it writes to aren't closed before any final writes.
	<-doneChan

	return snapshot, err
}

// watchForChangeEvent watches for a change to the wiki content.
func watchForChangeEvent(ctx context.Context, contentDir string, clean bool, version string, snapshot []fileSnapshot) error {
	// Create and initialize watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher for '%s': %v", contentDir, err)
	}
	defer watcher.Close()
	if err = watchDirRecursive(contentDir, watcher); err != nil {
		return fmt.Errorf("failed to initialize watcher for %s: %v", contentDir, err)
	}

	// Make sure files haven't changed in between when wiki update started and new watch started.
	if snapshot != nil {
		newSnapshot, err := takeFilesSnapshot(contentDir)
		if err != nil {
			return fmt.Errorf("failed to take new snapshot for %s: %v", contentDir, err)
		}
		if !filesSnapshotsAreEqual(snapshot, newSnapshot) {
			// Files have changed. Start a new update.
			util.PrintVerbose("About to watch for changes but file snapshots differ. Starting a new update.")
			return nil
		}
	}

	// Watch for changes.
	select {
	case event, ok := <-watcher.Events:
		if !ok {
			return fmt.Errorf("watcher unexpectedly closed while watching '%s' %v", contentDir, err)
		}
		util.PrintDebug("Watcher event detected for %s: %v", contentDir, event)
		return nil
	case err, ok := <-watcher.Errors:
		if !ok {
			return fmt.Errorf("failed to read watcher error for %s", contentDir)
		}
		return fmt.Errorf("watcher error for %s: %v", contentDir, err)
	case <-ctx.Done():
		// Regen timer has expired.
		return nil
	}
}
