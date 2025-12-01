// Package wiki generates HTML from markdown for a given wiki.
package wiki

import (
	"path/filepath"
	"strings"
)

// IgnorePattern represents a single gitignore-style pattern.
type IgnorePattern struct {
	original      string // Original pattern string for error messages
	pattern       string // Processed pattern for matching
	isDir         bool   // Pattern ends with / (matches directories only)
	isAnchored    bool   // Pattern starts with / (anchored to content root)
	isNegation    bool   // Pattern starts with ! (negates previous matches)
	hasDoubleGlob bool   // Pattern contains **/ (recursive directory matching)
}

// ParseIgnorePattern parses a gitignore-style pattern string.
func ParseIgnorePattern(line string) (*IgnorePattern, error) {
	line = strings.TrimSpace(line)

	// Skip empty lines and comments
	if line == "" || strings.HasPrefix(line, "#") {
		return nil, nil
	}

	pattern := &IgnorePattern{
		original: line,
	}

	// Check for negation (must be first character)
	if strings.HasPrefix(line, "!") {
		pattern.isNegation = true
		line = strings.TrimPrefix(line, "!")
	}

	// Check for anchored pattern (starts with /)
	if strings.HasPrefix(line, "/") {
		pattern.isAnchored = true
		line = strings.TrimPrefix(line, "/")
	}

	// Check for directory-only pattern (ends with /)
	if strings.HasSuffix(line, "/") {
		pattern.isDir = true
		line = strings.TrimSuffix(line, "/")
	}

	// Check for **/ recursive directory matching
	if strings.Contains(line, "**/") {
		pattern.hasDoubleGlob = true
	}

	pattern.pattern = line
	return pattern, nil
}

// Matches returns true if the pattern matches the given path.
// relPath should be relative to the content directory.
// isDir indicates whether the path is a directory.
func (p *IgnorePattern) Matches(relPath string, isDir bool) bool {
	// Normalize path separators for matching
	relPath = filepath.ToSlash(relPath)
	pattern := filepath.ToSlash(p.pattern)

	// Handle **/ recursive matching
	if p.hasDoubleGlob {
		return p.matchRecursive(relPath)
	}

	// Handle anchored patterns (must match from root)
	if p.isAnchored {
		if p.isDir {
			// Pattern like "/backup/" - match directory or anything inside it
			if relPath == pattern {
				return isDir
			}
			return strings.HasPrefix(relPath, pattern+"/")
		}
		return p.matchPath(relPath, pattern)
	}

	// For directory-only patterns like "backup/", match the directory and its contents
	if p.isDir {
		// Check if path is the directory itself
		basename := filepath.Base(relPath)
		if p.matchPath(basename, pattern) {
			return true
		}

		// Check if path is inside the directory
		// e.g., pattern "backup/" should match "backup/file.md" or "dir/backup/file.md"
		parts := strings.Split(relPath, "/")
		for i := range parts {
			if p.matchPath(parts[i], pattern) {
				// Found a matching directory in the path
				return true
			}
		}
		return false
	}

	// Default behavior: match against basename OR any path component
	// This matches git's behavior where "logs" matches "logs", "foo/logs", "foo/logs/file.txt"
	basename := filepath.Base(relPath)

	// Try matching the basename first
	if p.matchPath(basename, pattern) {
		return true
	}

	// If pattern contains /, it's a path pattern - match against full path
	if strings.Contains(pattern, "/") {
		return p.matchPath(relPath, pattern)
	}

	// Also check if any path component matches (for directory patterns)
	parts := strings.Split(relPath, "/")
	for i := range parts {
		subPath := strings.Join(parts[i:], "/")
		if p.matchPath(subPath, pattern) {
			return true
		}
	}

	return false
}

// matchRecursive handles patterns with **/ for recursive directory matching.
func (p *IgnorePattern) matchRecursive(relPath string) bool {
	pattern := filepath.ToSlash(p.pattern)
	relPath = filepath.ToSlash(relPath)

	// Split on **/ to get prefix and suffix
	parts := strings.SplitN(pattern, "**/", 2)

	if len(parts) != 2 {
		// No **/ found (shouldn't happen, but handle gracefully)
		return p.matchPath(relPath, pattern)
	}

	prefix := parts[0]
	suffix := parts[1]

	// If there's a prefix, path must start with it
	if prefix != "" {
		if !strings.HasPrefix(relPath, prefix) {
			return false
		}
		// Remove the prefix from the path
		relPath = strings.TrimPrefix(relPath, prefix)
	}

	// Now check if any part of the remaining path matches the suffix
	if suffix == "" {
		return true // **/ matches everything after prefix
	}

	// Try matching suffix at any depth
	pathParts := strings.Split(relPath, "/")
	for i := range pathParts {
		subPath := strings.Join(pathParts[i:], "/")
		if p.matchPath(subPath, suffix) {
			return true
		}
	}

	return false
}

// matchPath performs the actual pattern matching using filepath.Match.
func (p *IgnorePattern) matchPath(path, pattern string) bool {
	// Handle exact matches first (optimization)
	if path == pattern {
		return true
	}

	// Use filepath.Match for glob pattern matching
	// Note: filepath.Match doesn't handle /, so we need to handle path patterns specially
	if strings.Contains(pattern, "/") {
		// For path patterns, match each component
		pathParts := strings.Split(path, "/")
		patternParts := strings.Split(pattern, "/")

		// If pattern has more parts than path, can't match
		if len(patternParts) > len(pathParts) {
			return false
		}

		// Match from the end (for patterns like "logs/debug.log")
		pathOffset := len(pathParts) - len(patternParts)
		for i, patternPart := range patternParts {
			pathPart := pathParts[pathOffset+i]
			matched, err := filepath.Match(patternPart, pathPart)
			if err != nil || !matched {
				return false
			}
		}
		return true
	}

	// Simple pattern (no /), use filepath.Match directly
	matched, err := filepath.Match(pattern, path)
	return err == nil && matched
}

// IgnoreMatcher manages a list of ignore patterns and determines if paths should be ignored.
type IgnoreMatcher struct {
	patterns []*IgnorePattern
}

// NewIgnoreMatcher creates a new IgnoreMatcher from a list of pattern strings.
func NewIgnoreMatcher(lines []string) (*IgnoreMatcher, error) {
	matcher := &IgnoreMatcher{
		patterns: make([]*IgnorePattern, 0),
	}

	for _, line := range lines {
		pattern, err := ParseIgnorePattern(line)
		if err != nil {
			return nil, err
		}
		if pattern != nil {
			matcher.patterns = append(matcher.patterns, pattern)
		}
	}

	return matcher, nil
}

// Matches returns true if the path should be ignored based on the patterns.
// Patterns are evaluated in order, with later patterns overriding earlier ones.
// Negation patterns (starting with !) can un-ignore files.
func (m *IgnoreMatcher) Matches(relPath string, isDir bool) bool {
	if m == nil || len(m.patterns) == 0 {
		return false
	}

	matched := false

	// Process patterns in order - later patterns override earlier ones
	for _, pattern := range m.patterns {
		if pattern.Matches(relPath, isDir) {
			// If this is a negation pattern, un-ignore the file
			// Otherwise, ignore it
			matched = !pattern.isNegation
		}
	}

	return matched
}
