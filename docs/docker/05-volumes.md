# Docker Volumes

> Containers are ephemeral. Volumes are how you keep data alive across container lifecycles.

## Why volumes exist

A container's filesystem is destroyed when the container is removed. For Postgres data, uploaded files, or any persistent state, you need storage that **outlives the container**.

## Three kinds of mounts

| Kind | Lifecycle | Managed by | Best for |
|---|---|---|---|
| **Named volume** | Survives `docker rm`, wiped by `docker volume rm` | Docker | Databases, anything Docker should manage |
| **Bind mount** | Lives on the host filesystem | You | Live code reload in dev, log files, init scripts |
| **tmpfs** | RAM-only, gone on container stop | Docker | Secrets you don't want on disk |

## Named volumes — our Postgres pattern

```yaml
services:
  postgres-auth:
    image: postgres:16-alpine
    volumes:
      - pgdata-auth:/var/lib/postgresql/data    # named volume

volumes:
  pgdata-auth:                                  # declared at top level
```

Inspect:

```bash
docker volume ls                                # all volumes
docker volume inspect pgdata-auth               # mountpoint, driver, options
```

The actual data lives at the path shown in `Mountpoint` — typically `/var/lib/docker/volumes/<name>/_data` on Linux, inside the Docker VM on macOS/Windows.

## Bind mounts — dev-only

Map a host folder into the container — handy for live-edit workflows.

```yaml
# docker-compose.override.yml
services:
  auth-svc:
    volumes:
      - ./go-grpc-auth-svc:/src      # source mount for live reload
    command: go run ./cmd            # run the source, not the baked binary
```

> ⚠️ Don't ship bind mounts in production compose files — they tie the container to the host's directory layout, which won't exist on a server.

## Init scripts via bind mount

Postgres runs anything in `/docker-entrypoint-initdb.d/` on **first boot only** (i.e. when the data dir is empty). Useful for creating extra databases or seeding data:

```yaml
postgres-auth:
  volumes:
    - pgdata-auth:/var/lib/postgresql/data
    - ./infra/postgres/init:/docker-entrypoint-initdb.d:ro   # read-only mount
```

```sql
-- infra/postgres/init/01-create-extra.sql
CREATE DATABASE other_db;
```

> 📌 **It only runs once.** If you change the script later, you have to `docker compose down -v` to wipe the volume and re-init. Otherwise the script is silently ignored.

## Common commands

```bash
# List & inspect
docker volume ls
docker volume inspect pgdata-auth

# Remove
docker volume rm pgdata-auth                  # only if no container uses it

# Prune all unused volumes (DANGEROUS — read carefully)
docker volume prune

# Back up a volume to a tarball on the host
docker run --rm \
  -v pgdata-auth:/data \
  -v "$PWD":/backup \
  alpine tar czf /backup/pgdata-auth.tgz -C /data .

# Restore from tarball
docker run --rm \
  -v pgdata-auth:/data \
  -v "$PWD":/backup \
  alpine sh -c "cd /data && tar xzf /backup/pgdata-auth.tgz"
```

## `down` vs `down -v` — memorize this

| Command | Removes containers? | Removes volumes? | DB data survives? |
|---|---|---|---|
| `docker compose stop` | No | No | Yes |
| `docker compose down` | Yes | **No** | **Yes** |
| `docker compose down -v` | Yes | **Yes** | **NO — data wiped** |

> ⚠️ `-v` is destructive. Don't add it to a Makefile target called `down` — make it a separate target named `nuke` or `wipe`.

## Volume per service, or one shared?

We use **one volume per Postgres service** (`pgdata-auth`, `pgdata-product`, …) so each database is fully isolated. Backing up or wiping one doesn't affect the others — same reasoning as one DB per service.

## Pitfalls

- **Mount a non-empty host dir over a populated container dir** → container's contents are hidden by the host dir's contents. Common gotcha with `node_modules`.
- **Forgetting the `:ro` on an init script mount** → container could modify your repo files.
- **Different volume drivers in dev vs prod** → behavior differs; stick to default `local` unless you have a reason.

## Related
- [`database/01-postgres-setup.md`](../database/01-postgres-setup.md)
- [`database/04-backup-restore.md`](../database/04-backup-restore.md)
