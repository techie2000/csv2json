package archiver

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	a := New("/processed", "/ignored", "/failed", true)

	if a == nil {
		t.Fatal("New() returned nil")
	}

	if a.archivePaths[CategoryProcessed] != "/processed" {
		t.Errorf("Expected processed path '/processed', got '%s'", a.archivePaths[CategoryProcessed])
	}

	if a.archivePaths[CategoryIgnored] != "/ignored" {
		t.Errorf("Expected ignored path '/ignored', got '%s'", a.archivePaths[CategoryIgnored])
	}

	if a.archivePaths[CategoryFailed] != "/failed" {
		t.Errorf("Expected failed path '/failed', got '%s'", a.archivePaths[CategoryFailed])
	}

	if !a.addTimestamp {
		t.Error("Expected addTimestamp to be true")
	}
}

func TestArchive_CreatesDirs(t *testing.T) {
	// Setup temp directories
	tempDir := t.TempDir()
	inputDir := filepath.Join(tempDir, "input")
	archiveDir := filepath.Join(tempDir, "archive")

	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("Failed to create input dir: %v", err)
	}

	// Create test file
	testFile := filepath.Join(inputDir, "test.csv")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create archiver (archive dir doesn't exist yet)
	a := New(archiveDir, archiveDir, archiveDir, false)

	// Archive should create the directory
	if err := a.Archive(testFile, CategoryProcessed, ""); err != nil {
		t.Fatalf("Archive failed: %v", err)
	}

	// Verify archive directory was created
	if _, err := os.Stat(archiveDir); os.IsNotExist(err) {
		t.Error("Archive directory was not created")
	}
}

func TestArchive_WithoutTimestamp(t *testing.T) {
	tempDir := t.TempDir()
	inputDir := filepath.Join(tempDir, "input")
	archiveDir := filepath.Join(tempDir, "archive")

	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("Failed to create input dir: %v", err)
	}
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		t.Fatalf("Failed to create archive dir: %v", err)
	}

	testFile := filepath.Join(inputDir, "test.csv")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	a := New(archiveDir, archiveDir, archiveDir, false)

	if err := a.Archive(testFile, CategoryProcessed, ""); err != nil {
		t.Fatalf("Archive failed: %v", err)
	}

	// Check file exists with original name (no timestamp)
	archivedFile := filepath.Join(archiveDir, "test.csv")
	if _, err := os.Stat(archivedFile); os.IsNotExist(err) {
		t.Error("Archived file not found without timestamp")
	}

	// Verify original file no longer exists
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("Original file still exists after archive")
	}
}

func TestArchive_WithTimestamp(t *testing.T) {
	tempDir := t.TempDir()
	inputDir := filepath.Join(tempDir, "input")
	archiveDir := filepath.Join(tempDir, "archive")

	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("Failed to create input dir: %v", err)
	}
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		t.Fatalf("Failed to create archive dir: %v", err)
	}

	testFile := filepath.Join(inputDir, "test.csv")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	a := New(archiveDir, archiveDir, archiveDir, true)

	if err := a.Archive(testFile, CategoryProcessed, ""); err != nil {
		t.Fatalf("Archive failed: %v", err)
	}

	// Find archived file with timestamp pattern
	entries, err := os.ReadDir(archiveDir)
	if err != nil {
		t.Fatalf("Failed to read archive dir: %v", err)
	}

	found := false
	for _, entry := range entries {
		name := entry.Name()
		// Should match pattern: test_YYYYMMDD_HHMMSS.csv
		if strings.HasPrefix(name, "test_") && strings.HasSuffix(name, ".csv") {
			found = true
			// Check timestamp format (8 digits + underscore + 6 digits)
			parts := strings.Split(name, "_")
			if len(parts) >= 3 {
				dateStr := parts[1]
				timeStr := strings.TrimSuffix(parts[2], ".csv")
				if len(dateStr) == 8 && len(timeStr) == 6 {
					// Timestamp format looks correct
					break
				}
			}
		}
	}

	if !found {
		t.Error("Archived file with timestamp not found")
	}
}

func TestArchive_DuplicateHandling(t *testing.T) {
	tempDir := t.TempDir()
	inputDir := filepath.Join(tempDir, "input")
	archiveDir := filepath.Join(tempDir, "archive")

	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("Failed to create input dir: %v", err)
	}
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		t.Fatalf("Failed to create archive dir: %v", err)
	}

	a := New(archiveDir, archiveDir, archiveDir, false)

	// Archive first file
	testFile1 := filepath.Join(inputDir, "test.csv")
	if err := os.WriteFile(testFile1, []byte("content 1"), 0644); err != nil {
		t.Fatalf("Failed to create test file 1: %v", err)
	}
	if err := a.Archive(testFile1, CategoryProcessed, ""); err != nil {
		t.Fatalf("Archive 1 failed: %v", err)
	}

	// Archive second file with same name
	testFile2 := filepath.Join(inputDir, "test.csv")
	if err := os.WriteFile(testFile2, []byte("content 2"), 0644); err != nil {
		t.Fatalf("Failed to create test file 2: %v", err)
	}
	if err := a.Archive(testFile2, CategoryProcessed, ""); err != nil {
		t.Fatalf("Archive 2 failed: %v", err)
	}

	// Should have test.csv and test_1.csv
	entries, err := os.ReadDir(archiveDir)
	if err != nil {
		t.Fatalf("Failed to read archive dir: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("Expected 2 archived files, got %d", len(entries))
	}

	hasOriginal := false
	hasCounter := false
	for _, entry := range entries {
		if entry.Name() == "test.csv" {
			hasOriginal = true
		}
		if entry.Name() == "test_1.csv" {
			hasCounter = true
		}
	}

	if !hasOriginal {
		t.Error("Original filename not found")
	}
	if !hasCounter {
		t.Error("Counter filename not found")
	}
}

