# Producción: URLs y variables

Este archivo documenta la configuración operativa actual. No contiene secretos.

## URLs confirmadas

| Superficie | URL | Uso |
|---|---|---|
| Frontend | https://dydi-xi.vercel.app | URL pública estable |
| API Gateway | https://api-gateway-j3yi.onrender.com | Único backend público para web y móvil |
| Groups | https://groups-service-mev3.onrender.com | Destino del gateway y de Realtime |
| Habits | https://dydi-jeru.onrender.com | Destino del gateway y de Groups |
| Realtime | https://realtime-service-4mg6.onrender.com | Destino WebSocket y broadcasts internos |
| Supabase | https://iuorcsmjgyljyzebfhbc.supabase.co | Auth y PostgreSQL |

Los cuatro servicios viven en cuentas de Render separadas. Por eso no existe una
red privada compartida entre ellos: las llamadas “internas” usan estas URLs HTTPS
públicas y se protegen con el mismo `INTERNAL_TOKEN`.

## Distribución de variables

### Vercel — frontend

- `VITE_API_URL=https://api-gateway-j3yi.onrender.com`
- `VITE_WS_URL=wss://api-gateway-j3yi.onrender.com`
- `VITE_SUPABASE_URL=https://iuorcsmjgyljyzebfhbc.supabase.co`
- `VITE_SUPABASE_ANON_KEY=<publishable key de .env.supabase-new>`

La clave publishable/anon es pública por diseño, pero se administra como variable
de Vercel y nunca se copia a documentación, commits o mensajes.

### Render — api-gateway-j3yi

- `GROUPS_SERVICE_URL=https://groups-service-mev3.onrender.com`
- `HABITS_SERVICE_URL=https://dydi-jeru.onrender.com`
- `REALTIME_SERVICE_URL=https://realtime-service-4mg6.onrender.com`
- `SUPABASE_JWKS_URL=https://iuorcsmjgyljyzebfhbc.supabase.co/auth/v1/.well-known/jwks.json`
- `ALLOWED_ORIGINS=https://dydi-xi.vercel.app` — gobierna CORS **y** el allowlist
  de `Origin` del handshake WebSocket. El móvil no necesita ir en esta lista: su
  `Origin` es el de este mismo gateway y pasa por same-origin.
- `INTERNAL_TOKEN=<mismo secreto en los cuatro servicios>`
- `WAKE_TOKEN=<solo gateway y cron externo>`
- `RATE_LIMIT_RPS=5`
- `RATE_LIMIT_BURST=20`

### Render — groups-service-mev3

- `DATABASE_URL=<Supavisor transaction pooler, puerto 6543>`
- `DB_MAX_CONNS=10`
- `MAX_GROUP_SIZE=8`
- `HABITS_SERVICE_URL=https://dydi-jeru.onrender.com`
- `INTERNAL_TOKEN=<secreto compartido>`

### Render — dydi-jeru (Habits)

- `DATABASE_URL=<Supavisor transaction pooler, puerto 6543>`
- `DB_MAX_CONNS=10`
- `REALTIME_SERVICE_URL=https://realtime-service-4mg6.onrender.com`
- `PUNISHMENT_CATALOG_PATH=./punishments.json`
- `INTERNAL_TOKEN=<secreto compartido>`

### Render — realtime-service-4mg6

- `GROUPS_SERVICE_URL=https://groups-service-mev3.onrender.com`
- `INTERNAL_TOKEN=<secreto compartido>`
- `MAX_CONNECTIONS_PER_GROUP=8`
- `PING_INTERVAL_SECONDS=30`
- `WRITE_WAIT_SECONDS=10`

Realtime no recibe variables ni credenciales de Supabase. Tampoco
`ALLOWED_ORIGINS`: el allowlist de `Origin` del WebSocket lo aplica el gateway,
el único punto donde el `Host` sigue siendo el público real. Si la variable quedó
puesta en Render de antes, ya es inerte y se puede borrar.

## Cold starts

El mecanismo estable es un cron externo cada 12 minutos que hace
`POST https://api-gateway-j3yi.onrender.com/ops/wake` con
`X-Wake-Token`. El gateway despierta Groups, Habits y Realtime en paralelo.
`WAKE_TOKEN` nunca debe llegar al frontend, a una URL o a logs.

## Orden de despliegue

1. Groups, Habits y Realtime.
2. Gateway, después de confirmar los tres `/health`.
3. Vercel, después de confirmar gateway y JWKS.
4. Flujo funcional pequeño; no pruebas de carga.
