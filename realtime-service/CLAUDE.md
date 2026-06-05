# CLAUDE.md — realtime-service

## Purpose
WebSocket hub. Maintains persistent connections for all active group members
and broadcasts real-time events: check-ins, streak updates, roulette spins,
and debt assignments.

## This is the most infrastructure-sensitive service in the project.
Read this file fully before making any changes.

## Architecture: Hub Pattern

```
Client A --ws--+
Client B --ws--+-- Hub (in-memory) -- broadcasts to room
Client C --ws--+
```

The Hub is a single in-memory struct that manages all active connections
grouped by `groupID`. It runs in its own goroutine.

```
internal/
+-- hub/
|   +-- hub.go        <- Hub struct, Register/Unregister/Broadcast
|   +-- client.go     <- Client struct, read/write pumps
+-- handler/
    +-- ws_handler.go <- HTTP upgrade -> WebSocket handshake
```

## Event Types (do not add new types without updating frontend)

```go
const (
    EventCheckin        = "checkin"         // member completed a habit
    EventStreakUpdate   = "streak_update"   // streak count changed
    EventMemberOnline  = "member_online"   // member connected
    EventMemberOffline = "member_offline"  // member disconnected
    EventRouletteStart = "roulette_start"  // habits-service triggered spin
    EventRouletteResult = "roulette_result" // spin result to broadcast
    EventDebtCreated   = "debt_created"    // new debt assigned
)
```

## Event Payload (JSON)
```json
{
  "type": "checkin",
  "groupID": "uuid",
  "userID": "uuid",
  "payload": {}
}
```

## Internal Endpoint (called by habits-service only)

```
POST /internal/broadcast
Body: { "type": "...", "groupID": "...", "userID": "...", "payload": {} }
```

This endpoint is not exposed via api-gateway. Only internal services call it.

## Environment Variables
```
PORT=8084
SUPABASE_JWKS_URL=https://<project-ref>.supabase.co/auth/v1/.well-known/jwks.json
MAX_CONNECTIONS_PER_GROUP=8
PING_INTERVAL_SECONDS=30
PONG_WAIT_SECONDS=60
WRITE_WAIT_SECONDS=10
```

## Free-Tier Critical Patterns

### Connection limits
`MAX_CONNECTIONS_PER_GROUP` is set to 8 (max group size). This is a hard
limit enforced in `hub.go`. Do not raise this without testing on Render
free tier first.

### Ping/Pong keepalive — do not remove
WebSocket connections on Render free tier get dropped after ~55 seconds
of inactivity. The client pump sends a ping every 30 seconds to prevent this.
```go
// In client.go writePump — do not remove this block
case <-ticker.C:
    conn.SetWriteDeadline(time.Now().Add(writeWait))
    if err := conn.Ping(ctx); err != nil {
        return
    }
```

### Health endpoint — do not simplify
```go
r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    // Also return active connection count for metrics
    json.NewEncoder(w).Encode(map[string]int{
        "active_connections": hub.ConnectionCount(),
    })
})
```
This is pinged by cron-job.org AND used by Prometheus for metrics.
Never simplify to a plain 200 response.

### Graceful shutdown
On SIGTERM (Render sends this before stopping), the hub closes all
connections cleanly and broadcasts `member_offline` for each. This
prevents ghost connections on reconnect.

## Supabase Realtime as Fallback
If this service becomes unstable on Render free tier, the fallback plan
is to replace it with Supabase Realtime (already in our Supabase project).
The event types and payload structure above are designed to be compatible
with Supabase Realtime channel format to make migration straightforward.
Do not change the event structure without considering this fallback.

## Metrics Instrumentation — do not remove
The following are measured for the academic paper:
- `realtime_connections_total` — Prometheus counter
- `realtime_cold_start_ms` — logged on first connection after sleep
- `realtime_events_emitted_total` — Prometheus counter
- `realtime_events_delivered_total` — Prometheus counter

Delivery consistency = delivered / emitted. This is a key research metric.

## Dockerfile Notes
This service holds long-lived connections. The final Docker image must use
`CMD` not `ENTRYPOINT` to allow Render's graceful shutdown signal to reach
the Go process directly (not wrapped in a shell).

```dockerfile
# Correct
CMD ["/app/realtime-service"]

# Wrong — shell wrapping blocks SIGTERM
CMD ["sh", "-c", "/app/realtime-service"]
```
