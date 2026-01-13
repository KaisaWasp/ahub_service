package auth

import (
	"ahub/storage"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type AuthStorage struct {
	bd *storage.Storage
}

type RefreshTokenData struct {
	UserID    string
	ExpiresAt time.Time
}

func NewStorage(s *storage.Storage) *AuthStorage {
	return &AuthStorage{bd: s}
}

type RegistrationData struct {
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Login        string `json:"login"`
	PasswordHash string `json:"password_hash"`
	OTP          string `json:"otp"`
}

func (s *AuthStorage) SaveRegistration(ctx context.Context, data RegistrationData, ttl time.Duration) (string, error) {
	if s == nil {
		return "", fmt.Errorf("AuthStorage is nil")
	}
	if s.bd == nil {
		return "", fmt.Errorf("storage is nil")
	}
	if s.bd.Redis == nil {
		return "", fmt.Errorf("redis storage is nil")
	}
	if s.bd.Redis.Client == nil {
		return "", fmt.Errorf("redis client is nil")
	}

	token := uuid.NewString()
	key := fmt.Sprintf("registration:%s", token)

	payload, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	if err := s.bd.Redis.Client.Set(ctx, key, payload, ttl).Err(); err != nil {
		return "", err
	}

	val, err := s.bd.Redis.Client.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	fmt.Println("REDIS VALUE:", val)

	return token, nil
}

func (s *AuthStorage) GetRegistration(ctx context.Context, token string) (*RegistrationData, error) {
	key := fmt.Sprintf("registration:%s", token)

	val, err := s.bd.Redis.Client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var data RegistrationData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, err
	}

	return &data, err
}

func (s *AuthStorage) DeleteRegistration(ctx context.Context, token string) error {
	key := fmt.Sprintf("registration:%s", token)
	return s.bd.Redis.Client.Del(ctx, key).Err()
}

func (s *AuthStorage) CreateUser(ctx context.Context, data RegistrationData) (string, error) {
	return s.bd.Postgres.CreateUserWithLogin(ctx, data.FirstName, data.LastName, data.Login, data.PasswordHash)
}

func (s *AuthStorage) SaveRefreshToken(
	ctx context.Context,
	userID string,
	refreshToken string,
	expiresAt time.Time,
) error {

	if s == nil || s.bd == nil || s.bd.Postgres == nil {
		return fmt.Errorf("postgres storage is nil")
	}

	return s.bd.Postgres.SaveRefreshToken(ctx, userID, refreshToken, expiresAt)
}

func (s *AuthStorage) GetRefreshToken(ctx context.Context, token string) (*RefreshTokenData, error) {
	if s == nil || s.bd == nil || s.bd.Postgres == nil {
		return nil, fmt.Errorf("postgres storage is nil")
	}

	data, err := s.bd.Postgres.GetRefreshToken(ctx, token)
	if err != nil {
		return nil, err
	}

	return (*RefreshTokenData)(data), nil
}

func (s *AuthStorage) DeleteRefreshToken(ctx context.Context, token string) error {
	if s == nil || s.bd == nil || s.bd.Postgres == nil {
		return fmt.Errorf("postgres storage is nil")
	}

	return s.bd.Postgres.DeleteRefreshToken(ctx, token)
}
