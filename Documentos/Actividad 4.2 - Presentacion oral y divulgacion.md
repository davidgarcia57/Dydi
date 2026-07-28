---
title: "Actividad 4.2. Comunicación pública del proyecto: presentación oral y divulgación"
subtitle: "Proyecto Dydi · Equipo 3 · Universidad Tecnológica de Durango · Integradora 2026"
date: "28 de julio de 2026"
lang: es
---

PROYECTO DYDI · UTD · INTEGRADORA 2026 · Modalidad: equipo · Equipo 3, Resiliencia y
consumo de recursos en microservicios.

> **Qué entrega esta actividad.** (a) La preparación completa de la presentación oral,
> con su tipo, objetivos, estructura, guion cronometrado, reparto de voces y simulacro
> de preguntas, junto con el apoyo visual ya construido. (b) Un producto de divulgación
> dirigido a un público distinto al de la defensa, con su ficha de adaptación,
> tratamiento de datos y declaración de limitaciones.
>
> Las dos subtareas parten del mismo hallazgo central y solo cambian el lenguaje y el
> nivel de detalle, como pide la indicación.

## El hallazgo central (común a los dos públicos)

> **Una arquitectura de microservicios repartida en cuatro cuentas gratuitas sí
> sostiene tráfico real, pero se rompe antes y por otra razón de la que se esperaba: a
> 1 000 conexiones concurrentes se cae una de cada cuatro conexiones en vivo mientras
> la memoria del servicio más cargado va apenas al 46.6 % de su límite. Lo que lo tumba
> es el acoplamiento entre servicios, con recursos de sobra.**

Todo lo que sigue es ese enunciado dicho de dos maneras.

| | Defensa oral (4.2a) | Divulgación (4.2b) |
|---|---|---|
| Público | Profesora y jurado académico | Estudiantes y equipos pequeños que quieren desplegar sin presupuesto |
| Qué le importa | Que el método aguante preguntas | Si le sirve a *su* proyecto y qué le va a doler |
| Nivel de detalle | Umbrales, dispersión, criterios de exclusión, amenazas a la validez | Un número grande, tres límites prácticos y una advertencia honesta |
| Cómo mide el éxito | Que ninguna cifra quede sin respaldo | Que alguien tome una decisión mejor informada en 60 segundos |
| Formato | Exposición de 13 min más preguntas, con deck proyectado | Infografía de una página, imprimible y compartible |

---

# Subtarea a. Presentación oral

## a.1 Tipo de presentación y justificación

Tipo elegido: **ponencia académica breve en formato de simulación de defensa**, con 13
minutos de exposición y entre 5 y 8 de preguntas, repartida entre las cuatro voces del
equipo según el rol que cada quien tuvo en el experimento.

Las otras opciones del planteamiento se descartaron con estos argumentos:

- *Pitch técnico o demo del producto.* Dydi es el objeto medido. Una demo de la app
  respondería «¿funciona la aplicación?» cuando la pregunta evaluada es «¿aguanta la
  infraestructura, y cómo lo sabemos?». La app aparece 40 segundos, lo justo para que se
  entienda qué genera la carga.
- *Presentación interna.* El público no comparte el contexto del equipo, así que cada
  término (VU, P95, OOM, free tier) tiene que ganarse su lugar en la primera mención.

Hay una decisión de fondo sobre el formato: la exposición se organiza alrededor del
giro del experimento. Contar once corridas en orden cronológico consumiría el tiempo
sin llegar a lo interesante, que es que buscábamos un OOM y encontramos otra cosa. El
guion se apoya en el *storytelling* de la Actividad 4.1 §c.2.

## a.2 Objetivos de la presentación

Los tres objetivos están redactados como lo que la audiencia debe poder hacer al
terminar:

1. **Reproducir el hallazgo central en una frase**, incluido el matiz de que el quiebre
   llegó por calidad de servicio y con memoria de sobra.
2. **Reconocer que el método resiste el escrutinio**, por las repeticiones, los
   umbrales predefinidos, las exclusiones por causa asignable y un piloto que corrigió
   al propio instrumento.
3. **Distinguir lo medido de lo estimado.** Los niveles de 100 y 1 000 VUs son
   mediciones; los ~3 200 usuarios diarios son una estimación con supuestos declarados.

Si al terminar alguien pudiera confundir el punto 3, la presentación falló aunque todo
lo demás salga bien.

