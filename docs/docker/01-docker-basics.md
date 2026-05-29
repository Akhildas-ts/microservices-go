# Docker Basics

> Read this once before touching any Dockerfile or compose file.

## What Docker actually is

Docker packages an application together with **everything it needs to run** (OS libraries, runtime, binaries, config) into a single artifact called an **image**. A running instance of that image is a **container**.

Containers are isolated processes on the host — **NOT VMs**. They share the host kernel, which is why they start in milliseconds and use tiny amounts of RAM.

## Why Docker matters for microservices

A microservice system runs many small programs. Without Docker:

- "Works on my machine" — each dev has different Go versions, Postgres versions, OS quirks.
- Onboarding takes a day: install Go, install Postgres, configure 6 env files, start 6 terminals.
- Production deploys are fragile because prod drifts from dev.

With Docker:

- `docker compose up` brings the entire stack online identically on any OS.
- The image you test locally is byte-for-byte what runs in production.
- Each service can use a different language/runtime — they only talk over the network.

## Image vs container — the most common confusion

| Concept   | Analogy           | Mutable? | Storage              |
|-----------|-------------------|----------|----------------------|
| Image     | Class / blueprint | No       | Layers on disk       |
| Container | Object / instance | Yes      | Writable top layer   |

You **build** an image once. You **run** many containers from it.

```bash
docker build -t ecom/auth-svc:dev .       # produces an image
docker run --rm ecom/auth-svc:dev         # creates+runs a container
docker run --rm ecom/auth-svc:dev         # creates+runs ANOTHER container
```

## Image layers

Each `RUN`, `COPY`, `ADD` in a Dockerfile creates a **layer**. Layers are cached. If `go.mod` doesn't change, the `go mod download` layer is reused — that's why we copy `go.mod` **before** the source. See [`02-dockerfile-guide.md`](./02-dockerfile-guide.md).

## Docker networking basics

- Every container gets its own network namespace.
- Docker creates a **bridge network** so containers can talk to each other.
- Inside a Compose network, you reach another service by its **service name** as if it were a hostname:
  - ✅ `auth-svc:50051`
  - ❌ `localhost:50051` — `localhost` inside container = container itself.
- **Published ports** (`-p 3000:3000`) expose a container to the **host**. Two containers on the same Docker network do NOT need published ports to talk to each other.

More in [`04-networking.md`](./04-networking.md).

## Docker volumes

A container's filesystem is **ephemeral** — destroyed when the container is removed. Anything you need to keep must live on a volume.

Three kinds:

- **Named volume** — managed by Docker, best for DB data: `pgdata-auth:/var/lib/postgresql/data`
- **Bind mount** — a host folder mapped into the container, great for live code reload in dev: `./go-grpc-auth-svc:/app`
- **tmpfs** — in memory, gone on stop, for secrets you don't want on disk.

More in [`05-volumes.md`](./05-volumes.md).

## Docker daemon, client, registry

- **Docker daemon (`dockerd`)** — the long-running process that does the work.
- **Docker client (`docker` CLI)** — what you type. Talks to the daemon over a socket.
- **Registry** — image storage (Docker Hub, GHCR, ECR). `docker pull` / `docker push` move images between local and registry.

## Why Docker Compose?

`docker run …` is fine for one container. For six services + multiple databases + a network + volumes + dependency ordering, you need a declarative spec. That's `docker-compose.yml` — one file that captures the entire stack and lets you bring it up/down with one command.

Compose is also the conceptual stepping stone to Kubernetes — same vocabulary (service, network, volume, env, healthcheck).

## Mental model checklist

You're ready to write Dockerfiles when you can answer:

- [ ] What's the difference between an image and a container?
- [ ] Why is layer order important in a Dockerfile?
- [ ] Why can't a container reach another by `localhost`?
- [ ] What does `EXPOSE` actually do? (Hint: not much by itself.)
- [ ] What happens to data when a container is removed?
- [ ] What's the difference between `ENTRYPOINT` and `CMD`?

Answers live in [`interview/01-docker.md`](../interview/01-docker.md).
