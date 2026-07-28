---
title: "Actividad 4.1. Elaboración del informe final y comunicación visual de hallazgos"
subtitle: "Proyecto Dydi · Equipo 3 · Universidad Tecnológica de Durango · Integradora 2026"
date: "28 de julio de 2026"
lang: es
---

PROYECTO DYDI · UTD · INTEGRADORA 2026 · Modalidad: equipo · Equipo 3, Resiliencia y
consumo de recursos en microservicios.

> **Qué entrega esta actividad.** Las tres subtareas se completaron en orden y cada
> una dejó un entregable parcial verificable antes de integrar el informe final:
> (a) la estructura del informe, el tipo de documento y el reparto de redacción;
> (b) resultados, discusión y conclusiones redactados, con la declaración de
> confiabilidad, validez y limitaciones; (c) las visualizaciones seleccionadas, la
> narrativa que las conecta y el cuadro de mando con jerarquía de indicadores.
> El informe final integrado es `Documentos/articulo.md` en el repositorio del
> proyecto.

Pregunta de investigación: ¿puede una arquitectura de microservicios fragmentada en
cuatro cuentas de la capa gratuita de Render soportar tráfico concurrente y
procesamiento en tiempo real sin colapsar por falta de memoria (OOM) ni degradar
inaceptablemente su latencia?

---

# Subtarea a. Estructura del informe

## a.1 Tipo de documento y audiencia real

El proyecto tiene tres audiencias con necesidades distintas, y un solo documento no
las cubre.

| Audiencia | Qué necesita | Documento que la atiende |
|---|---|---|
| Evaluación académica (profesora y jurado de la Integradora) | Rigor metodológico, trazabilidad del dato y defensa de las decisiones | Informe académico (este entregable) más la defensa oral (Act. 4.2a) |
| Comunidad técnica y académica que consulte el trabajo | Replicabilidad, evidencia y contraste con la literatura | El mismo informe, en formato de artículo publicable |
| Equipos que evalúan desplegar sin presupuesto | El veredicto práctico y sus límites, sin metodología | Producto de divulgación (Act. 4.2b) |

Decisión tomada: **informe académico en formato de artículo científico (IMRyD
extendido)**, con extensión objetivo de ~6 páginas a dos columnas y normas APA 7.ª
edición. Las tres opciones del planteamiento se evaluaron así:

- *Informe ejecutivo.* Descartado. Este trabajo no sustenta una decisión de
  inversión; produce evidencia sobre una pregunta. El registro ejecutivo se conserva,
  pero como pieza de divulgación dirigida a otro público (Act. 4.2b).
- *Informe técnico.* Descartado. La documentación del sistema ya vive en el
  repositorio (`README.md`, `docs/architecture.md`, `CLAUDE.md`). Lo que este trabajo
  agrega es la medición, así que la arquitectura entra al informe solo como objeto
  bajo prueba (§3 del artículo).
- *Artículo académico.* Elegido. El estudio se formuló desde la Unidad I como
  cuantitativo, correlacional y evaluativo, con hipótesis y umbrales predefinidos
  (Protocolo v2, §6.4). El género que corresponde a ese diseño es el artículo
  empírico.

## a.2 Estructura completa del informe

Cada sección declara de dónde sale su evidencia y con qué unidad del curso se
articula. Las extensiones son objetivo, en formato de dos columnas.

