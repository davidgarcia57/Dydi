package db

import (
	"context"
	"crypto/rand"
	"math/big"
	"time"

	"github.com/dydi/groups-service/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

const inviteChars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func GenerateInviteCode() (string, error) {
	b := make([]byte, 8)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(inviteChars))))
		if err != nil {
			return "", err
		}
		b[i] = inviteChars[n.Int64()]
	}
	return string(b), nil
}

// CreateGroupWithOwner inserts a new group AND the owner membership in a single
// transaction, so a crash between the two can never leave an ownerless group.
func CreateGroupWithOwner(ctx context.Context, pool *pgxpool.Pool, name, inviteCode, createdBy string) (*model.Group, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	g := &model.Group{}
	err = tx.QueryRow(ctx,
		`INSERT INTO groups (name, invite_code, created_by) VALUES ($1, $2, $3)
		 RETURNING id, name, invite_code, created_at`,
		name, inviteCode, createdBy,
	).Scan(&g.ID, &g.Name, &g.InviteCode, &g.CreatedAt)
	if err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx,
		`INSERT INTO memberships (group_id, user_id, role, status)
		 VALUES ($1, $2, 'owner', 'active')`,
		g.ID, createdBy,
	); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return g, nil
}

func GetGroupByID(ctx context.Context, pool *pgxpool.Pool, id string) (*model.Group, error) {
	g := &model.Group{}
	err := pool.QueryRow(ctx,
		`SELECT id, name, invite_code, created_at FROM groups WHERE id = $1`,
		id,
	).Scan(&g.ID, &g.Name, &g.InviteCode, &g.CreatedAt)
	return g, err
}

func GetGroupByInviteCode(ctx context.Context, pool *pgxpool.Pool, code string) (*model.Group, error) {
	g := &model.Group{}
	err := pool.QueryRow(ctx,
		`SELECT id, name, invite_code, created_at FROM groups WHERE invite_code = $1`,
		code,
	).Scan(&g.ID, &g.Name, &g.InviteCode, &g.CreatedAt)
	return g, err
}

func GetMembers(ctx context.Context, pool *pgxpool.Pool, groupID string) ([]model.Member, error) {
	rows, err := pool.Query(ctx,
		`SELECT u.id, u.display_name, u.avatar_url, gm.joined_at
		 FROM memberships gm
		 JOIN users u ON u.id = gm.user_id
		 WHERE gm.group_id = $1 AND gm.status = 'active'
		 ORDER BY gm.joined_at ASC`,
		groupID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make([]model.Member, 0)
	for rows.Next() {
		var m model.Member
		if err := rows.Scan(&m.UserID, &m.DisplayName, &m.AvatarURL, &m.JoinedAt); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func CountMembers(ctx context.Context, pool *pgxpool.Pool, groupID string) (int, error) {
	var count int
	err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM memberships WHERE group_id = $1 AND status = 'active'`,
		groupID,
	).Scan(&count)
	return count, err
}

func IsMember(ctx context.Context, pool *pgxpool.Pool, groupID, userID string) (bool, error) {
	var exists bool
	err := pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM memberships
		               WHERE group_id = $1 AND user_id = $2 AND status = 'active')`,
		groupID, userID,
	).Scan(&exists)
	return exists, err
}

// AddMember inserts a membership (or reactivates a previously left/kicked one).
// role is 'owner' for the group creator and 'member' for everyone else.
func AddMember(ctx context.Context, pool *pgxpool.Pool, groupID, userID, role string) error {
	_, err := pool.Exec(ctx,
		`INSERT INTO memberships (group_id, user_id, role, status)
		 VALUES ($1, $2, $3, 'active')
		 ON CONFLICT (group_id, user_id)
		 DO UPDATE SET status = 'active', left_at = NULL`,
		groupID, userID, role,
	)
	return err
}

// SetMembershipStatus marks a membership as 'left' (voluntary) or 'kicked'
// (by proposal) instead of deleting the row, so the member's history
// (votes, suggestions, debts) stays referentially valid.
func SetMembershipStatus(ctx context.Context, pool *pgxpool.Pool, groupID, userID, status string) error {
	_, err := pool.Exec(ctx,
		`UPDATE memberships
		 SET status = $3,
		     left_at = CASE WHEN $3 = 'active' THEN NULL ELSE NOW() END
		 WHERE group_id = $1 AND user_id = $2`,
		groupID, userID, status,
	)
	return err
}

// GetGroupsByUserID fetches all groups a user belongs to, with their members,
// using a single JOIN query to avoid N+1 database round-trips.
func GetGroupsByUserID(ctx context.Context, pool *pgxpool.Pool, userID string) ([]model.GroupWithMembers, error) {
	rows, err := pool.Query(ctx,
		`SELECT g.id, g.name, g.invite_code, g.created_at,
		        u.id, u.display_name, u.avatar_url, gm2.joined_at
		 FROM groups g
		 JOIN memberships gm1 ON gm1.group_id = g.id AND gm1.user_id = $1 AND gm1.status = 'active'
		 JOIN memberships gm2 ON gm2.group_id = g.id AND gm2.status = 'active'
		 JOIN users u ON u.id = gm2.user_id
		 ORDER BY g.created_at ASC, gm2.joined_at ASC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	grouped := map[string]*model.GroupWithMembers{}
	order := []string{}

	for rows.Next() {
		var g model.Group
		var m model.Member
		var joinedAt time.Time
		if err := rows.Scan(
			&g.ID, &g.Name, &g.InviteCode, &g.CreatedAt,
			&m.UserID, &m.DisplayName, &m.AvatarURL, &joinedAt,
		); err != nil {
			return nil, err
		}
		m.JoinedAt = joinedAt

		if _, ok := grouped[g.ID]; !ok {
			gwm := &model.GroupWithMembers{Group: g, Members: make([]model.Member, 0)}
			grouped[g.ID] = gwm
			order = append(order, g.ID)
		}
		grouped[g.ID].Members = append(grouped[g.ID].Members, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := make([]model.GroupWithMembers, 0, len(order))
	for _, id := range order {
		result = append(result, *grouped[id])
	}
	return result, nil
}
