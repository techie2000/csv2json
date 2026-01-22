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
	flag.Parse()

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

		// Set route context if using queue output
		if route.Output.Type == "queue" {
			proc.SetRouteName(route.Name, route.Output.IncludeRouteContext)
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