| # | Sección | Contenido | Ext. | Origen de la evidencia | Articulación |
|---|---|---|---|---|---|
| 0 | Título, autores, resumen y palabras clave | Resumen estructurado (problema, método, resultado, aporte) | 0.3 p | Síntesis final | — |
| 1 | Introducción | Problema (barrera económica del despliegue), pregunta de investigación, tres contribuciones | 0.6 p | Protocolo v2 §1–§3 | Unidad I (planteamiento y objetivos) |
| 2 | Trabajo relacionado | Cuatro antecedentes y el hueco que deja la literatura | 0.5 p | Act. 3.1 (matriz de fuentes) | Unidad II (marco teórico) |
| 3 | Arquitectura del sistema bajo prueba | Topología, los 4 servicios, decisiones de diseño condicionadas por el free tier | 0.7 p | Código del repositorio | — |
| 4 | Metodología | Diseño, población/muestra, instrumentos, matriz experimental, procedimiento y validación del instrumento (piloto) | 1.2 p | Protocolo v2 · Act. 3.2 · Act. 3.3 | Unidad III (recolección) |
| 5 | Resultados | Línea base, matriz ejecutada y punto de quiebre, arranque en frío, hallazgos operativos del free tier | 1.4 p | Act. 3.4 (banco) · Act. 3.5 | Unidad III |
| 6 | Discusión | Mecanismo del quiebre, aislamiento frente a propagación, margen real de uso, traducción a usuarios, costo operativo | 1.0 p | Act. 3.5 y literatura de §2 | Unidad II y III |
| 7 | Amenazas a la validez | Constructo, interna, externa, instrumento, alcance, exclusiones | 0.5 p | Bitácora `matriz.log` | Unidad III |
| 8 | Conclusiones y trabajo futuro | Respuesta a la pregunta, viabilidad operativa, aporte metodológico, 4 líneas futuras | 0.5 p | Síntesis | Unidad I (cierre del objetivo) |
| 9 | Disponibilidad de datos y código | Repositorio, arnés y datos crudos con *commit hash* por corrida | 0.1 p | `load-tests/` | — |
| 10 | Referencias (APA 7) | 11 fuentes: 5 primarias arbitradas más documentación técnica | 0.4 p | Act. 3.1 | Unidad II |

Hay dos decisiones de estructura que conviene poder sostener en la defensa.

La primera es que §4.6, «Validación del instrumento (piloto)», ocupa una sección
propia. El piloto detectó que se estaba midiendo un limitador de tasa en lugar de la
arquitectura, y ese episodio figura entre las tres contribuciones declaradas del
trabajo, así que tiene su lugar dentro de la metodología.

La segunda es que §5.4, «Hallazgos operativos de la capa gratuita», se ubica en
Resultados. Las cuotas, las pausas por inactividad y los créditos de ráfaga son
propiedades del objeto estudiado y responden parte de la pregunta de investigación.
Tratarlas como limitaciones del experimento las sacaría del hallazgo.

## a.3 Asignación de responsabilidades de redacción

El reparto sigue los roles que ya declara el protocolo metodológico: quien ejecutó una
parte del experimento redacta esa parte y la defiende. Cada sección tiene un redactor
y un revisor distintos.

| Sección | Redacta | Revisa | Base de su autoridad |
|---|---|---|---|
| §0 Resumen · §1 Introducción | García Páez David | Casiano Gamzi Juan David | Coordinación del artículo y ejecución de corridas |
| §2 Trabajo relacionado | Casiano Gamzi Juan David | Cervantes Guerrero Keila | Responsable de análisis; construyó la matriz de fuentes (Act. 3.1) |
| §3 Arquitectura del sistema | Solis Flores Irvin | García Páez David | Responsable de setup y despliegue de las 4 cuentas |
| §4 Metodología (§4.1–§4.5) | Cervantes Guerrero Keila | Solis Flores Irvin | Responsable de datos y telemetría; autora del banco (Act. 3.4) |
| §4.6 Validación del instrumento | García Páez David | Cervantes Guerrero Keila | Ejecutó ambos pilotos y detectó el confusor |
| §5 Resultados | Cervantes Guerrero Keila | Casiano Gamzi Juan David | Consolidó los artefactos por corrida |
| §5.4 Hallazgos operativos | Solis Flores Irvin | García Páez David | Diagnosticó las suspensiones y la restauración de la BD |
| §6 Discusión | Equipo completo (sesión conjunta) | García Páez David | Interpretación acordada, no delegable |
| §7 Amenazas a la validez | Casiano Gamzi Juan David | Solis Flores Irvin | Custodia de los criterios de exclusión |
| §8 Conclusiones | García Páez David | Equipo completo | Cierre del objetivo declarado en la Unidad I |
| §9 Datos y código · §10 Referencias | Cervantes Guerrero Keila | Casiano Gamzi Juan David | Versionado de artefactos y bibliografía |
| Figuras y cuadro de mando (subtarea c) | Cervantes Guerrero Keila | García Páez David | Genera las figuras desde `analyze_results.py` |

