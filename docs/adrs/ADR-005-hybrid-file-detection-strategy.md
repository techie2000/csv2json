# ADR 005: Hybrid File Detection Strategy (Event-Driven + Polling Fallback)

**Status:** Accepted  
**Date:** 2026-01-22  
**Decision Makers:** Development Team  
**Context:** csv2json File Polling Service  
**Enhances:** ADR-004 (Multi-Ingress Routing)

## Context and Problem Statement

The current implementation uses **time-based polling** (default 5 seconds) to detect new files. While simple and universally compatible, polling has significant drawbacks:

**Current Pain Points:**
- **Latency**: 5-second delay before file detection (or more if poll interval is increased)
- **Resource Waste**: Continuous CPU cycles checking empty directories
- **Not Scalable**: Higher polling frequency increases CPU load
- **Event Blind**: System cannot react immediately to file arrival

**Modern Reality:**
All major operating systems support **event-based file notifications**:
- **Linux**: inotify
- **macOS**: FSEvents
- **Windows**: ReadDirectoryChangesW

Go's `fsnotify` library provides cross-platform abstraction over these mechanisms.

**The Question:**
Should we evolve from time-based polling to event-driven file detection, while maintaining compatibility with edge cases where events aren't available?

## Decision Drivers

- **Performance**: Immediate detection (milliseconds vs seconds)
- **Efficiency**: Zero CPU overhead when no files are arriving
- **Scalability**: OS-level watching scales better than application polling
- **Compatibility**: Must work with network file systems that don't support events
- **Flexibility**: Different routes may have different requirements

## Options Considered

### Option 1: Keep Polling-Only (Current)

**Pros:**
- Simple implementation
- Works everywhere (local, NFS, SMB, cloud storage)
- Predictable resource usage
- Easy to reason about

**Cons:**
- Fixed latency (5+ seconds)
- Continuous CPU usage even when idle
- Not scalable with high file volumes
- Cannot react immediately to urgent files

### Option 2: Event-Only (Replace Polling)

**Pros:**
- Immediate detection (sub-second)
- Zero CPU when idle
- Highly scalable
- Modern and efficient

**Cons:**
- **BREAKS network file systems**: NFS/SMB often don't support inotify/FSEvents
- **BREAKS cloud mounts**: S3 FUSE, Azure Files, etc. may not emit events
- **BREAKS certain containerized scenarios**: Docker volume mounts may not propagate events
- No fallback for unsupported platforms

### Option 3: Hybrid Strategy (Event-Driven with Polling Fallback) ✅

**Architecture:**
```
┌─────────────────────────────────────┐
│ File Detection Strategy (per route) │
└─────────────────────────────────────┘
           │
    ┌──────┴──────┐
    ▼             ▼
┌────────┐   ┌─────────┐
│ Event  │   │ Polling │
│ Driver │   │ Fallback│
└────────┘   └─────────┘
    │             │
    └──────┬──────┘
           ▼
   File Processing Pipeline
```

**Three Modes:**

1. **Event Mode** (default):
   - Uses fsnotify for OS-level file system notifications
   - Immediate detection (typically <100ms)
   - Zero CPU when idle
   - Falls back to polling if event setup fails

2. **Polling Mode**:
   - Current implementation (time-based scanning)
   - Configurable interval (default 5s)
   - Guaranteed compatibility with all file systems
   - Higher latency but universal

3. **Hybrid Mode**:
   - Primary: Event-driven monitoring
   - Backup: Periodic polling (longer interval, e.g., 60s)
   - Catches events that fsnotify might miss (network glitches, etc.)
   - Best reliability with good performance

**Pros:**
- **Performance**: Event-driven by default for modern systems
- **Compatibility**: Polling fallback for network/cloud file systems
- **Flexibility**: Per-route configuration (some routes event, some poll)
- **Reliability**: Hybrid mode provides redundancy
- **Graceful Degradation**: Auto-falls back if events fail
- **Backward Compatible**: Existing configs work unchanged (default to event)

**Cons:**
- More complex implementation
- Two code paths to maintain
- Need to handle fsnotify errors gracefully
- Slightly larger binary size (fsnotify dependency)

## Decision Outcome

**Chosen Option:** **Hybrid Strategy** (Option 3)

