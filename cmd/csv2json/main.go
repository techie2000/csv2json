package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"csv2json/internal/config"
	"csv2json/internal/processor"
	"csv2json/internal/version"
)

func main() {
	// Parse command-line flags
	versionFlag := flag.Bool("version", false, "Display version information")
	helpFlag := flag.Bool("help", false, "Display usage information")
	flag.Parse()

	// Handle help flag
	if *helpFlag {
		printHelp()
		os.Exit(0)
	}

	// Handle version flag
	if *versionFlag {
		fmt.Println(version.GetFullVersionInfo())
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize log file
	if cfg.LogFile != "" {
		// Create log directory if it doesn't exist
		logDir := filepath.Dir(cfg.LogFile)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			log.Fatalf("Failed to create log directory: %v", err)
		}

		// Open log file
		logFile, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("Failed to open log file: %v", err)
		}
		defer logFile.Close()

		// Write to both stdout and log file
		multiWriter := io.MultiWriter(os.Stdout, logFile)
		log.SetOutput(multiWriter)
		log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	}

	// Check if using multi-ingress routing mode
	if cfg.RoutesConfigPath != "" {
		log.Printf("Starting in MULTI-INGRESS ROUTING mode with config: %s", cfg.RoutesConfigPath)
		runMultiIngressMode(cfg.RoutesConfigPath)
	} else {
		log.Println("Starting in LEGACY SINGLE-INPUT mode")
		runLegacyMode(cfg)
	}
}

// runLegacyMode runs the service in single-input mode (original behavior)
func runLegacyMode(cfg *config.Config) {
	// Initialize processor
	proc, err := processor.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize processor: %v", err)
	}

	// Log startup configuration
	log.Println("========================================")
	log.Printf("%s", version.GetFullVersionInfo())
	log.Println("csv2json service starting with configuration:")
	log.Println("========================================")
	log.Printf("INPUT_FOLDER: %s", cfg.InputFolder)
	log.Printf("POLL_INTERVAL: %v", cfg.PollInterval)
	log.Printf("MAX_FILES_PER_POLL: %d", cfg.MaxFilesPerPoll)
	if len(cfg.FileSuffixFilter) > 0 {
		log.Printf("FILE_SUFFIX_FILTER: %v", cfg.FileSuffixFilter)
	} else {
		log.Println("FILE_SUFFIX_FILTER: * (all files)")
	}
	log.Printf("FILENAME_PATTERN: %s", cfg.FilenamePattern.String())
	log.Printf("DELIMITER: %q", cfg.Delimiter)
	log.Printf("QUOTECHAR: %q", cfg.QuoteChar)
	log.Printf("ENCODING: %s", cfg.Encoding)
	log.Printf("HAS_HEADER: %t", cfg.HasHeader)
	log.Printf("OUTPUT_TYPE: %s", cfg.OutputType)
	if cfg.OutputType == "file" {
		log.Printf("OUTPUT_FOLDER: %s", cfg.OutputFolder)
	} else {
		log.Printf("QUEUE_TYPE: %s", cfg.QueueType)
		log.Printf("QUEUE_HOST: %s", cfg.QueueHost)
		log.Printf("QUEUE_PORT: %d", cfg.QueuePort)
		log.Printf("QUEUE_NAME: %s", cfg.QueueName)
		if cfg.QueueUsername != "" {
			log.Printf("QUEUE_USERNAME: %s", cfg.QueueUsername)
		}
		log.Printf("LOG_QUEUE_MESSAGES: %t", cfg.LogQueueMessages)
	}
	log.Printf("ARCHIVE_PROCESSED: %s", cfg.ArchiveProcessed)
	log.Printf("ARCHIVE_IGNORED: %s", cfg.ArchiveIgnored)
	log.Printf("ARCHIVE_FAILED: %s", cfg.ArchiveFailed)
	log.Printf("ARCHIVE_TIMESTAMP: %t", cfg.ArchiveTimestamp)
	log.Printf("LOG_LEVEL: %s", cfg.LogLevel)
	log.Printf("LOG_FILE: %s", cfg.LogFile)
	log.Println("========================================")

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start processor in goroutine
	go func() {
		if err := proc.Start(); err != nil {
			log.Fatalf("Processor error: %v", err)
		}
	}()

	log.Println("Service ready. Monitoring for new files. Press Ctrl+C to stop.")

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutdown signal received, stopping gracefully...")

	proc.Stop()
	log.Println("Service stopped")
}

