// Package wiki generates HTML from markdown for a given wiki.
package wiki

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/stalexan/gomarkwiki/internal/util"
)

// Resource limits to prevent resource exhaustion
const (
	// MaxMarkdownFileSize is the maximum size in bytes for a markdown file that can be processed
	MaxMarkdownFileSize = 100 * 1024 * 1024 // 100 MB

	// MaxFilesProcessed is the maximum number of files that can be processed in a single wiki generation
	MaxFilesProcessed = 1000000 // 1 million files

	// MaxRecursionDepth is the maximum directory recursion depth allowed
	MaxRecursionDepth = 1000 // 1000 levels
)

// validatePlaceholder validates a placeholder name according to the rules:
// - Must not be empty
// - Must be at most 100 characters
// - Must contain at least one letter or digit
// - Can only contain letters, digits, underscore, and hyphen
func validatePlaceholder(placeholder string) error {
	if len(placeholder) == 0 {
		return fmt.Errorf("placeholder cannot be empty")
	}

	if len(placeholder) > 100 {
		return fmt.Errorf("placeholder too long (max 100 characters): %q", placeholder)
	}

	// Must have at least one alphanumeric
	hasAlnum := false
	for _, r := range placeholder {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			hasAlnum = true
			break
		}
	}
	if !hasAlnum {
		return fmt.Errorf("placeholder must contain at least one letter or digit: %q", placeholder)
	}

	// Character validation (catches everything else: braces, control chars, whitespace, special chars)
	for _, r := range placeholder {
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-') {
			return fmt.Errorf("placeholder can only contain letters, digits, underscore, and hyphen: %q", placeholder)
		}
	}

	return nil
}

// loadSubstitutionStrings loads substitution strings for a wiki, from its substitution-strings.csv
func (wiki *Wiki) loadSubstitutionStrings() error {
	// Start with no substitution strings.
	wiki.subStrings = nil

	// Is there a substitution strings file?
	const subsFileName = "substitution-strings.csv"
	candidateSubsPath := filepath.Join(wiki.SourceDir, subsFileName)
	wiki.subsPath = filepath.Clean(candidateSubsPath)

	var pairs [][2]string
	var err error
	if pairs, err = util.LoadStringPairs(candidateSubsPath); err != nil {
		return fmt.Errorf("failed to load substitution strings from '%s': %v", candidateSubsPath, err)
	}
	if len(pairs) == 0 {
		// There's either no substitution strings file or the file is empty.
		return nil
	}

	// Save substitutions.
	seenPlaceholders := make(map[string]int)
	for i, pair := range pairs {
		placeholder := strings.TrimSpace(pair[0])

		// Validate placeholder
		if err := validatePlaceholder(placeholder); err != nil {
			return fmt.Errorf("invalid placeholder at line %d: %v", i+1, err)
		}

		// Check for duplicates
		if existingLine, exists := seenPlaceholders[placeholder]; exists {
			return fmt.Errorf("duplicate placeholder %q found at line %d (first seen at line %d)", placeholder, i+1, existingLine)
		}
		seenPlaceholders[placeholder] = i + 1

		placeholder = fmt.Sprintf("{{%s}}", placeholder)
		substitution := pair[1]
		wiki.subStrings = append(wiki.subStrings, [2]string{placeholder, substitution})
	}

	return nil
}

// makeSubstitutions makes string substitutions in data.
func (wiki Wiki) makeSubstitutions(data []byte) []byte {
	for _, pair := range wiki.subStrings {
		data = bytes.ReplaceAll(data, []byte(pair[0]), []byte(pair[1]))
	}
	return data
}

// loadIngoreExpressions loads regular expressions that define which files to ingore.
func (wiki *Wiki) loadIgnoreExpressions() error {
	// Start with no ignore expressions.
	wiki.ignore = nil

	// Open ingore file, if there is one.
	const ignoreFileName = "ignore.txt"
	ignorePath := filepath.Join(wiki.SourceDir, ignoreFileName)
	wiki.ignorePath = filepath.Clean(ignorePath)
	var file *os.File
	var err error
	if file, err = os.Open(ignorePath); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("unable to open '%s': %v", ignorePath, err)
		} else {
			// There is no ignore file.
			return nil
		}
	}
	defer file.Close()

	// Read expressions.
	scanner := bufio.NewScanner(file)
	lineCount := 0
	for scanner.Scan() {
		lineCount++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Skip blank lines or comments (lines starting with '#')
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		expression, err := regexp.Compile(trimmed)
		if err != nil {
			return fmt.Errorf("error compiling regular expression '%s' on line %d: %v", trimmed, lineCount, err)
		}
		wiki.ignore = append(wiki.ignore, expression)
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading '%s': %v", ignorePath, err)
	}

	return nil
}

// ignoreFile returns true if the file at path should be ignored.
func (wiki Wiki) ignoreFile(path string) bool {
	for _, expr := range wiki.ignore {
		if expr.MatchString(path) {
			return true
		}
	}
	return false
}
