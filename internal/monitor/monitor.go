package monitor

import (
	"fmt"
	"log"
	"time"
)

// WatchMode defines the file detection strategy
type WatchMode string

const (
	WatchModeEvent  WatchMode = "event"
	WatchModePoll   WatchMode = "poll"
	WatchModeHybrid WatchMode = "hybrid"
)

// FileMonitor is the interface all monitor implementations must satisfy
type FileMonitor interface {
	Start(callback FileCallback) error
	Stop()
}

// NewMonitor creates the appropriate monitor based on watch mode
func NewMonitor(mode WatchMode, watchFolder string, pollInterval time.Duration, hybridPollInterval time.Duration, maxFilesPerPoll int) (FileMonitor, error) {
	switch mode {
	case WatchModeEvent:
		// Try event-driven, fallback to polling if it fails
		monitor, err := NewEventMonitor(watchFolder, maxFilesPerPoll)
		if err != nil {
			log.Printf("Warning: Failed to create event monitor (%v), falling back to polling", err)
			return NewPollingMonitor(watchFolder, pollInterval, maxFilesPerPoll), nil
		}
		return monitor, nil

	case WatchModePoll:
		return NewPollingMonitor(watchFolder, pollInterval, maxFilesPerPoll), nil

	case WatchModeHybrid:
		// Try hybrid, fallback to polling if it fails
		monitor, err := NewHybridMonitor(watchFolder, hybridPollInterval, maxFilesPerPoll)
		if err != nil {
			log.Printf("Warning: Failed to create hybrid monitor (%v), falling back to polling", err)
			return NewPollingMonitor(watchFolder, pollInterval, maxFilesPerPoll), nil
		}
		return monitor, nil

	default:
		return nil, fmt.Errorf("unsupported watch mode: %s (supported: event, poll, hybrid)", mode)
	}
}
