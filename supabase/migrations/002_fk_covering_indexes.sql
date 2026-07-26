-- Cover every foreign key whose column order is not already backed by a
-- primary key, unique constraint, or existing index. These indexes keep
-- referential checks and cascading deletes predictable as data grows.

CREATE INDEX IF NOT EXISTS idx_groups_created_by
    ON groups (created_by);

CREATE INDEX IF NOT EXISTS idx_debts_group_debtor
    ON debts (group_id, debtor_id);

CREATE INDEX IF NOT EXISTS idx_debts_entry_group
    ON debts (roulette_entry_id, group_id);

CREATE INDEX IF NOT EXISTS idx_suggestions_entry_group
    ON punishment_suggestions (roulette_entry_id, group_id);

CREATE INDEX IF NOT EXISTS idx_proposals_habit
    ON proposals (habit_id);

CREATE INDEX IF NOT EXISTS idx_proposals_resolved_by
    ON proposals (resolved_by);

CREATE INDEX IF NOT EXISTS idx_eligible_voters_proposal_group
    ON proposal_eligible_voters (proposal_id, group_id);

CREATE INDEX IF NOT EXISTS idx_proposal_votes_proposal_group
    ON proposal_votes (proposal_id, group_id);
