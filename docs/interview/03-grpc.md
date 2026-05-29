# gRPC Interview Questions

---

### Q1. Why gRPC instead of REST for service-to-service?

- **HTTP/2** — multiplexed streams over one TCP connection, lower latency.
- **Protobuf** — smaller binary payloads, schema-first, generated clients in any language.
- **Streaming** — server, client, and bidirectional streams native.
- **Performance** — typically 5–10x faster than REST+JSON for high-throughput RPC.

REST is still better at the **edge** — browsers, third-party clients, ad-hoc tools all speak HTTP/JSON.

---

### Q2. What is a `.proto` file?

The schema/contract between client and server. You define messages and service methods once; `protoc` generates strongly-typed code in Go, Java, Python, etc.

> In this project the generated code lives in each service's `pkg/pb/*.pb.go`.

---

### Q3. What are the four RPC styles in gRPC?

1. **Unary** — request/response (like REST).
2. **Server streaming** — one request, many responses.
3. **Client streaming** — many requests, one response.
4. **Bidirectional streaming** — both sides stream independently.

This project uses unary for everything currently.

---

### Q4. How does authentication work in gRPC?

- **Per-RPC**: send a JWT in metadata:
  ```go
  ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
  ```
  Server validates inside an **interceptor**.
- **Transport-level**: mTLS — both sides present TLS certificates.

In this project, the gateway validates JWTs at the edge and trusts the internal services.

---

### Q5. What are interceptors?

gRPC middleware. They wrap every RPC for logging, auth, metrics, retries.

- Server side: `grpc.UnaryInterceptor(...)`
- Client side: `grpc.WithUnaryInterceptor(...)`

Chain them with `grpc.ChainUnaryInterceptor(a, b, c)`.

---

### Q6. How do you version a gRPC API?

Protobuf is backward-compatible if you follow rules:
- Never change a field's number.
- Never reuse a removed field number.
- Never change a field's type.

For breaking changes, **version the package**: `package shop.v1;` → `shop.v2;`. Run both during migration.

---

### Q7. Common gRPC error codes — what do they mean?

| Code | Meaning |
|---|---|
| `OK` | Success |
| `UNAVAILABLE` | Couldn't reach the server (DNS, network, server down) |
| `DEADLINE_EXCEEDED` | Request took longer than client's context deadline |
| `UNAUTHENTICATED` | No / invalid auth |
| `PERMISSION_DENIED` | Authenticated but not allowed |
| `INVALID_ARGUMENT` | Request didn't match schema |
| `NOT_FOUND` | Resource doesn't exist |
| `INTERNAL` | Server-side bug |

---

### Q8. How do you set timeouts and propagate cancellation?

Use `context.WithTimeout` on every RPC:
```go
ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
defer cancel()
resp, err := productClient.FindOne(ctx, req)
```

The context is sent over the wire as a gRPC deadline. If the client cancels, the server can detect via `ctx.Done()` and stop early.

---

### Q9. Why is `grpc.Dial` per request bad?

It opens a fresh TCP+TLS+HTTP/2 connection every time:
- Kills HTTP/2 multiplexing.
- Exhausts file descriptors.
- Adds 3–10× latency.

**Create the client once, reuse it for the life of the process.** This project does this in `pkg/client/*.go`.

---

### Q10. How do you load-balance gRPC?

gRPC keeps long-lived HTTP/2 connections, so a normal L4 load balancer sends all RPCs from one client to one server. Solutions:

- **Client-side LB** — gRPC resolver returns multiple targets; client distributes RPCs.
- **L7 proxy** that understands gRPC (Envoy, Linkerd, nginx with `grpc_pass`).
- **K8s**: headless `Service` + client-side LB, OR a **service mesh** (Istio / Linkerd) doing L7 LB transparently.

---

### Q11. What's the gRPC Health Checking Protocol?

A standard service definition (`grpc.health.v1.Health`) every server can implement. Tools like `grpc_health_probe`, K8s, and load balancers can probe it. Far better than "TCP port is open" for liveness/readiness checks.

---

### Q12. gRPC vs REST in one sentence?

> gRPC is faster, strongly typed, and great between services you control; REST is universal and easier to debug at the edge.

---

### Q13. How do you debug a gRPC call without code?

- **grpcurl** — like `curl` for gRPC: `grpcurl -plaintext localhost:50051 list`.
- Enable **gRPC reflection** on the server to make grpcurl ergonomic:
  ```go
  reflection.Register(grpcServer)
  ```
- **BloomRPC / Kreya / Postman** — GUI clients that import `.proto` files.

---

### Q14. What happens to in-flight RPCs on `GracefulStop()`?

`GracefulStop()` stops accepting new RPCs and waits for in-flight ones to finish. Pair with SIGTERM handling so deploys don't drop requests. (See [`../docker/10-production-best-practices.md`](../docker/10-production-best-practices.md) §5.)

---

### Q15. What is the difference between `Stop()` and `GracefulStop()`?

- `Stop()` — closes connections immediately. Any in-flight RPC errors out.
- `GracefulStop()` — refuses new RPCs but waits for in-flight ones.

Always prefer `GracefulStop()` for production.

---

## Related
- [`../docker/08-grpc-inside-docker.md`](../docker/08-grpc-inside-docker.md)
- [`../architecture/02-service-communication.md`](../architecture/02-service-communication.md)
