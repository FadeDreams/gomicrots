
### Fixed
- **event consumer (RabbitMQ):** eliminated silent message loss caused by
  auto-ack and a discarded JSON unmarshal error (`1d62bca`)
  - Switched `Listen` from auto-ack to manual ack/nack so messages are only
    removed from the queue after successful handling
  - Malformed payloads (previously silently unmarshaled into a zero-value
    struct and processed as valid) now route straight to a terminal dead
    queue instead of failing silently
  - Added a dead-letter exchange with a 30s-TTL retry queue for transient
    `handlePayload` failures, bounded to 5 attempts via the `x-death` header
  - Replaced the anonymous/exclusive queue with a durable named queue, since
    ephemeral queues dropped in-flight retries on reconnect
  - Set consumer prefetch (`Qos`) to cap in-flight unacked messages
