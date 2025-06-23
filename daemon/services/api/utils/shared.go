package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/domalab/uma/daemon/logger"
)

// SharedUtilities provides common utility functions used across the application
type SharedUtilities struct{}

// NewSharedUtilities creates a new shared utilities instance
func NewSharedUtilities() *SharedUtilities {
	return &SharedUtilities{}
}

// String manipulation utilities

// TrimAndLower trims whitespace and converts to lowercase
func (su *SharedUtilities) TrimAndLower(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// SanitizeString removes potentially dangerous characters from a string
func (su *SharedUtilities) SanitizeString(s string) string {
	// Remove null bytes and control characters
	s = strings.ReplaceAll(s, "\x00", "")
	s = regexp.MustCompile(`[\x00-\x1f\x7f]`).ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}

// TruncateString truncates a string to a maximum length with ellipsis
func (su *SharedUtilities) TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	if maxLength <= 3 {
		return s[:maxLength]
	}
	return s[:maxLength-3] + "..."
}

// SplitAndTrim splits a string by delimiter and trims each part
func (su *SharedUtilities) SplitAndTrim(s, delimiter string) []string {
	parts := strings.Split(s, delimiter)
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// JoinNonEmpty joins non-empty strings with a delimiter
func (su *SharedUtilities) JoinNonEmpty(parts []string, delimiter string) string {
	nonEmpty := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.TrimSpace(part) != "" {
			nonEmpty = append(nonEmpty, part)
		}
	}
	return strings.Join(nonEmpty, delimiter)
}

// File and path utilities

// FileExists checks if a file exists and is not a directory
func (su *SharedUtilities) FileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// DirExists checks if a directory exists
func (su *SharedUtilities) DirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// PathExists checks if a path exists (file or directory)
func (su *SharedUtilities) PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// EnsureDir creates a directory if it doesn't exist
func (su *SharedUtilities) EnsureDir(path string) error {
	if !su.DirExists(path) {
		return os.MkdirAll(path, 0755)
	}
	return nil
}

// GetFileSize returns the size of a file in bytes
func (su *SharedUtilities) GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// GetFileModTime returns the modification time of a file
func (su *SharedUtilities) GetFileModTime(path string) (time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}

// SafeJoinPath safely joins path components and validates the result
func (su *SharedUtilities) SafeJoinPath(base string, parts ...string) (string, error) {
	result := filepath.Join(append([]string{base}, parts...)...)
	
	// Ensure the result is still within the base directory
	absBase, err := filepath.Abs(base)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for base: %v", err)
	}
	
	absResult, err := filepath.Abs(result)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for result: %v", err)
	}
	
	if !strings.HasPrefix(absResult, absBase) {
		return "", fmt.Errorf("path traversal detected: %s", result)
	}
	
	return result, nil
}

// Data conversion utilities

// StringToInt converts string to int with default value
func (su *SharedUtilities) StringToInt(s string, defaultValue int) int {
	if val, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
		return val
	}
	return defaultValue
}

// StringToInt64 converts string to int64 with default value
func (su *SharedUtilities) StringToInt64(s string, defaultValue int64) int64 {
	if val, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64); err == nil {
		return val
	}
	return defaultValue
}

// StringToFloat64 converts string to float64 with default value
func (su *SharedUtilities) StringToFloat64(s string, defaultValue float64) float64 {
	if val, err := strconv.ParseFloat(strings.TrimSpace(s), 64); err == nil {
		return val
	}
	return defaultValue
}

// StringToBool converts string to bool with default value
func (su *SharedUtilities) StringToBool(s string, defaultValue bool) bool {
	s = su.TrimAndLower(s)
	switch s {
	case "true", "yes", "1", "on", "enabled":
		return true
	case "false", "no", "0", "off", "disabled":
		return false
	default:
		return defaultValue
	}
}

// BytesToHumanReadable converts bytes to human readable format
func (su *SharedUtilities) BytesToHumanReadable(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	
	units := []string{"KB", "MB", "GB", "TB", "PB"}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}

// Collection utilities

// ContainsString checks if a slice contains a string
func (su *SharedUtilities) ContainsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ContainsStringIgnoreCase checks if a slice contains a string (case insensitive)
func (su *SharedUtilities) ContainsStringIgnoreCase(slice []string, item string) bool {
	lowerItem := strings.ToLower(item)
	for _, s := range slice {
		if strings.ToLower(s) == lowerItem {
			return true
		}
	}
	return false
}

// RemoveDuplicateStrings removes duplicate strings from a slice
func (su *SharedUtilities) RemoveDuplicateStrings(slice []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(slice))
	
	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	
	return result
}

// FilterStrings filters strings based on a predicate function
func (su *SharedUtilities) FilterStrings(slice []string, predicate func(string) bool) []string {
	result := make([]string, 0, len(slice))
	for _, item := range slice {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

// MapStrings transforms strings using a mapper function
func (su *SharedUtilities) MapStrings(slice []string, mapper func(string) string) []string {
	result := make([]string, len(slice))
	for i, item := range slice {
		result[i] = mapper(item)
	}
	return result
}

// Error handling utilities

// WrapError wraps an error with additional context
func (su *SharedUtilities) WrapError(err error, context string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", context, err)
}

// LogAndReturnError logs an error and returns it
func (su *SharedUtilities) LogAndReturnError(err error, context string) error {
	if err == nil {
		return nil
	}
	wrappedErr := su.WrapError(err, context)
	logger.Red("%v", wrappedErr)
	return wrappedErr
}

// IgnoreError logs an error but doesn't return it (for cleanup operations)
func (su *SharedUtilities) IgnoreError(err error, context string) {
	if err != nil {
		logger.Yellow("Ignoring error in %s: %v", context, err)
	}
}

// Time utilities

// FormatDuration formats a duration in a human-readable way
func (su *SharedUtilities) FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%.1fh", d.Hours())
	} else {
		return fmt.Sprintf("%.1fd", d.Hours()/24)
	}
}

// ParseDurationWithDefault parses a duration string with a default value
func (su *SharedUtilities) ParseDurationWithDefault(s string, defaultValue time.Duration) time.Duration {
	if d, err := time.ParseDuration(strings.TrimSpace(s)); err == nil {
		return d
	}
	return defaultValue
}

// Global shared utilities instance
var globalSharedUtilities = NewSharedUtilities()

// GetSharedUtilities returns the global shared utilities instance
func GetSharedUtilities() *SharedUtilities {
	return globalSharedUtilities
}

// Convenience functions that use the global instance

// TrimAndLower trims whitespace and converts to lowercase
func TrimAndLower(s string) string {
	return globalSharedUtilities.TrimAndLower(s)
}

// SanitizeString removes potentially dangerous characters from a string
func SanitizeString(s string) string {
	return globalSharedUtilities.SanitizeString(s)
}

// FileExists checks if a file exists and is not a directory
func FileExists(path string) bool {
	return globalSharedUtilities.FileExists(path)
}

// DirExists checks if a directory exists
func DirExists(path string) bool {
	return globalSharedUtilities.DirExists(path)
}

// ContainsString checks if a slice contains a string
func ContainsString(slice []string, item string) bool {
	return globalSharedUtilities.ContainsString(slice, item)
}

// BytesToHumanReadable converts bytes to human readable format
func BytesToHumanReadable(bytes int64) string {
	return globalSharedUtilities.BytesToHumanReadable(bytes)
}
