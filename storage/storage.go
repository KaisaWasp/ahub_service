package storage

import (
	"ahub/internal/config"
	"ahub/storage/postgres"
	"ahub/storage/redis"
	"log/slog"
)

type Storage struct {
	Postgres *postgres.Storage
	Redis    *redis.Storage
	Log      *slog.Logger
}

func New(cfg *config.Config, log *slog.Logger) (*Storage, error) {
	var err error

	s := &Storage{
		Log: log,
	}

	s.Postgres, err = postgres.New(cfg.Postgres, log)
	if err != nil {
		return nil, err
	}

	s.Redis, err = redis.New(cfg.Redis, log)
	if err != nil {
		return nil, err
	}

	return s, nil
}
