-- Dydi initial schema
-- Source of truth: supabase/migrations/001_initial.sql
-- Apply via Supabase dashboard or supabase CLI (supabase db push)

-- users: extended profile (auth identity lives in Supabase Auth)
CREATE TABLE IF NOT EXISTS users (
    id          UUID PRIMARY KEY,  -- matches Supabase Auth user id
    display_name TEXT NOT NULL,
    avatar_url  TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- groups
CREATE TABLE IF NOT EXISTS groups (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    invite_code TEXT NOT NULL UNIQUE,
    owner_id    UUID NOT NULL REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- group_members
CREATE TABLE IF NOT EXISTS group_members (
    group_id    UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (group_id, user_id)
);

-- habits catalog
CREATE TABLE IF NOT EXISTS habits (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    description TEXT,
    icon        TEXT
);

-- user_habits: which habit a user tracks in a group
CREATE TABLE IF NOT EXISTS user_habits (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id),
    group_id    UUID NOT NULL REFERENCES groups(id),
    habit_id    UUID NOT NULL REFERENCES habits(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, group_id, habit_id)
);

-- checkins: daily habit completions
CREATE TABLE IF NOT EXISTS checkins (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_habit_id   UUID NOT NULL REFERENCES user_habits(id),
    checked_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    note            TEXT
);

-- debts: penalties assigned after roulette spin
CREATE TABLE IF NOT EXISTS debts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id        UUID NOT NULL REFERENCES groups(id),
    debtor_id       UUID NOT NULL REFERENCES users(id),
    punishment_text TEXT NOT NULL,
    punishment_emoji TEXT,
    resolved        BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at     TIMESTAMPTZ
);
