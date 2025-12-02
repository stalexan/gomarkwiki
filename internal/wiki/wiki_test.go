package wiki

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const PACKAGE_DIR = "../.."
const TMP_DIR = "tmp"
const TESTDATA_DIR = "./testdata"
const STYLE_PATH = "./static/style.css"
const GITHUB_STYLE_PATH = "./static/github-style.css"

var (
	packageDir      string
	stylePath       string
	githubStylePath string
	tempDir         string
	testDataDir     string
)

func messageFatal(message string) {
	fmt.Println(message)
	os.Exit(1)
}

func TestMain(m *testing.M) {
	// Create temp dir.
	var err error
	if packageDir, err = filepath.Abs(PACKAGE_DIR); err != nil {
		messageFatal(fmt.Sprintf("Error creating absolute path: %v", err))
	}
	tempDir = filepath.Join(packageDir, TMP_DIR)
	if err = os.MkdirAll(tempDir, 0755); err != nil {
		messageFatal(fmt.Sprintf("Error creating temp directory: %v", err))
	}

	// Find testdata dir.
	testDataDir = TESTDATA_DIR

	// Find style.css.
	stylePath = STYLE_PATH
	githubStylePath = GITHUB_STYLE_PATH

	// Run the tests.
	code := m.Run()

	os.Exit(code)
}

// TestA01GenerateTinyWiki tests generation of a simple wiki.
func TestA01GenerateTinyWiki(t *testing.T) {
	// Create temp dir.
	var testCaseTempDir string
	var err error
	var success bool
	if testCaseTempDir, err = os.MkdirTemp(tempDir, "a01"); err != nil {
		t.Fatalf("Error creating test case temp directory: %v", err)
	}
	defer func() {
		if success {
			t.Logf("Removing test case temp directory %s", testCaseTempDir)
			if err := os.RemoveAll(testCaseTempDir); err != nil {
				t.Fatalf("Error removing test case temp directory: %v", err)
			}
		}
	}()
	t.Logf("Created test case temp directory %s", testCaseTempDir)

	// Create output dir.
	outputDir := filepath.Join(testCaseTempDir, "output")
	if err = os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create directory %s: %v", outputDir, err)
	}

	// Create Wiki instance
	testCaseDataDir := filepath.Join(testDataDir, "a01-tiny-wiki")
	sourceDir := filepath.Join(testCaseDataDir, "source")
	var theWiki *Wiki
	if theWiki, err = NewWiki(sourceDir, outputDir); err != nil {
		t.Fatalf("Error creating Wiki instance: %v", err)
	}

	// Generate wiki
	if err = theWiki.Generate(context.Background(), true, false, false, "test"); err != nil {
		t.Fatalf("Error generating wiki: %v", err)
	}

	// Create expected output.
	expectedOutputSourceDir := filepath.Join(testCaseDataDir, "expected-output")
	expectedOutputDir := filepath.Join(testCaseTempDir, "expected-output")
	if err := copyDir(expectedOutputSourceDir, expectedOutputDir); err != nil {
		t.Fatalf("Failed to create expected output dir %s: %v", expectedOutputDir, err)
	}
	if err := copyFile(stylePath, expectedOutputDir); err != nil {
		t.Fatalf("Failed to create style.css for expected output dir %s: %v", expectedOutputDir, err)
	}
	if err := copyFile(githubStylePath, expectedOutputDir); err != nil {
		t.Fatalf("Failed to create github-style.css for expected output dir %s: %v", expectedOutputDir, err)
	}

	// Check output.
	var report string
	if report, err = diffDirs(expectedOutputDir, outputDir); err != nil {
		t.Fatalf("Error diffing dirs: %v", err)
	}
	if len(report) > 0 {
		t.Logf("Directories differ: expected output %s and actual output %s", sourceDir, testCaseTempDir)
		t.Fatalf("Diff report:\n%s", strings.TrimRight(report, "\n"))
	}

	success = true
}

