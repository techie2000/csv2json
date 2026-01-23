package version

import (
	"os"
	"strings"
)

// Version is the fallback version if VERSION file cannot be read
// In normal operation, version is read from VERSION file via GetVersion()
// This constant should remain "unknown" - do not hardcode versions here
const Version = "unknown"

// BuildDate is set at compile time via -ldflags
var BuildDate = "unknown"

// GitCommit is set at compile time via -ldflags
var GitCommit = "unknown"

// ReadVersionFromFile reads the VERSION file from project root
// Falls back to const Version if file cannot be read
func ReadVersionFromFile() string {
	// Try multiple paths to find VERSION file
	paths := []string{
		"VERSION",
		"../VERSION",
		"../../VERSION",
		"../../../VERSION",
	}

	for _, path := range paths {
		if content, err := os.ReadFile(path); err == nil {
			return strings.TrimSpace(string(content))
		}
	}

	// Fallback to const if file not found
	return Version
}

// GetVersion returns the version, preferring VERSION file over const
func GetVersion() string {
	return ReadVersionFromFile()
}

// GetVersionInfo returns formatted version information
func GetVersionInfo() string {
	return "csv2json v" + GetVersion()
}

// GetFullVersionInfo returns detailed version information including build metadata
func GetFullVersionInfo() string {
	info := "csv2json v" + GetVersion()
	if GitCommit != "unknown" {
		info += " (commit: " + GitCommit + ")"
	}
	if BuildDate != "unknown" {
		info += " (built: " + BuildDate + ")"
	}
	return info
}
