package monitor

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

// HybridMonitor combines event-driven and polling strategies
type HybridMonitor struct {
	watchFolder     string
	pollInterval    time.Duration
	maxFilesPerPoll int
	processedFiles  map[string]bool
	running         bool
	stopChan        chan struct{}
	watcher         *fsnotify.Watcher
}

// NewHybridMonitor creates a hybrid monitor with event-driven primary and polling backup
func NewHybridMonitor(watchFolder string, pollInterval time.Duration, maxFilesPerPoll int) (*HybridMonitor, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &HybridMonitor{
		watchFolder:     watchFolder,
		pollInterval:    pollInterval,
		maxFilesPerPoll: maxFilesPerPoll,
		processedFiles:  make(map[string]bool),
		stopChan:        make(chan struct{}),
		watcher:         watcher,
	}, nil
}

// Start begins hybrid monitoring (events + periodic polling backup)
func (m *HybridMonitor) Start(callback FileCallback) error {
	m.running = true

	// Initial scan to mark existing files as processed
	m.scanExisting()

	// Add watch on the folder
	if err := m.watcher.Add(m.watchFolder); err != nil {
		log.Printf("Failed to add watch on %s: %v", m.watchFolder, err)
		return err
	}

	log.Printf("Hybrid file monitor started (events + %v polling backup)", m.pollInterval)

	// Polling ticker for backup
	ticker := time.NewTicker(m.pollInterval)
	defer ticker.Stop()

	// Process events and periodic polls
	for {
		select {
		case event, ok := <-m.watcher.Events:
			if !ok {
				return nil
			}

			// Only care about Create and Write events
			if event.Has(fsnotify.Create) || event.Has(fsnotify.Write) {
				m.handleFileEvent(event.Name, callback)
			}

		case err, ok := <-m.watcher.Errors:
			if !ok {
				return nil
			}
			log.Printf("Watcher error: %v", err)

		case <-ticker.C:
			// Backup polling to catch any missed events
			if err := m.scanForNew(callback); err != nil {
				log.Printf("Error during backup scan: %v", err)
			}

		case <-m.stopChan:
			log.Println("Hybrid file monitor stopped")
			m.watcher.Close()
			return nil
		}
	}
}

// Stop terminates the hybrid monitor
func (m *HybridMonitor) Stop() {
	if m.running {
		close(m.stopChan)
		m.running = false
	}
}

func (m *HybridMonitor) scanExisting() {
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

func (m *HybridMonitor) handleFileEvent(filePath string, callback FileCallback) {
	// Extract filename
	filename := filepath.Base(filePath)

	// Skip directories
	info, err := os.Stat(filePath)
	if err != nil || info.IsDir() {
		return
	}

	// Skip already processed files
	if m.processedFiles[filename] {
		return
	}

	// Wait for file to be ready (not being written)
	if !m.isFileReady(filePath) {
		return
	}

	log.Printf("Detected new file (event): %s", filename)

	// Process file
	if err := callback(filePath); err != nil {
		log.Printf("Error processing %s: %v", filename, err)
	}

	// Mark as processed
	m.processedFiles[filename] = true
}

func (m *HybridMonitor) scanForNew(callback FileCallback) error {
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

		log.Printf("Detected new file (backup poll): %s", filename)

		// Process file
		if err := callback(filePath); err != nil {
			log.Printf("Error processing %s: %v", filename, err)
		}

		// Mark as processed
		m.processedFiles[filename] = true
		processedCount++
	}

	return nil
}

func (m *HybridMonitor) isFileReady(filePath string) bool {
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
