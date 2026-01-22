package monitor

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

type FileCallback func(string) error

type Monitor struct {
	watchFolder    string
	pollInterval   time.Duration
	maxFilesPerPoll int
	processedFiles map[string]bool
	running        bool
	stopChan       chan struct{}
}

func New(watchFolder string, pollInterval time.Duration, maxFilesPerPoll int) *Monitor {
	return &Monitor{
		watchFolder:    watchFolder,
		pollInterval:   pollInterval,
		maxFilesPerPoll: maxFilesPerPoll,
		processedFiles: make(map[string]bool),
		stopChan:       make(chan struct{}),
	}
}

func (m *Monitor) Start(callback FileCallback) error {
	m.running = true

	// Initial scan to mark existing files as processed
	m.scanExisting()

	log.Printf("File monitor started. Polling every %v", m.pollInterval)

	ticker := time.NewTicker(m.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := m.scan(callback); err != nil {
				log.Printf("Error during scan: %v", err)
			}
		case <-m.stopChan:
			log.Println("File monitor stopped")
			return nil
		}
	}
}

func (m *Monitor) Stop() {
	if m.running {
		close(m.stopChan)
		m.running = false
	}
}

func (m *Monitor) scanExisting() {
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

func (m *Monitor) scan(callback FileCallback) error {
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

func (m *Monitor) isFileReady(filePath string) bool {
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
