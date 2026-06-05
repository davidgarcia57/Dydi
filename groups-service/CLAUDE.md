# CLAUDE.md — groups-service

## Purpose
Manages group creation, invite codes, membership, and habit proposals.
Enforces the hard limit of 8 members per group.

A **proposal** is a democratic request to add or remove a shared habit for
the entire group. Any member can open a proposal; it passes when a simple
majority votes yes. When a proposal is approved, habits-service is notified
and assigns (or unassigns) the habit to all current group members.

## This service does NOT
- Handle check-ins or streak calculation (that is habits-service)
- Handle real-time events (that is realtime-service)
- Validate JWT — api-gateway does that; trust the `X-User-ID` header

## Endpoints

### Groups
| Method | Path | Auth | Description |
|---|---|---|---|
| POST | /groups | JWT | Create a new group |
| GET | /groups/:id | JWT | Get group details + members |
| POST | /groups/:id/join | JWT | Join via invite code |
| GET | /groups/:id/members | JWT | List members |
| DELETE | /groups/:id/leave | JWT | Leave group |

### Proposals
| Method | Path | Auth | Description |
|---|---|---|---|
| POST | /groups/:id/proposals | JWT | Open a new proposal (add_habit or remove_habit) |
| GET | /groups/:id/proposals | JWT | List open and recent proposals |
| POST | /proposals/:id/vote | JWT | Cast a vote (yes/no) |

### Health
| Method | Path | Auth | Description |
|---|---|---|---|
| GET | /health | None | Keepalive ping |

## Environment Variables
```
PORT=8082
DATABASE_URL=...             # Supabase PostgreSQL connection string
MAX_GROUP_SIZE=8
HABITS_SERVICE_URL=http://...  # notified when a proposal is approved
```

## Database Tables Owned
`groups` · `group_members` · `proposals` · `proposal_votes`

Other services must NOT write to these tables directly.
Cross-service reads are acceptable.

## Proposal Flow (business logic)

```
Member opens proposal (type: add_habit | remove_habit, habit_id)
  → status: open
  → other members vote yes/no
  → when votes from all members are cast OR majority threshold reached:
      → status changes to approved or rejected
  → on approved: POST to habits-service /internal/proposals/apply
      → habits-service assigns/unassigns the habit for all group members
```

**Quorum rule:** simple majority (> 50% of current members). The proposer's
vote counts as an implicit yes. A proposal with no votes after 7 days expires.

## Internal Package Structure

```
internal/
├── model/
│   └── group.go         <- Group, Member, GroupWithMembers, Proposal, Vote structs
├── db/
│   └── queries.go       <- all pgxpool queries + GenerateInviteCode
└── handler/
    ├── group_handler.go    <- GroupHandler (groups, membership)
    └── proposal_handler.go <- ProposalHandler (proposals, votes)
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
in `main()` and passed to handlers.

### Request context — userID
```go
userID := r.Header.Get("X-User-ID")
```
Read from header — do not re-validate the JWT here.

### Invite code generation
`db.GenerateInviteCode()` returns an 8-character alphanumeric code using
`crypto/rand`. Characters exclude O, 0, I, 1 to avoid visual ambiguity.

## Service-Specific Notes
- `MAX_GROUP_SIZE=8` is enforced in `JoinGroup` before any DB write.
  This is both a product constraint and a Render free-tier protection
  (limits concurrent WebSocket connections per room).
- `JoinGroup` returns 404 (not 403) on wrong invite code to avoid leaking
  group existence to non-members.
- Members list always returns `[]` (not `null`) for empty groups.
- The call to habits-service on proposal approval runs in a goroutine.
  A habits-service outage must not block or fail the vote response.
