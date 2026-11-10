package redisdb

import (
	"context"
	"fmt"
	"net"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Host string
	Port string
	User string
	Pass string
	DB   int
}

func NewRedis(ctx context.Context, cfg *RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     net.JoinHostPort(cfg.Host, cfg.Port),
		Username: cfg.User,
		Password: cfg.Pass,
		DB:       cfg.DB,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("client.Ping: %w", err)
	}

	return client, nil
}
