// Package wiki generates HTML from markdown for a given wiki.
package wiki

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/stalexan/gomarkwiki/internal/util"
)

// loadSubstitutionStrings loads substitution strings for a wiki, from its substitution-strings.csv
func (wiki *Wiki) loadSubstitutionStrings() error {
	// Start with no substitution strings.
	wiki.subStrings = nil
	wiki.subsPath = ""

	// Is there a substitution strings file?
	const subsFileName = "substitution-strings.csv"
	candidateSubsPath := filepath.Join(wiki.SourceDir, subsFileName)
	var pairs [][2]string
	var err error
	if pairs, err = util.LoadStringPairs(candidateSubsPath); err != nil {
		return fmt.Errorf("failed to load substitution strings from '%s': %v", candidateSubsPath, err)
	}
	if pairs != nil {
		// There's a substitution strings file. Remember its path.
		wiki.subsPath = filepath.Clean(candidateSubsPath)
	}
	if len(pairs) == 0 {
		// There's either no substitution strings file or the file is empty.
		return nil
	}

	// Save substitutions.
	for _, pair := range pairs {
		placeholder := pair[0]
		if len(placeholder) == 0 {
			continue
		}
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
	// Open ingore file, if there is one.
	const ignoreFileName = "ignore.txt"
	ignorePath := filepath.Join(wiki.SourceDir, ignoreFileName)
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
		expression, err := regexp.Compile(line)
		if err != nil {
			return fmt.Errorf("error compiling regular expression '%s' on line %d: %v", line, lineCount, err)
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
