// Package wiki generates HTML from markdown for a given wiki.
package wiki

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// fileSnapshot records the name, modification time, and size for a given file or directory.
//
// IMPORTANT: All fields must be value types or immutable types (strings are OK,
// but no pointers, slices, maps, or channels). This struct is used in the Watcher's
// snapshot slice, which uses a "copy and release" pattern that shares the underlying
// array between copies. See the Thread Safety comment on the Watcher struct for details.
type fileSnapshot struct {
	name      string
	timestamp int64 // nanoseconds since Unix epoch
	size      int64 // file size in bytes (0 for directories)
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
	contentDir    string
	subsPath      string
	ignorePath    string
	sourceDir     string // For error messages
	ignoreMatcher *IgnoreMatcher

	// Watcher instance (reused across cycles)
	fsWatcher *fsnotify.Watcher
	mu        sync.Mutex // Protects fsWatcher, snapshot, subsModTime, ignoreModTime, and ignoreMatcher

	// Current state
	snapshot         []fileSnapshot
	subsModTime      int64 // Last known modification time of substitution strings file
	ignoreModTime    int64 // Last known modification time of ignore.txt file
	subsFileExists   bool
	ignoreFileExists bool

	// Context management
	ctx    context.Context
	cancel context.CancelFunc
}

// Thread Safety:
// The mutex (mu) protects concurrent access to fsWatcher, snapshot, and subsModTime.
// For snapshot reads, we use a "copy and release" pattern: lock, copy the slice header
// to a local variable, then unlock immediately. This minimizes lock contention by
// holding the lock only long enough to ensure atomic read of the slice header (pointer,
// length, capacity). The local copy can then be used safely without holding the lock,
// allowing other goroutines to update the snapshot concurrently. This pattern ensures:
//   - Atomic assignment/read of the slice header (all 3 fields together)
//   - Memory visibility (proper memory barriers for cross-goroutine visibility)
//   - No data races (compliance with Go's memory model)
// Note: The slice header copy shares the underlying array with the original, but since
// we only read from it and the elements are value types, this is safe.

// WatchResult represents the result of waiting for a change.
type WatchResult struct {
	Snapshot      []fileSnapshot // New snapshot after changes stabilized
	Regen         bool           // Whether full regeneration is needed
	IgnoreChanged bool           // Whether ignore.txt changed
	Timeout       bool           // Whether the wait timed out
}

