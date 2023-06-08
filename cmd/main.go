package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"

	"github.com/stalexan/gomarkwiki/internal/util"
	"github.com/stalexan/gomarkwiki/internal/wiki"
)

// version holds the gomarkwiki version, and is set at build time.
var version string

// Usage message
const usagePart1 = `Usage: gomarkwiki [options] source_dir dest_dir
       gomarkwiki [options] -wikis wikis_file

Options:`

const usagePart2 = `
Description:
  gomarkwiki is a command-line program that converts Markdown to HTML.

  To generate a single wiki, use the source_dir and dest_dir parameters. Or to
  generate multiple wikis, use the -wikis option to specify a CSV file that
  defines one wiki per line formatted as source_dir,dest_dir.

Examples:
  gomarkwiki /path/to/source /path/to/destination
  gomarkwiki -wikis wikis.csv`

// commandLineArgs stores the arguments specified on the command line.
type commandLineArgs struct {
	dirs       [][2]string
	cpuProfile string
	regen      bool
	clean      bool
	watch      bool
}

func formatVersion() string {
	return fmt.Sprintf("gomarkwiki %s compiled with %s on %s/%s", version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
}

// parseCommandLine parses the command line.
func parseCommandLine() commandLineArgs {
	// Define command line flags.
	printHelp := flag.Bool("help", false, "Show help")
	printVersion := flag.Bool("version", false, "Print version information")
	regen := flag.Bool("regen", false, "Regenerate all files regardless of timestamps")
	clean := flag.Bool("clean", false, "Delete any files in dest_dir that do not have a corresponding file in source_dir")
	watch := flag.Bool("watch", false, "Remain running and watch for changes to regenerate files on the fly")
	var wikisCsvPath string
	flag.StringVar(&wikisCsvPath, "wikis", "", "Generate wikis specified in CSV file, with one wiki defined per line formatted as source_dir,dest_dir")
	flag.BoolVar(&util.Verbose, "verbose", false, "Print status messages")
	flag.BoolVar(&util.Debug, "debug", false, "Print debug messages")
	cpuProfile := flag.String("cpuprofile", "", "Write cpu profile to file")

	// Define custom usage message.
	flag.Usage = func() {
		fmt.Println(usagePart1)
		flag.PrintDefaults()
		fmt.Println(usagePart2)
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
		fmt.Println(formatVersion())
		os.Exit(0)
	}

	// What directories are specified?
	dirs := make([][2]string, 0)
	if wikisCsvPath != "" {
		// Dirs are specified in a CSV file.
		var err error
		if dirs, err = util.LoadStringPairs(wikisCsvPath); dirs == nil || err != nil {
			util.PrintFatalError(err, "Failed to read '%s'", wikisCsvPath)
		}
	} else if flag.NArg() == 2 {
		// Dirs were specified on the command line.
		dirs = append(dirs, [2]string{flag.Arg(0), flag.Arg(1)})
	} else {
		flag.Usage()
		os.Exit(1)
	}

	return commandLineArgs{
		dirs:       dirs,
		cpuProfile: *cpuProfile,
		regen:      *regen,
		clean:      *clean,
		watch:      *watch,
	}
}

func main() {
	// Parse command line
	args := parseCommandLine()

	// Start profiling.
	if args.cpuProfile != "" {
		util.PrintVerbose("Starting profiler")
		var file *os.File
		var err error
		if file, err = os.Create(args.cpuProfile); err != nil {
			util.PrintFatalError(err, "Failed to create file '%s'", args.cpuProfile)
		}
		defer file.Close()
		if err = pprof.StartCPUProfile(file); err != nil {
			util.PrintFatalError(err, "Failed to start profiler")
		}
		defer func() {
			util.PrintVerbose("Stopping profiler")
			pprof.StopCPUProfile()
		}()
	}

	// Create Wiki instances
	var wikis []*wiki.Wiki
	var err error
	for _, dirPair := range args.dirs {
		var theWiki *wiki.Wiki
		if theWiki, err = wiki.NewWiki(dirPair[0], dirPair[1]); err != nil {
			util.PrintFatalError(err, "")
		}
		wikis = append(wikis, theWiki)
	}

	// Generate wikis
	util.PrintVerbose("Starting %s", formatVersion())
	if err = generateWikis(wikis, args.regen, args.clean, args.watch, version); err != nil {
		util.PrintFatalError(err, "")
	}

	// Success.
	os.Exit(0)
}

// generateWikis generates the wikis and then optionally watch watches for
// changes in each wiki to regenerate files on the fly.
func generateWikis(wikis []*wiki.Wiki, regen, clean, watch bool, version string) error {
	// Create channels to watch for completions, errors, and terminate signal.
	doneChan := make(chan struct{})
	errorChan := make(chan error)
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)
	defer func() {
		close(doneChan)
		close(errorChan)
		close(termChan)
	}()

	// Define worker function.
	worker := func(wiki *wiki.Wiki) {
		// Generate wiki.
		if err := wiki.Generate(regen, clean, watch, version); err != nil {
			errorChan <- err
			return
		}
		doneChan <- struct{}{}
	}

	// Start workers.
	for _, wiki := range wikis {
		go worker(wiki)
	}

	// Watch for completions, errors, and terminate signal.
	workersCount := len(wikis)
	var err error
	for workersCount > 0 && err == nil {
		select {
		case <-doneChan:
			workersCount--
		case err = <-errorChan:
			return err
		case <-termChan:
			fmt.Println("Terminate signal received. Exiting...")
			return nil
		}
	}

	return nil
}
