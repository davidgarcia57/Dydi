package db

import (
	"context"

	"github.com/dydi/groups-service/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func scanProposal(row pgx.Row) (*model.Proposal, error) {
	p := &model.Proposal{}
	err := row.Scan(
		&p.ID, &p.GroupID, &p.ProposerID, &p.Type,
		&p.HabitID, &p.TargetUserID,
		&p.Status, &p.CreatedAt, &p.ExpiresAt,
		&p.VoteCount, &p.MemberCount,
	)
	return p, err
}

// CreateProposal inserts a new proposal and freezes its electorate (the set of
// active members at this moment) into proposal_eligible_voters, so the majority
// threshold cannot drift as people join or leave mid-vote. Runs in a transaction
// so the proposal and its electorate are created atomically. member_count is the
// frozen electorate size.
func CreateProposal(ctx context.Context, pool *pgxpool.Pool, groupID, proposerID string, pType model.ProposalType, habitID, targetUserID *string) (*model.Proposal, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Lazily close any proposal whose deadline has passed but is still 'open'.
	// Without this, the partial unique indexes (uq_proposals_one_open_*) would
	// permanently block re-proposing the same habit/kick/delete once a prior
	// proposal expired unresolved — there is no background expiry job.
	if _, err := tx.Exec(ctx,
		`UPDATE proposals SET status = 'expired', resolved_at = NOW()
		 WHERE group_id = $1 AND status = 'open' AND expires_at <= NOW()`,
		groupID,
	); err != nil {
		return nil, err
	}

	var id string
	if err := tx.QueryRow(ctx,
		`INSERT INTO proposals (group_id, proposer_id, type, habit_id, target_user_id)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		groupID, proposerID, pType, habitID, targetUserID,
	).Scan(&id); err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx,
		`INSERT INTO proposal_eligible_voters (proposal_id, group_id, user_id)
		 SELECT $1, $2, user_id FROM memberships
		 WHERE group_id = $2 AND status = 'active'`,
		id, groupID,
	); err != nil {
		return nil, err
	}

	p, err := scanProposal(tx.QueryRow(ctx,
		`SELECT p.id, p.group_id, p.proposer_id, p.type, p.habit_id, p.target_user_id,
		        p.status, p.created_at, p.expires_at,
		        COUNT(pv.voter_id) AS vote_count,
		        (SELECT COUNT(*) FROM proposal_eligible_voters WHERE proposal_id = p.id) AS member_count
		 FROM proposals p
		 LEFT JOIN proposal_votes pv ON pv.proposal_id = p.id AND pv.approved = TRUE
		 WHERE p.id = $1
		 GROUP BY p.id`,
		id,
	))
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return p, nil
}

// IsEligibleVoter reports whether a user belongs to a proposal's frozen electorate.
func IsEligibleVoter(ctx context.Context, pool *pgxpool.Pool, proposalID, userID string) (bool, error) {
	var exists bool
	err := pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM proposal_eligible_voters
		               WHERE proposal_id = $1 AND user_id = $2)`,
		proposalID, userID,
	).Scan(&exists)
	return exists, err
}

func GetProposal(ctx context.Context, pool *pgxpool.Pool, proposalID string) (*model.Proposal, error) {
	return scanProposal(pool.QueryRow(ctx,
		`SELECT p.id, p.group_id, p.proposer_id, p.type, p.habit_id, p.target_user_id,
		        p.status, p.created_at, p.expires_at,
		        COUNT(pv.voter_id) AS vote_count,
		        (SELECT COUNT(*) FROM proposal_eligible_voters WHERE proposal_id = p.id) AS member_count
		 FROM proposals p
		 LEFT JOIN proposal_votes pv ON pv.proposal_id = p.id AND pv.approved = TRUE
		 WHERE p.id = $1
		 GROUP BY p.id`,
		proposalID,
	))
}

