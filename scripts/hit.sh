#!/usr/bin/env bash
# hit.sh — golpea el backend de Dydi sin abrir el front. Dos caminos:
#
#   • Servicio directo (groups|habits|realtime): estampa X-Internal-Token +
#     X-User-ID, igual que hace el gateway tras validar el JWT. NO necesitas un
#     JWT — es el camino ideal para probar handlers en local.
#   • Gateway (puerto 8080): exige un JWT ES256 real de Supabase (el gateway lo
#     valida contra el JWKS; no se puede forjar). Cópialo del navegador
#     (DevTools → red → Authorization) y pásalo en DYDI_JWT.
#
# Requiere el stack levantado (docker-compose up) y curl (ya viene en WSL).
#
# Uso:
#   ./scripts/hit.sh GET  habits /habits
#   ./scripts/hit.sh POST groups /groups '{"name":"Test"}'
#   DYDI_USER=<uuid> ./scripts/hit.sh GET habits /habits/today
#   DYDI_JWT=<token> ./scripts/hit.sh GET gateway /groups
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

METHOD="${1:?uso: hit.sh METHOD SERVICIO RUTA [body-json]  (SERVICIO: gateway|groups|habits|realtime)}"
SVC="${2:?servicio: gateway|groups|habits|realtime}"
ROUTE="${3:?ruta, ej: /habits}"
BODY="${4:-}"

case "$SVC" in
  gateway)  PORT=8080 ;;
  groups)   PORT=8082 ;;
  habits)   PORT=8083 ;;
  realtime) PORT=8084 ;;
  *) echo "✗ servicio desconocido: $SVC (usa gateway|groups|habits|realtime)"; exit 2 ;;
esac

ARGS=(-sS -i -X "$METHOD")

if [ "$SVC" = "gateway" ]; then
  # El gateway valida el JWT; no acepta el token interno desde fuera.
  [ -n "${DYDI_JWT:-}" ] || { echo "✗ gateway exige un JWT real en DYDI_JWT (cópialo del navegador)"; exit 1; }
  ARGS+=(-H "Authorization: Bearer $DYDI_JWT")
else
  # Camino interno: el secreto compartido + el user que quieras simular.
  TOKEN="$(grep -E '^INTERNAL_TOKEN=' "$ROOT/.env" 2>/dev/null | head -1 | cut -d= -f2-)"
  [ -n "${TOKEN:-}" ] || { echo "✗ Falta INTERNAL_TOKEN en .env"; exit 1; }
  USER_ID="${DYDI_USER:-00000000-0000-0000-0000-000000000001}"
  ARGS+=(-H "X-Internal-Token: $TOKEN" -H "X-User-ID: $USER_ID")
fi

if [ -n "$BODY" ]; then
  ARGS+=(-H "Content-Type: application/json" -d "$BODY")
fi

curl "${ARGS[@]}" "http://localhost:${PORT}${ROUTE}"
echo
