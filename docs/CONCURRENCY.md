# Concurrency Architecture

This document describes the concurrency model and patterns used in gomarkwiki.

## Overview

The application uses **hierarchical concurrency** with multiple levels:

1. **Top Level**: Multi-wiki parallelism (one goroutine per wiki)
2. **Mid Level**: File system watching (one watcher per wiki in watch mode)
3. **Coordination**: Context-based cancellation and graceful shutdown

## Level 1: Multi-Wiki Parallelism

### Worker Pool Pattern

The main entry point (`cmd/main.go`) processes multiple wikis in parallel using a worker pool pattern:

```go
func generateWikis(wikis []*wiki.Wiki, ...) error {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    errorChan := make(chan error, len(wikis)*2)
    var wg sync.WaitGroup
    
    // Launch one goroutine per wiki
    wg.Add(len(wikis))
    for _, wiki := range wikis {
        go worker(wiki)
    }
    
    // Wait for completion or error
    wg.Wait()
}
```

### Key Features

- **One goroutine per wiki** - Each wiki is processed independently
- **Buffered error channel** - Collects errors without blocking workers
- **WaitGroup synchronization** - Ensures all workers complete before exit
- **Context-based cancellation** - Allows coordinated shutdown across all workers
- **Signal handling** - Graceful shutdown on SIGINT/SIGTERM

### Error Handling

Errors from any worker are:
1. Sent to a buffered channel
2. Trigger cancellation of all other workers
3. Collected and aggregated into a single error report
4. Returned after all workers have cleaned up

## Level 2: File System Watching

### Watcher Lifecycle

When running with the `-watch` flag, each wiki creates a single `Watcher` instance:

```go
func (wiki *Wiki) watch(ctx context.Context, ...) error {
    // Create watcher (one per wiki)
    watcher, err := NewWatcher(ctx, wiki.ContentDir, ...)
    if err != nil {
        return err
    }
    defer watcher.Close()
    
    // Main watch loop
    for {
        result, err := watcher.WaitForChange()
        // ... regenerate wiki ...
    }
}
```

### Watcher Architecture

Each `Watcher` instance:
- **Single fsnotify.Watcher** - Reused across all watch cycles for efficiency
- **State machine** - Manages change detection and stability waiting
- **Snapshot tracking** - Detects which files changed
- **Exponential backoff** - Waits for rapid file operations to stabilize

### State Machine

The watcher implements a state machine for change detection:

1. **Idle** - Listening for file system events
2. **EventDetected** - A change was detected
3. **WaitingForStability** - Comparing snapshots to ensure changes are complete
4. **Ready** - Changes stabilized, ready for regeneration
5. **Timeout** - Periodic regeneration timer expired (every 10 minutes)

## Synchronization Primitives

### sync.WaitGroup

Used to track the lifecycle of wiki worker goroutines:
- `wg.Add(len(wikis))` - Register all workers
- `defer wg.Done()` - Mark worker as completed
- `wg.Wait()` - Block until all workers finish

### sync.Mutex

Used in `Watcher` to protect shared state:
- `fsWatcher` - The underlying fsnotify watcher instance
- `snapshot` - File snapshots for change detection
- `subsModTime` / `ignoreModTime` - Configuration file modification times

**Note**: The mutex is primarily defensive, protecting against race conditions during shutdown.

### Channels

#### Error Channel
```go
errorChan := make(chan error, len(wikis)*2)  // Buffered
```
- Buffered to prevent worker blocking
- Collects errors from all workers
- Non-blocking send pattern with context check

#### Signal Channel
```go
termChan := make(chan os.Signal, 1)
signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)
```
- Receives OS signals for graceful shutdown
- Buffered to prevent signal loss

#### fsnotify Channels
- `fsWatcher.Events` - File system events
- `fsWatcher.Errors` - Watcher errors
- Read with proper shutdown handling

### context.Context

Contexts form a hierarchy for cancellation propagation:
```
context.Background()
    └─> Main context (generateWikis)
        └─> Wiki worker contexts
            └─> Watcher contexts
                └─> Timeout contexts (per wait cycle)
```

Cancellation flows down the hierarchy, allowing coordinated shutdown.

