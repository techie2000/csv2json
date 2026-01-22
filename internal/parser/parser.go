package parser

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
)

// OrderedMap represents a map that preserves insertion order
type OrderedMap struct {
	Keys   []string
	Values map[string]string
}

// ParseResult contains the headers and data rows
type ParseResult struct {
	Headers []string
	Rows    []OrderedMap
}

type Parser struct {
	delimiter rune
	quoteChar rune
	hasHeader bool
}

func New(delimiter, quoteChar rune, hasHeader bool) *Parser {
	return &Parser{
		delimiter: delimiter,
		quoteChar: quoteChar,
		hasHeader: hasHeader,
	}
}

// Parse reads a CSV file and returns headers and ordered data rows
func (p *Parser) ParseWithOrder(filename string) (*ParseResult, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = p.delimiter
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	var headers []string
	var records []OrderedMap

	rowNum := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV record at row %d: %w", rowNum, err)
		}

		// First row handling
		if rowNum == 0 {
			if p.hasHeader {
				headers = record
			} else {
				// Generate column names: col_0, col_1, etc.
				for i := range record {
					headers = append(headers, fmt.Sprintf("col_%d", i))
				}
				// Process this row as data
				row := OrderedMap{
					Keys:   headers,
					Values: make(map[string]string),
				}
				for i, value := range record {
					row.Values[headers[i]] = value
				}
				records = append(records, row)
			}
		} else {
			// Subsequent rows
			if len(record) != len(headers) {
				return nil, fmt.Errorf("row %d has %d columns, expected %d", rowNum, len(record), len(headers))
			}

			row := OrderedMap{
				Keys:   headers,
				Values: make(map[string]string),
			}
			for i, value := range record {
				row.Values[headers[i]] = value
			}
			records = append(records, row)
		}

		rowNum++
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("no data rows found in file")
	}

	return &ParseResult{Headers: headers, Rows: records}, nil
}

// Parse maintains backward compatibility with old signature
func (p *Parser) Parse(filename string) ([]map[string]string, error) {
	result, err := p.ParseWithOrder(filename)
	if err != nil {
		return nil, err
	}

	// Convert OrderedMap to plain map for backward compatibility
	records := make([]map[string]string, len(result.Rows))
	for i, row := range result.Rows {
		records[i] = row.Values
	}
	return records, nil
}

func (p *Parser) Validate(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("cannot open file: %w", err)
	}
	defer file.Close()

	// Read first 4KB to validate content
	buf := make([]byte, 4096)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return fmt.Errorf("cannot read file: %w", err)
	}

	content := string(buf[:n])
	if !strings.Contains(content, string(p.delimiter)) {
		return fmt.Errorf("file does not appear to contain delimiter '%c'", p.delimiter)
	}

	return nil
}