## a.4 Criterios editoriales comunes

Estas reglas se acordaron antes de repartir la redacción, para que el documento
conserve una sola voz pese a tener seis redactores.

Ninguna cifra entra al texto sin procedencia. Toda cifra debe existir en
`load-tests/analysis/stats.json` o en un artefacto de corrida; si no está ahí, no se
escribe. Nadie transcribe números a mano.

Se reporta mediana con rango mín–máx, nunca el promedio solo. En capa gratuita los
valores atípicos por «vecinos ruidosos» son esperables, y ocultarlos falsearía la
dispersión.

Las cifras llevan espacio fino como separador de millar (1 000, 64 706) y punto
decimal, con la unidad siempre explícita (ms, MB, %, VUs).

Los verbos van en pasado para lo medido y en presente para lo que el sistema hace. Una
estimación nunca se redacta como medición: la traducción a usuarios (§6) va marcada
como «estimación condicionada, no resultado observado».

Se citan según APA 7. Los sitios de documentación técnica (Render, Supabase, k6,
Prometheus) se tratan como fuentes secundarias fechadas.

El informe vive en `Documentos/articulo.md` dentro del repositorio, con una rama y un
*pull request* por sección y el revisor asignado como aprobador obligatorio.

> Entregable parcial verificable (a): esta sección, más el esqueleto de
> `Documentos/articulo.md` con las diez secciones y sus responsables asignados.

---

# Subtarea b. Resultados, discusión y conclusiones

## b.1 Organización de los resultados por objetivo

Los resultados se ordenan por objetivo específico y no por la cronología de las
corridas. La tabla siguiente es el índice de trazabilidad entre lo que se prometió
medir en la Unidad I y lo que se reporta.

| Objetivo específico (Unidad I) | Hallazgo | Evidencia | Sección |
|---|---|---|---|
| Medir consumo de recursos bajo carga concurrente | La memoria crece proporcional a las conexiones y se concentra en la ruta WS; el peor servicio llega a 46.6 % del límite | RAM pico por servicio, 2 niveles × 3 rep. | §5.1, §5.2 |
| Medir latencia y su degradación | El plano HTTP se degrada sin fallar: P95 de 826 a 1 045 ms (+26 %), 0 fallos en 64 706 peticiones | `summary.json` de k6 por corrida | §5.2 |
| Localizar el punto de quiebre (H4) | Aparece en el nivel 1 000 y en el plano de tiempo real, con 23.87 % de conexiones caídas y establecimiento P95 de ~20 s | Umbrales predefinidos frente a la medición | §5.2 |
| Caracterizar el costo de la capa gratuita | Cuatro restricciones operativas condicionan la viabilidad del sistema y también la del experimento | Bitácora y telemetría del incidente | §5.3, §5.4 |

## b.2 Síntesis de resultados

**Línea base (100 VUs).** 21 596 peticiones HTTP sin fallos (35.3 req/s sostenidas),
408 sesiones WebSocket sin caídas, conexión media de 682 ms y latencia HTTP P95 de
404 ms, con un consumo pico de 20.9 a 44.5 MB por servicio (4.1 % a 8.7 % de los
512 MB). A esta escala la arquitectura opera con holgura de un orden de magnitud.

**Matriz ejecutada (100 y 1 000 VUs, 6 corridas válidas de 6).**

| Métrica | 100 VUs | 1 000 VUs |
|---|---:|---:|
| Peticiones HTTP fallidas | 0.00 % | 0.00 % (0 / 64 706) |
| P95 HTTP | 826 ms (790–847) | 1 045 ms (955–1 138) |
| Conexiones WS caídas | 0.00 % (0–1.92) | **23.87 % (23.59–24.16)** |
| P95 establecimiento WS | 918 ms (894–932) | **19 974 ms (19 817–20 251)** |
| RAM pico api-gateway | 43.6 MB (43.5–43.8) | 238.5 MB (236.8–239.2) |
| RAM pico realtime | 42.8 MB (36.6–50.7) | 231.3 MB (164.1–293.0) |
| RAM pico groups | 51.6 MB (51.0–52.3) | 54.8 MB (49.9–60.4) |
| RAM pico habits | 20.8 MB (20.6–21.0) | 21.5 MB (20.7–21.5) |

