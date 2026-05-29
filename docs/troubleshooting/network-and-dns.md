# Network & DNS Issues

## Container A can't see container B at all

### Symptom
```
dial tcp: lookup auth-svc: no such host
```

### Diagnose
```bash
docker network inspect ecom-net | grep -E '"Name"|auth-svc|order-svc'
```
If only one of them shows up → they're not on the same network.

### Fix
Both services must declare:
```yaml
networks: [ecom-net]
```
Then `docker compose up -d` to recreate.

---

## Wrong IP returned for a service name

### Symptom
After recreating containers, DNS still returns an old, stale IP. Connection refused.

### Diagnose
```bash
docker compose exec api-gateway sh -c 'getent hosts auth-svc'
docker network inspect ecom-net --format='{{json .Containers}}' | jq
```

### Fix
This usually means the client has a cached connection. In Go, gRPC reconnects automatically — give it a few seconds. If the caller process pinned the IP at startup (some older HTTP libraries), restart the caller:
```bash
docker compose restart api-gateway
```

---

## Two containers, same `container_name`

### Symptom
```
ERROR: for ecom-auth-svc  Cannot create container … name is already in use
```

### Diagnose
```bash
docker ps -a | grep ecom-auth-svc
```

### Fix
- Remove the stale one: `docker rm -f ecom-auth-svc`.
- Or remove the duplicate `container_name` from compose.

---

## A new service joined the default network instead of `ecom-net`

### Symptom
The service starts but can't see anyone. `docker network inspect ecom-net` doesn't list it.

### Fix
Make sure the new service block has `networks: [ecom-net]`. Without it, Compose attaches it to the auto-created default network — which is isolated from `ecom-net`.

---

## DNS works but TCP connect fails

### Symptom
`getent hosts auth-svc` returns an IP, but `nc -zv auth-svc 50051` is refused.

### Diagnose
```bash
# Is the callee listening on all interfaces?
docker compose exec auth-svc sh
> apk add net-tools
> netstat -tln | grep 50051
```

### Fix
The callee must listen on `:50051` (all interfaces), not `127.0.0.1:50051`. See [`grpc-issues.md`](./grpc-issues.md) and [`../docker/08-grpc-inside-docker.md`](../docker/08-grpc-inside-docker.md).

---

## Compose creates a new network each time and breaks DNS

### Symptom
After `docker compose down && up`, services briefly resolve to old IPs.

### Fix
Usually harmless — the network is recreated atomically. If you see persistent issues, prune the old network:
```bash
docker network ls
docker network rm <stale-net-id>
```

---

## Related
- [`../docker/04-container-networking.md`](../docker/04-networking.md)
- [`grpc-issues.md`](./grpc-issues.md)
