package db

import (
	"context"
	"encoding/json"

	"github.com/dydi/groups-service/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateProposal(ctx context.Context, pool *pgxpool.Pool, groupID, proposerID string, pType model.ProposalType, payload json.RawMessage) (*model.Proposal, error) {
	p := &model.Proposal{}
	var rawPayload []byte
	err := pool.QueryRow(ctx,
		`INSERT INTO proposals (group_id, proposer_id, type, payload)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, group_id, proposer_id, type, payload, status, created_at, expires_at`,
		groupID, proposerID, pType, payload,
	).Scan(&p.ID, &p.GroupID, &p.ProposerID, &p.Type, &rawPayload, &p.Status, &p.CreatedAt, &p.ExpiresAt)
	if err != nil {
		return nil, err
	}
	p.Payload = rawPayload
	return p, nil
}

func GetProposal(ctx context.Context, pool *pgxpool.Pool, proposalID string) (*model.Proposal, error) {
	p := &model.Proposal{}
	var rawPayload []byte
	err := pool.QueryRow(ctx,
		`SELECT id, group_id, proposer_id, type, payload, status, created_at, expires_at
		 FROM proposals WHERE id = $1`,
		proposalID,
	).Scan(&p.ID, &p.GroupID, &p.ProposerID, &p.Type, &rawPayload, &p.Status, &p.CreatedAt, &p.ExpiresAt)
	if err != nil {
		return nil, err
	}
	p.Payload = rawPayload
	return p, nil
}

func ListOpenProposals(ctx context.Context, pool *pgxpool.Pool, groupID string) ([]model.Proposal, error) {
	rows, err := pool.Query(ctx,
		`SELECT p.id, p.group_id, p.proposer_id, p.type, p.payload, p.status, p.created_at, p.expires_at,
		        COUNT(pv.voter_id) AS vote_count,
		        (SELECT COUNT(*) FROM user_groups WHERE group_id = p.group_id) AS member_count
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

	var proposals []model.Proposal
	for rows.Next() {
		var p model.Proposal
		var rawPayload []byte
		if err := rows.Scan(&p.ID, &p.GroupID, &p.ProposerID, &p.Type, &rawPayload,
			&p.Status, &p.CreatedAt, &p.ExpiresAt, &p.VoteCount, &p.MemberCount); err != nil {
			return nil, err
		}
		p.Payload = rawPayload
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

func CastVote(ctx context.Context, pool *pgxpool.Pool, proposalID, voterID string, approved bool) error {
	_, err := pool.Exec(ctx,
		`INSERT INTO proposal_votes (proposal_id, voter_id, approved) VALUES ($1, $2, $3)`,
		proposalID, voterID, approved,
	)
	return err
}

func CountApprovalVotes(ctx context.Context, pool *pgxpool.Pool, proposalID string) (approvals, members int, err error) {
	err = pool.QueryRow(ctx,
		`SELECT
		   (SELECT COUNT(*) FROM proposal_votes WHERE proposal_id = $1 AND approved = TRUE),
		   (SELECT COUNT(*) FROM user_groups
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
