# Docker Compose Guide

> Compose is the single declarative file that boots the entire stack: 6 services + 5 Postgres instances + 1 network.

## Why Compose

Without Compose, starting the stack means six `docker run …` invocations with the right `--network`, `--env`, `-v`, `-p` flags. With Compose, all of that lives in `docker-compose.yml`. One `docker compose up`.

## The compose mental model

```
docker-compose.yml
├── networks:   declares the private bridge network
├── volumes:    declares named volumes (DB persistence)
└── services:   one block per container we want to run
     ├── build / image
     ├── environment (overrides .env values inside the container)
     ├── depends_on (start order + healthcheck gating)
     ├── ports (only when the container should be reachable from host)
     ├── networks (which network(s) the container joins)
     └── healthcheck (how Docker decides "is this container healthy?")
```

## Full `docker-compose.yml`

Save at the **repo root**.

```yaml
# docker-compose.yml
# One file, entire stack: 6 services + 5 Postgres instances + 1 network.

networks:
  ecom-net:
    driver: bridge

volumes:
  pgdata-auth:
  pgdata-product:
  pgdata-order:
  pgdata-admin:
  pgdata-cart:

services:
  # =========================================================
  # POSTGRES — one per service (true microservice isolation)
  # =========================================================
  postgres-auth:
    image: postgres:16-alpine
    container_name: ecom-postgres-auth
    restart: unless-stopped
    environment:
      POSTGRES_USER: ${PG_USER}
      POSTGRES_PASSWORD: ${PG_PASSWORD}
      POSTGRES_DB: e_auth_svc
    volumes:
      - pgdata-auth:/var/lib/postgresql/data
    networks: [ecom-net]
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${PG_USER} -d e_auth_svc"]
      interval: 5s
      timeout: 3s
      retries: 10

  postgres-product:
    image: postgres:16-alpine
    container_name: ecom-postgres-product
    restart: unless-stopped
    environment:
      POSTGRES_USER: ${PG_USER}
      POSTGRES_PASSWORD: ${PG_PASSWORD}
      POSTGRES_DB: e_product_svc
    volumes: [pgdata-product:/var/lib/postgresql/data]
    networks: [ecom-net]
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${PG_USER} -d e_product_svc"]
      interval: 5s
      timeout: 3s
      retries: 10

  postgres-order:
    image: postgres:16-alpine
    container_name: ecom-postgres-order
    restart: unless-stopped
    environment:
      POSTGRES_USER: ${PG_USER}
      POSTGRES_PASSWORD: ${PG_PASSWORD}
      POSTGRES_DB: e_order_svc
    volumes: [pgdata-order:/var/lib/postgresql/data]
    networks: [ecom-net]
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${PG_USER} -d e_order_svc"]
      interval: 5s
      timeout: 3s
      retries: 10

  postgres-admin:
    image: postgres:16-alpine
    container_name: ecom-postgres-admin
    restart: unless-stopped
    environment:
      POSTGRES_USER: ${PG_USER}
      POSTGRES_PASSWORD: ${PG_PASSWORD}
      POSTGRES_DB: e_admin_svc
    volumes: [pgdata-admin:/var/lib/postgresql/data]
    networks: [ecom-net]
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${PG_USER} -d e_admin_svc"]
      interval: 5s
      timeout: 3s
      retries: 10

  postgres-cart:
    image: postgres:16-alpine
    container_name: ecom-postgres-cart
    restart: unless-stopped
    environment:
      POSTGRES_USER: ${PG_USER}
      POSTGRES_PASSWORD: ${PG_PASSWORD}
      POSTGRES_DB: e_cart_svc
    volumes: [pgdata-cart:/var/lib/postgresql/data]
    networks: [ecom-net]
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${PG_USER} -d e_cart_svc"]
      interval: 5s
      timeout: 3s
      retries: 10

  # =========================================================
  # MICROSERVICES
  # =========================================================
  auth-svc:
    build:
      context: ./go-grpc-auth-svc
      dockerfile: Dockerfile
    image: ecom/auth-svc:dev
    container_name: ecom-auth-svc
    restart: unless-stopped
    # These env vars OVERRIDE values inside dev.env at runtime
    # because the service uses viper.AutomaticEnv().
    environment:
      PORT: ":50051"
      DB_URL: "postgres://${PG_USER}:${PG_PASSWORD}@postgres-auth:5432/e_auth_svc?sslmode=disable"
      JWT_SECRET_KEY: ${JWT_SECRET_KEY}
    depends_on:
      postgres-auth:
        condition: service_healthy   # waits for pg_isready, not just "started"
    networks: [ecom-net]

  product-svc:
    build: { context: ./go-grpc-product-svc }
    image: ecom/product-svc:dev
    container_name: ecom-product-svc
    restart: unless-stopped
    environment:
      PORT: ":50052"
      DB_URL: "postgres://${PG_USER}:${PG_PASSWORD}@postgres-product:5432/e_product_svc?sslmode=disable"
    depends_on:
      postgres-product: { condition: service_healthy }
    networks: [ecom-net]

  order-svc:
    build: { context: ./go-grpc-order-svc }
    image: ecom/order-svc:dev
    container_name: ecom-order-svc
    restart: unless-stopped
    environment:
      PORT: ":50053"
      DB_URL: "postgres://${PG_USER}:${PG_PASSWORD}@postgres-order:5432/e_order_svc?sslmode=disable"
      PRODUCT_SVC_URL: "product-svc:50052"      # ← service name, NOT localhost
    depends_on:
      postgres-order: { condition: service_healthy }
      product-svc:    { condition: service_started }
    networks: [ecom-net]

  admin-svc:
    build: { context: ./go-grpc-admin-svc }
    image: ecom/admin-svc:dev
    container_name: ecom-admin-svc
    restart: unless-stopped
    environment:
      PORT: ":50054"
      DB_URL: "postgres://${PG_USER}:${PG_PASSWORD}@postgres-admin:5432/e_admin_svc?sslmode=disable"
      JWT_SECRET_KEY: ${JWT_SECRET_KEY}
    depends_on:
      postgres-admin: { condition: service_healthy }
    networks: [ecom-net]

  cart-svc:
    build: { context: ./go-grpc-cart-svc }
    image: ecom/cart-svc:dev
    container_name: ecom-cart-svc
    restart: unless-stopped
    environment:
      PORT: ":50055"
      DB_URL: "postgres://${PG_USER}:${PG_PASSWORD}@postgres-cart:5432/e_cart_svc?sslmode=disable"
      PRODUCT_SVC_URL: "product-svc:50052"
    depends_on:
      postgres-cart: { condition: service_healthy }
      product-svc:   { condition: service_started }
    networks: [ecom-net]

  api-gateway:
    build: { context: ./go-grpc-api-gateway }
    image: ecom/api-gateway:dev
    container_name: ecom-api-gateway
    restart: unless-stopped
    # Only the gateway publishes a port to the host — single entry point.
    ports:
      - "3000:3000"
    environment:
      PORT: ":3000"
      AUTH_SVC_URL:    "auth-svc:50051"
      PRODUCT_SVC_URL: "product-svc:50052"
      ORDER_SVC_URL:   "order-svc:50053"
      ADMIN_SVC_URL:   "admin-svc:50054"
      CART_SVC_URL:    "cart-svc:50055"
    depends_on:
      auth-svc:    { condition: service_started }
      product-svc: { condition: service_started }
      order-svc:   { condition: service_started }
      admin-svc:   { condition: service_started }
      cart-svc:    { condition: service_started }
    networks: [ecom-net]
```

