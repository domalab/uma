package logger

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gookit/color"
)

func Red(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	printer(color.Red, format, args...)

	// Capture error messages in Sentry for production monitoring
	sentry.CaptureMessage(message)
}

func Blue(format string, args ...interface{}) {
	printer(color.Blue, format, args...)
}

func Green(format string, args ...interface{}) {
	printer(color.Green, format, args...)
}

func Yellow(format string, args ...interface{}) {
	printer(color.Yellow, format, args...)
}

func LightGreen(format string, args ...interface{}) {
	printer2(color.S256(106), format, args...)
}

func LightRed(format string, args ...interface{}) {
	printer2(color.S256(9), format, args...)
}

func NotBetter(format string, args ...interface{}) {
	printer2(color.S256(132), format, args...)
}

func Olive(format string, args ...interface{}) {
	printer2(color.S256(11), format, args...)
}

func LightBlue(format string, args ...interface{}) {
	printer2(color.S256(14), format, args...)
}

func printer(fn color.Color, format string, args ...interface{}) {
	line := fmt.Sprintf(format, args...)
	fn.Printf("%s %s\n", time.Now().Format("15:04"), line)
	log.Println(line)
}

func printer2(fn *color.Style256, format string, args ...interface{}) {
	line := fmt.Sprintf(format, args...)
	fn.Printf("%s %s\n", time.Now().Format("15:04"), line)
	log.Println(line)
}

// Production logging optimization variables
var (
	isProductionMode = false
	verbosePatterns  = []string{
		"HTTP GET /health",
		"HTTP GET /metrics",
		"HTTP GET /api/v1/system/stats",
		"Monitoring cycle completed",
		"Cache hit for",
		"WebSocket ping",
		"Heartbeat",
	}
)

// SetProductionMode enables production logging optimizations
func SetProductionMode(enabled bool) {
	isProductionMode = enabled
	if enabled {
		Green("Production logging mode enabled - verbose messages will be filtered")
	}
}

// IsProductionMode returns whether production mode is enabled
func IsProductionMode() bool {
	return isProductionMode
}

// shouldSkipVerboseLog determines if a log message should be skipped in production
func shouldSkipVerboseLog(message string) bool {
	if !IsProductionMode() {
		return false
	}

	for _, pattern := range verbosePatterns {
		if strings.Contains(message, pattern) {
			return true
		}
	}
	return false
}

// ProductionInfo logs info messages only if not verbose in production mode
func ProductionInfo(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	if !shouldSkipVerboseLog(message) {
		Blue(format, args...)
	}
}

// ProductionDebug logs debug messages only in non-production mode
func ProductionDebug(format string, args ...interface{}) {
	if !IsProductionMode() {
		LightGreen(format, args...)
	}
}
