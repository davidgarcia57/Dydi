-- Dydi — Schema inicial
-- week_start siempre es el lunes de la semana ISO: DATE_TRUNC('week', NOW())::DATE

-- =============================================================
-- USUARIOS Y GRUPOS
-- =============================================================

CREATE TABLE IF NOT EXISTS users (
    id           UUID PRIMARY KEY,  -- mismo id que auth.users de Supabase
    display_name TEXT NOT NULL,
    avatar_url   TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS groups (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name         TEXT NOT NULL,
    invite_code  TEXT NOT NULL UNIQUE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS group_members (
    group_id   UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
    joined_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (group_id, user_id)
);

-- =============================================================
-- HÁBITOS
-- =============================================================

-- Catálogo fijo gestionado por el equipo. Los usuarios eligen de aquí.
-- icon_key mapea al SVG/animación en el frontend.
CREATE TABLE IF NOT EXISTS habits (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    description TEXT,
    icon_key    TEXT NOT NULL,
    color       TEXT NOT NULL
);

-- Hábito asignado a un usuario dentro de un grupo.
-- Se crean N filas (una por miembro) cuando el grupo aprueba un add_habit.
-- scheduled_time es individual: cada miembro decide su horario personal.
-- status se deriva en el API — no se almacena:
--   done    = existe checkin para hoy
--   pending = no existe checkin y aún no pasó la medianoche
--   missed  = no existe checkin y ya pasó la medianoche
CREATE TABLE IF NOT EXISTS user_habits (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
    group_id       UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    habit_id       UUID NOT NULL REFERENCES habits(id) ON DELETE CASCADE,
    scheduled_time TIME,           -- recordatorio personal, ej: '06:00'. Nullable.
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, group_id, habit_id)
);

-- Check-in diario. checked_on es DATE para simplificar "¿hizo check hoy?".
CREATE TABLE IF NOT EXISTS checkins (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_habit_id UUID NOT NULL REFERENCES user_habits(id) ON DELETE CASCADE,
    checked_on    DATE NOT NULL DEFAULT CURRENT_DATE,
    note          TEXT,
    UNIQUE (user_habit_id, checked_on)
);

-- =============================================================
-- RULETA Y CASTIGOS
-- =============================================================

-- Una entrada por persona por semana en que falló al menos un hábito.
-- Se crea en cuanto se detecta la primera falla de la semana.
-- week_start = lunes de esa semana.
CREATE TABLE IF NOT EXISTS roulette_entries (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id   UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    debtor_id  UUID NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
    week_start DATE NOT NULL,
    status     TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'completed')),
    spun_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (group_id, debtor_id, week_start)
);

-- Sugerencias de castigo propuestas por los miembros del grupo durante la semana.
-- Son un pool compartido por grupo+semana — cualquiera puede caer en la ruleta
-- y recibir cualquier castigo del pool. Cada miembro puede agregar la suya.
CREATE TABLE IF NOT EXISTS punishment_suggestions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id     UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    week_start   DATE NOT NULL,
    suggester_id UUID NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
    text         TEXT NOT NULL,
    emoji        TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (group_id, week_start, suggester_id)  -- un castigo por persona por semana
);

-- Resultado del giro. debtor_id/group_id/week_start se leen vía entry_id (no se duplican).
-- punishment_text se copia para preservar el texto histórico si la sugerencia se elimina.
CREATE TABLE IF NOT EXISTS debts (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entry_id         UUID NOT NULL REFERENCES roulette_entries(id)      ON DELETE CASCADE,
    suggestion_id    UUID          REFERENCES punishment_suggestions(id) ON DELETE SET NULL,
    punishment_text  TEXT NOT NULL,
    punishment_emoji TEXT,
    resolved         BOOLEAN NOT NULL DEFAULT FALSE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at      TIMESTAMPTZ
);

-- =============================================================
-- PROPUESTAS Y VOTACIONES
-- =============================================================

