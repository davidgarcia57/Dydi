package db

import (
	"context"

	"github.com/dydi/groups-service/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

func UpsertUser(ctx context.Context, pool *pgxpool.Pool, id, displayName string, avatarURL *string) (*model.User, error) {
	u := &model.User{}
	err := pool.QueryRow(ctx,
		`INSERT INTO users (id, display_name, avatar_url)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (id) DO UPDATE
		   SET display_name = EXCLUDED.display_name,
		       avatar_url   = EXCLUDED.avatar_url
		 RETURNING id, display_name, avatar_url, created_at`,
		id, displayName, avatarURL,
	).Scan(&u.ID, &u.DisplayName, &u.AvatarURL, &u.CreatedAt)
	return u, err
}