## Thread Safety Patterns

### Copy-and-Release Pattern

The `Watcher` uses a "copy and release" pattern for snapshot reads:

```go
w.mu.Lock()
snapshot := w.snapshot  // Copy slice header (pointer, len, cap)
w.mu.Unlock()

// Use snapshot without holding lock
// Safe because elements are value types
```

**Why this works**:
- Slice header (pointer, length, capacity) is copied atomically
- Underlying array is shared but only read
- `fileSnapshot` elements are value types (no pointers)
- Minimizes lock contention

### Concurrent Close Safety

The watcher protects against race conditions during shutdown:

```go
// In waitForEvent:
w.mu.Lock()
if w.fsWatcher == nil {
    w.mu.Unlock()
    return error  // Already closed
}
eventsChan := w.fsWatcher.Events  // Copy channel reference
w.mu.Unlock()

select {
case event := <-eventsChan:  // Use copied reference
    // ...
case <-ctx.Done():  // Context cancellation also signals shutdown
    // ...
}
```

**Multi-layered defense**:
1. Check if watcher is already closed
2. Copy channel references under lock
3. Context cancellation signals shutdown
4. Closed channels return immediately (no panic)

## Goroutine Access Patterns

### Is Watcher Single-Threaded?

**Almost, but not quite.**

- **Normal operation**: Only the wiki's watch loop goroutine accesses the Watcher
- **Shutdown window**: Both the watch loop AND cleanup code may access it simultaneously
- **Mutex purpose**: Ensures safe access during this brief shutdown race window

The mutex exists primarily for defensive programming and clean shutdown, not for ongoing concurrent access during normal operation.

## Graceful Shutdown

### Shutdown Sequence

When shutdown is triggered (error, signal, or completion):

1. **Cancel context** - Signals all goroutines to stop
2. **Wait for workers** - `wg.Wait()` blocks until all workers finish
3. **Collect errors** - Drain error channel
4. **Return results** - Aggregate and return errors

### Cleanup Order

Resources are cleaned up in reverse order of creation:
1. Timeout contexts (via defer in each wait cycle)
2. Watcher contexts (cancelled by Watcher.Close)
3. Watcher fsnotify instances (closed by Watcher.Close)
4. Main context (cancelled by generateWikis)

All cleanup uses `defer` to ensure execution even on panic.

## Example: Three Wikis in Watch Mode

```
Main Goroutine
├── Worker Goroutine 1 (Wiki A)
│   └── Watcher 1 (fsnotify for Wiki A content)
├── Worker Goroutine 2 (Wiki B)
│   └── Watcher 2 (fsnotify for Wiki B content)
└── Worker Goroutine 3 (Wiki C)
    └── Watcher 3 (fsnotify for Wiki C content)
```

**Key properties**:
- All three wikis generate in parallel
- Each has its own file watcher
- Errors in any wiki trigger shutdown of all
- Signal handling shuts down all gracefully

## Performance Considerations

### Parallelism Benefits

- Multiple wikis process simultaneously
- File watching doesn't block generation
- I/O operations can overlap across wikis

### Lock Contention

Lock contention is minimal because:
- Each `Watcher` has its own mutex
- Wikis don't share state
- Copy-and-release pattern minimizes lock hold time
- Locks are only held during brief critical sections

### Buffered Channels

Error channel is buffered (`len(wikis)*2`) to:
- Prevent workers from blocking on error send
- Allow workers to continue cleanup
- Collect all errors before returning

## Testing Concurrency

When testing concurrent behavior, consider:

1. **Signal handling** - Test SIGINT/SIGTERM during generation
2. **Error propagation** - Ensure one wiki's error stops others
3. **Resource cleanup** - Verify no goroutine leaks
4. **Race conditions** - Run with `-race` flag
5. **File watching** - Test rapid file changes and bulk operations

## Summary

The concurrency model is:
- **Simple**: Clear hierarchy of goroutines and contexts
- **Safe**: Proper synchronization with mutexes and channels
- **Robust**: Graceful shutdown handles all scenarios
- **Efficient**: Minimal lock contention and good parallelism

This design allows the application to process multiple wikis efficiently while
maintaining clean shutdown semantics and proper error handling.

