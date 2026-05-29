# Dydi

SaaS de accountability social donde grupos de amigos rastrean hábitos diarios y gamifican las consecuencias.

Proyecto académico — UTD Integradora 2025.

---

## Arquitectura

```
frontend/           → Vue 3 + Vite + Tailwind        (Vercel)
api-gateway/        → Go 1.22 + chi v5               (Render — Cuenta 1)
groups-service/     → Go 1.22 + chi v5               (Render — Cuenta 2)
habits-service/     → Go 1.22 + chi v5               (Render — Cuenta 3)
realtime-service/   → Go 1.22 + WebSocket            (Render — Cuenta 4)
```

Auth y base de datos via **Supabase** (free tier).

---

## Requisitos para desarrollo local

- [Docker Desktop](https://www.docker.com/products/docker-desktop/) con WSL2 habilitado
- Git

No necesitas Go ni Node instalados localmente — Docker los provee.

---

## Configuración del entorno

### 1. Clonar el repositorio

```bash
git clone <url-del-repo>
cd dydi
```

### 2. Crear los archivos de variables de entorno

Copia los `.env.example` de cada servicio y llena los valores:

```bash
cp api-gateway/.env.example      api-gateway/.env
cp groups-service/.env.example   groups-service/.env
cp habits-service/.env.example   habits-service/.env
cp realtime-service/.env.example realtime-service/.env
cp frontend/.env.example         frontend/.env
```

Los valores que necesitas del proyecto Supabase:

| Variable | Dónde encontrarla |
|---|---|
| `SUPABASE_JWT_SECRET` | Supabase Dashboard → Project Settings → API → JWT Secret |
| `DATABASE_URL` | Supabase Dashboard → Project Settings → Database → Connection string (URI) |
| `VITE_SUPABASE_URL` | Supabase Dashboard → Project Settings → API → Project URL |
| `VITE_SUPABASE_ANON_KEY` | Supabase Dashboard → Project Settings → API → anon public |

### 3. Levantar todos los servicios

```bash
docker-compose up --build
```

La primera vez tarda varios minutos mientras descarga las imágenes base. Las siguientes ejecuciones son mucho más rápidas.

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

> **Nota:** En local se usa PostgreSQL en Docker. En producción se usa la base de datos de Supabase.

---

## Verificar que todo está corriendo

```bash
docker-compose ps
```

Todos los servicios deben aparecer como `Up`. También puedes probar los health checks:

```bash
curl http://localhost:8080/health   # api-gateway
curl http://localhost:8082/health   # groups-service
curl http://localhost:8083/health   # habits-service
curl http://localhost:8084/health   # realtime-service
```

---

## Correr los tests

### Servicios Go

```bash
# Requiere Go instalado: sudo snap install go --classic
cd api-gateway && go test ./...
cd ../groups-service && go test ./...
cd ../habits-service && go test ./...
cd ../realtime-service && go test ./...
```

### Frontend

```bash
cd frontend
npm install
npm run test
```

Los tests también corren automáticamente en GitHub Actions en cada PR.

---

## Flujo de trabajo en equipo

1. Crea una rama desde `main`: `git checkout -b feature/nombre-feature`
2. Haz tus cambios **solo dentro del directorio del servicio que te corresponde**
3. Abre un Pull Request a `main`
4. GitHub Actions corre los tests y el build automáticamente
5. Solo merge cuando el CI esté en verde ✅

> Cada servicio es independiente. **No modifiques archivos fuera de tu directorio de servicio** sin avisar al equipo.

---

## Estructura del proyecto

```
dydi/
├── api-gateway/        ← Único punto de entrada público
├── groups-service/     ← Grupos y membresías
├── habits-service/     ← Hábitos, check-ins, rachas y penalizaciones
├── realtime-service/   ← WebSocket hub para eventos en tiempo real
├── frontend/           ← Vue 3 SPA
├── supabase/
│   └── migrations/     ← Schema de la base de datos
└── docker-compose.yml  ← Orquestación local
```

---

## Variables de entorno de producción

Cada servicio tiene un `.env.example` con las variables necesarias. En producción, configura estas variables en el dashboard de Render (Environment → Environment Variables).

**Nunca subas archivos `.env` al repositorio.**
