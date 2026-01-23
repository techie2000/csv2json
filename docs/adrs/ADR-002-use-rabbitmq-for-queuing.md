# ADR 002: Use RabbitMQ for Message Queue Output

**Status:** Accepted  
**Date:** 2026-01-20  
**Decision Makers:** Development Team  
**Context:** csv2json File Polling Service

## Context and Problem Statement

The csv2json service needs to support outputting converted JSON data to a message queue for downstream processing,
in addition to file-based output. This enables asynchronous, decoupled processing and integration with event-driven
architectures. The system must support a message broker that is:

- Reliable and production-ready
- Compatible with standard protocols for portability
- Suitable for moderate message volumes (thousands per day, not millions per second)
- Deployable in both self-hosted and cloud environments
- Simple to integrate with Go applications

## Decision Drivers

- **Reliability**: Messages must be delivered reliably with acknowledgment support
- **Simplicity**: Easy to set up, configure, and operate for small-to-medium workloads
- **Portability**: Not locked into a single cloud provider
- **Protocol Support**: Standards-based protocol (AMQP) for interoperability
- **Go Integration**: Mature, well-maintained Go client library
- **Operational Overhead**: Acceptable balance between control and maintenance burden
- **Message Guarantees**: Support for durable queues and message persistence
- **Multi-cloud Strategy**: Ability to deploy across different environments

## Options Considered

### Option 1: RabbitMQ

**Pros:**

- Industry-standard AMQP protocol ensures portability
- Mature, battle-tested with extensive production usage
- Excellent Go client library (`streadway/amqp`)
- Supports durable queues, message persistence, acknowledgments
- Can run self-hosted (Docker, Kubernetes) or cloud-managed
- Lower operational complexity compared to Kafka
- Rich management UI and monitoring tools
- Suitable for moderate message volumes (10k-100k msgs/day)
- Multi-cloud portable (not vendor-locked)

**Cons:**

- Requires self-hosting or managed service setup
- Not as high-throughput as Kafka (but sufficient for use case)
- Operational overhead compared to fully-managed cloud services
- Need to manage clustering/HA for production resilience
- Monitoring and maintenance responsibility

### Option 2: Apache Kafka

**Pros:**

- Extremely high throughput (millions of messages/second)
- Built for streaming and event sourcing use cases
- Strong ordering guarantees and log-based storage
- Popular in enterprise architectures
- Mature Go clients available

**Cons:**

- **Significant operational complexity** (ZooKeeper/KRaft, partitions, replication)
- **Overkill for this use case** (thousands, not millions of messages)
- Steeper learning curve for setup and maintenance
- Heavier resource requirements (memory, disk, CPU)
- More complex integration compared to simple queue semantics
- Designed for streaming, not simple work queues

### Option 3: AWS SQS (Simple Queue Service)

**Pros:**

- Fully managed, no infrastructure to maintain
- Scales automatically with demand
- Pay-per-use pricing model
- Native AWS integration
- High availability built-in

**Cons:**

- **Vendor lock-in**: Tightly coupled to AWS ecosystem
- Multi-cloud portability compromised
- Different API semantics (not AMQP standard)
- Potential for higher long-term costs at scale
- Less control over queue configuration
- Development/testing requires AWS credentials or LocalStack

### Option 4: Azure Service Bus

**Pros:**

- Fully managed Azure service
- Enterprise-grade features (dead-letter queues, sessions, transactions)
- Automatic scaling and high availability
- Deep Azure integration

**Cons:**

- **Vendor lock-in**: Azure-specific
- No AMQP 1.0 pure compatibility (uses proprietary extensions)
- Multi-cloud portability issues
- Higher cost than self-hosted solutions
- Overkill for simple use case
- Requires Azure subscription for dev/test

### Option 5: Google Cloud Pub/Sub

**Pros:**

- Fully managed GCP service
- Global distribution and scalability
- Strong at-least-once delivery guarantees
- Automatic scaling

**Cons:**

- **Vendor lock-in**: Google Cloud specific
- Not AMQP protocol, proprietary API
- Multi-cloud portability compromised
- Higher cost for moderate workloads
- Requires GCP credentials for all environments

## Decision Outcome

**Chosen Option:** RabbitMQ

