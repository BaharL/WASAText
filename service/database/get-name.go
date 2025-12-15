package database

import (
	"context"
	"database/sql"
)

func (db *appdbimpl) GetName(ctx context.Context, identifier string) (string, error) {
	var name string
	err := db.c.QueryRowContext(ctx, `
		SELECT name
		FROM users
		WHERE identifier = ?
	`, identifier).Scan(&name)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", sql.ErrNoRows
		}
		return "", err
	}

	return name, nil
}
