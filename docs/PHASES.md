# Project Phases — Roadmap

> This documentation is intentionally split into phases. Each phase has a clear "done" definition and a body of docs that grows with it.

## Phase 1 — Docker (current) 🐳

**Goal:** containerize the entire stack end-to-end. Learn Docker deeply. Build a body of interview-ready knowledge.

**Done when:**
- Every service has a multi-stage Dockerfile + `.dockerignore`.
- Root `docker-compose.yml` boots the full stack with `make up`.
- All gRPC inter-service calls work through Docker DNS (service names).
- Postgres persists data across `docker compose down`.
- Healthchecks gate `depends_on`.
- `.env` is gitignored; secrets are not in the image.
- All troubleshooting docs are populated from real issues hit.

**Docs that grow during this phase:**
- `docker/` — all 12 files.
- `troubleshooting/` — every bug becomes an entry.
- `interview/01-docker.md`, `02-microservices.md`, `03-grpc.md`.
- `learning-notes/` — daily notes.

**Recommended order:** follow [`docker/12-step-by-step-walkthrough.md`](./docker/12-step-by-step-walkthrough.md) top to bottom.

---

## Phase 2 — Kubernetes ☸️ (next)

**Goal:** translate every Docker concept into K8s manifests. Run the same stack on a local K8s (kind / minikube / Docker Desktop K8s).

**Will add:**
- `kubernetes/01-preparation.md`
- `kubernetes/02-deployments-services.md`
- `kubernetes/03-configmaps-secrets.md`
- `kubernetes/04-persistent-storage.md`
- `kubernetes/05-probes.md`
- `kubernetes/06-ingress.md`
- `kubernetes/07-helm-vs-raw.md`
- `kubernetes/08-step-by-step-walkthrough.md`
- `interview/06-kubernetes.md`

**Done when:**
- `kubectl apply -f manifests/` boots the stack on a local cluster.
- Probes are real (livenessProbe + readinessProbe).
- Secrets are K8s `Secret` objects (or external secret store).
- One service can scale independently with `kubectl scale`.
- Rolling updates work without dropping requests.

---

## Phase 3 — CI/CD 🔁

**Goal:** every commit builds, tests, scans, and ships images automatically.

**Will add:**
- `deployment/01-environments.md`
- `deployment/02-ci-cd.md` (GitHub Actions / GitLab CI)
- `deployment/03-image-tagging.md`
- `deployment/04-rollback.md`
- `security/01-image-scanning.md` (Trivy / grype)
- `interview/07-cicd.md`

**Done when:**
- Push to `main` → tested + linted + image built + pushed to GHCR.
- Tag → automated deploy to staging.
- Manual approval → deploy to prod.
- Failed deploy auto-rolls back.

---

## Phase 4 — Observability 📊

**Goal:** see what the system is doing in production. Logs, metrics, traces.

**Will add:**
- `observability/01-structured-logging.md` (slog/zap → JSON)
- `observability/02-metrics-prometheus.md`
- `observability/03-tracing-opentelemetry.md`
- `observability/04-dashboards-grafana.md`
- `observability/05-alerting.md`
- `interview/08-observability.md`

**Done when:**
- Every log line is JSON with `service`, `trace_id`, `request_id`.
- Prometheus scrapes all services; key SLI metrics graphed in Grafana.
- A request can be traced end-to-end through gateway → service → DB.
- An alert fires on a real (synthetic) problem.

---

## Phase 5 — Security & Hardening 🔒

**Goal:** turn the stack into something you'd actually run in production.

**Will add:**
- `security/02-secrets-management.md`
- `security/03-network-policies.md`
- `security/04-rbac.md`
- `security/05-supply-chain.md` (SBOM, signed images)
- `interview/09-security.md`

---

## Phase 6 — Event-driven / async ⚡ (stretch)

**Goal:** introduce a message bus (NATS / Kafka). Migrate one synchronous flow to events.

**Will add:**
- `microservices/04-async-events.md`
- `microservices/05-outbox-pattern.md`

---

## Phase 7 — Scale & resilience 🛡️ (stretch)

**Goal:** chaos testing, circuit breakers, rate limiting, service mesh.

---

## How to use this roadmap

- **Don't skip ahead.** Phase 2 stands on Phase 1's foundations.
- **Each phase is a learning unit.** Write notes in `learning-notes/` as you go.
- **Each phase has interview value.** A complete Phase 1 = strong junior backend interview. Complete Phase 1+2+3 = strong mid-level interview.
- **Update this file as scope changes.** It's a living roadmap.
