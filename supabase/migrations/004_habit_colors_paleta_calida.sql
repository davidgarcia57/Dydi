-- Los colores del catálogo de hábitos venían de la paleta ANTERIOR, la que se
-- descartó al pasar a la paleta cálida de Dydi. Quedaron once de doce fuera del
-- sistema: un azul cobalto (#5B9BD5 para "Agua 2 L"), un morado (#9B7FD4 para
-- "Foco"), un rojo (#E07070), un verde oscuro (#3D6B5E), un dorado (#D4A847) y
-- un gris azulado (#7B8FA1). Por eso los círculos de la lista se veían pegados
-- con cinta contra el resto de la interfaz.
--
-- Aquí se remapean a los tokens que SÍ existen en tailwind.config.js (los mismos
-- valores en web y móvil), para que no quede ningún color fuera del sistema:
--
--   terracotta   #C26F4D    sage         #A8C39A    sage-deep    #7CA39D
--   amber        #E9C281    amber-deep   #A57B33    coral        #EDA48F
--   coral-deep   #BC5C42    accent-deep  #4C736C    ink-soft     #6F6557
--
-- Hay tres colores repetidos porque la paleta cálida tiene nueve tonos usables y
-- el catálogo tiene doce hábitos; el icono es lo que los distingue en esos casos.
-- Cualquier pareja concreta se puede mover con otro UPDATE: lo que importa es que
-- ninguno salga de la lista de arriba.
--
-- Se hace por nombre, no por id: los ids son UUID generados y cambian entre
-- entornos. Los nombres son la clave natural del seed de 001 (que también usa
-- `name` para su idempotencia). Un hábito que no exista simplemente no se toca.
UPDATE habits AS h
   SET color = v.color
  FROM (VALUES
    ('Ejercicio 30 min',       '#C26F4D'), -- terracotta
    ('8,000 pasos',            '#BC5C42'), -- coral-deep
    ('Agua 2 L',               '#7CA39D'), -- sage-deep  (era azul cobalto)
    ('2 frutas o verduras',    '#A8C39A'), -- sage       (ya estaba bien)
    ('Sin comida chatarra',    '#EDA48F'), -- coral      (era rojo puro)
    ('Leer 20 min',            '#4C736C'), -- accent-deep
    ('Foco 1 h sin teléfono',  '#A57B33'), -- amber-deep (era morado)
    ('Escribir un diario',     '#E9C281'), -- amber      (era gris azulado)
    ('Sin redes en la mañana', '#6F6557'), -- ink-soft   (era dorado)
    ('Sin teléfono de noche',  '#4C736C'), -- accent-deep
    ('Tender la cama',         '#E9C281'), -- amber
    ('15 min al aire libre',   '#A8C39A')  -- sage       (ya estaba bien)
  ) AS v(name, color)
 WHERE h.name = v.name
   AND h.color <> v.color;
