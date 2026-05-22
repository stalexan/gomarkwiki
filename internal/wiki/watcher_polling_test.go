package wiki

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// pollTestSetup builds a temp source/dest tree with a single index.md, starts
// wiki.Generate in a goroutine with watch=true and the given poll interval,
// and waits for the initial HTML to land. It returns the wiki dirs and a
// cleanup func that cancels the context and joins the goroutine.
func pollTestSetup(t *testing.T, pollInterval time.Duration, initialContent string) (sourceDir, contentDir, destDir string, cleanup func()) {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "watcher-polling")
	if err != nil {
		t.Fatal(err)
	}

	sourceDir = filepath.Join(tempDir, "source")
	contentDir = filepath.Join(sourceDir, "content")
	destDir = filepath.Join(tempDir, "dest")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(contentDir, "index.md"), []byte(initialContent), 0644); err != nil {
		t.Fatal(err)
	}

	w, err := NewWiki(sourceDir, destDir)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("NewWiki failed: %v", err)
	}
	w.PollInterval = pollInterval

	ctx, cancel := context.WithCancel(context.Background())
	errChan := make(chan error, 1)
	go func() {
		errChan <- w.Generate(ctx, true, false, true, "test")
	}()

	// Wait for the initial generation to produce index.html.
	indexHTML := filepath.Join(destDir, "index.html")
	if !waitForFile(indexHTML, 2*time.Second) {
		cancel()
		<-errChan
		os.RemoveAll(tempDir)
		t.Fatalf("initial generation did not produce %s within timeout", indexHTML)
	}

	cleanup = func() {
		cancel()
		// Drain the goroutine; ignore context-cancelled errors.
		<-errChan
		os.RemoveAll(tempDir)
	}
	return sourceDir, contentDir, destDir, cleanup
}

// waitForFile polls for path to exist, returning true if seen before timeout.
func waitForFile(path string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

// waitForFileContent polls for path to exist AND contain substr.
func waitForFileContent(path, substr string, timeout time.Duration) (string, bool) {
	deadline := time.Now().Add(timeout)
	var last string
	for time.Now().Before(deadline) {
		if data, err := os.ReadFile(path); err == nil {
			last = string(data)
			if strings.Contains(last, substr) {
				return last, true
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	return last, false
}

// waitForFileMtimeChange polls until path's mtime differs from baseline.
func waitForFileMtimeChange(path string, baseline time.Time, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if info, err := os.Stat(path); err == nil {
			if !info.ModTime().Equal(baseline) {
				return true
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

func TestPollingDetectsFileModification(t *testing.T) {
	_, contentDir, destDir, cleanup := pollTestSetup(t, 50*time.Millisecond, "# Original heading")
	defer cleanup()

	indexHTML := filepath.Join(destDir, "index.html")

	// Sanity: initial content reflected in output.
	if _, ok := waitForFileContent(indexHTML, "Original heading", 2*time.Second); !ok {
		t.Fatalf("initial HTML did not reflect source content")
	}

	// Modify the source.
	if err := os.WriteFile(filepath.Join(contentDir, "index.md"), []byte("# Updated heading"), 0644); err != nil {
		t.Fatal(err)
	}

	// Expect regen to pick up the change. Generous timeout accounts for
	// poll interval (50ms) + stability backoff (>=100ms initial wait, 2
	// snapshots).
	if got, ok := waitForFileContent(indexHTML, "Updated heading", 3*time.Second); !ok {
		t.Errorf("polling did not regenerate after modification.\nLatest HTML: %s", got)
	}
}

func TestPollingDetectsFileAdd(t *testing.T) {
	_, contentDir, destDir, cleanup := pollTestSetup(t, 50*time.Millisecond, "# Index")
	defer cleanup()

	// Add a new markdown file.
	if err := os.WriteFile(filepath.Join(contentDir, "new.md"), []byte("# Newly added"), 0644); err != nil {
		t.Fatal(err)
	}

	newHTML := filepath.Join(destDir, "new.html")
	if !waitForFile(newHTML, 3*time.Second) {
		t.Errorf("polling did not regenerate after file add: %s not created", newHTML)
	}
}

func TestPollingDetectsFileDelete(t *testing.T) {
	_, contentDir, destDir, cleanup := pollTestSetup(t, 50*time.Millisecond, "# Index")
	defer cleanup()

	// Create an extra file and wait for it to regenerate.
	extraMD := filepath.Join(contentDir, "extra.md")
	if err := os.WriteFile(extraMD, []byte("# Extra"), 0644); err != nil {
		t.Fatal(err)
	}
	extraHTML := filepath.Join(destDir, "extra.html")
	if !waitForFile(extraHTML, 3*time.Second) {
		t.Fatalf("setup failed: %s not created", extraHTML)
	}

	// Delete the source file. The HTML output is not removed (we did not
	// run with -clean), so instead we verify the watcher keeps responding:
	// a follow-up unrelated edit should still trigger regen.
	if err := os.Remove(extraMD); err != nil {
		t.Fatal(err)
	}

	// Give the polling loop time to observe the delete + stabilize.
	time.Sleep(300 * time.Millisecond)

	// Now do a modification and verify it still propagates.
	indexMD := filepath.Join(contentDir, "index.md")
	if err := os.WriteFile(indexMD, []byte("# After delete"), 0644); err != nil {
		t.Fatal(err)
	}
	indexHTML := filepath.Join(destDir, "index.html")
	if got, ok := waitForFileContent(indexHTML, "After delete", 3*time.Second); !ok {
		t.Errorf("watcher stopped responding after file deletion.\nLatest HTML: %s", got)
	}
}

func TestPollingDebouncesRapidEdits(t *testing.T) {
	_, contentDir, destDir, cleanup := pollTestSetup(t, 50*time.Millisecond, "# Initial")
	defer cleanup()

	indexHTML := filepath.Join(destDir, "index.html")
	if _, ok := waitForFileContent(indexHTML, "Initial", 2*time.Second); !ok {
		t.Fatalf("initial HTML missing")
	}

	indexMD := filepath.Join(contentDir, "index.md")

	// Fire several rapid writes well within the stability window. The
	// debounce code in waitForStability should coalesce these into a single
	// regen with the final content.
	rapidVersions := []string{
		"# v1",
		"# v2",
		"# v3",
		"# v4",
		"# Final version",
	}
	for _, v := range rapidVersions {
		if err := os.WriteFile(indexMD, []byte(v), 0644); err != nil {
			t.Fatal(err)
		}
		time.Sleep(20 * time.Millisecond)
	}

	// Final-state assertion: the last write must appear in the HTML.
	if got, ok := waitForFileContent(indexHTML, "Final version", 5*time.Second); !ok {
		t.Errorf("debounced regen did not converge on final source.\nLatest HTML: %s", got)
	}
}

func TestPollingNoRegenWhenIdle(t *testing.T) {
	_, _, destDir, cleanup := pollTestSetup(t, 50*time.Millisecond, "# Static")
	defer cleanup()

	indexHTML := filepath.Join(destDir, "index.html")
	if _, ok := waitForFileContent(indexHTML, "Static", 2*time.Second); !ok {
		t.Fatalf("initial HTML missing")
	}

	// Capture baseline mtime, then wait for several poll intervals with no
	// edits. The HTML mtime must not advance.
	info, err := os.Stat(indexHTML)
	if err != nil {
		t.Fatal(err)
	}
	baseline := info.ModTime()

	if waitForFileMtimeChange(indexHTML, baseline, 400*time.Millisecond) {
		t.Errorf("HTML regenerated despite no source changes (mtime advanced from %v)", baseline)
	}
}
