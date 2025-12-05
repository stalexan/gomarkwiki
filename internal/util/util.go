// Package util implements utility routines for printing messages, warnings, and errors.
package util

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
)

var Verbose bool
var Debug bool

// Resource limits for CSV file processing
const (
	// MaxCSVFileSize is the maximum size in bytes for a CSV file that can be processed
	MaxCSVFileSize = 10 * 1024 * 1024 // 10 MB

	// MaxCSVFieldSize is the maximum size in bytes for a single CSV field
	MaxCSVFieldSize = 64 * 1024 // 64 KB

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

	// Read file incrementally to check field sizes and prevent memory exhaustion
	reader := csv.NewReader(file)
	reader.Comma = ','
	reader.FieldsPerRecord = -1 // Allow variable fields per record for flexible parsing
	reader.TrimLeadingSpace = true
	reader.Comment = '#' // Allow comment lines starting with #

	result := make([][2]string, 0)
	lineNum := 0

	for {
		record, err := reader.Read()
		if err != nil {
			// Check if we've reached end of file
			if errors.Is(err, io.EOF) {
				break
			}
			// Handle CSV parse errors
			if parseErr, ok := err.(*csv.ParseError); ok {
				return nil, fmt.Errorf("CSV parse error in '%s' at line %d: %v", csvPath, parseErr.Line, err)
			}
			return nil, fmt.Errorf("unable to read '%s': %v", csvPath, err)
		}

		lineNum++

		// Skip blank lines
		if len(record) == 0 || (len(record) == 1 && record[0] == "") {
			continue
		}

		// Validate that each non-empty record has exactly 2 fields
		if len(record) != 2 {
			return nil, fmt.Errorf("CSV file '%s' has wrong number of fields at line %d: expected 2 fields, got %d (blank lines and # comments are allowed)", csvPath, lineNum, len(record))
		}

		// Check number of entries limit
		if len(result) >= MaxSubstitutionStrings {
			return nil, fmt.Errorf("CSV file '%s' has too many entries (max %d entries)", csvPath, MaxSubstitutionStrings)
		}

		// Validate field sizes to prevent memory exhaustion from oversized fields
		field0Size := len(record[0])
		field1Size := len(record[1])

		if field0Size > MaxCSVFieldSize {
			return nil, fmt.Errorf("CSV file '%s' has field too large at line %d, field 1 (%d bytes, max %d bytes)", csvPath, lineNum, field0Size, MaxCSVFieldSize)
		}

		if field1Size > MaxCSVFieldSize {
			return nil, fmt.Errorf("CSV file '%s' has field too large at line %d, field 2 (%d bytes, max %d bytes)", csvPath, lineNum, field1Size, MaxCSVFieldSize)
		}

		// Save pair
		result = append(result, [2]string{record[0], record[1]})
	}

	return result, nil
}
