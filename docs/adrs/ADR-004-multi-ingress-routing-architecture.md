# ADR 004: Multi-Ingress Routing Architecture

**Status:** Accepted  
**Date:** 2026-01-22  
**Decision Makers:** Development Team  
**Context:** csv2json File Polling Service  
**Supersedes:** Portions of ADR-003 (single input assumption)

## Context and Problem Statement

The initial design assumed **one service instance = one input folder = one output destination**. This creates
operational complexity when handling multiple data sources:

**Current Pain Points:**

- Need separate service instances for products, orders, accounts, etc.
- Each instance is near-identical (duplicate Docker containers, configs, monitoring)
- Adding new input types requires deploying new service instances
- No easy way to partition load or route by source characteristics
- Configuration sprawl across multiple deployments

**The Question:**
Should we evolve from "one service per input" to "one service handling N routes"?

## Decision Drivers

- **Operational Simplicity**: Fewer moving parts, less configuration duplication
- **Scalability**: Partition by route when needed, not before
- **Evolvability**: Add new input types via configuration, not code deployment
- **Context Preservation**: Every output message should know its source route
- **Clarity**: Service doesn't understand "products" or "orders" - it understands **routes**

## Options Considered

### Option 1: Keep Current Single-Input Design

**Architecture:**

```text
Service Instance A → products/ → products_queue
Service Instance B → orders/ → orders_queue
Service Instance C → accounts/ → accounts_queue
```

**Pros:**

- Simple per-instance configuration
- Complete isolation between inputs
- Easy to reason about individual service behavior
- No shared state or route coordination

**Cons:**

- Operational overhead: N inputs = N service instances
- Configuration duplication (same code, different env vars)
- Resource waste (each instance needs CPU/memory allocation)
- Deployment friction (new input = new deployment)
- Hard to enforce consistent behavior across instances

### Option 2: Multi-Ingress Router (Configuration-Driven)

**Architecture:**

```text
Single Service Instance
├─ Route: products
│  ├─ Input: /data/input/products/
│  ├─ Filters: *.csv, products_*
│  ├─ Destination: rabbitmq://products_queue
│  └─ Semantics: header=true, delimiter=','
├─ Route: orders
│  ├─ Input: /data/input/orders/
│  ├─ Filters: *.csv, orders_*
│  ├─ Destination: rabbitmq://orders_queue
│  └─ Semantics: header=true, delimiter=','
└─ Route: accounts
   ├─ Input: /data/input/accounts/
   ├─ Filters: *.csv
   ├─ Destination: file:///data/output/accounts/
   └─ Semantics: header=false, delimiter='|'
```

**Configuration via `routes.json`:**

```json
{
  "routes": [
    {
      "name": "products",
      "input": {
        "path": "/data/input/products",
        "filenamePattern": "products_.*\\.csv",
        "pollIntervalSeconds": 10
      },
      "parsing": {
        "hasHeader": true,
        "delimiter": ",",
        "encoding": "utf-8"
      },
      "output": {
        "type": "queue",
        "destination": "rabbitmq://products_queue",
        "includeRouteContext": true
      },
      "archive": {
        "processedPath": "/data/archive/products/processed",
        "failedPath": "/data/archive/products/failed",
        "ignoredPath": "/data/archive/products/ignored"
      }
    },
    {
      "name": "orders",
      "input": {
        "path": "/data/input/orders",
        "filenamePattern": "orders_.*\\.csv",
        "pollIntervalSeconds": 5
      },
      "parsing": {
        "hasHeader": true,
        "delimiter": ","
      },
      "output": {
        "type": "queue",
        "destination": "rabbitmq://orders_queue"
      },
      "archive": {
        "processedPath": "/data/archive/orders/processed",
        "failedPath": "/data/archive/orders/failed"
      }
    }
  ]
}
```

**Pros:**

- **One binary, many behaviors**: Single codebase handles all routes
- **Config-only additions**: New input = edit routes.json, restart service
- **Load partitioning when needed**: Run multiple instances with route subsets
- **Consistent behavior**: Same code guarantees same semantics
- **Route context in output**: Every message knows its source
- **Resource efficiency**: Single process handles multiple inputs
- **Operational sanity**: One thing to monitor, one thing to debug

**Cons:**

- **Shared failure domain**: Bug affects all routes (mitigated by restart/rollback)
- **Config restart required**: Adding routes needs service restart (acceptable trade-off)
- **More complex configuration**: routes.json vs simple env vars
- **Route isolation**: Need careful error handling so one route doesn't poison others

### Option 3: Hybrid (Service per Cluster)

Run multiple instances, each handling a subset of routes (e.g., Instance A handles products + orders, Instance B
handles accounts + payments).

