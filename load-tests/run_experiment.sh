#!/usr/bin/env bash
# run_experiment.sh — una corrida completa del experimento (tesis), reproducible.
#
# Orquesta: scraper de /metrics (fondo) → k6 en Docker (grafana/k6) → guarda
# TODO en load-tests/results/<fecha>-peak<N>-rep<M>/ para el análisis:
#   metadata.json   qué se corrió (peak, grupos, commit, targets, horario)
#   metrics.csv     serie de tiempo de los 4 servicios (RAM/CPU/pool/WS caídos)
#   summary.json    agregados finales de k6 (P95, tasas de error, checks)
#   k6_output.txt   salida completa de k6 (progreso + thresholds)
#   raw.csv         (solo con --raw) todos los datapoints de k6 — pesa MUCHO
#
# Prerrequisitos:
#   1. ./load-tests/seed.sh up -u <uuid> -n 1000   (grupos reales sembrados)
#   2. Token JWT del MISMO usuario sembrado (caduca ~1h; saca uno fresco)
#
# Uso (contra Render):
#   DYDI_JWT=<token> ./load-tests/run_experiment.sh -p 5000 -r 1 \
#     gateway=https://tu-gateway.onrender.com \
#     groups=https://tu-groups.onrender.com \
#     habits=https://tu-habits.onrender.com \
#     realtime=https://tu-realtime.onrender.com
#
# Sin targets usa el docker-compose local (para pilotos). El target 'gateway'
# también define BASE_URL/WS_URL de k6. Flags: -p <peak> · -r <repetición>
# · -t <token> (alternativa a DYDI_JWT) · --raw (datapoints completos de k6)
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LT="$ROOT/load-tests"

PEAK=5000
REP=1
TOKEN="${DYDI_JWT:-}"
RAW=0
TARGETS=()

while [ $# -gt 0 ]; do
  case "$1" in
    -p) PEAK="${2:?falta el número después de -p}"; shift 2 ;;
    -r) REP="${2:?falta el número después de -r}"; shift 2 ;;
    -t) TOKEN="${2:?falta el token después de -t}"; shift 2 ;;
    --raw) RAW=1; shift ;;
    *=*) TARGETS+=("$1"); shift ;;
    *) echo "arg no reconocido: $1"; exit 2 ;;
  esac
done

# Sin token explícito: acuña uno fresco con la cuenta loadtest del .env.
if [ -z "$TOKEN" ]; then
  TOKEN="$("$LT/token.sh" || true)"
  [ -n "$TOKEN" ] && echo "→ JWT fresco acuñado vía token.sh (caduca ~1h)"
fi
[ -n "$TOKEN" ] || { echo "✗ Falta el JWT: DYDI_JWT=<token>, -t <token>, o LOADTEST_EMAIL/PASSWORD en .env"; exit 1; }
[ -s "$LT/loadtest_groups.json" ] || {
  echo "✗ Falta loadtest_groups.json — corre primero: ./load-tests/seed.sh up"; exit 1; }

# Sin targets en args: primero busca las URLs de Render en .env; si tampoco
# están, cae al docker-compose local (pilotos).
if [ ${#TARGETS[@]} -eq 0 ]; then
  for n in gateway groups habits realtime; do
    up="$(echo "$n" | tr '[:lower:]' '[:upper:]')"
    v="$(grep -E "^LOADTEST_${up}_URL=" "$ROOT/.env" 2>/dev/null | head -1 | cut -d= -f2-)"
    [ -n "$v" ] && TARGETS+=("$n=$v")
  done
fi
if [ ${#TARGETS[@]} -eq 0 ]; then
  TARGETS=(
    gateway=http://localhost:8080
    groups=http://localhost:8082
    habits=http://localhost:8083
    realtime=http://localhost:8084
  )
fi

# El target 'gateway' es la puerta de entrada de k6 (HTTP y WS).
GATEWAY=""
for t in "${TARGETS[@]}"; do
  [ "${t%%=*}" = "gateway" ] && GATEWAY="${t#*=}"
done
[ -n "$GATEWAY" ] || { echo "✗ Falta el target gateway=<url>"; exit 1; }
WS_GATEWAY="${GATEWAY/http:/ws:}"; WS_GATEWAY="${WS_GATEWAY/https:/wss:}"

GROUP_TOTAL=$(tr -cd ',' < "$LT/loadtest_groups.json" | wc -c | tr -d ' ')
GROUP_TOTAL=$((GROUP_TOTAL + 1))

STAMP=$(date +%Y%m%d-%H%M%S)
DIR="$LT/results/${STAMP}-peak${PEAK}-rep${REP}"
mkdir -p "$DIR"

COMMIT=$(git -C "$ROOT" rev-parse --short HEAD 2>/dev/null || echo "desconocido")
{
  echo "{"
  echo "  \"started_at\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\","
  echo "  \"peak\": $PEAK,"
  echo "  \"repetition\": $REP,"
  echo "  \"groups_seeded\": $GROUP_TOTAL,"
  echo "  \"commit\": \"$COMMIT\","
  echo "  \"gateway\": \"$GATEWAY\","
  echo "  \"targets\": \"${TARGETS[*]}\""
  echo "}"
} > "$DIR/metadata.json"

echo "══ Corrida peak=$PEAK rep=$REP → $DIR"

# 1) Scraper en fondo, muestreando ANTES/DURANTE/DESPUÉS de la rampa.
"$LT/scrape_metrics.sh" -i 5 -o "$DIR/metrics.csv" "${TARGETS[@]}" &
SCRAPER=$!
trap 'kill -TERM $SCRAPER 2>/dev/null || true' EXIT

# 2) k6 en Docker (no hay k6 local). --network host para que localhost funcione
#    en pilotos; contra Render da igual. results/ va montado para los exports.
K6_ARGS=(run
  -e "BASE_URL=$GATEWAY" -e "WS_URL=$WS_GATEWAY" -e "TOKEN=$TOKEN" -e "PEAK=$PEAK"
  --summary-export "/lt/results/${STAMP}-peak${PEAK}-rep${REP}/summary.json"
)
[ "$RAW" -eq 1 ] && K6_ARGS+=(--out "csv=/lt/results/${STAMP}-peak${PEAK}-rep${REP}/raw.csv")

K6_EXIT=0
# --user: el contenedor de k6 corre como uid 12345 y no puede escribir
# summary.json en results/ (propiedad del usuario del host).
docker run --rm -i --network host --user "$(id -u):$(id -g)" -v "$LT":/lt -w /lt grafana/k6 \
  "${K6_ARGS[@]}" k6_stress_test.js 2>&1 | tee "$DIR/k6_output.txt" || K6_EXIT=$?

# 3) Cierra el scraper con gracia y reporta.
kill -TERM $SCRAPER 2>/dev/null || true
wait $SCRAPER 2>/dev/null || true
trap - EXIT

echo "══ Artefactos en: $DIR"
ls -lh "$DIR"
# k6 sale ≠0 si un threshold falló: para la tesis eso también es un resultado
# válido (encontraste el límite), así que se reporta pero no se oculta.
[ "$K6_EXIT" -ne 0 ] && echo "⚠ k6 terminó con código $K6_EXIT (¿threshold reventado? revisa summary.json)"
exit 0
