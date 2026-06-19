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
  El gateway valida JWT **ES256** vía **JWKS** (no shared secret). Servicios
  internos confían en el header `X-User-ID` que pone el gateway; no revalidan.
- **Sin penalties-service**: las penitencias viven en `habits-service`.
- **Realtime propio** con `nhooyr.io/websocket` → migrado a
  `github.com/coder/websocket` (misma API). No usar Supabase Realtime — es la
  variable más importante del experimento.
- **Endpoints internos** (`/internal/*`) exigen header `X-Internal-Token` ==
  env `INTERNAL_TOKEN` (debe ser el MISMO en groups/habits/realtime).
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
| Móvil | `cd mobile && expo start` (no usa Docker) |
| Pruebas de estrés | k6 con `load-tests/k6_stress_test.js` |

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
