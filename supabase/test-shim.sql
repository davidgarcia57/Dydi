-- =============================================================
-- Dydi — shim de Supabase para Postgres pelón (SOLO tests)
--
-- Las migraciones de supabase/migrations/ asumen un proyecto Supabase:
-- el schema `extensions`, la tabla `auth.users` (a la que public.users
-- referencia por FK) y los roles `anon`/`authenticated`. Un contenedor
-- postgres:15 limpio no trae nada de eso, así que 001_initial.sql
-- falla en la primera línea.
--
-- Este archivo crea el mínimo indispensable para que las migraciones
-- reales se apliquen sin tocarlas. NO se aplica nunca a Supabase:
-- lo usa scripts/test-db.sh contra una BD efímera y desechable.
--
-- auth.users aquí es una maqueta con solo las columnas que el trigger
-- handle_new_user() lee (id, email, raw_user_meta_data). No pretende
-- imitar el resto del esquema de GoTrue.
-- =============================================================

CREATE SCHEMA IF NOT EXISTS extensions;
CREATE SCHEMA IF NOT EXISTS auth;

-- Los roles de Supabase existen solo para que los GRANT/REVOKE de las
-- migraciones no revienten. En los tests nos conectamos como owner.
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'anon') THEN
        CREATE ROLE anon NOLOGIN;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'authenticated') THEN
        CREATE ROLE authenticated NOLOGIN;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'service_role') THEN
        CREATE ROLE service_role NOLOGIN;
    END IF;
END
$$;

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA extensions;

-- Solo las columnas que handle_new_user() consume.
CREATE TABLE IF NOT EXISTS auth.users (
    id                 UUID PRIMARY KEY DEFAULT extensions.gen_random_uuid(),
    email              TEXT,
    raw_user_meta_data JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- gen_random_uuid() sin cualificar: las migraciones lo llaman así en los
-- DEFAULT de cada tabla, y en Supabase resuelve porque `extensions` va en
-- el search_path del rol. Lo replicamos a nivel de BD.
DO $$
BEGIN
    EXECUTE format(
        'ALTER DATABASE %I SET search_path TO public, extensions',
        current_database()
    );
END
$$;
