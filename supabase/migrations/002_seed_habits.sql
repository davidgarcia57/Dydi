-- =============================================================
-- Dydi — catalog seed
-- Run in Supabase SQL Editor (safe to re-run, idempotente por nombre).
-- =============================================================

INSERT INTO habits (name, description, icon_key, color)
SELECT v.name, v.description, v.icon_key, v.color
FROM (VALUES
  ('Ejercicio 30 min',   'Cualquier actividad física continua por al menos 30 minutos',  'exercise',  '#C9714A'),
  ('Leer 20 páginas',    'Lectura de cualquier libro, no pantallas',                     'read',      '#3D6B5E'),
  ('Meditar 10 min',     'Meditación guiada o en silencio, sin distracciones',           'meditate',  '#A8C39A'),
  ('Dormir antes de 11', 'Estar en cama con luces apagadas antes de las 11:00 pm',       'sleep',     '#7B8FA1'),
  ('Sin redes sociales', 'Cero scroll pasivo en IG, TikTok, Twitter durante el día',     'no_social', '#D4A847'),
  ('Agua 2 L',           'Completar al menos 2 litros de agua durante el día',           'water',     '#5B9BD5'),
  ('Sin azúcar',         'Sin refrescos, dulces ni postres procesados',                  'no_sugar',  '#E07070'),
  ('Journaling',         'Escribir al menos media página sobre el día o reflexiones',    'journal',   '#9B7FD4')
) AS v(name, description, icon_key, color)
WHERE NOT EXISTS (
  SELECT 1 FROM habits WHERE habits.name = v.name
);
