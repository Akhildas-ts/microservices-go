# Engineering Documentation

Knowledge base for the **Go gRPC E-Commerce Microservices** project. This folder grows in **phases** — each phase has a clear goal, a defined "done", and its own body of docs.

> **You are currently in Phase 1 — Docker.** See [`PHASES.md`](./PHASES.md) for the full roadmap.

---

## Phase 1 — Docker 🐳 (current)

**Goal:** dockerize the entire stack end-to-end, learn Docker deeply, build interview-ready knowledge.

**Start here:**
1. [`architecture/01-overview.md`](./architecture/01-overview.md) — what the system is.
2. [`docker/01-docker-basics.md`](./docker/01-docker-basics.md) — Docker mental model.
3. [`docker/12-step-by-step-walkthrough.md`](./docker/12-step-by-step-walkthrough.md) — the step-by-step implementation.

---

## How to use this folder

| If you are…                       | Start here                                              |
|----------------------------------|---------------------------------------------------------|
| New to the project               | [`architecture/01-overview.md`](./architecture/01-overview.md) |
| Trying to run the stack          | [`docker/12-step-by-step-walkthrough.md`](./docker/12-step-by-step-walkthrough.md) |
| Hit an error                     | [`troubleshooting/README.md`](./troubleshooting/README.md) |
| Preparing for an interview       | [`interview/`](./interview/) |
| Learning Docker concepts         | [`docker/`](./docker/) |
| Writing a new doc                | [`STYLE.md`](./STYLE.md) |
| Wondering what's next            | [`PHASES.md`](./PHASES.md) |

---

## Folder structure (Phase 1)

```
docs/
├── README.md                          ← you are here
├── PHASES.md                          ← roadmap across all phases
├── STYLE.md                           ← how to write docs in this repo
│
├── architecture/                      ← the system we are dockerizing
│   ├── 01-overview.md
│   ├── 02-service-communication.md
│   ├── 03-data-flow.md
│   └── 04-decisions.md                ← ADRs
│
├── docker/                            ← Phase 1 focus
│   ├── 01-docker-basics.md
│   ├── 02-dockerfile-guide.md
│   ├── 03-compose-guide.md            (docker-compose setup)
│   ├── 04-networking.md               (container networking)
│   ├── 05-volumes.md                  (persistence)
│   ├── 06-postgres-container-setup.md
│   ├── 07-environment-variables.md
│   ├── 08-grpc-inside-docker.md       ← the most important page
│   ├── 09-debugging-docker.md
│   ├── 10-production-best-practices.md
│   ├── 11-commands-cheatsheet.md
│   └── 12-step-by-step-walkthrough.md ← the implementation guide
│
├── troubleshooting/                   ← symptom → cause → fix
│   ├── README.md
│   ├── grpc-issues.md
│   ├── postgres-issues.md
│   ├── network-and-dns.md
│   ├── container-crashes.md
│   ├── port-conflicts.md
│   └── docker-cache-issues.md
│
├── interview/                         ← interview prep
│   ├── 01-docker.md
│   ├── 02-microservices.md
│   ├── 03-grpc.md
│   ├── 04-scenario-based.md
│   └── 05-debugging-stories.md
│
└── learning-notes/                    ← your engineering journal
    ├── README.md
    ├── TEMPLATE.md
    └── mistakes-and-lessons.md
```

Folders for `kubernetes/`, `deployment/` (CI/CD), `observability/`, `security/` will be added in their respective phases — see [`PHASES.md`](./PHASES.md).

---

## The stack at a glance

```
                 ┌──────────────────────────┐
   Client ─HTTP─►│  API Gateway (Gin :3000) │
                 └──────────────┬───────────┘
                                │ gRPC
       ┌────────────┬───────────┼────────────┬────────────┐
       ▼            ▼           ▼            ▼            ▼
   ┌────────┐  ┌─────────┐  ┌────────┐  ┌────────┐  ┌────────┐
   │  Auth  │  │ Product │  │ Order  │  │ Admin  │  │  Cart  │
   │ :50051 │  │ :50052  │  │ :50053 │  │ :50054 │  │ :50055 │
   └────┬───┘  └────┬────┘  └───┬────┘  └────┬───┘  └────┬───┘
        │           │           │            │           │
        ▼           ▼           ▼            ▼           ▼
   ┌────────┐  ┌─────────┐  ┌────────┐  ┌────────┐  ┌────────┐
   │ PG     │  │ PG      │  │ PG     │  │ PG     │  │ PG     │
   │ auth   │  │ product │  │ order  │  │ admin  │  │ cart   │
   └────────┘  └─────────┘  └────────┘  └────────┘  └────────┘

   order-svc and cart-svc also call product-svc over gRPC.
```

See [`architecture/01-overview.md`](./architecture/01-overview.md) for the full picture.

---

## Contributing to the docs

- Read [`STYLE.md`](./STYLE.md) once before adding a page.
- Every problem you solve → an entry in `troubleshooting/` AND a one-liner in `learning-notes/mistakes-and-lessons.md`.
- Every architectural decision → an ADR in `architecture/04-decisions.md`.
- Docs go stale fast. Update them in the same PR as the code change.