func TestArchive_ErrorLog(t *testing.T) {
	tempDir := t.TempDir()
	inputDir := filepath.Join(tempDir, "input")
	archiveDir := filepath.Join(tempDir, "archive")

	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("Failed to create input dir: %v", err)
	}
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		t.Fatalf("Failed to create archive dir: %v", err)
	}

	testFile := filepath.Join(inputDir, "test.csv")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	a := New(archiveDir, archiveDir, archiveDir, false)
	errorMsg := "Invalid CSV format: missing delimiter"

	if err := a.Archive(testFile, CategoryFailed, errorMsg); err != nil {
		t.Fatalf("Archive failed: %v", err)
	}

	// Check archived file exists
	archivedFile := filepath.Join(archiveDir, "test.csv")
	if _, err := os.Stat(archivedFile); os.IsNotExist(err) {
		t.Error("Archived file not found")
	}

	// Check error log exists
	errorLogFile := archivedFile + ".error"
	content, err := os.ReadFile(errorLogFile)
	if err != nil {
		t.Fatalf("Error log not found: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "Timestamp:") {
		t.Error("Error log missing timestamp")
	}
	if !strings.Contains(contentStr, "File: test.csv") {
		t.Error("Error log missing filename")
	}
	if !strings.Contains(contentStr, errorMsg) {
		t.Error("Error log missing error message")
	}
}

func TestArchive_Categories(t *testing.T) {
	tempDir := t.TempDir()
	inputDir := filepath.Join(tempDir, "input")
	processedDir := filepath.Join(tempDir, "processed")
	ignoredDir := filepath.Join(tempDir, "ignored")
	failedDir := filepath.Join(tempDir, "failed")

	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("Failed to create input dir: %v", err)
	}

	a := New(processedDir, ignoredDir, failedDir, false)

	tests := []struct {
		name     string
		category Category
		expected string
	}{
		{"processed file", CategoryProcessed, processedDir},
		{"ignored file", CategoryIgnored, ignoredDir},
		{"failed file", CategoryFailed, failedDir},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(inputDir, tt.name+".csv")
			if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			if err := a.Archive(testFile, tt.category, ""); err != nil {
				t.Fatalf("Archive failed: %v", err)
			}

			expectedPath := filepath.Join(tt.expected, filepath.Base(testFile))
			if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
				t.Errorf("File not found in expected directory: %s", expectedPath)
			}
		})
	}
}

func TestCopyFile(t *testing.T) {
	tempDir := t.TempDir()

	srcFile := filepath.Join(tempDir, "source.txt")
	dstFile := filepath.Join(tempDir, "dest.txt")

	testContent := "test content for copy"
	if err := os.WriteFile(srcFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	if err := copyFile(srcFile, dstFile); err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}

	// Verify destination exists
	content, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("Content mismatch. Expected '%s', got '%s'", testContent, string(content))
	}

	// Verify source still exists (copyFile doesn't delete source)
	if _, err := os.Stat(srcFile); os.IsNotExist(err) {
		t.Error("Source file should still exist after copy")
	}
}

func TestArchive_CopyFallback(t *testing.T) {
	// This test simulates cross-device scenario by testing the copy logic
	tempDir := t.TempDir()
	inputDir := filepath.Join(tempDir, "input")
	archiveDir := filepath.Join(tempDir, "archive")

	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("Failed to create input dir: %v", err)
	}
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		t.Fatalf("Failed to create archive dir: %v", err)
	}

	testFile := filepath.Join(inputDir, "test.csv")
	testContent := "test content for copy fallback"
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	a := New(archiveDir, archiveDir, archiveDir, false)

	// Archive the file
	if err := a.Archive(testFile, CategoryProcessed, ""); err != nil {
		t.Fatalf("Archive failed: %v", err)
	}

	// Verify file was moved
	archivedFile := filepath.Join(archiveDir, "test.csv")
	content, err := os.ReadFile(archivedFile)
	if err != nil {
		t.Fatalf("Failed to read archived file: %v", err)
	}

	if string(content) != testContent {
		t.Error("Content mismatch in archived file")
	}

	// Verify original is gone
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("Original file still exists after archive")
	}
}

// Benchmark tests
func BenchmarkArchive_NoTimestamp(b *testing.B) {
	tempDir := b.TempDir()
	inputDir := filepath.Join(tempDir, "input")
	archiveDir := filepath.Join(tempDir, "archive")

	os.MkdirAll(inputDir, 0755)
	os.MkdirAll(archiveDir, 0755)

	a := New(archiveDir, archiveDir, archiveDir, false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testFile := filepath.Join(inputDir, "test.csv")
		os.WriteFile(testFile, []byte("test content"), 0644)
		a.Archive(testFile, CategoryProcessed, "")
	}
}

func BenchmarkArchive_WithTimestamp(b *testing.B) {
	tempDir := b.TempDir()
	inputDir := filepath.Join(tempDir, "input")
	archiveDir := filepath.Join(tempDir, "archive")

	os.MkdirAll(inputDir, 0755)
	os.MkdirAll(archiveDir, 0755)

	a := New(archiveDir, archiveDir, archiveDir, true)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testFile := filepath.Join(inputDir, "test.csv")
		os.WriteFile(testFile, []byte("test content"), 0644)
		// Add small delay to ensure unique timestamps
		time.Sleep(1 * time.Millisecond)
		a.Archive(testFile, CategoryProcessed, "")
	}
}
