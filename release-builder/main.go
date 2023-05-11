package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

const OUTPUT_DIR = "/output"
const EXTRACTED_SOURCE_DIR = "/output/source"

var versionRegex = regexp.MustCompile(`^v\d+\.\d+\.\d+`)

func formatMessage(format string, args []interface{}) string {
	message := fmt.Sprintf(format, args...)
	if !strings.HasSuffix(message, "\n") {
		message += "\n"
	}
	return message
}

func printMessage(format string, args ...interface{}) {
	message := formatMessage(format, args)
	message = "\x1b[32m" + message + "\x1b[0m"
	fmt.Print(message)
}

func printErrorAndExit(format string, args ...interface{}) {
	message := formatMessage(format, args)
	message = "\x1b[31m" + "ERROR: " + message + "\x1b[0m"
	fmt.Fprint(os.Stderr, message)
	os.Exit(1)
}

func run(name string, args ...string) {
	command := exec.Command(name, args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		printErrorAndExit("Failed to run %s %s: %v", name, args, err)
	}
}

func mkdir(dir string) {
	if err := os.Mkdir(dir, 0755); err != nil {
		printErrorAndExit("Failed to mkdir %v: %v", dir, err)
	}
}

func rmdir(dir string) {
	if err := os.RemoveAll(dir); err != nil {
		printErrorAndExit("Failed to rmdir %v: %v", dir, err)
	}
}

func rm(file string) {
	if err := os.Remove(file); err != nil {
		if !os.IsNotExist(err) {
			printErrorAndExit("Failed to remove %v: %v", file, err)
		}
	}
}

func readdir(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		printErrorAndExit("Failed to read directory %v: %v", dir, err)
	}
	filenames := make([]string, 0, len(entries))
	for _, entry := range entries {
		filenames = append(filenames, entry.Name())
	}
	return filenames
}

func chmod(file string, mode os.FileMode) {
	if err := os.Chmod(file, mode); err != nil {
		printErrorAndExit("Failed to chmod %v to %s: %v", file, mode, err)
	}
}

func preCheckOutputDir() {
	// Check that output directory is empty.
	filenames := readdir(OUTPUT_DIR)
	if len(filenames) != 0 {
		printErrorAndExit("Output directory is not empty")
	}
}

func preCheckBranchMain() {
	// Check that we're on main branch.
	output, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		printErrorAndExit("Failed to run 'git': %v", err)
	}
	branch := strings.TrimSpace(string(output))
	if branch != "main" {
		printErrorAndExit("Wrong branch: %s", branch)
	}
}

func preCheckUncommittedChanges() {
	// Check that there are no uncommitted changes.
	changes := uncommittedChanges()
	if len(changes) > 0 {
		printErrorAndExit("Uncommitted changes found:\n%s", changes)
	}
}

func uncommittedChanges(dirs ...string) string {
	args := []string{"status", "--porcelain", "--untracked-files=no"}
	if len(dirs) > 0 {
		args = append(args, dirs...)
	}
	changes, err := exec.Command("git", args...).Output()
	if err != nil {
		printErrorAndExit("Failed to run command: %v", err)
	}
	return string(changes)
}

func determineVersion(commit string) string {
	output, err := exec.Command("git", "describe", "--tags", commit).Output()
	if err != nil {
		printErrorAndExit("Commit %v was not found: %v", commit, err)
	}
	version := strings.TrimSpace(string(output))
	if !versionRegex.MatchString(version) {
		printErrorAndExit("Invalid version: %v", version)
	}
	return version
}

func tarSource(version string) string {
	filename := fmt.Sprintf("gomarkwiki-%s.tar.gz", version)
	path := filepath.Join(OUTPUT_DIR, filename)
	cmd := fmt.Sprintf("git archive --format=tar --prefix=gomarkwiki-%s/ %s | gzip -n > %s",
		version, version, path)
	run("sh", "-c", cmd)
	printMessage("Created %s", filename)
	return path
}

func extractTar(tarPath string) {
	mkdir(EXTRACTED_SOURCE_DIR)
	run("tar", "xz", "--strip-components=1", "-f", tarPath, "-C", EXTRACTED_SOURCE_DIR)
}

