package postgres

import (
	"context"
	"regexp"

	"github.com/google/uuid"
)

func (s *Storage) CreateUserWithLogin(
	ctx context.Context,
	firstName, lastName, login, passwordHash string,
) (string, error) {

	var email *string
	var phone *string

	emailRegex := regexp.MustCompile(`^[\w._%+\-]+@[\w.\-]+\.[a-zA-Z]{2,}$`)
	if emailRegex.MatchString(login) {
		email = &login
	} else {
		phone = &login
	}

	user := User{
		ID:           uuid.New().String(),
		FirstName:    firstName,
		LastName:     lastName,
		Email:        email,
		Phone:        phone,
		PasswordHash: passwordHash,
	}

	if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
		return "", err
	}

	return user.ID, nil
}
