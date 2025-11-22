package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gofrs/uuid"
)

// LoginOrCreateUser handles user login/creation.
// If a user with the given name exists, it returns its identifier.
// If not, it creates a new user with a fresh UUID identifier.
func (db *appdbimpl) LoginOrCreateUser(ctx context.Context, name string) (string, error) {
	// 1. Check whether the user already exists.
	var identifier string
	err := db.c.QueryRowContext(ctx, `
		SELECT identifier FROM users WHERE name = ?
	`, name).Scan(&identifier)

	if err == nil {
		// User found → return existing identifier.
		return identifier, nil
	}

	if err != sql.ErrNoRows {
		// Unexpected DB error.
		return "", fmt.Errorf("query error: %w", err)
	}

	// 2. User does not exist → create a new one.
	id, err := uuid.NewV4()
	if err != nil {
		return "", fmt.Errorf("cannot generate uuid: %w", err)
	}
	newID := id.String()

	_, err = db.c.ExecContext(ctx, `
		INSERT INTO users (name, identifier) VALUES (?, ?)
	`, name, newID)
	if err != nil {
		return "", fmt.Errorf("cannot create user: %w", err)
	}

	return newID, nil
}
