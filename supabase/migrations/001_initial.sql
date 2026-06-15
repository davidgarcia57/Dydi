-- =============================================================
-- Dydi — canonical schema (consolidated redesign)
-- Source of truth. If code disagrees with this file, fix the code.
--
-- Design principle (the spine of this schema):
--   Every row that carries a group_id must PROVE, via a composite
--   foreign key, that the user / proposal / vote / suggestion / debt
--   actually belongs to that same group. The database must make an
--   invalid cross-group state impossible — we never trust the app to
--   send coherent data.
--
-- This file assumes a fresh database. To re-apply over an existing
-- dev database, uncomment the RESET block below first (drops all data).
-- =============================================================

-- ┌───────────────────────────────────────────────────────────┐
-- │ RESET (dev only — destroys data). Uncomment to recreate.    │
-- │ Includes the OLD names (group_members) so it cleanly        │
-- │ replaces a previous deploy. auth.users is NOT touched.      │
-- └───────────────────────────────────────────────────────────┘
-- DROP TABLE IF EXISTS
--   proposal_votes, proposal_eligible_voters, proposals,
--   debts, punishment_suggestions, roulette_entries,
--   checkins, user_habits, group_habits,
--   group_members, memberships,        -- old name + new name
--   habits, groups, users
-- CASCADE;
-- DROP FUNCTION IF EXISTS public.handle_new_user()  CASCADE;
-- DROP FUNCTION IF EXISTS public.set_updated_at()   CASCADE;
-- DROP FUNCTION IF EXISTS public.enforce_group_size() CASCADE;

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA extensions;

-- Generic updated_at trigger, reused across mutable tables.
CREATE OR REPLACE FUNCTION public.set_updated_at()
RETURNS trigger AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- =============================================================
-- IDENTITY
-- =============================================================

-- Public profile for an authenticated user. The id is a hard FK to
-- auth.users(id): no profile can exist without an auth account, and
-- deleting the auth account cascades the profile away.
-- Populated/kept in sync by handle_new_user() (see bottom of file).
CREATE TABLE IF NOT EXISTS users (
    id           UUID PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,
    display_name TEXT        NOT NULL,
    avatar_url   TEXT,
    -- Used to derive checkins.checked_on in the member's local day,
    -- so an 11:30pm check-in is not bumped to "tomorrow" in UTC.
    timezone     TEXT        NOT NULL DEFAULT 'America/Monterrey',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_users_display_name_len
        CHECK (length(btrim(display_name)) BETWEEN 1 AND 50)
);

CREATE TABLE IF NOT EXISTS groups (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT        NOT NULL,
    invite_code TEXT        NOT NULL UNIQUE,
    -- Who created the group. Kept even if the system is democratic, so an
    -- owner exists for emergency operations (e.g. last-member edge cases).
    created_by  UUID        REFERENCES users(id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_groups_name_len
        CHECK (length(btrim(name)) BETWEEN 1 AND 60)
);

-- Membership = the join table between users and groups, modeled as a
-- first-class entity (it has its own attributes and lifecycle), so it is
-- named as an entity rather than as a generic pivot (groups_users).
-- Members are never hard-deleted on leave/kick: status preserves history
-- (their past votes, suggestions and debts stay valid).
-- Hard ceiling of 8 active members enforced by trigger below.
CREATE TABLE IF NOT EXISTS memberships (
    group_id   UUID        NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    user_id    UUID        NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
    role       TEXT        NOT NULL DEFAULT 'member'
        CHECK (role IN ('owner', 'admin', 'member')),
    status     TEXT        NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'left', 'kicked')),
    joined_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    left_at    TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (group_id, user_id),
    -- An active member has no left_at; an inactive one must record when.
    CONSTRAINT chk_memberships_left_state
        CHECK (
            (status =  'active' AND left_at IS NULL)
            OR
            (status <> 'active' AND left_at IS NOT NULL)
        )
);

CREATE TRIGGER trg_memberships_updated_at
    BEFORE UPDATE ON memberships
    FOR EACH ROW EXECUTE FUNCTION public.set_updated_at();

-- =============================================================
-- HABITS
-- =============================================================

