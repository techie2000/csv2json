# ADR-006: Message Envelope and Provenance Metadata

**Status:** Accepted  
**Date:** 2026-01-23  
**Decision Makers:** Architecture Team  
**Context:** csv2json Multi-Ingress Routing Architecture

## Context and Problem Statement

In distributed systems, messages traveling through queues often lose critical context about their origin, ingestion
path, and intended processing contract. This leads to "context amnesia" where downstream services must infer intent
from payload structure alone—a fragile pattern prone to:

- **Ambiguous routing**: Services guess processing logic based on payload shape
- **Contract drift**: Queue names change but semantics remain unclear
- **Debugging nightmares**: No audit trail of message origin or transformation
- **Brittle integrations**: Downstream services break when queue topology changes
- **Schema confusion**: Structurally identical payloads with different meanings

**Medieval Astrology Problem**: Downstream services shouldn't divine meaning from payload shapes like astrologers
reading tea leaves. They should be explicitly told:

- Where the message came from (provenance)
- What contract it satisfies (schema/version)
- When and how it was ingested (audit trail)

## Decision Drivers

- **Provenance is not business logic**: Queue names, source files, and ingestion timestamps are metadata, not payload concerns
- **Contracts outlive infrastructure**: Queue names change, contracts shouldn't
- **Downstream autonomy**: Services should branch on declared contracts, not inferred structure
- **Audit compliance**: Full lineage tracking from source file to queue message
- **Schema evolution**: Support versioned contracts and schema registries
- **Time-travel debugging**: Reproduce issues by replaying messages with full context
- **Clean architecture**: Separate concerns (envelope vs payload)

## Options Considered

### Option 1: Minimal Context (Current)

**Structure:**

```json
{
  "route": {"name": "products", "source": "/path/to/file.csv"},
  "identifier": "file.csv",
  "data": [...]
}
```

**Pros:**

- Simple to implement
- Minimal message overhead

**Cons:**

- No queue name (loses provenance)
- No contract/schema identifier
- No service version (debugging issues)
- No timestamp (ordering/replay issues)
- Downstream services guess intent from structure

### Option 2: Enhanced Envelope (Chosen)

**Structure:**

```json
{
  "meta": {
    "ingestionContract": "customers.csv.v1",
    "source": {
      "type": "file",
      "name": "products_20260122_103045.csv",
      "path": "/data/input/products/products_20260122_103045.csv",
      "queue": "products_queue",
      "broker": "rabbitmq://localhost:5672",
      "route": "products"
    },
    "ingestion": {
      "service": "csv2json",
      "version": "0.2.0",
      "timestamp": "2026-01-22T10:30:45Z"
    }
  },
  "data": [...]
}
```

**Pros:**

- Full provenance chain (file → route → queue)
- Contract-based routing (not structure inference)
- Service version enables compatibility checks
- Timestamp for ordering/replay
- Supports schema registries
- Clean separation of concerns

**Cons:**

- Larger message size (~200 bytes overhead)
- Requires code changes in producer and consumers

### Option 3: Headers-Only Metadata

**Structure:**

- Use AMQP message headers for metadata
- Keep payload unchanged

**Pros:**

- Zero payload overhead
- Protocol-level metadata

**Cons:**

- Headers lost when messages persisted to files
- Not portable across queue systems
- Harder to debug (invisible metadata)
- Breaks OUTPUT_TYPE=both pattern

## Decision Outcome

**Chosen Option:** Enhanced Envelope (Option 2)

### Rationale

1. **Provenance is Essential**: Queue name is part of message meaning—two queues with identical payloads
   can have different semantics
2. **Contracts Enable Evolution**: `ingestionContract` allows schema versioning independent of infrastructure changes
3. **Audit Trail**: Full lineage from source file → route → queue → timestamp
4. **Debugging**: Service version + timestamp enable time-travel debugging
5. **Downstream Clarity**: Services declare "I accept customers.csv.v1" not "I accept whatever comes from that queue today"
6. **Cross-System Portability**: Works with OUTPUT_TYPE=both (files + queues), headers don't

### Trade-offs Accepted

- **Message Size**: ~200 bytes overhead per message (acceptable for audit/debug value)
- **Migration Cost**: Existing consumers must handle new envelope format (mitigated by backward compatibility period)
- **Complexity**: More fields to populate (mitigated by library/helper functions)

## Implementation Details

### Contract Identifier Format

**Recommended Pattern:** `<domain>.<entity>.<format>.<version>`

