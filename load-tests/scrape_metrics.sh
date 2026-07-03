#!/usr/bin/env bash
# scrape_metrics.sh — captura /metrics de los 4 servicios en serie de tiempo (CSV).
#
# Es la evidencia central del paper: memoria RSS creciendo hacia el OOM kill,
# goroutines, saturación del pool de pgx y eventos WS caídos, muestreados cada
# N segundos MIENTRAS k6 mete carga. El dashboard de Render solo da una gráfica
# gruesa; esto da datos tabulares para el análisis (Keila).
#
# Solo usa curl + awk (nada que instalar). Un scrape fallido también se registra
# (metric=scrape_error): si el servicio muere por OOM, ese hueco ES el dato.
#
# Uso:
#   ./load-tests/scrape_metrics.sh -o resultados.csv \
#     gateway=https://tu-gateway.onrender.com \
#     groups=https://tu-groups.onrender.com \
#     habits=https://tu-habits.onrender.com \
#     realtime=https://tu-realtime.onrender.com
#
# Sin targets usa el docker-compose local (localhost:8080/8082/8083/8084).
# Flags: -i <segundos entre scrapes, default 5> · -o <csv, default metrics.csv>
#        -d <duración total en segundos; sin -d corre hasta Ctrl+C/SIGTERM>
set -euo pipefail

INTERVAL=5
OUT="metrics.csv"
DURATION=0
TARGETS=()

while [ $# -gt 0 ]; do
  case "$1" in
    -i) INTERVAL="${2:?falta el número después de -i}"; shift 2 ;;
    -o) OUT="${2:?falta el archivo después de -o}"; shift 2 ;;
    -d) DURATION="${2:?falta el número después de -d}"; shift 2 ;;
    *=*) TARGETS+=("$1"); shift ;;
    *) echo "arg no reconocido: $1 (targets van como nombre=url)"; exit 2 ;;
  esac
done

if [ ${#TARGETS[@]} -eq 0 ]; then
  TARGETS=(
    gateway=http://localhost:8080
    groups=http://localhost:8082
    habits=http://localhost:8083
    realtime=http://localhost:8084
  )
fi

# Familias que le importan al experimento; el resto del dump se ignora.
# (bucket de histogramas fuera: el P95 del paper lo da k6 del lado cliente;
#  count/sum bastan para throughput y latencia media del lado servidor.)
METRIC_RE='^(process_resident_memory_bytes|process_cpu_seconds_total|go_goroutines|go_memstats_heap_alloc_bytes|http_request_duration_seconds_(count|sum)|db_pool_|realtime_)'

[ -s "$OUT" ] || echo "unix_ts,iso_ts,service,metric,labels,value" > "$OUT"

STOP=0
trap 'STOP=1' INT TERM

START=$(date +%s)
echo "→ scrapeando ${#TARGETS[@]} servicios cada ${INTERVAL}s → $OUT (Ctrl+C para parar)"

while [ "$STOP" -eq 0 ]; do
  NOW=$(date +%s)
  ISO=$(date -u -d "@$NOW" +%Y-%m-%dT%H:%M:%SZ)
  [ "$DURATION" -gt 0 ] && [ $((NOW - START)) -ge "$DURATION" ] && break

  for t in "${TARGETS[@]}"; do
    name="${t%%=*}"; url="${t#*=}"
    if body=$(curl -sf --max-time 4 "$url/metrics"); then
      printf '%s\n' "$body" | awk -v ts="$NOW" -v iso="$ISO" -v svc="$name" -v re="$METRIC_RE" '
        /^#/ { next }
        $1 !~ re { next }
        {
          # separa nombre{labels} valor  |  nombre valor
          val = $NF
          head = $0; sub(/ [^ ]+$/, "", head)
          labels = ""
          if (match(head, /\{.*\}/)) {
            labels = substr(head, RSTART+1, RLENGTH-2)
            metric = substr(head, 1, RSTART-1)
          } else {
            metric = head
          }
          gsub(/"/, "\"\"", labels)
          printf "%s,%s,%s,%s,\"%s\",%s\n", ts, iso, svc, metric, labels, val
        }' >> "$OUT"
    else
      # El servicio no respondió: probable cold start, throttling u OOM kill.
      echo "$NOW,$ISO,$name,scrape_error,\"\",1" >> "$OUT"
    fi
  done

  sleep "$INTERVAL" &
  wait $! || true   # sleep en bg para que Ctrl+C/SIGTERM corte sin esperar el intervalo
done

echo "✓ scrape terminado → $OUT ($(wc -l < "$OUT" | tr -d ' ') filas)"
