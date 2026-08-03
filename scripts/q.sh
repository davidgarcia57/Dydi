#!/usr/bin/env bash
# q.sh — query READ-ONLY a la BD de Supabase, vía Docker (psql). No instala nada.
# Es el equivalente "database-query" casero: la sesión va forzada a
# default_transaction_read_only=on, así un INSERT/UPDATE/DELETE falla por diseño.
#
# En Windows, córrelo desde la distro WSL:
#   wsl -d ubuntu bash -lc './scripts/q.sh "select id, display_name from users limit 5;"'
#
# Uso:
#   ./scripts/q.sh "select count(*) from groups;"
#   ./scripts/q.sh -f consulta.sql
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# Lee DATABASE_URL del .env sin sourcearlo (la contraseña puede traer caracteres raros).
DB="$(grep -E '^DATABASE_URL=' "$ROOT/.env" 2>/dev/null | head -1 | cut -d= -f2-)"
[ -n "${DB:-}" ] || { echo "✗ Falta DATABASE_URL en .env (copia .env.example)"; exit 1; }

if [ "${1:-}" = "-f" ]; then
  FILE="${2:?uso: ./scripts/q.sh -f archivo.sql}"
  [ -f "$FILE" ] || { echo "✗ No existe el archivo: $FILE"; exit 1; }
  exec docker run --rm -i -e PGOPTIONS='-c default_transaction_read_only=on' \
    -v "$(cd "$(dirname "$FILE")" && pwd)":/sql postgres:15 \
    psql "$DB" -v ON_ERROR_STOP=1 -f "/sql/$(basename "$FILE")"
fi

SQL="${1:-}"
[ -n "$SQL" ] || { echo 'uso: ./scripts/q.sh "SELECT ..."   |   ./scripts/q.sh -f archivo.sql'; exit 2; }

exec docker run --rm -i -e PGOPTIONS='-c default_transaction_read_only=on' \
  postgres:15 psql "$DB" -v ON_ERROR_STOP=1 -c "$SQL"
