package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/natefinch/lumberjack.v2"
)

// FileLoggerConfig holds configuration for file-based logging
type FileLoggerConfig struct {
	Filename   string `json:"filename"`
	MaxSize    int    `json:"max_size"`    // megabytes
	MaxBackups int    `json:"max_backups"` // number of backup files
	MaxAge     int    `json:"max_age"`     // days
	Compress   bool   `json:"compress"`    // compress backup files
}

// DefaultFileLoggerConfig returns optimized configuration for Unraid systems
func DefaultFileLoggerConfig(logsDir string) FileLoggerConfig {
	return FileLoggerConfig{
		Filename:   filepath.Join(logsDir, "uma.log"),
		MaxSize:    5,     // 5MB limit for minimal disk usage
		MaxBackups: 0,     // DISABLED - no backup files to prevent disk space issues
		MaxAge:     0,     // DISABLED - no age-based retention
		Compress:   false, // DISABLED - no compression to avoid backup files
	}
}

// UnraidOptimizedConfig returns configuration specifically optimized for Unraid systems
// where /var/log space is limited and should be conserved
func UnraidOptimizedConfig(logsDir string) FileLoggerConfig {
	return FileLoggerConfig{
		Filename:   filepath.Join(logsDir, "uma.log"),
		MaxSize:    5,     // 5MB maximum file size for minimal disk usage
		MaxBackups: 0,     // No backup files - when limit reached, truncate
		MaxAge:     0,     // No age-based rotation
		Compress:   false, // No compression to avoid creating additional files
	}
}

// SetupFileLogger configures the global logger with file output and disk space optimization
func SetupFileLogger(config FileLoggerConfig) error {
	// Ensure the log directory exists
	logDir := filepath.Dir(config.Filename)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory %s: %w", logDir, err)
	}

	// Create lumberjack logger with optimized settings
	fileLogger := &lumberjack.Logger{
		Filename:   config.Filename,
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
		Compress:   config.Compress,
	}

	// Set up multi-writer to log to both file and stdout
	multiWriter := io.MultiWriter(os.Stdout, fileLogger)
	log.SetOutput(multiWriter)

	// Log the configuration for verification
	log.Printf("File logging configured: %s (max_size: %dMB, max_backups: %d, max_age: %d days, compress: %t)",
		config.Filename, config.MaxSize, config.MaxBackups, config.MaxAge, config.Compress)

	return nil
}

// CleanupOldLogFiles removes any existing backup log files to ensure clean state
func CleanupOldLogFiles(logsDir string) error {
	// Pattern to match UMA log backup files (uma.log.1, uma.log.2, etc.)
	pattern := filepath.Join(logsDir, "uma.log.*")

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to find log backup files: %w", err)
	}

	for _, file := range matches {
		if err := os.Remove(file); err != nil {
			log.Printf("Warning: failed to remove old log file %s: %v", file, err)
		} else {
			log.Printf("Removed old log backup file: %s", file)
		}
	}

	return nil
}

// GetLogFileSize returns the current size of the main log file in bytes
func GetLogFileSize(filename string) (int64, error) {
	info, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil // File doesn't exist yet
		}
		return 0, err
	}
	return info.Size(), nil
}

// LogDiskUsageInfo logs information about current log file usage
func LogDiskUsageInfo(logsDir string) {
	logFile := filepath.Join(logsDir, "uma.log")

	size, err := GetLogFileSize(logFile)
	if err != nil {
		log.Printf("Warning: could not get log file size: %v", err)
		return
	}

	sizeMB := float64(size) / (1024 * 1024)
	log.Printf("Current log file size: %.2f MB (limit: 10 MB)", sizeMB)

	// Check for any backup files that shouldn't exist
	pattern := filepath.Join(logsDir, "uma.log.*")
	matches, err := filepath.Glob(pattern)
	if err == nil && len(matches) > 0 {
		log.Printf("Warning: Found %d backup log files that should not exist: %v", len(matches), matches)
		log.Printf("Consider running cleanup to remove these files and free disk space")
	}
}

// ValidateLogConfiguration validates the logging configuration for Unraid compatibility
func ValidateLogConfiguration(config FileLoggerConfig) error {
	// Validate max size
	if config.MaxSize <= 0 {
		return fmt.Errorf("max_size must be greater than 0, got: %d", config.MaxSize)
	}
	if config.MaxSize > 100 {
		log.Printf("Warning: max_size %dMB is quite large for Unraid systems, consider reducing to 10MB or less", config.MaxSize)
	}

	// Warn about backup files on Unraid
	if config.MaxBackups > 0 {
		log.Printf("Warning: max_backups is set to %d, which will create backup files and consume additional disk space", config.MaxBackups)
		log.Printf("For Unraid systems, it's recommended to set max_backups to 0 to prevent disk space issues")
	}

	// Warn about compression on Unraid
	if config.Compress {
		log.Printf("Warning: compression is enabled, which may create additional files during rotation")
		log.Printf("For Unraid systems, it's recommended to disable compression to minimize disk usage")
	}

	// Validate filename
	if config.Filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	return nil
}