// runMultiIngressMode runs the service in multi-ingress routing mode (ADR-004)
func runMultiIngressMode(routesConfigPath string) {
	// Load routes configuration
	routesConfig, err := config.LoadRoutes(routesConfigPath)
	if err != nil {
		log.Fatalf("Failed to load routes configuration: %v", err)
	}

	if len(routesConfig.Routes) == 0 {
		log.Fatal("No routes configured in routes.json")
	}

	log.Printf("Loaded %d route(s) from configuration", len(routesConfig.Routes))

	// Create a processor for each route
	processors := make([]*processor.Processor, 0, len(routesConfig.Routes))

	for i, route := range routesConfig.Routes {
		log.Printf("Initializing route %d/%d: %s", i+1, len(routesConfig.Routes), route.Name)

		// Convert route to legacy config
		routeCfg := route.ToLegacyConfig()

		// Initialize processor for this route
		proc, err := processor.New(routeCfg)
		if err != nil {
			log.Fatalf("Failed to initialize processor for route '%s': %v", route.Name, err)
		}

		// Set envelope context for queue output (ADR-006)
		if route.Output.Type == "queue" || route.Output.Type == "both" {
			includeEnvelope := true // Default
			if route.Output.IncludeEnvelope != nil {
				includeEnvelope = *route.Output.IncludeEnvelope
			}
			proc.SetEnvelopeContext(route.Name, route.IngestionContract, includeEnvelope)
		}

		processors = append(processors, proc)

		// Log route configuration
		log.Println("----------------------------------------")
		log.Printf("Route: %s", route.Name)
		log.Printf("  Input: %s", route.Input.Path)
		log.Printf("  Output: %s -> %s", route.Output.Type, route.Output.Destination)
		if route.Input.FilenamePattern != "" {
			log.Printf("  Pattern: %s", route.Input.FilenamePattern)
		}
		log.Printf("  PollInterval: %ds", route.Input.PollIntervalSec)
		log.Println("----------------------------------------")
	}

	// Log startup summary
	log.Println("========================================")
	log.Printf("%s", version.GetFullVersionInfo())
	log.Printf("Multi-Ingress Routing Mode: %d active routes", len(processors))
	log.Println("========================================")

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start all processors in goroutines
	for i, proc := range processors {
		routeName := routesConfig.Routes[i].Name
		go func(p *processor.Processor, name string) {
			log.Printf("Starting route processor: %s", name)
			if err := p.Start(); err != nil {
				log.Printf("ERROR: Route '%s' processor failed: %v", name, err)
			}
		}(proc, routeName)
	}

	log.Println("All routes active. Monitoring for new files. Press Ctrl+C to stop.")

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutdown signal received, stopping all routes gracefully...")

	// Stop all processors
	for i, proc := range processors {
		routeName := routesConfig.Routes[i].Name
		log.Printf("Stopping route: %s", routeName)
		proc.Stop()
	}

	log.Println("All routes stopped. Service shutdown complete.")
}

// printHelp displays comprehensive usage information
func printHelp() {
	fmt.Printf(`%s

DESCRIPTION:
    csv2json is a high-performance CSV-to-JSON conversion service that monitors
    directories for incoming CSV files, converts them to JSON, and outputs to
    files or message queues. Supports both single-input and multi-ingress routing
    modes with event-driven or polling-based file detection.

USAGE:
    csv2json [OPTIONS]

OPTIONS:
    --help              Display this help information
    --version           Display version information and exit

OPERATIONAL MODES:
    The service operates in one of two modes based on configuration:

    1. LEGACY SINGLE-INPUT MODE (default)
       - Monitors a single input directory
       - Configured via environment variables
       - Suitable for simple, single-source scenarios

    2. MULTI-INGRESS ROUTING MODE (ADR-004)
       - Monitors multiple input directories with independent configurations
       - Configured via routes.json file
       - Suitable for multi-source, multi-destination scenarios
       - Set ROUTES_CONFIG environment variable to enable

CONFIGURATION:
    All configuration is managed through environment variables or routes.json.

    Key Environment Variables (Legacy Mode):
        ROUTES_CONFIG              Path to routes.json (enables Multi-Ingress Mode)
        INPUT_FOLDER               Directory to monitor (default: ./input)
        WATCH_MODE                 File detection: event|poll|hybrid (default: event)
        POLL_INTERVAL_SECONDS      Polling interval (default: 5)
        OUTPUT_TYPE                Output: file|queue (default: file)
        OUTPUT_FOLDER              JSON output directory (default: ./output)
        QUEUE_TYPE                 Queue system: rabbitmq (default)
        QUEUE_HOST                 Queue server host (default: localhost)
        QUEUE_PORT                 Queue server port (default: 5672)
        QUEUE_NAME                 Queue name (required for queue mode)
        HAS_HEADER                 CSV has header row (default: true)
        DELIMITER                  Field delimiter (default: ,)
        ARCHIVE_PROCESSED          Archive for processed files
        ARCHIVE_FAILED             Archive for failed files
        LOG_LEVEL                  Logging level (default: INFO)

    See .env.example for complete configuration options.

FILE DETECTION STRATEGIES (ADR-005):
    event (default, recommended):
        - Uses OS-level file system notifications (fsnotify)
        - Immediate detection (~100ms latency)
        - Zero CPU overhead when idle
        - Automatically falls back to polling if unavailable

    poll (legacy compatibility):
        - Time-based directory scanning
        - Works with all file systems (NFS, SMB, cloud mounts)
        - Higher latency (5+ seconds)
        - Continuous CPU usage

    hybrid (maximum reliability):
        - Primary: Event-driven monitoring
        - Backup: Periodic polling (default: 60s)
        - Best for critical systems requiring redundancy

EXAMPLES:
    # Display version
    csv2json --version

    # Run with default configuration (.env or environment variables)
    csv2json

    # Run in Multi-Ingress Routing Mode
    export ROUTES_CONFIG=./routes.json
    csv2json

    # Run with custom poll interval
    export POLL_INTERVAL_SECONDS=10
    csv2json

    # Run with queue output
    export OUTPUT_TYPE=queue
    export QUEUE_NAME=my_queue
    export QUEUE_HOST=rabbitmq.example.com
    csv2json

ARCHITECTURE DECISION RECORDS:
    ADR-001: Use Go over Python for performance and concurrency
    ADR-002: Use RabbitMQ for message queuing (extensible to Kafka, SQS, etc.)
    ADR-003: Core system principles (string values, empty strings, array structure)
    ADR-004: Multi-ingress routing architecture for multi-source scenarios
    ADR-005: Hybrid file detection strategy (event-driven with polling fallback)

DOCUMENTATION:
    README.md           Comprehensive documentation
    TESTING.md          Testing strategy and guidelines
    docs/adrs/          Architecture decision records
    routes.json.example Example multi-ingress configuration
    .env.example        Complete environment variable reference

PROJECT:
    Repository: github.com/techie2000/csv2json (pending first push)
    License:    [To be added]
    Version:    %s

`, version.GetVersionInfo(), version.GetVersionInfo())
}