// NewWatcher creates a new Watcher instance for the given directories.
// The parent context is used for cancellation - when it's cancelled, the watcher will stop.
func NewWatcher(parentCtx context.Context, contentDir, subsPath, ignorePath, sourceDir string, ignoreMatcher *IgnoreMatcher) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %v", err)
	}

	// Set up recursive watching
	if err := watchDirRecursive(parentCtx, contentDir, fsWatcher); err != nil {
		fsWatcher.Close()
		return nil, fmt.Errorf("failed to initialize watcher for %s: %v", contentDir, err)
	}

	// Watch source directory to monitor config file creation/deletion
	if err := fsWatcher.Add(sourceDir); err != nil {
		fsWatcher.Close()
		return nil, fmt.Errorf("failed to watch source directory '%s': %v", sourceDir, err)
	}

	// Watch substitution strings file if provided
	var subsModTime int64
	var subsExists bool
	if subsPath != "" {
		if info, err := os.Stat(subsPath); err == nil {
			subsModTime = info.ModTime().Unix()
			subsExists = true
			if err := fsWatcher.Add(subsPath); err != nil {
				fsWatcher.Close()
				return nil, fmt.Errorf("failed to watch '%s': %v", subsPath, err)
			}
		} else if !os.IsNotExist(err) {
			fsWatcher.Close()
			return nil, fmt.Errorf("failed to stat '%s': %v", subsPath, err)
		}
	}

	// Watch ignore.txt file if provided
	var ignoreModTime int64
	var ignoreExists bool
	if ignorePath != "" {
		if info, err := os.Stat(ignorePath); err == nil {
			ignoreModTime = info.ModTime().Unix()
			ignoreExists = true
			if err := fsWatcher.Add(ignorePath); err != nil {
				fsWatcher.Close()
				return nil, fmt.Errorf("failed to watch '%s': %v", ignorePath, err)
			}
		} else if !os.IsNotExist(err) {
			fsWatcher.Close()
			return nil, fmt.Errorf("failed to stat '%s': %v", ignorePath, err)
		}
	}

	// Create a child context that will be cancelled when the watcher is closed
	ctx, cancel := context.WithCancel(parentCtx)

	return &Watcher{
		contentDir:       contentDir,
		subsPath:         subsPath,
		ignorePath:       ignorePath,
		sourceDir:        sourceDir,
		ignoreMatcher:    ignoreMatcher,
		fsWatcher:        fsWatcher,
		subsModTime:      subsModTime,
		ignoreModTime:    ignoreModTime,
		subsFileExists:   subsExists,
		ignoreFileExists: ignoreExists,
		ctx:              ctx,
		cancel:           cancel,
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
	w.mu.Lock()
	snapshot := w.snapshot
	w.mu.Unlock()

	if snapshot != nil {
		newSnapshot, err := takeFilesSnapshot(ctx, w.contentDir, w.ignoreMatcher)
		if err != nil {
			return nil, fmt.Errorf("failed to take snapshot for %s: %v", w.contentDir, err)
		}
		if !filesSnapshotsAreEqual(snapshot, newSnapshot) {
			util.PrintVerbose("Files changed between generation cycles. Starting update.")

			// Check if substitution strings file or ignore.txt also changed
			subsChanged := w.checkSubsFileChanged()
			ignoreChanged := w.checkIgnoreFileChanged()
			regen := subsChanged || ignoreChanged

			// Files changed, wait for stability and return
			stableSnapshot, err := w.waitForStability(ctx)
			if err != nil {
				return nil, err
			}
			return &WatchResult{
				Snapshot:      stableSnapshot,
				Regen:         regen,
				IgnoreChanged: ignoreChanged,
				Timeout:       false,
			}, nil
		}

		// Even if content files didn't change, check if substitution strings file or ignore.txt changed
		subsChanged := w.checkSubsFileChanged()
		ignoreChanged := w.checkIgnoreFileChanged()
		if subsChanged || ignoreChanged {
			if subsChanged {
				util.PrintVerbose("Substitution strings file changed between generation cycles. Starting update.")
			}
			if ignoreChanged {
				util.PrintVerbose("Ignore expressions file changed between generation cycles. Starting update.")
			}
			w.mu.Lock()
			currentSnapshot := w.snapshot
			w.mu.Unlock()
			return &WatchResult{
				Snapshot:      currentSnapshot, // No content changes, so snapshot is still valid
				Regen:         true,
				IgnoreChanged: ignoreChanged,
				Timeout:       false,
			}, nil
		}
	}

	// Wait for a file system event (or timeout)
	regen, ignoreChanged, err := w.waitForEvent(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to watch for change event in %s: %v", w.sourceDir, err)
	}

	// Check if context expired (timeout for periodic regeneration)
	if ctx.Err() == context.DeadlineExceeded {
		util.PrintDebug("Regen timer expired for %s", w.sourceDir)

		// Take fresh snapshot so next cycle has up-to-date state.
		// Use parent context (w.ctx) since the timeout context already expired.
		freshSnapshot, err := takeFilesSnapshot(w.ctx, w.contentDir, w.ignoreMatcher)
		if err != nil {
			// Fall back to old snapshot if we can't take a fresh one
			util.PrintDebug("Failed to take fresh snapshot on timeout, using old snapshot: %v", err)
			w.mu.Lock()
			snapshot := w.snapshot
			w.mu.Unlock()
			return &WatchResult{
				Snapshot:      snapshot,
				Regen:         false,
				IgnoreChanged: false,
				Timeout:       true,
			}, nil
		}

		return &WatchResult{
			Snapshot:      freshSnapshot,
			Regen:         false,
			IgnoreChanged: false,
			Timeout:       true,
		}, nil
	}

	// Wait for changes to stabilize before regenerating
	stableSnapshot, err := w.waitForStability(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for changes to finish for %s: %v", w.sourceDir, err)
	}

	return &WatchResult{
		Snapshot:      stableSnapshot,
		Regen:         regen,
		IgnoreChanged: ignoreChanged,
		Timeout:       false,
	}, nil
}

