# Docker Build / Cache Issues

## "I changed the code but the container still runs the old version"

### Symptom
You edit `main.go`, run `docker compose up -d`, but logs show the previous behavior.

### Diagnose

```bash
# 1. Did compose rebuild?
docker compose up -d --build       # the --build is the key

# 2. Is the binary in the image actually new?
docker compose exec auth-svc md5sum /app/app     # compare before/after

# 3. Is your change actually inside the build context?
# (i.e. is it in a folder that .dockerignore is excluding?)
cat go-grpc-auth-svc/.dockerignore
```

### Fixes (in order)

```bash
# (a) Force a rebuild
docker compose up -d --build

# (b) Bypass the layer cache
docker compose build --no-cache auth-svc
docker compose up -d auth-svc

# (c) Nuclear — wipe the image and rebuild
docker compose down
docker image rm ecom/auth-svc:dev
docker compose build --no-cache auth-svc
docker compose up -d
```

---

## Dependency download re-runs every build (slow)

### Likely cause
`go.mod` / `go.sum` are being copied **after** the source, so any source change busts the dependency-download layer.

### Fix
In the Dockerfile, copy module files first:

```dockerfile
COPY go.mod go.sum ./    # ← these first
RUN go mod download
COPY . .                 # ← source last
```

That way `go mod download` is cached as long as `go.mod` doesn't change.

---

## `COPY failed: file not found in build context`

### Likely cause
You're trying to copy a file that's excluded by `.dockerignore`, or the path is wrong relative to the build context.

### Diagnose
```bash
# What IS in the build context that Docker actually sees?
docker build --no-cache --progress=plain -t debug -f - go-grpc-auth-svc <<'EOF'
FROM alpine
COPY . /ctx
RUN ls -la /ctx
EOF
```

### Fix
- Adjust `.dockerignore` (don't be too aggressive with `*`).
- Use relative paths from the build-context root, not absolute.

---

## Build context is huge / build takes forever

### Diagnose
The first line of `docker build` output:
```
=> [internal] load build context           5.0s
=> => transferring context: 850MB          ← huge!
```

### Fix
- Add a stricter `.dockerignore`.
- Especially exclude `.git`, `bin/`, `tmp/`, IDE folders, large media files.
- Confirm with: `du -sh go-grpc-auth-svc/.* go-grpc-auth-svc/*` to find the big folders.

---

## "It builds on my machine but not in CI"

### Likely cause
You're relying on a layer your local Docker has cached. CI starts fresh.

### Fix
Test locally with `docker compose build --no-cache` periodically. Pin versions (`golang:1.23.4-alpine3.20`). Lock `go.sum` (`go mod tidy`).

---

## Disk full from old images / volumes

### Diagnose
```bash
docker system df
```

### Fix
```bash
docker image prune -f                  # dangling images
docker container prune -f              # stopped containers
docker builder prune -f                # build cache
docker volume prune                    # unused volumes (asks first)
# Or the big hammer:
docker system prune -a --volumes
```

> ⚠️ `--volumes` wipes ALL unused volumes — including `pgdata-*` if no container is currently using them.

---

## Related
- [`../docker/02-dockerfile-guide.md`](../docker/02-dockerfile-guide.md)
- [`../docker/09-debugging-docker.md`](../docker/09-debugging-docker.md)