-- Tipos de propuesta y sus payloads:
--   add_habit     → { "habit_id": "uuid", "scheduled_time": "06:00" }
--                    scheduled_time es el default que se asigna a todos los miembros.
--                    Cada usuario puede ajustar el suyo después.
--   remove_habit  → { "habit_id": "uuid" }
--                    Elimina user_habits de todos los miembros (cascada a checkins futuros).
--   kick_member   → { "user_id": "uuid" }
--   delete_group  → {}
--
-- Aprobación: > 50 % de miembros vota TRUE dentro de expires_at.
-- Si expires_at pasa sin mayoría → status = 'expired'.
CREATE TABLE IF NOT EXISTS proposals (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id    UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    proposer_id UUID NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
    type        TEXT NOT NULL CHECK (type IN ('add_habit', 'remove_habit', 'kick_member', 'delete_group')),
    payload     JSONB NOT NULL DEFAULT '{}',
    status      TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'approved', 'rejected', 'expired')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at  TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '24 hours'
);

CREATE TABLE IF NOT EXISTS proposal_votes (
    proposal_id UUID NOT NULL REFERENCES proposals(id) ON DELETE CASCADE,
    voter_id    UUID NOT NULL REFERENCES users(id)     ON DELETE CASCADE,
    approved    BOOLEAN NOT NULL,
    voted_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (proposal_id, voter_id)
);

-- =============================================================
-- ÍNDICES
-- =============================================================

-- checkins: query principal = "checkins de hoy para este grupo"
CREATE INDEX IF NOT EXISTS idx_checkins_user_habit    ON checkins (user_habit_id, checked_on DESC);
CREATE INDEX IF NOT EXISTS idx_checkins_date          ON checkins (checked_on DESC);

-- user_habits: hábitos de un usuario / hábitos de un grupo
CREATE INDEX IF NOT EXISTS idx_user_habits_user       ON user_habits (user_id);
CREATE INDEX IF NOT EXISTS idx_user_habits_group      ON user_habits (group_id, user_id);

-- group_members: grupos a los que pertenece un usuario
CREATE INDEX IF NOT EXISTS idx_group_members_user     ON group_members (user_id);

-- ruleta y deudas
CREATE INDEX IF NOT EXISTS idx_roulette_group_week    ON roulette_entries (group_id, week_start);
CREATE INDEX IF NOT EXISTS idx_roulette_debtor        ON roulette_entries (debtor_id, week_start);
CREATE INDEX IF NOT EXISTS idx_debts_entry            ON debts (entry_id);

-- sugerencias de castigo
CREATE INDEX IF NOT EXISTS idx_suggestions_group_week ON punishment_suggestions (group_id, week_start);

-- propuestas
CREATE INDEX IF NOT EXISTS idx_proposals_group_status ON proposals (group_id, status);

-- =============================================================
-- TRIGGER: sincronizar auth.users → public.users
-- =============================================================
-- Corre dentro de Postgres en la misma transacción que el registro.
-- No requiere llamada HTTP desde el frontend — escala con Supabase.

CREATE OR REPLACE FUNCTION public.handle_new_user()
RETURNS trigger AS $$
BEGIN
  INSERT INTO public.users (id, display_name, avatar_url)
  VALUES (
    new.id,
    COALESCE(new.raw_user_meta_data->>'display_name', split_part(new.email, '@', 1)),
    new.raw_user_meta_data->>'avatar_url'
  )
  ON CONFLICT (id) DO UPDATE
    SET display_name = COALESCE(EXCLUDED.display_name, public.users.display_name),
        avatar_url   = COALESCE(EXCLUDED.avatar_url,   public.users.avatar_url);
  RETURN new;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

CREATE OR REPLACE TRIGGER on_auth_user_created
  AFTER INSERT OR UPDATE ON auth.users
  FOR EACH ROW EXECUTE FUNCTION public.handle_new_user();