// waitForEvent waits for a file system event from the watcher.
//
// State: Idle -> EventDetected
//
// This method blocks until:
//   - A file system event occurs (returns regen flag and ignoreChanged flag)
//   - The watcher encounters an error (returns error)
//   - The context times out (returns false, false, caller handles timeout)
func (w *Watcher) waitForEvent(ctx context.Context) (bool, bool, error) {
	// Check context first - if already cancelled (e.g., by Close()), return early
	// This prevents the race condition where Close() is called between
	// releasing the lock and entering the select statement
	select {
	case <-ctx.Done():
		return false, false, nil // Context cancelled, caller will handle
	default:
	}

	// Get references to channels while holding the lock to prevent
	// concurrent Close() from causing a race condition.
	//
	// Note on concurrent Close() safety: Even if Close() is called after we copy
	// the channel references but before the select below, this is safe because:
	// 1. Close() calls w.cancel(), triggering ctx.Done() in the select
	// 2. Close() closes the fsnotify channels, which we detect via ok == false
	// 3. When ok == false, we check ctx.Done() to distinguish graceful shutdown
	//    from unexpected channel closure, preventing spurious errors on Ctrl-C
	// 4. Reading from closed channels in Go returns immediately (no panic)
	// This multi-layered defense ensures graceful shutdown without races.
	w.mu.Lock()
	if w.fsWatcher == nil {
		w.mu.Unlock()
		return false, false, fmt.Errorf("watcher closed while waiting for event in '%s'", w.contentDir)
	}
	eventsChan := w.fsWatcher.Events
	errorsChan := w.fsWatcher.Errors
	w.mu.Unlock()

	select {
	case event, ok := <-eventsChan:
		if !ok {
			// Channel closed - check if this is due to graceful shutdown
			select {
			case <-ctx.Done():
				return false, false, nil // Graceful shutdown
			default:
				return false, false, fmt.Errorf("watcher unexpectedly closed while watching '%s'", w.contentDir)
			}
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

		// Check if substitution strings file or ignore.txt changed
		subsChanged := eventMatchesConfigFile(event, w.subsPath)
		if subsChanged {
			util.PrintVerbose("Substitution strings file '%s' write seen", w.subsPath)
		}

		ignoreChanged := eventMatchesConfigFile(event, w.ignorePath)
		if ignoreChanged {
			util.PrintVerbose("Ignore expressions file '%s' write seen", w.ignorePath)
		}

		return subsChanged || ignoreChanged, ignoreChanged, nil

	case err, ok := <-errorsChan:
		if !ok {
			// Channel closed - check if this is due to graceful shutdown
			select {
			case <-ctx.Done():
				return false, false, nil // Graceful shutdown
			default:
				return false, false, fmt.Errorf("failed to read watcher error for %s", w.contentDir)
			}
		}
		return false, false, fmt.Errorf("watcher error for %s: %v", w.contentDir, err)

	case <-ctx.Done():
		return false, false, nil // Timeout, caller will handle
	}
}

func eventMatchesConfigFile(event fsnotify.Event, path string) bool {
	if path == "" {
		return false
	}

	if filepath.Clean(event.Name) != path {
		return false
	}

	return event.Has(fsnotify.Create) ||
		event.Has(fsnotify.Write) ||
		event.Has(fsnotify.Remove) ||
		event.Has(fsnotify.Rename) ||
		event.Has(fsnotify.Chmod)
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
			w.mu.Lock()
			snapshot := w.snapshot
			w.mu.Unlock()
			resultChan <- result{snapshot: snapshot}
			return
		case <-time.After(CHANGE_WAIT):
		}

		var snapshot1, snapshot2 []fileSnapshot
		var err error

		// Loop until snapshots match (indicating stability)
		// Note: While this loop has no explicit pass limit, it's bounded by the parent
		// context timeout (MAX_REGEN_INTERVAL = 10 minutes), which ensures a clean exit
		// even if files never stabilize (e.g., continuous filesystem modifications).
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
					w.mu.Lock()
					snapshot := w.snapshot
					w.mu.Unlock()
					resultChan <- result{snapshot: snapshot}
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
				snapshot1, err = takeFilesSnapshot(ctx, w.contentDir, w.ignoreMatcher)
				if err != nil {
					resultChan <- result{err: fmt.Errorf("failed to take files snapshot for %s: %v", w.contentDir, err)}
					return
				}
			}

			// Calculate wait time (exponential backoff: 2^n)
			// MAX_CHANGE_WAIT caps the wait time to prevent excessive delays.
			// For the current values, this kicks in at waitPass >= 7.
			shift := waitPass - 1
			waitTime := CHANGE_WAIT * (1 << shift)
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
			snapshot2, err = takeFilesSnapshot(ctx, w.contentDir, w.ignoreMatcher)
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
		if os.IsNotExist(err) {
			w.mu.Lock()
			changed := w.subsFileExists
			w.subsFileExists = false
			w.subsModTime = 0
			w.mu.Unlock()
			return changed
		}
		util.PrintWarning("Failed to stat substitution strings file '%s': %v", w.subsPath, err)
		return false
	}

	currentModTime := info.ModTime().Unix()
	w.mu.Lock()
	changed := !w.subsFileExists || currentModTime != w.subsModTime
	w.subsFileExists = true
	w.subsModTime = currentModTime
	w.mu.Unlock()
	return changed
}

