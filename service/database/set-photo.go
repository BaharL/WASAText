package database

import (
	"context"
	"fmt"
)

// SetUserPhoto updates the photo_url of the user identified by identifier.
func (db *appdbimpl) SetUserPhoto(
	ctx context.Context,
	identifier string,
	photoURL string,
) error {

	userID, err := db.getUserIDByIdentifier(ctx, identifier)
	if err != nil {
		return err
	}

	_, err = db.c.ExecContext(ctx, `
		UPDATE users
		SET photo_url = ?
		WHERE id = ?
	`, photoURL, userID)
	if err != nil {
		return fmt.Errorf("cannot update user photo: %w", err)
	}

	return nil
}
