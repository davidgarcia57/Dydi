# CLAUDE.md — [SERVICE NAME]
# Copy this file into each remaining service and fill in the blanks.
# Lines marked with <- FILL IN need to be completed per service.

## Purpose
<- FILL IN: One paragraph describing what this service owns and is responsible for.

## This service does NOT
<- FILL IN: Explicit list of what this service must NOT do (prevents scope creep).

## Endpoints

| Method | Path | Auth | Description |
|---|---|---|---|
<- FILL IN: one row per endpoint

## Environment Variables
```
PORT=808X                       <- FILL IN port number
DATABASE_URL=...                # Supabase PostgreSQL connection string
SUPABASE_JWKS_URL=https://...   # only if this service validates JWTs directly
<- FILL IN: any service-specific vars
```

## Database Tables Owned
<- FILL IN: List which tables this service reads/writes.
Other services must NOT write to these tables directly.
Cross-service data needs go through HTTP calls to this service.

## Key Patterns (apply to ALL Go services)

### Health endpoint — do not remove or alter
```go
r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("ok"))
})
```

### Port binding — always use $PORT
```go
port := os.Getenv("PORT")
if port == "" {
    port = "808X" // <- FILL IN default port
}
http.ListenAndServe(":"+port, r)
```

### Error handling convention
Always return structured JSON errors, never plain text:
```go
func writeError(w http.ResponseWriter, status int, msg string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
```

### Database connection pool (pgx)
Use `pgxpool`, not single connections. Pool is initialized once at startup
and passed via dependency injection, never as a global variable.
```go
pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
```

### Request context — userID
The API gateway attaches the authenticated user's ID as a header.
Read it from the header — do not re-validate the JWT in this service:
```go
userID := r.Header.Get("X-User-ID")
```

### Folder structure (standard for all Go services)
```
[service-name]/
├── CLAUDE.md
├── Dockerfile
├── .env.example
├── go.mod
├── go.sum
├── main.go              ← wiring only, no business logic
└── internal/
    ├── handler/         ← HTTP handlers (thin, delegate to db layer)
    ├── model/           ← structs / domain types
    └── db/              ← pgxpool queries
```

## Dockerfile Template (multistage — do not simplify)
```dockerfile
# Stage 1: Build
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/service ./main.go

# Stage 2: Final (keep image small for fast Render cold start)
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/service .
CMD ["./service"]
```

## Service-Specific Notes
<- FILL IN: Any patterns, decisions, or constraints unique to this service.

---

# Filled examples below for reference (active services only):

---

# CLAUDE.md — groups-service

## Purpose
Manages group creation, invite codes, membership, and habit proposals.
Enforces the hard limit of 8 members per group.

A proposal is a democratic request to add or remove a shared habit for the
entire group. Any member can open a proposal; it passes when a simple majority
votes yes. When approved, groups-service notifies habits-service to assign or
unassign the habit for all current members.

## This service does NOT
- Handle check-ins or habit tracking (that is habits-service)
- Handle real-time events (that is realtime-service)
- Validate JWT — api-gateway does that; trust the `X-User-ID` header

## Endpoints

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | /groups | JWT | Create a new group |
| GET | /groups/:id | JWT | Get group details + members |
| POST | /groups/:id/join | JWT | Join via invite code |
| GET | /groups/:id/members | JWT | List members |
| DELETE | /groups/:id/leave | JWT | Leave group |
| POST | /groups/:id/proposals | JWT | Open a new proposal (add_habit or remove_habit) |
| GET | /groups/:id/proposals | JWT | List open and recent proposals |
| POST | /proposals/:id/vote | JWT | Cast a vote (yes/no) |
| GET | /health | None | Keepalive ping |

## Environment Variables
```
PORT=8082
DATABASE_URL=...
MAX_GROUP_SIZE=8
HABITS_SERVICE_URL=http://...   # notified when a proposal is approved
```

## Database Tables Owned
`groups` · `memberships` · `proposals` · `proposal_eligible_voters` · `proposal_votes`

## Service-Specific Notes
- MAX_GROUP_SIZE=8 is enforced at the service layer before any DB write.
  This is both a product constraint and a Render free-tier protection measure.
- JoinGroup returns 404 (not 403) on wrong invite code to avoid leaking group existence.
- The call to habits-service on proposal approval runs in a goroutine — a
  habits-service outage must not block or fail the vote response.

---

# CLAUDE.md — habits-service

## Purpose
Owns the habits catalog, user habit assignments per group, daily check-ins,
streak calculation, and the Saturday roulette / penalty mechanic.
This is the core data service of the app.

## This service does NOT
- Send real-time events directly — after a spin it fires a goroutine that
  POSTs to realtime-service `/internal/broadcast`; if realtime is down the
  spin still succeeds
- Handle group membership validation — it reads `memberships` directly
  (cross-service read, not write)
- Validate JWT — api-gateway does that; trust the `X-User-ID` header
- Own proposal logic — proposals live in groups-service. This service
  exposes `/internal/proposals/apply` for groups-service to call.

## Endpoints

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | /habits | JWT | List available habits catalog |
| POST | /checkins | JWT | Submit daily check-in |
| GET | /checkins/:groupID/today | JWT | Get today's check-ins for group |
| GET | /streaks/:userID | JWT | Get user's current streaks |
| GET | /penalties/:groupID/eligible | JWT | Members who failed habits this week |
| POST | /penalties/spin | JWT | Spin roulette, assign punishment, trigger broadcast |
| GET | /penalties/:groupID | JWT | List active debts for the group |
| POST | /internal/proposals/apply | none | Apply approved proposal (called by groups-service) |
| GET | /health | None | Keepalive ping |

## Environment Variables
```
PORT=8083
DATABASE_URL=...
REALTIME_SERVICE_URL=http://...
PUNISHMENT_CATALOG_PATH=./punishments.json
```

## Database Tables Owned
`habits` · `user_habits` · `checkins` · `roulette_entries` · `punishment_suggestions` · `debts`

## Service-Specific Notes
- Debts auto-expire at the end of the following week. There is no manual
  resolution endpoint. Queries filter by `week_start` to exclude stale debts.
- The punishment catalog is a JSON file loaded at startup (`punishments.json`).
  Randomness uses `crypto/rand`, not `math/rand`.
- `POST /internal/proposals/apply` receives `{ "group_id", "habit_id", "action": "add"|"remove" }`.
  `add` inserts user_habits rows for all current members. `remove` deletes them.
