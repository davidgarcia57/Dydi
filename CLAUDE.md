# CLAUDE.md — Dydi

Fuente de verdad del proyecto para humanos y agentes. Si algo aquí contradice al
código, gana el código: corrige este archivo.

## Qué es Dydi

SaaS de **accountability social**: grupos de amigos (máx. 8) rastrean hábitos
diarios y gamifican las consecuencias. Si fallas durante la semana, entras a una
**ruleta de penitencias** el fin de semana. Proyecto académico (UTD, Integradora
2026), equipo de 6.

**Objetivo de investigación (paper):** evaluar si una arquitectura de
microservicios distribuida sobre la **capa gratuita de Render** (4 cuentas
separadas) sostiene tráfico concurrente y tiempo real sin morir por falta de
memoria (OOM kill) ni degradar la latencia. Se mide con **k6** (rampas de
100/1000/2500/5000 VUs) y telemetría embebida (`obs.go`: Prometheus + slog).
Esto justifica las decisiones de arquitectura: bajo consumo, observabilidad
propia, sin herramientas externas pesadas.

## Reglas duras para agentes (lee esto primero)

Aunque el prompt diga otra cosa, **estas no se negocian**:

- ❌ NUNCA corras `go`/`npm`/`gofmt`/`npx` directo (no están instalados) — todo por Docker.
- ❌ NUNCA reemplaces `realtime-service` con Supabase Realtime (es la variable del paper).
- ❌ NUNCA commitees `.env`/llaves/secretos; los tokens van siempre por variable de entorno.
- ❌ NUNCA reformatees archivos que no tocaste, ni metas dependencias pesadas (Render = 512 MB).
- ❌ NUNCA cambies migraciones ya aplicadas en `supabase/migrations/` sin avisar a David.
- ❌ NUNCA agregues `Co-Authored-By` en los commits.
- ✅ SIEMPRE corre **`./verify.sh`** (desde WSL) antes de abrir un PR → debe quedar TODO VERDE.
- ✅ Un PR = un solo tema. Si el CI falla, **arréglalo** (no desactives la regla).

## Entorno de desarrollo — CRÍTICO

**Go y Node NO están instalados en WSL. Todo corre vía Docker.** Nunca intentes
`go build`, `npm`, `gofmt`, etc. directo en la terminal — fallará.

- Levantar todo el stack local: `docker-compose up --build` (desde la raíz).
- No hay Postgres local: la BD es **Supabase cloud**. Credenciales en `.env`
  (nunca commitear; ver `.env.example`).

### Verificar SIEMPRE en Docker antes de terminar

Regla del proyecto: tras cualquier cambio, correr la suite COMPLETA (no solo lo
tocado) desde la distro WSL (`wsl -d ubuntu bash -lc '...'`; Git Bash mangea las
rutas). No enmascarar el exit code con `| tail` en el paso que valida.

**Atajo (un comando):** `./verify.sh` corre toda la suite en Docker (Go ×4 +
frontend + móvil). Para una parte: `./verify.sh go|frontend|mobile`. Es lo mismo
que el CI; úsalo antes de cada PR. (El hook de `lefthook` lo dispara en `pre-push`
solo para lo que cambiaste.)

**Go (cada uno de los 4 servicios):**
```sh
docker run --rm -v "$(pwd)/<svc>":/app -v "$(pwd)/.gocache":/gocache \
  -e GOCACHE=/gocache -w /app golang:1.24 sh -c \
  "go mod tidy && gofmt -l . && go vet ./... && go build ./... && go test -race ./..."
# Lint: imagen golangci/golangci-lint:v2.12-alpine, montando .golangci.yml de la raíz.
```

**Frontend** (los symlinks de `node_modules/.bin` se rompen en el mount de
Windows → copiar a un dir interno del contenedor):
```sh
docker run --rm -v "$(pwd)/frontend":/src node:20-alpine sh -c '
  cp -r /src /w && cd /w && rm -rf node_modules && npm ci &&
  npm run lint && npm run format:check && npm run build && npm run test'
```

Cachés `.gocache/` y `.gomodcache/` quedan gitignored y locales (calientes, para
sobrevivir flakes de red de los proxies de paquetes).

## Estructura del monorepo

