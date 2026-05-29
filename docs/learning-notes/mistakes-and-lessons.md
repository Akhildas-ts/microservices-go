# Mistakes & Lessons

> One-liner log. Keep it scannable — append, never delete.

| Date | Mistake | Lesson |
|------|---------|--------|
| 2026-05-29 | Used `localhost:50051` inside container, got `connection refused` | Inside a container, `localhost` = that container. Use compose service names. |
| 2026-05-29 | `docker compose down -v` wiped my DB | `-v` removes volumes. Use plain `down` to keep data. |
| 2026-05-29 | Password `akhil@123` broke the Postgres URL parser | URL-encode `@` as `%40` or use alphanumeric passwords only. |
| 2026-05-29 | Service couldn't see another despite both running | Both must declare `networks: [ecom-net]` — missing one drops it on the default network. |
| 2026-05-29 | App started before Postgres ready, crashed | Add `healthcheck:` on the DB + `depends_on.condition: service_healthy` on the app. |
| 2026-05-29 | `POSTGRES_PASSWORD` change didn't take effect | Postgres bakes it on FIRST boot only. Use `ALTER USER` on the existing instance, or `down -v` to re-init. |

## Add yours below as they happen

<!--
| YYYY-MM-DD | <what went wrong> | <principle to remember> |
-->
