# PostgreSQL in a Docker Container

> How to run Postgres in Docker the right way: persistent data, healthy startup, one DB per service.

## Why this matters

Postgres is the only **stateful** component in this stack. If you treat its container like the others (ephemeral, restartable, throwaway) you'll wipe your data the moment you run `docker compose down -v`. We have to be deliberate about three things: **persistence**, **initialization**, and **readiness**.

## The Postgres image — what it already does

We use `postgres:16-alpine`. On first start it:

1. Reads `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB` env vars.
2. Creates the user + the initial database.
3. Runs any `*.sql` or `*.sh` files in `/docker-entrypoint-initdb.d/` — **once, only if data dir is empty**.
4. Starts listening on port `5432` inside the container.
5. Persists data to `/var/lib/postgresql/data`.

## Minimal compose block

```yaml
postgres-auth:
  image: postgres:16-alpine
  container_name: ecom-postgres-auth
  restart: unless-stopped
  environment:
    POSTGRES_USER: ${PG_USER}
    POSTGRES_PASSWORD: ${PG_PASSWORD}
    POSTGRES_DB: e_auth_svc
  volumes:
    - pgdata-auth:/var/lib/postgresql/data       # ← persistence
  networks: [ecom-net]
  healthcheck:
    test: ["CMD-SHELL", "pg_isready -U ${PG_USER} -d e_auth_svc"]
    interval: 5s
    timeout: 3s
    retries: 10

volumes:
  pgdata-auth:                                    # declared at the bottom
```

## The three things that go wrong (and how to avoid them)

### 1. Data is lost on `down`

> ⚠️ `docker compose down -v` removes the volume → wipes your DB.
> `docker compose down` (no `-v`) is safe.

Always pair the `volumes:` block with a `pgdata-*` named volume. If you don't mount one, Postgres writes to the container's ephemeral layer and everything is gone the moment the container is removed.

### 2. The app starts before Postgres is ready

A bare `depends_on: [postgres-auth]` only waits for the **container** to start — not for Postgres to accept connections. The app then crashes with `connection refused`.

Fix: combine a `healthcheck:` on the Postgres service with `depends_on.condition: service_healthy` on the app:

```yaml
auth-svc:
  depends_on:
    postgres-auth:
      condition: service_healthy
```

`pg_isready` returns 0 only when Postgres is actually accepting connections.

### 3. Changing init scripts has no effect

`/docker-entrypoint-initdb.d/` runs **only on the very first boot** (empty data dir). If you change the script later, Postgres ignores it because the volume already has data.

Two options:
- **Wipe and re-init**: `docker compose down -v && docker compose up -d`. You lose data.
- **Apply the change manually** with `psql` against the running DB.

## One database per service

We give every microservice its **own database** (and in dev, its own Postgres container) — see [`architecture/04-decisions.md`](../architecture/04-decisions.md) for the rationale.

| Container             | DB              | Owner service |
|-----------------------|-----------------|---------------|
| `ecom-postgres-auth`    | `e_auth_svc`    | auth-svc      |
| `ecom-postgres-product` | `e_product_svc` | product-svc   |
| `ecom-postgres-order`   | `e_order_svc`   | order-svc     |
| `ecom-postgres-admin`   | `e_admin_svc`   | admin-svc     |
| `ecom-postgres-cart`    | `e_cart_svc`    | cart-svc      |

**Alternative for resource-constrained laptops:** one Postgres container hosting all 5 DBs (use an init script to `CREATE DATABASE`). Less microservice-correct, but cheap.

## Connecting from another container

Inside the service, the DB URL becomes:

```
postgres://${PG_USER}:${PG_PASSWORD}@postgres-auth:5432/e_auth_svc?sslmode=disable
```

Key points:
- Hostname is the **compose service name** (`postgres-auth`), NOT `localhost`.
- `sslmode=disable` is fine inside the private compose network; required by some configurations.
- Special characters in the password must be URL-encoded (`@` → `%40`).

## Connecting from your host machine (optional)

For a DB GUI or `psql` on your laptop, **temporarily** publish the port — only bound to loopback for safety:

```yaml
postgres-auth:
  ports:
    - "127.0.0.1:5433:5432"   # host:container — host port 5433 to avoid clashing with a local pg
```

Then on the host:

```bash
psql -h 127.0.0.1 -p 5433 -U postgres -d e_auth_svc
```

> 📌 Remove this in production compose. DB ports should never be public.

## Useful psql commands inside the container

```bash
# Open a shell into the running DB container
docker compose exec postgres-auth psql -U $PG_USER -d e_auth_svc

# Once in psql:
\dt          -- list tables
\d users     -- describe a table
\du          -- list users/roles
\l           -- list databases
\q           -- quit
```

## Backup / restore (quick reference)

```bash
# Dump
docker compose exec -T postgres-auth \
  pg_dump -U $PG_USER e_auth_svc > backup-auth-$(date +%F).sql

# Restore (DB must exist)
cat backup-auth-2026-05-29.sql | \
  docker compose exec -T postgres-auth psql -U $PG_USER -d e_auth_svc
```

## Pitfalls

- **Special chars in passwords** break DB URLs. Use `[a-zA-Z0-9_-]` only, or URL-encode (`@`→`%40`, `#`→`%23`).
- **Changing `POSTGRES_PASSWORD` after first boot** doesn't update the existing user — the password was baked in on first start. To change it: `psql` and `ALTER USER`, or wipe the volume.
- **Forgetting the `networks: [ecom-net]`** → app can't resolve `postgres-auth`.
- **Mounting your host's `/var/lib/postgresql/data`** on macOS → painful permission errors. Use a named volume.

## Related
- [`05-volumes-and-persistence.md`](./05-volumes.md)
- [`07-environment-variables.md`](./07-environment-variables.md)
- [`../troubleshooting/postgres-issues.md`](../troubleshooting/postgres-issues.md)
