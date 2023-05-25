// Package util implements utility routines for printing messages, warnings, and errors.
package util

import (
	"fmt"
	"os"
)

var Verbose bool

// PrintMessage prints a message to stdout.
func PrintMessage(message string) {
	fmt.Println(message)
}

// PrintMessage prints a message to stdout if the Verbose flag is set.
func PrintVerboseMessage(message string) {
	if Verbose {
		fmt.Println(message)
	}
}

// PrintWarning prints a warning to stderr.
func PrintWarning(message string) {
	fmt.Fprintf(os.Stderr, "WARNING: %s\n", message)
}

// PrintError prints message and error to stderr.
func PrintError(message string, err error) {
	if err != nil {
		if message != "" {
			message += ": "
		}
		message += err.Error()
	}
	fmt.Fprintf(os.Stderr, "ERROR: %s\n", message)
}

// PrintFatalError prints message and error to stderr and exits.
func PrintFatalError(message string, err error) {
	PrintError(message, err)
	os.Exit(1)
}
