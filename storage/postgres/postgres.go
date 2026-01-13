package postgres

import (
	"ahub/internal/config"
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Storage struct {
	db  *sql.DB
	log *slog.Logger
}

type RefreshTokenData struct {
	UserID    string
	ExpiresAt time.Time
}

type UserInfo struct {
	ID           string
	Email        sql.NullString
	Phone        sql.NullString
	PasswordHash string
}

func New(cfg config.PostgresConfig, log *slog.Logger) (*Storage, error) {
	op := "storage.postgres.New"

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DBName,
		cfg.SSLMode,
	)

	fmt.Println("Postgres DSN:", dsn)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: open db: %w", op, err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("%s: ping db: %w", op, err)
	}

	log.Info("postgres connected")

	return &Storage{
		db:  db,
		log: log,
	}, nil
}

func (s *Storage) SaveRefreshToken(
	ctx context.Context,
	userID string,
	token string,
	expiresAt time.Time,
) error {

	query := `
		INSERT INTO refresh_tokens (token, user_id, expires_at)
		VALUES ($1, $2, $3)
	`

	_, err := s.db.ExecContext(ctx, query, token, userID, expiresAt)
	if err != nil {
		return fmt.Errorf("save refresh token: %w", err)
	}

	return nil
}

func (s *Storage) GetRefreshToken(ctx context.Context, token string) (*RefreshTokenData, error) {
	query := `
		SELECT user_id, expires_at
		FROM refresh_tokens
		WHERE token = $1
	`

	var userID string
	var expiresAt time.Time

	err := s.db.QueryRowContext(ctx, query, token).Scan(&userID, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("refresh token not found")
		}
		return nil, fmt.Errorf("get refresh token: %w", err)
	}

	return &RefreshTokenData{
		UserID:    userID,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *Storage) GetUserByLogin(ctx context.Context, login string) (*UserInfo, error) {
	const query = `
		SELECT id, email, phone, password_hash
		FROM users
		WHERE email = $1 OR phone = $1
		LIMIT 1
	`

	var user UserInfo

	err := s.db.QueryRowContext(ctx, query, login).Scan(
		&user.ID,
		&user.Email,
		&user.Phone,
		&user.PasswordHash,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("get user by login: %w", err)
	}

	return &user, nil
}

func (s *Storage) DeleteRefreshToken(ctx context.Context, token string) error {
	query := `
		DELETE FROM refresh_tokens
		WHERE token = $1
	`

	result, err := s.db.ExecContext(ctx, query, token)
	if err != nil {
		return fmt.Errorf("delete refresh token: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("refresh token not found")
	}

	return nil
}
