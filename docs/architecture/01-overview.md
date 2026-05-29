# System Architecture — Overview

> A Go-based e-commerce backend built as 6 microservices communicating over gRPC, fronted by a Gin HTTP API gateway.

## Why microservices for this project?

The same e-commerce backend could be a monolith and would probably ship faster. We chose microservices to learn:
- **Service boundaries** — what belongs in auth vs. admin vs. product?
- **Inter-service communication** — gRPC, retries, timeouts.
- **Independent deployability** — fix admin without redeploying cart.
- **Per-service data ownership** — every service owns its own DB.
- **Production-grade operational concerns** — containers, healthchecks, secrets, observability.

Trade-off accepted: more moving parts, network failure modes, and operational complexity.

## Components

| Service       | Tech           | Port    | Owns                                  |
|---------------|----------------|---------|---------------------------------------|
| API Gateway   | Gin (HTTP)     | `:3000` | Routing, HTTP↔gRPC translation, auth middleware |
| Auth Service  | gRPC + GORM    | `:50051`| Users, signup, login, JWT issuance     |
| Product Service | gRPC + GORM | `:50052`| Product catalog, inventory             |
| Order Service | gRPC + GORM    | `:50053`| Orders, order items                    |
| Admin Service | gRPC + GORM    | `:50054`| Admin users, admin login, admin-only ops |
| Cart Service  | gRPC + GORM    | `:50055`| Per-user carts, cart items             |

All services use:
- **Go 1.23**
- **Viper** for config (`pkg/config/envs/dev.env` + env-var overrides)
- **GORM** with the PostgreSQL driver
- **JWT** for authentication (auth + admin issue tokens; gateway validates)

## High-level diagram

```
                       ┌────────────────────────┐
        Browser/API ──►│ API Gateway   :3000    │
                       │  (Gin HTTP)            │
                       └──┬─────┬─────┬───┬───┬─┘
              ┌───────────┘     │     │   │   └──────────┐
              │ gRPC            │gRPC │   │ gRPC         │ gRPC
              ▼                 ▼     ▼   ▼              ▼
        ┌─────────┐       ┌─────────┐ ┌─────────┐  ┌──────────┐
        │  Auth   │       │ Product │ │ Admin   │  │   Cart   │
        │ :50051  │       │ :50052  │ │ :50054  │  │  :50055  │
        └────┬────┘       └────┬────┘ └────┬────┘  └────┬─────┘
             │                 │           │            │
             │       ┌─────────┘           │            │
             │       │                     │            │
             │       │      ┌──────────┐   │            │
             │       │      │  Order   │   │            │
             │       │      │  :50053  │◄──────────────┐│
             │       │      └────┬─────┘   │            │
             │       │           │         │      gRPC  │
             ▼       ▼           ▼         ▼            ▼
        ┌────────┐ ┌──────────┐ ┌────────┐ ┌────────┐ ┌────────┐
        │ PG     │ │ PG       │ │ PG     │ │ PG     │ │ PG     │
        │ auth   │ │ product  │ │ order  │ │ admin  │ │ cart   │
        └────────┘ └──────────┘ └────────┘ └────────┘ └────────┘

   Internal gRPC calls:
     order-svc  → product-svc (validate product, fetch price)
     cart-svc   → product-svc (validate product, fetch price)
```

## Folder layout in the repo

```
e-commerce-micro/
├── docker-compose.yml          # orchestrates the entire stack
├── .env / .env.example         # secrets for compose
├── Makefile                    # daily commands
├── docs/                       # ← you are here
│
├── go-grpc-api-gateway/
├── go-grpc-auth-svc/
├── go-grpc-product-svc/
├── go-grpc-order-svc/
├── go-grpc-admin-svc/
└── go-grpc-cart-svc/

Each service:
  ├── Dockerfile
  ├── .dockerignore
  ├── cmd/main.go
  ├── go.mod / go.sum
  └── pkg/
      ├── config/   (viper-based)
      ├── db/       (GORM init)
      ├── models/   (GORM models)
      ├── pb/       (generated protobuf code)
      └── services/ (gRPC service implementations)
```

## What this doc does NOT cover
- Specific request flow → see [`03-data-flow.md`](./03-data-flow.md)
- Why we picked gRPC, postgres-per-service, etc. → see [`04-decisions.md`](./04-decisions.md)
- How services find each other in Docker → see [`02-service-communication.md`](./02-service-communication.md)
