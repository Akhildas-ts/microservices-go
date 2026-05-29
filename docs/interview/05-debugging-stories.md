# Debugging War Stories

> The most powerful interview answers are **specific true stories**, not abstract knowledge. Write your own here as they happen. The format below is what to follow.

## Story template

```markdown
## YYYY-MM-DD — One-line title

**Context:** what I was doing and why.
**Symptom:** the exact error or behavior I saw.
**Investigation:** the commands I ran, the dead-ends I hit, the breakthrough.
**Root cause:** what was actually wrong.
**Fix:** the actual change I made.
**Lesson:** the principle I now apply.
```

---

## Example stories to copy the shape of

### 2026-05-29 — Auth service couldn't reach Postgres

**Context.** Bringing up the stack for the first time with the new compose file.

**Symptom.** auth-svc kept crashing at startup with:
```
dial tcp: lookup postgres-auth: no such host
```

**Investigation.** Ran `docker compose ps` → auth-svc was in `Restarting`. Ran `docker network inspect ecom-net` and noticed auth-svc was listed but `postgres-auth` was NOT.

**Root cause.** I'd forgotten `networks: [ecom-net]` on the `postgres-auth` block. Compose dropped it on the default network, isolated from `ecom-net`.

**Fix.** Added `networks: [ecom-net]` to the postgres block. `docker compose up -d` recreated it on the right network.

**Lesson.** Every service on a custom network MUST declare it. Without `networks:`, Compose attaches it to the default network — which is a separate network from yours.

---

### 2026-05-29 — `down -v` wiped the dev DB

**Context.** Wanted a fresh stack for testing.

**Symptom.** All previously seeded users were gone after `docker compose down -v && docker compose up -d`.

**Investigation.** Reread the docker compose docs. `-v` removes named volumes — and `pgdata-auth` was a named volume.

**Root cause.** Habit of typing `down -v` from a previous project where I had no persistent data.

**Fix.** Renamed the destructive Makefile target from `down` to `nuke`. Now muscle memory points at a name that screams "danger".

**Lesson.** Name dangerous commands with dangerous names.

---

### 2026-05-29 — Password `akhil@123` broke the DB URL

**Context.** Tried to migrate from local-only dev to dockerized stack.

**Symptom.** `pq: invalid URL: parse … net/url: invalid userinfo`

**Investigation.** Logged the resolved `DB_URL` and saw `postgres://postgres:akhil@123@postgres-auth:5432/...`. Two `@` signs.

**Root cause.** `@` inside the password collides with the URL `userinfo@host` separator.

**Fix.** URL-encoded `@` as `%40`. Updated the password to alphanumeric long-term.

**Lesson.** Database URLs are URIs. Anything not in `[a-zA-Z0-9_-]` needs encoding. Avoid special chars in passwords altogether.

---

## Add YOUR stories below as they happen

Each one becomes a 2-minute interview answer. Specific stories with real commands beat memorized definitions every time.

<!-- ## YYYY-MM-DD — ... -->
