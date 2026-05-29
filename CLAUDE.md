# CLAUDE.md — Dydi (Root)

## What is this project?

Dydi is a social accountability SaaS where friend groups track daily habits and
gamify consequences. Built as a university research project (UTD Integradora 2025)
to validate the hypothesis: **can a microservices architecture deployed exclusively
on free-tier platforms maintain acceptable quality metrics for a real-time SaaS?**

---

## Architecture Overview

```
frontend/           → Vue 3 + Vite + Pinia + Tailwind (Vercel)
api-gateway/        → Go 1.22 + chi v5 (Render — Account 1)
groups-service/     → Go 1.22 + chi v5 (Render — Account 2)
habits-service/     → Go 1.22 + chi v5 (Render — Account 3)
realtime-service/   → Go 1.22 + nhooyr/websocket (Render — Account 4)
```

**Auth** is handled entirely by Supabase Auth — no auth-service exists.
The frontend uses the Supabase JS SDK for login/register/logout.
`api-gateway` validates Supabase JWTs and forwards `X-User-ID` to downstream services.

**Penalty logic** lives inside `habits-service` (same domain: check-ins trigger debts).
There is no separate penalties-service.

Each service is a **fully independent Go module** with its own `go.mod`,
`Dockerfile`, and `.env.example`. They communicate via HTTP internally.
The only public-facing entry point is `api-gateway`.

---

## CRITICAL RULES — Read before touching anything

### 1. Service isolation
**Never modify files outside the service directory you are currently working in.**
If a change requires touching another service, stop and ask explicitly.
Each service is deployed independently; a change in one must not break others.

### 2. Ask before "fixing" unconventional patterns
Some patterns in this codebase look wrong but are intentional due to
Render free-tier constraints. Examples:
- Health check endpoints that seem redundant → they keep the service awake
- Explicit sleep/retry logic on startup → Render cold starts can take 30s+
- Non-standard port binding using `$PORT` env var → required by Render
- Ping-friendly lightweight responses on `/health` → used by cron-job.org

**If something looks unconventional, ask before changing it.**

### 3. Environment variables
**Never hardcode secrets, URLs, or credentials.** Always use environment
variables. Every service has a `.env.example` file. Never commit `.env` files.

### 4. Dockerfiles are sacred
Each service has a multistage `Dockerfile` optimized for Render free tier
(small final image = faster cold start). Do not simplify or "clean up"
Dockerfiles without explicit instruction.

### 5. No global dependencies
Do not add a shared Go package or module that multiple services import.
Each service must be fully self-contained. Code duplication between services
is acceptable and intentional.

---

## Tech Stack (Official — do not deviate)

| Layer | Technology | Version |
|---|---|---|
| Frontend | Vue 3 (Composition API, `<script setup>`) | latest |
| Frontend state | Pinia | latest |
| Frontend styling | Tailwind CSS (utility classes only) | v3 |
| Frontend realtime | @vueuse/core `useWebSocket` | latest |
| Frontend hosting | Vercel (free) | — |
| Backend language | Go | 1.22 |
| HTTP router | go-chi/chi | v5 |
| WebSocket | nhooyr.io/websocket | latest |
| DB driver | jackc/pgx | v5 |
| Query generation | sqlc | latest |
| Database | PostgreSQL 15 | — |
| DB hosting | Supabase (free tier) | — |
| Auth | Supabase Auth + JWT | — |
| Backend hosting | Render (free tier) | — |
| Local dev | Docker + docker-compose | — |
| Metrics | Prometheus + Grafana | — |

---

## Naming Conventions

| Context | Convention | Example |
|---|---|---|
| Go files | snake_case | `habit_handler.go` |
| Go functions/types | PascalCase | `CreateHabit()` |
| Go variables | camelCase | `habitID` |
| Go constants | UPPER_SNAKE | `MAX_GROUP_SIZE` |
| Database tables | snake_case plural | `group_members` |
| Database columns | snake_case | `created_at` |
| API endpoints | kebab-case | `/api/group-members` |
| Vue components | PascalCase | `HabitCard.vue` |
| Vue composables | camelCase with `use` prefix | `useGroupSocket.js` |
| Env variables | UPPER_SNAKE | `DATABASE_URL` |

---

## Folder Structure (Expected)

```
dydi/
├── CLAUDE.md                     ← you are here
├── docker-compose.yml            ← local dev only
├── .gitignore
├── supabase/
│   └── migrations/
│       └── 001_initial.sql
├── api-gateway/
│   ├── CLAUDE.md
│   ├── Dockerfile
│   ├── .env.example
│   ├── go.mod
│   ├── go.sum
│   ├── main.go
│   └── internal/
│       ├── proxy/
│       └── middleware/
├── groups-service/
│   ├── CLAUDE.md
│   ├── Dockerfile
│   ├── .env.example
│   ├── go.mod
│   ├── go.sum
│   ├── main.go
│   └── internal/
│       ├── handler/
│       ├── model/
│       └── db/
├── habits-service/               ← also owns penalty/debt logic
│   ├── CLAUDE.md
│   ├── Dockerfile
│   ├── .env.example
│   ├── go.mod
│   ├── go.sum
│   ├── main.go
│   └── internal/
│       ├── handler/
│       ├── model/
│       └── db/
├── realtime-service/
│   ├── CLAUDE.md
│   ├── Dockerfile
│   ├── .env.example
│   ├── go.mod
│   ├── go.sum
│   ├── main.go
│   └── internal/
│       ├── hub/
│       └── handler/
└── frontend/
    ├── CLAUDE.md
    ├── Dockerfile
    ├── .env.example
    ├── package.json
    ├── vite.config.js
    └── src/
        ├── components/
        ├── composables/
        ├── stores/
        ├── views/
        └── router/
```

---

## Local Development

All services run locally via `docker-compose.yml` in the root.
To start everything:

```bash
docker-compose up --build
```

Local ports (do not change without updating docker-compose):

| Service | Local port |
|---|---|
| api-gateway | 8080 |
| groups-service | 8082 |
| habits-service | 8083 |
| realtime-service | 8084 |
| frontend | 5173 |
| postgres (local) | 5432 |

---

## Deployment Notes (Render Free Tier)

- Each service is deployed to a **separate Render account** to maximize
  free-tier hours.
- Services operate in an **active window: 08:00–22:00 CST**.
  Outside this window, cold starts are expected and acceptable.
- A cron job on **cron-job.org** pings each service's `/health` endpoint
  every 10 minutes between 08:00 and 22:00 to prevent sleep.
- The active window constraint is **documented as a research variable**
  in the academic paper, not hidden as a limitation.
- Render requires the app to bind on `$PORT` env var, not a hardcoded port.
  Every service must respect this.

---

## Research Metrics Being Collected

These are measured for the academic paper. Do not remove or alter the
instrumentation code related to these:

- HTTP response latency (P95)
- WebSocket cold start time
- Monthly uptime per service
- Concurrent WebSocket connections per group
- Check-in event delivery consistency (emitted vs received)

---

## Database Schema (Source of truth)

Schema lives in `supabase/migrations/`. Never modify the database directly.
All schema changes go through migration files. The canonical table list:

`users` · `groups` · `group_members` · `habits` · `user_habits` · `checkins` · `debts`

Full schema definition is in `supabase/migrations/001_initial.sql`.
