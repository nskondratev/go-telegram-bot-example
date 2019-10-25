package pg

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
)

func New(ctx context.Context, connStr string, maxConnections int32) (*pgxpool.Pool, error) {
	cnf, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PG config: %w", err)
	}
	cnf.MaxConns = maxConnections
	pool, err := pgxpool.ConnectConfig(ctx, cnf)
	if err != nil {
		return nil, err
	}
	return pool, nil
}
