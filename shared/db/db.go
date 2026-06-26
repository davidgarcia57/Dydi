package db

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewDBPool builds the pgx pool. Tuned for Supabase's Supavisor transaction
// pooler (port 6543): MaxConns is bounded (and tunable per k6 ramp via
// DB_MAX_CONNS) so the four services don't exhaust the shared connection budget,
// and QueryExecModeExec avoids implicit prepared statements, which transaction
// pooling can't keep alive across multiplexed backend connections.
func NewDBPool(ctx context.Context) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, err
	}
	cfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec
	cfg.MaxConns = EnvInt32("DB_MAX_CONNS", 10)
	cfg.MinConns = 0
	cfg.MaxConnIdleTime = time.Minute
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.HealthCheckPeriod = 30 * time.Second
	return pgxpool.NewWithConfig(ctx, cfg)
}

// EnvInt32 reads an environment variable as an int32, returning def if empty or invalid.
func EnvInt32(key string, def int32) int32 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return int32(n)
		}
	}
	return def
}
