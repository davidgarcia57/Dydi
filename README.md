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

---

## Compilar y Verificar Cambios Go

Go NO esta instalado localmente; corre dentro de Docker. Para verificar que un cambio compila:

```bash
docker compose build <servicio>   # ej. docker compose build habits-service
docker compose up -d <servicio>
docker compose logs -f <servicio>
```

---

## Despliegue en Producción (Render Free Tier)

El backend está diseñado para desplegarse en la capa gratuita de Render (un microservicio por cuenta). Para lidiar con el límite de inactividad de 15 minutos de Render, el proyecto cuenta con dos mecanismos:

1. **Despertar en Paralelo:** El endpoint `/health` del `api-gateway` hace peticiones asíncronas a los demás microservicios (`groups`, `habits`, `realtime`) para despertarlos simultáneamente sin agotar el timeout de Render.
2. **Cron Job (Keep-Awake):** 
   - Se recomienda configurar un job gratuito en [cron-job.org](https://cron-job.org) que haga un ping HTTP GET a `https://<tu-api-gateway>.onrender.com/health` cada **12 minutos** durante las horas de mayor tráfico.
   - Esto mantiene despiertos a **todos** los servicios automáticamente gracias a la arquitectura del Gateway.
   - *(Opcional)* Existe un workflow de respaldo en `.github/workflows/keep-awake.yml` que cumple esta misma función, pero consume minutos de GitHub Actions.

---

## Flujo De Trabajo En Equipo

1. Crea tu rama desde `main`: `git checkout -b feature/nombre-feature`
2. Trabaja unicamente dentro del directorio del servicio que te corresponde
3. Abre un Pull Request a `main`
4. GitHub Actions corre build automaticamente
5. Merge solo cuando el CI este en verde

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
|-- frontend/                 -> Vue 3 SPA
`-- supabase/
    `-- migrations/           -> schema de la base de datos (fuente de verdad)
```

---

## Variables De Entorno Por Servicio

Cada servicio tiene su propio `.env.example` con las variables necesarias para correrlo de forma independiente o configurar Render.

Nunca subas archivos `.env` al repositorio.
