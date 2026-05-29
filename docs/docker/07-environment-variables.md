# Environment Variables & Configuration

> One config story that works for `go run` AND `docker compose up`, with secrets kept out of git.

## The problem

Each service currently loads config from `pkg/config/envs/dev.env` via Viper. That works for local `go run` — but inside Docker we need:

- Different hostnames (`postgres-auth` instead of `localhost`).
- Different secrets per environment (dev / staging / prod).
- Real secrets never committed to git.

We solve this **without changing application code**, because the existing config layer already calls `viper.AutomaticEnv()` → environment variables override file values.

## The three layers of config (precedence)

```
┌──────────────────────────────────────────────┐
│ 1. Process env vars (set by docker-compose)  │  ← highest
├──────────────────────────────────────────────┤
│ 2. pkg/config/envs/dev.env (file in image)   │
├──────────────────────────────────────────────┤
│ 3. Go zero values                            │  ← lowest
└──────────────────────────────────────────────┘
```

This means:
- Outside Docker → `dev.env` wins.
- Inside Docker → compose `environment:` block wins.
- **No code change to switch** between environments.

## The two-file pattern at the repo root

| File | Committed? | Purpose |
|---|---|---|
| `.env.example` | ✅ yes | Template — variable names, no real values |
| `.env`         | ❌ no (gitignore!) | Real values for your local stack |

Compose automatically loads `.env` from the directory you run it in, and substitutes `${VAR}` placeholders in `docker-compose.yml`.

### `.env.example`

```bash
# Postgres credentials (used by every postgres-* container)
PG_USER=postgres
PG_PASSWORD=ChangeMe_StrongPass_123

# JWT secret shared by auth-svc and admin-svc
JWT_SECRET_KEY=replace-with-long-random-string
```

### `.gitignore` additions

```
.env
**/*.local.env
```

## How compose passes env vars to a service

```yaml
auth-svc:
  environment:
    PORT: ":50051"
    DB_URL: "postgres://${PG_USER}:${PG_PASSWORD}@postgres-auth:5432/e_auth_svc?sslmode=disable"
    JWT_SECRET_KEY: ${JWT_SECRET_KEY}
```

What happens:

1. Compose substitutes `${PG_USER}`, `${PG_PASSWORD}`, `${JWT_SECRET_KEY}` from `.env`.
2. The resulting key-value pairs are injected into the container as process env vars.
3. Inside the container, the Go app's `os.Getenv("PORT")` returns `":50051"`.
4. Viper sees this and uses it instead of whatever's in `dev.env`.

## Why `localhost` doesn't work inside containers

```
DB_URL=postgres://...@localhost:5432/...   ❌  inside container, localhost = the container itself
DB_URL=postgres://...@postgres-auth:5432/... ✅  uses compose service-name DNS
```

Same rule for service-to-service URLs:

```
AUTH_SVC_URL=localhost:50051                ❌
AUTH_SVC_URL=auth-svc:50051                 ✅
```

See [`04-container-networking.md`](./04-networking.md) for the full networking explanation.

## Per-environment overrides

Two approaches — pick one.

### Approach A — Multiple compose files (recommended)

```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml up
```

`docker-compose.prod.yml` only contains the differences:

```yaml
services:
  auth-svc:
    environment:
      JWT_SECRET_KEY: ${JWT_SECRET_KEY_PROD}
    deploy:
      resources:
        limits: { cpus: "0.5", memory: 256M }
```

### Approach B — Multiple env files

```bash
docker compose --env-file .env.staging up
```

## Inspecting env vars at runtime

```bash
# What did the auth-svc container actually receive?
docker compose exec auth-svc env | sort

# Just the ones we care about
docker compose exec auth-svc env | grep -E 'PORT|DB_URL|JWT'
```

If a value is missing or wrong, the answer is in this output.

## Secrets — what NOT to do

- ❌ Commit `.env` with real production secrets.
- ❌ Bake secrets into the Docker image with `ENV PASSWORD=xxx`.
- ❌ Log env vars (`fmt.Println("config:", c)` will print your JWT secret).
- ❌ Echo them in CI logs.

## Secrets — what to do

- ✅ Local: `.env` in `.gitignore`.
- ✅ CI: encrypted secrets in GitHub Actions → injected as env at build/deploy.
- ✅ Prod: AWS Secrets Manager / GCP Secret Manager / Vault. Fetch at startup.
- ✅ For Docker-only deployments, use Compose's `secrets:` block (writes files into `/run/secrets/<name>`).

## Common mistakes

- Forgetting `.env` exists → compose substitutes empty strings, services start with bad config.
- Using `$` inside a value without escaping → compose tries to substitute. Escape with `$$`.
- Quoting `:50051` in compose YAML correctly: YAML treats `:` in unquoted strings carefully. Safest: `PORT: ":50051"` with quotes.

## Related
- [`03-docker-compose-setup.md`](./03-compose-guide.md)
- [`06-postgres-container-setup.md`](./06-postgres-container-setup.md)
- [`10-production-best-practices.md`](./10-production-best-practices.md)
