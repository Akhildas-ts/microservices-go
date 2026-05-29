# Dockerization — Complete Step-by-Step Walkthrough

> The single guide that takes the project from "works on my laptop with `go run`" to "boots the full stack with `make up`".

Work through the steps in order. After each step there's a **verify** command — run it before moving on.

## Prerequisites

- Docker Desktop installed and running.
- The project compiles locally (`go build ./...` succeeds in every service).
- You've read [`01-docker-basics.md`](./01-docker-basics.md) and [`03-docker-compose-setup.md`](./03-compose-guide.md).

---

## Step 0 — Bug fix: URL-encode special chars in passwords

`product-svc`, `order-svc`, and `cart-svc` ship with `DB_URL` that contains a raw `@` in the password (`akhil@123`). Postgres URL parsing reads `@` as the user/host separator and silently breaks. Auth and admin already encode it as `%40`.

**Fix:** open each of these files and URL-encode the password (`@` → `%40`), OR change the password to alphanumeric:

```
go-grpc-product-svc/pkg/config/envs/dev.env
go-grpc-order-svc/pkg/config/envs/dev.env
go-grpc-cart-svc/pkg/config/envs/dev.env
```

Why do this now: even though compose env vars will override these in Docker, you'll be testing locally too. Leave the local broken and you'll waste hours debugging the wrong thing.

**Verify:** `grep -rn 'akhil@123' .` returns nothing.

---

## Step 1 — Create the `docs/` folder

You're reading it. ✅

---

## Step 2 — Write `.dockerignore` for every service

In **each** of the 6 service folders create `.dockerignore`:

```
.git
.gitignore
*.md
.DS_Store
.idea/
.vscode/
bin/
tmp/
*.exe
*.test
*.out
.cache/
**/*.local.env
```

**Verify:** `ls -la go-grpc-auth-svc/.dockerignore` exists.

---

## Step 3 — Write the Dockerfile for `auth-svc` (do ONE first)

Use the template from [`02-dockerfile-guide.md`](./02-dockerfile-guide.md). Save as `go-grpc-auth-svc/Dockerfile`. The key things to confirm:

- `EXPOSE 50051`
- Healthcheck targets `localhost:50051`
- The `COPY --from=builder /src/pkg/config/envs/dev.env …` line is present.

**Verify (build only — no compose yet):**

```bash
cd go-grpc-auth-svc
docker build -t test-auth .
```

You should see the multi-stage build finish in 1–3 min on first run. Final image size:

```bash
docker images test-auth
# Should be ~25-40 MB, NOT 800+ MB.
```

If it's huge, you forgot multi-stage or `.dockerignore`.

---

## Step 4 — Try running the image standalone (it will fail, that's expected)

```bash
docker run --rm test-auth
```

