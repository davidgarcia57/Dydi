# Evaluación de una arquitectura de microservicios en tiempo real sobre la capa gratuita de Render: un estudio de caso bajo inyección de carga

> **Estado del borrador (2026-07-02):** metodología y arquitectura completas; la
> sección de Resultados contiene la línea base del piloto y espera los datos de
> la matriz experimental (4 niveles × 3 repeticiones). Los pendientes están
> marcados como `[PENDIENTE: ...]`. Objetivo de extensión: ~6 páginas en formato
> de dos columnas.

**Autores:** García Páez David Israel · Solis Flores Irvin Alfonso · Cervantes
Guerrero Keila Yuridia · Casiano Gamzi Juan David
*Universidad Tecnológica de Durango (UTD) — Proyecto Integradora 2026*

---

## Resumen

Las capas gratuitas de plataformas como servicio (PaaS) imponen restricciones
severas de memoria (512 MB), suspensión por inactividad y horas de cómputo
limitadas, por lo que suelen descartarse para sistemas distribuidos en tiempo
real. Este trabajo evalúa empíricamente si una arquitectura de microservicios
—cuatro servicios en Go repartidos en cuatro cuentas independientes de la capa
gratuita de Render, con WebSockets propios para difusión en tiempo real— puede
sostener tráfico concurrente sin morir por agotamiento de memoria (OOM kill) ni
degradar su latencia de forma inaceptable. Mediante experimentación controlada
con k6 (rampas de 100, 1 000, 2 500 y 5 000 usuarios virtuales, tres
repeticiones por nivel) y telemetría embebida de bajo costo (Prometheus + logs
estructurados), se midió la relación entre la carga inyectada y el consumo de
recursos, el percentil 95 de latencia HTTP y la tasa de conexiones WebSocket
caídas. `[PENDIENTE: una oración con el hallazgo principal de la matriz —
p. ej. el punto de quiebre observado o la resistencia hasta 5 000 VUs.]` Los
resultados aportan un modelo replicable para que equipos académicos desplieguen
prácticas profesionales en infraestructura sin costo.

**Palabras clave:** microservicios · pruebas de estrés · WebSockets · PaaS ·
capa gratuita · k6 · observabilidad

`[PENDIENTE: traducción del resumen al inglés (Abstract) si la venue lo pide.]`

---

## 1. Introducción

Los proyectos académicos de ingeniería de software rara vez llegan a producción:
el costo de la infraestructura es una barrera que las instituciones y los
estudiantes no siempre pueden absorber. Las capas gratuitas de PaaS (Render,
Railway, Fly.io) ofrecen una salida, pero sus restricciones —512 MB de RAM por
servicio, suspensión tras 15 minutos de inactividad, horas mensuales acotadas—
generan la percepción de que solo sirven para demos triviales, no para
arquitecturas distribuidas con requisitos de tiempo real.

Este estudio somete esa percepción a prueba con un sistema real: **Dydi**, una
aplicación SaaS de *accountability* social en la que grupos de amigos rastrean
hábitos diarios y gamifican las consecuencias de incumplirlos. Dydi opera sobre
cuatro microservicios en Go desplegados en **cuatro cuentas independientes** de
la capa gratuita de Render (una estrategia deliberada para multiplicar los
recursos gratuitos disponibles), con difusión en tiempo real implementada sobre
WebSockets propios —no delegada a servicios gestionados— precisamente para que
el costo de esa pieza quede dentro del experimento.

**Pregunta de investigación:** ¿puede esta infraestructura fragmentada y gratuita
soportar tráfico concurrente y procesamiento en tiempo real sin colapsar por
falta de memoria (OOM kill) ni degradar inaceptablemente su latencia?

Las contribuciones del trabajo son tres:

1. **Evidencia empírica** de la relación entre carga concurrente (100–5 000
   usuarios virtuales) y consumo de recursos/latencia/entrega en tiempo real en
   una PaaS gratuita, medida sobre un sistema completo con autenticación,
   base de datos gestionada y WebSockets.
2. **Un arnés experimental reproducible y de costo cero** (k6 + telemetría
   embebida en ~200 líneas de Go por servicio), publicado como artefacto abierto
   junto con los datos crudos de todas las corridas.
3. **Un hallazgo metodológico** sobre la validación de instrumentos en este tipo
   de estudios: el piloto reveló que un límite de tasa por usuario en el gateway
   invalidaba por completo la medición (§4.6), un confusor que habría pasado
   inadvertido sin telemetría del lado del servidor.