-- Global pre-seeded catalog. Users never create habits — they propose
-- adding/removing a catalog habit to/from their group via proposals.
CREATE TABLE IF NOT EXISTS habits (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT    NOT NULL,
    description TEXT,
    icon_key    TEXT    NOT NULL,
    color       TEXT    NOT NULL,
    is_active   BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- A catalog habit ADOPTED by a group. This is the missing layer:
-- "this habit is available/active in this group". Created when an
-- add_habit proposal is approved; archived on remove_habit.
-- user_habits then reference (group_id, habit_id) here, guaranteeing a
-- member can only hold habits the group has actually adopted.
CREATE TABLE IF NOT EXISTS group_habits (
    group_id    UUID        NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    habit_id    UUID        NOT NULL REFERENCES habits(id) ON DELETE RESTRICT,
    added_by    UUID        NOT NULL,
    added_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    archived_at TIMESTAMPTZ,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (group_id, habit_id),
    -- The adder must be a member of the same group.
    FOREIGN KEY (group_id, added_by)
        REFERENCES memberships(group_id, user_id)
        ON DELETE CASCADE
);

CREATE TRIGGER trg_group_habits_updated_at
    BEFORE UPDATE ON group_habits
    FOR EACH ROW EXECUTE FUNCTION public.set_updated_at();

-- One row per member per adopted habit. Created/deleted by the proposals
-- system. The two composite FKs guarantee BOTH that the user is an active
-- member of the group AND that the habit is adopted by that group.
CREATE TABLE IF NOT EXISTS user_habits (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID NOT NULL,
    group_id       UUID NOT NULL,
    habit_id       UUID NOT NULL,
    scheduled_time TIME,                              -- single daily reminder (MVP)
    is_active      BOOLEAN NOT NULL DEFAULT true,     -- pause without deleting history
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (group_id, user_id, habit_id),
    FOREIGN KEY (group_id, user_id)
        REFERENCES memberships(group_id, user_id)
        ON DELETE CASCADE,
    FOREIGN KEY (group_id, habit_id)
        REFERENCES group_habits(group_id, habit_id)
        ON DELETE CASCADE
);

CREATE TRIGGER trg_user_habits_updated_at
    BEFORE UPDATE ON user_habits
    FOR EACH ROW EXECUTE FUNCTION public.set_updated_at();

-- One check-in per user_habit per calendar day (member's local day).
-- Presence of a row = done. Absence = pending/missed (derived at query time).
-- checked_on must be computed by the app as
--   (now() AT TIME ZONE u.timezone)::date
CREATE TABLE IF NOT EXISTS checkins (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_habit_id UUID        NOT NULL REFERENCES user_habits(id) ON DELETE CASCADE,
    checked_on    DATE        NOT NULL,
    note          TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_habit_id, checked_on),
    CONSTRAINT chk_checkins_note_len
        CHECK (note IS NULL OR length(note) <= 500)
);

-- =============================================================
-- ROULETTE AND DEBTS
-- =============================================================

-- Opened for an eligible offender (missed >=1 habit Mon–Fri of the ISO week).
-- Members submit suggestions until suggestion_deadline; after that the
-- offender spins. No suggestions at deadline -> collective debt.
-- NOTE: "deadline has passed" is checked at runtime (in the spin RPC/handler),
-- never as a CHECK — CHECKs are evaluated on write, they do not turn false
-- as time advances.
CREATE TABLE IF NOT EXISTS roulette_entries (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id            UUID        NOT NULL,
    debtor_id           UUID        NOT NULL,
    week_start          DATE        NOT NULL,
    suggestion_deadline TIMESTAMPTZ NOT NULL,
    spun_at             TIMESTAMPTZ,                   -- NULL = not yet spun
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (group_id, debtor_id, week_start),
    UNIQUE (id, group_id),                             -- target for composite FKs below
    CONSTRAINT chk_roulette_week_start_monday
        CHECK (EXTRACT(ISODOW FROM week_start) = 1),
    CONSTRAINT chk_roulette_deadline_after_creation
        CHECK (suggestion_deadline > created_at),
    CONSTRAINT chk_roulette_spun_after_creation
        CHECK (spun_at IS NULL OR spun_at >= created_at),
    FOREIGN KEY (group_id, debtor_id)
        REFERENCES memberships(group_id, user_id)
        ON DELETE CASCADE
);

-- One suggestion per member per roulette_entry (the offender may self-suggest).
-- Locked once submitted (no updates).
CREATE TABLE IF NOT EXISTS punishment_suggestions (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    roulette_entry_id UUID        NOT NULL,
    group_id          UUID        NOT NULL,   -- denormalized to enforce same-group FK
    suggester_id      UUID        NOT NULL,
    text              TEXT        NOT NULL,
    emoji             TEXT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (roulette_entry_id, suggester_id),
    UNIQUE (id, roulette_entry_id),            -- target for debts.winning FK
    CONSTRAINT chk_suggestion_text_len
        CHECK (length(btrim(text)) BETWEEN 1 AND 300),
    -- Suggestion belongs to a roulette of this exact group...
    FOREIGN KEY (roulette_entry_id, group_id)
        REFERENCES roulette_entries(id, group_id)
        ON DELETE CASCADE,
    -- ...and the suggester is a member of that same group.
    FOREIGN KEY (group_id, suggester_id)
        REFERENCES memberships(group_id, user_id)
        ON DELETE CASCADE
);

-- One debt per (roulette_entry, debtor).
--   scope = 'individual' : winning_suggestion_id set, one row (the offender).
--   scope = 'collective' : winning_suggestion_id NULL, one row per member.
-- punishment_text is always snapshotted, so deleting the suggestion never
-- erases the debt description. Debts auto-expire (expires_at); status can
-- also be moved to completed/forgiven explicitly.
CREATE TABLE IF NOT EXISTS debts (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    roulette_entry_id     UUID        NOT NULL,
    group_id              UUID        NOT NULL,
    debtor_id             UUID        NOT NULL,
    week_start            DATE        NOT NULL,
    winning_suggestion_id UUID,
    punishment_text       TEXT        NOT NULL,
    punishment_emoji      TEXT,
    scope                 TEXT        NOT NULL DEFAULT 'individual'
        CHECK (scope IN ('individual', 'collective')),
    status                TEXT        NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'completed', 'expired', 'forgiven')),
    completed_at          TIMESTAMPTZ,
    expires_at            DATE        NOT NULL,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (roulette_entry_id, debtor_id),     -- no two debts per member per entry
    CONSTRAINT chk_debts_week_start_monday
        CHECK (EXTRACT(ISODOW FROM week_start) = 1),
    CONSTRAINT chk_debts_expires_following_week
        CHECK (expires_at >= week_start + INTERVAL '14 days'),
    CONSTRAINT chk_debts_completed_state
        CHECK (
            (status =  'completed' AND completed_at IS NOT NULL)
            OR
            (status <> 'completed' AND completed_at IS NULL)
        ),
    CONSTRAINT chk_debts_collective_has_no_winner
        CHECK (scope <> 'collective' OR winning_suggestion_id IS NULL),
    -- Debt's roulette belongs to this exact group...
    FOREIGN KEY (roulette_entry_id, group_id)
        REFERENCES roulette_entries(id, group_id)
        ON DELETE CASCADE,
    -- ...the debtor is a member of that same group...
    FOREIGN KEY (group_id, debtor_id)
        REFERENCES memberships(group_id, user_id)
        ON DELETE CASCADE,
    -- ...and the winning suggestion (if any) belongs to that same roulette.
    -- (Composite FK is skipped when winning_suggestion_id IS NULL — MATCH SIMPLE.)
    FOREIGN KEY (winning_suggestion_id, roulette_entry_id)
        REFERENCES punishment_suggestions(id, roulette_entry_id)
);