*Mediana de 3 repeticiones por nivel; mín–máx entre paréntesis. En negritas, los
valores que cruzan un umbral operacional predefinido (caídas WS < 10 %,
establecimiento WS P95 < 2 s, fallos HTTP < 5 %).*

**Arranque en frío.** Tras la suspensión por inactividad, la primera respuesta de
`/health` tardó entre 10.7 y 13.5 s por servicio, contra 0.3 a 0.6 s en caliente.
El efecto compuesto es peor que el individual, porque mientras un servicio proxeado
despierta el gateway agota su espera y responde 502 en cadena; un cliente que llega en
frío percibe la aplicación caída durante 30 a 60 s.

**Hallazgos operativos del free tier.** Fueron cuatro. La cuota de egreso: la Sesión 1
movió unos 17.8 GB contra los 5 GB mensuales por cuenta y provocó la suspensión de
dos cuentas, y la compresión gzip midió 270 788 a 28 114 bytes, un 89.6 % menos. La
pausa de la capa de datos tras 8 días de inactividad, con restauración manual de unos
2.5 minutos. Los créditos de ráfaga de la instancia de base de datos, donde dos
corridas colapsaron con 87.9 % de caídas, memoria al tope y CPU ociosa, la firma de
una inanición de E/S. Y el modo de *pooling*, que quedó descartado como causa raíz en
este caso pero documentado como multiplicador de riesgo.

## b.3 Discusión: vínculo con el marco teórico y con la hipótesis

La discusión del informe responde cuatro preguntas y ata cada una a la literatura
seleccionada en la Unidad II.

| Pregunta de la discusión | Respuesta | Vínculo con el marco teórico |
|---|---|---|
| ¿Dónde se quiebra primero y por qué? | Por calidad de servicio del canal en vivo. El handshake WS valida membresía contra `groups-service`, de modo que la salud del tiempo real depende de un servicio transaccional y de su capa de datos | Sobri et al. (2022): el *pool* de conexiones a la BD es determinante bajo estrés, y aquí es el eslabón que rompe la cadena |
| ¿La fragmentación en 4 cuentas aísla o propaga fallos? | Hace las dos cosas. Aísla presupuestos (el egreso suspendió 2 cuentas y dejó intactas las otras 2) y propaga por dependencias (un realtime sano entregó 87 % de fallos por un groups bloqueado) | Newman (2021): aislamiento de recursos ≠ aislamiento de fallos |
| ¿Qué margen real ofrece para uso académico? | A 100 conexiones concurrentes, holgura de un orden de magnitud; el umbral se cruza en algún punto entre 100 y 1 000 | Blinowski et al. (2022): línea base de expectativas en hardware restringido |
| ¿Cuántos usuarios reales son? | Unos 3 200 UAD (≈400 grupos llenos), como **estimación condicionada** por Ley de Little con supuestos declarados. Una cota independiente por cuota de egreso (1 100 a 3 700 UAD) converge al mismo orden | Little (1961) |

**Contraste con la hipótesis.** H4, que postulaba un nivel ≤ 5 000 conexiones donde el
sistema incumpliera sus umbrales, queda confirmada. Los dos matices que la precisan
son la parte interesante. El quiebre llegó un orden de magnitud antes de lo esperado,
en el nivel 1 000. Y llegó por una vía distinta a la hipotetizada: calidad de servicio
con la memoria al 46.6 %, en lugar del agotamiento de memoria que se anticipaba. La
elección de Go se sostiene con Fernando y Engel (2025), aunque el estudio muestra que
en esta clase de despliegue el factor limitante resultó ser el acoplamiento
arquitectónico, por encima del lenguaje.

## b.4 Conclusiones y recomendaciones

