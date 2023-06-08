// Package wiki generates HTML from markdown for a given wiki.
package wiki

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
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
var htmlHeaderTemplate *template.Template

//go:embed static/style.css
var embeddedFileSystem embed.FS

const CHANGE_WAIT = 100 // milliseconds
const MAX_WAIT = 5000   // milliseconds

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

// Wiki stores data about a wiki to generate
type Wiki struct {
	// Directories
	SourceDir  string // Wiki source directory
	ContentDir string // Content directory within wiki source
	DestDir    string // Dest dir where generated wiki will be stored

	styleCssCopyNeeded bool // Whether style.css nees to be copied to dest

	subStrings [][2]string // Substitution strings. Each pair is the string to look for and what to replace it with.
}

// NewWikiDirs constructs a new instance of WikiDirs.
func NewWiki(sourceDir, destDir string) (*Wiki, error) {
	wiki := Wiki{
		SourceDir:          sourceDir,
		ContentDir:         filepath.Join(sourceDir, "content"),
		DestDir:            destDir,
		styleCssCopyNeeded: true,
		subStrings:         nil,
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

	return &wiki, nil
}

// loadSubstitionStrings loads substition strings for a wiki, from substition-strings.csv
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
		// There are no substition strings.
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

// makeSubstitions makes substitions in data.
func (wiki Wiki) makeSubstitions(data []byte) []byte {
	for _, pair := range wiki.subStrings {
		data = bytes.ReplaceAll(data, []byte(pair[0]), []byte(pair[1]))
	}
	return data
}

// Generate generates a wiki and then optionally watches for changes in the
// wiki to regenerate files on the fly.
func (wiki *Wiki) Generate(regen, clean, watch bool, version string) error {
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
	// Generate the part of the wiki that comes from the source content.
	var relDestPaths map[string]bool
	var err error
	if relDestPaths, err = wiki.generateFromContent(regen, version); err != nil {
		return err
	}

	// Copy styles.css to destDir.
	if err = wiki.copyStyleCss(); err != nil {
		return err
	}
	relDestPaths["style.css"] = true

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

// generateFromContent generates the part of the wiki that comes from the source content.
func (wiki Wiki) generateFromContent(regen bool, version string) (map[string]bool, error) {
	// Iterate recursively over the source directory and generate the wiki from the files found.
	util.PrintVerbose("Generating wiki '%s' from '%s'", wiki.DestDir, wiki.SourceDir)
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
	if err = htmlHeaderTemplate.Execute(html, templateData{title, version, rootRelPath}); err != nil {
		return "", fmt.Errorf("failed to create HTML header for '%s': %v", outPath, err)
	}

	// Generate the body of the HTML from markdown.
	if err = markdown.Convert(data, html); err != nil {
		return "", fmt.Errorf("failed to generate HTML body for '%s': %v", outPath, err)
	}

	// Generate end of HTML file.
	html.WriteString("</body>\n</html>")

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

// copyFile copies source to the file at destPath.
func copyToFile(destPath string, source io.Reader) error {
	// Create dest file.
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

// cleanDestDir cleans the dest dir by any deleting files that don't have any
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

// copyStyleCss copies styles.css to dest dir.
func (wiki *Wiki) copyStyleCss() error {
	// Is copy neeeded?
	if !wiki.styleCssCopyNeeded {
		return nil
	}

	// Copy style.css.
	var styleCss []byte
	var err error
	util.PrintVerbose("Copying style.css to '%s'", wiki.DestDir)
	if styleCss, err = embeddedFileSystem.ReadFile("static/style.css"); err != nil {
		return fmt.Errorf("failed to read style.css: %v", err)
	}
	styleCssPath := fmt.Sprintf("%s/style.css", wiki.DestDir)
	if err := copyToFile(styleCssPath, bytes.NewReader(styleCss)); err != nil {
		return err
	}
	wiki.styleCssCopyNeeded = false

	return nil
}

// watch watches for changes in the wiki content directory and regenerates files on the fly.
func (wiki *Wiki) watch(clean bool, version string) error {
	util.PrintVerbose("Watching for changes in '%s'", wiki.ContentDir)

	var snapshot []fileSnapshot // Latest snapshot
	for phaseId := 1; ; phaseId++ {
		// Watch for changes.
		var err error
		if err = watchForChangeEvent(phaseId, wiki.ContentDir, clean, version, snapshot); err != nil {
			return fmt.Errorf("watch phase %d failed: %v", phaseId, err)
		}

		// Wait for changes to finish.
		if snapshot, err = waitForChangesToFinish(wiki.ContentDir); err != nil {
			return fmt.Errorf("wait for changes phase %d failed: %v", phaseId, err)
		}

		// Update wiki.
		if err := wiki.generate(false, clean, version); err != nil {
			return fmt.Errorf("failed to update wiki: %v", err)
		}
	}
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
func waitForChangesToFinish(dir string) ([]fileSnapshot, error) {
	// Initial wait.
	time.Sleep(CHANGE_WAIT * time.Millisecond)

	// Create channels.
	snapshotsMatchChan := make(chan []fileSnapshot, 1) // Signals that changes are complete.
	errorChan := make(chan error, 1)                   // Signals that an error happened while waiting.
	defer func() {
		close(snapshotsMatchChan)
		close(errorChan)
	}()

	// Wait for changes to complete.
	go func() {
		var snapshot1, snapshot2 []fileSnapshot
		var err error
		for waitPass := 1; !filesSnapshotsAreEqual(snapshot1, snapshot2); waitPass++ {
			util.PrintDebug("Wait pass %d", waitPass)

			// Take a before snapshot.
			if snapshot2 != nil {
				snapshot1 = snapshot2
			} else {
				snapshot1, err = takeFilesSnapshot(dir)
				if err != nil {
					errorChan <- fmt.Errorf("failed to take files snapshot: %v", err)
				}
			}

			// Wait
			waitTime := waitPass * waitPass * CHANGE_WAIT
			if waitTime > MAX_WAIT {
				waitTime = MAX_WAIT
			}
			time.Sleep(time.Duration(waitTime) * time.Millisecond)

			// Take an after snapshot.
			if snapshot2, err = takeFilesSnapshot(dir); err != nil {
				errorChan <- fmt.Errorf("failed to take files snapshot: %v", err)
			}
		}

		// Snapshots match. Remember last snapshot.
		snapshotsMatchChan <- snapshot2
	}()

	// Wait on changes to complete, or exit if there's a ctrl-c.
	for {
		select {
		case snapshot := <-snapshotsMatchChan:
			util.PrintDebug("Snapshots match")
			return snapshot, nil
		case err := <-errorChan:
			return nil, err
		}
	}
}

// watchForChangeEvent watches for a change to the wiki content.
func watchForChangeEvent(phaseId int, contentDir string, clean bool, version string, snapshot []fileSnapshot) error {
	// Create and initialize watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher for '%s': %v", contentDir, err)
	}
	defer watcher.Close()
	if err = watchDirRecursive(contentDir, watcher); err != nil {
		return fmt.Errorf("failed to initialize watcher: %v", err)
	}

	// Make sure files haven't changed in between when wiki update started and new watch started.
	if snapshot != nil {
		newSnapshot, err := takeFilesSnapshot(contentDir)
		if err != nil {
			return fmt.Errorf("failed to take new snapshot: %v", err)
		}
		if !filesSnapshotsAreEqual(snapshot, newSnapshot) {
			// Start a new update.
			return nil
		}
	}

	// Watch for changes.
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return fmt.Errorf("watcher unexpectedly closed while watching '%s' %v", contentDir, err)
			}
			util.PrintDebug("Watcher event detected: %v", event)
			return nil
		case err, ok := <-watcher.Errors:
			if !ok {
				return errors.New("failed to read watcher error")
			}
			return fmt.Errorf("watcher error: %v", err)
		}
	}
}
