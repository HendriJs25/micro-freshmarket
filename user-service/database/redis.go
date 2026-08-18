package database

import (
	"context"
	"fmt"
	"net"
	"time"
	"user-service/config"

	redislib "github.com/redis/go-redis/v9"
)

const redisPingTimeout = 5 * time.Second

type Redis struct {
	Client *redislib.Client
}

func NewRedis(cfg config.Redis) (*Redis, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate redis configuration: %w", err)
	}

	client := redislib.NewClient(&redislib.Options{
		Addr:     net.JoinHostPort(cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), redisPingTimeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &Redis{
		Client: client,
	}, nil
}

func (r *Redis) Close() error {
	return r.Client.Close()
}
