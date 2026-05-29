# Scenario-Based Questions

> These are the "tell me about a time…" / "what would you do if…" questions. Use the diagnostic ladder from [`../docker/09-debugging-docker.md`](../docker/09-debugging-docker.md) as your structure.

---

### Scenario 1 — "After `docker compose up`, the gateway returns 502 for /login."

Walk through it like this:

1. `docker compose ps` — is `auth-svc` running and healthy?
2. `docker compose logs auth-svc` — did it crash at startup? Likely culprit: DB not ready, or env vars wrong.
3. `docker compose exec api-gateway env | grep AUTH` — is `AUTH_SVC_URL` set to `auth-svc:50051`, not `localhost`?
4. `docker compose exec api-gateway sh -c 'getent hosts auth-svc'` — does DNS resolve?
5. `docker compose exec postgres-auth pg_isready` — is the DB accepting connections?

Likely fixes:
- Postgres still booting → add healthcheck + `condition: service_healthy`.
- Wrong service-name URL → fix compose env block.
- URL-encoded password issue.

---

### Scenario 2 — "Postgres data disappears every morning."

- Someone is running `docker compose down -v` instead of `down`. The `-v` removes volumes.
- Or the volume isn't actually mounted — verify with `docker inspect ecom-postgres-auth | grep -A5 Mounts`.

Fix: rename the destructive Makefile target (e.g. `nuke`) and educate the team.

---

### Scenario 3 — "We need to upgrade `product-svc` with zero downtime."

In Compose this is hard — Compose is single-host, single-replica per service.

Approach:
1. Build a new image with a new tag (`v1.1.0`).
2. Run BOTH versions behind a load balancer.
3. Shift traffic gradually (blue/green or canary).

This is exactly why we move to **Kubernetes** in Phase 2 — `Deployment` with `RollingUpdate` strategy gives this out of the box.

---

### Scenario 4 — "One service is hogging 100% CPU and slowing the others."

1. `docker stats` to confirm which service.
2. `docker compose logs <svc>` — anything weird?
3. Set resource limits to contain the blast radius:
   ```yaml
   deploy:
     resources:
       limits: { cpus: "0.5", memory: 256M }
   ```
4. Profile the service (`pprof`) to find the hot path.
5. Long-term: fix the leak / inefficiency.

---

### Scenario 5 — "How do you add a new microservice (e.g. payment-svc)?"

1. Scaffold `go-grpc-payment-svc/` mirroring an existing service.
2. Add `Dockerfile` + `.dockerignore`.
3. Append `payment-svc` + `postgres-payment` blocks to `docker-compose.yml`.
4. Add `PAYMENT_SVC_URL: payment-svc:50056` to the gateway's env block.
5. Define `payment.proto`, regenerate stubs.
6. Wire the gRPC client in the gateway and any other caller.
7. Update [`../architecture/01-overview.md`](../architecture/01-overview.md) and add an ADR.

---

### Scenario 6 — "Tests pass locally but fail in CI."

Common causes:
- Local Docker has cached layers CI doesn't.
- Local has a stale `.env` with values CI doesn't have.
- Different host architecture (M1 vs amd64).

Fixes:
- Periodically `docker compose build --no-cache` locally.
- Pin base images: `golang:1.23.4-alpine3.20`.
- In CI, run `docker buildx build --platform=linux/amd64` explicitly.

---

### Scenario 7 — "All services are running but `order-svc` can't talk to `product-svc`."

Walk:
1. `docker network inspect ecom-net | grep -E 'order-svc|product-svc'` — both on the network?
2. `docker compose exec order-svc sh -c 'getent hosts product-svc'` — DNS resolves?
3. `docker compose exec order-svc env | grep PRODUCT_SVC_URL` — set correctly?
4. `docker compose exec product-svc netstat -tln | grep 50052` — listening on all interfaces?

Likely root cause: missing `networks: [ecom-net]` on one of them, OR product-svc listening on `127.0.0.1` instead of `:50052`.

---

### Scenario 8 — "How do you handle a database password rotation?"

1. Update the secret store (or `.env` locally).
2. `docker compose exec postgres-auth psql -U postgres -c "ALTER USER postgres WITH PASSWORD 'NewPass';"`.
3. Recreate the dependent containers so they pick up the new env: `docker compose up -d --force-recreate auth-svc`.
4. **Note**: changing `POSTGRES_PASSWORD` env var alone does NOT update the existing user — that variable is consumed only on first boot.

---

### Scenario 9 — "How would you migrate from Docker Compose to Kubernetes?"

| Compose         | K8s |
|-----------------|-----|
| `services:`     | one `Deployment` + `Service` per app |
| `environment:`  | `ConfigMap` (non-secret) + `Secret` |
| `volumes:`      | `PersistentVolumeClaim` |
| `depends_on`    | initContainers / readiness gating |
| `healthcheck:`  | `livenessProbe` + `readinessProbe` |
| Compose DNS     | K8s Service DNS |
| `docker compose up` | `kubectl apply -f` (or Helm) |

The Dockerfiles don't change. The compose file becomes a set of manifests. Concepts transfer.

---

### Scenario 10 — "A container is healthy but the app is broken."

Healthcheck only verifies what you tell it to verify. If your healthcheck is "is the TCP port open?" the app can be totally broken (DB lost, downstream unavailable) and still pass.

Fix: make the healthcheck **meaningful**:
- gRPC: register the Health service, use `grpc_health_probe`.
- HTTP: implement `/readyz` that actually checks DB ping + downstream pings.

---

## Related
- [`05-debugging-stories.md`](./05-debugging-stories.md)
- [`../docker/09-debugging-docker.md`](../docker/09-debugging-docker.md)
- [`../troubleshooting/README.md`](../troubleshooting/README.md)