## 2. Trabajo relacionado

`[PENDIENTE: 3–5 referencias por subtema. Sugerencias de búsqueda:
"microservices performance evaluation", "serverless/PaaS free tier benchmark",
"WebSocket scalability measurement", "load testing k6 methodology".]`

La literatura sobre evaluación de desempeño de microservicios se concentra en
infraestructura dedicada o nubes de pago (AWS, GCP), donde el cuello de botella
es el diseño del sistema y no las cuotas del proveedor. Los estudios sobre
plataformas gratuitas son mayormente informales (entradas de blog, foros) y
carecen de metodología replicable. Este trabajo se distingue por (a) medir una
capa gratuita con rigor experimental, (b) usar un sistema real en producción y
no un *benchmark* sintético, y (c) instrumentar la telemetría dentro del propio
sistema bajo prueba, respetando sus restricciones de memoria.

## 3. Arquitectura del sistema bajo prueba

### 3.1 Topología

```
frontend (Vercel) ─┐
mobile (Expo)     ─┼─► api-gateway ─► groups-service    (Render, cuenta 2)
                   │  (Render, c.1) ├─► habits-service   (Render, cuenta 3)
                   │   JWT ES256    └─► realtime-service (Render, cuenta 4)
                   │   vía JWKS            WebSockets
                   └────────────────► Supabase (PostgreSQL 15 + Auth)
```

Cuatro servicios en Go 1.24 (enrutador chi v5, driver pgx v5):

- **api-gateway** — único punto de entrada; valida JWT (ES256 vía JWKS de
  Supabase), estampa la identidad (`X-User-ID`) y un secreto interno
  (`X-Internal-Token`) en cada petición proxeada, de modo que los servicios
  —públicos por vivir en cuentas separadas— rechazan todo tráfico que no venga
  del gateway.
- **groups-service** — grupos, membresías y propuestas con votación.
- **habits-service** — hábitos, check-ins, rachas y penitencias.
- **realtime-service** — difusión de eventos por WebSocket
  (`github.com/coder/websocket`), con verificación de membresía en el handshake.

### 3.2 Decisiones de diseño condicionadas por la capa gratuita

| Restricción de Render Free | Decisión de diseño |
|---|---|
| 512 MB de RAM por servicio | Go compilado (binarios ~20 MB de RSS en reposo); sin sidecars ni agentes de monitoreo externos |
| Horas de cómputo por cuenta | 4 cuentas independientes, un servicio por cuenta |
| Suspensión tras 15 min de inactividad | Endpoint `/ops/wake` en el gateway que despierta a los tres servicios en cascada; cron externo cada 12 min en horario pico |
| Sin métricas exportables nativas | Telemetría embebida: módulo `obs.go` por servicio (histogramas Prometheus en `/metrics` + logs JSON con `slog`) |

La pieza de tiempo real se implementó **a propósito** sobre WebSockets propios en
lugar de un servicio gestionado (p. ej. Supabase Realtime): el objetivo del
estudio es medir el costo de sostener conexiones persistentes dentro de los
512 MB, y delegarla lo habría sacado del sistema bajo prueba.

## 4. Metodología

### 4.1 Diseño

Estudio **cuantitativo, correlacional y evaluativo**: mide el grado de relación
entre la **variable independiente** (número de conexiones concurrentes
inyectadas en rampas progresivas) y las **variables dependientes** (consumo de
RAM/CPU por servicio, percentil 95 de latencia HTTP, tiempo de conexión
WebSocket y tasa de conexiones caídas), con el fin de localizar el punto de
quiebre de la capa gratuita.

### 4.2 Población, muestra y unidad de análisis

La población es el universo teórico de peticiones HTTP y conexiones WebSocket
que la arquitectura puede recibir en producción. La unidad de análisis es la
transacción individual registrada por k6. Se emplea **muestreo no probabilístico
intencional**: el equipo inyecta deliberadamente los perfiles de carga extrema
necesarios para comprobar la hipótesis (100, 1 000, 2 500 y 5 000 usuarios
virtuales). Se excluyen del análisis el tráfico orgánico incidental, los errores
de red del lado del inyector y los pings de keep-alive del cron externo —este
último se **pausa durante las corridas** para no contaminar percentiles.

### 4.3 Instrumentos

- **Inyección:** `k6_stress_test.js` con dos escenarios simultáneos: tráfico
  HTTP constante (20 iteraciones/s sobre endpoints REST autenticados) para medir
  latencia bajo estrés, y una rampa de WebSockets hasta el pico configurado,
  repartida entre 1 000 grupos sembrados (límite de 8 conexiones por sala).