**Respuesta a la pregunta de investigación.** La arquitectura sí sostiene tráfico
concurrente con tiempo real, con un límite medido. A 100 conexiones concurrentes opera
con holgura; a 1 000, el plano HTTP sigue sin fallar mientras el canal en vivo cruza su
umbral de servicio. El punto de quiebre está por debajo de las 1 000 conexiones y llegó
por calidad de servicio del canal en vivo.

Recomendaciones derivadas de la evidencia, ordenadas por la relación costo/beneficio
que se pudo medir:

1. **Comprimir las respuestas antes de cualquier otra optimización.** Es la única
   mitigación cuyo efecto se midió directamente (89.6 % menos bytes en el cable) y
   ataca la restricción que efectivamente dejó el sistema fuera de línea.
2. **Desacoplar la verificación de membresía del handshake WebSocket**, con caché con
   expiración o token firmado de sala. Es el mecanismo identificado del quiebre.
3. **Tratar el ciclo de la capa de datos como decisión de diseño.** Los créditos de
   ráfaga hacen que el mismo sistema rinda distinto según su historial de consumo, así
   que dejarlo en manos de la operación deja el rendimiento al azar.
4. **Presupuestar el arranque en frío dentro de la experiencia de usuario**, con
   despertar en cascada y pinger externo, porque 30 a 60 s de percepción de caída pesan
   más que 200 ms de P95.

## b.5 Confiabilidad, validez y limitaciones

Esta declaración es requisito del entregable, y se sostiene con la dispersión medida
en lugar de con afirmaciones cualitativas.

### Confiabilidad (repetibilidad de la medición)

Tres repeticiones por nivel, mismo *commit*, misma ventana horaria. El coeficiente de
variación (CV = desviación estándar / media) entre repeticiones quedó así:

| Métrica | CV a 100 VUs | CV a 1 000 VUs | Lectura |
|---|---:|---:|---|
| Conexiones WS caídas | 173.4 % | **1.2 %** | A 100 VUs el CV es artificial, porque la media ronda cero (dos ceros y un 1.92 %). A 1 000 la medición es muy estable |
| P95 establecimiento WS | 2.1 % | 1.1 % | Altamente repetible en ambos niveles |
| P95 HTTP | 3.5 % | 8.8 % | Repetible; la dispersión crece con la carga, como se esperaba |
| RAM pico api-gateway | 0.3 % | 0.5 % | Prácticamente determinista |
| RAM pico groups | 1.3 % | 9.5 % | Estable |
| RAM pico habits | 1.0 % | 2.2 % | Estable |
| RAM pico realtime | 16.3 % | **28.1 %** | La métrica menos repetible del conjunto (164 a 293 MB). Se reporta con su rango y queda fuera de las conclusiones finas |

La conclusión central del estudio descansa en la métrica más estable del conjunto,
las caídas WS a 1 000 VUs con un CV de 1.2 %, lo que descarta el ruido de plataforma
como explicación del quiebre. La métrica más volátil, la RAM de realtime, se usa
únicamente para afirmar el patrón grueso, el crecimiento ×5 en la ruta WS, y queda
fuera de cualquier proyección puntual.

La confiabilidad se apoya además en otros tres soportes: el instrumento versionado con
el *commit hash* del código medido en cada `metadata.json`; el banco, los estadísticos
y las figuras regenerables con un solo comando desde los datos crudos
(`analyze_results.py`), de modo que ninguna cifra depende de una transcripción manual;
y la bitácora de las 11 corridas (7 válidas y 4 excluidas) con la causa asignable de
cada exclusión.

### Validez

**De constructo.** Se mide lo que se dice medir en el plano de recursos y latencia,
con telemetría del servidor e inyector independiente como dos fuentes que se cruzan.
La amenaza reconocida es que todos los usuarios virtuales se autentican con una sola
cuenta, por lo que los patrones de caché de base de datos no representan a miles de
usuarios distintos.

