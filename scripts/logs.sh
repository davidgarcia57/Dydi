#!/usr/bin/env bash
# logs.sh — tail de logs (slog JSON de obs.go) de un servicio del compose.
# Acepta nombre corto (gateway|groups|habits|realtime) o el nombre del compose.
#
# Uso:
#   ./scripts/logs.sh habits           # últimas 100 líneas
#   ./scripts/logs.sh gateway -f       # follow en vivo
#   ./scripts/logs.sh frontend --tail=20
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

RAW="${1:?uso: ./scripts/logs.sh <servicio> [flags de docker compose logs]}"
shift || true

# Mapea los alias cortos a los nombres reales del compose.
case "$RAW" in
  gateway)  SVC=api-gateway ;;
  groups)   SVC=groups-service ;;
  habits)   SVC=habits-service ;;
  realtime) SVC=realtime-service ;;
  *)        SVC="$RAW" ;;  # frontend, mobile o el nombre exacto que se pase
esac

# Si no se pasa -f ni --tail, default a las últimas 100 líneas.
case "$*" in
  *-f*|*--follow*|*--tail*) docker compose logs "$@" "$SVC" ;;
  *)                        docker compose logs --tail=100 "$@" "$SVC" ;;
esac