// TestCollisionDetectionDeterministic tests that collision detection is deterministic
// across multiple generation cycles. When multiple markdown files would generate the
// same HTML path, the same file should always win (based on lexicographic ordering).
func TestCollisionDetectionDeterministic(t *testing.T) {
	// Create temp dir.
	var testCaseTempDir string
	var err error
	var success bool
	if testCaseTempDir, err = os.MkdirTemp(tempDir, "collision-test"); err != nil {
		t.Fatalf("Error creating test case temp directory: %v", err)
	}
	defer func() {
		if success {
			t.Logf("Removing test case temp directory %s", testCaseTempDir)
			if err := os.RemoveAll(testCaseTempDir); err != nil {
				t.Fatalf("Error removing test case temp directory: %v", err)
			}
		}
	}()
	t.Logf("Created test case temp directory %s", testCaseTempDir)

	// Create source directory structure
	sourceDir := filepath.Join(testCaseTempDir, "source")
	contentDir := filepath.Join(sourceDir, "content")
	if err = os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatalf("Failed to create content directory %s: %v", contentDir, err)
	}

	// Create multiple markdown files that would all generate "test.html"
	// Using different extensions to ensure they collide
	files := []struct {
		name     string
		content  string
		expected string // Which file should win (lexicographically first)
	}{
		{"test.markdown", "# Test Markdown", "test.markdown"}, // "markdown" < "md" < "mdwn" lexicographically
		{"test.md", "# Test MD", "test.markdown"},             // Should be skipped
		{"test.mdwn", "# Test MDWN", "test.markdown"},         // Should be skipped
	}

	for _, f := range files {
		filePath := filepath.Join(contentDir, f.name)
		if err = os.WriteFile(filePath, []byte(f.content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", filePath, err)
		}
	}

	// Create static directory (required by NewWiki)
	staticDir := filepath.Join(sourceDir, "static")
	if err = os.MkdirAll(staticDir, 0755); err != nil {
		t.Fatalf("Failed to create static directory %s: %v", staticDir, err)
	}

	// Create output dir
	outputDir := filepath.Join(testCaseTempDir, "output")
	if err = os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create directory %s: %v", outputDir, err)
	}

	// Create Wiki instance
	var theWiki *Wiki
	if theWiki, err = NewWiki(sourceDir, outputDir); err != nil {
		t.Fatalf("Error creating Wiki instance: %v", err)
	}

	// Run generation multiple times to verify deterministic behavior
	expectedWinner := "test.markdown"
	for i := 0; i < 3; i++ {
		// Clean output directory between runs
		if err = os.RemoveAll(outputDir); err != nil {
			t.Fatalf("Failed to clean output directory: %v", err)
		}
		if err = os.MkdirAll(outputDir, 0755); err != nil {
			t.Fatalf("Failed to recreate output directory: %v", err)
		}

		// Generate wiki
		if err = theWiki.Generate(context.Background(), true, false, false, "test"); err != nil {
			t.Fatalf("Error generating wiki (run %d): %v", i+1, err)
		}

		// Verify that only one HTML file was created
		expectedHtmlPath := filepath.Join(outputDir, "test.html")
		if _, err = os.Stat(expectedHtmlPath); err != nil {
			t.Fatalf("Expected HTML file %s not found (run %d): %v", expectedHtmlPath, i+1, err)
		}

		// Verify that the content matches the expected winner (test.markdown)
		htmlContent, err := os.ReadFile(expectedHtmlPath)
		if err != nil {
			t.Fatalf("Failed to read HTML file: %v", err)
		}

		// The HTML should contain content from test.markdown (the lexicographically first file)
		if !strings.Contains(string(htmlContent), "Test Markdown") {
			t.Fatalf("HTML file does not contain expected content from %s (run %d). Content: %s", expectedWinner, i+1, string(htmlContent)[:min(200, len(htmlContent))])
		}

		// Verify that only one HTML file exists (no duplicates)
		htmlFiles := 0
		err = filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && filepath.Ext(path) == ".html" {
				htmlFiles++
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Error walking output directory: %v", err)
		}
		if htmlFiles != 1 {
			t.Fatalf("Expected exactly 1 HTML file, found %d (run %d)", htmlFiles, i+1)
		}
	}

	success = true
}

// TestCollisionDetectionNestedDirs tests that collision detection works correctly
// with nested directory structures. It verifies:
// 1. Files in the same nested directory collide (e.g., dir/test.md and dir/test.markdown)
// 2. Files with the same name in different directories don't collide (e.g., a/test.md and b/test.md)
// 3. Collision resolution is deterministic based on lexicographic ordering by full path
func TestCollisionDetectionNestedDirs(t *testing.T) {
	// Create temp dir.
	var testCaseTempDir string
	var err error
	var success bool
	if testCaseTempDir, err = os.MkdirTemp(tempDir, "collision-nested-test"); err != nil {
		t.Fatalf("Error creating test case temp directory: %v", err)
	}
	defer func() {
		if success {
			t.Logf("Removing test case temp directory %s", testCaseTempDir)
			if err := os.RemoveAll(testCaseTempDir); err != nil {
				t.Fatalf("Error removing test case temp directory: %v", err)
			}
		}
	}()
	t.Logf("Created test case temp directory %s", testCaseTempDir)

	// Create source directory structure
	sourceDir := filepath.Join(testCaseTempDir, "source")
	contentDir := filepath.Join(sourceDir, "content")
	if err = os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatalf("Failed to create content directory %s: %v", contentDir, err)
	}

	// Create nested directories
	dirA := filepath.Join(contentDir, "dirA")
	dirB := filepath.Join(contentDir, "dirB")
	dirC := filepath.Join(contentDir, "dirC")
	if err = os.MkdirAll(dirA, 0755); err != nil {
		t.Fatalf("Failed to create directory %s: %v", dirA, err)
	}
	if err = os.MkdirAll(dirB, 0755); err != nil {
		t.Fatalf("Failed to create directory %s: %v", dirB, err)
	}
	if err = os.MkdirAll(dirC, 0755); err != nil {
		t.Fatalf("Failed to create directory %s: %v", dirC, err)
	}

	// Create test files:
	// 1. In dirA: collision between test.markdown and test.md (test.markdown should win)
	// 2. In dirB: single file page.md (no collision)
	// 3. In dirC: collision between page.md and page.mdwn (page.md should win)
	// 4. Files named "common.md" in both dirA and dirB (no collision - different output paths)
	files := []struct {
		path             string
		content          string
		shouldBeInOutput bool   // Whether this file should generate output
		outputPath       string // Expected output path relative to output dir
	}{
		// dirA files - test.markdown should win the collision
		{"dirA/test.markdown", "# DirA Test Markdown", true, "dirA/test.html"},
		{"dirA/test.md", "# DirA Test MD", false, ""},

		// Both directories have common.md - no collision (different output paths)
		{"dirA/common.md", "# DirA Common", true, "dirA/common.html"},
		{"dirB/common.md", "# DirB Common", true, "dirB/common.html"},

		// dirB single file
		{"dirB/page.md", "# DirB Page", true, "dirB/page.html"},

		// dirC files - page.md should win the collision (lexicographically before page.mdwn)
		{"dirC/page.md", "# DirC Page MD", true, "dirC/page.html"},
		{"dirC/page.mdwn", "# DirC Page MDWN", false, ""},
	}

	for _, f := range files {
		filePath := filepath.Join(contentDir, f.path)
		if err = os.WriteFile(filePath, []byte(f.content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", filePath, err)
		}
	}

	// Create static directory (required by NewWiki)
	staticDir := filepath.Join(sourceDir, "static")
	if err = os.MkdirAll(staticDir, 0755); err != nil {
		t.Fatalf("Failed to create static directory %s: %v", staticDir, err)
	}

	// Create output dir
	outputDir := filepath.Join(testCaseTempDir, "output")
	if err = os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create directory %s: %v", outputDir, err)
	}

	// Create Wiki instance
	var theWiki *Wiki
	if theWiki, err = NewWiki(sourceDir, outputDir); err != nil {
		t.Fatalf("Error creating Wiki instance: %v", err)
	}

	// Run generation multiple times to verify deterministic behavior
	for i := 0; i < 3; i++ {
		// Clean output directory between runs
		if err = os.RemoveAll(outputDir); err != nil {
			t.Fatalf("Failed to clean output directory: %v", err)
		}
		if err = os.MkdirAll(outputDir, 0755); err != nil {
			t.Fatalf("Failed to recreate output directory: %v", err)
		}

		// Generate wiki
		if err = theWiki.Generate(context.Background(), true, false, false, "test"); err != nil {
			t.Fatalf("Error generating wiki (run %d): %v", i+1, err)
		}

		// Verify expected outputs
		for _, f := range files {
			if f.shouldBeInOutput {
				outputPath := filepath.Join(outputDir, f.outputPath)

				// Check file exists
				if _, err = os.Stat(outputPath); err != nil {
					t.Fatalf("Expected output file %s not found (run %d): %v", outputPath, i+1, err)
				}

				// Check content matches expected source
				htmlContent, err := os.ReadFile(outputPath)
				if err != nil {
					t.Fatalf("Failed to read output file %s: %v", outputPath, err)
				}

				// Extract the first line of the markdown content for verification
				expectedContent := f.content[2:] // Remove "# " prefix
				if !strings.Contains(string(htmlContent), expectedContent) {
					t.Fatalf("Output file %s does not contain expected content '%s' (run %d). Content: %s",
						outputPath, expectedContent, i+1, string(htmlContent)[:min(200, len(htmlContent))])
				}
			}
		}

		// Count total HTML files to ensure we have exactly the expected number
		htmlFiles := 0
		err = filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && filepath.Ext(path) == ".html" {
				htmlFiles++
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Error walking output directory: %v", err)
		}

		// Should have exactly 5 HTML files:
		// dirA/test.html, dirA/common.html, dirB/common.html, dirB/page.html, dirC/page.html
		expectedCount := 5
		if htmlFiles != expectedCount {
			t.Fatalf("Expected exactly %d HTML files, found %d (run %d)", expectedCount, htmlFiles, i+1)
		}
	}

	success = true
}