**Pros:**

- Some consolidation without full shared failure domain
- Can partition by criticality or load characteristics

**Cons:**

- Still operational complexity (which routes on which instance?)
- Configuration coordination challenges
- Worst of both worlds: complexity without full benefit

## Decision Outcome

### Chosen Option: Option 2 - Multi-Ingress Router

This service is **configuration-driven plumbing**. It routes files from source folders to destination queues/files
based on declared routes. It does not understand "products" or "orders" - it understands **route contracts**.

### Rationale

1. **Operational Simplicity Wins**: Managing one service instance is vastly simpler than managing N near-identical
   instances
2. **Evolvability**: Business needs change faster than deployment pipelines. Config-driven routing decouples business
   logic from infrastructure.
3. **Scalability Path**: Start with one instance. When load demands, partition routes across instances (e.g., critical
   routes on dedicated instances).
4. **Context Preservation**: Route metadata in output messages enables downstream systems to make routing decisions
   without hardcoding.

### Core Principle: The Service Knows Routes, Not Domains

The service does NOT know what "products" or "orders" mean. It knows:

- **Route Name**: Arbitrary identifier (could be "route_a", "ingestion_17")
- **Input Path**: Where to poll
- **Filters**: What to process
- **Parsing Rules**: How to interpret CSV
- **Destination**: Where to send output
- **Archive Paths**: Where to move processed files

Domain semantics (what products ARE) belong in downstream consumers, not the ingestion service.

## Implementation Plan

### Phase 1: Configuration Schema

**Environment Variable:**

```bash
ROUTES_CONFIG=/etc/csv2json/routes.json
```

**routes.json Schema:**

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "required": ["routes"],
  "properties": {
    "routes": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["name", "input", "output"],
        "properties": {
          "name": {
            "type": "string",
            "description": "Unique route identifier"
          },
          "input": {
            "type": "object",
            "required": ["path"],
            "properties": {
              "path": { "type": "string" },
              "filenamePattern": { "type": "string" },
              "suffixFilter": { "type": "string" },
              "pollIntervalSeconds": { "type": "integer", "default": 10 }
            }
          },
          "parsing": {
            "type": "object",
            "properties": {
              "hasHeader": { "type": "boolean", "default": true },
              "delimiter": { "type": "string", "default": "," },
              "encoding": { "type": "string", "default": "utf-8" }
            }
          },
          "output": {
            "type": "object",
            "required": ["type", "destination"],
            "properties": {
              "type": { "enum": ["file", "queue"] },
              "destination": { "type": "string" },
              "includeRouteContext": { "type": "boolean", "default": true }
            }
          },
          "archive": {
            "type": "object",
            "required": ["processedPath", "failedPath"],
            "properties": {
              "processedPath": { "type": "string" },
              "failedPath": { "type": "string" },
              "ignoredPath": { "type": "string" }
            }
          }
        }
      }
    }
  }
}
```

### Phase 2: Output Envelope Specification

**Queue Output with Route Context:**

```json
{
  "route": {
    "name": "products",
    "source": "/data/input/products/products_20260122_103045.csv"
  },
  "sourceFile": "products_20260122_103045.csv",
  "rowNumber": 1,
  "timestamp": "2026-01-22T10:31:45Z",
  "payload": {
    "sku": "ABC123",
    "name": "Widget",
    "price": "29.99"
  }
}
```

**Key Fields:**

- **route.name**: Which route processed this file (enables downstream routing)
- **route.source**: Full source file path for audit trail
- **sourceFile**: Filename only (for compatibility with ADR-003)
- **rowNumber, timestamp, payload**: Per ADR-003 contract

**File Output (unchanged):**
File output remains as pure JSON array per ADR-003. Route context is NOT embedded in file output (files are for
data, queues are for events).

### Phase 3: Service Behavior

**Startup:**

1. Load `routes.json` from `ROUTES_CONFIG` path
2. Validate schema and required paths exist
3. Initialize one poller per route (independent goroutines)
4. Each poller runs on its configured `pollIntervalSeconds`

**Per-Route Processing:**
Each route operates independently:

- Polls its input folder
- Applies its filters
- Parses with its semantics
- Outputs to its destination
- Archives to its paths

**Error Isolation:**

- One route's failure does NOT stop other routes
- Failed routes log errors but keep retrying
- Structured logs include route name for filtering

**Configuration Reload:**

- ❌ Hot reload NOT supported in v1 (adds complexity)
- ✅ Requires service restart to pick up route changes
- Future: Consider SIGHUP handler for zero-downtime reload

### Phase 4: Operational Considerations

**Single Instance (Default):**

```yaml
# docker-compose.yml
services:
  csv2json:
    image: csv2json:latest
    environment:
      - ROUTES_CONFIG=/config/routes.json
    volumes:
      - ./routes.json:/config/routes.json:ro
      - ./data:/data
