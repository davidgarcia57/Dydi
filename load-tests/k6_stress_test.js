import ws from 'k6/ws';
import { check, sleep } from 'k6';
import { Trend, Rate, Counter } from 'k6/metrics';

// --- MÉTRICAS PERSONALIZADAS ---
// Esto es lo que usarás para generar las gráficas de tu tesis
const connectionTime = new Trend('ws_connection_time');
const droppedConnections = new Rate('ws_dropped_rate');
const messagesSent = new Counter('ws_messages_sent');
const messagesReceived = new Counter('ws_messages_received');

// --- RAMPAS DE CARGA (Escenario de Estrés) ---
// Simula el ingreso progresivo de usuarios hasta llegar a 5,000 conexiones concurrentes
export const options = {
  stages: [
    { duration: '30s', target: 100 },   // Rampa 1: 100 usuarios (Línea base)
    { duration: '1m',  target: 1000 },  // Rampa 2: Salto a 1,000 usuarios
    { duration: '2m',  target: 2500 },  // Rampa 3: Salto a 2,500 usuarios
    { duration: '3m',  target: 5000 },  // Rampa 4: El objetivo de la tesis (5,000 usuarios)
    { duration: '2m',  target: 5000 },  // Meseta: Mantener 5,000 usuarios por 2 min (si el server sobrevive)
    { duration: '30s', target: 0 },     // Enfriamiento
  ],
  // Nota: k6 permite exportar resultados a CSV o JSON usando flags de terminal
  // ej: k6 run --out json=resultados.json script.js
};

// --- CONFIGURACIÓN ---
// Estas variables se pueden pasar por la terminal, por ejemplo:
// k6 run -e WS_URL=wss://tu-app.onrender.com -e TOKEN=tu_jwt k6_stress_test.js
const API_URL  = __ENV.WS_URL   || 'ws://localhost:8080';
const GROUP_ID = __ENV.GROUP_ID || '00000000-0000-0000-0000-000000000000'; // Pon el UUID de un grupo real
const TOKEN    = __ENV.TOKEN    || 'TU_JWT_TOKEN_AQUI';

export default function () {
  const url = `${API_URL}/ws/${GROUP_ID}?token=${TOKEN}`;

  const res = ws.connect(url, null, function (socket) {
    const startTime = Date.now();

    socket.on('open', function () {
      // Registrar el tiempo que tardó en hacer el Handshake HTTP -> WS
      connectionTime.add(Date.now() - startTime); 
      
      // Simular comportamiento humano: Enviar un "latido" (ping/heartbeat) aleatorio
      // para mantener la conexión viva y obligar al CPU de Go a trabajar
      socket.setInterval(function timeout() {
        socket.send(JSON.stringify({ 
          type: 'ping', 
          timestamp: Date.now() 
        }));
        messagesSent.add(1);
      }, Math.random() * 15000 + 10000); // Cada 10-25 segundos
    });

    socket.on('message', function (msg) {
      messagesReceived.add(1);
    });

    socket.on('error', function (e) {
      // Ocurre cuando Render hace OOM Kill o Throttling y tumba el socket de golpe
      if (e.error() != 'websocket: close sent') {
        droppedConnections.add(1);
      }
    });

    socket.on('close', function () {
      // Fin natural o forzado de la conexión
    });
  });

  // Verificamos que el handshake haya sido exitoso (Código 101 Switching Protocols)
  check(res, { 'Handshake WS exitoso (HTTP 101)': (r) => r && r.status === 101 });
  
  // Pequeña pausa simulando el "think time" de nuevos usuarios entrando
  sleep(0.5);
}
