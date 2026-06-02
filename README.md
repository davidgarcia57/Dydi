# Dydi

SaaS de accountability social donde grupos de amigos rastrean hábitos diarios y gamifican las consecuencias. Proyecto académico — UTD Integradora 2025.

---

## Arquitectura

```
frontend/           → Vue 3 + Vite + Tailwind        (Vercel)
api-gateway/        → Go 1.22 + chi v5               (Render — Cuenta 1)
groups-service/     → Go 1.22 + chi v5               (Render — Cuenta 2)
habits-service/     → Go 1.22 + chi v5               (Render — Cuenta 3)
realtime-service/   → Go 1.22 + WebSocket            (Render — Cuenta 4)
```

Auth y base de datos via **Supabase** (free tier). El único punto de entrada público es `api-gateway`.

---

## Requisitos

- [Docker Desktop](https://www.docker.com/products/docker-desktop/) con WSL2 habilitado
- Git

No necesitas Go ni Node instalados localmente — Docker los provee.

---

## Configuración inicial (una sola vez)

### 1. Clonar el repositorio

```bash
git clone <url-del-repo>
cd dydi
```

### 2. Crear el archivo de variables de entorno

```bash
cp .env.example .env
```

Abre `.env` y llena las tres variables con los valores de tu proyecto en Supabase:

| Variable | Dónde encontrarla |
|---|---|
| `SUPABASE_JWT_SECRET` | Dashboard → Settings → API → **JWT Secret** |
| `VITE_SUPABASE_URL` | Dashboard → Settings → API → **Project URL** |
| `VITE_SUPABASE_ANON_KEY` | Dashboard → Settings → API → **anon public** |

> El resto de las variables (puertos, DB local, URLs internas) ya están configuradas en `docker-compose.yml` y no requieren ajuste.

### 3. Levantar todos los servicios

```bash
docker-compose up --build
```

La primera vez tarda varios minutos descargando imágenes. Las siguientes son mucho más rápidas.

---

## Puertos locales

| Servicio | URL |
|---|---|
| Frontend | http://localhost:5173 |
| API Gateway | http://localhost:8080 |
| groups-service | http://localhost:8082 |
| habits-service | http://localhost:8083 |
| realtime-service | http://localhost:8084 |
| PostgreSQL | localhost:5432 |

> En local se usa PostgreSQL corriendo en Docker. En producción se usa la base de datos de Supabase.

---

## Verificar que todo está corriendo

```bash
docker-compose ps
```

Todos los servicios deben aparecer como `Up`. También puedes probar los health checks manualmente:

```bash
curl http://localhost:8080/health
curl http://localhost:8082/health
curl http://localhost:8083/health
curl http://localhost:8084/health
```

---

## Correr los tests

Los tests corren automáticamente en GitHub Actions en cada PR. Para correrlos localmente necesitas Go instalado:

```bash
# instalar Go si no lo tienes
sudo snap install go --classic

# correr tests por servicio
cd api-gateway      && go test ./... && cd ..
cd groups-service   && go test ./... && cd ..
cd habits-service   && go test ./... && cd ..
cd realtime-service && go test ./... && cd ..
```

---

## Flujo de trabajo en equipo

1. Crea tu rama desde `main`: `git checkout -b feature/nombre-feature`
2. Trabaja **únicamente dentro del directorio del servicio que te corresponde**
3. Abre un Pull Request a `main`
4. GitHub Actions corre build y tests automáticamente
5. Merge solo cuando el CI esté en verde ✅

> Cada servicio se despliega de forma independiente. **No modifiques archivos fuera de tu directorio sin avisar al equipo.**

---

## Estructura del proyecto

```
dydi/
├── .env.example        ← copia esto como .env y llena los valores
├── docker-compose.yml  ← orquestación local
├── api-gateway/        ← único punto de entrada público
├── groups-service/     ← grupos y membresías
├── habits-service/     ← hábitos, check-ins, rachas y penalizaciones
├── realtime-service/   ← WebSocket hub para eventos en tiempo real
├── frontend/           ← Vue 3 SPA (PWA-ready)
└── supabase/
    └── migrations/     ← schema de la base de datos
```

---

## Variables de entorno por servicio (producción / standalone)

Cada servicio tiene su propio `.env.example` con las variables necesarias para correrlo de forma independiente (sin docker-compose) o para configurar las variables en Render.

**Nunca subas archivos `.env` al repositorio.**
