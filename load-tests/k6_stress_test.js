import ws from 'k6/ws';
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Rate, Counter } from 'k6/metrics';

// =============================================================
// Dydi — prueba de carga (tesis)
//
// Dos escenarios corriendo en paralelo:
//   1) websockets : N usuarios concurrentes conectados al realtime-service,
//      REPARTIDOS entre muchos grupos (porque el server topa a 8 conexiones por
//      grupo: meter 5000 al mismo grupo solo mediría rechazos, no capacidad).
//   2) http_p95   : tráfico HTTP constante a la API REST mientras los WS cargan,
//      para medir el P95 de latencia del camino REST bajo estrés.
//
// Variables (pásalas con -e):
//   -e BASE_URL=https://api-gateway.onrender.com   (HTTP)
//   -e WS_URL=wss://api-gateway.onrender.com        (WebSocket)
//   -e TOKEN=<JWT de Supabase de un usuario real>   (ver nota abajo)
//   -e GROUP_COUNT=1000   (cuántas salas distintas; mantén PEAK/GROUP_COUNT < 8)
//   -e PEAK=5000          (usuarios WS pico)
//
// Cómo sacar el TOKEN: entra a la app, abre DevTools → Console y corre
//   (await window.supabase?.auth.getSession())   — o copia el header
//   Authorization: Bearer <...> de cualquier request en la pestaña Network.
// El access_token de Supabase caduca ~1h; esta prueba dura ~9 min, alcanza.
//
// Exportar resultados:  k6 run --out json=resultados.json k6_stress_test.js
// =============================================================

// --- MÉTRICAS PERSONALIZADAS (para las gráficas de la tesis) ---
const wsConnectTime  = new Trend('ws_connect_time', true); // ms hasta el 'open'
const wsDroppedRate  = new Rate('ws_dropped_rate');        // % conexiones caídas
const wsMsgsReceived = new Counter('ws_msgs_received');    // eventos recibidos del server
const wsMsgsSent     = new Counter('ws_msgs_sent');

// --- CONFIG ---
const BASE_URL    = __ENV.BASE_URL || 'http://localhost:8080';
const WS_URL      = __ENV.WS_URL   || 'ws://localhost:8080';
const TOKEN       = __ENV.TOKEN    || 'TU_JWT_TOKEN_AQUI';
const PEAK        = parseInt(__ENV.PEAK || '5000', 10);
const GROUP_COUNT = parseInt(__ENV.GROUP_COUNT || '1000', 10);

const authHeaders = { headers: { Authorization: `Bearer ${TOKEN}` } };

export const options = {
  scenarios: {
    // ── 1) Carga de WebSockets concurrentes (rampa hasta PEAK) ──────────────
    websockets: {
      executor: 'ramping-vus',
      exec: 'wsConnect',
      startVUs: 0,
      stages: [
        { duration: '30s', target: Math.round(PEAK * 0.02) }, // línea base
        { duration: '1m',  target: Math.round(PEAK * 0.2) },
        { duration: '2m',  target: Math.round(PEAK * 0.5) },
        { duration: '3m',  target: PEAK },                     // objetivo de la tesis
        { duration: '2m',  target: PEAK },                     // meseta
        { duration: '30s', target: 0 },                        // enfriamiento
      ],
    },
    // ── 2) Tráfico HTTP constante para medir P95 bajo carga ─────────────────
    http_p95: {
      executor: 'constant-arrival-rate',
      exec: 'httpRead',
      rate: 20,            // 20 req/s
      timeUnit: '1s',
      duration: '9m',      // cubre todo el ciclo de WS
      preAllocatedVUs: 50,
      maxVUs: 200,
    },
  },
  thresholds: {
    // Free tier: umbrales generosos; ajústalos según tu hipótesis.
    http_req_duration: ['p(95)<800', 'p(99)<2000'],
    http_req_failed:   ['rate<0.05'], // tasa de error REST (variable del paper)
    ws_connect_time:   ['p(95)<2000'],
    ws_dropped_rate:   ['rate<0.10'],
    checks:            ['rate>0.95'],
  },
};

// --- ESCENARIO 1: WebSocket ---
export function wsConnect() {
  // Reparte cada VU en una sala distinta para no chocar con el tope de 8/grupo.
  // El handshake WS no valida membresía (solo JWT válido + cupo en la sala),
  // así que un solo token sirve para muchas salas y el groupID puede ser libre.
  const groupID = `loadtest-${__VU % GROUP_COUNT}`;
  const url = `${WS_URL}/ws/${groupID}?token=${TOKEN}`;

  const start = Date.now(); // medir el handshake completo (antes de connect)

  const res = ws.connect(url, null, function (socket) {
    socket.on('open', function () {
      wsConnectTime.add(Date.now() - start);
      wsDroppedRate.add(false);

      // Latido para mantener viva la conexión y forzar el read-loop del server.
      socket.setInterval(function () {
        socket.send(JSON.stringify({ type: 'ping', ts: Date.now() }));
        wsMsgsSent.add(1);
      }, Math.random() * 15000 + 10000); // cada 10-25s

      // Sesión de ~60-120s y cierre → simula churn real y libera el VU.
      socket.setTimeout(function () {
        socket.close();
      }, Math.random() * 60000 + 60000);
    });

    socket.on('message', function () {
      wsMsgsReceived.add(1);
    });

    socket.on('error', function (e) {
      if (e.error() !== 'websocket: close sent') {
        wsDroppedRate.add(true); // OOM kill / throttling de Render tumbó el socket
      }
    });
  });

  check(res, { 'WS handshake 101': (r) => r && r.status === 101 });
}

// --- ESCENARIO 2: HTTP (mide P95 del camino REST bajo carga) ---
export function httpRead() {
  // GETs sin efectos secundarios (no ensucian datos): catálogo y grupos del user.
  const habits = http.get(`${BASE_URL}/api/habits`, {
    ...authHeaders,
    tags: { name: 'list_habits' },
  });
  check(habits, { 'GET /habits 200': (r) => r.status === 200 });

  const groups = http.get(`${BASE_URL}/api/groups`, {
    ...authHeaders,
    tags: { name: 'list_groups' },
  });
  check(groups, { 'GET /groups 200': (r) => r.status === 200 });

  sleep(1);
}
