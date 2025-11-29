package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadStringPairs(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	t.Run("valid CSV", func(t *testing.T) {
		csvPath := filepath.Join(tmpDir, "valid.csv")
		os.WriteFile(csvPath, []byte("key1,value1\nkey2,value2"), 0644)

		pairs, err := LoadStringPairs(csvPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(pairs) != 2 {
			t.Errorf("expected 2 pairs, got %d", len(pairs))
		}
		if pairs[0][0] != "key1" || pairs[0][1] != "value1" {
			t.Errorf("expected [key1, value1], got [%s, %s]", pairs[0][0], pairs[0][1])
		}
	})

	t.Run("quoted value with comma", func(t *testing.T) {
		csvPath := filepath.Join(tmpDir, "quoted.csv")
		os.WriteFile(csvPath, []byte(`key,"value, with comma"`), 0644)

		pairs, err := LoadStringPairs(csvPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if pairs[0][1] != "value, with comma" {
			t.Errorf("expected quoted value to be parsed correctly, got %q", pairs[0][1])
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		pairs, err := LoadStringPairs(filepath.Join(tmpDir, "nonexistent.csv"))
		if err != nil {
			t.Errorf("expected nil error for nonexistent file, got %v", err)
		}
		if pairs != nil {
			t.Error("expected nil pairs for nonexistent file")
		}
	})

	t.Run("wrong number of fields", func(t *testing.T) {
		csvPath := filepath.Join(tmpDir, "wrong-fields.csv")
		os.WriteFile(csvPath, []byte("key1,value1,extra"), 0644)

		_, err := LoadStringPairs(csvPath)
		if err == nil {
			t.Error("expected error for wrong number of fields")
		}
	})

	t.Run("file too large", func(t *testing.T) {
		csvPath := filepath.Join(tmpDir, "large.csv")
		// Create file larger than MaxCSVFileSize
		largeContent := strings.Repeat("a,b\n", MaxCSVFileSize/4+1)
		os.WriteFile(csvPath, []byte(largeContent), 0644)

		_, err := LoadStringPairs(csvPath)
		if err == nil {
			t.Error("expected error for file too large")
		}
		if !strings.Contains(err.Error(), "too large") {
			t.Errorf("expected 'too large' in error message, got: %v", err)
		}
	})

	t.Run("max entries limit", func(t *testing.T) {
		csvPath := filepath.Join(tmpDir, "too-many.csv")
		// Create file with MaxSubstitutionStrings + 1 entries
		var content strings.Builder
		for i := 0; i <= MaxSubstitutionStrings; i++ {
			content.WriteString(fmt.Sprintf("key%d,value%d\n", i, i))
		}
		os.WriteFile(csvPath, []byte(content.String()), 0644)

		_, err := LoadStringPairs(csvPath)
		if err == nil {
			t.Error("expected error for too many entries")
		}
		if !strings.Contains(err.Error(), "too many entries") {
			t.Errorf("expected 'too many entries' in error message, got: %v", err)
		}
	})

	t.Run("oversized field", func(t *testing.T) {
		csvPath := filepath.Join(tmpDir, "big-field.csv")
		// Create a field larger than MaxCSVFieldSize
		bigValue := strings.Repeat("x", MaxCSVFieldSize+1)
		os.WriteFile(csvPath, []byte(fmt.Sprintf("key,%s", bigValue)), 0644)

		_, err := LoadStringPairs(csvPath)
		if err == nil {
			t.Error("expected error for oversized field")
		}
		if !strings.Contains(err.Error(), "field too large") {
			t.Errorf("expected 'field too large' in error message, got: %v", err)
		}
	})

	t.Run("empty file", func(t *testing.T) {
		csvPath := filepath.Join(tmpDir, "empty.csv")
		os.WriteFile(csvPath, []byte(""), 0644)

		pairs, err := LoadStringPairs(csvPath)
		if err != nil {
			t.Errorf("unexpected error for empty file: %v", err)
		}
		if len(pairs) != 0 {
			t.Errorf("expected 0 pairs for empty file, got %d", len(pairs))
		}
	})
}
