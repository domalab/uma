package version

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

// BuildInfo contains version and build information
type BuildInfo struct {
	Version    string    `json:"version"`
	GitCommit  string    `json:"git_commit,omitempty"`
	GitTag     string    `json:"git_tag,omitempty"`
	BuildTime  time.Time `json:"build_time"`
	GoVersion  string    `json:"go_version"`
	Platform   string    `json:"platform"`
	Dirty      bool      `json:"dirty,omitempty"`
}

// GetBuildInfo returns comprehensive build information
func GetBuildInfo(version string) *BuildInfo {
	info := &BuildInfo{
		Version:   sanitizeVersion(version),
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		BuildTime: time.Now(), // Fallback to current time
	}

	// Try to get build info from debug.ReadBuildInfo()
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		// Extract VCS information
		for _, setting := range buildInfo.Settings {
			switch setting.Key {
			case "vcs.revision":
				if len(setting.Value) >= 7 {
					info.GitCommit = setting.Value[:7] // Short commit hash
				} else {
					info.GitCommit = setting.Value
				}
			case "vcs.time":
				if t, err := time.Parse(time.RFC3339, setting.Value); err == nil {
					info.BuildTime = t
				}
			case "vcs.modified":
				info.Dirty = setting.Value == "true"
			}
		}
	}

	return info
}

// GetFormattedVersion returns a formatted version string for display
func GetFormattedVersion(version string) string {
	info := GetBuildInfo(version)
	
	if info.GitCommit != "" {
		if info.Dirty {
			return fmt.Sprintf("%s-%s-dirty", info.Version, info.GitCommit)
		}
		return fmt.Sprintf("%s-%s", info.Version, info.GitCommit)
	}
	
	return info.Version
}

// GetAPIVersion returns the API version for OpenAPI specification
func GetAPIVersion(version string) string {
	sanitized := sanitizeVersion(version)
	if sanitized == "" || sanitized == "unknown" {
		// Fallback to current date format
		return time.Now().Format("2006.01.02")
	}
	return sanitized
}

// sanitizeVersion cleans up the version string
func sanitizeVersion(version string) string {
	if version == "" {
		return "unknown"
	}
	
	// Remove any leading 'v' prefix
	version = strings.TrimPrefix(version, "v")
	
	// If it's just "unknown", return as is
	if version == "unknown" {
		return version
	}
	
	return version
}

// GetUserAgent returns a formatted User-Agent string
func GetUserAgent(version string) string {
	info := GetBuildInfo(version)
	return fmt.Sprintf("UMA/%s (%s; %s)", info.Version, info.Platform, info.GoVersion)
}

// IsDevVersion checks if this is a development version
func IsDevVersion(version string) bool {
	sanitized := sanitizeVersion(version)
	return sanitized == "unknown" || 
		   strings.Contains(sanitized, "dev") || 
		   strings.Contains(sanitized, "dirty")
}
