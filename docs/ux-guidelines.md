# Guías de UX de Dydi (con fuente)

Este documento existe porque el APK "se sentía al azar" y no queríamos arreglarlo
a gusto. Cada regla de aquí viene de una fuente normativa o de investigación
empírica, con su URL, y **es la razón por la que el código está como está**. Si
vas a cambiar distribución o jerarquía en el móvil, discute contra esto.

Fuentes consultadas (136 reglas extraídas, 88 de prioridad alta):

| Fuente | Qué autoridad tiene | Base |
|---|---|---|
| **W3C — WCAG 2.2 + notas de la WAI** | Norma. Los criterios A/AA son exigibles | <https://www.w3.org/TR/WCAG22/> |
| **Material Design 3** | Guía de la plataforma donde corre el APK | <https://m3.material.io> |
| **Apple HIG** | La referencia más madura de jerarquía móvil | <https://developer.apple.com/design/human-interface-guidelines> |
| **Nielsen Norman Group** | Investigación empírica de usabilidad | <https://www.nngroup.com> |

## Números que no se negocian

| Qué | Valor | Fuente |
|---|---|---|
| Área tocable | **48×48 dp** | M3 pide 48, Apple 44 pt, WCAG 2.5.8 mínimo 24. Con 48 cumples las tres |
| Separación entre targets | **≥8 dp** (12 dp si tienen fondo/bisel) | M3 + Apple HIG |
| Márgenes laterales | **16 dp** en compact (todo teléfono en vertical) | M3 Layout |
| Escala de espaciado | múltiplos de **8** (con 2/4/6/10 permitidos) | M3, base `space100 = 8dp` |
| Piso tipográfico | **12 px** | Ver `mobile/tailwind.config.js` |
| Texto de apoyo | máximo **3 líneas** | M3 Lists |
| Destinos en la barra | **3 a 5**, cinco es el techo | M3: *"Navigation bars provide access to three to five destinations"* |
| Sub-tabs dentro de una pantalla | máximo **4** | M3: *"At five or more tabs, the container becomes cramped"* |
| Elementos "grandes" por pantalla | máximo **2**, y máximo **3** tamaños | NN/g jerarquía visual |
| Indicadores de espera | <200 ms nada · 200 ms–5 s spinner · >5 s progreso explícito | M3. Cruza con el cold start de Render |
| Reflow | debe funcionar a **320 px** de ancho | WCAG 1.4.10 (AA) |
| Interlineado forzado | el layout sobrevive a **1.5×** | WCAG 1.4.12 (AA) — mata las alturas fijas |

## Distribución y jerarquía

**La acción principal no va bajo el pliegue.** Apple HIG (Layout): *"place the
most important items near the top and leading side"*. W3C Mobile Accessibility
Note 4.3 lo pide explícitamente. Y NN/g lo midió: *"the 100 pixels just above the
fold were viewed **102% more** than the 100 pixels just below"*, con una
diferencia promedio del **84%** en cómo se trata la información arriba vs. abajo.

→ Por eso `app/(tabs)/index.tsx` va en el orden **tu turno → tu riesgo → el squad
→ countdown**, y no al revés.

**En pantalla chica va menos información, no la misma comprimida.** W3C Mobile
Accessibility Note 2.1: *"fewer content modules, fewer images"*. NN/g: lo que en
un monitor de 30" es una pantalla, en un teléfono de 4" son **cinco**.

**Cada bloque extra le quita visibilidad al importante.** NN/g H8: *"Every extra
unit of information in an interface competes with the relevant units of
information and diminishes their relative visibility"*. Es la explicación técnica
de "todo se siente al azar": si todo contrasta, nada destaca.

**Cuidado con el falso piso.** NN/g llama *illusion of completeness* al efecto de
un bloque grande con divisor de ancho completo: la gente cree que el contenido
terminó ahí y no hace scroll.

