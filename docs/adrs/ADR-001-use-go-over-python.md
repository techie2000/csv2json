# ADR 001: Use Go Instead of Python for Always-On File Processing Service

**Status:** Accepted  
**Date:** 2026-01-20  
**Decision Makers:** Development Team  
**Context:** csv2json project

## Context and Problem Statement

We need to build a file polling service that runs continuously (24/7) to monitor directories, parse CSV/delimited files, convert them to JSON, and output to files or message queues. The service must be reliable, performant, and easy to deploy in production environments.

Two primary language options were considered: **Python** and **Go**.

## Decision Drivers

- **Operational Profile:** Always-on, long-running service (not ad-hoc or one-time scripts)
- **Performance:** Must efficiently handle continuous file I/O and polling
- **Resource Usage:** Memory footprint and CPU efficiency matter for cost optimization
- **Deployment:** Should be easy to distribute and run in various environments
- **Reliability:** Must run for extended periods without degradation
- **Concurrency:** Must handle multiple files and I/O operations efficiently
- **Maintenance:** Long-term maintainability and operational simplicity

## Options Considered

### Option 1: Python

**Pros:**

- ✅ Rapid prototyping and development
- ✅ Rich ecosystem for data processing (pandas, csv module)
- ✅ Easy to read and modify
- ✅ Good for ad-hoc scripts and one-time processing

**Cons:**

- ❌ Requires Python runtime installation on all target systems
- ❌ Global Interpreter Lock (GIL) limits true concurrency
- ❌ Higher memory footprint (~50-100MB+ for simple services)
- ❌ Performance degrades over long runtimes (memory leaks, GC pressure)
- ❌ Dependency management complexity (pip, virtualenv, version conflicts)
- ❌ Slower I/O operations compared to compiled languages
- ❌ Larger Docker images (~500MB base)

**Best For:** Ad-hoc scripts, data science notebooks, internal tooling used occasionally

### Option 2: Go

**Pros:**

- ✅ Compiled to single binary (no runtime dependencies)
- ✅ Superior performance: 10-50x faster for I/O-heavy workloads
- ✅ Low memory footprint (~10-20MB for simple services)
- ✅ Built-in concurrency with goroutines (lightweight threads)
- ✅ Predictable performance over extended runtime
- ✅ Fast startup time (<1 second)
- ✅ Cross-platform compilation (build for any OS from any OS)
- ✅ Excellent for long-running services
- ✅ Small Docker images (~15-20MB with Alpine)
- ✅ Built-in race detection and profiling tools
- ✅ Strong standard library for networking, file I/O, and concurrency

**Cons:**

- ❌ Slightly more verbose than Python
- ❌ Steeper learning curve for developers unfamiliar with compiled languages
- ❌ Less flexible for rapid experimentation

**Best For:** Production services, APIs, always-on daemons, high-performance systems

## Decision Outcome

**Chosen Option:** **Go**

### Rationale

Since this service will run **continuously in production (24/7)**, Go is the clear winner:

1. **Performance at Scale:** For a service processing files continuously, Go's I/O performance and low overhead means:
   - Faster file processing (less time per file)
   - Lower latency (quicker response to new files)
   - More throughput (handle more files per second)

2. **Resource Efficiency:**
   - 10x lower memory usage = cost savings in cloud environments
   - Predictable resource consumption over days/weeks of runtime
   - No GC pauses affecting responsiveness

3. **Operational Simplicity:**
   - Deploy single binary (no "works on my machine" issues)
   - No dependency hell or version conflicts
   - Trivial to containerize (tiny images)
   - Easy to distribute across multiple environments

4. **Reliability:**
   - Designed for long-running processes
   - Better error handling patterns
   - Built-in support for graceful shutdown
   - Memory safety without GC overhead

5. **Concurrency:**
   - Goroutines allow true parallel file processing
   - Efficient polling without blocking
   - Easy to add parallel processing as load grows

### Trade-offs Accepted

- **Development Time:** Go may take slightly longer to write initially, but the production benefits far outweigh this one-time cost
- **Team Familiarity:** Team must learn Go if unfamiliar, but Go's simplicity makes this investment worthwhile
- **Rapid Changes:** For this use case (stable file processing service), we don't need Python's rapid iteration benefits

## Performance Comparison

| Metric                | Python             | Go                     | Winner           |
|-----------------------|--------------------|------------------------|------------------|
| Startup Time          | ~500ms             | <50ms                  | Go (10x faster)  |
| Memory (Idle)         | ~50MB              | ~5MB                   | Go (10x better)  |
| File I/O Speed        | Baseline           | 10-50x faster          | Go               |
| Concurrent Processing | Limited (GIL)      | Excellent (goroutines) | Go               |
| Runtime Stability     | Degrades over time | Consistent             | Go               |
| Docker Image Size     | ~500MB             | ~15MB                  | Go (33x smaller) |
| Deployment            | Requires runtime   | Single binary          | Go               |

## Consequences

### Positive

- ✅ Production service will be fast, reliable, and resource-efficient
- ✅ Easier deployment and distribution
- ✅ Lower operational costs (compute, memory, storage)
- ✅ Better monitoring and profiling tools
- ✅ Scales effortlessly as file volume grows

### Negative

- ❌ Team needs to learn Go (if unfamiliar)
- ❌ Cannot leverage Python-specific libraries (but Go's stdlib is sufficient)

### Mitigation

- Provide Go training/resources for team
- Leverage Go's excellent documentation and tooling
- Use this project as a learning opportunity

## When Python Would Be Better

Python would be the right choice if:

- Service is run ad-hoc or infrequently (not 24/7)
- Rapid experimentation is more important than performance
- Complex data science libraries are required (pandas, numpy, scikit-learn)
- The service is internal tooling with low usage

For our use case (always-on production service), **Go is the optimal choice**.

## References

- [Go vs Python Performance Benchmarks](https://benchmarksgame-team.pages.debian.net/benchmarksgame/)
- [Why Go for Services](https://go.dev/solutions/cloud/)
- [Go Standard Library - CSV Package](https://pkg.go.dev/encoding/csv)
- [Goroutines vs Threads](https://go.dev/doc/effective_go#goroutines)

## Revision History

- **2026-01-20:** Initial decision - Choose Go over Python for production service
