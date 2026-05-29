# CLAUDE.md — habits-service

## Purpose
Owns the habits catalog, user habit assignments per group, daily check-ins,
streak calculation, and the Saturday roulette / penalty mechanic.
This is the core data service of the app.

## This service does NOT
- Send real-time events directly — it returns 200 and the caller (api-gateway
  or a background job) notifies realtime-service separately
- Handle group membership validation — it calls groups-service for that
- Validate JWT — api-gateway does that, trust X-User-ID header

## Endpoints

### Habits
| Method | Path | Auth | Description |
|---|---|---|---|
| GET | /habits | JWT | List available habits catalog |
| POST | /habits/assign | JWT | Assign habit to self in a group |
| POST | /checkins | JWT | Submit daily check-in |
| GET | /checkins/:groupID/today | JWT | Get today's check-ins for group |
| GET | /streaks/:userID | JWT | Get user's current streaks |

### Penalties (owned by this service — no separate penalties-service)
| Method | Path | Auth | Description |
|---|---|---|---|
| GET | /penalties/:groupID/eligible | JWT | Members who failed habits this week |
| POST | /penalties/spin | JWT | Spin roulette, assign punishment, trigger broadcast |
| GET | /penalties/:groupID | JWT | List pending debts for group |
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
`habits` · `user_habits` · `checkins` · `debts`

Other services must NOT write to these tables directly.

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

### Error handling convention
```go
func writeError(w http.ResponseWriter, status int, msg string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
```

### Database connection pool (pgx)
Use `pgxpool`. Pool is initialized once at startup, passed via dependency
injection, never as a global variable.

### Request context — userID
```go
userID := r.Header.Get("X-User-ID")
```

## Service-Specific Notes
Streak calculation runs on every check-in. It counts consecutive days with
at least one check-in. A streak breaks if no check-in was recorded the
previous calendar day (CST timezone).

The punishment catalog is a JSON file loaded at startup (`punishments.json`),
not stored in the DB. Randomness uses `crypto/rand`, not `math/rand`.

After a roulette spin, this service calls realtime-service's HTTP internal
endpoint to trigger the `roulette_result` and `debt_created` broadcast.
