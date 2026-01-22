package config

import (
	"os"
	"testing"
	"time"
)

// TestLoadDefaultConfig validates default configuration values
func TestLoadDefaultConfig(t *testing.T) {
	// Clear environment
	os.Clearenv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Expected successful load with defaults, got error: %v", err)
	}

	// Validate defaults
	if cfg.InputFolder != "./input" {
		t.Errorf("Expected default InputFolder './input', got '%s'", cfg.InputFolder)
	}

	if cfg.PollInterval != 5*time.Second {
		t.Errorf("Expected default PollInterval 5s, got %v", cfg.PollInterval)
	}

	if cfg.MaxFilesPerPoll != 0 {
		t.Errorf("Expected default MaxFilesPerPoll 0, got %d", cfg.MaxFilesPerPoll)
	}

	if cfg.Delimiter != ',' {
		t.Errorf("Expected default Delimiter ',', got '%c'", cfg.Delimiter)
	}

	if cfg.QuoteChar != '"' {
		t.Errorf("Expected default QuoteChar '\"', got '%c'", cfg.QuoteChar)
	}

	if cfg.OutputType != "file" {
		t.Errorf("Expected default OutputType 'file', got '%s'", cfg.OutputType)
	}
}

// TestLoadCustomConfig validates environment variable loading
func TestLoadCustomConfig(t *testing.T) {
	os.Clearenv()

	os.Setenv("INPUT_FOLDER", "/custom/input")
	os.Setenv("POLL_INTERVAL_SECONDS", "10")
	os.Setenv("MAX_FILES_PER_POLL", "50")
	os.Setenv("DELIMITER", "|")
	os.Setenv("OUTPUT_TYPE", "queue")
	os.Setenv("QUEUE_TYPE", "rabbitmq")
	os.Setenv("QUEUE_HOST", "localhost")
	os.Setenv("QUEUE_PORT", "5672")
	os.Setenv("QUEUE_NAME", "test")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Expected successful load, got error: %v", err)
	}

	if cfg.InputFolder != "/custom/input" {
		t.Errorf("Expected InputFolder '/custom/input', got '%s'", cfg.InputFolder)
	}

	if cfg.PollInterval != 10*time.Second {
		t.Errorf("Expected PollInterval 10s, got %v", cfg.PollInterval)
	}

	if cfg.MaxFilesPerPoll != 50 {
		t.Errorf("Expected MaxFilesPerPoll 50, got %d", cfg.MaxFilesPerPoll)
	}

	if cfg.Delimiter != '|' {
		t.Errorf("Expected Delimiter '|', got '%c'", cfg.Delimiter)
	}

	if cfg.OutputType != "queue" {
		t.Errorf("Expected OutputType 'queue', got '%s'", cfg.OutputType)
	}
}

// TestValidateQueueConfig validates queue configuration validation
// ADR-003: QUEUE_TYPE must be explicit enum value
func TestValidateQueueConfig(t *testing.T) {
	os.Clearenv()

	testCases := []struct {
		name        string
		queueType   string
		queueHost   string
		queuePort   string
		shouldError bool
	}{
		{"valid rabbitmq", "rabbitmq", "localhost", "5672", false},
		{"valid kafka", "kafka", "localhost", "9092", false},
		{"valid sqs", "sqs", "localhost", "9000", false},
		{"valid azure-servicebus", "azure-servicebus", "localhost", "5672", false},
		{"invalid type", "invalid", "localhost", "5672", true},
		{"invalid port low", "rabbitmq", "localhost", "0", true},
		{"invalid port high", "rabbitmq", "localhost", "99999", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Clearenv()
			os.Setenv("OUTPUT_TYPE", "queue")
			os.Setenv("QUEUE_TYPE", tc.queueType)
			os.Setenv("QUEUE_HOST", tc.queueHost)
			os.Setenv("QUEUE_PORT", tc.queuePort)
			os.Setenv("QUEUE_NAME", "test")

			cfg, err := Load()

			if tc.shouldError && err == nil {
				t.Errorf("Expected error for %s, got success", tc.name)
			}

			if !tc.shouldError && err != nil {
				t.Errorf("Expected success for %s, got error: %v", tc.name, err)
			}

			if !tc.shouldError && cfg.QueueType != tc.queueType {
				t.Errorf("Expected QueueType '%s', got '%s'", tc.queueType, cfg.QueueType)
			}
		})
	}
}

