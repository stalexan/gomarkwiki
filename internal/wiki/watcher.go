// Package wiki generates HTML from markdown for a given wiki.
package wiki

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/stalexan/gomarkwiki/internal/util"
)

// Watcher timing configuration constants.
// These control the file system change detection and stability waiting behavior.
const (
	// CHANGE_WAIT is the initial wait time before the first snapshot comparison
	// when detecting file changes. This gives file operations time to complete
	// before we start checking for stability.
	CHANGE_WAIT = 100 * time.Millisecond

	// MAX_CHANGE_WAIT is the maximum wait time between snapshot comparisons
	// during the exponential backoff algorithm. This caps the wait time to
	// prevent excessive delays when waiting for rapid file operations to stabilize
	// (e.g., bulk edits, git operations).
	MAX_CHANGE_WAIT = 5000 * time.Millisecond

	// MAX_REGEN_INTERVAL is the maximum time interval before forcing a periodic
	// regeneration of the wiki, even if no file changes are detected. This ensures
	// the wiki stays up-to-date even if file system events are missed.
	MAX_REGEN_INTERVAL = 10 * time.Minute
)

// fileSnapshot records the name and modification time for a given file or directory.
type fileSnapshot struct {
	name      string
	timestamp int64
	isDir     bool
}

// Watcher manages file system watching for a wiki content directory.
// It encapsulates the complex state machine for detecting changes and
// waiting for file operations to stabilize.
//
// State Machine:
//   - Idle: Waiting for file changes (listening to fsnotify events)
//   - EventDetected: A file change was detected
//   - WaitingForStability: Comparing snapshots to ensure changes are complete
//   - Ready: Changes stabilized, ready for regeneration
//   - Timeout: Periodic regeneration timer expired
//
// The watcher reuses a single fsnotify.Watcher instance across cycles
// for efficiency, and maintains snapshot state to detect rapid changes.
type Watcher struct {
	// Configuration
	contentDir string
	subsPath   string
	sourceDir  string // For error messages

	// Watcher instance (reused across cycles)
	fsWatcher *fsnotify.Watcher
	mu        sync.Mutex // Protects all fsWatcher access

	// Current state
	snapshot    []fileSnapshot
	subsModTime int64 // Last known modification time of substitution strings file

	// Context management
	ctx    context.Context
	cancel context.CancelFunc
}

// WatchResult represents the result of waiting for a change.
type WatchResult struct {
	Snapshot []fileSnapshot // New snapshot after changes stabilized
	Regen    bool           // Whether full regeneration is needed
	Timeout  bool           // Whether the wait timed out
}

// NewWatcher creates a new Watcher instance for the given directories.
// The parent context is used for cancellation - when it's cancelled, the watcher will stop.
func NewWatcher(parentCtx context.Context, contentDir, subsPath, sourceDir string) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %v", err)
	}

	// Set up recursive watching
	if err := watchDirRecursive(parentCtx, contentDir, fsWatcher); err != nil {
		fsWatcher.Close()
		return nil, fmt.Errorf("failed to initialize watcher for %s: %v", contentDir, err)
	}

	// Watch substitution strings file if provided
	var subsModTime int64
	if subsPath != "" {
		if err := fsWatcher.Add(subsPath); err != nil {
			fsWatcher.Close()
			return nil, fmt.Errorf("failed to watch '%s': %v", subsPath, err)
		}
		// Record initial modification time of substitution strings file
		if info, err := os.Stat(subsPath); err == nil {
			subsModTime = info.ModTime().Unix()
		}
	}

	// Create a child context that will be cancelled when the watcher is closed
	ctx, cancel := context.WithCancel(parentCtx)

	return &Watcher{
		contentDir:  contentDir,
		subsPath:    subsPath,
		sourceDir:   sourceDir,
		fsWatcher:   fsWatcher,
		subsModTime: subsModTime,
		ctx:         ctx,
		cancel:      cancel,
	}, nil
}

// Close releases resources associated with the watcher.
func (w *Watcher) Close() error {
	w.cancel()
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.fsWatcher != nil {
		err := w.fsWatcher.Close()
		w.fsWatcher = nil // Prevent double-close and signal that watcher is closed
		return err
	}
	return nil
}

