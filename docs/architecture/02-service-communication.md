# Service Communication

> How services talk to each other in this project, and the rules around it.

## External vs internal communication

| Boundary | Protocol | Why |
|---|---|---|
| Client → Gateway | HTTP/JSON (REST) | Browsers and 3rd-party clients speak HTTP natively |
| Gateway → Services | gRPC over HTTP/2 | Binary, schema-typed, fast, native streaming |
| Service → Service | gRPC over HTTP/2 | Same reasons |
| Service → Postgres | TCP (pq protocol) | GORM handles the driver |

The gateway is the **only** component published to the host. Everything internal stays on the private Docker network.

## Who calls whom

```
api-gateway   →  auth-svc, product-svc, order-svc, admin-svc, cart-svc
order-svc     →  product-svc            (validate product, fetch price)
cart-svc      →  product-svc            (validate product, fetch price)
auth-svc      →  (nobody — leaf)
admin-svc     →  (nobody — leaf)
product-svc   →  (nobody — leaf, but called by gateway, order, cart)
```

Mental model: **only the gateway is a fan-out node**. Fan-out between internal services is intentionally limited — keeps the graph simple.

## How services address each other

Inside Docker, every service is reachable by its **compose service name**. Docker's embedded DNS does the lookup.

| What you set in config       | Resolves to                                |
|------------------------------|--------------------------------------------|
| `AUTH_SVC_URL=auth-svc:50051`   | the container running `auth-svc`        |
| `PRODUCT_SVC_URL=product-svc:50052` | the `product-svc` container          |

> ⚠️ **Never use `localhost` for inter-service URLs inside containers.** Inside container A, `localhost` is container A itself. The connection will be refused. See [`troubleshooting/network-and-dns.md`](../troubleshooting/network-and-dns.md).

Outside Docker (running services with `go run`), use `localhost:50051` etc. as before. The same env-var name covers both worlds — different values per environment.

## Connection lifecycle

gRPC clients in this project are created **once at startup** (see `pkg/client/*.go` in order and cart services) and reused for every RPC. gRPC keeps a long-lived HTTP/2 connection and multiplexes streams over it.

```go
// pkg/client/product_client.go (simplified)
cc, err := grpc.Dial(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
return pb.NewProductServiceClient(cc)   // reuse this client for the life of the process
```

> 💡 Don't `grpc.Dial` per request. Each Dial opens a TCP connection, defeats HTTP/2 multiplexing, and quickly exhausts file descriptors.

## Failure modes you must reason about

1. **Callee unreachable** (`Unavailable`) → callee down or network broken.
2. **Slow callee** → without a timeout, the caller hangs.
3. **Partial failure** → product-svc DB is up, product-svc is up, but auth-svc is down.
4. **Cascading failure** → product-svc gets slow, order-svc and cart-svc both pile up requests, eventually starve.

Mitigations to add as you mature the project (see [`microservices/02-inter-service-comm.md`](../microservices/02-inter-service-comm.md)):
- Per-RPC **deadlines** (`ctx, cancel := context.WithTimeout(ctx, 2*time.Second)`).
- **Retries** with backoff for idempotent calls.
- **Circuit breakers** (e.g. `sony/gobreaker`) to fail fast when downstream is sick.
- **Health checks** so orchestrators stop sending traffic to bad instances.

## Authentication across services
- The gateway validates JWTs at the edge.
- For now, downstream services trust the gateway (no per-RPC auth).
- Production hardening: gateway propagates the validated identity via gRPC **metadata** (`authorization: Bearer …` or a signed internal header), and each service verifies.

## Related docs
- [`grpc/01-grpc-basics.md`](../grpc/01-grpc-basics.md)
- [`microservices/02-inter-service-comm.md`](../microservices/02-inter-service-comm.md)
- [`docker/04-networking.md`](../docker/04-networking.md)
