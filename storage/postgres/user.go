package postgres

import (
	"context"
	"fmt"
	"regexp"

	"github.com/google/uuid"
)

func (s *Storage) CreateUserWithLogin(ctx context.Context, firstName, lastName, login, passwordHash string) (string, error) {
	var email, phone string

	emailRegex := regexp.MustCompile(`^[\w._%+\-]+@[\w.\-]+\.[a-zA-Z]{2,}$`)
	if emailRegex.MatchString(login) {
		email = login
	} else {
		phone = login
	}

	id := uuid.New().String()

	query := `
    INSERT INTO users (id, first_name, last_name, email, phone, password_hash)
    VALUES ($1, $2, $3, $4, $5, $6)
    RETURNING id
    `

	_, err := s.db.ExecContext(ctx, query, id, firstName, lastName, email, phone, passwordHash)
	if err != nil {
		return "", fmt.Errorf("create user: %w", err)
	}

	return id, nil
}
