package output

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"txt2json/internal/converter"
	"txt2json/internal/parser"
)

type FileHandler struct {
	outputFolder string
	converter    *converter.Converter
}

func NewFileHandler(outputFolder string) *FileHandler {
	return &FileHandler{
		outputFolder: outputFolder,
		converter:    converter.New(),
	}
}

func (h *FileHandler) Send(data []map[string]string, identifier string) error {
	// Generate output filename
	ext := filepath.Ext(identifier)
	base := identifier[:len(identifier)-len(ext)]
	outputFilename := base + ".json"
	outputPath := filepath.Join(h.outputFolder, outputFilename)

	// Marshal to JSON
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write to file
	if err := os.WriteFile(outputPath, jsonBytes, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

func (h *FileHandler) SendOrdered(result *parser.ParseResult, identifier string) error {
	// Generate output filename
	ext := filepath.Ext(identifier)
	base := identifier[:len(identifier)-len(ext)]
	outputFilename := base + ".json"
	outputPath := filepath.Join(h.outputFolder, outputFilename)

	// Convert to ordered JSON (preserves CSV column order per ADR-003)
	jsonBytes, err := h.converter.ToJSONOrdered(result)
	if err != nil {
		return fmt.Errorf("failed to marshal ordered JSON: %w", err)
	}

	// Write to file
	if err := os.WriteFile(outputPath, jsonBytes, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

func (h *FileHandler) Close() error {
	return nil
}
