# Docker Commands Cheatsheet

> Daily-use commands. Keep this open in a tab when working.

## Lifecycle

```bash
# Build
docker compose build                       # build all images
docker compose build auth-svc              # build one
docker compose build --no-cache            # ignore layer cache

# Start
docker compose up -d                       # all services, background
docker compose up                          # foreground (see logs streamed)
docker compose up -d auth-svc              # one service + its deps
docker compose up -d --build               # build then start

# Stop
docker compose stop                        # stop without removing
docker compose down                        # stop AND remove containers (KEEP volumes)
docker compose down -v                     # stop + remove + WIPE volumes ⚠️ DATA LOSS

# Restart
docker compose restart                     # restart everything
docker compose restart auth-svc            # restart one
```

## Visibility

```bash
docker compose ps                          # running containers + status + ports
docker compose top                         # processes inside containers
docker stats                               # live CPU/MEM per container
docker system df                           # disk usage by images/containers/volumes
docker events --since '10m'                # recent docker engine events
```

## Logs

```bash
docker compose logs -f                     # follow all
docker compose logs -f --tail=100 auth-svc # follow one with last 100 lines
docker compose logs --since=5m             # last 5 minutes
docker compose logs auth-svc | grep -i error
```

## Inspecting

```bash
docker inspect ecom-auth-svc | less
docker inspect ecom-auth-svc --format='{{.State.Status}} {{.State.ExitCode}}'
docker inspect ecom-auth-svc --format='{{json .Config.Env}}' | jq

docker network ls
docker network inspect ecom-net

docker volume ls
docker volume inspect pgdata-auth

docker image ls
docker image history ecom/auth-svc:dev     # see the layers
```

## Getting inside

```bash
# Shell into a RUNNING container
docker compose exec auth-svc sh

# Run a fresh container (when the normal one keeps crashing)
docker compose run --rm auth-svc sh

# One-off commands
docker compose exec auth-svc env | sort
docker compose exec auth-svc cat /app/pkg/config/envs/dev.env
docker compose exec auth-svc ls -la /app

# psql into a DB
docker compose exec postgres-auth psql -U postgres -d e_auth_svc
```

## Network probing

```bash
docker compose exec api-gateway sh
> getent hosts auth-svc                    # DNS lookup
> nc -zv auth-svc 50051                    # TCP probe (apk add netcat-openbsd)
> wget -qO- http://auth-svc:50051/         # HTTP probe (will fail for gRPC, but tests reachability)
```

## Cleanup

```bash
# Safe-ish
docker image prune -f                      # remove dangling images
docker container prune -f                  # remove stopped containers
docker network prune -f                    # remove unused networks

# Dangerous — read carefully
docker volume prune                        # remove unused volumes (asks)
docker system prune -a --volumes           # NUCLEAR — wipe almost everything unused
```

## Recommended `Makefile` (root of repo)

```makefile
COMPOSE := docker compose

.PHONY: help build rebuild up down nuke logs ps shell-auth psql-auth

help:           ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	  awk 'BEGIN{FS=":.*## "} {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build:          ## Build all images
	$(COMPOSE) build

rebuild:        ## Rebuild ignoring cache
	$(COMPOSE) build --no-cache

up:             ## Start the stack (background)
	$(COMPOSE) up -d

down:           ## Stop + remove containers (KEEP volumes)
	$(COMPOSE) down

nuke:           ## Stop + remove containers + WIPE volumes ⚠️
	$(COMPOSE) down -v --rmi local --remove-orphans

logs:           ## Tail logs from all services
	$(COMPOSE) logs -f --tail=100

logs-auth:      ## Tail auth-svc logs
	$(COMPOSE) logs -f auth-svc

ps:             ## Show running containers
	$(COMPOSE) ps

shell-auth:     ## Exec sh into auth-svc
	$(COMPOSE) exec auth-svc sh

psql-auth:      ## psql into auth DB
	$(COMPOSE) exec postgres-auth psql -U $${PG_USER:-postgres} -d e_auth_svc
```

Then daily: `make up`, `make logs`, `make down`. `make help` lists everything.

## Power-user one-liners

```bash
# Show env vars of every running container, side by side
for c in $(docker compose ps -q); do
  echo "=== $(docker inspect $c --format='{{.Name}}') ==="
  docker exec $c env | grep -E 'PORT|DB_URL|SVC_URL|JWT' | sort
done

# Tail logs of one service while restarting it
docker compose restart auth-svc && docker compose logs -f auth-svc

# Find which container holds a specific port on the host
docker ps --format '{{.Names}}\t{{.Ports}}' | grep 3000

# Quick "is everything healthy?" check
docker compose ps --format json | jq -r '.[] | "\(.Name)\t\(.Status)"'
```

## Related
- [`09-debugging-docker.md`](./09-debugging-docker.md)
- [`03-docker-compose-setup.md`](./03-compose-guide.md)
