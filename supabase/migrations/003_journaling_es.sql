-- "Journaling" era el único nombre del catálogo que quedó en inglés; los otros
-- once ya estaban en español. Se renombra aquí en vez de editar el INSERT de
-- 001_initial.sql porque esa migración ya está aplicada en producción.
--
-- Renombrar es seguro para los datos existentes: group_habits y checkins apuntan
-- a habits.id (UUID), nunca al nombre. En una base nueva el orden también queda
-- bien, porque 001 siembra la fila y 003 corre después.
--
-- El NOT EXISTS solo hace la migración idempotente y evita quedar con dos
-- hábitos llamados igual si alguien ya creó uno con el nombre nuevo a mano
-- (habits.name no tiene restricción UNIQUE que lo impida).
UPDATE habits
   SET name = 'Escribir un diario'
 WHERE name = 'Journaling'
   AND NOT EXISTS (SELECT 1 FROM habits h2 WHERE h2.name = 'Escribir un diario');
