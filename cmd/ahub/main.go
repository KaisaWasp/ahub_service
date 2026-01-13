package main

import (
	"ahub/internal/auth"
	"ahub/internal/config"
	"ahub/internal/migrations"
	storagebd "ahub/storage"
	"fmt"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
)

const (
	envLocal = "local"
	envDev   = "dev"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("start", slog.String("env", cfg.Env))

	migrationsPath := "file://D:/MyProjects/AHUB/migrations"
	pg := cfg.Postgres
	dbURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		pg.User, pg.Password, pg.Host, pg.Port, pg.DBName, pg.SSLMode,
	)

	migrations.RunMigrations(migrationsPath, dbURL)

	storage, err := storagebd.New(cfg, log)
	if err != nil {
		log.Error("failed to initialize storage:", err)
		return
	}

	authStorage := auth.NewStorage(storage)

	jwtManager := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWTTTLDuration())

	authService := auth.NewAuthService(authStorage, cfg.Redis.TTLDuration(), jwtManager)
	authHandler := auth.NewHandler(authService)

	r := gin.Default()

	auth.RegisterRoutes(r, authHandler, jwtManager)

	if err := r.Run(cfg.HTTPServer.Address); err != nil {
		log.Error("failed to run server:", err)
	}
}

func setupLogger(env string) *slog.Logger {
	var handler slog.Handler

	switch env {
	case envLocal:
		handler = slog.NewTextHandler(
			os.Stdout,
			&slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		)

	case envDev:
		handler = slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{
				Level: slog.LevelInfo,
			},
		)

	default:
		handler = slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{
				Level: slog.LevelInfo,
			},
		)
	}

	logger := slog.New(handler)

	slog.SetDefault(logger)

	return logger
}
