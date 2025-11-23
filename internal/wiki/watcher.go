// Package wiki generates HTML from markdown for a given wiki.
package wiki

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/stalexan/gomarkwiki/internal/util"
)

// Times to wait while waiting for changes to finish.
const CHANGE_WAIT = 100      // milliseconds
const MAX_CHANGE_WAIT = 5000 // milliseconds

// Max time before regenerating wiki while watching for changes.
const MAX_REGEN_INTERVAL = 10 // minutes

// fileSnapshot records the name and modification time for a given file or directory.
type fileSnapshot struct {
	name      string
	timestamp int64
	isDir     bool
}

// watch watches for changes in the wiki content directory and regenerates files on the fly.
func (wiki *Wiki) watch(clean bool, version string) error {
	util.PrintVerbose("Watching for changes in '%s'", wiki.ContentDir)

	var snapshot []fileSnapshot // Latest snapshot
	for {
		// Wait for when wiki needs to be updated.
		var err error
		var regen bool
		if snapshot, regen, err = wiki.waitForWhenGenerateNeeded(clean, version, snapshot); err != nil {
			return fmt.Errorf("failed waiting to update %s wiki: %v", wiki.SourceDir, err)
		}

		// Reload substition strings if the substition strings file has changed.
		if regen {
			util.PrintVerbose("Reloading substitution strings from '%s'", wiki.subsPath)
			if err := wiki.loadSubstitutionStrings(); err != nil {
				return fmt.Errorf("failed to reload substition strings file: %v", err)
			}
		}

		// Update wiki.
		if err = wiki.generate(regen, clean, version); err != nil {
			return fmt.Errorf("failed to update %s wiki: %v", wiki.SourceDir, err)
		}
	}
}

func (wiki *Wiki) waitForWhenGenerateNeeded(clean bool, version string, snapshot []fileSnapshot) ([]fileSnapshot, bool, error) {
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

	// Define the result struct
	type result struct {
		snapshot []fileSnapshot
		regen    bool
	}

	// Create a chan to return snapshot and regen value.
	resultChan := make(chan result, 1)
	defer func() {
		close(resultChan)
	}()

	go func() {
		// Say when this goroutine is done.
		defer func() {
			doneChan <- struct{}{}
		}()

		// Watch for changes.
		var err error
		var regen bool
		if regen, err = watchForChangeEvent(ctx, wiki.ContentDir, wiki.subsPath, clean, version, snapshot); err != nil {
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
		resultChan <- result{newSnapshot, regen}
	}()

	// Wait on results.
	var err error
	var res result
	select {
	case res = <-resultChan:
		snapshot = res.snapshot
	case err = <-errorChan:
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			util.PrintDebug("Regen timer expired for %s", wiki.SourceDir)
		}
	}

	// Wait for goroutine to finish, so that the chan it writes to isn't closed before any final writes.
	<-doneChan

	return snapshot, res.regen, err
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

// takeFilesSnapshot records the names and  modification times of all files and directories in dir recursively.
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

// filesSnapshotsAreEqual compares two snapshots and returns true if they exist and are equal.
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
func watchForChangeEvent(ctx context.Context, contentDir string, subsPath string, clean bool, version string, snapshot []fileSnapshot) (bool, error) {
	// Create and initialize watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return false, fmt.Errorf("failed to create watcher for '%s': %v", contentDir, err)
	}
	defer watcher.Close()
	if err = watchDirRecursive(contentDir, watcher); err != nil {
		return false, fmt.Errorf("failed to initialize watcher for %s: %v", contentDir, err)
	}
	if subsPath != "" {
		// Watch substitution-strings.csv too.
		err := watcher.Add(subsPath)
		if err != nil {
			return false, fmt.Errorf("failed to watch '%s': '%s'", subsPath, err)
		}
	}

	// Make sure files haven't changed in between when wiki update started and new watch started.
	if snapshot != nil {
		newSnapshot, err := takeFilesSnapshot(contentDir)
		if err != nil {
			return false, fmt.Errorf("failed to take new snapshot for %s: %v", contentDir, err)
		}
		if !filesSnapshotsAreEqual(snapshot, newSnapshot) {
			// Files have changed. Start a new update.
			util.PrintVerbose("About to watch for changes but file snapshots differ. Starting a new update.")
			return false, nil
		}
	}

	// Watch for changes.
	select {
	case event, ok := <-watcher.Events:
		if !ok {
			return false, fmt.Errorf("watcher unexpectedly closed while watching '%s' %v", contentDir, err)
		}
		util.PrintDebug("Watcher event detected for %s: %v", contentDir, event)

		// Regenerate all files if the substitution strings file has changed.
		regen := subsPath != "" &&
			(event.Has(fsnotify.Write) || event.Has(fsnotify.Rename)) &&
			filepath.Clean(event.Name) == subsPath
		if regen {
			util.PrintVerbose("Substitution strings file '%s' write seen", subsPath)
		}

		return regen, nil
	case err, ok := <-watcher.Errors:
		if !ok {
			return false, fmt.Errorf("failed to read watcher error for %s", contentDir)
		}
		return false, fmt.Errorf("watcher error for %s: %v", contentDir, err)
	case <-ctx.Done():
		// Regen timer has expired.
		return false, nil
	}
}
