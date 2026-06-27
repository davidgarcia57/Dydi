# Dydi

SaaS de accountability social donde grupos de amigos rastrean habitos diarios y gamifican las consecuencias. Proyecto academico UTD Integradora 2026.

---

## Que hace Dydi

Un usuario crea un grupo, invita a sus amigos (maximo 8), y entre todos proponen y votan que habitos quieren rastrear juntos (ej. "30 min de ejercicio", "leer 20 paginas"). Cada dia, cada miembro hace check-in de sus habitos. El sabado, quien haya fallado entra a la ruleta de penitencias: el grupo vota que castigo le toca y se sortea de forma aleatoria entre las opciones aprobadas. La deuda caduca automaticamente al final de la semana siguiente si no se cumple.

---

## Arquitectura

```text
frontend/           -> Vue 3 + Vite + Tailwind        (Vercel)
api-gateway/        -> Go 1.24 + chi v5               (Render - Cuenta 1)
groups-service/     -> Go 1.24 + chi v5               (Render - Cuenta 2)
habits-service/     -> Go 1.24 + chi v5               (Render - Cuenta 3)
realtime-service/   -> Go 1.24 + WebSocket            (Render - Cuenta 4)
mobile/             -> Expo + React Native            (APK via EAS)
```

Auth y base de datos viven en Supabase cloud. El unico punto de entrada publico es `api-gateway`.

---

## Requisitos