```

**Partitioned Instances (High Load):**

```yaml
# docker-compose.yml
services:
  csv2json-critical:
    image: csv2json:latest
    environment:
      - ROUTES_CONFIG=/config/routes-critical.json  # products, orders only
    volumes:
      - ./routes-critical.json:/config/routes-critical.json:ro

  csv2json-bulk:
    image: csv2json:latest
    environment:
      - ROUTES_CONFIG=/config/routes-bulk.json  # accounts, reports, etc.
    volumes:
      - ./routes-bulk.json:/config/routes-bulk.json:ro
```

**Monitoring:**
Structured logs include route name:

```json
{
  "level": "INFO",
  "event": "file_processed",
  "route": "products",
  "file": "products_20260122_103045.csv",
  "rows": 150,
  "durationMs": 234
}
```

Query by route: `SELECT * FROM logs WHERE route = 'products' AND event = 'processing_failed'`

## Trade-offs Accepted

### Positive

- ✅ **Operational simplicity**: One service to deploy, monitor, debug
- ✅ **Config-driven evolution**: New inputs via config, not code
- ✅ **Resource efficiency**: Single process handles multiple routes
- ✅ **Scalability path**: Partition routes when needed
- ✅ **Consistent behavior**: Same code guarantees same semantics
- ✅ **Context preservation**: Route metadata enables intelligent downstream routing
- ✅ **Deployment friction reduced**: No redeploy for new input types

### Negative

- ❌ **Shared failure domain**: Bug affects all routes (until restart/rollback)
- ❌ **Restart required**: Config changes need service restart
- ❌ **Configuration complexity**: routes.json vs simple env vars
- ❌ **Route isolation**: Must ensure one route's failure doesn't poison others

### Mitigation

1. **Graceful Error Handling**: Wrap each route's processing in error recovery
2. **Route-Level Circuit Breakers**: Disable failing routes temporarily
3. **Comprehensive Testing**: Test route isolation and error propagation
4. **Monitoring Per Route**: Alert on per-route failure rates
5. **Fast Restart**: Optimize startup time for quick config reloads
6. **Config Validation**: Fail fast on invalid routes.json at startup

## Canonical Route + Envelope Specification

This design establishes a **pattern** for all future ingestion services:

### Universal Principles

1. **Service knows routes, not domains**: Generic plumbing, not business logic
2. **Config-driven routing**: Behavior defined by configuration, not code
3. **Route context in output**: Every message carries its source route
4. **Error isolation**: One route's failure doesn't stop others
5. **Structured logs with route tags**: Queryable, filterable logging

### Envelope Standard (All Ingestion Services)

```json
{
  "route": {
    "name": "string",
    "source": "string"
  },
  "sourceFile": "string",
  "rowNumber": "integer",
  "timestamp": "ISO8601",
  "payload": {}
}
```

**Contract:**

- **route.name**: Always present, identifies which route processed this
- **route.source**: Full source file path for audit trail
- **timestamp**: Always ISO8601 UTC
- **payload**: Route-specific data (CSV row, JSON object, XML element, etc.)

Future ingestion services (JSON ingestion, XML ingestion, Parquet ingestion) will follow this envelope standard
for consistency.

## Consequences

### What This Enables

✅ Add new input types via configuration  
✅ Single service handles multiple data sources  
✅ Partition load by route when needed  
✅ Route context enables downstream routing decisions  
✅ Consistent behavior across all routes  
✅ Operational simplicity (one thing to monitor)  
✅ Future ingestion services follow same pattern  

### What This Constrains

❌ Config changes require restart (no hot reload in v1)  
❌ All routes share same service lifecycle  
❌ Configuration format is more complex than simple env vars  
❌ Must ensure route isolation in error handling  

### Migration Path from ADR-003

ADR-003 principles remain valid **per-route**:

- Each route still follows three-step pipeline (poll, convert, archive)
- Each route still has one outcome (processed, ignored, failed)
- Each route still enforces CSV validation per its parsing config
- Each route still streams JSON (no memory explosion)

**ADR-003 is now per-route, not per-service.**

## References

- [ADR-003: Core System Principles](./ADR-003-core-system-principles.md) - Per-route behavior
- [ADR-001: Why Go](./ADR-001-use-go-over-python.md) - Concurrency model supports multi-route
- [ADR-002: Why RabbitMQ](./ADR-002-use-rabbitmq-for-queuing.md) - Queue routing per destination

## Revision History

- **2026-01-22:** Initial proposal for multi-ingress routing architecture