// checkIgnoreFileChanged checks if the ignore.txt file has been modified
// since it was last recorded. Returns true if the file changed, and updates
// the stored modification time.
func (w *Watcher) checkIgnoreFileChanged() bool {
	if w.ignorePath == "" {
		return false
	}

	info, err := os.Stat(w.ignorePath)
	if err != nil {
		if os.IsNotExist(err) {
			w.mu.Lock()
			changed := w.ignoreFileExists
			w.ignoreFileExists = false
			w.ignoreModTime = 0
			w.mu.Unlock()
			return changed
		}
		util.PrintWarning("Failed to stat ignore expressions file '%s': %v", w.ignorePath, err)
		return false
	}

	currentModTime := info.ModTime().Unix()
	w.mu.Lock()
	changed := !w.ignoreFileExists || currentModTime != w.ignoreModTime
	w.ignoreFileExists = true
	w.ignoreModTime = currentModTime
	w.mu.Unlock()
	return changed
}

// UpdateSnapshot updates the internal snapshot state.
// This should be called after successful generation.
func (w *Watcher) UpdateSnapshot(snapshot []fileSnapshot) {
	w.mu.Lock()
	w.snapshot = snapshot
	w.mu.Unlock()
	// Also update substitution strings file and ignore.txt mod times when snapshot is updated
	w.checkSubsFileChanged()
	w.checkIgnoreFileChanged()
}

// UpdateIgnoreMatcher updates the ignore matcher used for snapshot comparisons.
// This should be called when ignore.txt is reloaded.
func (w *Watcher) UpdateIgnoreMatcher(matcher *IgnoreMatcher) {
	w.mu.Lock()
	w.ignoreMatcher = matcher
	w.mu.Unlock()
}

// GetSnapshot returns the current snapshot.
func (w *Watcher) GetSnapshot() []fileSnapshot {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.snapshot
}

