# CLAUDE.md — habits-service

## Purpose
Owns the habits catalog, user habit assignments per group, daily check-ins,
streak calculation, and the Saturday roulette / penalty mechanic.
This is the core data service of the app.

## This service does NOT
- Send real-time events directly — after a spin it fires a goroutine that
  POSTs to realtime-service `/internal/broadcast`; if realtime is down the
  spin still succeeds
- Handle group membership validation for most endpoints — it reads
  `group_members` directly (cross-service read, not write)
- Validate JWT — api-gateway does that, trust X-User-ID header

## Endpoints

### Habits
| Method | Path | Auth | Description |
|---|---|---|---|
| GET | /habits | JWT | List available habits catalog |
| POST | /habits/assign | JWT | Assign habit to self in a group |
| POST | /checkins | JWT | Submit daily check-in |
| GET | /checkins/:groupID/today | JWT | Get today's check-ins for group |
| GET | /streaks/:userID | JWT | Get user's current streaks across all groups |

### Penalties (owned by this service — no separate penalties-service)
| Method | Path | Auth | Description |
|---|---|---|---|
| GET | /penalties/:groupID/eligible | JWT | Members who failed habits this week |
| POST | /penalties/spin | JWT | Spin roulette, assign punishment, trigger broadcast |
| GET | /penalties/:groupID | JWT | List pending debts for group (last 7 days) |
| PATCH | /penalties/:id/resolve | JWT | Mark debt as resolved |

### Health
| Method | Path | Auth | Description |
|---|---|---|---|
| GET | /health | None | Keepalive ping |

## Environment Variables
```
PORT=8083
DATABASE_URL=...                  # Supabase PostgreSQL connection string
REALTIME_SERVICE_URL=http://...   # To trigger broadcast after spin
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
│   └── habit.go          ← Habit, UserHabit, TodayCheckin, Streak,
│                            RouletteEntry, Debt, EligibleMember, Punishment
├── db/
│   └── queries.go        ← all pgxpool queries + calculateStreak + CurrentWeekStart
└── handler/
    ├── habit_handler.go  ← HabitHandler (habits, checkins, streaks)
    ├── penalty_handler.go← PenaltyHandler (roulette, debts)
    └── response.go       ← shared writeError / writeJSON
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

## Service-Specific Notes
- `POST /habits/assign` is a direct assignment. The proposals/voting system
  (`proposals`, `proposal_votes` tables) is not yet implemented in this service.
- `GET /penalties/:groupID` returns only debts from the last 7 days to
  avoid loading stale data (matches the `idx_debts_week` index).
- The `suggestion_id` column in `debts` is always NULL for now — user-submitted
  punishment suggestions are captured in `punishment_suggestions` but the spin
  currently draws from the JSON catalog only.