CREATE TRIGGER trg_debts_updated_at
    BEFORE UPDATE ON debts
    FOR EACH ROW EXECUTE FUNCTION public.set_updated_at();

-- =============================================================
-- PROPOSALS AND VOTES
-- =============================================================

-- All group changes go through proposals.
-- Typed columns (not JSONB) so referential integrity holds:
--   add_habit / remove_habit -> habit_id required, target_user_id forbidden
--   kick_member              -> target_user_id required, habit_id forbidden
--   delete_group             -> both forbidden
CREATE TABLE IF NOT EXISTS proposals (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id       UUID        NOT NULL,
    proposer_id    UUID        NOT NULL,
    type           TEXT        NOT NULL
        CHECK (type IN ('add_habit', 'remove_habit', 'kick_member', 'delete_group')),
    habit_id       UUID        REFERENCES habits(id) ON DELETE CASCADE,
    target_user_id UUID,
    status         TEXT        NOT NULL DEFAULT 'open'
        CHECK (status IN ('open', 'approved', 'rejected', 'expired')),
    resolved_at    TIMESTAMPTZ,
    resolved_by    UUID        REFERENCES users(id) ON DELETE SET NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at     TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '7 days',
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Exact payload per type — nothing extra, nothing missing.
    CONSTRAINT chk_proposals_payload
        CHECK (
            (type IN ('add_habit', 'remove_habit')
                AND habit_id IS NOT NULL AND target_user_id IS NULL)
            OR
            (type = 'kick_member'
                AND target_user_id IS NOT NULL AND habit_id IS NULL)
            OR
            (type = 'delete_group'
                AND habit_id IS NULL AND target_user_id IS NULL)
        ),
    -- Resolved proposals must record when.
    CONSTRAINT chk_proposals_resolved_state
        CHECK (
            (status =  'open' AND resolved_at IS NULL)
            OR
            (status <> 'open' AND resolved_at IS NOT NULL)
        ),
    UNIQUE (id, group_id),                     -- target for composite FKs below
    FOREIGN KEY (group_id, proposer_id)
        REFERENCES memberships(group_id, user_id)
        ON DELETE CASCADE,
    -- A kick target must be a member of THIS group (not just any user).
    -- Composite FK is skipped when target_user_id IS NULL — MATCH SIMPLE.
    FOREIGN KEY (group_id, target_user_id)
        REFERENCES memberships(group_id, user_id)
        ON DELETE CASCADE
);