// watch watches for changes in the wiki content directory and regenerates files on the fly.
func (wiki *Wiki) watch(ctx context.Context, clean bool, version string) error {
	util.PrintVerbose("Watching for changes in '%s'", wiki.ContentDir)

	// Create watcher with parent context
	watcher, err := NewWatcher(ctx, wiki.ContentDir, wiki.subsPath, wiki.ignorePath, wiki.SourceDir, wiki.ignoreMatcher)
	if err != nil {
		return fmt.Errorf("failed to create watcher: %v", err)
	}
	defer watcher.Close()

	// Take initial snapshot
	initialSnapshot, err := takeFilesSnapshot(ctx, wiki.ContentDir, wiki.ignoreMatcher)
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
		// Note: UpdateSnapshot also updates subsModTime and ignoreModTime if files changed

		// Reload substitution strings if needed
		if result.Regen {
			util.PrintVerbose("Reloading substitution strings from '%s'", wiki.subsPath)
			if err := wiki.loadSubstitutionStrings(); err != nil {
				return fmt.Errorf("failed to reload substitution strings file: %v", err)
			}
			// Mod time already updated by UpdateSnapshot above
		}

		// Reload ignore expressions if needed
		if result.IgnoreChanged {
			util.PrintVerbose("Reloading ignore expressions from '%s'", wiki.ignorePath)
			if err := wiki.loadIgnoreExpressions(); err != nil {
				return fmt.Errorf("failed to reload ignore expressions file: %v", err)
			}
			// Update watcher's ignore matcher with the new one
			watcher.UpdateIgnoreMatcher(wiki.ignoreMatcher)
			// Mod time already updated by UpdateSnapshot above
		}

		// Skip generation if this was just a timeout (periodic regen)
		if result.Timeout {
			util.PrintDebug("Periodic regeneration for %s", wiki.SourceDir)
		}

		// Update wiki
		if err = wiki.generate(ctx, result.Regen, clean, version); err != nil {
			// In watch mode, log the error but continue watching
			util.PrintError(err, "failed to update %s wiki", wiki.SourceDir)
			// Continue the loop instead of returning
			continue
		}
	}
}

// watchDirRecursive sets up watches on the specified directory and all subdirectories recursively.
func watchDirRecursive(ctx context.Context, path string, watcher *fsnotify.Watcher) error {
	err := watcher.Add(path)
	if err != nil {
		return fmt.Errorf("failed to watch directory '%s': '%s'", path, err)
	}

	baseDepth := strings.Count(path, string(filepath.Separator))
	err = filepath.Walk(path, func(subPath string, info os.FileInfo, err error) error {
		// Check for cancellation periodically during walk
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Check recursion depth
		currentDepth := strings.Count(subPath, string(filepath.Separator)) - baseDepth
		if currentDepth > MaxRecursionDepth {
			return fmt.Errorf("directory recursion depth exceeded at '%s' (depth %d, max %d)", subPath, currentDepth, MaxRecursionDepth)
		}

		if err != nil {
			// If we can't access a path during setup, warn and skip it but continue setup
			util.PrintWarning("Failed to access '%s' during watch setup: %v", subPath, err)
			if info != nil && info.IsDir() {
				return filepath.SkipDir
			}
			return nil
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
// Files matching the ignore patterns are excluded from the snapshot.
func takeFilesSnapshot(ctx context.Context, dir string, ignoreMatcher *IgnoreMatcher) ([]fileSnapshot, error) {
	var snapshots []fileSnapshot
	baseDepth := strings.Count(dir, string(filepath.Separator))

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		// Check for cancellation periodically during walk
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Check recursion depth
		currentDepth := strings.Count(path, string(filepath.Separator)) - baseDepth
		if currentDepth > MaxRecursionDepth {
			return fmt.Errorf("directory recursion depth exceeded at '%s' (depth %d, max %d)", path, currentDepth, MaxRecursionDepth)
		}

		if err != nil {
			util.PrintWarning("Failed to access '%s' during snapshot: %v", path, err)
			if info != nil && info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if this file should be ignored
		if ignoreMatcher != nil {
			relPath, err := filepath.Rel(dir, path)
			if err == nil && ignoreMatcher.Matches(relPath, info.IsDir()) {
				util.PrintDebug("Ignoring '%s' in snapshot", path)
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Look up file size. (Directory size is filesystem-dependent and meaningless for change detection
		size := info.Size()
		if info.IsDir() {
			size = 0
		}

		snapshot := fileSnapshot{
			name:      path,
			timestamp: info.ModTime().UnixNano(),
			size:      size,
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

	// Quick length check
	if len(snapshot1) != len(snapshot2) {
		return false
	}

	// Compare each snapshot
	for i := range snapshot1 {
		s1, s2 := &snapshot1[i], &snapshot2[i]
		if s1.name != s2.name ||
			s1.timestamp != s2.timestamp ||
			s1.size != s2.size ||
			s1.isDir != s2.isDir {
			return false
		}
	}

	return true
}