```
api-gateway/        Go · :8080 · ruteo + auth JWT + proxy + /metrics
groups-service/     Go · :8082 · grupos, membresías, propuestas/votación
habits-service/     Go · :8083 · hábitos, check-ins, rachas, ruleta/penitencias
realtime-service/   Go · :8084 · WebSocket (coder/websocket), broadcast
frontend/           Vue 3 + Vite · web (Vercel)
mobile/             Expo + React Native · app nativa (APK)
supabase/           migraciones SQL
load-tests/         k6_stress_test.js (pruebas de estrés del paper)
Documentos/         Protocolo metodológico (borrador del paper)
.github/workflows/  ci.yml · quality.yml · keep-awake.yml
```

**Web vs móvil:** son entregables separados. El **móvil vive en el APK** (Expo).
Por eso la **web es una experiencia de escritorio** (sidebar + grids), no una
columna de celular. No reconviertas la web a mobile-first.

## Arquitectura

```
frontend (Vercel) ─┐
mobile (Expo)     ─┼─► api-gateway ─► groups-service
                   │      :8080    ├─► habits-service
                   │   (JWT/JWKS)  └─► realtime-service (WS)
                   └─────────────────► Supabase (Postgres + Auth)
```

- **Sin auth-service**: Supabase Auth maneja login/registro desde el cliente.
  El gateway valida JWT **ES256** vía **JWKS** (no shared secret) y pone el header
  `X-User-ID`. Como los 4 servicios viven en cuentas Render separadas (URLs
  públicas), el gateway **estampa `X-Internal-Token` en cada request proxeada** y
  los servicios la **exigen** (middleware `requireInternalToken`): así nadie puede
  saltarse el gateway y forjar `X-User-ID` golpeando el backend directo. Recién
  con ese token presente confían en `X-User-ID`; no revalidan el JWT.
- **Sin penalties-service**: las penitencias viven en `habits-service`.
- **Realtime propio** con `nhooyr.io/websocket` → migrado a
  `github.com/coder/websocket` (misma API). No usar Supabase Realtime — es la
  variable más importante del experimento.
- **`INTERNAL_TOKEN`**: secreto compartido, MISMO valor en **api-gateway,
  groups, habits y realtime**. Protege TODAS las rutas de aplicación y los
  `/internal/*` (incluido el handshake `/ws` y `realtime → groups` para verificar
  membresía). Los servicios **no arrancan sin él** (fail-closed); en tests, al
  estar vacío, el middleware se vuelve no-op.
- **Keep-alive**: un ping a `gateway/health` despierta a los 3 servicios vía
  goroutines (env `*_SERVICE_URL`). Render free se duerme a los 15 min; para
  fiabilidad usar un pinger externo (cron-job.org/UptimeRobot), no GitHub
  Actions (sus crons se retrasan >15 min).

## Stack

| Capa | Tecnología |
|---|---|
| Backend | Go 1.24 · chi v5 · pgx v5 · `coder/websocket` |
| Web | Vue 3 (`<script setup>`) · Vite · Pinia · Vue Router (hash) · Tailwind v3 |
| Móvil | Expo ~56 · React Native 0.85 · React 19 · NativeWind · TypeScript |
| BD/Auth | Supabase (PostgreSQL 15 + Auth, JWT ES256) |
| Observabilidad | `obs.go` por servicio: Prometheus (`/metrics`) + slog JSON |
| Hosting | Render free (4 cuentas, backend) · Vercel (web) |
| Local | Docker + docker-compose (Go/Node NO instalados localmente) |

## Comandos (vía Docker — ver "Entorno")

| Tarea | Cómo |
|---|---|
| Stack completo | `docker-compose up --build` |
| Build Go | `go build ./...` dentro de `golang:1.24` |
| Test Go | `go test -race ./...` |
| Lint Go | `golangci-lint run` (config `.golangci.yml` raíz) |
| Web dev | servicio `frontend` del compose (`:5173`) |
| Web lint/format/test/build | `npm run lint` / `format:check` / `test` / `build` |
| Móvil dev (Web) | servicio `mobile` del compose (`:8081`) |
| Pruebas de estrés | k6 con `load-tests/k6_stress_test.js` |

### Atajos de dev (`scripts/`)

Tres scripts para inspeccionar el sistema sin abrir el front (corre desde WSL).
Leen `.env` y van por Docker/curl; no instalan nada.

