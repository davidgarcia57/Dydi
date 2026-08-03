#!/usr/bin/env bash
# test-db.sh — corre los tests de integración de Go contra un Postgres
# EFÍMERO en Docker. Nunca toca Supabase.
#
# Por qué existe: los tests unitarios del repo pasan `nil` como pool, así que
# solo ejercitan las guardas que retornan antes de llegar a una query. Nada del
# SQL real se ejecuta en CI. Esto levanta una BD desechable, le aplica el shim
# de Supabase + las migraciones reales, y corre los tests marcados con la build
# tag `integration`.
#
# En Windows, córrelo desde la distro WSL:
#   wsl -d ubuntu bash -lc './scripts/test-db.sh'
#
# Uso:
#   ./scripts/test-db.sh                    # todos los servicios con tests de integración
#   ./scripts/test-db.sh habits-service     # solo uno
#   ./scripts/test-db.sh habits-service TestSpinIsIdempotent   # un solo test
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
NET=dydi-test-net
PG=dydi-test-pg
PGPASS=testpass
# Nombre de BD y usuario fijos: la BD vive y muere con este script.
DSN="postgres://postgres:${PGPASS}@${PG}:5432/postgres?sslmode=disable"

SVC="${1:-}"
RUN="${2:-}"

cleanup() {
  docker rm -f "$PG" >/dev/null 2>&1 || true
  docker network rm "$NET" >/dev/null 2>&1 || true
}
trap cleanup EXIT

echo "▶ levantando Postgres efímero…"
cleanup
docker network create "$NET" >/dev/null
docker run -d --name "$PG" --network "$NET" \
  -e POSTGRES_PASSWORD="$PGPASS" \
  -e POSTGRES_DB=postgres \
  postgres:15 >/dev/null

# pg_isready en vez de un sleep fijo: el arranque varía por máquina.
echo -n "▶ esperando a que acepte conexiones"
for _ in $(seq 1 60); do
  if docker exec "$PG" pg_isready -U postgres -q 2>/dev/null; then break; fi
  echo -n "."
  sleep 1
done
echo
docker exec "$PG" pg_isready -U postgres -q || { echo "✗ Postgres no arrancó"; exit 1; }

echo "▶ aplicando shim + migraciones…"
docker exec -i "$PG" psql -U postgres -d postgres -v ON_ERROR_STOP=1 -q \
  < "$ROOT/supabase/test-shim.sql"
for m in "$ROOT"/supabase/migrations/*.sql; do
  echo "  · $(basename "$m")"
  docker exec -i "$PG" psql -U postgres -d postgres -v ON_ERROR_STOP=1 -q < "$m"
done

# Servicios que tienen tests de integración (los que hablan con la BD).
SERVICES=(habits-service groups-service)
if [ -n "$SVC" ]; then SERVICES=("$SVC"); fi

FAILED=0
for s in "${SERVICES[@]}"; do
  echo
  echo "▶ tests de integración: $s"
  ARGS=(-tags=integration -race -count=1)
  [ -n "$RUN" ] && ARGS+=(-run "$RUN")
  # V=1 ./scripts/test-db.sh → muestra cada test (útil para confirmar que
  # corrieron de verdad y no se saltaron por falta de TEST_DATABASE_URL).
  [ "${V:-}" = "1" ] && ARGS+=(-v)
  # --network: el contenedor de Go tiene que resolver el hostname del de Postgres.
  if ! docker run --rm --network "$NET" \
      -v "$ROOT/$s":/app -v "$ROOT/shared":/shared \
      -v "$ROOT/.gocache":/gocache -v "$ROOT/.gomodcache":/gomodcache \
      -e GOCACHE=/gocache -e GOMODCACHE=/gomodcache \
      -e TEST_DATABASE_URL="$DSN" \
      -w /app golang:1.24 go test "${ARGS[@]}" ./...; then
    FAILED=1
  fi
done

echo
if [ "$FAILED" -eq 0 ]; then
  echo -e "\033[1;32mOK — tests de integración en verde\033[0m"
else
  echo -e "\033[1;31m✗ hubo tests de integración en rojo\033[0m"
  exit 1
fi
