# Dydi

SaaS de accountability social donde grupos de amigos rastrean habitos diarios y gamifican las consecuencias. Proyecto academico - UTD Integradora 2025.

---

## Arquitectura

```text
frontend/           -> Vue 3 + Vite + Tailwind        (Vercel)
api-gateway/        -> Go 1.22 + chi v5               (Render - Cuenta 1)
groups-service/     -> Go 1.22 + chi v5               (Render - Cuenta 2)
habits-service/     -> Go 1.22 + chi v5               (Render - Cuenta 3)
realtime-service/   -> Go 1.22 + WebSocket            (Render - Cuenta 4)
```

Auth vive en Supabase Auth. En local se usa Supabase CLI para Auth y PostgreSQL de `docker-compose` para los datos de la app. El unico punto de entrada publico es `api-gateway`.

---

## Requisitos

- [Docker Desktop](https://www.docker.com/products/docker-desktop/) con WSL2 habilitado
- [Supabase CLI](https://supabase.com/docs/guides/local-development/cli/getting-started/)
- Git

No necesitas Go ni Node instalados localmente; Docker los provee.

---

## Configuracion Inicial Local

### 1. Clonar el repositorio

```bash
git clone <url-del-repo>
cd dydi
```

### 2. Levantar Supabase Auth local

No necesitas crear cuenta de Supabase para desarrollo local. El CLI levanta Auth, API, Studio e Inbucket usando Docker.

```bash
supabase init
supabase start
supabase status
```

Del output de `supabase status`, copia:

| Valor del CLI | Variable en `.env` |
|---|---|
| `API URL` | `VITE_SUPABASE_URL` |
| `anon key` | `VITE_SUPABASE_ANON_KEY` |
| `JWT secret` | `SUPABASE_JWT_SECRET` |

En local normalmente `VITE_SUPABASE_URL` queda como:

```env
VITE_SUPABASE_URL=http://127.0.0.1:54321
```

Importante: el `anon key` y el `JWT secret` deben salir del mismo `supabase status`. Si no coinciden, el frontend podra iniciar sesion, pero `api-gateway` rechazara el token.

### 3. Crear el archivo de variables de entorno

```bash
cp .env.example .env
```

Abre `.env` y llena las tres variables con los valores locales que imprimio `supabase status`.

> El resto de las variables (puertos, DB local, URLs internas) ya estan configuradas en `docker-compose.yml` y no requieren ajuste.

### 4. Levantar todos los servicios

```bash
docker compose up --build
```

La primera vez tarda varios minutos descargando imagenes. Las siguientes son mucho mas rapidas.

### 5. Probar login y registro

Abre:

```text
http://localhost:5173/#/login
```

Si Supabase local requiere confirmacion de correo, revisa el inbox local que aparece en `supabase status` como Inbucket.

---

## Puertos Locales

| Servicio | URL |
|---|---|
| Frontend | http://localhost:5173 |
| API Gateway | http://localhost:8080 |
| groups-service | http://localhost:8082 |
| habits-service | http://localhost:8083 |
| realtime-service | http://localhost:8084 |
| PostgreSQL app local | localhost:5432 |
| Supabase API local | http://127.0.0.1:54321 |
| Supabase Studio local | ver `supabase status` |
| Inbucket local | ver `supabase status` |

En local, los datos de la app usan PostgreSQL de `docker-compose`. Supabase local se usa para Auth.

---

## Verificar Servicios

```bash
docker compose ps
```

Todos los servicios deben aparecer como `Up`. Tambien puedes probar los health checks manualmente:

```bash
curl http://localhost:8080/health
curl http://localhost:8082/health
curl http://localhost:8083/health
curl http://localhost:8084/health
```

---

## Correr Tests

Los tests corren automaticamente en GitHub Actions. Para correrlos localmente necesitas Go instalado o usar los contenedores.

```bash
cd api-gateway      && go test ./... && cd ..
cd groups-service   && go test ./... && cd ..
cd habits-service   && go test ./... && cd ..
cd realtime-service && go test ./... && cd ..
```

Frontend:

```bash
cd frontend
npm test
```

---

## Flujo De Trabajo En Equipo

1. Crea tu rama desde `main`: `git checkout -b feature/nombre-feature`
2. Trabaja unicamente dentro del directorio del servicio que te corresponde
3. Abre un Pull Request a `main`
4. GitHub Actions corre build y tests automaticamente
5. Merge solo cuando el CI este en verde

Cada servicio se despliega de forma independiente. No modifiques archivos fuera de tu directorio sin avisar al equipo.

---

## Estructura Del Proyecto

```text
dydi/
├── .env.example        -> copia esto como .env y llena los valores
├── docker-compose.yml  -> orquestacion local
├── api-gateway/        -> unico punto de entrada publico
├── groups-service/     -> grupos y membresias
├── habits-service/     -> habitos, check-ins, rachas y penalizaciones
├── realtime-service/   -> WebSocket hub para eventos en tiempo real
├── frontend/           -> Vue 3 SPA (PWA-ready)
└── supabase/
    └── migrations/     -> schema de la base de datos
```

---

## Variables De Entorno Por Servicio

Cada servicio tiene su propio `.env.example` con las variables necesarias para correrlo de forma independiente o configurar Render.

Nunca subas archivos `.env` al repositorio.
