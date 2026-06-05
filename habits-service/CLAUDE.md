# CLAUDE.md — habits-service

## Purpose
Owns the habits catalog, user habit assignments per group, daily check-ins,
streak calculation, and the Saturday roulette / penalty mechanic.
This is the core data service of the app.

## Core loop (what this service enables day-to-day)

```
1. Daily  → member submits check-in for each assigned habit
2. Daily  → service calculates streak and broadcasts via realtime-service
3. Saturday → any member triggers the roulette for the group
              → service finds who failed habits during the week (Mon–Fri)
              → picks a punishment from the catalog (JSON) at random
              → creates a debt record for the offender
              → broadcasts result via realtime-service
4. Debt auto-expires at the end of the following week (no manual intervention needed)
```

## This service does NOT
- Send real-time events directly — after a spin it fires a goroutine that
  POSTs to realtime-service `/internal/broadcast`; if realtime is down the
  spin still succeeds
- Handle group membership validation for most endpoints — it reads
  `group_members` directly (cross-service read, not write)
- Validate JWT — api-gateway does that; trust the `X-User-ID` header
- Own proposal logic — proposals/votes live in groups-service.
  This service exposes `/internal/proposals/apply` for groups-service
  to call when a proposal is approved.

## Endpoints

### Habits
| Method | Path | Auth | Description |
|---|---|---|---|
| GET | /habits | JWT | List available habits catalog |
| POST | /checkins | JWT | Submit daily check-in |
| GET | /checkins/:groupID/today | JWT | Get today's check-ins for the group |
| GET | /streaks/:userID | JWT | Get user's current streaks across all groups |

### Penalties
| Method | Path | Auth | Description |
|---|---|---|---|
| GET | /penalties/:groupID/eligible | JWT | Members who failed habits this week |
| POST | /penalties/spin | JWT | Spin roulette, assign punishment, trigger broadcast |
| GET | /penalties/:groupID | JWT | List active debts for the group |

### Internal (called by other services only — not exposed via api-gateway)
| Method | Path | Auth | Description |
|---|---|---|---|
| POST | /internal/proposals/apply | none | Apply approved proposal: assign or unassign a habit for all group members |

### Health
| Method | Path | Auth | Description |
|---|---|---|---|
| GET | /health | None | Keepalive ping |

## Environment Variables
```
PORT=8083
DATABASE_URL=...                  # Supabase PostgreSQL connection string
REALTIME_SERVICE_URL=http://...   # to trigger broadcast after spin
PUNISHMENT_CATALOG_PATH=./punishments.json
```

## Database Tables Owned
`habits` · `user_habits` · `checkins` · `roulette_entries` · `punishment_suggestions` · `debts`

Other services must NOT write to these tables directly.
Cross-service reads of `group_members` and `users` are acceptable.

## Internal Package Structure

```
internal/
├── model/
│   └── habit.go          <- Habit, UserHabit, TodayCheckin, Streak,
│                            RouletteEntry, Debt, EligibleMember, Punishment
├── db/
│   └── queries.go        <- all pgxpool queries + calculateStreak + CurrentWeekStart
└── handler/
    ├── habit_handler.go  <- HabitHandler (habits, checkins, streaks)
    ├── penalty_handler.go <- PenaltyHandler (roulette, debts)
    └── response.go       <- shared writeError / writeJSON
```

## Compiling & Running

Go is NOT installed locally — it runs inside Docker. To verify changes compile:
```bash
docker-compose build habits-service
docker-compose up -d habits-service
docker-compose logs -f habits-service
```

## Key Patterns

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
    port = "8083"
}
http.ListenAndServe(":"+port, r)
```

### Punishment catalog
Loaded from JSON at startup via `loadCatalog()` in `main.go`.
Panics on missing or empty file — a service without a catalog is invalid.
Random selection uses `crypto/rand`, not `math/rand`.

### Checkin body shape
`POST /checkins` accepts `{ "group_id": "...", "habit_id": "...", "note": "..." }`.
The service looks up `user_habit_id` internally — the frontend never needs it.

### Streak calculation
`calculateStreak` in `db/queries.go` counts consecutive days ending on
today OR yesterday (streak is still active if you haven't checked in yet today).
Dates must be sorted DESC. Uses UTC; Supabase also stores in UTC.

### Eligible members query
Members are eligible if they missed at least one habit on any day this week
(Mon–yesterday). On Monday the list is always empty — nothing to check yet.
The query uses `generate_series` to enumerate weekdays.

### Realtime broadcast after spin
`notifyRealtime` runs in a goroutine and ignores errors. A realtime-service
outage must never fail a spin. Timeout is 3 seconds.

### Debt expiry
Debts do NOT require manual resolution. Each debt has a `week_start` date;
the frontend filters and the query only returns debts whose
`week_start >= current_week_start - 1 week` (i.e. active and previous week).
Older debts are considered expired and are excluded from all queries.

### Proposal apply (internal endpoint)
`POST /internal/proposals/apply` receives `{ "group_id": "...", "habit_id": "...", "action": "add" | "remove" }`.
- `add`: inserts a `user_habits` row for every current group member who doesn't have it yet.
- `remove`: deletes `user_habits` rows for all group members for that habit.

## Service-Specific Notes
- `punishment_suggestions` captures punishments proposed by users (future feature).
  The spin currently draws from the JSON catalog only. `suggestion_id` in `debts`
  is always NULL until user-submitted suggestions are integrated into the spin.
- `GET /penalties/:groupID` returns only debts from the current and previous week
  to match the `idx_debts_week` index and avoid loading stale data.
