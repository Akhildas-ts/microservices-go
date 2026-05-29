# Port Conflicts

## `bind: address already in use`

### Symptom
```
Error response from daemon: driver failed programming external connectivity on endpoint ecom-api-gateway: Bind for 0.0.0.0:3000 failed: port is already allocated
```

### Likely cause
Something on the host is already listening on the port you're trying to publish (3000, or 5432 if you publish Postgres).

### Diagnose

```bash
# macOS / Linux — who holds the port?
lsof -i :3000
sudo lsof -i :5432

# Or with ss (Linux)
ss -tlnp | grep 3000

# Maybe it's another docker container?
docker ps --format '{{.Names}}\t{{.Ports}}' | grep 3000
```

### Fixes

Pick whichever fits the situation:

```bash
# (a) Stop the conflicting process (often a previous run of this very stack)
docker compose down
# or kill the PID lsof showed

# (b) Change the host-side port in compose
ports:
  - "3001:3000"   # host 3001 → container 3000

# (c) Remove the published port entirely if no host access needed
# (delete the ports: block — internal services don't need it anyway)
```

---

## Conflict between two compose projects

If you run two e-commerce stacks at once (one from `main`, one from a feature branch), both try to bind 3000.

### Fix
- Run them with different `-p <project-name>`:
  ```bash
  docker compose -p ecom-main up -d
  docker compose -p ecom-feat up -d
  ```
- And set unique host ports per project:
  ```yaml
  ports:
    - "${HOST_GW_PORT:-3000}:3000"
  ```
  Set `HOST_GW_PORT=3001` in the feature branch's `.env`.

---

## Postgres host port collision

### Symptom
Failed to start `postgres-auth` because port 5432 already used by a local Postgres.

### Fix
Don't publish Postgres to the host. The microservices reach it over the Docker network at `postgres-auth:5432` — there's no need for host access.

If you genuinely need host access (DB GUI), use a different host port and bind to loopback only:
```yaml
ports:
  - "127.0.0.1:5433:5432"
```

---

## Related
- [`../docker/04-container-networking.md`](../docker/04-networking.md)