// WaitForChange waits for a file change event or timeout.
// It returns a WatchResult indicating what happened and whether regeneration is needed.
//
// State machine transitions:
//  1. [Optional] Check if files changed since last snapshot (race condition protection)
//  2. Wait for file system event OR timeout (MAX_REGEN_INTERVAL)
//  3. If event occurred, wait for changes to stabilize (exponential backoff)
//  4. Return result with new snapshot and regeneration flag
//
// Each call creates a fresh timeout context to ensure periodic regeneration
// happens at least every MAX_REGEN_INTERVAL minutes, even if no file changes occur.
func (w *Watcher) WaitForChange() (*WatchResult, error) {
	// Create timeout context for this cycle.
	// Fresh timeout ensures periodic regeneration even without file changes.
	ctx, cancel := context.WithTimeout(w.ctx, MAX_REGEN_INTERVAL)
	defer cancel()

	// Quick check: if files changed since last snapshot (race condition protection).
	// This handles the case where files changed between generation completing
	// and the watcher starting to listen again. While this check might seem
	// redundant if we just updated the snapshot, it protects against race
	// conditions where files change during the brief window between generation
	// completing and the next watch cycle starting.
	if w.snapshot != nil {
		newSnapshot, err := takeFilesSnapshot(ctx, w.contentDir)
		if err != nil {
			return nil, fmt.Errorf("failed to take snapshot for %s: %v", w.contentDir, err)
		}
		if !filesSnapshotsAreEqual(w.snapshot, newSnapshot) {
			util.PrintVerbose("Files changed between generation cycles. Starting update.")

			// Check if substitution strings file also changed
			regen := w.checkSubsFileChanged()

			// Files changed, wait for stability and return
			stableSnapshot, err := w.waitForStability(ctx)
			if err != nil {
				return nil, err
			}
			return &WatchResult{
				Snapshot: stableSnapshot,
				Regen:    regen,
				Timeout:  false,
			}, nil
		}

		// Even if content files didn't change, check if substitution strings file changed
		if w.checkSubsFileChanged() {
			util.PrintVerbose("Substitution strings file changed between generation cycles. Starting update.")
			return &WatchResult{
				Snapshot: w.snapshot, // No content changes, so snapshot is still valid
				Regen:    true,
				Timeout:  false,
			}, nil
		}
	}

	// Wait for a file system event (or timeout)
	regen, err := w.waitForEvent(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to watch for change event in %s: %v", w.sourceDir, err)
	}

	// Check if context expired (timeout for periodic regeneration)
	if ctx.Err() == context.DeadlineExceeded {
		util.PrintDebug("Regen timer expired for %s", w.sourceDir)
		return &WatchResult{
			Snapshot: w.snapshot,
			Regen:    false,
			Timeout:  true,
		}, nil
	}

	// Wait for changes to stabilize before regenerating
	stableSnapshot, err := w.waitForStability(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for changes to finish for %s: %v", w.sourceDir, err)
	}

	return &WatchResult{
		Snapshot: stableSnapshot,
		Regen:    regen,
		Timeout:  false,
	}, nil
}

// waitForEvent waits for a file system event from the watcher.
//
// State: Idle -> EventDetected
//
// This method blocks until:
//   - A file system event occurs (returns regen flag)
//   - The watcher encounters an error (returns error)
//   - The context times out (returns false, caller handles timeout)
func (w *Watcher) waitForEvent(ctx context.Context) (bool, error) {
	// Check context first - if already cancelled (e.g., by Close()), return early
	// This prevents the race condition where Close() is called between
	// releasing the lock and entering the select statement
	select {
	case <-ctx.Done():
		return false, nil // Context cancelled, caller will handle
	default:
	}

	// Get references to channels while holding the lock to prevent
	// concurrent Close() from causing a race condition
	w.mu.Lock()
	if w.fsWatcher == nil {
		w.mu.Unlock()
		return false, fmt.Errorf("watcher closed while waiting for event in '%s'", w.contentDir)
	}
	eventsChan := w.fsWatcher.Events
	errorsChan := w.fsWatcher.Errors
	w.mu.Unlock()

	select {
	case event, ok := <-eventsChan:
		if !ok {
			return false, fmt.Errorf("watcher unexpectedly closed while watching '%s'", w.contentDir)
		}
		util.PrintDebug("Watcher event detected for %s: %v", w.contentDir, event)

		// Handle new directory creation - add it to watch list
		if event.Has(fsnotify.Create) {
			if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
				w.mu.Lock()
				if w.fsWatcher != nil {
					if err := watchDirRecursive(ctx, event.Name, w.fsWatcher); err != nil {
						util.PrintVerbose("Failed to watch newly created directory '%s': %v", event.Name, err)
						// Continue anyway - snapshot comparison will catch changes
					} else {
						util.PrintDebug("Added newly created directory '%s' to watch list", event.Name)
					}
				}
				w.mu.Unlock()
			}
		}

		// Check if substitution strings file changed
		regen := w.subsPath != "" &&
			(event.Has(fsnotify.Write) || event.Has(fsnotify.Rename)) &&
			filepath.Clean(event.Name) == w.subsPath
		if regen {
			util.PrintVerbose("Substitution strings file '%s' write seen", w.subsPath)
		}

		return regen, nil

	case err, ok := <-errorsChan:
		if !ok {
			return false, fmt.Errorf("failed to read watcher error for %s", w.contentDir)
		}
		return false, fmt.Errorf("watcher error for %s: %v", w.contentDir, err)

	case <-ctx.Done():
		return false, nil // Timeout, caller will handle
	}
}

