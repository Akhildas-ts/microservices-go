# Dockerfile Guide — Multi-Stage Builds for Go

> Goal: produce a tiny, secure, production-grade image for each Go service.

## Why multi-stage builds?

| Without multi-stage              | With multi-stage           |
|----------------------------------|----------------------------|
| Final image ~ 1 GB (golang base) | Final image ~ 20–40 MB     |
| Contains Go compiler + full source | Contains only the binary |
| Bigger attack surface            | Almost nothing to attack   |
| Slow image pulls in CI/prod      | Fast pulls, fast scale-up  |

Pattern:

1. **Stage 1 (builder)** — use the full Go image, compile a static binary.
2. **Stage 2 (runtime)** — start from a minimal base, `COPY --from=builder` only the binary.

## Template — gRPC service (auth / product / order / admin / cart)

Save as `go-grpc-<name>-svc/Dockerfile`. Inline comments are the lesson.

```dockerfile
# =========================================================
# Stage 1: Builder — full Go toolchain, compiles the binary.
# =========================================================
FROM golang:1.23-alpine AS builder

# git is needed for some `go mod download` operations.
RUN apk add --no-cache git

WORKDIR /src

# Copy go.mod / go.sum first so this layer is cached as long
# as dependencies don't change. Rebuilds become MUCH faster.
COPY go.mod go.sum ./
RUN go mod download

# Now copy the rest of the source.
COPY . .

# Produce a fully static binary:
#   CGO_ENABLED=0   -> no libc, runs in any minimal image
#   -trimpath       -> reproducible: strip local file paths
#   -ldflags "-s -w"-> strip debug info, shrinks binary ~30%
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath -ldflags="-s -w" \
    -o /out/app ./cmd

# =========================================================
# Stage 2: Runtime — tiny final image, only what's needed.
# =========================================================
FROM alpine:3.20

# ca-certificates -> required if the service makes outbound HTTPS calls.
# tzdata          -> avoids "unknown timezone" issues if you set TZ.
# wget            -> tiny tool used for the HEALTHCHECK below.
RUN apk add --no-cache ca-certificates tzdata wget && \
    addgroup -S app && adduser -S app -G app

WORKDIR /app

# The compiled binary from stage 1.
COPY --from=builder /out/app /app/app

# Viper expects this file at this path. Compose env vars will
# OVERRIDE whatever values are inside it (because AutomaticEnv()).
COPY --from=builder /src/pkg/config/envs/dev.env /app/pkg/config/envs/dev.env

# Run as non-root.
USER app

# CHANGE THIS PER SERVICE ↓↓↓
EXPOSE 50051

# Simple TCP healthcheck — upgrade to grpc_health_probe later.
HEALTHCHECK --interval=15s --timeout=3s --start-period=10s --retries=3 \
  CMD wget -q --spider --timeout=2 tcp://localhost:50051 || exit 1

ENTRYPOINT ["/app/app"]
```

### Per-service port table

| Service        | `EXPOSE` & healthcheck port |
|----------------|------------------------------|
| auth-svc       | `50051` |
| product-svc    | `50052` |
| order-svc      | `50053` |
| admin-svc      | `50054` |
| cart-svc       | `50055` |

## Template — API Gateway (HTTP)

The gateway is HTTP (Gin), so its healthcheck uses real HTTP.

```dockerfile
FROM golang:1.23-alpine AS builder
RUN apk add --no-cache git
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath -ldflags="-s -w" \
    -o /out/api-gateway ./cmd

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata wget && \
    addgroup -S app && adduser -S app -G app
WORKDIR /app
COPY --from=builder /out/api-gateway /app/api-gateway
COPY --from=builder /src/pkg/config/envs/dev.env /app/pkg/config/envs/dev.env
USER app
EXPOSE 3000
HEALTHCHECK --interval=15s --timeout=3s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:3000/ || exit 1
ENTRYPOINT ["/app/api-gateway"]
```

## `.dockerignore` (put one in EACH service folder)

```
# VCS, IDE, OS junk
.git
.gitignore
*.md
.DS_Store
.idea/
.vscode/

# Local-build binaries
bin/
tmp/
*.exe
*.test
*.out
.cache/

# Other env files
**/*.local.env
```

A small build context = faster `docker build` + fewer accidental file leaks.

## Layer-order rule (most-stable → least-stable)

```dockerfile
COPY go.mod go.sum ./    ← changes rarely
RUN go mod download      ← cached unless go.mod changes
COPY . .                 ← changes constantly
RUN go build             ← cached only when nothing above changed
```

If you swap these and `COPY . .` first, you re-download every module on every code change.

## What to leave OUT of the runtime image

- The Go compiler.
- The source code (`.go` files).
- `go.mod` / `go.sum` (the binary doesn't need them).
- Tests, fixtures, docs.
- Any `.git` directory.

If `docker run --rm ecom/auth-svc:dev ls -la /app` shows source files, your `.dockerignore` or `COPY` is wrong.

## Stretch goals

- Switch final stage to **distroless** (`gcr.io/distroless/static-debian12`) for ~10 MB images. Tradeoff: no shell to `exec` into.
- Install **grpc_health_probe** in the runtime stage for proper gRPC healthchecks.
- Pin the alpine version (`alpine:3.20.3`) for reproducibility.

## Related
- [`03-compose-guide.md`](./03-compose-guide.md) — how compose builds and runs these images
- [`06-debugging.md`](./06-debugging.md) — when your image is built but the container won't run
