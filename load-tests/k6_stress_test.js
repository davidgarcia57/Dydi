import ws from 'k6/ws';
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Rate, Counter } from 'k6/metrics';
import { SharedArray } from 'k6/data';

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
// ⚠️ PRERREQUISITO: el handshake WS valida membresía contra groups-service
// (fail-closed), así que los grupos deben EXISTIR y el usuario del TOKEN debe
// ser miembro activo de todos. Antes de correr:
//   ./load-tests/seed.sh up -u <uuid-del-usuario-del-token> -n 1000
// Eso genera load-tests/loadtest_groups.json, que este script reparte entre VUs.
//
// Variables (pásalas con -e):
//   -e BASE_URL=https://api-gateway.onrender.com   (HTTP)
//   -e WS_URL=wss://api-gateway.onrender.com        (WebSocket)
//   -e TOKEN=<JWT de Supabase del usuario sembrado> (ver nota abajo)
//   -e PEAK=5000          (usuarios WS pico; mantén PEAK/grupos ≤ 8)
//
// Cómo sacar el TOKEN: entra a la app, abre DevTools → Console y corre
//   (await window.supabase?.auth.getSession())   — o copia el header
//   Authorization: Bearer <...> de cualquier request en la pestaña Network.
// El access_token de Supabase caduca ~1h; esta prueba dura ~9 min, alcanza.
//
// Exportar resultados:  k6 run --out json=resultados.json k6_stress_test.js
// =============================================================

// UUIDs reales sembrados por seed.sh (SharedArray: una sola copia en memoria
// aunque haya 5000 VUs).
const GROUPS = new SharedArray('loadtest groups', function () {
  try {
    return JSON.parse(open('./loadtest_groups.json'));
  } catch (e) {
    return []; // setup() aborta con mensaje claro
  }
});

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

// Accept-Encoding: gzip para que el server comprima las respuestas igual que a
// un navegador real (k6 no lo manda por defecto). k6 descomprime en el cliente
// y reporta tamaños en claro, pero los BYTES EN EL CABLE —lo que Render factura
// como egreso— viajan comprimidos.
const authHeaders = {
  headers: { Authorization: `Bearer ${TOKEN}`, 'Accept-Encoding': 'gzip' },
};

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

// Aborta ANTES de la rampa si falta el seed o los grupos no alcanzan:
// correr 9 minutos para medir puros 403/409 no genera datos útiles.
export function setup() {
  if (GROUPS.length === 0) {
    throw new Error(
      'loadtest_groups.json vacío o ausente — corre primero: ./load-tests/seed.sh up -u <uuid> -n 1000'
    );
  }
  const worstCase = Math.ceil(PEAK / GROUPS.length);
  if (worstCase > 8) {
    throw new Error(
      `PEAK=${PEAK} sobre ${GROUPS.length} grupos = ${worstCase} conexiones/grupo, ` +
        'pero el server topa a 8 — siembra más grupos (seed.sh up -n <más>)'
    );
  }
}

// --- ESCENARIO 1: WebSocket ---
export function wsConnect() {
  // Reparte cada VU en una sala distinta para no chocar con el tope de 8/grupo.
  // El handshake valida que el usuario del TOKEN sea miembro activo del grupo
  // (realtime → groups, fail-closed), por eso los UUIDs vienen del seed.
  const groupID = GROUPS[__VU % GROUPS.length];
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

  const ok = check(res, { 'WS handshake 101': (r) => r && r.status === 101 });
  if (!ok) {
    // Rechazo en el handshake (403 membresía, 409 sala llena, timeout de cold
    // start u OOM): nunca dispara 'open' ni 'error', hay que contarlo aquí.
    wsDroppedRate.add(true);
  }
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
