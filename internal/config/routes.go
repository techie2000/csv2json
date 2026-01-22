package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

// Route represents a single ingestion route configuration
type Route struct {
	Name    string        `json:"name"`
	Input   InputConfig   `json:"input"`
	Parsing ParsingConfig `json:"parsing"`
	Output  OutputConfig  `json:"output"`
	Archive ArchiveConfig `json:"archive"`
}

// InputConfig defines input folder and filtering
type InputConfig struct {
	Path               string `json:"path"`
	FilenamePattern    string `json:"filenamePattern,omitempty"`
	SuffixFilter       string `json:"suffixFilter,omitempty"`
	PollIntervalSec    int    `json:"pollIntervalSeconds"`
	MaxFilesPerPoll    int    `json:"maxFilesPerPoll,omitempty"`
	compiledPattern    *regexp.Regexp
	compiledSuffixList []string
}

// ParsingConfig defines CSV parsing semantics
type ParsingConfig struct {
	HasHeader bool   `json:"hasHeader"`
	Delimiter string `json:"delimiter"`
	QuoteChar string `json:"quoteChar,omitempty"`
	Encoding  string `json:"encoding,omitempty"`
}

// OutputConfig defines destination and type
type OutputConfig struct {
	Type                string `json:"type"` // "file" or "queue"
	Destination         string `json:"destination"`
	IncludeRouteContext bool   `json:"includeRouteContext"` // Default: true
}

// ArchiveConfig defines archive paths
type ArchiveConfig struct {
	ProcessedPath string `json:"processedPath"`
	FailedPath    string `json:"failedPath"`
	IgnoredPath   string `json:"ignoredPath,omitempty"`
}

// RoutesConfig represents the complete routes.json structure
type RoutesConfig struct {
	Routes []Route `json:"routes"`
}

// LoadRoutes loads routes from the JSON configuration file
func LoadRoutes(configPath string) (*RoutesConfig, error) {
	if configPath == "" {
		return nil, fmt.Errorf("ROUTES_CONFIG path is empty")
	}

	// Read file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read routes config: %w", err)
	}

	// Parse JSON
	var routesConfig RoutesConfig
	if err := json.Unmarshal(data, &routesConfig); err != nil {
		return nil, fmt.Errorf("failed to parse routes JSON: %w", err)
	}

	// Validate and compile patterns
	for i := range routesConfig.Routes {
		route := &routesConfig.Routes[i]

		// Validate required fields
		if route.Name == "" {
			return nil, fmt.Errorf("route at index %d missing required field 'name'", i)
		}
		if route.Input.Path == "" {
			return nil, fmt.Errorf("route '%s': missing required field 'input.path'", route.Name)
		}
		if route.Output.Type == "" || route.Output.Destination == "" {
			return nil, fmt.Errorf("route '%s': missing required output configuration", route.Name)
		}
		if route.Archive.ProcessedPath == "" || route.Archive.FailedPath == "" {
			return nil, fmt.Errorf("route '%s': missing required archive paths", route.Name)
		}

		// Verify paths exist
		if _, err := os.Stat(route.Input.Path); os.IsNotExist(err) {
			return nil, fmt.Errorf("route '%s': input path does not exist: %s", route.Name, route.Input.Path)
		}

		// Set defaults
		if route.Input.PollIntervalSec == 0 {
			route.Input.PollIntervalSec = 10
		}
		if route.Parsing.Delimiter == "" {
			route.Parsing.Delimiter = ","
		}
		if route.Parsing.QuoteChar == "" {
			route.Parsing.QuoteChar = "\""
		}
		if route.Parsing.Encoding == "" {
			route.Parsing.Encoding = "utf-8"
		}

		// Compile filename pattern if specified
		if route.Input.FilenamePattern != "" {
			compiled, err := regexp.Compile(route.Input.FilenamePattern)
			if err != nil {
				return nil, fmt.Errorf("route '%s': invalid filename pattern: %w", route.Name, err)
			}
			route.Input.compiledPattern = compiled
		}

		// Parse suffix filter if specified
		if route.Input.SuffixFilter != "" {
			route.Input.compiledSuffixList = parseSuffixFilter(route.Input.SuffixFilter)
		}

		// Create archive directories
		for _, archivePath := range []string{
			route.Archive.ProcessedPath,
			route.Archive.FailedPath,
			route.Archive.IgnoredPath,
		} {
			if archivePath != "" {
				if err := os.MkdirAll(archivePath, 0755); err != nil {
					return nil, fmt.Errorf("route '%s': failed to create archive directory %s: %w", route.Name, archivePath, err)
				}
			}
		}
	}

	return &routesConfig, nil
}

// ToLegacyConfig converts a Route to the legacy Config structure for compatibility
func (r *Route) ToLegacyConfig() *Config {
	delimiter := ','
	if len(r.Parsing.Delimiter) > 0 {
		delimiter = rune(r.Parsing.Delimiter[0])
	}

	quoteChar := '"'
	if len(r.Parsing.QuoteChar) > 0 {
		quoteChar = rune(r.Parsing.QuoteChar[0])
	}

	cfg := &Config{
		InputFolder:      r.Input.Path,
		PollInterval:     time.Duration(r.Input.PollIntervalSec) * time.Second,
		MaxFilesPerPoll:  r.Input.MaxFilesPerPoll,
		FilenamePattern:  r.Input.compiledPattern,
		Delimiter:        delimiter,
		QuoteChar:        quoteChar,
		Encoding:         r.Parsing.Encoding,
		HasHeader:        r.Parsing.HasHeader,
		ArchiveProcessed: r.Archive.ProcessedPath,
		ArchiveIgnored:   r.Archive.IgnoredPath,
		ArchiveFailed:    r.Archive.FailedPath,
		ArchiveTimestamp: true, // Always timestamp in routing mode
	}

	// Parse suffix filter
	if len(r.Input.compiledSuffixList) > 0 {
		cfg.FileSuffixFilter = r.Input.compiledSuffixList
	}

	// Parse output configuration
	cfg.OutputType = r.Output.Type
	if r.Output.Type == "file" {
		cfg.OutputFolder = r.Output.Destination
	} else if r.Output.Type == "queue" {
		// Parse queue destination (e.g., "rabbitmq://products_queue")
		cfg.QueueName = parseQueueDestination(r.Output.Destination)
		cfg.QueueType = "rabbitmq" // Default to RabbitMQ
		// Use global queue connection settings from environment
		cfg.QueueHost = getEnv("QUEUE_HOST", "localhost")
		cfg.QueuePort = getIntEnv("QUEUE_PORT", 5672)
		cfg.QueueUsername = getEnv("QUEUE_USERNAME", "")
		cfg.QueuePassword = getEnv("QUEUE_PASSWORD", "")
	}

	return cfg
}

// parseQueueDestination extracts queue name from destination string
// Examples:
//   - "rabbitmq://products_queue" -> "products_queue"
//   - "products_queue" -> "products_queue"
func parseQueueDestination(dest string) string {
	// Simple parsing: remove protocol prefix if present
	if idx := filepath.Base(dest); idx != "" {
		return idx
	}
	return dest
}

// parseSuffixFilter parses comma-separated suffix filter
func parseSuffixFilter(filter string) []string {
	if filter == "" {
		return nil
	}
	suffixes := []string{}
	for _, suffix := range regexp.MustCompile(`\s*,\s*`).Split(filter, -1) {
		if suffix != "" {
			suffixes = append(suffixes, suffix)
		}
	}
	return suffixes
}
