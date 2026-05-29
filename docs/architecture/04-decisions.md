# Architecture Decision Records (ADRs)

> Lightweight ADRs — one paragraph per decision. Format: **Context → Decision → Consequences**.

When you change direction on something significant, add a new ADR here. Never delete old ADRs — supersede them so history is preserved.

---

## ADR-001 — gRPC over REST for service-to-service

**Context.** Internal services need to call each other frequently. REST + JSON would be easy but verbose, slow to (de)serialize, and untyped.
**Decision.** Use gRPC (HTTP/2 + protobuf) for all internal service-to-service traffic. Keep REST only at the edge (gateway ↔ client).
**Consequences.** +Strong typing, smaller payloads, generated clients. +Streaming when we need it. −Harder to debug from `curl`; need grpcurl or BloomRPC. −Browsers can't call gRPC directly — must go through the gateway.

---

## ADR-002 — One database per service

**Context.** Microservices traditionally own their data. We could share one Postgres database with one schema per service to save resources.
**Decision.** Each service has its own DATABASE (`e_auth_svc`, `e_product_svc`, etc.). In dev we even run separate Postgres containers per service for true isolation.
**Consequences.** +Schema changes don't ripple. +Real microservice discipline — no cross-service JOINs. −More DB containers to manage. −Need gRPC calls for what would have been JOINs (e.g. order → product).

---

## ADR-003 — Viper with `AutomaticEnv` for config

**Context.** Need config that works for local `go run` and inside Docker without code changes.
**Decision.** Each service reads `pkg/config/envs/dev.env` via Viper, with `viper.AutomaticEnv()` enabled so environment variables override file values.
**Consequences.** +Same code path works locally and in containers. +Compose `environment:` block is the override source. −The env file must still exist for `ReadInConfig()` to succeed (we keep `dev.env` baked into the image; compose overrides at runtime).

---

## ADR-004 — API Gateway pattern (single edge)

**Context.** Six services means six potential URLs for the client. Auth has to happen somewhere.
**Decision.** Single Gin-based API gateway handles: HTTP↔gRPC translation, JWT validation, routing.
**Consequences.** +One entry point, one place for cross-cutting concerns. +Internal services don't need their own auth code. −Gateway can become a bottleneck or single point of failure (mitigated later with replicas + LB).

---

## ADR-005 — Compose for dev, K8s for prod (future)

**Context.** We need orchestration locally that resembles production.
**Decision.** Use Docker Compose for local development. Design Dockerfiles, env vars, healthchecks, and service naming to map cleanly onto Kubernetes later. See [`kubernetes/01-preparation.md`](../kubernetes/01-preparation.md).
**Consequences.** +Same images run in dev and prod. +Concepts (env vars, healthchecks, restart policies) transfer to K8s. −Compose ≠ K8s — features like HPA, RollingUpdate, NetworkPolicy don't exist in Compose. We'll learn them later.

---

## ADR-006 — Non-root containers

**Context.** A compromised container running as root has wider blast radius.
**Decision.** Every Dockerfile creates an `app` user and runs as that user.
**Consequences.** +Privilege reduction. −Some image operations (binding ports <1024, writing to system dirs) become impossible — which is actually a feature.

---

## Template for new ADRs

```markdown
## ADR-NNN — <short title>

**Context.** What problem are we solving? What were the constraints?
**Decision.** What did we decide?
**Consequences.** Pros, cons, and what becomes possible/impossible because of this.
```
