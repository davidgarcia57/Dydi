-- Dydi — Schema inicial (versión definitiva)
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

-- Tabla de unión groups × users.
-- Representa la membresía de un usuario en un grupo.
CREATE TABLE IF NOT EXISTS user_groups (
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

-- Asignación de un hábito a un usuario dentro de un grupo.
-- Tabla de unión ternaria: users × habits × groups.
-- scheduled_time es individual: cada miembro decide su horario personal.
-- El estado (done/pending/missed) se deriva en el API, no se almacena.
CREATE TABLE IF NOT EXISTS habit_assignments (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
    group_id       UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    habit_id       UUID NOT NULL REFERENCES habits(id) ON DELETE CASCADE,
    scheduled_time TIME,           -- recordatorio personal, ej: '06:00'. Nullable.
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, group_id, habit_id)
);

-- Check-in diario.
-- checked_on DATE registra el día de negocio (¿hizo check hoy?).
-- created_at TIMESTAMPTZ registra el instante exacto (para métricas y analytics).
CREATE TABLE IF NOT EXISTS checkins (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    habit_assignment_id UUID NOT NULL REFERENCES habit_assignments(id) ON DELETE CASCADE,
    checked_on          DATE NOT NULL DEFAULT CURRENT_DATE,
    note                TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (habit_assignment_id, checked_on)
);

-- =============================================================
-- RULETA Y CASTIGOS
-- =============================================================

-- Una entrada por persona por semana en que falló al menos un hábito.
-- spun_at IS NOT NULL indica que ya se giró la ruleta para esta persona esta semana.
-- week_start = lunes de esa semana (ISODOW = 1).
CREATE TABLE IF NOT EXISTS roulette_draws (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id   UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    debtor_id  UUID NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
    week_start DATE NOT NULL,
    spun_at    TIMESTAMPTZ,   -- NULL = pendiente de girar, NOT NULL = ya girado
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (group_id, debtor_id, week_start),
    CONSTRAINT chk_week_start_monday CHECK (EXTRACT(ISODOW FROM week_start) = 1)
);

-- Sugerencias de castigo propuestas por los miembros del grupo.
-- Son un pool compartido por grupo+semana — cualquier miembro puede recibir
-- cualquier castigo del pool. Un castigo por miembro por semana.
CREATE TABLE IF NOT EXISTS group_suggestions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id     UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    week_start   DATE NOT NULL,
    suggester_id UUID NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
    text         TEXT NOT NULL,
    emoji        TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (group_id, week_start, suggester_id),
    CONSTRAINT chk_suggestions_week_start_monday CHECK (EXTRACT(ISODOW FROM week_start) = 1)
);

-- Resultado del giro. group_id/debtor_id/week_start se leen vía draw_id (no se duplican).
-- punishment_text se copia para preservar el texto histórico si la sugerencia se elimina.
-- UNIQUE(draw_id): un giro produce exactamente una deuda; previene duplicados por race condition.
CREATE TABLE IF NOT EXISTS debts (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    draw_id               UUID NOT NULL UNIQUE REFERENCES roulette_draws(id) ON DELETE CASCADE,
    penalty_suggestion_id UUID         REFERENCES group_suggestions(id) ON DELETE SET NULL,
    punishment_text       TEXT NOT NULL,
    punishment_emoji      TEXT,
    resolved              BOOLEAN NOT NULL DEFAULT FALSE,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at           TIMESTAMPTZ,
    CONSTRAINT chk_resolved_time CHECK (
        (resolved = FALSE AND resolved_at IS NULL) OR
        (resolved = TRUE  AND resolved_at IS NOT NULL)
    )
);

-- =============================================================
-- PROPUESTAS Y VOTACIONES
-- =============================================================

-- Tipos de propuesta y sus payloads:
--   add_habit     → { "habit_id": "uuid", "scheduled_time": "06:00" }
--   remove_habit  → { "habit_id": "uuid" }
--   kick_member   → { "user_id": "uuid" }
--   delete_group  → {}
--
-- Aprobación: > 50 % de miembros vota TRUE dentro de expires_at.
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
CREATE INDEX IF NOT EXISTS idx_checkins_habit_assignment ON checkins (habit_assignment_id, checked_on DESC);
CREATE INDEX IF NOT EXISTS idx_checkins_date             ON checkins (checked_on DESC);

-- habit_assignments: hábitos de un usuario / hábitos de un grupo
CREATE INDEX IF NOT EXISTS idx_habit_assignments_user  ON habit_assignments (user_id);
CREATE INDEX IF NOT EXISTS idx_habit_assignments_group ON habit_assignments (group_id, user_id);

-- user_groups: grupos a los que pertenece un usuario
CREATE INDEX IF NOT EXISTS idx_user_groups_user ON user_groups (user_id);

-- ruleta y deudas
CREATE INDEX IF NOT EXISTS idx_roulette_draws_group_week ON roulette_draws (group_id, week_start);
CREATE INDEX IF NOT EXISTS idx_roulette_draws_debtor     ON roulette_draws (debtor_id, week_start);
CREATE INDEX IF NOT EXISTS idx_debts_draw                ON debts (draw_id);

-- sugerencias de castigo
CREATE INDEX IF NOT EXISTS idx_group_suggestions_group_week ON group_suggestions (group_id, week_start);

-- propuestas
CREATE INDEX IF NOT EXISTS idx_proposals_group_status ON proposals (group_id, status);

-- =============================================================
-- TRIGGER: sincronizar auth.users → public.users
-- =============================================================
-- Corre dentro de Postgres en la misma transacción que el registro.
-- search_path fijo para prevenir ataques de manipulación del search_path.

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
$$ LANGUAGE plpgsql SECURITY DEFINER SET search_path = '';

CREATE OR REPLACE TRIGGER on_auth_user_created
  AFTER INSERT OR UPDATE ON auth.users
  FOR EACH ROW EXECUTE FUNCTION public.handle_new_user();

-- Solo el trigger (ejecutado como superuser) puede invocar esta función.
REVOKE EXECUTE ON FUNCTION public.handle_new_user() FROM PUBLIC, anon, authenticated;
