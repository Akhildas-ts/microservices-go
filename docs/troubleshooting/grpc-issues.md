# gRPC Issues

## `rpc error: code = Unavailable … connection refused`

### Likely cause
Callee container isn't listening yet, crashed, OR the URL points to `localhost` inside the container.

### Diagnose
```bash
docker compose ps                                 # is the callee Up?
docker compose logs --tail=100 auth-svc           # did it crash?
docker compose exec order-svc env | grep SVC_URL  # is URL = service name?
docker compose exec order-svc sh -c 'getent hosts auth-svc'   # DNS resolves?
```

### Fix
- If callee not Up → see [`container-crashes.md`](./container-crashes.md).
- If URL is `localhost:50051` → change to `auth-svc:50051` in compose `environment:`.
- If DNS fails → both containers must declare `networks: [ecom-net]`.

---

## `lookup auth-svc: no such host`

### Likely cause
The two containers are on different Docker networks (or the callee isn't running).

### Diagnose
```bash
docker network inspect ecom-net | grep -E 'Name|auth-svc|order-svc'
```

### Fix
Add `networks: [ecom-net]` to the misbehaving service block in `docker-compose.yml`, then `docker compose up -d` to recreate.

---

## `transport: Error while dialing dial tcp: i/o timeout`

### Likely cause
Server started but is listening on `127.0.0.1`, not all interfaces.

### Diagnose
```bash
docker compose exec auth-svc sh
> apk add net-tools
> netstat -tln | grep 50051
```
- Good:  `:::50051` or `0.0.0.0:50051`
- Bad:   `127.0.0.1:50051`

### Fix
In `cmd/main.go`, `net.Listen("tcp", c.Port)` with `c.Port = ":50051"` (leading colon). The leading colon means "all interfaces". Don't change to `"127.0.0.1:50051"`.

---

## `rpc error: code = DeadlineExceeded`

### Likely cause
The callee is slow OR not responding. Client's `context.WithTimeout` fired.

### Diagnose
```bash
docker compose logs --since=5m product-svc        # any errors / slow queries?
docker stats                                      # CPU/MEM pressure?
docker compose exec postgres-product psql -U postgres -c 'select * from pg_stat_activity;'
```

### Fix
- If product-svc is overloaded → add resource limits, scale, optimize the slow query.
- If the deadline is too aggressive → bump it (`context.WithTimeout(ctx, 5*time.Second)`).
- Add retries with backoff for idempotent calls.

---

## `rpc error: code = Unauthenticated / PermissionDenied`

### Likely cause
JWT validation failed at the gateway, OR the token wasn't propagated downstream.

### Diagnose
```bash
docker compose exec api-gateway env | grep JWT_SECRET   # gateway uses one secret
docker compose exec auth-svc    env | grep JWT_SECRET   # auth uses the SAME secret?
```

### Fix
Both services must use the same `JWT_SECRET_KEY`. Set once in `.env`, reference from compose.

---

## Connection works at startup, fails later

### Likely cause
Stale client connection after server restart. gRPC reconnects but the first call may surface the old failure.

### Fix
- Configure keepalive on the client:
  ```go
  grpc.Dial(addr, grpc.WithKeepaliveParams(keepalive.ClientParameters{
      Time: 30 * time.Second, Timeout: 10 * time.Second,
  }))
  ```
- Add a simple retry interceptor for idempotent calls.

---

## Related
- [`../docker/08-grpc-inside-docker.md`](../docker/08-grpc-inside-docker.md)
- [`network-and-dns.md`](./network-and-dns.md)