### Rationale

The hybrid approach provides the best balance:

1. **Default Performance**: Event-driven detection gives immediate response for 95% of deployments
2. **Universal Compatibility**: Polling fallback ensures it works on ALL platforms
3. **Operational Flexibility**: Teams can choose per-route based on their infrastructure
4. **Future-Proof**: As cloud file systems improve event support, we automatically benefit
5. **Graceful Degradation**: If fsnotify fails (permissions, limits, etc.), system continues working

### Configuration Design

**routes.json (Multi-Ingress Mode):**
```json
{
  "routes": [
    {
      "name": "products",
      "input": {
        "folder": "/data/input/products",
        "watchMode": "event",           // "event", "poll", or "hybrid"
        "pollInterval": "5s",           // Used in poll/hybrid modes
        "hybridPollInterval": "60s"     // Backup polling in hybrid mode
      }
    }
  ]
}
```

**Environment Variables (Legacy Mode):**
```bash
# WATCH_MODE: event, poll, or hybrid
WATCH_MODE=event
POLL_INTERVAL_SECONDS=5
HYBRID_POLL_INTERVAL_SECONDS=60
```

**Default Behavior:**
- **watchMode**: `"event"` (with automatic fallback to polling if fsnotify fails)
- **pollInterval**: `5s` (used in poll mode or event fallback)
- **hybridPollInterval**: `60s` (used only in hybrid mode for backup polling)

### Implementation Strategy

**Phase 1: Core Event Monitor** ✅
- Add `fsnotify` dependency
- Create `internal/monitor/event_monitor.go` with fsnotify integration
- Refactor `internal/monitor/monitor.go` to support strategy pattern
- Add automatic fallback logic (event → poll if fsnotify fails)

**Phase 2: Configuration** ✅
- Add `WatchMode`, `PollInterval`, `HybridPollInterval` to routes.go InputConfig
- Update config.go for legacy mode watch mode settings
- Add validation for watch mode values

**Phase 3: Integration** ✅
- Update main.go to instantiate correct monitor based on watchMode
- Update README.md with watch mode documentation
- Update routes.json.example with watch mode examples

**Phase 4: Testing** ✅
- Unit tests for event monitor
- Integration tests for all three modes
- Verify fallback behavior when fsnotify unavailable

## Consequences

### Positive

1. **Immediate Detection**: Sub-second file detection in event mode (vs 5+ seconds)
2. **Resource Efficiency**: Near-zero CPU when idle in event mode
3. **Scalability**: Can handle high file volumes without polling overhead
4. **Flexibility**: Per-route configuration supports mixed environments
5. **Reliability**: Hybrid mode provides redundancy against event system failures
6. **Backward Compatible**: Existing deployments continue working (auto-upgrade to events)

### Negative

1. **Complexity**: Two monitoring strategies to maintain
2. **Dependency**: Adds fsnotify library (~150KB)
3. **Testing**: Need to validate all three modes across platforms
4. **Edge Cases**: fsnotify has OS-specific quirks (inotify limits on Linux, etc.)

### Mitigation

**Complexity Management:**
- Strategy pattern keeps code paths separate and testable
- Clear abstraction layer in Monitor interface
- Comprehensive unit tests for each mode

**Dependency Risk:**
- fsnotify is mature, widely-used, well-maintained
- Pure Go implementation (no CGO dependencies)
- Fallback to polling if library unavailable/broken

**OS Limits (Linux inotify):**
- Document inotify limits in README (default 8192 watches)
- Provide instructions for increasing: `fs.inotify.max_user_watches`
- Automatic fallback to polling if watch limit exceeded

**Event Reliability:**
- Hybrid mode catches missed events via periodic polling
- File readiness check still applies (2-second stability check)
- Graceful error handling with logging

## References

- **fsnotify Library**: https://github.com/fsnotify/fsnotify
- **Linux inotify**: https://man7.org/linux/man-pages/man7/inotify.7.html
- **macOS FSEvents**: https://developer.apple.com/documentation/coreservices/file_system_events
- **Windows ReadDirectoryChangesW**: https://docs.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-readdirectorychangesw
- **ADR-004**: Multi-Ingress Routing Architecture

## Revision History

- **2026-01-22**: Initial decision - Hybrid strategy accepted
