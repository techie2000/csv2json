package archiver

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type Category string

const (
	CategoryProcessed Category = "processed"
	CategoryIgnored   Category = "ignored"
	CategoryFailed    Category = "failed"
)

type Archiver struct {
	archivePaths map[Category]string
	addTimestamp bool
}

func New(processed, ignored, failed string, addTimestamp bool) *Archiver {
	return &Archiver{
		archivePaths: map[Category]string{
			CategoryProcessed: processed,
			CategoryIgnored:   ignored,
			CategoryFailed:    failed,
		},
		addTimestamp: addTimestamp,
	}
}

func (a *Archiver) Archive(filePath string, category Category, errorMsg string) error {
	archiveDir := a.archivePaths[category]

	// Ensure archive directory exists
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		return fmt.Errorf("failed to create archive directory: %w", err)
	}

	// Generate archive filename
	filename := filepath.Base(filePath)
	var archiveName string

	if a.addTimestamp {
		timestamp := time.Now().Format("20060102_150405")
		ext := filepath.Ext(filename)
		base := filename[:len(filename)-len(ext)]
		archiveName = fmt.Sprintf("%s_%s%s", base, timestamp, ext)
	} else {
		archiveName = filename
	}

	archivePath := filepath.Join(archiveDir, archiveName)

	// Handle duplicate names
	counter := 1
	for {
		if _, err := os.Stat(archivePath); os.IsNotExist(err) {
			break
		}
		ext := filepath.Ext(filename)
		base := filename[:len(filename)-len(ext)]
		if a.addTimestamp {
			timestamp := time.Now().Format("20060102_150405")
			archiveName = fmt.Sprintf("%s_%s_%d%s", base, timestamp, counter, ext)
		} else {
			archiveName = fmt.Sprintf("%s_%d%s", base, counter, ext)
		}
		archivePath = filepath.Join(archiveDir, archiveName)
		counter++
	}

	// Move file (try rename first, fallback to copy+delete for cross-device links)
	if err := os.Rename(filePath, archivePath); err != nil {
		// Rename failed (likely cross-device link in Docker volumes)
		// Fallback to copy + delete
		if err := copyFile(filePath, archivePath); err != nil {
			return fmt.Errorf("failed to copy file to archive: %w", err)
		}
		if err := os.Remove(filePath); err != nil {
			return fmt.Errorf("failed to remove original file after copy: %w", err)
		}
	}

	// Create error log if error message provided
	if errorMsg != "" {
		if err := a.logError(archivePath, errorMsg); err != nil {
			// Log error but don't fail the archive operation
			fmt.Printf("Warning: failed to create error log: %v\n", err)
		}
	}

	return nil
}

func (a *Archiver) logError(archivePath, errorMsg string) error {
	errorLogPath := archivePath + ".error"

	content := fmt.Sprintf("Timestamp: %s\nFile: %s\nError: %s\n",
		time.Now().Format(time.RFC3339),
		filepath.Base(archivePath),
		errorMsg,
	)

	return os.WriteFile(errorLogPath, []byte(content), 0644)
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	// Sync to ensure data is written to disk
	return destFile.Sync()
}