// waitForStability waits for file changes to stabilize by comparing snapshots.
//
// State: EventDetected -> WaitingForStability -> Ready
//
// Uses exponential backoff (2^n * CHANGE_WAIT) to avoid
// regenerating during rapid file operations (e.g., bulk edits, git operations).
// Continues until two consecutive snapshots match, indicating changes are complete.
func (w *Watcher) waitForStability(ctx context.Context) ([]fileSnapshot, error) {
	util.PrintDebug("Waiting for changes to finish in %s", w.contentDir)

	// Use a result channel pattern for cleaner coordination
	type result struct {
		snapshot []fileSnapshot
		err      error
	}

	resultChan := make(chan result, 1)

	go func() {
		defer close(resultChan)

		// Initial wait before first snapshot comparison (with cancellation support)
		select {
		case <-ctx.Done():
			resultChan <- result{snapshot: w.snapshot}
			return
		case <-time.After(CHANGE_WAIT):
		}

		var snapshot1, snapshot2 []fileSnapshot
		var err error

		// Loop until snapshots match (indicating stability)
		for waitPass := 1; !filesSnapshotsAreEqual(snapshot1, snapshot2); waitPass++ {
			// Check for cancellation
			select {
			case <-ctx.Done():
				// Return current snapshot if available, or last known snapshot
				if snapshot2 != nil {
					resultChan <- result{snapshot: snapshot2}
				} else if snapshot1 != nil {
					resultChan <- result{snapshot: snapshot1}
				} else {
					resultChan <- result{snapshot: w.snapshot}
				}
				return
			default:
			}

			// Log wait status
			message := fmt.Sprintf("Wait for change pass %d for %s", waitPass, w.contentDir)
			if waitPass > 1 {
				util.PrintVerbose(message)
			} else {
				util.PrintDebug(message)
			}

			// Take before snapshot
			if snapshot2 != nil {
				snapshot1 = snapshot2
			} else {
				snapshot1, err = takeFilesSnapshot(ctx, w.contentDir)
				if err != nil {
					resultChan <- result{err: fmt.Errorf("failed to take files snapshot for %s: %v", w.contentDir, err)}
					return
				}
			}

			// Calculate wait time (exponential backoff: 2^n)
			waitTime := CHANGE_WAIT * (1 << (waitPass - 1))
			if waitTime > MAX_CHANGE_WAIT {
				waitTime = MAX_CHANGE_WAIT
			}
			util.PrintDebug("Waiting %d ms for %s", waitTime.Milliseconds(), w.contentDir)

			// Wait with cancellation support
			select {
			case <-ctx.Done():
				if snapshot2 != nil {
					resultChan <- result{snapshot: snapshot2}
				} else {
					resultChan <- result{snapshot: snapshot1}
				}
				return
			case <-time.After(waitTime):
			}

			// Take after snapshot
			snapshot2, err = takeFilesSnapshot(ctx, w.contentDir)
			if err != nil {
				resultChan <- result{err: fmt.Errorf("failed to take files snapshot for %s: %v", w.contentDir, err)}
				return
			}
		}

		// Snapshots match - changes have stabilized
		util.PrintDebug("Snapshots match for %s", w.contentDir)
		resultChan <- result{snapshot: snapshot2}
	}()

	// Wait for result
	res := <-resultChan
	if res.err != nil {
		return nil, res.err
	}

	return res.snapshot, nil
}

