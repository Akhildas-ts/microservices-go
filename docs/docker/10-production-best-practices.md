# Production Best Practices

> Things you should layer on AFTER basic compose works. Each item below also makes you sound senior in interviews.

## 1. Multi-stage builds (already done)

Final images are 20–40 MB instead of 1 GB. See [`02-dockerfile-guide.md`](./02-dockerfile-guide.md).

## 2. Run as non-root user

Already in our template:

```dockerfile
RUN addgroup -S app && adduser -S app -G app
USER app
```

**Why:** if the container is compromised, the attacker isn't root inside it — they can't bind low ports, write to system paths, or install packages.

## 3. Pin image versions

```dockerfile
FROM golang:1.23-alpine          # ✅ good (minor version pinned)
FROM golang:1.23.4-alpine3.20    # ✅✅ better (fully pinned, reproducible)
FROM golang:latest               # ❌ bad — silently moves under you
```

Same for `postgres:16-alpine` vs `postgres:latest`. **`latest` belongs in tutorials, never in production.**

## 4. Healthchecks for every service

```yaml
healthcheck:
  test: ["CMD", "/bin/grpc_health_probe", "-addr=localhost:50051"]
  interval: 15s
  timeout: 3s
  start_period: 10s
  retries: 3
```

Three flags that matter:
- `interval` — how often to probe.
- `start_period` — grace window during which failures don't count (booting up).
- `retries` — consecutive failures before marking unhealthy.

Pair with `depends_on.condition: service_healthy` so dependents wait for **ready**, not just **started**.

## 5. Graceful shutdown

Docker sends `SIGTERM`, waits 10s, then `SIGKILL`. Use those 10 seconds:

```go
sig := make(chan os.Signal, 1)
signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
go func() {
    <-sig
    log.Println("shutting down…")
    grpcServer.GracefulStop()              // for gRPC
    // srv.Shutdown(ctx)                   // for Gin
}()
```

Without this, in-flight requests get dropped on every deploy.

## 6. Structured logging

Switch `log.Fatalln` / `fmt.Println` to `slog` (Go 1.21+) or `zap`. Output JSON. Include:

```json
{"time":"…","level":"INFO","service":"auth-svc","trace_id":"abc","msg":"login ok","user":"u_42"}
```

Future-you and your log aggregator (Datadog/CloudWatch/Loki) will thank you. Plain text logs are unparseable at scale.

## 7. Resource limits

```yaml
deploy:
  resources:
    limits:
      cpus: "0.5"
      memory: 256M
    reservations:
      memory: 64M
```

Prevents one runaway service from starving the rest. Catches memory leaks early.

## 8. Restart policies

| Policy             | Meaning                                          |
|--------------------|--------------------------------------------------|
| `no`               | Never restart                                    |
| `on-failure`       | Restart only on non-zero exit                    |
| `always`           | Restart even if you stop it manually             |
| `unless-stopped`   | Restart unless YOU stopped it — ✅ **recommended** |

`unless-stopped` is the right default for our services.

## 9. Image tagging strategy

| Tag         | When to use                                  |
|-------------|----------------------------------------------|
| `:dev`      | Local development                            |
| `:sha-<git-short-sha>` | Every CI build, fully unique         |
| `:v1.2.3`   | Releases — semver, never overwritten         |
| `:latest`   | Avoid in production — it's ambiguous         |

Combine: push `ecom/auth-svc:sha-abc1234` + `ecom/auth-svc:v1.2.3` from the same build.

## 10. Smaller images (distroless)

Switch the final stage to **distroless**:

```dockerfile
FROM gcr.io/distroless/static-debian12
COPY --from=builder /out/app /app/app
USER nonroot:nonroot
ENTRYPOINT ["/app/app"]
```

Result: ~10 MB images. Tradeoff: no shell, harder to debug. Worth it once you have grpc_health_probe and good logs.

## 11. Secrets — never in image, never in git

| Where | Pattern |
|---|---|
| Local | `.env` (gitignored) |
| CI | GitHub Actions encrypted secrets |
| Prod | AWS Secrets Manager / GCP Secret Manager / Vault, fetched on startup |
| Docker-only prod | Compose `secrets:` block writes to `/run/secrets/<name>` |

Never:
- `ENV PASSWORD=…` in Dockerfile.
- `RUN echo "PASSWORD=…" >> /etc/env` in Dockerfile.
- Logging the full config struct.

## 12. Environment separation via compose files

```bash
# Dev (auto applies override file)
docker compose up

# Prod
docker compose -f docker-compose.yml -f docker-compose.prod.yml up
```

Override files only contain the differences (replicas, limits, no source mounts, secrets-as-files).

## 13. Build context hygiene

Every service has a `.dockerignore` (see [`02-dockerfile-guide.md`](./02-dockerfile-guide.md)) so `.git`, IDE config, and local binaries never enter the build context.

## 14. Reproducible builds

In Go: `-trimpath`, `-ldflags="-s -w"`. In Docker: pinned base images. Lock your `go.sum`. With these, the same source → byte-identical image, every time.

## 15. CI/CD outline (Phase 3 preview)

When you reach the CI/CD phase, each service will follow:

```
1. lint (golangci-lint)
2. test (go test ./...)
3. build image with tag :sha-<git>
4. push to registry (GHCR / ECR)
5. deploy: kubectl set image / docker compose pull && up -d
6. smoke test (probe /healthz, run a representative gRPC call)
7. rollback on failure
```

## Production-readiness checklist for THIS project

Tick these off as you go. Each one is interview-worthy.

- [ ] Multi-stage Dockerfile for every service
- [ ] `.dockerignore` in every service folder
- [ ] Non-root `app` user in every image
- [ ] Pinned base image versions
- [ ] Healthcheck on every container
- [ ] `depends_on.condition: service_healthy` where it matters
- [ ] `restart: unless-stopped` on every service
- [ ] No secrets in git, no secrets in images
- [ ] Resource limits on every service
- [ ] Graceful shutdown wired up in every Go service
- [ ] Structured (JSON) logging
- [ ] grpc_health_probe + Health gRPC service registered
- [ ] CI builds + pushes tagged images on every commit

## Related
- [`02-dockerfile-guide.md`](./02-dockerfile-guide.md)
- [`08-grpc-inside-docker.md`](./08-grpc-inside-docker.md)
- [`../architecture/04-decisions.md`](../architecture/04-decisions.md)
