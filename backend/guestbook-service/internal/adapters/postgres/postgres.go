package postgresdb

import (
	"context"
	"fmt"
	"net"
	"net/url"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresConfig struct {
	DB   string
	User string
	Pass string
	Port string
	Host string
}

func NewPool(ctx context.Context, cfg *PostgresConfig) (*pgxpool.Pool, error) {
	conn := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.User, cfg.Pass),
		Host:   net.JoinHostPort(cfg.Host, cfg.Port),
		Path:   cfg.DB,
	}

	pool, err := pgxpool.New(ctx, conn.String())
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("pool.Ping: %w", err)
	}

	return pool, nil
}

func ClosePool(pool *pgxpool.Pool) {
	pool.Close()
}