**Interna.** Las condiciones quedaron fijadas y declaradas: mismo *commit*, keep-alive
pausado, servicios despertados antes de cada corrida y 10 minutos de reposo entre
corridas. El confusor más grave, el limitador de tasa por usuario, fue detectado y
eliminado por el piloto antes de medir, y el episodio queda documentado en §4.6. Los
criterios de exclusión se fijaron antes de las corridas y se aplicaron por causa
asignable, nunca por el valor de las métricas.

**De conclusión estadística.** Con n = 3 por nivel no se aplican pruebas de
significancia; se reporta estadística descriptiva (mediana, rango, CV). La distancia
entre lo medido y el umbral es de tal magnitud (23.87 % contra un límite de 10 %, con
rango de 23.59 a 24.16) que la conclusión no depende de un contraste inferencial.

**Externa.** Es un estudio de caso de una combinación concreta, Go con Render y
Supabase. Los hallazgos informan sobre esta clase de arquitectura y no sobre cualquier
PaaS gratuita. El aporte se sostiene en la replicabilidad, que descansa en el artefacto
publicado.

### Limitaciones declaradas

1. **Dos de los cuatro niveles.** Las restricciones operativas del propio free tier
   agotaron la ventana experimental antes de 2 500 y 5 000 VUs. Como el quiebre apareció
   en el nivel 1 000, los niveles superiores habrían caracterizado el comportamiento
   *posterior* al quiebre, incluida la búsqueda del OOM, sin cambiar la conclusión
   sobre su existencia. Queda declarado como trabajo futuro.
2. **Inyector único.** k6 corre desde una sola máquina y red; su ancho de banda podría
   haber sido cuello de botella en los niveles altos que no se ejecutaron.
3. **Usuario único**, la amenaza de constructo descrita arriba.
4. **La traducción a usuarios activos diarios es una estimación.** Depende de supuestos
   de duración de sesión, sesiones por día y concentración en hora pico que no se
   midieron sobre usuarios reales. Variarlos en rangos plausibles la mueve entre 10³ y
   10⁴ UAD por el lado de concurrencia; la cota de egreso, independiente, la regresa al
   orden de 10³.
5. **Ruido de plataforma no eliminable.** Se observaron errores transitorios del borde
   (Cloudflare 502/520) a razón de aproximadamente 1 de cada 40 peticiones incluso en
   reposo.

> Entregable parcial verificable (b): §5, §6, §7 y §8 de `articulo.md`, más el borrador
> previo de la Actividad 3.5 del que derivan.

---

# Subtarea c. Visualización y storytelling de datos

## c.1 Criterio de selección de las visualizaciones

Se generaron dos alternativas por hallazgo, seis figuras en total, y de cada par se
eligió una por la pregunta que responde. Las descartadas se conservan versionadas;
algunas sirven para otro público aunque no funcionen dentro del informe.

| Hallazgo | Pregunta que debe responder la figura | Forma elegida | Alternativa descartada y por qué |
|---|---|---|---|
| H1 · El canal en vivo cruza su umbral | ¿Cuánto se rompe y contra qué límite? | Barras por nivel con línea de umbral QoS y los 3 puntos de repetición superpuestos (`fig_h1_barras`) | Dona de composición (`fig_h1_pastel`). Muestra una sola corrida, así que ni compara niveles ni exhibe dispersión. Se reserva para la infografía, donde el mensaje es «1 de cada 4» |
| H2 · La presión se concentra en la ruta WS | ¿Qué servicio paga la carga y qué tan lejos está del límite? | *Dumbbell* por servicio (100 a 1 000 VUs) con la línea vertical de 512 MB (`fig_h2_dumbbell`) | Barras agrupadas (`fig_h2_barras`). Convierte cuatro servicios en ocho barras y obliga a reconstruir el salto comparando alturas; además la altura de la barra compite con la línea del límite en lugar de apoyarse en ella |
| H3 · La memoria sigue a la rampa y se aplana | ¿El consumo se dispara o se estabiliza? ¿Hay fuga? | Serie de tiempo por servicio con la línea de 512 MB (`fig_h3_linea`) | Barras del pico (`fig_h3_barras`). Reduce la corrida a un punto y borra justamente lo que se quiere mostrar, la forma de la curva y su meseta |

