package database

import (
	"context"
	"database/sql"
	"fmt"
)

// SetUserName updates the user's name given their identifier (Bearer token value).
func (db *appdbimpl) SetUserName(ctx context.Context, identifier string, newName string) error {
	res, err := db.c.ExecContext(ctx, `
		UPDATE users
		SET name = ?
		WHERE identifier = ?
	`, newName, identifier)
	if err != nil {
		return fmt.Errorf("cannot update username: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("cannot read rows affected: %w", err)
	}

	if rows == 0 {
		// Nessun utente con quell’identifier (eventualmente dopo gestiamo meglio).
		return sql.ErrNoRows
	}

	return nil
}
