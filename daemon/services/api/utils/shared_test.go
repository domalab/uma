package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSharedUtilities_StringManipulation(t *testing.T) {
	su := NewSharedUtilities()

	t.Run("TrimAndLower", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"  Hello World  ", "hello world"},
			{"UPPERCASE", "uppercase"},
			{"", ""},
			{"   ", ""},
		}

		for _, test := range tests {
			result := su.TrimAndLower(test.input)
			if result != test.expected {
				t.Errorf("TrimAndLower(%q) = %q, expected %q", test.input, result, test.expected)
			}
		}
	})

	t.Run("SanitizeString", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"hello\x00world", "helloworld"},
			{"normal string", "normal string"},
			{"  spaced  ", "spaced"},
			{"\x01\x02\x03", ""},
		}

		for _, test := range tests {
			result := su.SanitizeString(test.input)
			if result != test.expected {
				t.Errorf("SanitizeString(%q) = %q, expected %q", test.input, result, test.expected)
			}
		}
	})

	t.Run("TruncateString", func(t *testing.T) {
		tests := []struct {
			input     string
			maxLength int
			expected  string
		}{
			{"hello world", 5, "he..."},
			{"short", 10, "short"},
			{"exact", 5, "exact"},
			{"a", 1, "a"},
		}

		for _, test := range tests {
			result := su.TruncateString(test.input, test.maxLength)
			if result != test.expected {
				t.Errorf("TruncateString(%q, %d) = %q, expected %q", test.input, test.maxLength, result, test.expected)
			}
		}
	})

	t.Run("SplitAndTrim", func(t *testing.T) {
		result := su.SplitAndTrim("  apple  ,  banana  ,  cherry  ", ",")
		expected := []string{"apple", "banana", "cherry"}

		if len(result) != len(expected) {
			t.Errorf("SplitAndTrim length mismatch: got %d, expected %d", len(result), len(expected))
		}

		for i, item := range result {
			if item != expected[i] {
				t.Errorf("SplitAndTrim[%d] = %q, expected %q", i, item, expected[i])
			}
		}
	})
}

func TestSharedUtilities_FileOperations(t *testing.T) {
	su := NewSharedUtilities()

	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "shared_utils_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("FileExists", func(t *testing.T) {
		// Create a test file
		testFile := filepath.Join(tempDir, "test.txt")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		if !su.FileExists(testFile) {
			t.Error("FileExists should return true for existing file")
		}

		if su.FileExists(filepath.Join(tempDir, "nonexistent.txt")) {
			t.Error("FileExists should return false for non-existent file")
		}

		// Test with directory (should return false)
		if su.FileExists(tempDir) {
			t.Error("FileExists should return false for directory")
		}
	})

	t.Run("DirExists", func(t *testing.T) {
		if !su.DirExists(tempDir) {
			t.Error("DirExists should return true for existing directory")
		}

		if su.DirExists(filepath.Join(tempDir, "nonexistent")) {
			t.Error("DirExists should return false for non-existent directory")
		}
	})

	t.Run("EnsureDir", func(t *testing.T) {
		newDir := filepath.Join(tempDir, "new", "nested", "dir")

		if err := su.EnsureDir(newDir); err != nil {
			t.Errorf("EnsureDir failed: %v", err)
		}

		if !su.DirExists(newDir) {
			t.Error("EnsureDir should create the directory")
		}
	})

	t.Run("SafeJoinPath", func(t *testing.T) {
		tests := []struct {
			base      string
			parts     []string
			shouldErr bool
		}{
			{tempDir, []string{"safe", "path"}, false},
			{tempDir, []string{"..", "unsafe"}, true},
			{tempDir, []string{"safe", "..", "..", "unsafe"}, true},
			{tempDir, []string{"normal", "path"}, false},
		}

		for _, test := range tests {
			result, err := su.SafeJoinPath(test.base, test.parts...)
			if test.shouldErr && err == nil {
				t.Errorf("SafeJoinPath should have failed for unsafe path: %v", test.parts)
			}
			if !test.shouldErr && err != nil {
				t.Errorf("SafeJoinPath should not have failed for safe path: %v, error: %v", test.parts, err)
			}
			if !test.shouldErr && err == nil {
				expected := filepath.Join(append([]string{test.base}, test.parts...)...)
				if result != expected {
					t.Errorf("SafeJoinPath result mismatch: got %q, expected %q", result, expected)
				}
			}
		}
	})
}

func TestSharedUtilities_DataConversion(t *testing.T) {
	su := NewSharedUtilities()

	t.Run("StringToInt", func(t *testing.T) {
		tests := []struct {
			input    string
			default_ int
			expected int
		}{
			{"123", 0, 123},
			{"invalid", 42, 42},
			{"  456  ", 0, 456},
			{"", 10, 10},
		}

		for _, test := range tests {
			result := su.StringToInt(test.input, test.default_)
			if result != test.expected {
				t.Errorf("StringToInt(%q, %d) = %d, expected %d", test.input, test.default_, result, test.expected)
			}
		}
	})

	t.Run("StringToBool", func(t *testing.T) {
		tests := []struct {
			input    string
			default_ bool
			expected bool
		}{
			{"true", false, true},
			{"yes", false, true},
			{"1", false, true},
			{"false", true, false},
			{"no", true, false},
			{"0", true, false},
			{"invalid", true, true},
			{"invalid", false, false},
		}

		for _, test := range tests {
			result := su.StringToBool(test.input, test.default_)
			if result != test.expected {
				t.Errorf("StringToBool(%q, %t) = %t, expected %t", test.input, test.default_, result, test.expected)
			}
		}
	})

	t.Run("BytesToHumanReadable", func(t *testing.T) {
		tests := []struct {
			input    int64
			expected string
		}{
			{512, "512 B"},
			{1024, "1.0 KB"},
			{1536, "1.5 KB"},
			{1048576, "1.0 MB"},
			{1073741824, "1.0 GB"},
		}

		for _, test := range tests {
			result := su.BytesToHumanReadable(test.input)
			if result != test.expected {
				t.Errorf("BytesToHumanReadable(%d) = %q, expected %q", test.input, result, test.expected)
			}
		}
	})
}