// TestValidateMaxFilesPerPoll validates MAX_FILES_PER_POLL handling
// ADR-003: Rate limiting to prevent overwhelming downstream
func TestValidateMaxFilesPerPoll(t *testing.T) {
	testCases := []struct {
		name     string
		value    string
		expected int
	}{
		{"no limit", "0", 0},
		{"limit 1", "1", 1},
		{"limit 50", "50", 50},
		{"limit 1000", "1000", 1000},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Clearenv()
			os.Setenv("MAX_FILES_PER_POLL", tc.value)

			cfg, err := Load()
			if err != nil {
				t.Fatalf("Expected successful load, got error: %v", err)
			}

			if cfg.MaxFilesPerPoll != tc.expected {
				t.Errorf("Expected MaxFilesPerPoll %d, got %d", tc.expected, cfg.MaxFilesPerPoll)
			}
		})
	}
}

// TestValidateFileSuffixFilter validates file suffix filter parsing
func TestValidateFileSuffixFilter(t *testing.T) {
	testCases := []struct {
		name     string
		value    string
		expected []string
	}{
		{"all files", "*", []string{}},
		{"single suffix", ".csv", []string{".csv"}},
		{"multiple suffixes", ".csv,.txt,.dat", []string{".csv", ".txt", ".dat"}},
		{"with spaces", ".csv, .txt, .dat", []string{".csv", ".txt", ".dat"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Clearenv()
			os.Setenv("FILE_SUFFIX_FILTER", tc.value)

			cfg, err := Load()
			if err != nil {
				t.Fatalf("Expected successful load, got error: %v", err)
			}

			if len(cfg.FileSuffixFilter) != len(tc.expected) {
				t.Fatalf("Expected %d suffixes, got %d", len(tc.expected), len(cfg.FileSuffixFilter))
			}

			for i, expected := range tc.expected {
				if cfg.FileSuffixFilter[i] != expected {
					t.Errorf("Expected suffix[%d] '%s', got '%s'", i, expected, cfg.FileSuffixFilter[i])
				}
			}
		})
	}
}

// TestValidateFilenamePattern validates regex pattern compilation
func TestValidateFilenamePattern(t *testing.T) {
	testCases := []struct {
		name        string
		pattern     string
		shouldError bool
	}{
		{"all files", ".*", false},
		{"prefix", "data.*", false},
		{"suffix", ".*\\.csv", false},
		{"invalid regex", "[", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Clearenv()
			os.Setenv("FILENAME_PATTERN", tc.pattern)

			cfg, err := Load()

			if tc.shouldError && err == nil {
				t.Error("Expected error for invalid regex, got success")
			}

			if !tc.shouldError && err != nil {
				t.Errorf("Expected success, got error: %v", err)
			}

			if !tc.shouldError && cfg.FilenamePattern == nil {
				t.Error("Expected compiled pattern, got nil")
			}
		})
	}
}

// TestValidateOutputType validates output type enum
// ADR-003: Only "file" or "queue" allowed
func TestValidateOutputType(t *testing.T) {
	testCases := []struct {
		name        string
		outputType  string
		shouldError bool
	}{
		{"file", "file", false},
		{"queue", "queue", false},
		{"invalid", "invalid", true},
		{"whitespace", "   ", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Clearenv()
			os.Setenv("OUTPUT_TYPE", tc.outputType)

			// For queue type, need valid queue config
			if tc.outputType == "queue" {
				os.Setenv("QUEUE_TYPE", "rabbitmq")
				os.Setenv("QUEUE_HOST", "localhost")
				os.Setenv("QUEUE_PORT", "5672")
				os.Setenv("QUEUE_NAME", "test")
			}

			_, err := Load()

			if tc.shouldError && err == nil {
				t.Error("Expected error for invalid output type, got success")
			}

			if !tc.shouldError && err != nil {
				t.Errorf("Expected success, got error: %v", err)
			}
		})
	}
}

