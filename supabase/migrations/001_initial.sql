-- Dydi — Schema inicial
-- week_start siempre es el lunes de la semana ISO (DATE_TRUNC('week', NOW()))

-- =============================================================
-- USUARIOS Y GRUPOS
-- =============================================================

-- Perfil extendido (identidad vive en Supabase Auth)
CREATE TABLE IF NOT EXISTS users (
    id           UUID PRIMARY KEY,  -- mismo id que Supabase Auth
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

-- Catálogo fijo — no lo crean los usuarios.
-- icon_key mapea a la animación/SVG en el frontend.
CREATE TABLE IF NOT EXISTS habits (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    description TEXT,
    icon_key    TEXT NOT NULL,  -- ej: "water_bottle", "running_shoes"
    color       TEXT NOT NULL   -- ej: "#22D3EE" para tomar agua
);

-- Hábitos activos por usuario dentro de un grupo.
-- Se crean/eliminan a través de propuestas (ver tabla proposals).
CREATE TABLE IF NOT EXISTS user_habits (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
    group_id   UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    habit_id   UUID NOT NULL REFERENCES habits(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, group_id, habit_id)
);

-- Check-in diario. checked_on es DATE para simplificar queries de "¿hizo check hoy?".
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

-- Una entrada por persona por semana en que falló algún hábito.
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

-- Sugerencias de castigo que los miembros proponen para cada entrada.
-- Cualquier miembro puede sugerir mientras status='open'.
CREATE TABLE IF NOT EXISTS punishment_suggestions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entry_id    UUID NOT NULL REFERENCES roulette_entries(id) ON DELETE CASCADE,
    suggester_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    text        TEXT NOT NULL,
    emoji       TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Resultado de la ruleta. Solo se guarda la semana actual (filtro en queries).
-- Se consulta siempre con: WHERE week_start >= CURRENT_DATE - INTERVAL '7 days'
CREATE TABLE IF NOT EXISTS debts (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id      UUID NOT NULL REFERENCES groups(id)               ON DELETE CASCADE,
    debtor_id     UUID NOT NULL REFERENCES users(id)                ON DELETE CASCADE,
    entry_id      UUID NOT NULL REFERENCES roulette_entries(id)     ON DELETE CASCADE,
    suggestion_id UUID          REFERENCES punishment_suggestions(id) ON DELETE SET NULL,
    punishment_text  TEXT NOT NULL,   -- copia del texto al momento del spin
    punishment_emoji TEXT,
    week_start    DATE NOT NULL,
    resolved      BOOLEAN NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at   TIMESTAMPTZ
);

-- =============================================================
-- PROPUESTAS Y VOTACIONES (consenso grupal)
-- =============================================================

-- Tipos de propuesta:
--   add_habit     → payload: { "habit_id": "uuid" }          (agrega a todos los miembros)
--   remove_habit  → payload: { "habit_id": "uuid" }          (quita a todos los miembros)
--   kick_member   → payload: { "user_id": "uuid" }
--   delete_group  → payload: {}
--
-- Aprobación: más del 50% de miembros vota TRUE dentro de expires_at.
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

CREATE INDEX IF NOT EXISTS idx_checkins_user_habit    ON checkins (user_habit_id, checked_on DESC);
CREATE INDEX IF NOT EXISTS idx_checkins_date          ON checkins (checked_on DESC);
CREATE INDEX IF NOT EXISTS idx_debts_week             ON debts (group_id, week_start DESC);
CREATE INDEX IF NOT EXISTS idx_roulette_group_week    ON roulette_entries (group_id, week_start);
CREATE INDEX IF NOT EXISTS idx_proposals_group_status ON proposals (group_id, status);
CREATE INDEX IF NOT EXISTS idx_user_habits_group      ON user_habits (group_id, user_id);