func TestSharedUtilities_Collections(t *testing.T) {
	su := NewSharedUtilities()

	t.Run("ContainsString", func(t *testing.T) {
		slice := []string{"apple", "banana", "cherry"}

		if !su.ContainsString(slice, "banana") {
			t.Error("ContainsString should return true for existing item")
		}

		if su.ContainsString(slice, "grape") {
			t.Error("ContainsString should return false for non-existing item")
		}
	})

	t.Run("ContainsStringIgnoreCase", func(t *testing.T) {
		slice := []string{"Apple", "BANANA", "cherry"}

		if !su.ContainsStringIgnoreCase(slice, "apple") {
			t.Error("ContainsStringIgnoreCase should return true for existing item (different case)")
		}

		if !su.ContainsStringIgnoreCase(slice, "CHERRY") {
			t.Error("ContainsStringIgnoreCase should return true for existing item (different case)")
		}

		if su.ContainsStringIgnoreCase(slice, "grape") {
			t.Error("ContainsStringIgnoreCase should return false for non-existing item")
		}
	})

	t.Run("RemoveDuplicateStrings", func(t *testing.T) {
		input := []string{"apple", "banana", "apple", "cherry", "banana"}
		result := su.RemoveDuplicateStrings(input)
		expected := []string{"apple", "banana", "cherry"}

		if len(result) != len(expected) {
			t.Errorf("RemoveDuplicateStrings length mismatch: got %d, expected %d", len(result), len(expected))
		}

		for _, item := range expected {
			if !su.ContainsString(result, item) {
				t.Errorf("RemoveDuplicateStrings missing expected item: %s", item)
			}
		}
	})

	t.Run("FilterStrings", func(t *testing.T) {
		input := []string{"apple", "banana", "apricot", "cherry"}
		result := su.FilterStrings(input, func(s string) bool {
			return len(s) > 5
		})
		expected := []string{"banana", "apricot", "cherry"}

		if len(result) != len(expected) {
			t.Errorf("FilterStrings length mismatch: got %d, expected %d", len(result), len(expected))
		}

		for i, item := range result {
			if item != expected[i] {
				t.Errorf("FilterStrings[%d] = %q, expected %q", i, item, expected[i])
			}
		}
	})

	t.Run("MapStrings", func(t *testing.T) {
		input := []string{"apple", "banana", "cherry"}
		result := su.MapStrings(input, func(s string) string {
			return strings.ToUpper(s)
		})
		expected := []string{"APPLE", "BANANA", "CHERRY"}

		if len(result) != len(expected) {
			t.Errorf("MapStrings length mismatch: got %d, expected %d", len(result), len(expected))
		}

		for i, item := range result {
			if item != expected[i] {
				t.Errorf("MapStrings[%d] = %q, expected %q", i, item, expected[i])
			}
		}
	})
}

func TestSharedUtilities_TimeUtilities(t *testing.T) {
	su := NewSharedUtilities()

	t.Run("FormatDuration", func(t *testing.T) {
		tests := []struct {
			input    time.Duration
			expected string
		}{
			{30 * time.Second, "30.0s"},
			{2 * time.Minute, "2.0m"},
			{1 * time.Hour, "1.0h"},
			{25 * time.Hour, "1.0d"},
		}

		for _, test := range tests {
			result := su.FormatDuration(test.input)
			if result != test.expected {
				t.Errorf("FormatDuration(%v) = %q, expected %q", test.input, result, test.expected)
			}
		}
	})

	t.Run("ParseDurationWithDefault", func(t *testing.T) {
		tests := []struct {
			input    string
			default_ time.Duration
			expected time.Duration
		}{
			{"30s", time.Minute, 30 * time.Second},
			{"invalid", time.Minute, time.Minute},
			{"2h", time.Second, 2 * time.Hour},
		}

		for _, test := range tests {
			result := su.ParseDurationWithDefault(test.input, test.default_)
			if result != test.expected {
				t.Errorf("ParseDurationWithDefault(%q, %v) = %v, expected %v", test.input, test.default_, result, test.expected)
			}
		}
	})
}

func TestSharedUtilities_ErrorHandling(t *testing.T) {
	su := NewSharedUtilities()

	t.Run("WrapError", func(t *testing.T) {
		originalErr := fmt.Errorf("original error")
		wrappedErr := su.WrapError(originalErr, "test context")

		if wrappedErr == nil {
			t.Error("WrapError should not return nil for non-nil error")
		}

		if !strings.Contains(wrappedErr.Error(), "test context") {
			t.Error("WrapError should include context in error message")
		}

		if !strings.Contains(wrappedErr.Error(), "original error") {
			t.Error("WrapError should include original error in message")
		}

		// Test with nil error
		nilWrapped := su.WrapError(nil, "test context")
		if nilWrapped != nil {
			t.Error("WrapError should return nil for nil error")
		}
	})
}
