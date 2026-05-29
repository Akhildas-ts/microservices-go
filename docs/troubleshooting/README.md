# Troubleshooting Index

> Symptom → likely cause → diagnose → fix. Always look here before guessing.

## Format we use in every entry

```markdown
### Symptom
<exact text of the error or the user-visible behavior>

### Likely cause
<one or two sentences>

### Diagnose
<commands or steps to confirm>

### Fix
<the actual change to make>
```

This is the **only** format. Don't write essay-style troubleshooting docs — they don't get re-read.

## Files in this folder

| File | Covers |
|---|---|
| [`grpc-issues.md`](./grpc-issues.md) | `Unavailable`, `DEADLINE_EXCEEDED`, `no such host`, listen-address bugs |
| [`postgres-issues.md`](./postgres-issues.md) | Connection refused, auth failed, DB doesn't exist, password rot |
| [`network-and-dns.md`](./network-and-dns.md) | Containers can't see each other, wrong network membership |
| [`container-crashes.md`](./container-crashes.md) | Crash loops, exit codes, config load failures |
| [`port-conflicts.md`](./port-conflicts.md) | `bind: address already in use` |
| [`docker-cache-issues.md`](./docker-cache-issues.md) | "I changed the code but it's running the old version" |

## The universal diagnostic ladder

Before you open a specific file, walk this ladder — it solves most things:

```bash
# 1. What's running?
docker compose ps

# 2. Why did anything stop?
docker inspect ecom-auth-svc --format='{{.State.Status}} exit={{.State.ExitCode}}'

# 3. Logs
docker compose logs --tail=200 auth-svc

# 4. Env vars actually received
docker compose exec auth-svc env | sort

# 5. Network/DNS
docker compose exec auth-svc sh -c 'getent hosts postgres-auth'
```

See [`docker/09-debugging-docker.md`](../docker/09-debugging-docker.md) for the long version.

## When you fix something new

1. Add an entry to the relevant file using the format above.
2. Add a one-liner to `learning-notes/mistakes-and-lessons.md`.

Your future self will love you.