## a.3 Estructura, tiempos y reparto de voces

Objetivo: 13:00. Tope duro: 15:00. Cada quien defiende la parte que ejecutó.

| # | Lámina | Mensaje único de la lámina | Tiempo | Voz |
|---:|---|---|---:|---|
| 1 | Portada | Quiénes somos y qué pregunta respondimos | 0:20 | David |
| 2 | La barrera | Los proyectos académicos no llegan a producción porque la infraestructura cuesta | 0:50 | David |
| 3 | Dydi y su arquitectura | Cuatro servicios en Go, cuatro cuentas gratuitas, WebSockets propios | 1:20 | Irvin |
| 4 | La apuesta metodológica | El tiempo real es *nuestro* a propósito; si lo delegamos, sale del experimento | 0:50 | Irvin |
| 5 | Cómo se midió | k6 por fuera, telemetría propia por dentro, 4 niveles × 3 repeticiones | 1:20 | Keila |
| 6 | El piloto que casi nos engaña | 88 % de fallos causados por nuestro propio limitador de tasa | 1:10 | David |
| 7 | Hallazgo 1: el quiebre | A 1 000 VUs se cae 1 de cada 4 conexiones, más del doble del umbral | 1:30 | Keila |
| 8 | Hallazgo 2: la memoria sobra | El peor servicio va al 46.6 % del límite y aun así el sistema falla | 1:00 | Keila |
| 9 | El mecanismo | El handshake en vivo depende de un servicio transaccional | 1:10 | Juan David |
| 10 | La factura del free tier | Cuotas, pausas y créditos de ráfaga tumbaron el propio experimento | 1:00 | Irvin |
| 11 | Traducción a usuarios | ~3 200 usuarios diarios, como estimación declarada | 0:50 | Juan David |
| 12 | Conclusiones y límites | Sirve, con límite medido, y esto es lo que no podemos afirmar | 1:20 | David |
| 13 | Cierre | Todo el arnés y los datos crudos son públicos y reproducibles | 0:20 | David |

Las láminas 7 y 8 van juntas y sin pausa, porque el contraste entre ambas es el
hallazgo. Si el tiempo se va, se recortan la 3 y la 11, en ese orden. Las láminas 7, 8
y 12 no se recortan.

## a.4 Guion

Marcas de tiempo acumuladas. El texto es guía; nadie lo lee en voz alta.

**[0:00 · L1, Portada · David]**
«Buenas tardes. Somos el Equipo 3. Durante seis semanas intentamos responder una
pregunta incómoda: ¿se puede poner en producción una arquitectura de microservicios en
tiempo real sin pagar un peso de infraestructura? La pregunta no era si *arranca*, eso
lo sabe cualquiera, sino hasta dónde aguanta y por dónde se rompe.»

**[0:20 · L2, La barrera · David]**
«El problema de fondo es económico. Los proyectos de la carrera casi nunca llegan a
producción porque un servidor cuesta. Las capas gratuitas existen, Render, Railway,
Fly, pero tienen fama de servir solo para demos: 512 MB de RAM, el servicio se duerme a
los 15 minutos y horas de cómputo contadas. Nuestra sospecha era que esa fama estaba
repetida pero no medida. Así que la medimos.»

**[1:10 · L3, Dydi y su arquitectura · Irvin]**
«El sistema bajo prueba es Dydi, una app real. Grupos de amigos, máximo ocho, que se
comprometen a hábitos diarios; el que falla entra a una ruleta de penitencias el
sábado, y eso se transmite en vivo a todo el grupo. Por dentro son cuatro
microservicios en Go: el gateway, que valida el token y es la única puerta; grupos;
hábitos; y el de tiempo real, que sostiene los WebSockets. Cada uno vive en una cuenta
gratuita distinta de Render, una por integrante. Así compusimos el sistema con los
recursos gratuitos que legítimamente teníamos.»

**[2:30 · L4, La apuesta metodológica · Irvin]**
«Una decisión que va a parecer rara. El tiempo real lo programamos nosotros, con
WebSockets propios, pudiendo haber usado el servicio administrado de Supabase. Fue a
propósito. Lo que queremos medir es cuánto cuesta sostener conexiones vivas dentro de
512 MB; si delegamos esa pieza, se sale del sistema bajo prueba y ya no medimos nada.
Es la variable central del experimento.»

