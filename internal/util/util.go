// Package util implements utility routines for printing messages, warnings, and errors.
package util

import (
	"encoding/csv"
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

// LoadStringPairs loads string pairs from a CSV file, where each line is two
// comma separated strings.
func LoadStringPairs(csvPath string) ([][2]string, error) {
	// Does the CSV file exist?
	var err error
	if _, err = os.Stat(csvPath); err != nil {
		return nil, nil
	}

	// Open file.
	var file *os.File
	if file, err = os.Open(csvPath); err != nil {
		return nil, fmt.Errorf("unable to open '%s': %v", csvPath, err)
	}
	defer file.Close()

	// Read file.
	reader := csv.NewReader(file)
	reader.Comma = ','
	reader.FieldsPerRecord = 2
	var records [][]string
	if records, err = reader.ReadAll(); err != nil {
		return nil, fmt.Errorf("unable to read '%s': %v", csvPath, err)
	}

	// Save pairs.
	result := make([][2]string, 0)
	for _, record := range records {
		result = append(result, [2]string{record[0], record[1]})
	}

	return result, nil
}