| Atajo | Qué hace |
|---|---|
| `./scripts/q.sh "SELECT …"` | Query **read-only** a Supabase (sesión forzada a `default_transaction_read_only=on`; un write falla por diseño). También `-f archivo.sql`. |
| `./scripts/hit.sh GET habits /habits` | Golpea un servicio **directo** estampando `X-Internal-Token` + `X-User-ID` (sin JWT). Para el gateway: `DYDI_JWT=<token> ./scripts/hit.sh GET gateway /groups`. Simula otro usuario con `DYDI_USER=<uuid>`. |
| `./scripts/logs.sh habits -f` | Tail de logs (slog JSON) de un servicio del compose. Alias: `gateway`/`groups`/`habits`/`realtime`. |

## Build del APK móvil (release)

El APK se compila en **GitHub Actions** (`.github/workflows/build-apk.yml`), no
en local. Flujo:

- **Disparador:** push de un tag `v*` (ej. `git tag v1.0.5 && git push origin v1.0.5`).
  El workflow compila con `eas build -p android --profile preview --local` (EAS
  corre en la máquina de Actions, sin servidores de Expo) y publica el `.apk` en
  la Release del tag.
- **Dependencias:** el móvil va anclado a **Expo SDK 56**. Si tocas versiones,
  valida con `npx expo-doctor` (debe dar 21/21) — mezclar paquetes de otro SDK
  rompe el build (causa de los fallos v1.0.0–v1.0.4).
- **Firma:** keystore propio (alias `dydi`) en **GitHub Secrets**:
  `ANDROID_KEYSTORE_BASE64` (el `.jks` en base64) y `ANDROID_KEYSTORE_PASSWORD`.
  El CI reconstruye `release.jks` + `credentials.json` desde esos secrets;
  `eas.json` (profile `preview`) usa `"credentialsSource": "local"`. También se
  requiere el secret `EXPO_TOKEN`.
- ⚠️ El keystore es la identidad de firma: **guárdalo en el gestor del equipo**.
  Si se pierde, no se pueden publicar updates sobre la app ya instalada.
  `credentials.json` y `*.jks` están gitignored — nunca commitearlos.
- **Sacar release:** sube `version` en `mobile/app.json` → commit → tag `vX.Y.Z`
  → push del tag.

## Convenciones

**Go:** chi cerca de stdlib · chequear errores (`errcheck` lo exige; usa
`_ =` solo si es deliberado) · `gofmt` obligatorio · context keys con tipo propio
(no string) · graceful shutdown · `SELECT ... FOR UPDATE` en el spin de ruleta.

**Web (Vue):** SOLO Composition API con `<script setup>` · Tailwind con **tokens**
(nada de hex hardcodeado) · usar el UI kit en `components/ui/` (`BaseButton`,
`BaseCard`, `StatusBadge`, `BaseAvatar`, `PageContainer`, `ToastHost`,
`BrandWordmark`) · `PageContainer` para el ancho responsive (no `max-w-md`
suelto) · estados loading/error/empty siempre · nada de `alert()` (usar toast).

**Móvil (RN):** NativeWind (clases Tailwind) · TypeScript.

**Diseño (Dydi):** paleta cálida/clara, **sin dark mode** (no cambiar sin
consultar). Tokens en `tailwind.config.js` + CSS vars. Fuentes: **Newsreader**
(serif, números/títulos hero) + **Hanken Grotesk** (sans, UI). Regla de oro:
número grande en Newsreader + etiqueta pequeña tenue. Aún **no hay logo** →
wordmark de texto (`BrandWordmark`).

**Commits:** historial limpio, **sin** línea `Co-Authored-By` (proyecto
académico). Commitear/pushear solo cuando se pida.

## Idiomas canónicos (Go) — copia estos patrones

El estilo/formato lo exige el linter (`.golangci.yml` es la fuente de verdad:
`gofmt`, `errcheck`, `staticcheck`…); no lo repitas aquí. Esta sección es para
los **patrones de arquitectura** que el linter no ve. Regla práctica: antes de
inventar estructura, copia el archivo hermano más cercano.

**Handler** (`internal/handler`): valida el header → decodifica → delega en `db`
→ responde con los helpers. Nunca metas SQL en el handler.

