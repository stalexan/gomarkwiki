// Package util implements utility routines for printing messages, warnings, and errors.
package util

import (
	"fmt"
	"os"
)

var Verbose bool
var Debug bool

func formatMessage(format string, args []interface{}) string {
	return fmt.Sprintf(format, args...)
}

// PrintMessage prints a message to stdout.
func PrintMessage(format string, args ...interface{}) {
	fmt.Println(formatMessage(format, args))
}

// PrintVerbose prints a message to stdout if either the Verbose or Debug flag is set.
func PrintVerbose(format string, args ...interface{}) {
	if Verbose || Debug {
		fmt.Println(formatMessage(format, args))
	}
}

// PrintDebug prints a message to stdout if the Debug flag is set.
func PrintDebug(format string, args ...interface{}) {
	if Debug {
		fmt.Println(formatMessage(format, args))
	}
}

// PrintWarning prints a warning to stderr.
func PrintWarning(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "WARNING: %s\n", formatMessage(format, args))
}

// PrintFatalError prints a error message to stderr and exits.
func PrintFatalError(err error, format string, args ...interface{}) {
	message := formatMessage(format, args)
	if err != nil {
		if message != "" {
			message += ": "
		}
		message += err.Error()
	}
	fmt.Fprintf(os.Stderr, "ERROR: %s\n", message)

	os.Exit(1)
}