**Una tarea primaria por pantalla, y debe ser el elemento más grande.** M3:
*"The most important action or the main call to action should be the largest
element"*. Apple HIG: máximo uno o dos botones prominentes por vista.

**No metas todo en tarjetas.** M3: *"Don't force content into cards when spacing,
headlines, or dividers would create a simpler visual hierarchy"*. Una card es de
un solo tema.

## Navegación

**Una tab es un lugar, no una acción.** Apple HIG: *"Use a tab bar to support
navigation, **not to provide actions**"*. M3: *"Navigation bars shouldn't be used
for accessing single tasks"*.

→ Por eso la tab "Votar" (un verbo) pasó a llamarse **"Propuestas"**. Además la
pantalla ya se titulaba así, y esa contradicción entre barra y encabezado es un
fallo de **WCAG 3.2.4 Consistent Identification** (AA).

**Las tabs no modelan secuencias.** M3: *"Use tabs to group related content, **not
sequential content**"*. Un ciclo de vida (catálogo → propuesta → historial) no es
contenido relacionado: es un flujo.

**La navegación principal se queda visible.** NN/g midió >20% menos
descubribilidad con navegación escondida, y usuarios 15% más lentos en móvil. La
barra inferior de Dydi es correcta: no migrar a hamburguesa.

**Los iconos de navegación llevan etiqueta siempre visible.** NN/g: *"a word is
worth a thousand pictures"*; solo home, print y la lupa se reconocen
universalmente. Ninguno de los conceptos de Dydi tiene glifo estándar.

**Nunca escondas ni deshabilites una tab.** Apple HIG: *"If a section is empty,
**explain why** its content is unavailable"*.

## Enseñar dentro de la app

**Los tutoriales no funcionan.** NN/g, medido: *"Tutorials didn't improve task
performance"*. Apple HIG propone la alternativa: *"a collection of
context-specific tips instead of a single onboarding flow"*.

**El estado vacío es el lugar más barato para enseñar.** NN/g: úsalo para
educación contextual y ofrece la vía directa a la tarea que lo llenaría (su
ejemplo: *"Star your favorites to list them here"*). Dydi arranca vacío en 4 de 5
tabs: ese espacio ya estaba desperdiciado.

**No obligues a recordar entre pantallas.** NN/g H6: la memoria de trabajo
aguanta ~7 elementos y ~20 segundos. La razón por la que entras a la ruleta vive
en Hoy; la consecuencia se ejecuta en Ruleta. La pantalla debe cargar con el
recuerdo, no el usuario.

## Errores y estado

**El texto del servidor no es un mensaje para el usuario.** WCAG 3.3.1 (A) exige
describir el error **en texto** al usuario; 3.3.3 (AA) exige ofrecer la corrección
cuando se conoce. NN/g H9: lenguaje llano, sin códigos, con solución. Apple HIG
prohíbe títulos tipo "Error" o un código.

→ Por eso existe **`mobile/lib/errorMessage.ts`** y es la única fuente de
mensajes de error del APK. Un `{"error":"could not sync user"}` nunca llega a la
pantalla.

**Snackbar vs diálogo es una tabla, no gusto.** M3: snackbar = prioridad baja y
acción opcional; diálogo = prioridad alta y acción obligatoria. Un snackbar admite
**2 líneas** en compact y **una** acción, va **por encima** de la barra de
navegación, y uno a la vez.

**Los mensajes de estado deben anunciarse sin robar el foco.** WCAG 4.1.3 (AA).
En Dydi varios llegan solos por WebSocket (propuesta aprobada, deuda creada).

**Lo accionable se distingue de lo informativo.** W3C Mobile Accessibility Note
4.5, por forma, posición, etiqueta o iconografía. Y la nota 4.4: lo que dispara la
misma acción va en **un solo** elemento tocable.

→ Por eso en `index.tsx` la fila de un hábito pendiente es un solo `Pressable` de
48 dp con chevron, y la de un hábito ya cumplido no es tocable.