func ListOpenProposals(ctx context.Context, pool *pgxpool.Pool, groupID, userID string) ([]model.Proposal, error) {
	rows, err := pool.Query(ctx,
		`SELECT p.id, p.group_id, p.proposer_id, p.type, p.habit_id, p.target_user_id,
		        p.status, p.created_at, p.expires_at,
		        COUNT(pv.voter_id) AS vote_count,
		        (SELECT COUNT(*) FROM proposal_eligible_voters WHERE proposal_id = p.id) AS member_count,
		        EXISTS(SELECT 1 FROM proposal_votes WHERE proposal_id = p.id AND voter_id = $2) AS user_voted
		 FROM proposals p
		 LEFT JOIN proposal_votes pv ON pv.proposal_id = p.id AND pv.approved = TRUE
		 WHERE p.group_id = $1 AND p.status = 'open' AND p.expires_at > NOW()
		 GROUP BY p.id
		 ORDER BY p.created_at DESC`,
		groupID, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	proposals := make([]model.Proposal, 0)
	for rows.Next() {
		var p model.Proposal
		if err := rows.Scan(
			&p.ID, &p.GroupID, &p.ProposerID, &p.Type, &p.HabitID, &p.TargetUserID,
			&p.Status, &p.CreatedAt, &p.ExpiresAt, &p.VoteCount, &p.MemberCount,
			&p.UserVoted,
		); err != nil {
			return nil, err
		}
		proposals = append(proposals, p)
	}
	return proposals, rows.Err()
}

// ListResolvedProposals returns the group's closed proposals (newest first,
// capped at 30) so the squad can audit past decisions. An 'open' proposal past
// its deadline is presented as 'expired' even if no vote closed it lazily yet.
func ListResolvedProposals(ctx context.Context, pool *pgxpool.Pool, groupID, userID string) ([]model.Proposal, error) {
	rows, err := pool.Query(ctx,
		`SELECT p.id, p.group_id, p.proposer_id, p.type, p.habit_id, p.target_user_id,
		        CASE WHEN p.status = 'open' AND p.expires_at <= NOW() THEN 'expired'
		             ELSE p.status END AS status,
		        p.created_at, p.expires_at,
		        COUNT(pv.voter_id) AS vote_count,
		        (SELECT COUNT(*) FROM proposal_eligible_voters WHERE proposal_id = p.id) AS member_count,
		        EXISTS(SELECT 1 FROM proposal_votes WHERE proposal_id = p.id AND voter_id = $2) AS user_voted
		 FROM proposals p
		 LEFT JOIN proposal_votes pv ON pv.proposal_id = p.id AND pv.approved = TRUE
		 WHERE p.group_id = $1 AND (p.status <> 'open' OR p.expires_at <= NOW())
		 GROUP BY p.id
		 ORDER BY p.created_at DESC
		 LIMIT 30`,
		groupID, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	proposals := make([]model.Proposal, 0)
	for rows.Next() {
		var p model.Proposal
		if err := rows.Scan(
			&p.ID, &p.GroupID, &p.ProposerID, &p.Type, &p.HabitID, &p.TargetUserID,
			&p.Status, &p.CreatedAt, &p.ExpiresAt, &p.VoteCount, &p.MemberCount,
			&p.UserVoted,
		); err != nil {
			return nil, err
		}
		proposals = append(proposals, p)
	}
	return proposals, rows.Err()
}

func HasVoted(ctx context.Context, pool *pgxpool.Pool, proposalID, voterID string) (bool, error) {
	var exists bool
	err := pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM proposal_votes WHERE proposal_id = $1 AND voter_id = $2)`,
		proposalID, voterID,
	).Scan(&exists)
	return exists, err
}

// CastVote inserts a vote. group_id is looked up from the proposal so the
// composite FK to memberships is satisfied. The (proposal_id, voter_id) FK to
// proposal_eligible_voters guarantees only the frozen electorate can vote.
func CastVote(ctx context.Context, pool *pgxpool.Pool, proposalID, voterID string, approved bool) error {
	_, err := pool.Exec(ctx,
		`INSERT INTO proposal_votes (proposal_id, group_id, voter_id, approved)
		 SELECT $1, group_id, $2, $3 FROM proposals WHERE id = $1`,
		proposalID, voterID, approved,
	)
	return err
}

// CountApprovalVotes returns the current yes-vote count and the frozen electorate
// size for the proposal. Used to check quorum: yes_votes * 2 >= member_count.
func CountApprovalVotes(ctx context.Context, pool *pgxpool.Pool, proposalID string) (approvals, members int, err error) {
	err = pool.QueryRow(ctx,
		`SELECT
		     (SELECT COUNT(*) FROM proposal_votes WHERE proposal_id = $1 AND approved = TRUE),
		     (SELECT COUNT(*) FROM proposal_eligible_voters WHERE proposal_id = $1)`,
		proposalID,
	).Scan(&approvals, &members)
	return
}

// SetProposalStatus updates a proposal's status. When the status leaves 'open',
// resolved_at is stamped (required by the schema's resolved-state CHECK) and
// resolved_by records who triggered it (nil for system/expiry resolutions).
func SetProposalStatus(ctx context.Context, pool *pgxpool.Pool, proposalID string, status model.ProposalStatus, resolvedBy *string) error {
	_, err := pool.Exec(ctx,
		`UPDATE proposals
		 SET status = $1,
		     resolved_at = CASE WHEN $1 = 'open' THEN NULL ELSE NOW() END,
		     resolved_by = CASE WHEN $1 = 'open' THEN NULL ELSE $3 END
		 WHERE id = $2`,
		status, proposalID, resolvedBy,
	)
	return err
}
