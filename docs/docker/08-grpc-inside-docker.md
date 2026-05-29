# gRPC Communication Inside Docker

> The single most confusing part of dockerizing a gRPC microservices stack. Read this carefully — most outages are explained on this page.

## The two-line summary

1. **A gRPC server must listen on `0.0.0.0`** (or just `:50051`) — not `127.0.0.1` — or no other container can reach it.
2. **A gRPC client must dial the callee by compose service name** — `auth-svc:50051`, not `localhost:50051`.

If your stack is broken, 80% chance one of these is wrong.

## Why `localhost` fails inside a container

Each container has its own network namespace. Inside container A:

| What you write           | What it points to                  |
|--------------------------|------------------------------------|
| `localhost`              | Container A itself                 |
| `127.0.0.1`              | Container A itself                 |
| `auth-svc`               | The `auth-svc` container (via Docker DNS) |
| `host.docker.internal`   | The host machine                   |

So `grpc.Dial("localhost:50051")` from `order-svc` tries to reach a gRPC server inside `order-svc` itself — there isn't one, hence `connection refused`.

## Where this shows up in YOUR code

In `go-grpc-order-svc/pkg/client/product_client.go` (and equivalents), the URL comes from config:

```go
cc, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
```

The `addr` is `c.ProductSvcUrl`, which is loaded from env. In Docker:

```yaml
order-svc:
  environment:
    PRODUCT_SVC_URL: "product-svc:50052"     # ← service name
```

Outside Docker (local `go run`):

```env
# pkg/config/envs/dev.env
PRODUCT_SVC_URL=localhost:50052
```

Same code, two environments — see [`07-environment-variables.md`](./07-environment-variables.md).

## Why the server must listen on `0.0.0.0`

```go
lis, err := net.Listen("tcp", c.Port)   // c.Port = ":50051"
```

Your code uses `":50051"` — the **leading colon** means "all interfaces", which is correct. If you ever change it to `"127.0.0.1:50051"`, the server only listens on its own loopback and **no other container can reach it**.

Verify inside the container:

```bash
docker compose exec auth-svc sh
# install netstat first if alpine doesn't have it: apk add net-tools
> netstat -tln | grep 50051
tcp        0      0 :::50051                :::*                    LISTEN
```

`:::50051` (or `0.0.0.0:50051`) = good. `127.0.0.1:50051` = broken.

## Diagnosing a "can't reach the service" error

Symptom: `rpc error: code = Unavailable desc = connection error: ... dial tcp ...: connect: connection refused`

Step-by-step:

```bash
# 1. Is the callee actually running?
docker compose ps                     # is auth-svc up?

# 2. Did it crash earlier?
docker compose logs --tail=200 auth-svc

# 3. Can the caller see the callee on DNS?
docker compose exec order-svc sh
> getent hosts auth-svc               # returns an IP if DNS works

# 4. Are both on the same network?
docker network inspect ecom-net | grep -E '"Name"|auth-svc|order-svc'

# 5. What address is the caller actually using?
docker compose exec order-svc env | grep SVC_URL

# 6. Is the callee listening on the right port?
docker compose exec auth-svc sh
> apk add net-tools
> netstat -tln                        # should show :::50051
```

Nine out of ten times, the issue is in step 5 (URL wrong) or the callee crashed at step 1-2.

## Health checking gRPC services

`wget --spider tcp://localhost:50051` (what our basic Dockerfile uses) just checks that something is listening on the TCP port. It doesn't verify the gRPC server is responsive.

The proper solution is **grpc_health_probe** + the standard gRPC Health Checking Protocol.

### In the Dockerfile

```dockerfile
# Install grpc_health_probe in the runtime stage
ARG GRPC_HEALTH_PROBE_VERSION=v0.4.24
RUN wget -qO/bin/grpc_health_probe \
      https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 \
  && chmod +x /bin/grpc_health_probe

HEALTHCHECK --interval=15s --timeout=3s --start-period=10s --retries=3 \
  CMD /bin/grpc_health_probe -addr=localhost:50051 || exit 1
```

### In the gRPC server (Go)

```go
import (
    "google.golang.org/grpc/health"
    healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

grpcServer := grpc.NewServer()
healthSrv := health.NewServer()
healthpb.RegisterHealthServer(grpcServer, healthSrv)
healthSrv.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
```

K8s readiness/liveness probes will use the exact same probe later — investment pays off twice.

## Long-lived connections — performance gotcha

gRPC clients in this project are created **once** at startup and reused:

```go
// pkg/client/product_client.go
cc, _ := grpc.Dial(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
productClient := pb.NewProductServiceClient(cc)
```

Don't call `grpc.Dial` per request. Each Dial opens a new TCP connection, kills HTTP/2 multiplexing, and quickly exhausts file descriptors. Reuse the client struct.

## Timeouts — set them on the client

The default has no deadline. A slow product-svc means order-svc requests pile up forever. Always wrap the call:

```go
ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
defer cancel()
resp, err := productClient.FindOne(ctx, &pb.FindOneRequest{...})
```

In Docker this is essential — restart events, slow boots, and brief network glitches happen all the time.

## TLS between services?

Inside the private compose network it's typically OK to use **insecure** (`insecure.NewCredentials()`). Don't expose internal gRPC ports to the host. When you move to K8s with a service mesh (Istio, Linkerd) you get mTLS automatically.

## Pitfalls

- Listening on `127.0.0.1` → other containers can't connect (`connection refused`).
- Hard-coded `localhost` URLs → same.
- Forgetting `WithBlock()` during testing → `Dial` returns immediately even if the server is unreachable; the failure shows up on the first RPC.
- Reusing a client across goroutines is fine; **dialing per request is not**.
- No deadline on RPCs → cascading slowdowns.

## Related
- [`04-container-networking.md`](./04-networking.md)
- [`../troubleshooting/grpc-issues.md`](../troubleshooting/grpc-issues.md)
- [`../architecture/02-service-communication.md`](../architecture/02-service-communication.md)