It will fail because there's no Postgres yet. **What you want to see:** the binary started and crashed with a Postgres dial error. **What's bad:** it crashed on "Failed at config" (means the env file wasn't copied properly).

**Verify:** the failure mentions `dial tcp` or `connect: connection refused` — not config.

---

## Step 5 — Repeat steps 3–4 for the other 4 gRPC services

Copy `go-grpc-auth-svc/Dockerfile` into each of:

- `go-grpc-product-svc/Dockerfile`     → change `EXPOSE 50052`, healthcheck port `50052`
- `go-grpc-order-svc/Dockerfile`       → `EXPOSE 50053`
- `go-grpc-admin-svc/Dockerfile`       → `EXPOSE 50054`
- `go-grpc-cart-svc/Dockerfile`        → `EXPOSE 50055`

That's literally the only delta between them.

**Verify:** all 5 build cleanly: `docker build -t test-<name> ./go-grpc-<name>-svc`.

---

## Step 6 — Write the API Gateway Dockerfile

The gateway is HTTP, not gRPC — slightly different healthcheck. See [`02-dockerfile-guide.md`](./02-dockerfile-guide.md) for the exact file.

**Verify:** `docker build -t test-gateway ./go-grpc-api-gateway` succeeds and image is small.

---

## Step 7 — Root `.env.example` and `.env`

Create `.env.example` at repo root (committed) and `.env` (gitignored):

```bash
# .env.example  → copy to .env and fill in real values
PG_USER=postgres
PG_PASSWORD=ChangeMe_StrongPass_123
JWT_SECRET_KEY=replace-with-long-random-string
```

```bash
cp .env.example .env
# edit .env with real values
echo ".env" >> .gitignore
```

**Verify:** `cat .gitignore | grep -E '^\.env$'` returns the line.

---

## Step 8 — Write a MINIMAL `docker-compose.yml` (just auth + its DB)

Start small. Don't paste the full 11-service compose file yet — verify the pattern works for one service first.

```yaml
networks:
  ecom-net:
    driver: bridge

volumes:
  pgdata-auth:

services:
  postgres-auth:
    image: postgres:16-alpine
    container_name: ecom-postgres-auth
    restart: unless-stopped
    environment:
      POSTGRES_USER: ${PG_USER}
      POSTGRES_PASSWORD: ${PG_PASSWORD}
      POSTGRES_DB: e_auth_svc
    volumes:
      - pgdata-auth:/var/lib/postgresql/data
    networks: [ecom-net]
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${PG_USER} -d e_auth_svc"]
      interval: 5s
      timeout: 3s
      retries: 10

  auth-svc:
    build: { context: ./go-grpc-auth-svc }
    image: ecom/auth-svc:dev
    container_name: ecom-auth-svc
    restart: unless-stopped
    environment:
      PORT: ":50051"
      DB_URL: "postgres://${PG_USER}:${PG_PASSWORD}@postgres-auth:5432/e_auth_svc?sslmode=disable"
      JWT_SECRET_KEY: ${JWT_SECRET_KEY}
    depends_on:
      postgres-auth: { condition: service_healthy }
    networks: [ecom-net]
```

**Verify:**

```bash
docker compose up -d
docker compose ps
# Expected:
#   ecom-postgres-auth   Up (healthy)
#   ecom-auth-svc        Up
docker compose logs auth-svc
# Expected to see: "Auth Svc on :50051"
```

If something fails, walk the [debugging ladder](./09-debugging-docker.md#the-standard-diagnostic-ladder) before moving on.

---

## Step 9 — Add the other 4 gRPC services + their DBs

Append blocks to `docker-compose.yml` one at a time:

1. `postgres-product` + `product-svc`
2. `postgres-order` + `order-svc` (also needs `PRODUCT_SVC_URL: product-svc:50052`)
3. `postgres-admin` + `admin-svc`
4. `postgres-cart` + `cart-svc` (also needs `PRODUCT_SVC_URL: product-svc:50052`)

**Verify after each:**

```bash
docker compose up -d <new-svc>
docker compose logs <new-svc>
```

---

## Step 10 — Add the API Gateway

```yaml
api-gateway:
  build: { context: ./go-grpc-api-gateway }
  image: ecom/api-gateway:dev
  container_name: ecom-api-gateway
  restart: unless-stopped
  ports:
    - "3000:3000"               # ← ONLY service published to host
  environment:
    PORT: ":3000"
    AUTH_SVC_URL:    "auth-svc:50051"
    PRODUCT_SVC_URL: "product-svc:50052"
    ORDER_SVC_URL:   "order-svc:50053"
    ADMIN_SVC_URL:   "admin-svc:50054"
    CART_SVC_URL:    "cart-svc:50055"
  depends_on:
    auth-svc:    { condition: service_started }
    product-svc: { condition: service_started }
    order-svc:   { condition: service_started }
    admin-svc:   { condition: service_started }
    cart-svc:    { condition: service_started }
  networks: [ecom-net]
```

**Verify:**

```bash
docker compose up -d
docker compose ps              # all 11 containers Up
curl -i http://localhost:3000/ # something (probably 404) — proves the gateway is reachable
```

Hit a real endpoint (signup/login) to verify the gateway→auth-svc→Postgres chain works end-to-end.

---

## Step 11 — Add the root `Makefile`

Copy the Makefile from [`11-commands-cheatsheet.md`](./11-commands-cheatsheet.md). Daily commands become:

```bash
make up
make logs
make down
make ps
```

---

## Step 12 — Document what you learned

Open `docs/learning-notes/` and write a dated note. Capture:

- Anything that surprised you.
- Any error messages you hit and how you fixed them.
- Add a one-liner to `mistakes-and-lessons.md`.

This is the most valuable step. The docs you write during this Phase are the ones you'll skim before your interview.

---

## Stretch goals (optional Phase 1 polish)

- [ ] **grpc_health_probe** in each service Dockerfile + Health gRPC service registered in code → real gRPC healthchecks.
- [ ] **Graceful shutdown** in each service's `main.go`.
- [ ] **Distroless** final image stage → ~10 MB images.
- [ ] **Resource limits** in compose (`cpus`, `memory`).
- [ ] **Structured logging** (slog) emitting JSON.

Each of these is its own learning unit and good interview material.

---

## Phase 2 preview — Kubernetes

When this is solid, Phase 2 will translate every concept here into K8s manifests:

| Compose concept    | K8s equivalent                          |
|--------------------|-----------------------------------------|
| `services:` block  | `Deployment` + `Service`                 |
| `environment:`     | `ConfigMap` + `Secret`                   |
| `volumes:`         | `PersistentVolumeClaim`                  |
| `depends_on`       | `initContainers` / readiness gating      |
| `healthcheck:`     | `livenessProbe` + `readinessProbe`       |
| Compose DNS        | K8s Service DNS                          |

That's why getting Phase 1 right matters — the muscle memory transfers.

---

## Final verify — the full smoke test

```bash
make down
make build
make up
docker compose ps                   # 11/11 Up, postgres-* show (healthy)
sleep 5
curl -i http://localhost:3000/      # gateway reachable
# Hit a real route to exercise the chain:
curl -s -X POST http://localhost:3000/auth/signup \
  -H 'Content-Type: application/json' \
  -d '{"email":"test@example.com","password":"secret123"}'
```

If that returns a token, **you're done with Phase 1**. Take a screenshot. Update your resume. Celebrate.
