package wiki

import (
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
const STYLES_PATH = "./static/style.css"

var (
	packageDir  string
	stylesPath  string
	tempDir     string
	testDataDir string
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
	stylesPath = STYLES_PATH

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
	if err = theWiki.Generate(true, false, false, "test"); err != nil {
		t.Fatalf("Error generating wiki: %v", err)
	}

	// Create expected output.
	expectedOutputSourceDir := filepath.Join(testCaseDataDir, "expected-output")
	expectedOutputDir := filepath.Join(testCaseTempDir, "expected-output")
	if err := copyDir(expectedOutputSourceDir, expectedOutputDir); err != nil {
		t.Fatalf("Failed to create expected output dir %s: %v", expectedOutputDir, err)
	}
	if err := copyFile(stylesPath, expectedOutputDir); err != nil {
		t.Fatalf("Failed to create style.css for expected output dir %s: %v", expectedOutputDir, err)
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