**[3:20 · L5, Cómo se midió · Keila]**
«Dos fuentes independientes. Por fuera, k6 inyecta carga con rampas de 100, 1 000,
2 500 y 5 000 usuarios virtuales, y tres repeticiones por nivel para no confundir un
mal día con un resultado. Por dentro, telemetría propia: unas doscientas líneas de Go
por servicio que exponen memoria, latencia por ruta y estado del pool de base de datos,
que un *scraper* muestrea cada cinco segundos. Y una regla que nos sirvió mucho: si un
servicio muere, el hueco en la serie también es dato. Cada corrida deja cuatro
archivos, incluido el *commit* exacto del código que se midió.»

**[4:40 · L6, El piloto que casi nos engaña · Keila]**
«Aquí está nuestro mejor aprendizaje, y viene de un tropiezo. El primer piloto dio 88 %
de peticiones fallidas y 94 % de conexiones rechazadas. La conclusión fácil habría sido
“el free tier no aguanta, listo”. Pero la telemetría del servidor mostraba 44 MB de
RAM, o sea que el sistema estaba *aburrido*. Cruzando las dos fuentes encontramos la
causa: nuestro propio gateway limita cinco peticiones por segundo por usuario, y los
mil usuarios virtuales usaban la misma cuenta de prueba. Estábamos midiendo el
limitador. Lo hicimos configurable, repetimos el piloto y bajó a 0 % de fallos. Sin
telemetría del lado del servidor habríamos publicado una mentira con gráficas
bonitas.»

**[5:50 · L7, Hallazgo 1: el quiebre · Keila]**
«Este es el resultado. A 100 usuarios concurrentes, cero fallos y cero conexiones
caídas. A 1 000, el plano HTTP sigue sin fallar, con cero errores en 64 706 peticiones y
apenas 26 % más lento, pero el canal en vivo se rompe. Se cae el 23.87 % de las
conexiones, una de cada cuatro. Nuestro umbral, fijado *antes* de correr, era 10 %. Y
establecer una conexión pasó de novecientos milisegundos a veinte segundos. Miren la
dispersión: 23.59, 23.87 y 24.16 por ciento en las tres repeticiones. Con esa
consistencia, la explicación del mal día de la plataforma queda descartada. Es un
límite del sistema.»

**[7:20 · L8, Hallazgo 2: la memoria sobra · Keila]**
«Y aquí está lo que no esperábamos. Nuestra hipótesis decía que moriría por falta de
memoria. Cuando el sistema incumplió sus umbrales, el servicio más cargado usaba 238 de
512 MB, el 46.6 %. Falló con la mitad de la memoria libre. Además la curva en el tiempo
sigue a la rampa de conexiones y se aplana en la meseta, o sea que hay techo de carga y
no fuga.»

**[8:20 · L9, El mecanismo · Juan David]**
«¿Por qué se cae entonces? Por una decisión de seguridad nuestra. Cada vez que alguien
abre una conexión en vivo, el servicio de tiempo real le pregunta al de grupos “¿esta
persona pertenece a esta sala?”. Está bien hecho, cierra el sistema por defecto, pero
ata la salud del canal en vivo a la latencia de un servicio transaccional y de su base
de datos. Cuando esa ruta se congestiona, las conexiones caen; los clientes reintentan;
cada reintento es un handshake nuevo sobre un gateway que ya estaba al 100 % de su CPU.
Se convierte en tormenta. El cuello de botella está en cómo amarramos las piezas.»

**[9:30 · L10, La factura del free tier · Irvin]**
«Hay una segunda capa de resultados que no esperábamos reportar, y son las reglas de
operación. Nuestras corridas movieron 17.8 GB cuando la cuota es de 5 GB al mes por
cuenta, y Render nos suspendió dos cuentas. Ocho días después, Supabase pausó la base
de datos por inactividad y todo respondía errores estando sano. Y cuando la
restauramos, dos corridas colapsaron con una firma rarísima: memoria al tope y CPU
ociosa. Eran los créditos de ráfaga de la instancia gratuita, agotados. Esas dos
corridas están excluidas del dataset por criterio predefinido, pero siguen publicadas,
y una de ellas terminó convertida en hallazgo. La conclusión práctica es que la capa
gratuita te limita por cuotas antes que por CPU. La compresión que activamos bajó el
tráfico 89.6 % medido.»

