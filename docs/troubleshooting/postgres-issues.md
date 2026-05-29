# Postgres Issues

## `dial tcp: lookup postgres-auth: no such host`

### Likely cause
The app container isn't on the same Docker network as `postgres-auth`.

### Diagnose
```bash
docker network inspect ecom-net | grep -E 'Name|postgres-auth|auth-svc'
```

### Fix
Add `networks: [ecom-net]` to both services in `docker-compose.yml`. Recreate: `docker compose up -d`.

---

## `connection refused` to `postgres-auth:5432`

### Likely cause
The app started before Postgres was ready to accept connections.

### Diagnose
```bash
docker compose ps                       # is postgres-auth (healthy)?
docker compose logs postgres-auth | tail -50
```

### Fix
Add a healthcheck on the Postgres service and `condition: service_healthy` on the dependent:

```yaml
postgres-auth:
  healthcheck:
    test: ["CMD-SHELL", "pg_isready -U ${PG_USER} -d e_auth_svc"]
    interval: 5s
    timeout: 3s
    retries: 10

auth-svc:
  depends_on:
    postgres-auth: { condition: service_healthy }
```

---

## `FATAL: password authentication failed for user "postgres"`

### Likely cause
The password baked into the Postgres data dir (set on first boot) doesn't match what the app is sending now.

### Diagnose
```bash
docker compose exec postgres-auth env | grep POSTGRES_PASSWORD
docker compose exec auth-svc env | grep DB_URL
```

### Fix
Postgres bakes the password on **first boot only**. Changing `POSTGRES_PASSWORD` later does nothing to the existing user. Options:
1. Wipe and re-init: `docker compose down -v` (loses data) then `up`.
2. Keep data, update password manually:
   ```bash
   docker compose exec postgres-auth psql -U postgres -c \
     "ALTER USER postgres WITH PASSWORD 'NewPass';"
   # Then update .env and recreate the app container.
   ```

---

## `database "e_auth_svc" does not exist`

### Likely cause
Previous Postgres container booted with a different `POSTGRES_DB`. The old data dir was reused, so the DB you want was never created.

### Diagnose
```bash
docker compose exec postgres-auth psql -U postgres -l   # list databases
```

### Fix
- Cleanest: `docker compose down -v` to wipe the volume, then `up` (Postgres will create `e_auth_svc` fresh).
- Or create it by hand:
  ```bash
  docker compose exec postgres-auth psql -U postgres -c "CREATE DATABASE e_auth_svc;"
  ```

---

## `pq: invalid URL: parse … net/url: invalid userinfo`

### Likely cause
Special character (typically `@`) in the password isn't URL-encoded.

### Diagnose
```bash
docker compose exec auth-svc env | grep DB_URL
```

Spot the password section; `@` in the password collides with the user/host separator.

### Fix
Either:
- URL-encode: `akhil@123` → `akhil%40123`.
- Or change the password to alphanumeric (`[a-zA-Z0-9_-]`).

This is the most common silent bug in this project.

---

## `too many connections`

### Likely cause
App opens a fresh DB connection per request instead of pooling, OR you have many containers + a small `max_connections`.

### Diagnose
```bash
docker compose exec postgres-auth psql -U postgres -c \
  "select count(*) from pg_stat_activity;"
```

### Fix
- GORM uses a pool; ensure you're using one shared `*gorm.DB` (we do — `db.Handler` is shared).
- If genuinely too many: raise `max_connections` in Postgres or use PgBouncer.

---

## Data disappeared after `docker compose down`

### Likely cause
You ran `down -v` (the `-v` removes named volumes).

### Diagnose
Check `Makefile` and shell history for `-v`.

### Fix
- Use `docker compose down` (no `-v`) when you want to keep data.
- In the `Makefile`, name the destructive target `nuke` to make it obvious.

---

## Related
- [`../docker/06-postgres-container-setup.md`](../docker/06-postgres-container-setup.md)
- [`../docker/05-volumes.md`](../docker/05-volumes.md)