// checkSubsFileChanged checks if the substitution strings file has been modified
// since it was last recorded. Returns true if the file changed, and updates
// the stored modification time.
func (w *Watcher) checkSubsFileChanged() bool {
	if w.subsPath == "" {
		return false
	}

	info, err := os.Stat(w.subsPath)
	if err != nil {
		// File might have been deleted or is inaccessible
		// Don't update mod time, but don't trigger regen either
		return false
	}

	currentModTime := info.ModTime().Unix()
	if currentModTime != w.subsModTime {
		w.subsModTime = currentModTime
		return true
	}

	return false
}

// UpdateSnapshot updates the internal snapshot state.
// This should be called after successful generation.
func (w *Watcher) UpdateSnapshot(snapshot []fileSnapshot) {
	w.snapshot = snapshot
	// Also update substitution strings file mod time when snapshot is updated
	w.checkSubsFileChanged()
}

// GetSnapshot returns the current snapshot.
func (w *Watcher) GetSnapshot() []fileSnapshot {
	return w.snapshot
}

// watch watches for changes in the wiki content directory and regenerates files on the fly.
func (wiki *Wiki) watch(ctx context.Context, clean bool, version string) error {
	util.PrintVerbose("Watching for changes in '%s'", wiki.ContentDir)

	// Create watcher with parent context
	watcher, err := NewWatcher(ctx, wiki.ContentDir, wiki.subsPath, wiki.SourceDir)
	if err != nil {
		return fmt.Errorf("failed to create watcher: %v", err)
	}
	defer watcher.Close()

	// Take initial snapshot
	initialSnapshot, err := takeFilesSnapshot(ctx, wiki.ContentDir)
	if err != nil {
		return fmt.Errorf("failed to take initial snapshot: %v", err)
	}
	watcher.UpdateSnapshot(initialSnapshot)

	// Main watch loop
	for {
		// Check for cancellation before waiting
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Wait for a change
		result, err := watcher.WaitForChange()
		if err != nil {
			// Check if error is due to context cancellation
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fmt.Errorf("failed waiting to update %s wiki: %v", wiki.SourceDir, err)
		}

		// Update snapshot
		watcher.UpdateSnapshot(result.Snapshot)
		// Note: UpdateSnapshot also updates subsModTime if subs file changed

		// Reload substitution strings if needed
		if result.Regen {
			util.PrintVerbose("Reloading substitution strings from '%s'", wiki.subsPath)
			if err := wiki.loadSubstitutionStrings(); err != nil {
				return fmt.Errorf("failed to reload substitution strings file: %v", err)
			}
			// Mod time already updated by UpdateSnapshot above
		}

		// Skip generation if this was just a timeout (periodic regen)
		if result.Timeout {
			util.PrintDebug("Periodic regeneration for %s", wiki.SourceDir)
		}

		// Update wiki
		if err = wiki.generate(ctx, result.Regen, clean, version); err != nil {
			return fmt.Errorf("failed to update %s wiki: %v", wiki.SourceDir, err)
		}
	}
}

// watchDirRecursive sets up watches on the specified directory and all subdirectories recursively.
func watchDirRecursive(ctx context.Context, path string, watcher *fsnotify.Watcher) error {
	err := watcher.Add(path)
	if err != nil {
		return fmt.Errorf("failed to watch directory '%s': '%s'", path, err)
	}

	err = filepath.Walk(path, func(subPath string, info os.FileInfo, err error) error {
		// Check for cancellation periodically during walk
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

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

// takeFilesSnapshot records the names and  modification times of all files and directories in dir recursively.
func takeFilesSnapshot(ctx context.Context, dir string) ([]fileSnapshot, error) {
	var snapshots []fileSnapshot

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		// Check for cancellation periodically during walk
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

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

// filesSnapshotsAreEqual compares two snapshots and returns true if they exist and are equal.
func filesSnapshotsAreEqual(snapshot1, snapshot2 []fileSnapshot) bool {
	if snapshot1 == nil || snapshot2 == nil {
		return false
	}
	return reflect.DeepEqual(snapshot1, snapshot2)
}