- **Telemetría del servidor:** módulo `obs.go` en cada servicio expone
  `/metrics` (Prometheus): histogramas de latencia por ruta,
  `process_resident_memory_bytes`, goroutines, métricas del pool de conexiones
  a PostgreSQL, `realtime_cold_start_seconds` y contador de eventos descartados.
  Un scraper (`scrape_metrics.sh`) muestrea los cuatro servicios cada 5 s y
  serializa a CSV; un scrape fallido también se registra —si un servicio muere
  por OOM, ese hueco en la serie **es** el dato.
- **Orquestación:** `run_experiment.sh` ejecuta cada corrida de forma
  reproducible y deja por corrida: `metadata.json` (parámetros, *commit hash*
  del código medido), `metrics.csv` (serie de tiempo del servidor),
  `summary.json` (agregados de k6) y la salida íntegra del inyector.

### 4.4 Matriz experimental

4 niveles de carga × **3 repeticiones** por nivel, en ventanas horarias
similares (una sola corrida en capa gratuita, con vecinos ruidosos, no es
evidencia suficiente):

| Nivel (VUs pico) | Repeticiones | Duración por corrida | Reposo entre corridas |
|---:|---:|---|---|
| 100 | 3 | ~10 min (rampa 9 min) | 10 min |
| 1 000 | 3 | ~10 min | 10 min |
| 2 500 | 3 | ~10 min | 10 min |
| 5 000 | 3 | ~10 min | 10 min |

### 4.5 Procedimiento

1. **Preparación** (resp. Solis Flores): verificación de despliegues, variables
   de entorno y registro de telemetría en los cuatro servicios.
2. **Ejecución** (resp. García Páez): servicios despiertos previamente (el
   arranque en frío de Render, ~11 s medidos, contaminaría la primera rampa);
   cron de keep-alive pausado; inyección de las rampas.
3. **Recolección** (resp. Cervantes Guerrero): consolidación de los artefactos
   por corrida y cálculo de percentiles desde los histogramas.
4. **Análisis** (equipo completo): cruce de series —VUs vs. RAM por servicio
   vs. tasa de caída vs. P95—.

### 4.6 Validación del instrumento (piloto)

El piloto (100 VUs contra el despliegue real) produjo resultados anómalos:
88.30 % de peticiones HTTP fallidas y 93.68 % de conexiones WebSocket
rechazadas, pese a que la telemetría del servidor mostraba consumo mínimo
(≤ 44 MB de RAM). El cruce de ambas fuentes reveló la causa: el gateway aplica
un **límite de tasa por usuario** (5 req/s, ráfaga 20) y todos los usuarios
virtuales comparten la misma cuenta de pruebas, por lo que el experimento
completo compartía una sola cubeta de *tokens* — se estaba midiendo el limitador,
no la arquitectura.

Se corrigió haciendo los límites configurables por variable de entorno
(elevados a 2 000 req/s solo durante las corridas; los valores de producción se
conservan por defecto) y se repitió el piloto:

| Métrica | Piloto 1 (artefacto) | Piloto 2 (válido) |
|---|---:|---:|
| Peticiones HTTP fallidas | 88.30 % | **0.00 %** (0 / 21 596) |
| Conexiones WS caídas | 93.68 % | **0.00 %** (0 / 408) |
| Duración media de sesión WS | 4.96 s | 88 s |
| P95 HTTP (exitosas) | 425 ms | 404 ms |

Este episodio ilustra por qué la telemetría del lado del servidor no es
opcional en estudios de carga: sin ella, el 88 % de fallos se habría atribuido
erróneamente a la capa gratuita.

## 5. Resultados

### 5.1 Línea base (100 VUs, piloto válido)

Corrida `20260702-195005-peak100-rep1` (commit `af38ae4`): 21 596 peticiones
HTTP (35.3 req/s sostenidas) sin fallos; 408 sesiones WebSocket sin caídas,
tiempo de conexión promedio 682 ms (P95 805 ms); latencia HTTP promedio 221 ms,
P95 404 ms. Consumo pico de memoria por servicio (de 512 MB disponibles):

| Servicio | RAM pico | % del límite |
|---|---:|---:|
| groups-service | 44.5 MB | 8.7 % |
| api-gateway | 43.2 MB | 8.4 % |
| realtime-service | 34.7 MB | 6.8 % |
| habits-service | 20.9 MB | 4.1 % |

