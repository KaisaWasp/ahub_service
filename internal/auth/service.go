package auth

import (
	_ "ahub/storage"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	storage *AuthStorage
	otpTTL  time.Duration
	jwt     *JWTManager
}

func NewAuthService(storage *AuthStorage, otpTTL time.Duration, jwtManager *JWTManager) *AuthService {
	return &AuthService{storage: storage, otpTTL: otpTTL, jwt: jwtManager}
}

func (s *AuthService) StartRegistration(ctx context.Context, firstName, lastName, login, password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	//otp := generationOTP()
	otp := "123456"

	token, err := s.storage.SaveRegistration(ctx, RegistrationData{
		FirstName:    firstName,
		LastName:     lastName,
		Login:        login,
		PasswordHash: string(hash),
		OTP:          otp,
	}, s.otpTTL)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *AuthService) ConfirmRegistration(ctx context.Context, token, code string) (string, string, error) {
	data, err := s.storage.GetRegistration(ctx, token)
	if err != nil {
		return "", "", err
	}

	if code != data.OTP {
		return "", "", errors.New("invalid confirmation code")
	}

	userId, err := s.storage.CreateUser(ctx, *data)
	if err != nil {
		return "", "", err
	}

	if err := s.storage.DeleteRegistration(ctx, token); err != nil {
		return "", "", err
	}

	accessToken, err := s.jwt.GenerateAccessToken(userId)
	if err != nil {
		return "", "", err
	}

	refreshToken := uuid.NewString()
	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	if err := s.storage.SaveRefreshToken(ctx, userId, refreshToken, expiresAt); err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) Refresh(ctx context.Context, oldRefreshToken string) (string, string, error) {
	tokenData, err := s.storage.GetRefreshToken(ctx, oldRefreshToken)
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}

	if time.Now().After(tokenData.ExpiresAt) {
		return "", "", errors.New("refresh token expired")
	}

	if err := s.storage.DeleteRefreshToken(ctx, oldRefreshToken); err != nil {
		return "", "", err
	}

	newRefreshToken := uuid.NewString()
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	if err := s.storage.SaveRefreshToken(ctx, tokenData.UserID, newRefreshToken, expiresAt); err != nil {
		return "", "", err
	}

	accessToken, err := s.jwt.GenerateAccessToken(tokenData.UserID)
	if err != nil {
		return "", "", err
	}

	return accessToken, newRefreshToken, nil
}

func (s *AuthService) Login(ctx context.Context, login, password string) (string, string, error) {
	user, err := s.storage.bd.Postgres.GetUserByLogin(ctx, login)
	if err != nil {
		return "", "", errors.New("invalid login or password")
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(password),
	); err != nil {
		return "", "", errors.New("invalid login or password")
	}

	accessToken, err := s.jwt.GenerateAccessToken(user.ID)
	if err != nil {
		return "", "", err
	}

	refreshToken := uuid.NewString()
	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	if err := s.storage.SaveRefreshToken(ctx, user.ID, refreshToken, expiresAt); err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func generationOTP() string {
	const digits = "0123456789"
	otp := make([]byte, 6)
	for i := range otp {
		otp[i] = digits[randomInt(len(digits))]
	}

	return string(otp)
}

func randomInt(max int) int {
	return int(time.Now().UnixNano() % int64(max))
}