A las tres figuras finales se les aplicaron cuatro reglas de forma. El límite o umbral
siempre queda visible, porque una medición sin su criterio no dice nada. La escala de
magnitud arranca en cero. El valor va etiquetado junto a la marca, sin rejilla densa. Y
los puntos de las repeticiones se dibujan junto a la mediana, para que la dispersión se
vea sin ir a la tabla.

En el *dumbbell* de H2 esa última regla rinde un beneficio extra. Como el eje llega
hasta 512 MB, el espacio vacío a la derecha de los puntos es la holgura de memoria: el
hallazgo, que el sistema falló con más de la mitad del límite libre, se ve sin leer un
solo número.

## c.2 Narrativa: cómo se conectan los hallazgos con el problema

El informe cuenta una sola historia en cinco movimientos. Cada figura entra en el
movimiento donde hace falta y no antes; ninguna se muestra «porque ya la teníamos».

| # | Movimiento | Frase que lo sostiene | Apoyo visual |
|---|---|---|---|
| 1 | Contexto: la barrera | Los proyectos académicos no llegan a producción porque la infraestructura cuesta; el free tier existe pero se presume que solo sirve para demos | Diagrama de topología (§3) |
| 2 | Conflicto: la apuesta | Cuatro microservicios en cuatro cuentas gratuitas, con WebSockets propios *a propósito*, para que el costo del tiempo real quede dentro de la medición | Tabla de decisiones condicionadas por el free tier (§3.2) |
| 3 | Falsa pista: lo que casi medimos mal | El primer piloto reportó 88 % de fallos mientras la telemetría del servidor mostraba consumo mínimo. La causa estaba en nuestro propio limitador de tasa | Tabla piloto 1 frente a piloto 2 (§4.6) |
| 4 | Giro: el quiebre no está donde se esperaba | Se buscaba un OOM y apareció un colapso de calidad de servicio con la memoria al 46.6 % | Figura 1 (caídas frente al umbral) y Figura 2 (RAM frente a 512 MB), juntas: el contraste entre ambas *es* el hallazgo |
| 5 | Resolución: el veredicto con su precio | Sirve hasta cientos de usuarios concurrentes, y la factura llega por cuotas, pausas y créditos antes que por RAM | Figura 3 (la meseta) y §5.4 |

**Regla de una frase por figura.** El pie de cada figura afirma algo. Describir los
ejes es trabajo de la propia figura.

- Figura 1: «A 1 000 conexiones concurrentes se cae ~1 de cada 4 conexiones, más del
  doble del umbral, con dispersión mínima entre repeticiones.»
- Figura 2: «El costo se concentra en gateway y realtime (×5) mientras los servicios
  transaccionales permanecen planos, y aun así el peor se queda a menos de la mitad del
  límite.»
- Figura 3: «El consumo sigue a la rampa y se aplana en la meseta, señal de techo de
  carga y no de fuga de memoria.»

El par Figura 1 y Figura 2 es el núcleo narrativo, porque juntas dicen que el sistema
falla mientras la memoria sobra. Deben verse en la misma página del informe y en la
misma lámina de la presentación.

## c.3 Cuadro de mando con jerarquía de indicadores

Tres niveles de lectura, para que la misma evidencia sirva a quien quiere el veredicto
en cinco segundos y a quien quiere la causa raíz.

**Nivel 1, veredicto (¿cumple o no cumple?).** Un solo indicador de titular.

| Indicador | Umbral | 100 VUs | 1 000 VUs | Estado |
|---|---|---:|---:|---|
| Conexiones WebSocket caídas | < 10 % | 0.00 % | 23.87 % | 🟢 → 🔴 |

**Nivel 2, diagnóstico (¿qué plano se rompió?).** Separa tiempo real, HTTP y recursos.

| Plano | Indicador | Umbral | 100 VUs | 1 000 VUs | Estado |
|---|---|---|---:|---:|---|
| Tiempo real | P95 de establecimiento WS | < 2 s | 918 ms | 19 974 ms | 🔴 |
| HTTP | Peticiones fallidas | < 5 % | 0.00 % | 0.00 % | 🟢 |
| HTTP | P95 de latencia | (sin umbral; se observa) | 826 ms | 1 045 ms | 🟡 +26 % |
| Recursos | RAM del peor servicio, % de 512 MB | < 90 % | 10.1 % | 46.6 % | 🟢 |

