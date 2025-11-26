package wiki

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
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
		"\\.tmp$\n"
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
