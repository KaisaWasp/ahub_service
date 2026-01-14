package postgres

import (
	"ahub/internal/config"
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Storage struct {
	db  *gorm.DB
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

type RefreshToken struct {
	Token     string    `gorm:"primaryKey;column:token"`
	UserID    string    `gorm:"column:user_id;not null"`
	ExpiresAt time.Time `gorm:"column:expires_at;not null"`
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

type User struct {
	ID           string  `gorm:"column:id;primaryKey"`
	FirstName    string  `gorm:"column:first_name;not null"`
	LastName     string  `gorm:"column:last_name;not null"`
	Email        *string `gorm:"column:email"`
	Phone        *string `gorm:"column:phone"`
	PasswordHash string  `gorm:"column:password_hash;not null"`
}

func (User) TableName() string {
	return "users"
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

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("%s: open db: %w", op, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("%s: get sql db: %w", op, err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

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

	rt := RefreshToken{
		Token:     token,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}

	if err := s.db.WithContext(ctx).Create(&rt).Error; err != nil {
		return fmt.Errorf("save refresh token: %w", err)
	}

	return nil
}

func (s *Storage) GetRefreshToken(
	ctx context.Context,
	token string,
) (*RefreshTokenData, error) {

	var rt RefreshToken

	err := s.db.WithContext(ctx).
		Where("token = ?", token).
		First(&rt).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("refresh token not found")
		}
		return nil, fmt.Errorf("get refresh token: %w", err)
	}

	return &RefreshTokenData{
		UserID:    rt.UserID,
		ExpiresAt: rt.ExpiresAt,
	}, nil
}

func (s *Storage) GetUserByLogin(
	ctx context.Context,
	login string,
) (*UserInfo, error) {

	var user User

	err := s.db.WithContext(ctx).
		Where("email = ? OR phone = ?", login, login).
		Limit(1).
		First(&user).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("get user by login: %w", err)
	}

	return &UserInfo{
		ID:           user.ID,
		Email:        toNullString(user.Email),
		Phone:        toNullString(user.Phone),
		PasswordHash: user.PasswordHash,
	}, nil
}

func (s *Storage) DeleteRefreshToken(
	ctx context.Context,
	token string,
) error {

	result := s.db.WithContext(ctx).
		Where("token = ?", token).
		Delete(&RefreshToken{})

	if result.Error != nil {
		return fmt.Errorf("delete refresh token: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("refresh token not found")
	}

	return nil
}

func toNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{
		String: *s,
		Valid:  true,
	}
}
