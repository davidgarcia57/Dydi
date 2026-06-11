# Pruebas de Estrés (Load Testing) con k6

Este directorio contiene los scripts para poner a prueba la resiliencia y el consumo de recursos de la arquitectura de microservicios de Dydi (Tesis / Actividad 1.4 y 1.5).

## Requisitos

1. Descargar e instalar [k6 de Grafana](https://k6.io/docs/get-started/installation/).
2. Asegúrate de tener los microservicios corriendo (ya sea en local con Docker o desplegados en Render).
3. Necesitas un **Token JWT válido** y un **ID de Grupo válido**.

## ¿Cómo obtener el Token JWT?

1. Entra a tu frontend de Dydi e inicia sesión.
2. Abre la consola del navegador (F12) -> pestaña Application / Local Storage.
3. Copia el token de la llave que guarda Supabase (ej. `sb-*-auth-token`).
4. Ve al dashboard de Dydi, selecciona un grupo y copia el UUID de la URL.

## Ejecución del Experimento

Abre tu terminal en esta carpeta y ejecuta el siguiente comando, reemplazando las variables con tus datos reales:

```bash
k6 run \
  -e WS_URL=wss://tu-api-gateway.onrender.com \
  -e GROUP_ID=c73bcdcc-2669-4bf6-81d3-e4ae73fb11fd \
  -e TOKEN=eyJhbGciOiJIUzI1NiIs... \
  k6_stress_test.js
```

### Exportar Datos para las Gráficas (Tesis)

Para que puedas construir las gráficas de dispersión (Conexiones vs Latencia), puedes exportar la salida a un archivo CSV que luego puedes abrir en Excel o Python (Matplotlib/Pandas):

```bash
k6 run --out csv=resultados_tesis.csv -e WS_URL=... -e GROUP_ID=... -e TOKEN=... k6_stress_test.js
```

### ¿Qué sucederá?

k6 inyectará usuarios virtuales en "rampas":
1. Subirá a 100 usuarios en 30s.
2. Escalará a 1,000 en 1m.
3. Escalará a 2,500 en 2m.
4. Escalará a 5,000 en 3m.

Mientras tanto, debes monitorear la plataforma (Docker stats o el Dashboard de Render) para ver **el consumo exacto de Memoria RAM (MB) y CPU (%)**. 

Según la hipótesis, en algún punto antes de las 5,000 conexiones, los 512 MB de Render se llenarán, provocando un *OOM Kill* (reinicio). Esto generará errores en k6 que se registrarán en la métrica `ws_dropped_rate`, demostrando exactamente cuál es la capacidad límite de la arquitectura Go hiper-optimizada en un entorno gratuito.
