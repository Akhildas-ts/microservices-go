# Container Crashes

## Container is in a restart loop

### Symptom
`docker compose ps` shows `Restarting (1) X seconds ago`.

### Diagnose
```bash
# 1. Exit code & error
docker inspect ecom-auth-svc --format='{{.State.Status}} exit={{.State.ExitCode}} err={{.State.Error}}'

# 2. Last logs before the crash
docker compose logs --tail=200 auth-svc

# 3. If logs are empty, run interactively to see the real error
docker compose run --rm auth-svc sh
> ./app
```

### Common causes

| Log message                                | Cause                                 | Fix |
|--------------------------------------------|---------------------------------------|-----|
| `Failed at config`                         | env file missing in image             | Check `COPY --from=builder /src/pkg/config/envs/dev.env …` in Dockerfile |
| `dial tcp: lookup postgres-auth: no such host` | DNS / network                     | See [`network-and-dns.md`](./network-and-dns.md) |
| `dial tcp: connection refused` (to DB)     | DB not ready                          | Add healthcheck + `condition: service_healthy` |
| `pq: invalid URL`                          | password URL-encoding                 | See [`postgres-issues.md`](./postgres-issues.md) |
| `panic: …`                                 | actual code bug                       | Read the stack trace, fix the line |

---

## Container exits with code 0 immediately

### Likely cause
The container's main process finished. For long-running services this means `main()` returned — usually because `grpcServer.Serve(lis)` returned an error and you didn't `log.Fatalln`.

### Fix
Wrap the serve call so any error is fatal:
```go
if err := grpcServer.Serve(lis); err != nil {
    log.Fatalln("Failed to serve:", err)
}
```
(Your services already do this — confirm.)

---

## Container exits with code 137

### Likely cause
SIGKILL — typically OOM (out of memory) or `docker kill`.

### Diagnose
```bash
docker inspect ecom-auth-svc --format='{{.State.OOMKilled}}'   # true if killed for OOM
docker stats                                                    # observe memory live
dmesg | grep -i oom                                             # Linux only
```

### Fix
- Raise the limit if you set one too low.
- Find the leak: `pprof`, log allocations, simplify.

---

## Container exits with code 139

### Likely cause
SIGSEGV — segfault. Usually CGO or a corrupt binary.

### Fix
Confirm `CGO_ENABLED=0` in the build step (it is, in our template). Rebuild without cache.

---

## Container starts but does nothing visible

### Likely cause
App is writing logs to a file inside the container instead of stdout.

### Fix
Log to stdout/stderr. That's the convention for containerized apps; any other behavior is invisible to `docker logs`.

---

## "Healthcheck never goes healthy"

### Diagnose
```bash
docker inspect ecom-auth-svc --format='{{json .State.Health}}' | jq
```

Shows the most recent healthcheck output and exit code.

### Common reasons
- Healthcheck command doesn't exist in the image (`wget` missing → install in Dockerfile).
- Healthcheck targets a port the server isn't listening on (`127.0.0.1` bind issue).
- `start_period` too short — service still booting on every probe.

---

## Related
- [`../docker/09-debugging-docker.md`](../docker/09-debugging-docker.md)
