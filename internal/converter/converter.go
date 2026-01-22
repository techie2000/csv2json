package converter

import (
	"bytes"
	"csv2json/internal/parser"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Converter struct {
	indent string
}

func New() *Converter {
	return &Converter{
		indent: "  ",
	}
}

// ToJSON converts unordered maps to JSON (field order not preserved)
func (c *Converter) ToJSON(data []map[string]string) ([]byte, error) {
	jsonBytes, err := json.MarshalIndent(data, "", c.indent)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return jsonBytes, nil
}

// ToJSONOrdered converts ParseResult to JSON preserving CSV column order per ADR-003
func (c *Converter) ToJSONOrdered(result *parser.ParseResult) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("[\n")

	for i, row := range result.Rows {
		if i > 0 {
			buf.WriteString(",\n")
		}
		buf.WriteString(c.indent + "{\n")

		for j, key := range row.Keys {
			value := row.Values[key]
			// Escape JSON string
			valueJSON, err := json.Marshal(value)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal value: %w", err)
			}
			keyJSON, err := json.Marshal(key)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal key: %w", err)
			}

			buf.WriteString(c.indent + c.indent)
			buf.Write(keyJSON)
			buf.WriteString(": ")
			buf.Write(valueJSON)
			if j < len(row.Keys)-1 {
				buf.WriteString(",")
			}
			buf.WriteString("\n")
		}

		buf.WriteString(c.indent + "}")
	}

	buf.WriteString("\n]")
	return buf.Bytes(), nil
}

func (c *Converter) ToJSONFile(data []map[string]string, outputPath string) error {
	jsonBytes, err := c.ToJSON(data)
	if err != nil {
		return err
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if err := os.WriteFile(outputPath, jsonBytes, 0644); err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}

	return nil
}

func GetOutputFilename(inputFilename string) string {
	ext := filepath.Ext(inputFilename)
	base := inputFilename[:len(inputFilename)-len(ext)]
	return base + ".json"
}