var buildTargets = map[string][]string{
	"aix":     {"ppc64"},
	"darwin":  {"amd64", "arm64"},
	"freebsd": {"386", "amd64", "arm"},
	"linux":   {"386", "amd64", "arm", "arm64", "ppc64le", "mips", "mipsle", "mips64", "mips64le", "riscv64", "s390x"},
	"netbsd":  {"386", "amd64"},
	"openbsd": {"386", "amd64"},
	"windows": {"386", "amd64"},
	"solaris": {"amd64"},
}

func runBuild(version string) {
	// Download dependencies.
	if err := os.Chdir(EXTRACTED_SOURCE_DIR); err != nil {
		printErrorAndExit("Failed to cd to %s: %v", EXTRACTED_SOURCE_DIR, err)
	}
	run("go", "mod", "download")

	// Build each target.
	start := time.Now()
	for goos, goarchs := range buildTargets {
		for _, goarch := range goarchs {
			targetStart := time.Now()
			targetName := fmt.Sprintf("%v/%v", goos, goarch)
			printMessage("Building %v", targetName)
			buildForTarget(goos, goarch, version)
			printMessage("Built %v in %.3fs", targetName, time.Since(targetStart).Seconds())
		}
	}
	printMessage("Build finished in %.3fs", time.Since(start).Seconds())
}

func buildForTarget(goos, goarch, version string) {
	// Determine output filename.
	binaryFilename := fmt.Sprintf("gomarkwiki_%v_%v_%v", version, goos, goarch)
	if goos == "windows" {
		binaryFilename += ".exe"
	}
	binaryPath := filepath.Join(OUTPUT_DIR, binaryFilename)

	// Build.
	command := exec.Command("go", "build",
		"-o", binaryPath,
		"-ldflags", fmt.Sprintf("-s -w -X 'main.version=%s'", version),
		"./cmd/main.go",
	)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Dir = EXTRACTED_SOURCE_DIR
	command.Env = append(os.Environ(),
		"CGO_ENABLED=0",
		"GOOS="+goos,
		"GOARCH="+goarch,
	)
	if err := command.Run(); err != nil {
		printErrorAndExit("Failed to build %v/%v: %v", goos, goarch, err)
	}

	// Make executable.
	chmod(binaryPath, 0755)

	// Compress.
	compress(goos, binaryFilename)
}

func compress(goos, binaryFilename string) {
	// Create command.
	var command *exec.Cmd
	switch goos {
	case "windows":
		outputFile := strings.TrimSuffix(binaryFilename, ".exe") + ".zip"
		command = exec.Command("zip", "-q", "-X", outputFile, binaryFilename)
	default:
		command = exec.Command("bzip2", binaryFilename)
	}
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Dir = OUTPUT_DIR

	// Run command.
	if err := command.Run(); err != nil {
		printErrorAndExit("Failed to compress: %v", err)
	}

	// Remove uncompressed binary.
	rm(filepath.Join(OUTPUT_DIR, binaryFilename))
}

func computeSums() {
	// Get list of files to compute sums for.
	filenames := readdir(OUTPUT_DIR)

	// Create sums file.
	const shaFileName = "SHA256SUMS"
	shaPath := filepath.Join(OUTPUT_DIR, shaFileName)
	file, err := os.Create(shaPath)
	if err != nil {
		printErrorAndExit("Failed to create create %v: %v", shaFileName, err)
	}
	defer file.Close()

	// Compute sums.
	command := exec.Command("sha256sum", filenames...)
	command.Stdout = file
	command.Stderr = os.Stderr
	command.Dir = OUTPUT_DIR
	if err = command.Run(); err != nil {
		printErrorAndExit("Failed to compute sums: %v", err)
	}
}

func main() {
	// Which version of Go is this?
	printMessage("Go version: %s", runtime.Version())

	// Which commit of gomarkwiki to use for build?
	if len(os.Args) < 2 {
		printMessage("USAGE: %s [commit]", os.Args[0])
		os.Exit(1)
	}
	commit := os.Args[1]
	printMessage("gomarkwiki commit: %s", commit)

	// Check output dir and repo.
	preCheckOutputDir()
	preCheckBranchMain()
	preCheckUncommittedChanges()

	// Tar source.
	version := determineVersion(commit)
	printMessage("gomarkwiki version: %s", version)
	tarPath := tarSource(version)

	// Extract source.
	extractTar(tarPath)

	// Build.
	runBuild(version)
	rmdir(EXTRACTED_SOURCE_DIR)

	// Compute sums.
	computeSums()
}