**[10:30 · L11, Traducción a usuarios · Juan David]**
«La pregunta que todos hacen: ¿cuántos usuarios reales son? Aplicamos la Ley de Little
sobre la concurrencia que sí medimos, con supuestos que declaramos: sesiones de cinco
minutos, una y media al día, 25 % en hora pico. Salen unos 3 200 usuarios activos
diarios, unos 400 grupos llenos. Y por un camino independiente, la cuota de
transferencia da entre 1 100 y 3 700. Que dos límites distintos converjan en el mismo
orden de magnitud da confianza. Pero lo decimos claro: esto es una estimación con
supuestos y no un resultado medido.»

**[11:20 · L12, Conclusiones y límites · David]**
«Respuesta a la pregunta: sí aguanta, con límite medido. Cientos de conexiones
concurrentes con holgura de un orden de magnitud; a mil, el canal en vivo cruza su
umbral. El quiebre llegó diez veces antes de lo que esperábamos y por acoplamiento
entre servicios. Ahora, lo que no podemos afirmar. No corrimos 2 500 ni 5 000, porque la
propia capa gratuita agotó nuestra ventana. Todos los usuarios virtuales usaron una
sola cuenta. Y el inyector fue una sola máquina. Los tres quedan declarados como
trabajo futuro en el artículo.»

**[12:40 · L13, Cierre · David]**
«Todo es reproducible: el arnés, los datos crudos de las once corridas con su *commit*,
la bitácora y el script que regenera banco, estadísticas y figuras con un comando.
Gracias. Quedamos a sus preguntas.»

## a.5 Lenguaje accesible: reglas de traducción

Cada término técnico se dice una vez en llano antes de usar la sigla. Nadie usa una
sigla que no haya presentado.

| No decir | Decir |
|---|---|
| «El P95 de latencia es 1 045 ms» | «El 95 % de las peticiones respondió en menos de un segundo» |
| «23.87 % de *drop rate* en WS» | «Se cae una de cada cuatro conexiones en vivo» |
| «No hubo OOM kill» | «El sistema nunca se quedó sin memoria; iba a la mitad» |
| «Rampa de 1 000 VUs» | «Mil usuarios simulados conectándose al mismo tiempo» |
| «*Cold start* de 13.5 s» | «El servicio dormido tarda trece segundos en despertar» |
| «Coeficiente de variación de 1.2 %» | «Las tres repeticiones dieron casi el mismo número» |

## a.6 Apoyos visuales

El deck es `Documentos/presentacion/dydi-defensa.html`, un solo archivo HTML
autocontenido con las figuras incrustadas y sin dependencias de red, con la paleta y
las tipografías de Dydi. Se abre en cualquier navegador y se proyecta a pantalla
completa.

Se navega con `→` o espacio para avanzar, `←` para retroceder, `F` para pantalla
completa y `P` para imprimir o exportar a PDF, que es el respaldo si falla el proyector.

La regla de diseño aplicada es una idea por lámina y un número grande por lámina. El
texto de la lámina no repite el guion y no hay viñetas para leer en voz alta. Las
Figuras 1 y 2 (caídas frente al umbral, RAM frente a 512 MB) se muestran a sangre
completa porque son el núcleo narrativo.

## a.7 Ensayo, contingencias y criterios de calidad

Se harán dos pasadas completas cronometradas antes de la fecha, la segunda con público
ajeno al equipo. Si sobra contenido se corta contenido; nadie habla más rápido. Cada
quien ensaya la transición hacia la voz siguiente, que es donde se pierde tiempo.

| Riesgo | Plan |
|---|---|
| Falla el proyector o el equipo | PDF del deck en llave USB y en el correo de dos integrantes |
| No hay internet | El deck es autocontenido y ninguna lámina depende de la red |
| Falta un integrante | Cada lámina tiene voz suplente, la del bloque contiguo |
| Se agota el tiempo | Saltar L3 y L11, marcadas en el deck. L7, L8 y L12 se mantienen siempre |
| Preguntan por un número que no está en el deck | Está en el banco, `Actividad 3.4 - Banco de datos (corridas).csv`, abierto en una pestaña |

## a.8 Simulacro de preguntas

Cada pregunta tiene un responsable de primera respuesta. La regla del equipo: si un
número no se puede rastrear hasta un artefacto, se responde «lo verificamos y le
respondemos». Nadie improvisa cifras.

