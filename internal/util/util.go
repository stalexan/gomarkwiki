// Package util implements utility routines for printing messages, warnings, and errors.
package util

import (
	"encoding/csv"
	"fmt"
	"os"
)

var Verbose bool
var Debug bool

// Resource limits for CSV file processing
const (
	// MaxCSVFileSize is the maximum size in bytes for a CSV file that can be processed
	MaxCSVFileSize = 10 * 1024 * 1024 // 10 MB

	// MaxSubstitutionStrings is the maximum number of substitution string pairs allowed
	MaxSubstitutionStrings = 10000 // 10,000 pairs
)

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

func formatErrorMessage(err error, format string, args []interface{}) string {
	message := "ERROR"
	if format != "" {
		message += fmt.Sprintf(": %s", formatMessage(format, args))
	}
	if err != nil {
		message += fmt.Sprintf(": %v", err)
	}
	return message
}

// PrintError prints a error message to stderr.
func PrintError(err error, format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s\n", formatErrorMessage(err, format, args))
}

// PrintFatalError prints a error message to stderr and exits.
func PrintFatalError(err error, format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s\n", formatErrorMessage(err, format, args))
	os.Exit(1)
}

// LoadStringPairs loads string pairs from a CSV file, where each line is two
// comma separated strings.
func LoadStringPairs(csvPath string) ([][2]string, error) {
	// Open file.
	var file *os.File
	var err error
	if file, err = os.Open(csvPath); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("unable to open '%s': %v", csvPath, err)
		} else {
			// There is no string pairs file.
			return nil, nil
		}
	}
	defer file.Close()

	// Check file size before reading to prevent resource exhaustion
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("unable to stat '%s': %v", csvPath, err)
	}

	if fileInfo.Size() > MaxCSVFileSize {
		return nil, fmt.Errorf("CSV file '%s' is too large (%d bytes, max %d bytes)", csvPath, fileInfo.Size(), MaxCSVFileSize)
	}

	// Read file.
	reader := csv.NewReader(file)
	reader.Comma = ','
	reader.FieldsPerRecord = 2
	var records [][]string
	if records, err = reader.ReadAll(); err != nil {
		if parseErr, ok := err.(*csv.ParseError); ok {
			return nil, fmt.Errorf("CSV parse error in '%s' at line %d: %v", csvPath, parseErr.Line, err)
		}
		return nil, fmt.Errorf("unable to read '%s': %v", csvPath, err)
	}

	// Check number of entries limit
	if len(records) > MaxSubstitutionStrings {
		return nil, fmt.Errorf("CSV file '%s' has too many entries (%d entries, max %d entries)", csvPath, len(records), MaxSubstitutionStrings)
	}

	// Save pairs.
	result := make([][2]string, 0, len(records))
	for _, record := range records {
		result = append(result, [2]string{record[0], record[1]})
	}

	return result, nil
}
