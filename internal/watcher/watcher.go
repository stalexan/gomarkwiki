// Package watcher watches for changes and regenerates files on the fly.
package watcher

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/stalexan/gomarkwiki/internal/generator"
	"github.com/stalexan/gomarkwiki/internal/util"
)

const CHANGE_WAIT = 100 // milliseconds
const MAX_WAIT = 5000   // milliseconds

// watchDirRecursive sets up watches on the specified directory and all subdirectories recursively.
func watchDirRecursive(path string, watcher *fsnotify.Watcher) error {
	err := watcher.Add(path)
	if err != nil {
		return fmt.Errorf("failed to watch directory %s: %s", path, err)
	}

	err = filepath.Walk(path, func(subPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			err = watcher.Add(subPath)
			if err != nil {
				return fmt.Errorf("failed to watch subdirectory %s: %s", subPath, err)
			}
		}
		return nil
	})

	return err
}

// FileSnapshot records the name the modification time for a given file or directory.
type FileSnapshot struct {
	name      string
	timestamp int64
	isDir     bool
}

// TakeFilesSnapshot records the names and  modification times of all files and directories in dir recursively.
func TakeFilesSnapshot(dir string) ([]FileSnapshot, error) {
	var snapshots []FileSnapshot

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		snapshot := FileSnapshot{
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

// FilesSnapshotsAreEqual compares two snapshots and returns true if they are equal.
func FilesSnapshotsAreEqual(snapshot1, snapshot2 []FileSnapshot) bool {
	if snapshot1 == nil || snapshot2 == nil {
		return false
	}
	return reflect.DeepEqual(snapshot1, snapshot2)
}

// doWatchPhase watches for wiki content changes.
func doWatchPhase(phaseId int, dirs generator.WikiDirs, clean bool, version string, interruptChan chan os.Signal, snapshot []FileSnapshot) (bool, error) {
	// Create and initialize watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return false, fmt.Errorf("failed to create watcher for %s: %v", dirs.ContentDir, err)
	}
	defer watcher.Close()
	if err = watchDirRecursive(dirs.ContentDir, watcher); err != nil {
		return false, fmt.Errorf("failed to initialize watcher: %v", err)
	}

	// Make sure files haven't changed in between when wiki update started and new watch started.
	if snapshot != nil {
		newSnapshot, err := TakeFilesSnapshot(dirs.ContentDir)
		if err != nil {
			return false, fmt.Errorf("failed to take new snapshot: %v", err)
		}
		if !FilesSnapshotsAreEqual(snapshot, newSnapshot) {
			// Start a new update.
			return false, nil
		}
	}

	// Watch for changes.
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return false, fmt.Errorf("watcher unexpectedly closed while watching %s: %v", dirs.ContentDir, err)
			}
			util.PrintDebug("Watcher event detected: %v", event)
			return false, nil
		case <-interruptChan:
			fmt.Println("Interrupt signal received. Exiting...")
			return true, nil
		case err, ok := <-watcher.Errors:
			if !ok {
				return false, errors.New("failed to read watcher error")
			}
			return false, fmt.Errorf("watcher error: %v", err)
		}
	}
}

// waitOnChangesToFinish waits for changes in dir to finish.
func waitOnChangesToFinish(dir string, interruptChan chan os.Signal) (bool, []FileSnapshot, error) {
	// Initial wait.
	time.Sleep(CHANGE_WAIT * time.Millisecond)

	// Create channels.
	snapshotsMatchChan := make(chan []FileSnapshot, 1) // Signals that changes are complete.
	errorChan := make(chan error, 1)                   // Signals that an error happened while waiting.
	defer func() {
		close(snapshotsMatchChan)
		close(errorChan)
	}()

	// Wait for changes to complete.
	go func() {
		var snapshot1, snapshot2 []FileSnapshot
		var err error
		for waitPass := 1; !FilesSnapshotsAreEqual(snapshot1, snapshot2); waitPass++ {
			util.PrintDebug("Wait pass %d", waitPass)

			// Take a before snapshot.
			if snapshot2 != nil {
				snapshot1 = snapshot2
			} else {
				snapshot1, err = TakeFilesSnapshot(dir)
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
			if snapshot2, err = TakeFilesSnapshot(dir); err != nil {
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
			return false, snapshot, nil
		case err := <-errorChan:
			return false, nil, err
		case <-interruptChan:
			fmt.Println("Interrupt signal received. Exiting...")
			return true, nil, nil
		}
	}
}

// Watch watches for changes in the wiki content directory and regenerates files on the fly.
func Watch(dirs generator.WikiDirs, clean bool, version string) error {
	util.PrintVerbose("Watching for changes in %s", dirs.ContentDir)

	// Create a channel to listen for the interrupt signal
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGTERM)
	defer close(interruptChan)

	var snapshot []FileSnapshot // Latest snapshot
	for phaseId := 1; ; phaseId++ {
		// Watch for changes.
		exit, err := doWatchPhase(phaseId, dirs, clean, version, interruptChan, snapshot)
		if err != nil {
			return fmt.Errorf("watch phase %d failed: %v", phaseId, err)
		}
		if exit {
			return nil
		}

		// Wait for changes to finish.
		exit, snapshot, err = waitOnChangesToFinish(dirs.ContentDir, interruptChan)
		if err != nil {
			return fmt.Errorf("wait on changes phase %d failed: %v", phaseId, err)
		}
		if exit {
			return nil
		}

		// Update wiki.
		if err := generator.GenerateWiki(dirs, false, clean, version); err != nil {
			return fmt.Errorf("failed to update wiki: %v", err)
		}
	}
}
