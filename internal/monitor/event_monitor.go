package monitor

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

// EventMonitor uses fsnotify for event-driven file detection
type EventMonitor struct {
	watchFolder     string
	maxFilesPerPoll int
	processedFiles  map[string]bool
	running         bool
	stopChan        chan struct{}
	watcher         *fsnotify.Watcher
}

// NewEventMonitor creates an event-driven file monitor using fsnotify
func NewEventMonitor(watchFolder string, maxFilesPerPoll int) (*EventMonitor, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &EventMonitor{
		watchFolder:     watchFolder,
		maxFilesPerPoll: maxFilesPerPoll,
		processedFiles:  make(map[string]bool),
		stopChan:        make(chan struct{}),
		watcher:         watcher,
	}, nil
}

// Start begins event-driven monitoring
func (m *EventMonitor) Start(callback FileCallback) error {
	m.running = true

	// Initial scan to mark existing files as processed
	m.scanExisting()

	// Add watch on the folder
	if err := m.watcher.Add(m.watchFolder); err != nil {
		log.Printf("Failed to add watch on %s: %v", m.watchFolder, err)
		return err
	}

	log.Printf("Event-driven file monitor started on %s", m.watchFolder)

	// Process events
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

		case <-m.stopChan:
			log.Println("Event-driven file monitor stopped")
			m.watcher.Close()
			return nil
		}
	}
}

// Stop terminates the event monitor
func (m *EventMonitor) Stop() {
	if m.running {
		close(m.stopChan)
		m.running = false
	}
}

func (m *EventMonitor) scanExisting() {
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

func (m *EventMonitor) handleFileEvent(filePath string, callback FileCallback) {
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

	log.Printf("Detected new file: %s", filename)

	// Process file
	if err := callback(filePath); err != nil {
		log.Printf("Error processing %s: %v", filename, err)
	}

	// Mark as processed
	m.processedFiles[filename] = true
}

func (m *EventMonitor) isFileReady(filePath string) bool {
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
