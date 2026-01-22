package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// Routing settings
	RoutesConfigPath string // Path to routes.json (if using multi-ingress mode)

	// Input settings
	InputFolder           string
	PollInterval          time.Duration
	MaxFilesPerPoll       int
	FileSuffixFilter      []string
	FilenamePattern       *regexp.Regexp
	WatchMode             string // "event", "poll", or "hybrid"
	HybridPollInterval    time.Duration

	// Parsing settings
	Delimiter rune
	QuoteChar rune
	Encoding  string
	HasHeader bool

	// Output settings
	OutputType   string // "file" or "queue"
	OutputFolder string

	// Queue settings
	QueueType     string
	QueueHost     string
	QueuePort     int
	QueueName     string
	QueueUsername string
	QueuePassword string

	// Archive settings
	ArchiveProcessed string
	ArchiveIgnored   string
	ArchiveFailed    string
	ArchiveTimestamp bool

	// Logging settings
	LogLevel         string
	LogFile          string
	LogQueueMessages bool
}

func Load() (*Config, error) {
	// Load .env file if it exists (ignore error if not present)
	_ = godotenv.Load()

	cfg := &Config{
		RoutesConfigPath:   getEnv("ROUTES_CONFIG", ""), // Empty = legacy single-input mode
		InputFolder:        getEnv("INPUT_FOLDER", "./input"),
		PollInterval:       getDurationEnv("POLL_INTERVAL_SECONDS", 5) * time.Second,
		HybridPollInterval: getDurationEnv("HYBRID_POLL_INTERVAL_SECONDS", 60) * time.Second,
		MaxFilesPerPoll:    getIntEnv("MAX_FILES_PER_POLL", 0), // 0 = no limit
		WatchMode:          getEnv("WATCH_MODE", "event"),
		Delimiter:          rune(getEnv("DELIMITER", ",")[0]),
		QuoteChar:        rune(getEnv("QUOTECHAR", "\"")[0]),
		Encoding:         getEnv("ENCODING", "utf-8"),
		HasHeader:        getBoolEnv("HAS_HEADER", true),
		OutputType:       getEnv("OUTPUT_TYPE", "file"),
		OutputFolder:     getEnv("OUTPUT_FOLDER", "./output"),
		QueueType:        getEnv("QUEUE_TYPE", "rabbitmq"),
		QueueHost:        getEnv("QUEUE_HOST", "localhost"),
		QueuePort:        getIntEnv("QUEUE_PORT", 5672),
		QueueName:        getEnv("QUEUE_NAME", ""),
		QueueUsername:    getEnv("QUEUE_USERNAME", ""),
		QueuePassword:    getEnv("QUEUE_PASSWORD", ""),
		ArchiveProcessed: getEnv("ARCHIVE_PROCESSED", "./archive/processed"),
		ArchiveIgnored:   getEnv("ARCHIVE_IGNORED", "./archive/ignored"),
		ArchiveFailed:    getEnv("ARCHIVE_FAILED", "./archive/failed"),
		ArchiveTimestamp: getBoolEnv("ARCHIVE_TIMESTAMP", true),
		LogLevel:         getEnv("LOG_LEVEL", "INFO"),
		LogFile:          getEnv("LOG_FILE", "./logs/csv2json.log"),
		LogQueueMessages: getBoolEnv("LOG_QUEUE_MESSAGES", false),
	}

	// Parse file suffix filter
	suffixFilter := getEnv("FILE_SUFFIX_FILTER", "")
	if suffixFilter != "" && suffixFilter != "*" {
		suffixes := strings.Split(suffixFilter, ",")
		for _, s := range suffixes {
			s = strings.TrimSpace(s)
			if !strings.HasPrefix(s, ".") {
				s = "." + s
			}
			cfg.FileSuffixFilter = append(cfg.FileSuffixFilter, s)
		}
	}

	// Parse filename pattern
	pattern := getEnv("FILENAME_PATTERN", ".*")
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid FILENAME_PATTERN: %w", err)
	}
	cfg.FilenamePattern = re

	// Create required directories
	dirs := []string{
		cfg.InputFolder,
		cfg.OutputFolder,
		cfg.ArchiveProcessed,
		cfg.ArchiveIgnored,
		cfg.ArchiveFailed,
		filepath.Dir(cfg.LogFile),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Validate configuration
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.OutputType != "file" && c.OutputType != "queue" {
		return fmt.Errorf("OUTPUT_TYPE must be 'file' or 'queue', got: %s", c.OutputType)
	}

	if c.OutputType == "queue" {
		if c.QueueType == "" || c.QueueHost == "" || c.QueueName == "" {
			return fmt.Errorf("QUEUE_TYPE, QUEUE_HOST, and QUEUE_NAME must be set when OUTPUT_TYPE=queue")
		}
		if c.QueuePort < 1 || c.QueuePort > 65535 {
			return fmt.Errorf("QUEUE_PORT must be between 1 and 65535, got: %d", c.QueuePort)
		}
		validTypes := []string{"rabbitmq", "kafka", "sqs", "azure-servicebus"}
		valid := false
		for _, t := range validTypes {
			if c.QueueType == t {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("QUEUE_TYPE must be one of: rabbitmq, kafka, sqs, azure-servicebus, got: %s", c.QueueType)
		}
	}

	if c.PollInterval < time.Second {
		return fmt.Errorf("POLL_INTERVAL_SECONDS must be >= 1")
	}

	return nil
}

func (c *Config) ShouldProcessFile(filename string) bool {
	// Check suffix filter
	if len(c.FileSuffixFilter) > 0 {
		match := false
		for _, suffix := range c.FileSuffixFilter {
			if strings.HasSuffix(filename, suffix) {
				match = true
				break
			}
		}
		if !match {
			return false
		}
	}

	// Check filename pattern
	return c.FilenamePattern.MatchString(filename)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue int) time.Duration {
	if value := os.Getenv(key); value != "" {
		parsed, err := strconv.Atoi(value)
		if err == nil {
			return time.Duration(parsed)
		}
	}
	return time.Duration(defaultValue)
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		parsed, err := strconv.Atoi(value)
		if err == nil {
			return parsed
		}
	}
	return defaultValue
}
