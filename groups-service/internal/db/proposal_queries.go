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

// CreateProposal inserts a new proposal and returns it with current vote/member counts.
// habitID and targetUserID are nil for proposal types that don't need them.
func CreateProposal(ctx context.Context, pool *pgxpool.Pool, groupID, proposerID string, pType model.ProposalType, habitID, targetUserID *string) (*model.Proposal, error) {
	return scanProposal(pool.QueryRow(ctx,
		`INSERT INTO proposals (group_id, proposer_id, type, habit_id, target_user_id)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING
		     id, group_id, proposer_id, type, habit_id, target_user_id,
		     status, created_at, expires_at,
		     0 AS vote_count,
		     (SELECT COUNT(*) FROM group_members WHERE group_id = $1) AS member_count`,
		groupID, proposerID, pType, habitID, targetUserID,
	))
}

func GetProposal(ctx context.Context, pool *pgxpool.Pool, proposalID string) (*model.Proposal, error) {
	return scanProposal(pool.QueryRow(ctx,
		`SELECT p.id, p.group_id, p.proposer_id, p.type, p.habit_id, p.target_user_id,
		        p.status, p.created_at, p.expires_at,
		        COUNT(pv.voter_id) AS vote_count,
		        (SELECT COUNT(*) FROM group_members WHERE group_id = p.group_id) AS member_count
		 FROM proposals p
		 LEFT JOIN proposal_votes pv ON pv.proposal_id = p.id AND pv.approved = TRUE
		 WHERE p.id = $1
		 GROUP BY p.id`,
		proposalID,
	))
}

func ListOpenProposals(ctx context.Context, pool *pgxpool.Pool, groupID string) ([]model.Proposal, error) {
	rows, err := pool.Query(ctx,
		`SELECT p.id, p.group_id, p.proposer_id, p.type, p.habit_id, p.target_user_id,
		        p.status, p.created_at, p.expires_at,
		        COUNT(pv.voter_id) AS vote_count,
		        (SELECT COUNT(*) FROM group_members WHERE group_id = p.group_id) AS member_count
		 FROM proposals p
		 LEFT JOIN proposal_votes pv ON pv.proposal_id = p.id AND pv.approved = TRUE
		 WHERE p.group_id = $1 AND p.status = 'open' AND p.expires_at > NOW()
		 GROUP BY p.id
		 ORDER BY p.created_at DESC`,
		groupID,
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
// composite FK to group_members is satisfied.
func CastVote(ctx context.Context, pool *pgxpool.Pool, proposalID, voterID string, approved bool) error {
	_, err := pool.Exec(ctx,
		`INSERT INTO proposal_votes (proposal_id, group_id, voter_id, approved)
		 SELECT $1, group_id, $2, $3 FROM proposals WHERE id = $1`,
		proposalID, voterID, approved,
	)
	return err
}

// CountApprovalVotes returns the current yes-vote count and current member count
// for the proposal's group. Used to check quorum: yes_votes * 2 >= member_count.
func CountApprovalVotes(ctx context.Context, pool *pgxpool.Pool, proposalID string) (approvals, members int, err error) {
	err = pool.QueryRow(ctx,
		`SELECT
		     (SELECT COUNT(*) FROM proposal_votes WHERE proposal_id = $1 AND approved = TRUE),
		     (SELECT COUNT(*) FROM group_members
		      WHERE group_id = (SELECT group_id FROM proposals WHERE id = $1))`,
		proposalID,
	).Scan(&approvals, &members)
	return
}

func SetProposalStatus(ctx context.Context, pool *pgxpool.Pool, proposalID string, status model.ProposalStatus) error {
	_, err := pool.Exec(ctx,
		`UPDATE proposals SET status = $1 WHERE id = $2`,
		status, proposalID,
	)
	return err
}