**Nivel 3, causa raíz y operación (¿por qué, y qué lo condiciona?).**

| Dimensión | Indicador | Fuente | Valor observado |
|---|---|---|---|
| Acoplamiento | Dependencia del handshake WS con `groups-service` | Diseño e incidente de la Sesión 2 | Un realtime sano entregó 87.9 % de fallos por un groups bloqueado |
| Cómputo | CPU del gateway durante la corrida | `/metrics` | 100 % de su asignación (0.1 vCPU) en Sesión 1 |
| Capa de datos | Saturación del *pool* de conexiones | `/metrics` (pgx) | 10/10 en uso, cientos de esperas en el colapso |
| Cuota | Egreso por corrida frente a 5 GB/mes por cuenta | Panel de Render | ~3.03 GB por corrida, ~0.3 GB con gzip (89.6 % menos) |
| Disponibilidad | Arranque en frío de `/health` | Medición instrumentada | 10.7 a 13.5 s por servicio; 30 a 60 s de efecto compuesto |
| Estado de la BD | Créditos de ráfaga disponibles | Diagnóstico del incidente | Firma de inanición de E/S, con RAM al tope y CPU ociosa |

El cuadro se usa así: el Nivel 1 es lo que se responde en la defensa antes de cualquier
matiz; el Nivel 2 es la lámina de resultados de la presentación y la Tabla 3 del
informe; el Nivel 3 solo aparece si preguntan por el mecanismo. Ese último nivel es
donde se nota el valor de haber instrumentado telemetría propia, porque ninguno de sus
seis indicadores es observable desde el cliente.

## c.4 Producción y regeneración de las figuras

Las figuras se generan desde los datos crudos versionados con un único comando, de modo
que cualquier corrección de datos se propaga sola. Ninguna se edita a mano.

```sh
docker run --rm -v "$(pwd)/load-tests":/lt -w /lt python:3.12-slim \
  sh -c "pip install -q matplotlib && python analyze_results.py"
```

Las salidas quedan en `load-tests/analysis/`: `banco_corridas.csv` (el banco de la
Act. 3.4), `stats.json` (descriptivos y frecuencias) y las seis figuras `fig_*.png`.
Las tres elegidas se insertan en el informe como Figuras 1, 2 y 3.

Durante esta actividad se detectó y corrigió un defecto. En `fig_h1_pastel.png` la
leyenda se recortaba por el borde derecho del lienzo, porque `tight_layout` no
contempla una leyenda anclada fuera de los ejes. Esa figura no entra al informe, pero
sí se reutiliza en la infografía de la Actividad 4.2b, así que se corrigió en
`analyze_results.py` exportándola con `bbox_inches="tight"` y se regeneraron las seis
salidas.

> Entregable parcial verificable (c): las seis figuras en `load-tests/analysis/`, las
> tres seleccionadas ya insertadas en `articulo.md` §5, y este cuadro de mando.

---

# Cierre: integración del informe final

Con las tres subtareas completadas, el informe final integrado es
`Documentos/articulo.md`, de unas 6 páginas a dos columnas, con 10 secciones, 3
figuras, 4 tablas y 11 referencias en APA 7. Quedan dos pendientes editoriales
declarados en el propio documento: la traducción del resumen al inglés si la sede lo
exige, y la decisión sobre publicar el repositorio o un espejo para revisión.

**Trazabilidad con las actividades anteriores.** La Unidad I aportó pregunta, objetivos
e hipótesis. La Unidad II, el marco teórico y la matriz de fuentes (Act. 3.1). La
Unidad III aportó el instrumento (3.2), la recolección (3.3), el banco con las
visualizaciones (3.4) y el borrador de resultados (3.5). Esta actividad los integra en
un solo documento y define cómo se comunican; la Actividad 4.2 los lleva a la defensa
oral y a un público no académico.
