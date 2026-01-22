# ADR 003: Core System Principles and Behavior Contract

**Status:** Accepted  
**Date:** 2026-01-20  
**Decision Makers:** Development Team  
**Context:** txt2json File Polling Service

## Context and Problem Statement

This service is **pure plumbing** - it moves data from one place to another with minimal transformation. To prevent feature creep, scope drift, and complexity explosion, we must establish **non-negotiable core principles** that define what this system is and is not.

**What this system is:**

- A file-to-JSON conversion pipeline
- A routing and archiving mechanism
- A reliable, predictable data transport

**What this system is NOT:**

- Business logic processor
- Data enrichment engine
- Schema transformer
- "Just one more feature" collector

## Decision Drivers

- **Predictability**: Behavior must be deterministic and well-defined
- **Maintainability**: Simple rules are easy to understand and debug
- **Reliability**: Strict validation prevents silent failures
- **Clarity**: No ambiguity about file handling or outcomes
- **Scope Control**: Prevent feature creep and complexity growth

## Core Principles (Non-Negotiable)

### 1. Three-Step Pipeline

Every execution follows this exact sequence:

1. **Poll directory for files**
2. **Convert eligible CSV → JSON**
3. **Route outputs and archive inputs based on outcome**

No preprocessing, no post-processing, no side effects, no exceptions.

### 2. File Eligibility Rules

A file is considered for processing if and only if:

- ✅ It is a **regular file** (not a directory, symlink, device, or temp file)
- ✅ It **matches optional filename pattern** (regex)
- ✅ It **matches optional suffix filter** (e.g., `.csv`, `.txt`)
- ✅ It is **not locked or being written** (no `.processing` lock, no active writes)

**Defaults:**

- If no filename pattern specified → **accept all files**
- If no suffix specified → **accept all files**

**Filter Technology:** Use **regex, not globbing**. Regex scales; globs grow teeth.

### 3. CSV Expectations (Strict, Not Sorry)

The system expects:

- **Encoding**: UTF-8 only
- **Header row**: Optional via `HAS_HEADER` configuration
  - When `HAS_HEADER=true`: First row defines JSON keys
  - When `HAS_HEADER=false`: Auto-generates column names (`col_0`, `col_1`, etc.)
- **Delimiter**: Comma (or configured separator) - NO auto-detection
- **Quote fields**: Allowed (standard CSV escaping)
- **Empty fields**: Allowed (empty string, not null)

**If row column counts don't match → FAILED, not ignored**

**Rationale**: Auto-detecting CSV format is a trap. We are strict to prevent silent errors, misinterpretations, and debugging nightmares.

### 4. Every File Has Exactly One Outcome

No file exists in limbo. Every file ends in **exactly one** of these states:

| Outcome        | Definition                                               | Reason                                                   |
| -------------- | -------------------------------------------------------- | -------------------------------------------------------- |
| **Processed**  | Valid CSV, successfully converted, output delivered      | File matched filters, parsed correctly, output succeeded |
| **Ignored**    | File does not match filters (name/suffix)                | Intentionally skipped based on configuration             |
| **Failed**     | Matches filters but CSV cannot be parsed OR output fails | Validation error, parsing error, or delivery failure     |

**Files are MOVED, not copied.** The inbound folder is a queue, not a museum.

### 5. Exactly One Output Mode Active

Output modes are **mutually exclusive**. The system operates in one mode at a time:

#### File Output Mode

**Behavior:**

- One input CSV → One output JSON file
- JSON is an array of objects (header keys → row values)
- Preserve row order exactly
- No schema mutation, no transformations

**JSON Structure Contract:**

```json
[
  {
    "header1": "value1",
    "header2": "value2",
    "headerN": "valueN"
  }
]
```

**Properties:**

- **Top-level**: Array (always, even for single row)
- **Array items**: Objects with string keys and string values (cell data)
- **Key names**:
  - When `HAS_HEADER=true`: Exact CSV header text (case-sensitive, whitespace preserved)
  - When `HAS_HEADER=false`: Auto-generated (`col_0`, `col_1`, `col_2`, etc.)
