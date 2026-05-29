# Learning Notes

> Your engineering journal. Future-you will thank present-you.

## Why this folder exists

The fastest learners aren't the ones with the best memory — they're the ones who **write things down** as they happen. This folder is for:

- Daily learning sessions while dockerizing / migrating to K8s.
- Surprises, dead-ends, breakthroughs.
- One-liner lessons you don't want to forget.
- Decisions you made and the reasoning.

These notes become **gold during interviews** — they're the source of your war stories.

## Files in this folder

| File | Purpose |
|---|---|
| `TEMPLATE.md` | Copy this for every new dated note. |
| `YYYY-MM-DD-<topic>.md` | One file per learning session. Dated, kebab-case. |
| `mistakes-and-lessons.md` | Running one-liner log. Quick scan, big value. |

(Future) Subfolders for major decisions:
- `architecture-decisions/` — short ADRs ("Decided X because Y").
- `concepts-revision.md` — flashcard-style notes for spaced repetition.

## Daily routine

1. **Before you start.** Open today's file (`cp TEMPLATE.md 2026-05-29-dockerizing-day-1.md`). Write what you intend to learn.
2. **While working.** Jot down WTFs, error messages, surprises, the commands that worked.
3. **After.** Write a 3-line summary at the top: WHAT you learned, WHY it matters, HOW you'll apply it.
4. **Always.** Move any reusable principle into `mistakes-and-lessons.md`.

## What makes a good note?

- **Specific** beats general. "Postgres password rotation requires `ALTER USER`" beats "be careful with secrets".
- **Reproducible** beats vague. Include the commands you ran.
- **Honest** beats heroic. Write the dead-ends, not just the wins — that's where the lesson lives.

## How this folder supports each phase

| Phase | Note-keeping focus |
|---|---|
| Phase 1 (Docker) | Build/run/debug surprises, networking, persistence, secrets. |
| Phase 2 (K8s) | Manifest mistakes, probe tuning, RBAC traps, helm vs raw. |
| Phase 3 (CI/CD) | Pipeline failures, caching strategies, deploy rollback drills. |
| Phase 4 (Observability) | What a good log line looks like, what a useful metric looks like. |
