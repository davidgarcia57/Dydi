-- =============================================================
-- Dydi — canonical MVP schema
-- Source of truth. If code disagrees with this file, fix the code.
-- =============================================================

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA extensions;

-- =============================================================
-- IDENTITY
-- =============================================================

-- Mirrors auth.users. Populated and kept in sync by the trigger below.
CREATE TABLE IF NOT EXISTS users (
    id           UUID PRIMARY KEY,          -- same id as auth.users
    display_name TEXT        NOT NULL,
    avatar_url   TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS groups (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT        NOT NULL,
    invite_code TEXT        NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Max 8 members per group — enforced in application layer.
CREATE TABLE IF NOT EXISTS group_members (
    group_id  UUID        NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    user_id   UUID        NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (group_id, user_id)
);

-- =============================================================
-- HABITS
-- =============================================================

-- Global pre-seeded catalog. Users never create habits — they propose
-- adding or removing habits from this catalog via the proposals system.
CREATE TABLE IF NOT EXISTS habits (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    description TEXT,
    icon_key    TEXT NOT NULL,
    color       TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- One row per member per habit per group.
-- Created/deleted by the proposals system (add_habit / remove_habit).
-- The composite FK to group_members guarantees only active members
-- can hold habits in a group.
CREATE TABLE IF NOT EXISTS user_habits (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID NOT NULL,
    group_id       UUID NOT NULL,
    habit_id       UUID NOT NULL REFERENCES habits(id) ON DELETE CASCADE,
    scheduled_time TIME,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, group_id, habit_id),
    FOREIGN KEY (group_id, user_id)
        REFERENCES group_members(group_id, user_id)
        ON DELETE CASCADE
);

-- One check-in per user_habit per calendar day.
-- Presence of a row = done. Absence = pending/missed (derived at query time).
CREATE TABLE IF NOT EXISTS checkins (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_habit_id UUID        NOT NULL REFERENCES user_habits(id) ON DELETE CASCADE,
    checked_on    DATE        NOT NULL,
    note          TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_habit_id, checked_on)
);

-- =============================================================
-- ROULETTE AND DEBTS
-- =============================================================

-- Created when any group member opens a roulette for an eligible offender
-- (someone who missed at least one habit Mon–Fri of the current ISO week).
-- suggestion_deadline: members can submit suggestions until this timestamp.
-- After the deadline the offender can spin with whatever suggestions exist.
-- If no suggestions exist at deadline → collective debt (see debts table).
CREATE TABLE IF NOT EXISTS roulette_entries (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id            UUID        NOT NULL,
    debtor_id           UUID        NOT NULL,
    week_start          DATE        NOT NULL,
    suggestion_deadline TIMESTAMPTZ NOT NULL,
    spun_at             TIMESTAMPTZ,                   -- NULL = not yet spun
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (group_id, debtor_id, week_start),
    -- Required by punishment_suggestions composite FK below.
    UNIQUE (id, group_id),
    CONSTRAINT chk_roulette_week_start_monday
        CHECK (EXTRACT(ISODOW FROM week_start) = 1),
    FOREIGN KEY (group_id, debtor_id)
        REFERENCES group_members(group_id, user_id)
        ON DELETE CASCADE
);

-- One suggestion per group member per roulette_entry.
-- The offender can also suggest their own punishment.
-- Suggestions are locked once submitted (no updates).
CREATE TABLE IF NOT EXISTS punishment_suggestions (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    roulette_entry_id UUID        NOT NULL,
    group_id          UUID        NOT NULL,   -- denormalized for FK to group_members
    suggester_id      UUID        NOT NULL,
    text              TEXT        NOT NULL,
    emoji             TEXT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (roulette_entry_id, suggester_id),
    FOREIGN KEY (roulette_entry_id, group_id)
        REFERENCES roulette_entries(id, group_id)
        ON DELETE CASCADE,
    FOREIGN KEY (group_id, suggester_id)
        REFERENCES group_members(group_id, user_id)
        ON DELETE CASCADE
);

-- One debt per (roulette_entry, debtor).
-- Normal flow: winning_suggestion_id NOT NULL, is_collective = false,
--   one row with debtor_id = the offender.
-- Collective punishment: winning_suggestion_id NULL, is_collective = true,
--   one row per group member (all reference the same roulette_entry_id).
-- punishment_text is always snapshotted so deleting the suggestion
--   does not erase the debt description.
-- Debts auto-expire — there is no manual resolved state.
CREATE TABLE IF NOT EXISTS debts (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    roulette_entry_id     UUID        NOT NULL REFERENCES roulette_entries(id) ON DELETE CASCADE,
    group_id              UUID        NOT NULL,
    debtor_id             UUID        NOT NULL,
    week_start            DATE        NOT NULL,
    winning_suggestion_id UUID        REFERENCES punishment_suggestions(id) ON DELETE SET NULL,
    punishment_text       TEXT        NOT NULL,
    punishment_emoji      TEXT,
    is_collective         BOOLEAN     NOT NULL DEFAULT false,
    expires_at            DATE        NOT NULL,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (roulette_entry_id, debtor_id),
    CONSTRAINT chk_debts_week_start_monday
        CHECK (EXTRACT(ISODOW FROM week_start) = 1),
    CONSTRAINT chk_debts_expires_following_week
        CHECK (expires_at >= week_start + INTERVAL '14 days'),
    FOREIGN KEY (group_id, debtor_id)
        REFERENCES group_members(group_id, user_id)
        ON DELETE CASCADE
);

-- =============================================================
-- PROPOSALS AND VOTES
-- =============================================================

-- All habit changes go through proposals. Quorum: yes_votes * 2 >= member_count
-- (at least 50% of current members must approve; proposer's vote is implicit yes).
-- Typed columns instead of JSONB to preserve referential integrity:
--   add_habit / remove_habit → habit_id required
--   kick_member              → target_user_id required
--   delete_group             → neither required
CREATE TABLE IF NOT EXISTS proposals (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id       UUID        NOT NULL,
    proposer_id    UUID        NOT NULL,
    type           TEXT        NOT NULL
        CHECK (type IN ('add_habit', 'remove_habit', 'kick_member', 'delete_group')),
    habit_id       UUID        REFERENCES habits(id) ON DELETE CASCADE,
    target_user_id UUID        REFERENCES users(id)  ON DELETE CASCADE,
    status         TEXT        NOT NULL DEFAULT 'open'
        CHECK (status IN ('open', 'approved', 'rejected', 'expired')),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at     TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '7 days',
    CONSTRAINT chk_proposals_habit_required
        CHECK (type NOT IN ('add_habit', 'remove_habit') OR habit_id IS NOT NULL),
    CONSTRAINT chk_proposals_target_required
        CHECK (type != 'kick_member' OR target_user_id IS NOT NULL),
    -- Required so proposal_votes can FK to (id, group_id) and cascade-validate
    -- that voters belong to the same group as the proposal.
    UNIQUE (id, group_id),
    FOREIGN KEY (group_id, proposer_id)
        REFERENCES group_members(group_id, user_id)
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
    FOREIGN KEY (group_id, voter_id)
        REFERENCES group_members(group_id, user_id)
        ON DELETE CASCADE
);

-- =============================================================
-- INDEXES
-- =============================================================

CREATE INDEX IF NOT EXISTS idx_group_members_user
    ON group_members (user_id);

CREATE INDEX IF NOT EXISTS idx_user_habits_group_user
    ON user_habits (group_id, user_id);

CREATE INDEX IF NOT EXISTS idx_user_habits_user
    ON user_habits (user_id);

CREATE INDEX IF NOT EXISTS idx_checkins_user_habit_date
    ON checkins (user_habit_id, checked_on DESC);

CREATE INDEX IF NOT EXISTS idx_checkins_date
    ON checkins (checked_on DESC);

CREATE INDEX IF NOT EXISTS idx_roulette_entries_group_week
    ON roulette_entries (group_id, week_start);

CREATE INDEX IF NOT EXISTS idx_roulette_entries_debtor_week
    ON roulette_entries (debtor_id, week_start);

CREATE INDEX IF NOT EXISTS idx_suggestions_entry
    ON punishment_suggestions (roulette_entry_id);

CREATE INDEX IF NOT EXISTS idx_debts_group_active
    ON debts (group_id, expires_at DESC);

CREATE INDEX IF NOT EXISTS idx_debts_debtor_active
    ON debts (debtor_id, expires_at DESC);

CREATE INDEX IF NOT EXISTS idx_proposals_group_status
    ON proposals (group_id, status, created_at DESC);

-- Only one open proposal per habit type per group at a time.
CREATE UNIQUE INDEX IF NOT EXISTS idx_proposals_one_open_per_habit
    ON proposals (group_id, habit_id, type)
    WHERE status = 'open';

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
-- SUPABASE ROW LEVEL SECURITY
-- =============================================================
-- All product data is accessed exclusively through api-gateway using
-- the service role key (bypasses RLS). The frontend never touches
-- these tables directly — it only uses Supabase for Auth.
-- RLS is enabled with no policies as a defense-in-depth measure.

ALTER TABLE users               ENABLE ROW LEVEL SECURITY;
ALTER TABLE groups              ENABLE ROW LEVEL SECURITY;
ALTER TABLE group_members       ENABLE ROW LEVEL SECURITY;
ALTER TABLE habits              ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_habits         ENABLE ROW LEVEL SECURITY;
ALTER TABLE checkins            ENABLE ROW LEVEL SECURITY;
ALTER TABLE roulette_entries    ENABLE ROW LEVEL SECURITY;
ALTER TABLE punishment_suggestions ENABLE ROW LEVEL SECURITY;
ALTER TABLE debts               ENABLE ROW LEVEL SECURITY;
ALTER TABLE proposals           ENABLE ROW LEVEL SECURITY;
ALTER TABLE proposal_votes      ENABLE ROW LEVEL SECURITY;