| Pregunta previsible | Responde | Respuesta corta |
|---|---|---|
| ¿Por qué no usaron Supabase Realtime? | Irvin | Porque es la variable del experimento; delegarla saca del sistema bajo prueba justo lo que se quiere medir |
| ¿De dónde sale el 23.87 %? | Keila | Es la mediana de las tres repeticiones (23.59 / 23.87 / 24.16), medida por k6 y guardada en el `summary.json` de cada corrida. Se regenera con un comando |
| Excluyeron corridas, ¿no es escoger datos convenientes? | Juan David | El criterio estaba predefinido en el protocolo y la exclusión fue por causa asignable, nunca por el valor de la métrica. Las corridas excluidas siguen publicadas con su evidencia |
| ¿Por qué solo dos de los cuatro niveles? | David | Las reglas operativas del free tier agotaron la ventana, y documentarlo ya es un resultado. El quiebre había aparecido en el segundo nivel |
| ¿El primer piloto no prueba que medían mal? | David | El piloto hizo justo su trabajo, que es validar el instrumento antes de medir. Ese episodio es nuestro aporte metodológico |
| ¿Por qué medianas y no promedios? | Keila | Por los «vecinos ruidosos» de la capa gratuita. La mediana con rango resiste atípicos y muestra dispersión |
| ¿Por qué no llegó el OOM que esperaban? | Juan David | Porque la calidad del canal en vivo se degrada antes. La memoria crece proporcional a las conexiones, sin fuga, y proyectamos el OOM entre 2 500 y 5 000 |
| ¿Esto lo puede replicar alguien más? | Keila | Sí. Arnés, datos crudos de las once corridas con su *commit*, bitácora y script de regeneración, todo versionado |
| ¿Qué cambiarían para aguantar más? | Irvin | Aliviar la verificación de membresía por conexión, evaluar mitigaciones con corridas A/B y comparar contra un contenedor de pago único |

> Entregable parcial verificable (a): este guion con tiempos y reparto, más el deck
> `Documentos/presentacion/dydi-defensa.html` ya construido.

---

# Subtarea b. Producto de divulgación

## b.1 Producto elegido y justificación

Producto: **infografía de una página**, en
`Documentos/divulgacion/dydi-infografia.html`, autocontenida, imprimible en A4 y
exportable a PDF o PNG desde el navegador.

La indicación exige un público distinto al de la presentación oral, y ese requisito
descarta las otras opciones. El cartel científico se dirige al mismo público académico
en otro soporte. El resumen ejecutivo presupone un decisor con presupuesto, figura que
no existe en el escenario que el trabajo estudia. La infografía permite hablarle a
quien no va a leer un artículo de seis páginas ni asistir a una defensa, y que aun así
tiene que decidir en un minuto si esta ruta le sirve.

## b.2 Público objetivo y canal

El público son estudiantes de ingeniería y equipos pequeños (proyectos de titulación,
hackathones, productos en validación temprana) que tienen que desplegar algo real sin
presupuesto de infraestructura.

Sabe programar, ha oído «microservicios» y «WebSocket» y probablemente ya usó un free
tier para una demo. No sabe qué es un percentil 95, un OOM kill ni la Ley de Little. Lo
que quiere es saber si le va a servir y qué se le va a romper primero.

El canal es la publicación en comunidades técnicas y redes profesionales del equipo,
más impresión A4 para el pizarrón del laboratorio. El enlace al artículo completo y al
repositorio va en el pie de la pieza.

## b.3 Mensaje y arquitectura de la pieza

Mensaje central, la frase que debe quedar: *la capa gratuita aguanta más de lo que dice
su fama, se rompe por un lado inesperado y te limita por cuotas antes que por CPU.*

Jerarquía visual de arriba hacia abajo, pensada para lectura en cascada:

