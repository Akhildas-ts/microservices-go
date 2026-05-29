# Debugging Docker Issues — A Field Guide

> When something is broken, work top-to-bottom. Don't randomly try fixes — diagnose first.

## The standard diagnostic ladder

Run these in order, every time:

```bash
# 1. What's actually running?
docker compose ps

# 2. Why did anything stop? Look at exit codes.
docker inspect ecom-auth-svc --format='{{.State.Status}} exit={{.State.ExitCode}} err={{.State.Error}}'

# 3. What does the service say in its logs?
docker compose logs --tail=200 auth-svc

# 4. What env vars / config does it actually see?
docker compose exec auth-svc env | sort

# 5. Can it reach its dependencies?
docker compose exec auth-svc sh
> getent hosts postgres-auth      # DNS
> nc -zv postgres-auth 5432       # TCP connectivity (if nc installed)
```

That's the loop. 90% of bugs surface in one of these five steps.

## Reading `docker compose ps` like a pro

```
NAME                  STATUS                           PORTS
ecom-postgres-auth    Up 2 minutes (healthy)
ecom-auth-svc         Restarting (1) 12 seconds ago
ecom-api-gateway      Up 1 minute                      0.0.0.0:3000->3000/tcp
```

Decode the STATUS column:

| STATUS                          | Meaning |
|---------------------------------|---------|
| `Up X minutes`                  | Running. No healthcheck declared. |
| `Up X minutes (healthy)`        | Running AND last healthcheck passed. |
| `Up X minutes (unhealthy)`      | Running but healthcheck failing. Dependencies waiting. |
| `Up X minutes (health: starting)` | Within `start_period`, healthcheck not yet evaluated. |
| `Restarting (N) X seconds ago`  | Crashing in a loop. Exit code in parens. |
| `Exited (N)`                    | Stopped. N is exit code; 0 = clean, anything else = crashed. |

## Streaming logs the useful way

```bash
docker compose logs -f                     # follow all services
docker compose logs -f --tail=100 auth-svc # follow one with last 100 lines
docker compose logs --since=5m             # last 5 minutes across all
docker compose logs auth-svc | grep -i error
```

> 💡 If logs are silent: the container might be writing to a file inside instead of stdout. Convention for cloud-native apps: **log to stdout/stderr only**. Files inside containers are wasted bytes nobody can see.

## Getting INSIDE a container

```bash
# Open a shell in a running container
docker compose exec auth-svc sh

# If the container won't stay up, run a fresh one interactively
docker compose run --rm auth-svc sh
# inside: ./app          # run the binary manually to see the error

# One-off command without a shell
docker compose exec auth-svc cat /app/pkg/config/envs/dev.env
docker compose exec auth-svc ls -la /app
docker compose exec auth-svc env | sort
```

> 📌 `exec` requires the container to be running. `run` starts a new one — useful when `up` keeps crashing.

## Inspecting a container's full configuration

```bash
docker inspect ecom-auth-svc                   # giant JSON
docker inspect ecom-auth-svc --format='{{json .Config.Env}}' | jq
docker inspect ecom-auth-svc --format='{{.State.Status}} {{.State.ExitCode}}'
docker inspect ecom-auth-svc --format='{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}'
```

Common things to extract:
- Exit code (`.State.ExitCode`)
- Env vars (`.Config.Env`)
- Mounts (`.Mounts`)
- IP address on each network (`.NetworkSettings.Networks`)

## Network and DNS troubleshooting

```bash
# Who's on the network?
docker network inspect ecom-net

# Can container A resolve container B's name?
docker compose exec api-gateway sh
> getent hosts auth-svc        # returns "172.20.0.5 auth-svc" if OK

# Can container A reach container B's port?
> nc -zv auth-svc 50051        # "open" = success
# If `nc` isn't installed in alpine:
> apk add netcat-openbsd
```

If `getent hosts` returns nothing → DNS broken → both containers must be on the same network. Add `networks: [ecom-net]` to both.

## Crash loop diagnosis

```bash
# 1. Confirm it's looping
docker compose ps | grep ecom-auth-svc       # STATUS = "Restarting (1)"

# 2. Get the last error before the crash
docker compose logs --tail=100 auth-svc

# 3. If the logs are empty, try running interactively
docker compose run --rm auth-svc sh
> ./app
# Now you see the actual panic / log.Fatalln
```

Typical causes:
- Env file missing → `Failed at config` → check `COPY` step in Dockerfile.
- DB not ready → add `condition: service_healthy` on the depends_on.
- Bad DB URL → URL-encoding or wrong hostname.
- Panic at startup → look at the stack trace, find the line, fix the bug.

## Image / build cache problems

> Symptom: "I changed the code but the container still runs the old code."

```bash
# 1. Force a rebuild
docker compose up -d --build

# 2. If the layer cache is too aggressive
docker compose build --no-cache auth-svc

# 3. Confirm the binary is actually new
docker compose exec auth-svc md5sum /app/app   # compare across runs

# 4. Nuclear option
docker compose down
docker image rm ecom/auth-svc:dev
docker compose build --no-cache auth-svc
docker compose up -d
```

See [`../troubleshooting/docker-cache-issues.md`](../troubleshooting/docker-cache-issues.md).

## Resource starvation

```bash
docker stats                    # live CPU/MEM per container
docker system df                # disk usage by images/containers/volumes
```

If one container hogs CPU and slows others, add limits:

```yaml
deploy:
  resources:
    limits: { cpus: "0.5", memory: 256M }
```

## Events stream — what happened, and when?

```bash
docker events --since '10m'
```

Shows every container start/stop/health transition. Great for "I think my service died at 12:03 but I'm not sure why" investigations.

## A debugging checklist to live by

When something breaks, ask in this order:

1. **Is it running?** `docker compose ps`
2. **Why did it stop?** `docker inspect ... ExitCode`
3. **What does it say?** `docker compose logs`
4. **What config did it get?** `docker compose exec ... env`
5. **Can it reach its dependencies?** `getent hosts`, `nc -zv`
6. **Did my last change get picked up?** `--build`, `--no-cache`, `md5sum`

Memorize that ladder. It's the difference between "Docker is magic" and "Docker is debuggable".

## Related
- [`../troubleshooting/README.md`](../troubleshooting/README.md)
- [`11-commands-cheatsheet.md`](./11-commands-cheatsheet.md)