- [Docker Desktop](https://www.docker.com/products/docker-desktop/) con WSL2 habilitado
- Git
- Credenciales del proyecto Supabase cloud (pedir al lider del proyecto)

No necesitas Go ni Node instalados localmente; Docker los provee.

---

## Configuracion Inicial Local

### 1. Clonar el repositorio

```bash
git clone <url-del-repo>
cd dydi
```

### 2. Variables de entorno

Crea un archivo `.env` en la **raíz del proyecto** copiándolo del ejemplo:
```bash
cp .env.example .env
```

Llena las variables de la raíz con las credenciales del proyecto Supabase cloud:

| Variable | Donde encontrarla |
|---|---|
| `DATABASE_URL` | Supabase → Project Settings → Database → Connection string |
| `SUPABASE_JWKS_URL` | `https://<project-ref>.supabase.co/auth/v1/.well-known/jwks.json` |
| `VITE_SUPABASE_URL` | Supabase → Project Settings → API → Project URL |
| `VITE_SUPABASE_ANON_KEY` | Supabase → Project Settings → API → anon key |
| `INTERNAL_TOKEN` | Genéralo con `openssl rand -hex 32`. Protege las llamadas gateway-a-servicios y servicio-a-servicio |

> **Sobre `INTERNAL_TOKEN`:** es un secreto compartido que autentica las llamadas
> del `api-gateway` hacia los servicios y las llamadas internas entre servicios
> (`/internal/broadcast`, `/internal/proposals/apply`, etc.). Debe tener el
> **mismo valor exacto** en `api-gateway`, `groups-service`, `habits-service` y
> `realtime-service`: el gateway lo estampa en cada request proxeada y los
> servicios lo exigen antes de confiar en `X-User-ID`. Si los valores no
> coinciden, las llamadas se rechazan con 401. En local con Docker Compose basta
> ponerlo una vez en el `.env` de la raíz; en Render hay que configurarlo en las
> env vars de los 4 servicios.

> **Nota para despliegues o ejecución nativa:** Cada microservicio (`api-gateway`, `groups-service`, etc.) tiene su propio `.env.example`. Estos son necesarios si vas a correr los servicios con `go run main.go` o para configurar las variables en **Render**. Si solo vas a usar Docker Compose, el `.env` de la raíz es suficiente, ya que `docker-compose.yml` gestiona las rutas internas automáticamente.
> 
> Nunca subas archivos `.env` al repositorio.

### 3. Levantar todos los servicios

```bash
docker compose up --build
```

La primera vez tarda varios minutos descargando imagenes. Las siguientes son mucho mas rapidas.

### 4. Probar login y registro

```text
http://localhost:5173/#/login
```

---

## Puertos Locales

| Servicio | URL |
|---|---|
| Frontend | http://localhost:5173 |
| API Gateway | http://localhost:8080 |
| groups-service | http://localhost:8082 |
| habits-service | http://localhost:8083 |
| realtime-service | http://localhost:8084 |

La base de datos es Supabase cloud — no hay postgres local.

---

## Verificar Servicios

```bash
docker compose ps
```

Todos los servicios deben aparecer como `Up`. Tambien puedes probar los health checks:

```bash
curl http://localhost:8080/health
curl http://localhost:8082/health
curl http://localhost:8083/health
curl http://localhost:8084/health
```

Los 4 servicios exponen además `/metrics` (formato Prometheus) para medir latencia
P95, cold start de WebSocket y consistencia de entrega de eventos — útil para las
métricas del paper. Ej: `curl http://localhost:8080/metrics`.

---

## Verificar Cambios (`verify.sh`)

Go y Node no están instalados localmente — todo corre en Docker. `verify.sh`
ejecuta la **suite completa, la misma que el CI**: por cada servicio Go corre
`gofmt`, `go vet`, `go build`, `go test -race` y `golangci-lint`; en el frontend
`lint`, `format:check`, `build` y `test`; en el móvil `tsc --noEmit`.

```bash
./verify.sh                    # todo (4 Go + frontend + móvil)
./verify.sh go                 # solo los 4 servicios Go
./verify.sh go:habits-service  # un solo servicio Go
./verify.sh frontend           # solo el frontend
./verify.sh mobile             # solo el typecheck del móvil
```

> En Windows, córrelo desde WSL: `wsl -d ubuntu bash -lc './verify.sh'`
> (Git Bash mangea las rutas de los volúmenes Docker).

Déjalo **todo verde** antes de abrir un PR. Para iterar levantando un solo
servicio en vez de verificar:

```bash
docker compose build <servicio>   # ej. docker compose build habits-service
docker compose up -d <servicio>
docker compose logs -f <servicio>
```

---

## Atajos De Desarrollo (`scripts/`)

Para inspeccionar el sistema sin abrir el front (corre desde WSL; usan Docker/curl
y leen `.env`, no instalan nada):

| Atajo | Qué hace |
|---|---|
| `./scripts/q.sh "SELECT …"` | Query **read-only** a la BD de Supabase (la sesión va forzada a solo-lectura; un write falla). `-f archivo.sql` para correr un archivo. |
| `./scripts/hit.sh GET habits /habits` | Golpea un servicio backend **directo** (estampa `X-Internal-Token` + `X-User-ID`, sin JWT). Para el gateway: `DYDI_JWT=<token> ./scripts/hit.sh GET gateway /groups`. |
| `./scripts/logs.sh habits -f` | Tail de logs (slog JSON) de un servicio del compose. Alias: `gateway`/`groups`/`habits`/`realtime`. |

---

## Despliegue en Producción (Render Free Tier)

El backend se despliega en la capa gratuita de Render: **un microservicio por
cuenta**, los 4 apuntando a este mismo monorepo.

### Configuración de build de cada servicio en Render

Como el monorepo comparte el módulo `shared/`, los Dockerfiles de `groups`,
`habits` y `realtime` se construyen desde la **raíz** del repo:

| Campo (Render → Settings → Build) | Valor |
|---|---|
| **Root Directory** | *(vacío)* |
| **Docker Build Context Directory** | `.` |
| **Dockerfile Path** | `<servicio>/Dockerfile` — ej. `habits-service/Dockerfile` |

> ⚠️ **El `Dockerfile Path` debe coincidir EXACTO con el nombre de la carpeta**
> (`habits-service`, **con "s"**). Un typo como `habit-service/Dockerfile` no
> existe → Render falla el build **en silencio** y el servicio sigue sirviendo la
> última imagen que sí compiló (que puede ser de OTRO servicio). Síntoma clásico:
> el servicio responde `/health` 200 pero **todas** sus rutas de app dan **404**.
> Para saber qué binario corre realmente una URL:
> `curl https://<url>/<ruta-conocida>` debe dar **401** (la ruta existe y pide el
> token interno), no 404.

El `api-gateway` es la excepción: su Dockerfile no usa `shared/`, así que su build
context es `./api-gateway` con el `Dockerfile` por defecto. Además, en sus env
vars necesita las **URLs públicas** de los otros 3 servicios:

| Env var (api-gateway) | Apunta a |
|---|---|
| `GROUPS_SERVICE_URL` | URL pública de `groups-service` |
| `HABITS_SERVICE_URL` | URL pública de `habits-service` |
| `REALTIME_SERVICE_URL` | URL pública de `realtime-service` |
| `INTERNAL_TOKEN` | mismo valor que en los 3 servicios |

### Keep-Awake (límite de 15 min de inactividad)

Para lidiar con el límite de inactividad de 15 minutos de Render, el proyecto cuenta con dos mecanismos:

1. **Despertar en Paralelo:** El endpoint `/health` del `api-gateway` hace peticiones asíncronas a los demás microservicios (`groups`, `habits`, `realtime`) para despertarlos simultáneamente sin agotar el timeout de Render.
2. **Cron Job (Keep-Awake):** 
   - Se recomienda configurar un job gratuito en [cron-job.org](https://cron-job.org) que haga un ping HTTP GET a `https://<tu-api-gateway>.onrender.com/health` cada **12 minutos** durante las horas de mayor tráfico.
   - Esto mantiene despiertos a **todos** los servicios automáticamente gracias a la arquitectura del Gateway.
   - *(Opcional)* Existe un workflow de respaldo en `.github/workflows/keep-awake.yml` que cumple esta misma función, pero consume minutos de GitHub Actions.

> **Recuerda al desplegar:** configura `INTERNAL_TOKEN` con el **mismo valor** en
> las env vars de `api-gateway`, `groups-service`, `habits-service` y
> `realtime-service` en Render. Sin él (o con valores distintos), el gateway no
> arranca o los servicios rechazan las llamadas con 401.

---

## Flujo De Trabajo En Equipo

1. Crea tu rama desde `main`: `git checkout -b feature/nombre-feature`
2. Trabaja en el directorio del servicio que te corresponde
3. Corre `./verify.sh` localmente y déjalo **todo verde**
4. Abre un Pull Request a `main`
5. GitHub Actions corre la misma suite (`verify.sh`) en los jobs **Go**, **Frontend** y **Mobile**
6. Merge solo cuando **CI Success** (el job-gate que agrupa a todos) esté en verde

> El único check requerido en branch protection es **CI Success**; así renombrar
> o agregar jobs no rompe los PRs.

Cada servicio se despliega de forma independiente. No modifiques archivos fuera de tu directorio sin avisar al equipo.

---

## Estructura Del Proyecto

```text
dydi/
|-- .env.example              -> variables de entorno raiz (api-gateway)
|-- docker-compose.yml        -> orquestacion local
|-- api-gateway/              -> unico punto de entrada publico
|-- groups-service/           -> grupos, membresias y propuestas de habitos
|-- habits-service/           -> habitos, check-ins, rachas y penitencias
|-- realtime-service/         -> WebSocket hub para eventos en tiempo real
|-- frontend/                 -> Vue 3 SPA (web)
|-- mobile/                   -> app Expo / React Native (APK)
|-- scripts/                  -> atajos de dev (q.sh / hit.sh / logs.sh)
|-- verify.sh                 -> suite de verificacion en Docker (== CI)
|-- .github/                  -> CI (verify.sh), build del APK, keep-awake
`-- supabase/
    `-- migrations/           -> schema de la base de datos (fuente de verdad)
```

---

## Variables De Entorno Por Servicio

Cada servicio tiene su propio `.env.example` con las variables necesarias para correrlo de forma independiente o configurar Render.

Nunca subas archivos `.env` al repositorio.
