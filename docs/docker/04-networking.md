# Docker Networking

> The one rule: **inside a container, `localhost` means that container**. Use service names for inter-container traffic.

## Network types

| Driver | What it is | When to use |
|---|---|---|
| **bridge** (default) | Private virtual network on one host. Containers get private IPs and can reach each other. | Default for compose. What we use. |
| **host** | Container shares the host's network namespace. No isolation. | Performance-critical workloads on a single host. Avoid for our use. |
| **none** | No network. | Batch jobs that don't need to talk to anyone. |
| **overlay** | Spans multiple hosts (Swarm / K8s). | Multi-node clusters. Not needed locally. |
| **macvlan** | Container gets a MAC address on the physical LAN. | Edge cases (legacy network integration). |

Our `docker-compose.yml` declares one bridge network: `ecom-net`. Every container joins it.

## DNS inside Compose

When you define a service in compose, Docker:

1. Creates a container.
2. Attaches it to the declared networks.
3. **Registers its hostname = service name** in Docker's embedded DNS.

So `auth-svc:50051` resolves automatically from any other container on `ecom-net`.

```bash
# Verify from inside the api-gateway container:
docker compose exec api-gateway sh
> getent hosts auth-svc
172.20.0.5      auth-svc
```

## Why `localhost` doesn't work

Every container has its own loopback (`127.0.0.1`) and its own network stack. When `order-svc` resolves `localhost`, it gets ITSELF — not auth-svc, not the host. Result: `connection refused`.

| What you write                 | What it points to                    |
|--------------------------------|--------------------------------------|
| `localhost:50051` from container | The container itself              |
| `127.0.0.1:50051` from container | Same                              |
| `auth-svc:50051`               | The auth-svc container on `ecom-net` |
| `host.docker.internal:5432`    | The host machine (Docker Desktop)    |

## Published ports vs internal ports

```yaml
ports:
  - "3000:3000"     # host:container — opens host port 3000 to the world
```

- `EXPOSE 3000` in the Dockerfile is **documentation only** — it doesn't publish anything.
- `ports:` in compose actually opens a host port.
- For **internal-only** services (auth, product, order, admin, cart) we deliberately don't publish ports. Only the gateway is reachable from your laptop browser.

## Security implication

If you publish `5432:5432` for postgres-auth, then **anyone on your network can hit your DB** with just a username and password. Don't publish DB ports unless you genuinely need a host-side tool (psql, GUI) to connect — and even then, prefer `127.0.0.1:5432:5432` to bind only to loopback.

## Inspecting the network

```bash
docker network ls                              # all networks
docker network inspect ecom-net                # who's on it + their IPs

# Test name resolution from inside a container
docker compose exec api-gateway sh
> getent hosts auth-svc                        # alpine-friendly DNS check
> wget -qO- http://product-svc:50052/          # HTTP probe (will 4xx for gRPC)
```

## Two-network split (advanced)

For larger systems you split traffic onto two networks:

```yaml
networks:
  edge-net:        # public-facing: gateway + load balancer
  internal-net:    # private: services + DBs

services:
  api-gateway:
    networks: [edge-net, internal-net]
  auth-svc:
    networks: [internal-net]
```

The gateway bridges the two. Internal services have no path to the outside world.

## Pitfalls

- **Forgetting `networks: [ecom-net]`** on a new service → it joins the default network and can't reach the others. Symptom: `no such host: auth-svc`.
- **Re-using the same `container_name`** in two services → second one fails to start.
- **Publishing the same host port twice** → `bind: address already in use`.

## Related
- [`troubleshooting/network-and-dns.md`](../troubleshooting/network-and-dns.md)
- [`architecture/02-service-communication.md`](../architecture/02-service-communication.md)
