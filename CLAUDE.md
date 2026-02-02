# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Gomarkwiki is a Go CLI that converts Markdown to HTML for static site generation. It uses Goldmark for CommonMark/GFM parsing and fsnotify for file watching. The project intentionally minimizes dependencies (3 packages total) for security and auditability.

## Build & Development Commands

```bash
make build          # Build binary to ./build/gomarkwiki
make test           # Run all tests with verbose output
make fmt            # Format code with go fmt
make vet            # Run go vet
make staticcheck    # Run staticcheck (install: go install honnef.co/go/tools/cmd/staticcheck@latest)
make clean          # Remove build/ and tmp/ directories
```

Run a single test:
```bash
go test -v -run TestName ./internal/wiki
```

Run tests with race detection:
```bash
go test -race ./...
```

## Architecture

### Directory Structure
- `cmd/main.go` - CLI entry point, worker pool for multi-wiki generation, signal handling
- `internal/wiki/` - Core wiki functionality
  - `wiki.go` - Wiki struct, initialization, public Generate()/Watch() methods
  - `generator.go` - Markdown→HTML conversion using Goldmark, CSS handling
  - `filesystem.go` - File I/O, atomic writes, symlink handling (follows file symlinks, skips directory symlinks)
  - `watcher.go` - File watching state machine with exponential backoff
  - `config.go` - Configuration loading (substitution-strings.csv, ignore.txt)
  - `ignorepattern.go` - Gitignore-style pattern matching
- `internal/util/` - Utility functions (output formatting, CSV parsing)

### Concurrency Model

Hierarchical goroutine structure (see docs/CONCURRENCY.md for details):

1. **Multi-wiki parallelism**: One goroutine per wiki using worker pool pattern
2. **File watching**: Single fsnotify.Watcher per wiki, reused across watch cycles
3. **Coordination**: Context-based cancellation, buffered error channels, sync.WaitGroup

Key patterns:
- Copy-and-release for snapshot reads (minimizes lock contention)
- Exponential backoff for file change stability (100ms → 5000ms)
- 10-minute periodic regeneration timeout

### Source Directory Convention

When gomarkwiki processes a wiki, it expects:
```
source_dir/
├── content/                    # Required - markdown files here
│   ├── index.md
│   ├── local.css              # Optional - override default styles
│   └── favicon.ico            # Optional
├── substitution-strings.csv   # Optional - {{PLACEHOLDER}} → value
└── ignore.txt                 # Optional - gitignore-style patterns
```

### Resource Limits

Defined in internal/wiki/config.go and generator.go:
- MaxMarkdownFileSize: 100 MB
- MaxFilesProcessed: 1,000,000 per generation
- MaxRecursionDepth: 1,000 directory levels
- MaxCSVFileSize: 10 MB
- MaxSubstitutionStrings: 10,000 pairs

## Testing

Tests create a temporary `./tmp` directory. Test fixtures are in `internal/wiki/testdata/`.

The watcher tests (`watcher_robustness_test.go`) stress-test the file watching state machine.
