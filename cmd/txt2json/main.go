package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"csv2json/internal/config"
	"csv2json/internal/processor"
)

func main() {
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

	// Initialize processor
	proc, err := processor.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize processor: %v", err)
	}

	// Log startup configuration
	log.Println("========================================")
	log.Println("txt2json service starting with configuration:")
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