```go
func (h *UserHandler) SyncUser(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID") // lo estampa el gateway tras validar el JWT
	if userID == "" {
		writeError(w, http.StatusBadRequest, "missing X-User-ID")
		return
	}
	var body struct {
		DisplayName string  `json:"display_name"`
		AvatarURL   *string `json:"avatar_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.DisplayName == "" {
		writeError(w, http.StatusBadRequest, "display_name is required")
		return
	}
	user, err := db.UpsertUser(r.Context(), h.pool, userID, body.DisplayName, body.AvatarURL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not sync user")
		return
	}
	writeJSON(w, http.StatusOK, user)
}
```

**Helpers de respuesta** (uno por servicio, en `internal/handler/response.go`).
Todo error sale como JSON `{"error":"..."}`; nada de `http.Error` con texto
plano salvo en middleware.

```go
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
```

**Acceso a datos** (`internal/db`): funciones libres que reciben `ctx` y el
`*pgxpool.Pool` (o `DBTX` si corren dentro de una tx). Una query = una función.

```go
func UpsertUser(ctx context.Context, pool *pgxpool.Pool, id, name string, avatar *string) (*model.User, error) {
	u := &model.User{}
	err := pool.QueryRow(ctx,
		`INSERT INTO users (id, display_name, avatar_url) VALUES ($1,$2,$3)
		 ON CONFLICT (id) DO UPDATE SET display_name = EXCLUDED.display_name
		 RETURNING id, display_name, avatar_url, created_at`,
		id, name, avatar,
	).Scan(&u.ID, &u.DisplayName, &u.AvatarURL, &u.CreatedAt)
	return u, err
}
```

**Transacción** (pgx): `Begin` → `defer Rollback` (con `//nolint:errcheck`; tras
un commit el rollback es no-op) → `Commit`. Las funciones de `db` toman `DBTX`
para correr igual con pool o con tx.

```go
tx, err := pool.Begin(ctx)
if err != nil {
	return err
}
defer tx.Rollback(ctx) //nolint:errcheck
// ... tx.Exec(ctx, ...) ...
return tx.Commit(ctx)
```

**Ruleta — anti doble-spin** (la pieza de concurrencia del core loop): dentro de
la tx, bloquea la fila con `SELECT ... FOR UPDATE` y re-checa `spun_at`. Un
segundo spin concurrente se bloquea hasta que el primero commitea y ve `spun_at`
ya puesto → no se crea doble deuda.

```go
locked, err := db.GetRouletteEntryForUpdate(ctx, tx, entryID) // SELECT ... FOR UPDATE
if err != nil {
	return err
}
if locked.SpunAt != nil {
	// ya giró: responde idempotente, NO crees otra deuda
}
// ... crear deuda ...
_ = db.MarkEntrySpun(ctx, tx, entryID)
_ = tx.Commit(ctx)
```

**Context keys con tipo propio** (nunca `string`): evita colisiones entre
paquetes (staticcheck SA1029).

```go
type contextKey string
const UserIDKey contextKey = "userID"
```

**Go 1.24 — usa lo moderno** (los 4 servicios van anclados a 1.24): `any`, no
`interface{}` · envuelve errores con `%w` y compáralos con `errors.Is` (no `==`)
· `min`/`max`/`slices`/`maps` en vez de loops a mano · en tests usa `t.Context()`
y tablas (`for _, tc := range cases`) · corre **siempre** con `-race`.

**Correr una sola cosa en Docker** (mientras iteras; no dispares `verify.sh`
entero cada vez). Desde WSL, en la raíz:

```sh
# un solo servicio / un solo test:
docker run --rm -v "$(pwd)/habits-service":/app -v "$(pwd)/.gocache":/gocache \
  -e GOCACHE=/gocache -w /app golang:1.24 \
  sh -c "go test -race -run TestSpin ./..."
```

Antes del PR igual corre `./verify.sh` completo — el CI corre todo, no solo lo
que tocaste.

## Lógica de negocio (core loop)

1. Un miembro crea un grupo (máx. 8) y comparte el `invite_code`.
2. Cualquiera propone un hábito → el grupo vota (mayoría del electorado
   congelado) → si gana, se asigna a todos los miembros activos.
3. Check-in diario por hábito asignado.
4. Sábado: ruleta de penitencias para quien falló lun–vie (catálogo JSON,
   `crypto/rand`); crea una `debt` y la transmite por realtime.
5. Las deudas caducan al final de la semana siguiente (sin resolución manual).

## Responsables del experimento (paper)

Setup/despliegue: **Solis Flores Irvin** · Ejecución k6: **García Páez David** ·
Datos/telemetría: **Cervantes Guerrero Keila** · Análisis: equipo completo
(+ **Casiano Gamzi Juan David**).