// TestValidateArchiveSettings validates archive path configuration
func TestValidateArchiveSettings(t *testing.T) {
	os.Clearenv()
	os.Setenv("ARCHIVE_PROCESSED", "/custom/processed")
	os.Setenv("ARCHIVE_IGNORED", "/custom/ignored")
	os.Setenv("ARCHIVE_FAILED", "/custom/failed")
	os.Setenv("ARCHIVE_TIMESTAMP", "false")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Expected successful load, got error: %v", err)
	}

	if cfg.ArchiveProcessed != "/custom/processed" {
		t.Errorf("Expected ArchiveProcessed '/custom/processed', got '%s'", cfg.ArchiveProcessed)
	}

	if cfg.ArchiveIgnored != "/custom/ignored" {
		t.Errorf("Expected ArchiveIgnored '/custom/ignored', got '%s'", cfg.ArchiveIgnored)
	}

	if cfg.ArchiveFailed != "/custom/failed" {
		t.Errorf("Expected ArchiveFailed '/custom/failed', got '%s'", cfg.ArchiveFailed)
	}

	if cfg.ArchiveTimestamp != false {
		t.Error("Expected ArchiveTimestamp false, got true")
	}
}

// TestValidateDelimiterConfig validates delimiter configuration
func TestValidateDelimiterConfig(t *testing.T) {
	testCases := []struct {
		name      string
		delimiter string
		expected  rune
	}{
		{"comma", ",", ','},
		{"tab", "\t", '\t'},
		{"pipe", "|", '|'},
		{"semicolon", ";", ';'},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Clearenv()
			os.Setenv("DELIMITER", tc.delimiter)

			cfg, err := Load()
			if err != nil {
				t.Fatalf("Expected successful load, got error: %v", err)
			}

			if cfg.Delimiter != tc.expected {
				t.Errorf("Expected delimiter '%c', got '%c'", tc.expected, cfg.Delimiter)
			}
		})
	}
}

// TestValidateQueuePortRange validates port number range
// ADR-003: Port must be 1-65535
func TestValidateQueuePortRange(t *testing.T) {
	testCases := []struct {
		name        string
		port        string
		shouldError bool
	}{
		{"valid min", "1", false},
		{"valid mid", "5672", false},
		{"valid max", "65535", false},
		{"invalid zero", "0", true},
		{"invalid negative", "-1", true},
		{"invalid high", "65536", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Clearenv()
			os.Setenv("OUTPUT_TYPE", "queue")
			os.Setenv("QUEUE_TYPE", "rabbitmq")
			os.Setenv("QUEUE_HOST", "localhost")
			os.Setenv("QUEUE_PORT", tc.port)
			os.Setenv("QUEUE_NAME", "test")

			_, err := Load()

			if tc.shouldError && err == nil {
				t.Errorf("Expected error for port %s, got success", tc.port)
			}

			if !tc.shouldError && err != nil {
				t.Errorf("Expected success for port %s, got error: %v", tc.port, err)
			}
		})
	}
}

// TestConfigFailFast validates fail-fast behavior
// ADR-003: Invalid configuration must fail immediately with clear error
func TestConfigFailFast(t *testing.T) {
	os.Clearenv()
	os.Setenv("OUTPUT_TYPE", "queue")
	os.Setenv("QUEUE_TYPE", "invalid")

	_, err := Load()
	if err == nil {
		t.Fatal("Expected fail-fast error for invalid queue type, got success")
	}

	// Error message should be clear
	if err.Error() == "" {
		t.Error("Expected clear error message, got empty string")
	}
}
