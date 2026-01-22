package monitor

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

type FileCallback func(string) error

// PollingMonitor uses time-based polling for file detection
type PollingMonitor struct {
	watchFolder     string
	pollInterval    time.Duration
	maxFilesPerPoll int
	processedFiles  map[string]bool
	running         bool
	stopChan        chan struct{}
}

// NewPollingMonitor creates a polling-based file monitor
func NewPollingMonitor(watchFolder string, pollInterval time.Duration, maxFilesPerPoll int) *PollingMonitor {
	return &PollingMonitor{
		watchFolder:     watchFolder,
		pollInterval:    pollInterval,
		maxFilesPerPoll: maxFilesPerPoll,
		processedFiles:  make(map[string]bool),
		stopChan:        make(chan struct{}),
	}
}

// Start begins polling-based monitoring
func (m *PollingMonitor) Start(callback FileCallback) error {
	m.running = true

	// Initial scan to mark existing files as processed
	m.scanExisting()

	log.Printf("Polling-based file monitor started. Polling every %v", m.pollInterval)

	ticker := time.NewTicker(m.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := m.scan(callback); err != nil {
				log.Printf("Error during scan: %v", err)
			}
		case <-m.stopChan:
			log.Println("Polling-based file monitor stopped")
			return nil
		}
	}
}

// Stop terminates the polling monitor
func (m *PollingMonitor) Stop() {
	if m.running {
		close(m.stopChan)
		m.running = false
	}
}

func (m *PollingMonitor) scanExisting() {
	entries, err := os.ReadDir(m.watchFolder)
	if err != nil {
		log.Printf("Warning: unable to scan watch folder: %v", err)
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			m.processedFiles[entry.Name()] = true
		}
	}

	log.Printf("Found %d existing files (will not process)", len(m.processedFiles))
}

func (m *PollingMonitor) scan(callback FileCallback) error {
	entries, err := os.ReadDir(m.watchFolder)
	if err != nil {
		return err
	}

	processedCount := 0

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Check max files per poll limit
		if m.maxFilesPerPoll > 0 && processedCount >= m.maxFilesPerPoll {
			log.Printf("Reached max files per poll limit (%d), remaining files will be processed in next cycle", m.maxFilesPerPoll)
			break
		}

		filename := entry.Name()
		if m.processedFiles[filename] {
			continue
		}

		filePath := filepath.Join(m.watchFolder, filename)

		// Check if file is ready (not being written)
		if !m.isFileReady(filePath) {
			continue
		}

		log.Printf("Detected new file: %s", filename)

		// Process file
		if err := callback(filePath); err != nil {
			log.Printf("Error processing %s: %v", filename, err)
		}

		// Mark as processed even if there was an error
		// (archiver will have moved it anyway)
		m.processedFiles[filename] = true
		processedCount++
	}

	return nil
}

func (m *PollingMonitor) isFileReady(filePath string) bool {
	info1, err := os.Stat(filePath)
	if err != nil {
		return false
	}

	time.Sleep(2 * time.Second)

	info2, err := os.Stat(filePath)
	if err != nil {
		return false
	}

	// If size hasn't changed, file is probably ready
	return info1.Size() == info2.Size()
}
