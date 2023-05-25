package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/stalexan/gomarkwiki/internal/generator"
	"github.com/stalexan/gomarkwiki/internal/util"
)

// version holds the gomarkwiki version, and is set at build time.
var version string

func main() {
	// Parse command line
	args := parseCommandLine()

	// Start profiling.
	if args.cpuProfile != "" {
		util.PrintVerboseMessage("Starting profiler")
		var file *os.File
		var err error
		if file, err = os.Create(args.cpuProfile); err != nil {
			util.PrintFatalError(fmt.Sprintf("Failed to create file %s", args.cpuProfile), err)
		}
		defer file.Close()
		if err = pprof.StartCPUProfile(file); err != nil {
			util.PrintFatalError("Failed to start profiler", err)
		}
		defer func() {
			util.PrintVerboseMessage("Stopping profiler")
			pprof.StopCPUProfile()
		}()
	}

	// Generate wiki.
	if err := generator.GenerateWiki(args.dirs, args.regen, args.clean, version); err != nil {
		util.PrintFatalError("", err)
	}

	// Success.
	os.Exit(0)
}

// commandLineArgs stores the arguments specified on the command line.
type commandLineArgs struct {
	dirs       generator.WikiDirs
	cpuProfile string
	regen      bool
	clean      bool
}

// parseCommandLine parses the command line.
func parseCommandLine() commandLineArgs {
	// Define command line flags.
	printHelp := flag.Bool("help", false, "Show help")
	printVersion := flag.Bool("version", false, "Print version information")
	regen := flag.Bool("regen", false, "Regenerate all files regardless of timestamps")
	clean := flag.Bool("clean", false, "Delete any files in dest_dir that do not have a corresponding file in source_dir")
	flag.BoolVar(&util.Verbose, "verbose", false, "Print status messages")
	cpuProfile := flag.String("cpuprofile", "", "Write cpu profile to file")

	// Define custom usage message.
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] source_dir dest_dir\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "Options:")
		flag.PrintDefaults()
	}

	// Parse command line.
	flag.Parse()

	// Print help.
	if *printHelp {
		flag.Usage()
		os.Exit(0)
	}

	// Print version.
	if *printVersion {
		fmt.Printf("gomarkwiki %s compiled with %s on %s/%s\n", version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}

	// Were directories specified?
	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(1)
	}
	dirs := generator.NewWikiDirs(flag.Arg(0), flag.Arg(1))

	return commandLineArgs{
		dirs:       dirs,
		cpuProfile: *cpuProfile,
		regen:      *regen,
		clean:      *clean,
	}
}
