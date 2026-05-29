# Docker Interview Questions

> Beginner-friendly answers, grounded in this project. Read them, then practice saying them out loud — the muscle memory matters more than the perfect script.

---

### Q1. What's the difference between an image and a container?

An **image** is a read-only template — like a class. A **container** is a running instance of that image — like an object. You can have one image and 100 containers from it.

> Example in this project: `ecom/auth-svc:dev` is the image; `ecom-auth-svc` is the container.

---

### Q2. Why use multi-stage Dockerfiles?

To separate **build-time tools** from **runtime needs**.

- Stage 1 has the Go compiler (~800 MB).
- Stage 2 starts from `alpine:3.20` (~7 MB) and only copies the compiled binary.

Final image: ~25 MB vs ~1 GB. Smaller = faster pulls, less attack surface, faster scale-up.

---

### Q3. Why can't a container reach another by `localhost`?

Each container has its own network namespace. `localhost` always means "this container itself". To reach another container on the same Docker network, use its **service name** as the hostname (`auth-svc:50051`). Docker's embedded DNS resolves it.

---

### Q4. What's the difference between `ENTRYPOINT` and `CMD`?

- `ENTRYPOINT` is the command that always runs.
- `CMD` provides default arguments to that command, which the user can override at `docker run`.

Common pattern:
```dockerfile
ENTRYPOINT ["/app/app"]
CMD ["--config", "/etc/conf.yaml"]
```

---

### Q5. What happens to data when a container is removed?

The container's writable layer is destroyed. Anything NOT on a **volume** is gone forever.

> That's why our Postgres data lives on a named volume (`pgdata-auth:/var/lib/postgresql/data`) — the volume outlives the container.

---

### Q6. Why use `.dockerignore`?

Everything in the build context is sent to the Docker daemon at build time. A `.git/` folder or `node_modules/` slows the build and can leak secrets into the image. `.dockerignore` keeps the context tiny.

---

### Q7. What does `EXPOSE` actually do?

It's **documentation + a hint** for other tools. It does NOT publish the port to the host. Only `-p` / `ports:` does that. Two containers on the same Docker network can talk to each other on any port regardless of `EXPOSE`.

---

### Q8. Bridge vs Host vs Overlay networks?

- **Bridge** (default) — private network on one host. What we use.
- **Host** — container shares the host's network namespace. Fast, no isolation, port conflicts likely.
- **Overlay** — spans multiple hosts (Swarm / K8s). Needed for multi-node clusters.

---

### Q9. What does `docker compose down -v` do that `down` doesn't?

The `-v` flag **also removes named volumes**. So any data on those volumes (your Postgres data!) is gone. Plain `down` removes containers but keeps volumes.

> Rule of thumb: name your destructive Makefile target `nuke`, not `down`.

---

### Q10. Why run as a non-root user inside containers?

Defense in depth. If the container is compromised, the attacker isn't root inside it — they can't bind low ports, write to system directories, or install packages. Our Dockerfile creates an `app` user and uses `USER app`.

---

### Q11. What is a Docker layer? Why does layer order matter?

Each `RUN`, `COPY`, `ADD` creates a layer. Layers are cached. If a layer's inputs haven't changed, Docker reuses it.

We copy `go.mod` + `go.sum` **before** the source so `go mod download` is cached as long as dependencies don't change. Otherwise every code edit triggers a re-download.

---

### Q12. How would you reduce image size further?

- Switch the final stage to **distroless** (`gcr.io/distroless/static-debian12`) or even `scratch`. Combined with `CGO_ENABLED=0`, our Go binary needs nothing else.
- Strip the binary: `-ldflags="-s -w"`.
- Use a stricter `.dockerignore`.

Tradeoff with distroless: no shell — debugging is harder.

---

### Q13. How does `depends_on` work? Does it wait for the dependency to be ready?

By default, `depends_on` only waits for the dependency **container** to START, not to be ready. To wait for readiness:

```yaml
depends_on:
  postgres-auth:
    condition: service_healthy
```

This requires a `healthcheck:` on the dependency.

---

### Q14. What's a healthcheck and why does it matter?

A command Docker runs periodically inside the container. If it succeeds → healthy. If it fails enough times → unhealthy. Healthy/unhealthy gates `depends_on: condition: service_healthy` and informs orchestrators (K8s, load balancers) whether to send traffic.

---

### Q15. How do you debug a crashing container?

1. `docker compose ps` — confirm it's looping.
2. `docker inspect <name> --format='{{.State.ExitCode}}'` — get the exit code.
3. `docker compose logs --tail=200 <svc>` — read the last error.
4. `docker compose run --rm <svc> sh` then `./app` — run interactively to see the real failure.

See [`../docker/09-debugging-docker.md`](../docker/09-debugging-docker.md).

---

### Q16. Difference between bind mount and named volume?

- **Bind mount** — a host folder mapped into the container. Lives on the host. Good for dev (live source edits).
- **Named volume** — managed by Docker, stored under `/var/lib/docker/volumes/…`. Good for production data (e.g. Postgres).

---

### Q17. What's the default restart policy and which one should I usually pick?

Default is `no` (never restart). For long-running services, `unless-stopped` is the right choice — it restarts on crash but respects manual `docker compose stop`.

---

### Q18. Can you describe the lifecycle of a single `docker compose up`?

1. Compose parses `docker-compose.yml` + `.env`.
2. Builds any images marked `build:` (skips if unchanged).
3. Creates the network(s) if missing.
4. Creates volume(s) if missing.
5. Starts services in dependency order. Waits for `depends_on` conditions.
6. Each container's `ENTRYPOINT` runs.
7. Healthchecks start probing once `start_period` elapses.

---

### Q19. What's `docker-compose.override.yml`?

A second file that compose automatically merges on top of `docker-compose.yml` for local development. Lets you keep dev-only tweaks (bind mounts, exposed DB ports, debug env vars) out of the main file.

---

### Q20. How do you ship secrets to a container safely?

- Never bake them into the image (`ENV PASSWORD=…`).
- Local: `.env` (gitignored).
- CI: encrypted secrets injected as env at deploy time.
- Prod: secret manager (Vault / AWS SM / GCP SM) or compose `secrets:` block (writes to `/run/secrets/<name>`).

---

## Related
- [`02-microservices.md`](./02-microservices.md)
- [`03-grpc.md`](./03-grpc.md)
- [`04-scenario-based.md`](./04-scenario-based.md)