Examples:

- `customers.csv.v1` - Customer CSV ingestion contract v1
- `orders.json.v2` - Orders JSON ingestion contract v2
- `products.delimited.v1` - Products pipe-delimited contract v1

**Version Evolution:**

- **v1 → v2**: Breaking changes (new required fields, removed fields)
- **v1.1**: Minor changes (new optional fields, bug fixes)
- **v1.1.2**: Patches (no schema changes)

### routes.json Configuration

```json
{
  "routes": [
    {
      "name": "products",
      "ingestionContract": "products.csv.v1",
      "input": {
        "path": "./data/input/products"
      },
      "output": {
        "type": "queue",
        "destination": "products_queue",
        "includeEnvelope": true
      }
    }
  ]
}
```

### Message Envelope Fields

| Field | Required | Description |
| ----- | -------- | ----------- |
| `meta.ingestionContract` | ✅ | Schema/contract identifier (e.g., `products.csv.v1`) |
| `meta.source.type` | ✅ | Source type: `file`, `api`, `stream` |
| `meta.source.name` | ✅ | Original source filename |
| `meta.source.path` | ✅ | Full source file path |
| `meta.source.queue` | ✅* | Queue name (*required for queue output) |
| `meta.source.broker` | ✅* | Broker URI (*required for queue output) |
| `meta.source.route` | ✅ | Route name from configuration |
| `meta.ingestion.service` | ✅ | Service name (csv2json) |
| `meta.ingestion.version` | ✅ | Service version (semantic version) |
| `meta.ingestion.timestamp` | ✅ | ISO8601 ingestion timestamp (UTC) |

### Downstream Service Pattern

**Anti-Pattern (BAD):**

```go
// Inferring intent from structure
if hasField(msg, "sku") && hasField(msg, "price") {
    // Assume it's a product?
}
```

**Correct Pattern (GOOD):**

```go
// Explicit contract checking
switch msg.Meta.IngestionContract {
case "products.csv.v1":
    processProductsV1(msg.Data)
case "products.csv.v2":
    processProductsV2(msg.Data)
default:
    return ErrUnknownContract
}
```

## Consequences

### Positive

- ✅ **Zero ambiguity**: Downstream services know exactly what they're processing
- ✅ **Contract evolution**: Change queue topology without breaking consumers
- ✅ **Schema registries**: Enable centralized contract management
- ✅ **Audit compliance**: Full lineage tracking for regulatory requirements
- ✅ **Time-travel debugging**: Replay messages with exact context
- ✅ **Operational visibility**: Monitor by contract, not queue name
- ✅ **Safe migrations**: Transition consumers contract-by-contract
- ✅ **No more guessing**: No regex'ing filenames, no "this queue usually means X"

### Negative

- ⚠️ **Message size**: ~200 bytes overhead per message (0.2KB typical, 2% of 10KB payload)
- ⚠️ **Migration effort**: Update existing consumers to handle envelope
- ⚠️ **Configuration complexity**: Must define contracts in routes.json
- ⚠️ **Learning curve**: Teams must understand contract-based routing

### Mitigation

1. **Message Size**: Negligible for audit/debug value (0.2KB typical overhead)
2. **Migration**:
   - Phase 1: Producer emits both old and new formats (dual-write)
   - Phase 2: Consumers migrate to new format
   - Phase 3: Remove old format support
3. **Complexity**: Provide contract templates and validation tools
4. **Learning Curve**: Document patterns and provide examples

## References

- [ADR-004: Multi-Ingress Routing Architecture](ADR-004-multi-ingress-routing-architecture.md)
- [CloudEvents Specification](https://cloudevents.io/) - Industry standard for event metadata
- [Schema Registry Patterns](https://www.confluent.io/blog/schema-registry-patterns/) - Confluent best practices
- [Message Envelope Pattern](https://www.enterpriseintegrationpatterns.com/patterns/messaging/EnvelopeWrapper.html) - EIP

## Future Considerations

- **Schema Registry Integration**: Validate messages against registered schemas
- **Contract Testing**: Consumer-driven contract tests (Pact, Spring Cloud Contract)
- **Observability**: Trace messages across services using contract ID + timestamp
- **Dead Letter Queues**: Route unknown contracts to DLQ for investigation
- **Contract Deprecation**: Mark contracts deprecated with sunset date

## Revision History

- **2026-01-23**: Initial decision - Adopt message envelope with provenance metadata
