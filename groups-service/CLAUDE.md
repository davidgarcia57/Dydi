# CLAUDE.md — groups-service

## Purpose
Manages group creation, invite codes, membership, and group metadata.
Enforces the hard limit of 8 members per group.

## This service does NOT
- Handle check-ins or habit tracking (that is habits-service)
- Handle real-time events (that is realtime-service)
- Validate JWT (api-gateway does that — trust X-User-ID header)

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
DATABASE_URL=...             # Supabase PostgreSQL connection string
MAX_GROUP_SIZE=8
```

## Database Tables Owned
`groups` · `group_members`

`proposals` and `proposal_votes` are group-domain tables not yet implemented.
Other services must NOT write to `groups` or `group_members` directly.

## Internal Package Structure

```
internal/
├── model/
│   └── group.go         ← Group, Member, GroupWithMembers structs
├── db/
│   └── queries.go       ← all pgxpool queries + GenerateInviteCode
└── handler/
    └── group_handler.go ← GroupHandler, one method per endpoint
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
    port = "8082"
}
http.ListenAndServe(":"+port, r)
```

### Error response convention
```go
func writeError(w http.ResponseWriter, status int, msg string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
```

### Database connection pool (pgx)
Use `pgxpool`, not single connections. Pool is initialized once at startup
in `main()` and passed to `handler.NewGroupHandler(pool)`.

### Request context — userID
The API gateway attaches `X-User-ID` to every authenticated request.
Read it from the header — do not re-validate the JWT here.
```go
userID := r.Header.Get("X-User-ID")
```

### Invite code generation
`db.GenerateInviteCode()` returns an 8-character alphanumeric code using
`crypto/rand`. Characters exclude O, 0, I, 1 to avoid visual ambiguity.

## Service-Specific Notes
- MAX_GROUP_SIZE=8 is enforced in `JoinGroup` before any DB write.
  This is both a product constraint and a free-tier protection (limits
  concurrent WebSocket connections per room in realtime-service).
- `JoinGroup` returns 404 (not 403) on wrong invite code to avoid leaking
  group existence to non-members.
- Members list always returns `[]` (not `null`) for empty groups.
