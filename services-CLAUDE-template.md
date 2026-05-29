# CLAUDE.md — [SERVICE NAME]
# Copy this file into each remaining service and fill in the blanks.
# Lines marked with ← FILL IN need to be completed per service.

## Purpose
← FILL IN: One paragraph describing what this service owns and is responsible for.

## This service does NOT
← FILL IN: Explicit list of what this service must NOT do (prevents scope creep).

## Endpoints

| Method | Path | Auth | Description |
|---|---|---|---|
← FILL IN: one row per endpoint

## Environment Variables
```
PORT=808X                    ← FILL IN port number
DATABASE_URL=...             # Supabase PostgreSQL connection string
SUPABASE_JWT_SECRET=...      # For validating forwarded tokens (if needed)
← FILL IN: any service-specific vars
```

## Database Tables Owned
← FILL IN: List which tables this service reads/writes.
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
    port = "808X" // ← FILL IN default port
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
The API gateway attaches `userID` to every authenticated request context.
Extract it like this — do not re-validate the JWT in this service:
```go
userID := r.Context().Value("userID").(string)
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
    ├── handler/         ← HTTP handlers (thin, delegate to service layer)
    ├── service/         ← business logic
    ├── model/           ← structs / domain types
    └── db/              ← sqlc generated files + queries
```

## Dockerfile Template (multistage — do not simplify)
```dockerfile
# Stage 1: Build
FROM golang:1.22-alpine AS builder
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
← FILL IN: Any patterns, decisions, or constraints unique to this service.

---

# Filled examples below for reference (active services only):

---

# CLAUDE.md — groups-service

## Purpose
Manages group creation, invite codes, membership, and group metadata.
Enforces the hard limit of 8 members per group.

## This service does NOT
- Handle check-ins or habit tracking (that is habits-service)
- Handle real-time events (that is realtime-service)
- Validate JWT (api-gateway does that)

## Endpoints

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | /groups | JWT | Create a new group |
| GET | /groups/:id | JWT | Get group details + members |
| POST | /groups/:id/join | JWT | Join via invite code |
| GET | /groups/:id/members | JWT | List members |
| DELETE | /groups/:id/leave | JWT | Leave group |
| GET | /health | None | Keepalive ping |

## Environment Variables
```
PORT=8082
DATABASE_URL=...
MAX_GROUP_SIZE=8
```

## Database Tables Owned
`groups` · `group_members`

## Service-Specific Notes
MAX_GROUP_SIZE=8 is enforced at the service layer before any DB write.
This is both a product constraint and a free-tier protection measure.

---

# CLAUDE.md — habits-service

## Purpose
Owns the habits catalog, user habit assignments per group, daily check-ins,
and streak calculation. This is the core data service of the app.

## This service does NOT
- Send real-time events (after saving a check-in, it returns 200 and the
  api-gateway notifies realtime-service separately)
- Handle group membership validation (calls groups-service for that)

## Endpoints

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | /habits | JWT | List available habits catalog |
| POST | /habits/assign | JWT | Assign habit to self in a group |
| POST | /checkins | JWT | Submit daily check-in |
| GET | /checkins/:groupID/today | JWT | Get today's check-ins for group |
| GET | /streaks/:userID | JWT | Get user's current streaks |
| GET | /health | None | Keepalive ping |

## Environment Variables
```
PORT=8083
DATABASE_URL=...
```

## Database Tables Owned
`habits` · `user_habits` · `checkins` · `debts`

## Service-Specific Notes
Streak calculation runs on every check-in. It counts consecutive days
with at least one check-in. A streak breaks if no check-in was recorded
the previous calendar day (CST timezone).

**Penalty logic is owned by this service** (no separate penalties-service exists).
The Saturday roulette mechanic, debt assignment, and punishment catalog all
live here alongside the habits domain.

Endpoints for the penalty subdomain:

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | /penalties/:groupID/eligible | JWT | Members who failed habits this week |
| POST | /penalties/spin | JWT | Spin roulette, assign punishment, trigger broadcast |
| GET | /penalties/:groupID | JWT | List pending debts for group |
| PATCH | /penalties/:id/resolve | JWT | Mark debt as resolved |

The punishment catalog is a JSON file loaded at startup (`punishments.json`),
not stored in the DB. Randomness uses `crypto/rand`, not `math/rand`.