- **Values**: Always strings (no type coercion: "30" not 30, "true" not true)
- **Empty fields**: Empty string `""` (not null)
- **Order**: Row order matches CSV row order

**Example Input (CSV):**

```csv
name,age,email
John,30,john@example.com
Jane,25,jane@example.com
```

**Example Output (JSON):**

```json
[
  {"name": "John", "age": "30", "email": "john@example.com"},
  {"name": "Jane", "age": "25", "email": "jane@example.com"}
]
```

#### Queue Output Mode

**Behavior:**

- Each CSV row → One JSON message
- Headers become object keys
- File metadata included for lineage

**JSON Structure Contract:**

```json
{
  "sourceFile": "string",
  "rowNumber": "integer",
  "timestamp": "ISO8601 string",
  "payload": {
    "header1": "value1",
    "headerN": "valueN"
  }
}
```

**Properties:**

- **sourceFile**: Original filename (string)
- **rowNumber**: 1-based row index (integer, excludes header row)
- **timestamp**: Processing time in ISO8601 format (e.g., `2026-01-21T10:15:23Z`)
- **payload**: Object with CSV data (same rules as File Output Mode)
  - Keys: CSV headers (case-sensitive)
  - Values: Always strings
  - Empty fields: Empty string `""`

**Example Input (CSV):**

```csv
name,age,email
John,30,john@example.com
Jane,25,jane@example.com
```

**Example Output (2 messages):**

Message 1:

```json
{
  "sourceFile": "data.csv",
  "rowNumber": 1,
  "timestamp": "2026-01-21T10:15:23Z",
  "payload": {"name": "John", "age": "30", "email": "john@example.com"}
}
```

Message 2:

```json
{
  "sourceFile": "data.csv",
  "rowNumber": 2,
  "timestamp": "2026-01-21T10:15:23Z",
  "payload": {"name": "Jane", "age": "25", "email": "jane@example.com"}
}
```

**Rationale for metadata:** When something breaks downstream, you need lineage to trace back to source. This matters.

### 6. Polling & Safety Rules

**Atomic file handling:**

- Use atomic file moves OR lock files (`.processing` suffix during processing)
- **Never read files still being written** (check file stability before processing)
- **Process sequentially by default** (parallelism is optional and dangerous)
- **Optional rate limiting:** `MAX_FILES_PER_POLL` limits files processed per cycle (prevents overwhelming downstream systems)

**File stability check:**

- Verify file size hasn't changed over a short interval (e.g., 1 second)
- Or require sender to create a "done" marker file (e.g., `data.csv.ready`)

**Rationale:** Reading in-flight files causes corruption, partial processing, and race conditions. Safety first.

### 7. Error Handling Philosophy

**One bad row does not poison the file.**

- **Structural CSV errors** → File fails (no header, wrong delimiter, encoding issues)
- **Row-level parse errors** → Captured, counted, included in failure report, processing continues

**For failed files, generate `.error.json` alongside archive:**

```json
{
  "errorType": "StructuralError" | "RowParseError" | "OutputError",
  "message": "Human-readable error description",
  "rowNumbers": [5, 12, 47],
  "originalFilename": "data.csv",
  "timestamp": "2026-01-20T10:15:23Z",
  "totalRows": 100,
  "failedRows": 3
}
```

**Rationale:** Future-you will ask "why did this fail?" Present-you should be kind. Detailed error reports enable debugging without digging through logs.

### 8. Logging (Brief but Sharp)

**Structured logs only.** No unstructured text logs.

**Required fields:**

```json
{
  "level": "INFO" | "WARNING" | "ERROR",
  "event": "file_detected" | "processing_complete" | "processing_failed",
  "file": "data.csv",
  "rows": 150,
  "durationMs": 234
}
```

**Rationale:** Structured logs enable querying, alerting, and dashboards. Human-readable text is for documentation, not logs.

### 9. Streaming (No Monster Files in Memory)

**JSON output must stream.** Never load entire CSV into memory.

- Parse CSV row-by-row
- Write JSON incrementally (streaming JSON array)
- Constant memory footprint regardless of file size