| Bloque | Contenido | Función |
|---|---|---|
| 1 · Titular | «Medimos hasta dónde aguanta gratis» y una línea de contexto | Gancho |
| 2 · Los tres números | 100 conexiones con holgura · 1 de cada 4 caídas a 1 000 · 46.6 % de memoria usada al fallar | El hallazgo, legible en 5 segundos |
| 3 · El giro | «Esperábamos que muriera de memoria. Se murió de acoplamiento», con las dos figuras enfrentadas | La idea que se comparte |
| 4 · Lo que de verdad te va a doler | Cuota de transferencia · el servicio se duerme · la base de datos se pausa y tiene créditos | Utilidad práctica |
| 5 · Qué haríamos distinto | Comprimir · desacoplar la validación del handshake · presupuestar el arranque en frío | Accionable |
| 6 · Letra chica honesta | Qué se midió, qué se estimó y qué no se probó | Transparencia (§b.5) |
| 7 · Pie | Autores, institución, enlace al artículo y al repositorio | Atribución |

## b.4 Adaptación del lenguaje

Misma evidencia, otro registro. La comparación con el informe es explícita:

| En el informe académico | En la infografía |
|---|---|
| «Conexiones WS caídas: 23.87 % (23.59–24.16), mediana de 3 repeticiones» | «Se cae 1 de cada 4 conexiones» |
| «El servicio más cargado alcanzó el 46.6 % del límite de 512 MB» | «Le sobraba la mitad de la memoria cuando falló» |
| «P95 de establecimiento WS: 19 974 ms» | «Conectarse pasó de 1 a 20 segundos» |
| «Cuota de egreso de 5 GB/mes por cuenta; la Sesión 1 movió ~17.8 GB» | «Nos suspendieron dos cuentas por pasarnos de datos. La prueba tumbó al experimento» |
| «Estimación condicionada por Ley de Little bajo supuestos declarados» | «Nuestra cuenta: alcanza para ~3 200 personas al día, y es una estimación» |
| «Amenaza de constructo: carga de un solo usuario» | «Todos nuestros usuarios simulados usaban la misma cuenta, así que la base de datos la tuvo más fácil que en la vida real» |

De la pieza se omiten a propósito la matriz experimental completa, los criterios de
exclusión, el marco teórico y las referencias. Su ausencia no cambia la decisión que
este público va a tomar, y el enlace al artículo completo queda en el pie para quien
los necesite.

## b.5 Transparencia, datos sensibles y limitaciones

Requisito explícito del entregable, tratado en tres frentes.

**Datos personales.** No hay, y la pieza lo dice. Toda la carga fue sintética: usuarios
virtuales de k6 contra una cuenta de pruebas y grupos sembrados por el propio arnés. No
participaron personas, no se recolectaron datos de usuarios reales y el estudio no
requirió consentimiento informado, porque su unidad de análisis es la transacción. La
pieza lo declara en una línea para que nadie asuma que hubo observación de usuarios.

**Datos sensibles del sistema.** Se excluyen por diseño. Ni la infografía ni el deck
contienen tokens, llaves, cadenas de conexión, URLs internas de los servicios ni el
token de *keep-alive*. Es la misma regla del proyecto: los secretos van por variable de
entorno y nunca a un documento. Las capturas y figuras se revisan antes de exportar.

**Limitaciones.** La pieza las declara en su propia superficie, con el mismo peso
visual que los resultados:

1. Se probaron dos niveles de carga, 100 y 1 000, de los cuatro planeados.
2. Los ~3 200 usuarios diarios son una estimación con supuestos declarados, marcada
   visualmente distinta de los números medidos.
3. Es un caso concreto (Go con Render y Supabase), no un veredicto sobre toda capa
   gratuita.
4. Todos los usuarios simulados compartían una cuenta.

**Atribución y verificabilidad.** La pieza lleva autores, institución, fecha y enlace al
artículo y al repositorio con los datos crudos, de modo que cualquiera puede regenerar
las cifras por su cuenta.

## b.6 Especificación técnica del entregable

Un archivo HTML autocontenido, sin dependencias de red, con las figuras incrustadas en
base64 y tipografías con alternativas del sistema, más una hoja de estilo de impresión
para A4 vertical. Para exportarlo se abre en el navegador, se manda a imprimir y se
elige «Guardar como PDF» con los gráficos de fondo activados.

> Entregable parcial verificable (b): `Documentos/divulgacion/dydi-infografia.html` más
> esta ficha de adaptación.

---

# Cierre

Las dos subtareas parten del mismo hallazgo y divergen en registro. La defensa lo somete
al escrutinio metodológico; la infografía lo pone a disposición de quien tiene que
decidir dónde desplegar mañana. Las dos declaran las mismas limitaciones, la primera
con la terminología de la validez y la segunda en lenguaje llano, sin recortar ninguna.
