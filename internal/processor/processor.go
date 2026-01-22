package processor

import (
	"fmt"
	"log"
	"path/filepath"

	"csv2json/internal/archiver"
	"csv2json/internal/config"
	"csv2json/internal/monitor"
	"csv2json/internal/output"
	"csv2json/internal/parser"
)

type Processor struct {
	config   *config.Config
	parser   *parser.Parser
	archiver *archiver.Archiver
	output   output.Handler
	monitor  *monitor.Monitor
}

func New(cfg *config.Config) (*Processor, error) {
	// Initialize components
	p := parser.New(cfg.Delimiter, cfg.QuoteChar, cfg.HasHeader)

	arch := archiver.New(
		cfg.ArchiveProcessed,
		cfg.ArchiveIgnored,
		cfg.ArchiveFailed,
		cfg.ArchiveTimestamp,
	)

	out, err := output.CreateHandler(
		cfg.OutputType,
		cfg.OutputFolder,
		cfg.QueueType,
		cfg.QueueHost,
		cfg.QueuePort,
		cfg.QueueName,
		cfg.QueueUsername,
		cfg.QueuePassword,
		cfg.LogQueueMessages,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create output handler: %w", err)
	}

	mon := monitor.New(cfg.InputFolder, cfg.PollInterval, cfg.MaxFilesPerPoll)

	return &Processor{
		config:   cfg,
		parser:   p,
		archiver: arch,
		output:   out,
		monitor:  mon,
	}, nil
}

func (p *Processor) Start() error {
	return p.monitor.Start(p.processFile)
}

func (p *Processor) Stop() {
	p.monitor.Stop()
	if err := p.output.Close(); err != nil {
		log.Printf("Error closing output handler: %v", err)
	}
}

func (p *Processor) processFile(filePath string) error {
	filename := filepath.Base(filePath)
	log.Printf("Processing file: %s", filename)

	// Check if file should be processed based on filters
	if !p.config.ShouldProcessFile(filename) {
		log.Printf("File does not match filters, ignoring: %s", filename)
		return p.archiver.Archive(filePath, archiver.CategoryIgnored, "")
	}

	// Validate file content
	if err := p.parser.Validate(filePath); err != nil {
		log.Printf("File validation failed: %v", err)
		return p.archiver.Archive(filePath, archiver.CategoryFailed, err.Error())
	}

	// Parse file (preserves CSV column order per ADR-003)
	result, err := p.parser.ParseWithOrder(filePath)
	if err != nil {
		log.Printf("Parsing failed: %v", err)
		return p.archiver.Archive(filePath, archiver.CategoryFailed, err.Error())
	}

	if len(result.Rows) == 0 {
		log.Printf("No data parsed from file: %s", filename)
		return p.archiver.Archive(filePath, archiver.CategoryFailed, "No data parsed")
	}

	log.Printf("Parsed %d rows from %s", len(result.Rows), filename)

	// Send output with ordered fields
	if err := p.output.SendOrdered(result, filename); err != nil {
		log.Printf("Output failed: %v", err)
		return p.archiver.Archive(filePath, archiver.CategoryFailed, err.Error())
	}

	// Archive as processed
	if err := p.archiver.Archive(filePath, archiver.CategoryProcessed, ""); err != nil {
		log.Printf("Failed to archive file: %v", err)
		return err
	}

	log.Printf("Successfully processed: %s", filename)
	return nil
}