// TestIgnoreFileSkipsBlankAndCommentLines ensures ignore.txt parsing skips blank/comment lines.
func TestIgnoreFileSkipsBlankAndCommentLines(t *testing.T) {
	var err error

	// Create temp dir for test
	testCaseTempDir, err := os.MkdirTemp(tempDir, "ignore-test")
	if err != nil {
		t.Fatalf("Error creating test case temp directory: %v", err)
	}
	defer func() {
		t.Logf("Removing test case temp directory %s", testCaseTempDir)
		if err := os.RemoveAll(testCaseTempDir); err != nil {
			t.Fatalf("Error removing test case temp directory: %v", err)
		}
	}()

	sourceDir := filepath.Join(testCaseTempDir, "source")
	contentDir := filepath.Join(sourceDir, "content")
	staticDir := filepath.Join(sourceDir, "static")
	outputDir := filepath.Join(testCaseTempDir, "output")

	for _, dir := range []string{contentDir, staticDir, outputDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Required static files for generation
	if err := copyFile(stylePath, staticDir); err != nil {
		t.Fatalf("Failed to copy style.css: %v", err)
	}
	if err := copyFile(githubStylePath, staticDir); err != nil {
		t.Fatalf("Failed to copy github-style.css: %v", err)
	}

	// Create content files
	keepMarkdown := filepath.Join(contentDir, "keep.md")
	if err := os.WriteFile(keepMarkdown, []byte("# Keep\nThis file should be processed."), 0644); err != nil {
		t.Fatalf("Failed to create markdown file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(contentDir, "skip.tmp"), []byte("tmp"), 0644); err != nil {
		t.Fatalf("Failed to create tmp file: %v", err)
	}

	ignoreContents := "# Comment line should be ignored\n" +
		"\n" +
		"   # Another comment with leading space\n" +
		"  \t  \n" +
		"*.tmp\n"
	if err := os.WriteFile(filepath.Join(sourceDir, "ignore.txt"), []byte(ignoreContents), 0644); err != nil {
		t.Fatalf("Failed to write ignore.txt: %v", err)
	}

	theWiki, err := NewWiki(sourceDir, outputDir)
	if err != nil {
		t.Fatalf("Error creating Wiki instance: %v", err)
	}

	if err := theWiki.Generate(context.Background(), true, false, false, "test"); err != nil {
		t.Fatalf("Error generating wiki: %v", err)
	}

	expectedHtml := filepath.Join(outputDir, "keep.html")
	if _, err := os.Stat(expectedHtml); err != nil {
		t.Fatalf("Expected HTML file %s not found: %v", expectedHtml, err)
	}

	if _, err := os.Stat(filepath.Join(outputDir, "skip.tmp")); err == nil {
		t.Fatalf("skip.tmp should have been ignored and not copied to output")
	}
}

// TestConfigPathsNotSetWhenFilesMissing ensures Wiki does not set subsPath when file is absent.
// Note: ignorePath is still set even when file is missing (different behavior from subsPath).
func TestConfigPathsNotSetWhenFilesMissing(t *testing.T) {
	var err error

	testCaseTempDir, err := os.MkdirTemp(tempDir, "config-paths")
	if err != nil {
		t.Fatalf("Error creating test case temp directory: %v", err)
	}
	defer func() {
		t.Logf("Removing test case temp directory %s", testCaseTempDir)
		if err := os.RemoveAll(testCaseTempDir); err != nil {
			t.Fatalf("Error removing test case temp directory: %v", err)
		}
	}()

	sourceDir := filepath.Join(testCaseTempDir, "source")
	contentDir := filepath.Join(sourceDir, "content")
	staticDir := filepath.Join(sourceDir, "static")
	outputDir := filepath.Join(testCaseTempDir, "output")

	for _, dir := range []string{contentDir, staticDir, outputDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Required static files for generation
	if err := copyFile(stylePath, staticDir); err != nil {
		t.Fatalf("Failed to copy style.css: %v", err)
	}
	if err := copyFile(githubStylePath, staticDir); err != nil {
		t.Fatalf("Failed to copy github-style.css: %v", err)
	}

	theWiki, err := NewWiki(sourceDir, outputDir)
	if err != nil {
		t.Fatalf("Error creating Wiki instance: %v", err)
	}

	expectedIgnore := filepath.Join(sourceDir, "ignore.txt")
	expectedSubs := filepath.Join(sourceDir, "substitution-strings.csv")

	// subsPath is set even when file is missing (so watcher can detect creation)
	if theWiki.subsPath != expectedSubs {
		t.Fatalf("Expected subsPath to be %s, got %s", expectedSubs, theWiki.subsPath)
	}

	// ignorePath is set even when file is missing (so watcher can detect creation)
	if theWiki.ignorePath != expectedIgnore {
		t.Fatalf("Expected ignorePath to be %s, got %s", expectedIgnore, theWiki.ignorePath)
	}
}

// TestConfigPathsSetWhenFilesExist ensures Wiki sets subsPath when substitution file exists.
func TestConfigPathsSetWhenFilesExist(t *testing.T) {
	var err error

	testCaseTempDir, err := os.MkdirTemp(tempDir, "config-paths-exist")
	if err != nil {
		t.Fatalf("Error creating test case temp directory: %v", err)
	}
	defer func() {
		t.Logf("Removing test case temp directory %s", testCaseTempDir)
		if err := os.RemoveAll(testCaseTempDir); err != nil {
			t.Fatalf("Error removing test case temp directory: %v", err)
		}
	}()

	sourceDir := filepath.Join(testCaseTempDir, "source")
	contentDir := filepath.Join(sourceDir, "content")
	staticDir := filepath.Join(sourceDir, "static")
	outputDir := filepath.Join(testCaseTempDir, "output")

	for _, dir := range []string{contentDir, staticDir, outputDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Required static files for generation
	if err := copyFile(stylePath, staticDir); err != nil {
		t.Fatalf("Failed to copy style.css: %v", err)
	}
	if err := copyFile(githubStylePath, staticDir); err != nil {
		t.Fatalf("Failed to copy github-style.css: %v", err)
	}

	// Create substitution strings file
	subsPath := filepath.Join(sourceDir, "substitution-strings.csv")
	if err := os.WriteFile(subsPath, []byte("TEST,value"), 0644); err != nil {
		t.Fatalf("Failed to write substitution file: %v", err)
	}

	theWiki, err := NewWiki(sourceDir, outputDir)
	if err != nil {
		t.Fatalf("Error creating Wiki instance: %v", err)
	}

	expectedSubs := filepath.Join(sourceDir, "substitution-strings.csv")

	// subsPath should be set when file exists
	if theWiki.subsPath != expectedSubs {
		t.Fatalf("Expected subsPath to be %s when file exists, got %s", expectedSubs, theWiki.subsPath)
	}
}

// TestCheckSubsFileChangedDetectsCreateDelete verifies creation and deletion are detected.
func TestCheckSubsFileChangedDetectsCreateDelete(t *testing.T) {
	tempDir := t.TempDir()
	subsPath := filepath.Join(tempDir, "substitution-strings.csv")

	w := &Watcher{
		subsPath: subsPath,
	}

	// Create file
	if err := os.WriteFile(subsPath, []byte("A,B"), 0644); err != nil {
		t.Fatalf("Failed to write substitution file: %v", err)
	}
	if changed := w.checkSubsFileChanged(); !changed {
		t.Fatalf("Expected creation to be detected as change")
	}

	// No change
	if changed := w.checkSubsFileChanged(); changed {
		t.Fatalf("Expected no change when file unchanged")
	}

	// Delete file
	if err := os.Remove(subsPath); err != nil {
		t.Fatalf("Failed to remove substitution file: %v", err)
	}
	if changed := w.checkSubsFileChanged(); !changed {
		t.Fatalf("Expected deletion to be detected as change")
	}

	if w.subsFileExists {
		t.Fatalf("Expected subsFileExists to be false after deletion")
	}
}

// TestCheckIgnoreFileChangedDetectsCreateDelete verifies creation and deletion for ignore.txt.
func TestCheckIgnoreFileChangedDetectsCreateDelete(t *testing.T) {
	tempDir := t.TempDir()
	ignorePath := filepath.Join(tempDir, "ignore.txt")

	w := &Watcher{
		ignorePath: ignorePath,
	}

	// Create file
	if err := os.WriteFile(ignorePath, []byte("\\.tmp$"), 0644); err != nil {
		t.Fatalf("Failed to write ignore file: %v", err)
	}
	if changed := w.checkIgnoreFileChanged(); !changed {
		t.Fatalf("Expected creation to be detected as change")
	}

	// No change
	if changed := w.checkIgnoreFileChanged(); changed {
		t.Fatalf("Expected no change when file unchanged")
	}

	// Delete file
	if err := os.Remove(ignorePath); err != nil {
		t.Fatalf("Failed to remove ignore file: %v", err)
	}
	if changed := w.checkIgnoreFileChanged(); !changed {
		t.Fatalf("Expected deletion to be detected as change")
	}

	if w.ignoreFileExists {
		t.Fatalf("Expected ignoreFileExists to be false after deletion")
	}
}

func copyDir(dir1, dir2 string) error {
	// Make sure dir2 doesn't exit.
	_, err := os.Stat(dir2)
	if err == nil || !os.IsNotExist(err) {
		return fmt.Errorf("directory '%s' exists", dir2)
	}

	// Copy dir1 to dir2.
	command := exec.Command("cp", "-r", dir1, dir2)
	if _, err = command.Output(); err != nil {
		return fmt.Errorf("failed to copy dir '%s' to '%s'", dir1, dir2)
	}

	return nil
}

func copyFile(sourcePath, destPath string) error {
	command := exec.Command("cp", sourcePath, destPath)
	if _, err := command.Output(); err != nil {
		return fmt.Errorf("failed to copy file '%s' to '%s'", sourcePath, destPath)
	}
	return nil
}

func diffDirs(dir1, dir2 string) (string, error) {
	// Do diff.
	command := exec.Command("diff", "-r", dir1, dir2)
	var output []byte
	var err error
	if output, err = command.Output(); err != nil {
		if len(output) > 0 {
			// Return diff report.
			return string(output), nil
		} else {
			// there was an unexpected error
			return "", err
		}
	}

	// The directories are the same.
	return "", nil
}

// ============================================================================
// Config Tests (config.go)
// ============================================================================

func TestValidatePlaceholder(t *testing.T) {
	tests := []struct {
		name        string
		placeholder string
		wantErr     bool
	}{
		{"valid simple", "FOO", false},
		{"valid with underscore", "FOO_BAR", false},
		{"valid with hyphen", "foo-bar", false},
		{"valid with numbers", "foo123", false},
		{"empty", "", true},
		{"too long", strings.Repeat("a", 101), true},
		{"no alphanumeric", "___", true},
		{"contains braces", "{{FOO}}", true},
		{"contains space", "FOO BAR", true},
		{"contains special char", "FOO@BAR", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePlaceholder(tt.placeholder)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePlaceholder(%q) error = %v, wantErr %v", tt.placeholder, err, tt.wantErr)
			}
		})
	}
}

func TestMakeSubstitutions(t *testing.T) {
	wiki := Wiki{
		subStrings: [][2]string{
			{"{{SITE}}", "example.com"},
			{"{{YEAR}}", "2024"},
		},
	}

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"single substitution", "Visit {{SITE}}", "Visit example.com"},
		{"multiple substitutions", "{{SITE}} - {{YEAR}}", "example.com - 2024"},
		{"no substitutions", "plain text", "plain text"},
		{"repeated placeholder", "{{SITE}} and {{SITE}}", "example.com and example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wiki.makeSubstitutions([]byte(tt.input))
			if string(got) != tt.want {
				t.Errorf("makeSubstitutions() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIgnoreFile(t *testing.T) {
	// Create a temp directory to use as ContentDir
	tmpDir, err := os.MkdirTemp(tempDir, "ignore-test")
	if err != nil {
		t.Fatalf("Error creating test temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create ignore matcher with gitignore-style patterns
	patterns := []string{
		"*.tmp",          // Match all .tmp files
		".git/",          // Match .git directory
		"backup/",        // Match backup directory
		"*.log",          // Match all .log files
		"/README.md",     // Match README.md in root only
		"**/*.bak",       // Match .bak files at any depth
		"!important.log", // Don't ignore this specific file
	}

	matcher, err := NewIgnoreMatcher(patterns)
	if err != nil {
		t.Fatalf("Error creating ignore matcher: %v", err)
	}

	wiki := Wiki{
		ContentDir:    tmpDir,
		ignoreMatcher: matcher,
	}

	tests := []struct {
		path   string
		isDir  bool
		ignore bool
		reason string
	}{
		// Basic file patterns
		{"file.tmp", false, true, "*.tmp matches"},
		{"file.md", false, false, "no pattern matches"},
		{"data.log", false, true, "*.log matches"},
		{"important.log", false, false, "*.log matches but negated by !important.log"},

		// Directory patterns
		{".git", true, true, ".git/ matches directory"},
		{".git/config", false, true, ".git/ matches parent directory"},
		{".git", false, false, ".git/ should not match file named '.git'"},
		{"backup", true, true, "backup/ matches directory"},
		{"backup/file.md", false, true, "backup/ matches parent directory"},
		{"backup", false, false, "backup/ should not match file named 'backup'"},
		{"mybackup", true, false, "backup/ doesn't match without slash"},
		{"my-backup.md", false, false, "backup/ doesn't match substring in filename"},

		// Anchored patterns
		{"README.md", false, true, "/README.md matches at root"},
		{"docs/README.md", false, false, "/README.md doesn't match in subdir"},

		// Recursive patterns
		{"file.bak", false, true, "**/*.bak matches at root level"},
		{"dir/file.bak", false, true, "**/*.bak matches in subdir"},
		{"dir/subdir/file.bak", false, true, "**/*.bak matches in deep subdir"},

		// No match
		{"normal.md", false, false, "no pattern matches"},
		{"test.txt", false, false, "no pattern matches"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			// Create full path relative to ContentDir
			fullPath := filepath.Join(tmpDir, tt.path)

			got := wiki.ignoreFile(fullPath, tt.isDir)
			if got != tt.ignore {
				t.Errorf("ignoreFile(%q, isDir=%v) = %v, want %v (%s)",
					tt.path, tt.isDir, got, tt.ignore, tt.reason)
			}
		})
	}
}

func TestLoadSubstitutionStrings(t *testing.T) {
	tmpDir, err := os.MkdirTemp(tempDir, "subs-test")
	if err != nil {
		t.Fatalf("Error creating test temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	t.Run("valid substitutions", func(t *testing.T) {
		sourceDir := filepath.Join(tmpDir, "valid-source")
		os.MkdirAll(sourceDir, 0755)
		csvPath := filepath.Join(sourceDir, "substitution-strings.csv")
		if err := os.WriteFile(csvPath, []byte("SITE,example.com\nYEAR,2024"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		wiki := &Wiki{SourceDir: sourceDir}
		if err := wiki.loadSubstitutionStrings(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(wiki.subStrings) != 2 {
			t.Errorf("expected 2 substitution strings, got %d", len(wiki.subStrings))
		}
	})

	t.Run("duplicate placeholders", func(t *testing.T) {
		sourceDir := filepath.Join(tmpDir, "dup-source")
		os.MkdirAll(sourceDir, 0755)
		csvPath := filepath.Join(sourceDir, "substitution-strings.csv")
		if err := os.WriteFile(csvPath, []byte("FOO,value1\nFOO,value2"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		wiki := &Wiki{SourceDir: sourceDir}
		err := wiki.loadSubstitutionStrings()
		if err == nil {
			t.Error("expected error for duplicate placeholders")
		}
		if !strings.Contains(err.Error(), "duplicate") {
			t.Errorf("error should mention 'duplicate', got: %v", err)
		}
	})

	t.Run("invalid placeholder", func(t *testing.T) {
		sourceDir := filepath.Join(tmpDir, "invalid-source")
		os.MkdirAll(sourceDir, 0755)
		csvPath := filepath.Join(sourceDir, "substitution-strings.csv")
		if err := os.WriteFile(csvPath, []byte("FOO BAR,value"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		wiki := &Wiki{SourceDir: sourceDir}
		err := wiki.loadSubstitutionStrings()
		if err == nil {
			t.Error("expected error for invalid placeholder")
		}
	})
}

// ============================================================================
// Generator Tests (generator.go)
// ============================================================================

func TestIsPathMarkdown(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"file.md", true},
		{"file.mdwn", true},
		{"file.markdown", true},
		{"file.txt", false},
		{"file.html", false},
		{"file.MD", true}, // case-insensitive - .MD is recognized
		{"path/to/file.md", true},
		{"file", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := isPathMarkdown(tt.path); got != tt.want {
				t.Errorf("isPathMarkdown(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestRemoveFileExtension(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"file.md", "file"},
		{"path/to/file.md", "path/to/file"},
		{"file.tar.gz", "file.tar"},
		{"file", "file"},
		{".hidden", ".hidden"},           // Dotfile without extension should remain unchanged
		{".hidden.md", ".hidden"},        // Dotfile with extension should have extension removed
		{"path/.hidden", "path/.hidden"}, // Dotfile in path without extension
		{".gitignore", ".gitignore"},     // Another common dotfile
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := removeFileExtension(tt.input); got != tt.want {
				t.Errorf("removeFileExtension(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCheckForStyleDirective(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantStyle bool
		wantData  string
	}{
		{
			name:      "github directive at start",
			input:     "#[style(github)]\n# Title\nContent",
			wantStyle: true,
			wantData:  "# Title\nContent",
		},
		{
			name:      "github directive with whitespace",
			input:     "  #[style(github)]\n# Title",
			wantStyle: true,
			wantData:  "# Title",
		},
		{
			name:      "no directive",
			input:     "# Title\nContent",
			wantStyle: false,
			wantData:  "# Title\nContent",
		},
		{
			name:      "directive not at start",
			input:     "# Title\n#[style(github)]",
			wantStyle: false,
			wantData:  "# Title\n#[style(github)]",
		},
		{
			name:      "with BOM",
			input:     "\xef\xbb\xbf#[style(github)]\nContent",
			wantStyle: true,
			wantData:  "Content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStyle, gotData := checkForStyleDirective([]byte(tt.input))
			if gotStyle != tt.wantStyle {
				t.Errorf("style = %v, want %v", gotStyle, tt.wantStyle)
			}
			if string(gotData) != tt.wantData {
				t.Errorf("data = %q, want %q", gotData, tt.wantData)
			}
		})
	}
}

// ============================================================================
// Filesystem Tests (filesystem.go)
// ============================================================================

func TestSourceIsOlder(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp(tempDir, "test-source-older")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	t.Run("source newer than dest", func(t *testing.T) {
		// Create older dest file
		destPath := filepath.Join(tmpDir, "dest1.txt")
		oldTime := time.Now().Add(-2 * time.Second)
		if err := os.WriteFile(destPath, []byte("dest"), 0644); err != nil {
			t.Fatal(err)
		}
		os.Chtimes(destPath, oldTime, oldTime)

		time.Sleep(100 * time.Millisecond) // Ensure timestamps differ

		// Create source file
		sourcePath := filepath.Join(tmpDir, "source1.txt")
		if err := os.WriteFile(sourcePath, []byte("source"), 0644); err != nil {
			t.Fatal(err)
		}
		sourceInfo, _ := os.Stat(sourcePath)

		if sourceIsOlder(sourceInfo, destPath) {
			t.Error("expected source to be newer than dest")
		}
	})

	t.Run("dest does not exist", func(t *testing.T) {
		sourcePath := filepath.Join(tmpDir, "source2.txt")
		if err := os.WriteFile(sourcePath, []byte("source"), 0644); err != nil {
			t.Fatal(err)
		}
		sourceInfo, _ := os.Stat(sourcePath)

		if sourceIsOlder(sourceInfo, filepath.Join(tmpDir, "nonexistent.txt")) {
			t.Error("expected false when dest doesn't exist")
		}
	})
}

func TestIsDirectoryEmpty(t *testing.T) {
	tmpDir, err := os.MkdirTemp(tempDir, "test-empty-dir")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	t.Run("empty directory", func(t *testing.T) {
		emptyDir := filepath.Join(tmpDir, "empty")
		os.Mkdir(emptyDir, 0755)
		isEmpty, err := isDirectoryEmpty(emptyDir)
		if err != nil {
			t.Fatal(err)
		}
		if !isEmpty {
			t.Error("expected directory to be empty")
		}
	})

	t.Run("non-empty directory", func(t *testing.T) {
		nonEmptyDir := filepath.Join(tmpDir, "nonempty")
		os.Mkdir(nonEmptyDir, 0755)
		os.WriteFile(filepath.Join(nonEmptyDir, "file.txt"), []byte("content"), 0644)
		isEmpty, err := isDirectoryEmpty(nonEmptyDir)
		if err != nil {
			t.Fatal(err)
		}
		if isEmpty {
			t.Error("expected directory to not be empty")
		}
	})
}

// ============================================================================
// Wiki Tests (wiki.go)
// ============================================================================

func TestNewWikiValidation(t *testing.T) {
	tmpDir, err := os.MkdirTemp(tempDir, "test-wiki-validation")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create valid wiki structure
	validSource := filepath.Join(tmpDir, "source")
	validContent := filepath.Join(validSource, "content")
	validDest := filepath.Join(tmpDir, "dest")
	os.MkdirAll(validContent, 0755)
	os.MkdirAll(validDest, 0755)

	t.Run("valid wiki", func(t *testing.T) {
		wiki, err := NewWiki(validSource, validDest)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if wiki == nil {
			t.Error("expected wiki to be non-nil")
		}
	})

	t.Run("same source and dest", func(t *testing.T) {
		_, err := NewWiki(validSource, validSource)
		if err == nil {
			t.Error("expected error for same source and dest")
		}
	})

	t.Run("dest under source", func(t *testing.T) {
		nestedDest := filepath.Join(validSource, "output")
		os.MkdirAll(nestedDest, 0755)
		_, err := NewWiki(validSource, nestedDest)
		if err == nil {
			t.Error("expected error for dest under source")
		}
	})

	t.Run("source under dest", func(t *testing.T) {
		nestedSource := filepath.Join(validDest, "source")
		nestedContent := filepath.Join(nestedSource, "content")
		os.MkdirAll(nestedContent, 0755)
		_, err := NewWiki(nestedSource, validDest)
		if err == nil {
			t.Error("expected error for source under dest")
		}
	})

	t.Run("missing source dir", func(t *testing.T) {
		_, err := NewWiki(filepath.Join(tmpDir, "nonexistent"), validDest)
		if err == nil {
			t.Error("expected error for missing source")
		}
	})

	t.Run("missing content dir", func(t *testing.T) {
		noContentSource := filepath.Join(tmpDir, "no-content")
		os.MkdirAll(noContentSource, 0755)
		_, err := NewWiki(noContentSource, validDest)
		if err == nil {
			t.Error("expected error for missing content dir")
		}
	})
}

// ============================================================================
// Watcher Tests (watcher.go)
// ============================================================================

func TestFilesSnapshotsAreEqual(t *testing.T) {
	snapshot1 := []fileSnapshot{
		{name: "file1.md", timestamp: 1000, isDir: false},
		{name: "dir1", timestamp: 2000, isDir: true},
	}

	snapshot2 := []fileSnapshot{
		{name: "file1.md", timestamp: 1000, isDir: false},
		{name: "dir1", timestamp: 2000, isDir: true},
	}

	snapshot3 := []fileSnapshot{
		{name: "file1.md", timestamp: 1001, isDir: false}, // different timestamp
		{name: "dir1", timestamp: 2000, isDir: true},
	}

	snapshot4 := []fileSnapshot{
		{name: "file1.md", timestamp: 1000, isDir: false},
	}

	t.Run("equal snapshots", func(t *testing.T) {
		if !filesSnapshotsAreEqual(snapshot1, snapshot2) {
			t.Error("expected snapshots to be equal")
		}
	})

	t.Run("different timestamps", func(t *testing.T) {
		if filesSnapshotsAreEqual(snapshot1, snapshot3) {
			t.Error("expected snapshots to be different")
		}
	})

	t.Run("different lengths", func(t *testing.T) {
		if filesSnapshotsAreEqual(snapshot1, snapshot4) {
			t.Error("expected snapshots to be different")
		}
	})

	t.Run("nil snapshot", func(t *testing.T) {
		if filesSnapshotsAreEqual(nil, snapshot1) {
			t.Error("expected nil comparison to return false")
		}
	})

	t.Run("both nil", func(t *testing.T) {
		if filesSnapshotsAreEqual(nil, nil) {
			t.Error("expected nil-nil comparison to return false")
		}
	})
}

func TestTakeFilesSnapshot(t *testing.T) {
	tmpDir, err := os.MkdirTemp(tempDir, "snapshot-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test structure
	subdir := filepath.Join(tmpDir, "subdir")
	os.MkdirAll(subdir, 0755)
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content"), 0644)
	os.WriteFile(filepath.Join(subdir, "file2.txt"), []byte("content"), 0644)

	snapshot, err := takeFilesSnapshot(context.Background(), tmpDir, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(snapshot) < 3 { // tmpDir + subdir + at least 1 file
		t.Errorf("expected at least 3 entries in snapshot, got %d", len(snapshot))
	}

	// Verify snapshot contains expected paths
	foundFile1 := false
	foundSubdir := false
	for _, s := range snapshot {
		if s.name == filepath.Join(tmpDir, "file1.txt") {
			foundFile1 = true
			if s.isDir {
				t.Error("file1.txt should not be marked as directory")
			}
		}
		if s.name == subdir {
			foundSubdir = true
			if !s.isDir {
				t.Error("subdir should be marked as directory")
			}
		}
	}

	if !foundFile1 {
		t.Error("snapshot should contain file1.txt")
	}
	if !foundSubdir {
		t.Error("snapshot should contain subdir")
	}
}

func TestTakeFilesSnapshotWithIgnorePatterns(t *testing.T) {
	tmpDir, err := os.MkdirTemp(tempDir, "snapshot-ignore-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test structure with files that should be ignored
	os.WriteFile(filepath.Join(tmpDir, "include.txt"), []byte("content"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "exclude.log"), []byte("content"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "temp.tmp"), []byte("content"), 0644)

	logDir := filepath.Join(tmpDir, "logs")
	os.MkdirAll(logDir, 0755)
	os.WriteFile(filepath.Join(logDir, "debug.log"), []byte("content"), 0644)

	// Create ignore matcher with patterns to exclude .log, .tmp files and logs directory
	ignorePatterns := []string{"*.log", "*.tmp", "logs/"}
	ignoreMatcher, err := NewIgnoreMatcher(ignorePatterns)
	if err != nil {
		t.Fatalf("failed to create ignore matcher: %v", err)
	}

	// Take snapshot with ignore patterns
	snapshot, err := takeFilesSnapshot(context.Background(), tmpDir, ignoreMatcher)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify ignored files are not in snapshot
	for _, s := range snapshot {
		if s.name == filepath.Join(tmpDir, "exclude.log") {
			t.Error("snapshot should not contain exclude.log (it should be ignored)")
		}
		if s.name == filepath.Join(tmpDir, "temp.tmp") {
			t.Error("snapshot should not contain temp.tmp (it should be ignored)")
		}
		if s.name == logDir {
			t.Error("snapshot should not contain logs directory (it should be ignored)")
		}
		if s.name == filepath.Join(logDir, "debug.log") {
			t.Error("snapshot should not contain debug.log inside logs directory")
		}
	}

	// Verify included files are in snapshot
	foundInclude := false
	for _, s := range snapshot {
		if s.name == filepath.Join(tmpDir, "include.txt") {
			foundInclude = true
		}
	}
	if !foundInclude {
		t.Error("snapshot should contain include.txt")
	}
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestWikiWithSubstitutionStrings(t *testing.T) {
	tmpDir, err := os.MkdirTemp(tempDir, "test-substitutions")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create source structure
	sourceDir := filepath.Join(tmpDir, "source")
	contentDir := filepath.Join(sourceDir, "content")
	destDir := filepath.Join(tmpDir, "dest")
	os.MkdirAll(contentDir, 0755)
	os.MkdirAll(destDir, 0755)

	// Create substitution strings file
	subsPath := filepath.Join(sourceDir, "substitution-strings.csv")
	os.WriteFile(subsPath, []byte("SITE,example.com\nYEAR,2024"), 0644)

	// Create markdown with placeholders (including an undefined one)
	mdPath := filepath.Join(contentDir, "test.md")
	mdContent := "# Welcome to {{SITE}}\nCopyright {{YEAR}}\nUnknown: {{MISSING}}"
	os.WriteFile(mdPath, []byte(mdContent), 0644)

	// Generate wiki
	wiki, err := NewWiki(sourceDir, destDir)
	if err != nil {
		t.Fatalf("failed to create wiki: %v", err)
	}

	if err := wiki.Generate(context.Background(), true, false, false, "test"); err != nil {
		t.Fatalf("failed to generate wiki: %v", err)
	}

	// Check output
	htmlPath := filepath.Join(destDir, "test.html")
	content, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	contentStr := string(content)

	// Verify substitutions were applied
	if !strings.Contains(contentStr, "example.com") {
		t.Error("SITE substitution not applied")
	}
	if !strings.Contains(contentStr, "2024") {
		t.Error("YEAR substitution not applied")
	}

	// Verify undefined placeholder remains unchanged
	if !strings.Contains(contentStr, "{{MISSING}}") {
		t.Error("undefined placeholder should remain unchanged")
	}

	// Verify defined placeholders don't remain
	if strings.Contains(contentStr, "{{SITE}}") {
		t.Error("{{SITE}} placeholder should have been replaced")
	}
	if strings.Contains(contentStr, "{{YEAR}}") {
		t.Error("{{YEAR}} placeholder should have been replaced")
	}
}
