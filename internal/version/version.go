package version

// Version is the current version of csv2json
// This should be updated for each release following semantic versioning (MAJOR.MINOR.PATCH)
// Update this value when incrementing the VERSION file
const Version = "0.1.0"

// BuildDate is set at compile time via -ldflags
var BuildDate = "unknown"

// GitCommit is set at compile time via -ldflags
var GitCommit = "unknown"

// GetVersionInfo returns formatted version information
func GetVersionInfo() string {
	return "csv2json v" + Version
}

// GetFullVersionInfo returns detailed version information including build metadata
func GetFullVersionInfo() string {
	info := "csv2json v" + Version
	if GitCommit != "unknown" {
		info += " (commit: " + GitCommit + ")"
	}
	if BuildDate != "unknown" {
		info += " (built: " + BuildDate + ")"
	}
	return info
}
