# CLAUDE.md — api-gateway

## Purpose
Single entry point for all client requests. Validates JWT tokens, routes to
internal services, and handles CORS and rate limiting. This is the only
service exposed to the public internet (besides the frontend on Vercel).

## This service does NOT
- Contain business logic
- Write to the database directly
- Handle WebSocket connections (those upgrade and proxy to realtime-service)

## Routing Table

| Path prefix | Routes to | Notes |
|---|---|---|
| `/api/groups/*` | groups-service:8082 | JWT required |
| `/api/habits/*` | habits-service:8083 | JWT required |
| `/api/penalties/*` | habits-service:8083 | JWT required — penalty logic lives in habits-service |
| `/health` | local | Returns 200, used by cron-job.org ping |
| `/ws/*` | realtime-service:8084 | WebSocket upgrade, JWT validated before proxying |

**There is no `/api/auth/*` route.** Auth (login, register, logout, token refresh) is
handled directly by the frontend via the Supabase JS SDK. The gateway never sees
auth traffic; it only validates the resulting JWT on every other request.

## JWT Validation

- Tokens are issued by Supabase Auth using **ES256 (P-256)** signing
- The gateway validates against Supabase's JWKS endpoint — **no shared secret**
- JWKS URL: `SUPABASE_JWKS_URL` env var (see `.env.example`)
- After validation, attaches `X-User-ID` header and forwards the request
- Downstream services trust the gateway and do NOT re-validate the JWT

## Environment Variables

```
PORT=8080                          # Render injects this automatically
GROUPS_SERVICE_URL=http://...
HABITS_SERVICE_URL=http://...      # also handles /api/penalties/* routes
REALTIME_SERVICE_URL=http://...
SUPABASE_JWKS_URL=https://<project-ref>.supabase.co/auth/v1/.well-known/jwks.json
ALLOWED_ORIGINS=https://dydi.vercel.app,http://localhost:5173
```

## Key Patterns

### Health endpoint — do not remove or simplify
```go
r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("ok"))
})
```
This endpoint is pinged every 10 minutes by cron-job.org to prevent
Render free-tier sleep. It must be fast and never require auth.

### Port binding — always use $PORT
```go
port := os.Getenv("PORT")
if port == "" {
    port = "8080"
}
http.ListenAndServe(":"+port, r)
```
Render injects `$PORT` at runtime. Never hardcode the port.

### Startup retry — do not remove
On startup, the gateway retries connecting to downstream services
with exponential backoff. This handles the case where services are
waking up from Render cold start simultaneously (can take 30s+).

## Dockerfile Notes
Multistage build: builder (golang:1.24-alpine) → final (alpine:latest).
Final image must stay under 20MB for fast Render cold starts.
Do not add unnecessary dependencies to the final stage.

## CORS
Only allow origins listed in `ALLOWED_ORIGINS`. Do not use wildcard `*`
in production. The frontend URL is `https://dydi.vercel.app`.