CREATE TRIGGER trg_proposals_updated_at
    BEFORE UPDATE ON proposals
    FOR EACH ROW EXECUTE FUNCTION public.set_updated_at();

-- Frozen electorate: the active members at the moment a proposal opens.
-- Quorum is computed against THIS set, so it doesn't drift if people join
-- or leave mid-vote. Populated by the app when the proposal is created.
CREATE TABLE IF NOT EXISTS proposal_eligible_voters (
    proposal_id UUID NOT NULL,
    group_id    UUID NOT NULL,
    user_id     UUID NOT NULL,
    PRIMARY KEY (proposal_id, user_id),
    FOREIGN KEY (proposal_id, group_id)
        REFERENCES proposals(id, group_id)
        ON DELETE CASCADE,
    FOREIGN KEY (group_id, user_id)
        REFERENCES memberships(group_id, user_id)
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS proposal_votes (
    proposal_id UUID        NOT NULL,
    group_id    UUID        NOT NULL,
    voter_id    UUID        NOT NULL,
    approved    BOOLEAN     NOT NULL,
    voted_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (proposal_id, voter_id),
    FOREIGN KEY (proposal_id, group_id)
        REFERENCES proposals(id, group_id)
        ON DELETE CASCADE,
    -- A voter must be in the frozen electorate of this proposal.
    FOREIGN KEY (proposal_id, voter_id)
        REFERENCES proposal_eligible_voters(proposal_id, user_id)
        ON DELETE CASCADE
);

-- =============================================================
-- GROUP SIZE CEILING (hard rule: max 8 active members)
-- Mirrors MAX_GROUP_SIZE in groups-service; enforced here so the DB
-- itself refuses a 9th active member even under a race or direct write.
-- =============================================================
CREATE OR REPLACE FUNCTION public.enforce_group_size()
RETURNS trigger AS $$
BEGIN
    IF NEW.status = 'active' THEN
        IF (SELECT COUNT(*) FROM public.memberships
            WHERE group_id = NEW.group_id AND status = 'active') >= 8 THEN
            RAISE EXCEPTION 'group % already has 8 active members', NEW.group_id
                USING ERRCODE = 'check_violation';
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_memberships_size_limit
    BEFORE INSERT OR UPDATE OF status ON memberships
    FOR EACH ROW EXECUTE FUNCTION public.enforce_group_size();

-- =============================================================
-- INDEXES
-- Postgres auto-indexes PKs and UNIQUE constraints, but NOT foreign keys.
-- We index FK columns (for cascades + reverse lookups) and hot query paths.
-- =============================================================

-- memberships: PK(group_id,user_id) covers group lookups; add the reverse.
CREATE INDEX IF NOT EXISTS idx_memberships_user
    ON memberships (user_id, group_id);

-- group_habits FKs
CREATE INDEX IF NOT EXISTS idx_group_habits_habit
    ON group_habits (habit_id);
CREATE INDEX IF NOT EXISTS idx_group_habits_added_by
    ON group_habits (group_id, added_by);

-- user_habits FKs / hot paths
CREATE INDEX IF NOT EXISTS idx_user_habits_user
    ON user_habits (user_id);
CREATE INDEX IF NOT EXISTS idx_user_habits_group_habit
    ON user_habits (group_id, habit_id);

-- checkins
CREATE INDEX IF NOT EXISTS idx_checkins_date
    ON checkins (checked_on DESC);

-- roulette
CREATE INDEX IF NOT EXISTS idx_roulette_group_week
    ON roulette_entries (group_id, week_start);
CREATE INDEX IF NOT EXISTS idx_roulette_debtor_week
    ON roulette_entries (debtor_id, week_start);

-- suggestions FKs
CREATE INDEX IF NOT EXISTS idx_suggestions_entry
    ON punishment_suggestions (roulette_entry_id);
CREATE INDEX IF NOT EXISTS idx_suggestions_group_suggester
    ON punishment_suggestions (group_id, suggester_id);

-- debts FKs / hot paths
CREATE INDEX IF NOT EXISTS idx_debts_group_active
    ON debts (group_id, expires_at DESC);
CREATE INDEX IF NOT EXISTS idx_debts_debtor_active
    ON debts (debtor_id, expires_at DESC);
CREATE INDEX IF NOT EXISTS idx_debts_winner
    ON debts (winning_suggestion_id, roulette_entry_id);

-- proposals FKs / hot paths
CREATE INDEX IF NOT EXISTS idx_proposals_group_status
    ON proposals (group_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_proposals_proposer
    ON proposals (group_id, proposer_id);
CREATE INDEX IF NOT EXISTS idx_proposals_target
    ON proposals (group_id, target_user_id);

-- proposal_eligible_voters / votes FKs
CREATE INDEX IF NOT EXISTS idx_eligible_voters_member
    ON proposal_eligible_voters (group_id, user_id);
CREATE INDEX IF NOT EXISTS idx_proposal_votes_voter
    ON proposal_votes (group_id, voter_id);

-- Only ONE open proposal at a time per actionable target:
CREATE UNIQUE INDEX IF NOT EXISTS uq_proposals_one_open_habit
    ON proposals (group_id, habit_id, type)
    WHERE status = 'open' AND habit_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_proposals_one_open_kick
    ON proposals (group_id, target_user_id)
    WHERE status = 'open' AND type = 'kick_member';
CREATE UNIQUE INDEX IF NOT EXISTS uq_proposals_one_open_delete
    ON proposals (group_id)
    WHERE status = 'open' AND type = 'delete_group';

-- =============================================================
-- AUTH SYNC TRIGGER: auth.users -> public.users
-- =============================================================
CREATE OR REPLACE FUNCTION public.handle_new_user()
RETURNS trigger AS $$
BEGIN
    INSERT INTO public.users (id, display_name, avatar_url)
    VALUES (
        NEW.id,
        COALESCE(NEW.raw_user_meta_data->>'display_name', split_part(NEW.email, '@', 1)),
        NEW.raw_user_meta_data->>'avatar_url'
    )
    ON CONFLICT (id) DO UPDATE
        SET display_name = COALESCE(EXCLUDED.display_name, public.users.display_name),
            avatar_url   = COALESCE(EXCLUDED.avatar_url,   public.users.avatar_url);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER SET search_path = '';

DROP TRIGGER IF EXISTS on_auth_user_created ON auth.users;

CREATE TRIGGER on_auth_user_created
    AFTER INSERT OR UPDATE ON auth.users
    FOR EACH ROW EXECUTE FUNCTION public.handle_new_user();

REVOKE EXECUTE ON FUNCTION public.handle_new_user() FROM PUBLIC, anon, authenticated;

-- =============================================================
-- ROW LEVEL SECURITY
-- All product data is accessed exclusively through api-gateway using the
-- service role key (which BYPASSES RLS). The frontend never touches these
-- tables directly — it uses Supabase only for Auth. RLS is therefore
-- enabled with NO policies as defense-in-depth: if the anon/authenticated
-- key ever reaches these tables, every row is denied by default.
-- (If the frontend is ever pointed straight at the DB, real per-group
--  policies must be added before that switch — see notes in PR/docs.)
-- =============================================================
ALTER TABLE users                    ENABLE ROW LEVEL SECURITY;
ALTER TABLE groups                   ENABLE ROW LEVEL SECURITY;
ALTER TABLE memberships              ENABLE ROW LEVEL SECURITY;
ALTER TABLE habits                   ENABLE ROW LEVEL SECURITY;
ALTER TABLE group_habits             ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_habits              ENABLE ROW LEVEL SECURITY;
ALTER TABLE checkins                 ENABLE ROW LEVEL SECURITY;
ALTER TABLE roulette_entries         ENABLE ROW LEVEL SECURITY;
ALTER TABLE punishment_suggestions   ENABLE ROW LEVEL SECURITY;
ALTER TABLE debts                    ENABLE ROW LEVEL SECURITY;
ALTER TABLE proposals                ENABLE ROW LEVEL SECURITY;
ALTER TABLE proposal_eligible_voters ENABLE ROW LEVEL SECURITY;
ALTER TABLE proposal_votes           ENABLE ROW LEVEL SECURITY;

-- =============================================================
-- CATALOG SEED (idempotent by name)
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