**Rationale:** 10 MB file? Fine. 10 GB file? Also fine. Memory usage must be independent of input size.

### 10. Configuration Philosophy (12-Factor Compliant)

**Environment variables only.** No config files, no hardcoded values.

- All configuration via environment variables
- Sensible defaults for non-critical settings
- Required variables fail fast if missing
- No .ini, .yaml, .toml, .json config files

**Rationale:** 12-factor app methodology ensures portability across environments (dev, staging, prod) without code changes.

## System Summary

This system is:

- ✅ **Deterministic**: Same input always produces same output
- ✅ **Stateless**: No memory of previous runs
- ✅ **Strict CSV parsing**: No guessing, no auto-detection
- ✅ **Explicit routing**: Clear rules for every outcome
- ✅ **Observable failures**: Detailed error reports
- ✅ **No magic**: Predictable behavior, no surprises

## Decision Outcome

**These principles are the system contract.** They define boundaries that must not be violated.

### What We Will NOT Do

❌ Add business logic or data enrichment  
❌ Auto-detect CSV formats (guessing is forbidden)  
❌ Transform schemas or mutate data structures  
❌ Support multiple simultaneous output modes  
❌ Keep files in the input folder after processing  
❌ Process files that are still being written  
❌ Copy files instead of moving them  
❌ Ignore malformed CSV headers  
❌ Support non-UTF-8 encodings without explicit configuration  
❌ Add "just one more feature" without revisiting this ADR  

### What We Will Do

✅ Strictly validate CSV format  
✅ Fail fast on malformed input  
✅ Move files atomically to archive folders  
✅ Provide clear, deterministic outcomes  
✅ Maintain file lineage in queue mode  
✅ Process files safely with stability checks  
✅ Keep the system simple, predictable, and maintainable  
✅ Generate detailed error reports for failed files  
✅ Use structured logging for observability  
✅ Stream JSON to handle files of any size  
✅ Configure via environment variables only  
✅ Continue processing on row-level errors (don't poison entire file)  

## Consequences

### Positive

- **Predictable Behavior**: Every file follows the same deterministic path
- **Easy Debugging**: Clear rules make troubleshooting straightforward
- **No Feature Creep**: Strict scope prevents complexity explosion
- **Reliable Operation**: Atomic moves and validation prevent data loss
- **Maintainability**: Simple system is easy to understand and modify
- **Clear Contracts**: Users know exactly what to expect

### Negative

- **Rigid Constraints**: Cannot handle edge cases outside these rules
- **No Auto-Detection**: Users must configure delimiter explicitly
- **Auto-Generated Column Names**: Files without headers get generic names (`col_0`, `col_1`, etc.) instead of meaningful field names
- **UTF-8 Only**: Requires encoding conversion for other formats
- **Single Output Mode**: Cannot write to file AND queue simultaneously

### Mitigation

1. **Document Constraints Clearly**: Users must understand limitations upfront
2. **Fail Fast with Clear Errors**: Don't silently ignore problems
3. **Provide Configuration Options**: Allow customization within boundaries (delimiter, quote char, filters)
4. **Offer Pre-Processing Advice**: Document how to prepare files before ingestion (add meaningful headers for better JSON keys, convert encoding)
5. **Monitor for Violations**: Log and alert when files fail validation

## Enforcement

Any proposed change that violates these principles must:

1. **Supersede this ADR** with a new ADR documenting the rationale
2. **Justify the complexity increase** with concrete business value
3. **Update documentation** to reflect the new contract
4. **Add tests** to validate the new behavior

**Default answer to "Can we add...?" is NO, unless this ADR is updated.**

## References

- [Go Implementation](../../internal/)
- [Configuration Schema](../../.env.example)
- [README - Error Handling](../../README.md#error-handling)
- [ADR-001: Why Go](./ADR-001-use-go-over-python.md)
- [ADR-002: Why RabbitMQ](./ADR-002-use-rabbitmq-for-queuing.md)

## Revision History

- **2026-01-20:** Initial definition of core system principles and behavior contract
