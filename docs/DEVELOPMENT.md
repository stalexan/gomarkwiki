# Development Guide

This document describes how to set up and work on the gomarkwiki project.

## Prerequisites

- **Go 1.22 or later** - Required by the project (as specified in `go.mod`)
- **Make** - For running build tasks (optional, you can use `go` commands directly)

## Project Structure

```
gomarkwiki/
├── cmd/
│   └── main.go          # Main entry point
├── internal/
│   ├── wiki/            # Core wiki functionality (contains tests)
│   └── util/            # Utility functions
├── release-builder/     # Release build scripts
├── docs/                # Documentation
├── go.mod               # Go module dependencies
├── Makefile             # Build automation
└── README.md            # Project overview
```

## Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/stalexan/gomarkwiki
cd gomarkwiki
```

### 2. Install Dependencies

The project uses Go modules, so dependencies will be automatically downloaded when you build:

```bash
go mod download
```

## Development Workflow

### Building the Application

Build the binary using Make:

```bash
make build
```

This will:
- Create a `build/` directory
- Compile the binary with version information
- Place the executable at `./build/gomarkwiki`

You can also build directly with Go:

```bash
go build -ldflags "-X 'main.version=$(git describe --tags)'" -o build/gomarkwiki ./cmd/main.go
```

### Running Tests

Run the test suite:

```bash
make test
```

This runs tests in `./internal/wiki` with verbose output.

To run tests directly with Go:

```bash
go test -v ./internal/wiki
```

### Code Formatting

Format all Go code according to Go standards:

```bash
make fmt
```

This runs `go fmt ./...` to format all Go files in the project.

### Code Quality Checks

#### Go Vet

Run `go vet` to check for common errors:

```bash
make vet
```

#### Staticcheck

Run staticcheck for additional static analysis (requires staticcheck tool):

```bash
make staticcheck
```

If you don't have staticcheck installed, you can install it with:

```bash
go install honnef.co/go/tools/cmd/staticcheck@latest
```

### Cleaning Build Artifacts

Remove build directories:

```bash
make clean
```

This removes both the `build/` and `tmp/` directories.

### Running the Application

After building, you can run the binary:

```bash
./build/gomarkwiki [options] source_dir dest_dir
```

Example usage:

```bash
# Generate HTML for a markdown site
./build/gomarkwiki ~/example-site ~/wikis-html/example-site

# Watch for changes and regenerate automatically
./build/gomarkwiki -watch ~/example-site ~/wikis-html/example-site

# Generate multiple wikis from a CSV file
./build/gomarkwiki -wikis /etc/gomarkwiki/wikis.csv
```

## Development Cycle

A typical development cycle looks like this:

1. **Make code changes** - Edit files in `cmd/`, `internal/`, etc.
2. **Format code** - Run `make fmt` to ensure consistent formatting
3. **Check code quality** - Run `make vet` to catch common issues
4. **Run tests** - Run `make test` to verify tests pass
5. **Build** - Run `make build` to create the binary
6. **Test manually** - Run the binary with your markdown files to verify functionality

## Available Make Targets

Run `make help` to see all available targets:

- `build` - Build the binary
- `clean` - Clean the build directory
- `fmt` - Format the code
- `staticcheck` - Check the code using staticcheck
- `test` - Run the tests
- `vet` - Check the code using vet
- `help` - Print help message

## Testing Your Changes

1. **Unit Tests**: Run `make test` to ensure existing tests pass
2. **Manual Testing**: Build and run the binary with sample markdown files
3. **Integration Testing**: Test with real wiki directories to ensure end-to-end functionality

## Version Information

The build process automatically includes version information from git tags. The
version is embedded in the binary using the `-ldflags` flag:

```bash
-ldflags "-X 'main.version=$(VERSION)'"
```

Where `VERSION` is determined by `git describe --tags`.

## Dependencies

The project uses the following main dependencies (see `go.mod`):

- `github.com/yuin/goldmark` - Markdown parser
- `github.com/fsnotify/fsnotify` - File system watching (for `-watch` mode)

Dependencies are managed through Go modules and will be automatically downloaded when building.

## Contributing

When contributing to the project:

1. Follow Go code style conventions
2. Run `make fmt` before committing
3. Ensure all tests pass with `make test`
4. Run code quality checks (`make vet`, `make staticcheck`)
5. Add tests for new functionality
6. Update documentation as needed

## Troubleshooting

### Staticcheck not found

Install staticcheck:
```bash
go install honnef.co/go/tools/cmd/staticcheck@latest
```

Make sure `$GOPATH/bin` or `$HOME/go/bin` is in your `PATH`.

