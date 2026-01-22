package monitor

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	watchFolder := "/watch"
	pollInterval := 5 * time.Second
	maxFiles := 10

	m := NewPollingMonitor(watchFolder, pollInterval, maxFiles)

	if m == nil {
		t.Fatal("New() returned nil")
	}

	if m.watchFolder != watchFolder {
		t.Errorf("Expected watchFolder '%s', got '%s'", watchFolder, m.watchFolder)
	}

	if m.pollInterval != pollInterval {
		t.Errorf("Expected pollInterval %v, got %v", pollInterval, m.pollInterval)
	}

	if m.maxFilesPerPoll != maxFiles {
		t.Errorf("Expected maxFilesPerPoll %d, got %d", maxFiles, m.maxFilesPerPoll)
	}

	if m.processedFiles == nil {
		t.Error("processedFiles map should be initialized")
	}
}

func TestScanExisting(t *testing.T) {
	tempDir := t.TempDir()

	// Create some existing files
	existingFiles := []string{"file1.csv", "file2.csv", "file3.csv"}
	for _, filename := range existingFiles {
		filePath := filepath.Join(tempDir, filename)
		if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	m := NewPollingMonitor(tempDir, 1*time.Second, 10)
	m.scanExisting()

	if len(m.processedFiles) != len(existingFiles) {
		t.Errorf("Expected %d processed files, got %d", len(existingFiles), len(m.processedFiles))
	}

	for _, filename := range existingFiles {
		if !m.processedFiles[filename] {
			t.Errorf("File '%s' should be marked as processed", filename)
		}
	}
}

func TestScanExisting_IgnoresDirs(t *testing.T) {
	tempDir := t.TempDir()

	// Create files and directories
	if err := os.WriteFile(filepath.Join(tempDir, "file.csv"), []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tempDir, "subdir"), 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	m := NewPollingMonitor(tempDir, 1*time.Second, 10)
	m.scanExisting()

	if len(m.processedFiles) != 1 {
		t.Errorf("Expected 1 processed file, got %d", len(m.processedFiles))
	}

	if !m.processedFiles["file.csv"] {
		t.Error("file.csv should be marked as processed")
	}
}

func TestMaxFilesPerPoll(t *testing.T) {
	tempDir := t.TempDir()
	maxFiles := 3

	m := NewPollingMonitor(tempDir, 1*time.Second, maxFiles)
	m.running = true

	// Create more files than maxFiles limit
	totalFiles := 5
	for i := 0; i < totalFiles; i++ {
		filename := filepath.Join(tempDir, "file_"+string(rune('A'+i))+".csv")
		if err := os.WriteFile(filename, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	processedCount := 0
	var mu sync.Mutex

	callback := func(path string) error {
		mu.Lock()
		processedCount++
		mu.Unlock()
		return nil
	}

	// First scan should process maxFiles
	if err := m.scan(callback); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if processedCount != maxFiles {
		t.Errorf("Expected %d files processed in first scan, got %d", maxFiles, processedCount)
	}

	// Second scan should process remaining files
	processedCount = 0
	if err := m.scan(callback); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	expectedRemaining := totalFiles - maxFiles
	if processedCount != expectedRemaining {
		t.Errorf("Expected %d files processed in second scan, got %d", expectedRemaining, processedCount)
	}
}

func TestMaxFilesPerPoll_Zero(t *testing.T) {
	tempDir := t.TempDir()
	maxFiles := 0 // Zero means no limit

	m := NewPollingMonitor(tempDir, 1*time.Second, maxFiles)
	m.running = true

	// Create multiple files
	totalFiles := 10
	for i := 0; i < totalFiles; i++ {
		filename := filepath.Join(tempDir, "file_"+string(rune('A'+i))+".csv")
		if err := os.WriteFile(filename, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	processedCount := 0
	var mu sync.Mutex

	callback := func(path string) error {
		mu.Lock()
		processedCount++
		mu.Unlock()
		return nil
	}

	// Should process all files in one scan (no limit)
	if err := m.scan(callback); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if processedCount != totalFiles {
		t.Errorf("Expected all %d files processed, got %d", totalFiles, processedCount)
	}
}

func TestScan_SkipsProcessedFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create initial file
	file1 := filepath.Join(tempDir, "file1.csv")
	if err := os.WriteFile(file1, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}

	m := NewPollingMonitor(tempDir, 1*time.Second, 10)
	m.running = true

	processedCount := 0
	callback := func(path string) error {
		processedCount++
		return nil
	}

	// First scan processes file1
	if err := m.scan(callback); err != nil {
		t.Fatalf("First scan failed: %v", err)
	}

	if processedCount != 1 {
		t.Errorf("Expected 1 file processed, got %d", processedCount)
	}

	// Second scan should not process file1 again
	processedCount = 0
	if err := m.scan(callback); err != nil {
		t.Fatalf("Second scan failed: %v", err)
	}

	if processedCount != 0 {
		t.Errorf("Expected 0 files processed in second scan, got %d", processedCount)
	}
}

func TestScan_ProcessesNewFiles(t *testing.T) {
	tempDir := t.TempDir()

	m := NewPollingMonitor(tempDir, 1*time.Second, 10)
	m.running = true

	processedFiles := []string{}
	var mu sync.Mutex

	callback := func(path string) error {
		mu.Lock()
		processedFiles = append(processedFiles, filepath.Base(path))
		mu.Unlock()
		return nil
	}

	// Create file1 and scan
	file1 := filepath.Join(tempDir, "file1.csv")
	if err := os.WriteFile(file1, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}

	if err := m.scan(callback); err != nil {
		t.Fatalf("First scan failed: %v", err)
	}

	// Create file2 and scan
	file2 := filepath.Join(tempDir, "file2.csv")
	if err := os.WriteFile(file2, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	if err := m.scan(callback); err != nil {
		t.Fatalf("Second scan failed: %v", err)
	}

	// Should have processed both files (one in each scan)
	if len(processedFiles) != 2 {
		t.Errorf("Expected 2 files processed total, got %d", len(processedFiles))
	}

	hasFile1 := false
	hasFile2 := false
	for _, f := range processedFiles {
		if f == "file1.csv" {
			hasFile1 = true
		}
		if f == "file2.csv" {
			hasFile2 = true
		}
	}

	if !hasFile1 {
		t.Error("file1.csv was not processed")
	}
	if !hasFile2 {
		t.Error("file2.csv was not processed")
	}
}

func TestScan_CallbackError(t *testing.T) {
	tempDir := t.TempDir()

	file1 := filepath.Join(tempDir, "file1.csv")
	if err := os.WriteFile(file1, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}

	m := NewPollingMonitor(tempDir, 1*time.Second, 10)
	m.running = true

	// Callback that returns an error should not stop processing
	callbackCalled := false
	callback := func(path string) error {
		callbackCalled = true
		return nil // Scan continues even if callback has issues
	}

	if err := m.scan(callback); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if !callbackCalled {
		t.Error("Callback was not called")
	}
}

func TestScan_IgnoresDirectories(t *testing.T) {
	tempDir := t.TempDir()

	// Create file and directory
	file1 := filepath.Join(tempDir, "file1.csv")
	if err := os.WriteFile(file1, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}

	subDir := filepath.Join(tempDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	m := NewPollingMonitor(tempDir, 1*time.Second, 10)
	m.running = true

	processedCount := 0
	callback := func(path string) error {
		processedCount++
		return nil
	}

	if err := m.scan(callback); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Should only process the file, not the directory
	if processedCount != 1 {
		t.Errorf("Expected 1 file processed, got %d", processedCount)
	}
}

func TestStop(t *testing.T) {
	tempDir := t.TempDir()
	m := NewPollingMonitor(tempDir, 100*time.Millisecond, 10)

	// Start monitor in goroutine
	var wg sync.WaitGroup
	wg.Add(1)

	callback := func(path string) error {
		return nil
	}

	go func() {
		defer wg.Done()
		m.Start(callback)
	}()

	// Let it run for a bit
	time.Sleep(200 * time.Millisecond)

	// Stop should return quickly
	m.Stop()

	// Wait for Start to return
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Stop worked
	case <-time.After(2 * time.Second):
		t.Error("Stop did not terminate the monitor in reasonable time")
	}
}

// Benchmark tests
func BenchmarkScan_SmallFiles(b *testing.B) {
	tempDir := b.TempDir()

	// Create 10 small files
	for i := 0; i < 10; i++ {
		filename := filepath.Join(tempDir, "file_"+string(rune('A'+i))+".csv")
		os.WriteFile(filename, []byte("test"), 0644)
	}

	m := NewPollingMonitor(tempDir, 1*time.Second, 0)
	m.running = true

	callback := func(path string) error {
		return nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset processed files for each iteration
		m.processedFiles = make(map[string]bool)
		m.scan(callback)
	}
}

func BenchmarkScan_WithLimit(b *testing.B) {
	tempDir := b.TempDir()

	// Create 100 files
	for i := 0; i < 100; i++ {
		filename := filepath.Join(tempDir, "file_"+string(rune('A'+i%26))+string(rune('0'+i/26))+".csv")
		os.WriteFile(filename, []byte("test"), 0644)
	}

	m := NewPollingMonitor(tempDir, 1*time.Second, 10)
	m.running = true

	callback := func(path string) error {
		return nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.processedFiles = make(map[string]bool)
		m.scan(callback)
	}
}