### Rationale

RabbitMQ was selected as the primary message queue implementation because it provides the **optimal balance of
simplicity, portability, and reliability** for this service:

1. **Standards-Based Protocol**: AMQP ensures we're not locked to a specific vendor or platform. The same code can
   work with RabbitMQ, Azure Service Bus (with AMQP support), or other AMQP-compliant brokers.

2. **Right-Sized Solution**: Our use case involves thousands of messages per day, not millions per second. RabbitMQ
   handles this workload effortlessly without Kafka's operational complexity.

3. **Multi-Cloud Portability**: Can be deployed as:
   - Self-hosted Docker container (development)
   - Kubernetes deployment (on-prem or any cloud)
   - Managed RabbitMQ service (CloudAMQP, AWS MQ, Azure)
   - This flexibility aligns with avoiding vendor lock-in

4. **Mature Go Integration**: The `streadway/amqp` library is stable, well-documented, and widely used in production
   Go applications.

5. **Operational Simplicity**: Compared to Kafka, RabbitMQ is significantly easier to deploy, configure, and maintain
   for moderate workloads.

6. **Future-Proof**: If workload requirements change dramatically, we can:
   - Scale RabbitMQ horizontally with clustering
   - Migrate to cloud-managed AMQP services
   - Eventually replace with Kafka if streaming use cases emerge (architecture already supports pluggable output
     handlers)

### Trade-offs Accepted

1. **Operational Responsibility**: Unlike fully-managed cloud services, we take on responsibility for deployment,
   monitoring, and maintenance of RabbitMQ (mitigated by Docker/Kubernetes and managed hosting options).

2. **Not Maximum Throughput**: Kafka would handle higher throughput, but this is unnecessary for current requirements
   (mitigated by sufficient headroom for growth).

3. **Self-Hosted Infrastructure**: Requires hosting infrastructure unless using managed RabbitMQ service (mitigated by
   containerization and cloud hosting flexibility).

## Consequences

### Positive

- **Portable Architecture**: Not locked into any single cloud provider
- **Standard Protocol**: Easy to swap RabbitMQ for another AMQP broker if needed
- **Developer Productivity**: Simple queue semantics make integration straightforward
- **Cost Control**: Self-hosted option provides predictable infrastructure costs
- **Flexibility**: Can choose deployment model based on environment (Docker, K8s, managed)
- **Reliable Delivery**: Durable queues and acknowledgments ensure message safety
- **Ecosystem**: Rich tooling, monitoring, and operational best practices available

### Negative

- **Infrastructure Management**: Need to deploy, monitor, and maintain RabbitMQ instances (unless using managed
  service)
- **High Availability**: Must configure clustering/mirroring for production resilience
- **Monitoring Required**: Need to implement health checks, metrics collection, and alerting
- **Learning Curve**: Team needs to understand RabbitMQ concepts (exchanges, queues, bindings)
- **Backup/Recovery**: Must plan for message persistence and disaster recovery

### Mitigation

1. **Use Docker Compose for Development**: Simplifies local setup and testing
2. **Consider Managed RabbitMQ**: CloudAMQP, AWS MQ, or Azure Service Bus (AMQP mode) for production to reduce
   operational burden
3. **Implement Health Checks**: Monitor queue depth, connection status, message rates
4. **Document Operations**: Create runbooks for common operational tasks
5. **Stub Cloud Queues**: Maintain stubs for AWS SQS and Azure Service Bus in code, allowing future migration if
   multi-queue support is needed
6. **Abstract Output Interface**: Keep output handler interface generic, making broker replacement straightforward

## References

- [RabbitMQ Official Documentation](https://www.rabbitmq.com/documentation.html)
- [AMQP 0-9-1 Protocol Specification](https://www.rabbitmq.com/resources/specs/amqp0-9-1.pdf)
- [streadway/amqp Go Client](https://github.com/streadway/amqp)
- [RabbitMQ vs Kafka Comparison](https://www.cloudamqp.com/blog/when-to-use-rabbitmq-or-apache-kafka.html)
- [Go Queue Handler Implementation](../../internal/output/queue_handler.go)

## Revision History

- **2026-01-20:** Initial decision to use RabbitMQ as primary queue implementation