A 100 conexiones concurrentes la arquitectura opera con holgura de un orden de
magnitud, lo que establece la línea base contra la cual contrastar los niveles
superiores.

### 5.2 Matriz completa

`[PENDIENTE: correr la matriz 4×3 tras el congelamiento de funcionalidades.
Para cada nivel reportar: mediana de las 3 repeticiones de (a) RAM pico por
servicio, (b) P95 HTTP, (c) ws_dropped_rate, (d) tiempo de conexión WS;
- Figura 1: VUs vs. RAM por servicio (con línea horizontal en 512 MB).
- Figura 2: VUs vs. P95 HTTP y vs. ws_dropped_rate (eje doble).
- Tabla 3: resumen de la matriz con desviaciones entre repeticiones.
- Si hubo OOM kill: serie de tiempo de la corrida donde ocurre, mostrando el
  hueco de scrapes y el disparo de ws_dropped_rate (la "gráfica central").]`

### 5.3 Arranque en frío

`[PENDIENTE: corrida dedicada midiendo realtime_cold_start_seconds y el costo
del wake en cascada; hoy se tiene una observación informal de ~11 s para el
gateway.]`

## 6. Discusión

`[PENDIENTE tras la matriz. Guía: ¿dónde se quiebra primero la arquitectura y
por qué (realtime por memoria de conexiones persistentes vs. gateway por CPU de
proxy+TLS vs. pool de PostgreSQL)? ¿La fragmentación en 4 cuentas aísla los
fallos o los propaga? ¿Qué margen real ofrece para uso académico (usuarios
concurrentes soportados con calidad aceptable)? Contrastar con la hipótesis:
se esperaba OOM kill antes de 5 000 VUs.]`

## 7. Amenazas a la validez

- **Constructo — carga de un solo usuario:** todos los VUs se autentican con la
  misma cuenta; los patrones de consulta y caché de base de datos no representan
  a miles de usuarios distintos. El límite de tasa por usuario se elevó por
  configuración durante las corridas (§4.6), documentando los valores.
- **Interna — ruido de la plataforma:** en capa gratuita los recursos son
  compartidos con inquilinos desconocidos ("vecinos ruidosos"); se mitiga con
  3 repeticiones por nivel en ventanas horarias similares y reportando
  dispersión, no solo promedios. Errores transitorios del borde (Cloudflare
  520/502) se observaron a razón de ~1/40 peticiones incluso en reposo.
- **Externa — generalización:** es un estudio de caso de un sistema (Go +
  Render + Supabase); los hallazgos informan sobre esta clase de arquitectura,
  no sobre cualquier PaaS gratuita. La replicabilidad se apoya en el artefacto
  publicado.
- **Instrumento — inyector único:** k6 corre desde una sola máquina/red; el
  ancho de banda del inyector podría ser cuello de botella en los niveles altos.
  `[PENDIENTE: reportar el uso de red del inyector en 5 000 VUs.]`

## 8. Conclusiones y trabajo futuro

`[PENDIENTE tras la matriz. Estructura sugerida: (1) respuesta directa a la
pregunta de investigación con cifras; (2) el modelo replicable para equipos
académicos (costo $0, requisitos, límites observados); (3) el aporte
metodológico del §4.6; (4) futuro: comparar contra la misma arquitectura en un
solo contenedor de pago, medir consistencia de entrega de eventos, repetir con
usuarios distintos.]`

## Disponibilidad de datos y código

El código fuente completo, el arnés experimental (`load-tests/`) y los datos
crudos de todas las corridas (`load-tests/results/`, con el *commit hash* del
código medido en cada `metadata.json`) están versionados en el repositorio del
proyecto: `https://github.com/davidgarcia57/Dydi`. `[PENDIENTE: decidir si el
repo se hace público o se publica un espejo/release para la revisión.]`

## Referencias

`[PENDIENTE: completar en APA 7. Candidatas mínimas:]`

- Grafana Labs. (2025). *k6 documentation*. https://grafana.com/docs/k6/
- Newman, S. (2021). *Building microservices* (2.ª ed.). O'Reilly Media.
- Prometheus Authors. (2025). *Prometheus documentation*. https://prometheus.io/docs/
- Render. (2026). *Free instance types — Render docs*. https://render.com/docs/free
- Supabase. (2026). *Supabase documentation*. https://supabase.com/docs
- `[PENDIENTE: 2–3 artículos arbitrados de evaluación de desempeño de
  microservicios / WebSockets a escala.]`
