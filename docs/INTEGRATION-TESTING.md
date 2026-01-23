# Integration Testing Guide

This document describes how to run integration tests with actual RabbitMQ to validate the ADR-006 message envelope implementation.

## Prerequisites

- Docker and Docker Compose installed
- RabbitMQ service running (`docker compose up -d`)

## Running Integration Tests

### 1. Start RabbitMQ

```bash
docker compose up -d
```

### 2. Run Integration Tests

```bash
# Run all integration tests
go test ./internal/integration -v

# Run specific test
go test ./internal/integration -v -run TestRabbitMQIntegration_EnvelopeStructure

# Skip integration tests (for CI environments)
SKIP_INTEGRATION=true go test ./internal/integration -v
```

### 3. Test Coverage

The integration tests validate:

#### Envelope Structure (`TestRabbitMQIntegration_EnvelopeStructure`)
- **IngestionContract**: Verifies contract ID is correctly set
- **Source Type**: Confirms source.type = "file"
- **Source Name**: Validates filename extraction
- **Source Path**: Checks full file path is captured
- **Source Queue**: Verifies queue name is included
- **Source Broker**: Validates broker URI format (amqp://)
- **Source Route**: Confirms route name from configuration
- **Ingestion Service**: Checks service name = "csv2json"
- **Ingestion Version**: Verifies version is populated (from VERSION file)
- **Ingestion Timestamp**: Validates RFC3339 timestamp format and recency
- **Data Payload**: Confirms data array is correctly embedded

#### Multiple Messages (`TestRabbitMQIntegration_MultipleMessages`)
- Sends 3 messages with different contracts
- Verifies each message has unique contract ID
- Validates all messages received correctly

## Troubleshooting

### Authentication Error

If you see:
```
Exception (403) Reason: "username or password not allowed"
```

**Solution**: The tests assume RabbitMQ default credentials (guest/guest). Ensure docker-compose.yml doesn't override these.

### Connection Timeout

If tests skip with "Cannot connect to RabbitMQ":

1. Verify RabbitMQ is running:
   ```bash
   docker ps | grep rabbitmq
   ```

2. Check RabbitMQ logs:
   ```bash
   docker logs csv2json-rabbitmq-1
   ```

3. Ensure port 5672 is accessible:
   ```bash
   # Windows
   Test-NetConnection localhost -Port 5672
   
   # Linux/Mac
   nc -zv localhost 5672
   ```

### Custom RabbitMQ Host

If RabbitMQ is on a different host:

```bash
QUEUE_HOST=rabbitmq.example.com go test ./internal/integration -v
```

## Benchmark Results

Envelope marshaling performance (from `go test -bench=BenchmarkBuildMessageEnvelope`):

- **Small Payload** (3 records): ~3.4 µs per operation, 1.6 KB allocated
- **Large Payload** (100 records): ~71.6 µs per operation, 36 KB allocated

Overhead is minimal: envelope metadata adds ~30 allocations and ~1.5 KB for typical messages.

## CI/CD Integration

For continuous integration pipelines:

```yaml
# GitHub Actions example
- name: Run Unit Tests
  run: go test ./... -v -short

- name: Start RabbitMQ
  run: docker compose up -d rabbitmq

- name: Run Integration Tests
  run: go test ./internal/integration -v
  
- name: Stop RabbitMQ
  run: docker compose down
```

## Test Data

Integration tests use ephemeral test queues:
- `integration-test-queue` - Envelope structure validation
- `multi-message-test-queue` - Multiple message scenarios

These queues are created automatically by RabbitMQ and can be cleaned up via:

```bash
docker compose down -v  # Remove volumes
```

## Manual Verification

You can manually inspect queue messages using RabbitMQ Management UI:

1. Access http://localhost:15672
2. Login with guest/guest
3. Navigate to Queues tab
4. Click on test queue
5. Use "Get messages" to inspect envelope structure

Example message you should see:

```json
{
  "meta": {
    "ingestionContract": "integration.csv.v1",
    "source": {
      "type": "file",
      "name": "integration-test.csv",
      "path": "/data/input/integration-test.csv",
      "queue": "integration-test-queue",
      "broker": "amqp://localhost:5672/",
      "route": "integration-test-route"
    },
    "ingestion": {
      "service": "csv2json",
      "version": "0.2.0",
      "timestamp": "2026-01-23T14:48:32Z"
    }
  },
  "data": [
    {"id": "1", "name": "Integration Test", "status": "active"},
    {"id": "2", "name": "Envelope Test", "status": "complete"}
  ]
}
```

## References

- [ADR-006: Message Envelope and Provenance Metadata](../../docs/adrs/ADR-006-message-envelope-and-provenance-metadata.md)
- [RabbitMQ Docker Documentation](https://hub.docker.com/_/rabbitmq)
- [Go Testing Package](https://pkg.go.dev/testing)
