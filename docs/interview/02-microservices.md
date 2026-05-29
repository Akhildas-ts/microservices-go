# Microservices Interview Questions

> Use this project's architecture as the concrete example throughout.

---

### Q1. Why split this app into microservices instead of one monolith?

- **Independent deploys**: shipping a product-svc fix doesn't touch auth.
- **Failure isolation**: cart-svc crashing doesn't take down login.
- **Tech freedom**: each service could pick its own DB or language.
- **Team scaling**: clear ownership boundaries.

Trade-off: distributed-systems complexity — network calls, partial failures, eventual consistency. For a small team or a small app, a monolith is often the right call.

---

### Q2. Why does each service have its own database?

To enforce loose coupling. If two services share a DB, every schema change is a coordination event. With separate DBs, the only contract between services is the **API** (gRPC here).

Cost: no cross-service JOINs. You either:
- Call the other service over gRPC (we do this for order/cart → product).
- Denormalize / cache the data you need.
- Use events to keep local copies in sync.

---

### Q3. How do services communicate in this system?

Two protocols:
- **HTTP + JSON** between clients and the API gateway.
- **gRPC over HTTP/2** between the gateway and internal services, and between internal services.

The gateway is the single edge that translates one to the other.

---

### Q4. What if `product-svc` is down? How does `order-svc` behave?

Today: the gRPC call fails, the order request fails.

Better:
- **Timeouts** so we don't hang forever.
- **Retries with backoff** for idempotent calls.
- **Circuit breaker** to fail fast when downstream is sick.

Long-term: shift to an event-driven flow — order-svc emits "order placed", product-svc reconciles asynchronously.

---

### Q5. How do you handle distributed transactions?

Avoid them when possible. Patterns:

- **Saga** — chain of local transactions with compensating actions on failure.
- **Outbox pattern** — write your data + an event to the same DB transaction; a publisher reads the outbox table and ships events. Guarantees at-least-once delivery.
- **Two-phase commit** — works but slow and fragile; rarely used in modern systems.

---

### Q6. What's the API gateway's job?

- Single entry point for clients.
- Protocol translation (HTTP/JSON ↔ gRPC).
- Cross-cutting concerns: auth, rate limiting, CORS, logging.
- Optional: response aggregation, A/B routing.

---

### Q7. How do you trace a request across services?

Inject a **correlation ID** at the gateway (e.g. `X-Request-ID`), propagate it via gRPC **metadata**, log it in every service. The natural upgrade is **OpenTelemetry** for full distributed tracing.

---

### Q8. How do services find each other (service discovery)?

- **Docker Compose**: by service name (Docker's embedded DNS).
- **Kubernetes**: by Service DNS (`auth-svc.default.svc.cluster.local`).
- **Bare VMs**: Consul / Eureka / etcd.

Same idea: a name resolver gives the caller a reachable address.

---

### Q9. How do you version a microservice API?

For gRPC:
- Protobuf is **forward/backward compatible** if you follow rules (never change field numbers, never reuse a removed number, never change types).
- For breaking changes, version the package: `package shop.v1;` → `package shop.v2;`.

For HTTP:
- URL-versioning: `/v1/orders`, `/v2/orders`.
- Header-versioning: `Accept: application/vnd.shop.v2+json`.

---

### Q10. How do you keep services consistent if they have separate DBs?

- Eventual consistency is the default — accept it.
- Use **events** (Kafka / NATS / Postgres LISTEN/NOTIFY) to propagate state changes.
- Use the **outbox pattern** so you never have a state where DB and event log disagree.

---

### Q11. How do you scale one service without scaling the others?

In Docker Compose: run multiple replicas of just that service:
```bash
docker compose up -d --scale product-svc=3
```
Put a load balancer in front. In K8s: bump the `Deployment` replica count or use **HPA** (Horizontal Pod Autoscaler) based on CPU / custom metrics.

---

### Q12. How do you handle backward-incompatible DB migrations?

The **expand/contract** pattern:
1. **Expand**: add the new column/table — old code still works.
2. **Migrate code**: ship new code that uses both old + new.
3. **Backfill**: copy old data to new shape.
4. **Contract**: drop the old column.

Each step is independently safe. Pair with a real migration tool (not GORM AutoMigrate) like `golang-migrate` or `goose`.

---

### Q13. Synchronous vs asynchronous service communication?

- **Sync (gRPC, HTTP)** — simple, immediate response, tight coupling. We use this for "must happen now" calls.
- **Async (events / queues)** — loose coupling, resilient, eventual consistency. Use for notifications, audit, fan-out work.

Modern systems mix both.

---

### Q14. What's a "distributed monolith" and how do you avoid it?

A microservice system where every change requires deploying many services together — you get the complexity of distribution without the benefit of independence. Avoid by:
- Clear, stable APIs between services.
- No shared databases.
- Limited synchronous fan-out (1–2 hops max).
- Each service owns its data and emits events when state changes.

---

### Q15. When would you NOT use microservices?

- Small team (< 5 engineers).
- Small app / clear single domain.
- Early-stage product where boundaries aren't known yet — premature splitting locks in wrong lines.
- Tight latency budgets where network hops hurt.

**Start with a modular monolith. Split when pain demands it.**

---

## Related
- [`01-docker.md`](./01-docker.md)
- [`03-grpc.md`](./03-grpc.md)
- [`../architecture/04-decisions.md`](../architecture/04-decisions.md)
