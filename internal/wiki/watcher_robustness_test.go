package wiki

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatcherRobustness(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "watcher-robustness")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	sourceDir := filepath.Join(tempDir, "source")
	contentDir := filepath.Join(sourceDir, "content")
	destDir := filepath.Join(tempDir, "dest")
	os.MkdirAll(contentDir, 0755)
	os.MkdirAll(destDir, 0755)

	// Create a readable file
	os.WriteFile(filepath.Join(contentDir, "readable.md"), []byte("# Readable"), 0644)

	// Create a subdirectory
	unreadableDir := filepath.Join(contentDir, "unreadable")
	os.MkdirAll(unreadableDir, 0755)
	os.WriteFile(filepath.Join(unreadableDir, "file.md"), []byte("# Hidden"), 0644)

	// Make the subdirectory unreadable
	if err := os.Chmod(unreadableDir, 0000); err != nil {
		t.Skip("Skipping test because chmod failed (running as root/Windows?)")
	}
	defer os.Chmod(unreadableDir, 0755) // Restore permissions for cleanup

	// Test NewWiki and Generate with watch=true
	// We expect it to NOT fail completely, or at least handle it gracefully?
	// The current code suggests it will fail.

	wiki, err := NewWiki(sourceDir, destDir)
	if err != nil {
		t.Fatalf("NewWiki failed: %v", err)
	}

	// create a context that we can cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Try to start watcher (Generate calls watch if watch=true)
	// We run it in a goroutine because it blocks
	errChan := make(chan error, 1)
	go func() {
		// We expect this to fail if my hypothesis is correct
		errChan <- wiki.Generate(ctx, true, false, true, "test")
	}()

	// Wait a bit to see if it crashes immediately
	// If it succeeds, it blocks, so we cancel context to stop it
	select {
	case err := <-errChan:
		if err != nil && err != context.Canceled {
			t.Errorf("Generate failed with error: %v. It should have handled the permission error gracefully.", err)
		} else if err == nil {
			// It finished? That's weird if watch=true, it should block.
			// Unless we cancelled it? No, we haven't cancelled yet.
			// Maybe it returned nil immediately?
			t.Error("Generate finished unexpectedly (should block)")
		} else {
			// context.Canceled
			t.Log("Generate cancelled as expected (it was running)")
		}
	case <-time.After(500 * time.Millisecond):
		// It's still running, which is what we want!
		t.Log("Generate is running correctly despite permission error")
		cancel()
		err := <-errChan
		if err != nil && err != context.Canceled {
			t.Errorf("Generate returned error after cancel: %v", err)
		}
	}
}
