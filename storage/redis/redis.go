package redis

import (
	"context"
	"fmt"
	"time"

	"ahub/internal/config"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

type Storage struct {
	Client *redis.Client
	Log    *slog.Logger
}

func New(cfg config.RedisConfig, log *slog.Logger) (*Storage, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	return &Storage{
		Client: rdb,
		Log:    log,
	}, nil
}

func (s *Storage) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	return s.Client.Set(ctx, key, value, ttl).Err()
}

func (s *Storage) Get(ctx context.Context, key string) (string, error) {
	val, err := s.Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil // ключ не существует
	}
	return val, err
}
