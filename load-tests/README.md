# Pruebas de estrés (k6) — experimento de la tesis

Todo lo necesario para las corridas del paper: sembrar datos, meter carga
(rampas de WebSockets + HTTP) y capturar la telemetría de los 4 servicios.
**Nada se instala local**: k6 y psql corren vía Docker (regla del proyecto).

## Los 5 archivos

| Archivo | Qué hace |
|---|---|
| `seed.sh` | Siembra N grupos `loadtest-*` en Supabase con el usuario de prueba como miembro (el handshake WS valida membresía — sin esto, todo da 403). Exporta los UUIDs a `loadtest_groups.json`. |
| `token.sh` | Imprime un JWT fresco de la cuenta loadtest (password grant contra Supabase Auth). Se acabó copiar tokens del navegador. |
| `k6_stress_test.js` | La prueba: rampa de WS hasta `PEAK` repartida entre los grupos sembrados + tráfico HTTP constante para el P95. |
| `scrape_metrics.sh` | Muestrea `/metrics` de los 4 servicios cada 5s → CSV (RAM, CPU, goroutines, pool de pgx, WS caídos). Un scrape fallido también se registra: si el servicio muere por OOM, ese hueco ES el dato. |
| `run_experiment.sh` | Una corrida completa y reproducible: scraper + k6 + artefactos en `results/<fecha>-peak<N>-rep<M>/`. |

## Paso 0 — Configurar `.env` (una sola vez)

1. Registra una cuenta DEDICADA en la app (ej. `loadtest@...`). No uses tu
   cuenta personal: el seed le cuelga ~1000 grupos a su UI.
2. Agrega al `.env` (ver `.env.example`): `LOADTEST_EMAIL`, `LOADTEST_PASSWORD`
   y las 4 `LOADTEST_*_URL` de Render.

Con eso, todo lo demás se resuelve solo (uuid, token, targets).

## Paso 1 — Sembrar (una sola vez, o tras un `down`)

```bash
./load-tests/seed.sh up -n 1000        # resuelve el uuid vía LOADTEST_EMAIL
./load-tests/seed.sh up -n 50 --dry-run   # ensayo: valida todo y hace ROLLBACK
```

Regla: `PEAK / grupos ≤ 8` (tope de conexiones por sala). 1000 grupos aguantan
PEAK=5000 sobrado. Al terminar TODO el experimento: `./load-tests/seed.sh down`
(borra SOLO los grupos con `invite_code` `LOADTEST-%`).

## Paso 2 — Correr

```bash
# Contra Render (targets y token salen del .env):
./load-tests/run_experiment.sh -p 5000 -r 1

# Piloto local (docker-compose arriba; targets explícitos ganan al .env):
./load-tests/run_experiment.sh -p 100 -r 0 \
  gateway=http://localhost:8080 groups=http://localhost:8082 \
  habits=http://localhost:8083 realtime=http://localhost:8084

# Token manual si prefieres (o si la cuenta loadtest no está en .env):
DYDI_JWT=<token> ./load-tests/run_experiment.sh -p 5000 -r 1
```

Cada corrida deja en `results/<fecha>-peak<N>-rep<M>/`:
`metadata.json` (qué se corrió y contra qué commit) · `metrics.csv` (serie de
tiempo del servidor) · `summary.json` (agregados de k6: P95, error rate, checks)
· `k6_output.txt`. Con `--raw` además todos los datapoints de k6 (pesa mucho).

## La matriz del paper

4 niveles × **mínimo 3 repeticiones** por nivel, en horarios similares (una
sola corrida en free tier tiene demasiado ruido para revisión por pares):

```bash
for rep in 1 2 3; do
  for peak in 100 1000 2500 5000; do
    ./load-tests/run_experiment.sh -p $peak -r $rep   # token fresco por corrida
    sleep 600   # reposo: que la memoria regrese a su línea base
  done
done
```

Consejos para que los datos salgan limpios:

- **Despierta los servicios antes** (`curl gateway/health` y espera ~1 min): el
  cold start de Render contamina la primera rampa si no es lo que quieres medir.
  (O mídelo a propósito en una corrida aparte — `realtime_cold_start_seconds` ya
  se captura.)
- Deja **~10 min de reposo** entre corridas para que los servicios regresen a
  su línea base de memoria.
- El JWT caduca ~1h; `run_experiment.sh` acuña uno fresco por corrida vía
  `token.sh`, así que no hay que renovar nada a mano.
- `k6` saliendo con error por *threshold reventado* NO invalida la corrida: para
  la hipótesis, encontrar el límite es el resultado.

## Qué esperar (hipótesis)

k6 inyecta la rampa 100 → 1000 → 2500 → 5000 VUs (~9 min). Según la hipótesis,
antes de las 5000 conexiones los 512 MB de Render se llenan → OOM kill → k6 lo
registra en `ws_dropped_rate` y el scraper en `process_resident_memory_bytes` +
`scrape_error`. Ese cruce de series (RAM del servidor vs. conexiones vs. drops)
es la gráfica central del paper.