## Root `.env.example`

```bash
# Copy to .env (gitignored) and fill in real values.

# Postgres credentials (used by every postgres-* container)
PG_USER=postgres
PG_PASSWORD=ChangeMe_StrongPass_123

# JWT secret shared by auth-svc and admin-svc
JWT_SECRET_KEY=replace-with-long-random-string
```

Add to `.gitignore`:
```
.env
**/*.local.env
```

## How env-var overriding works

Your `pkg/config/config.go` does:

```go
viper.AddConfigPath("./pkg/config/envs")
viper.SetConfigName("dev")
viper.SetConfigType("env")
viper.AutomaticEnv()              // ← env vars override file values
viper.ReadInConfig()
viper.Unmarshal(&config)
```

Precedence (highest first):

1. Env vars set by docker-compose `environment:` block.
2. Values in `pkg/config/envs/dev.env` (copied into image).
3. Zero values.

So your local `dev.env` files keep working for `go run`, and compose overrides win inside containers — **no code change required**.

## Container naming convention used above

- DB containers: `ecom-postgres-<service>` (e.g. `ecom-postgres-auth`)
- App containers: `ecom-<service>` (e.g. `ecom-auth-svc`)
- Images: `ecom/<service>:<tag>` (e.g. `ecom/auth-svc:dev`, later `:v1.0.0`)
- Network: `ecom-net`
- Volumes: `pgdata-<service>`

Consistent prefix → easy `docker ps | grep ecom-`.

## Compose overrides for environments

You can layer files:

```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml up
```

Common pattern:

| File | Purpose |
|---|---|
| `docker-compose.yml` | Base — works for everyone |
| `docker-compose.override.yml` | Auto-applied local-dev tweaks (live reload, exposed DB ports) |
| `docker-compose.prod.yml` | Production overrides (replicas, no source mounts, secrets via files) |

## Related
- [`07-commands-cheatsheet.md`](./07-commands-cheatsheet.md)
- [`04-networking.md`](./04-networking.md)
- [`05-volumes.md`](./05-volumes.md)
