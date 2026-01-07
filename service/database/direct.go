package database

import (
	"context"
	"database/sql"
	"fmt"
)

func (db *appdbimpl) GetOrCreateDirectChat(
	ctx context.Context,
	requesterIdentifier string,
	otherIdentifier string,
) (int64, error) {

	if requesterIdentifier == otherIdentifier {
		return 0, fmt.Errorf("cannot create direct chat with self")
	}

	meID, err := db.getUserIDByIdentifier(ctx, requesterIdentifier)
	if err != nil {
		return 0, err
	}
	otherID, err := db.getUserIDByIdentifier(ctx, otherIdentifier)
	if err != nil {
		return 0, err
	}

	// 1) Find existing direct chat with exactly these 2 members
	var existingID int64
	err = db.c.QueryRowContext(ctx, `
		SELECT c.id
		FROM conversations c
		JOIN conversation_members cm ON cm.conversation_id = c.id
		WHERE c.type = 'direct' AND cm.user_id IN (?, ?)
		GROUP BY c.id
		HAVING
			COUNT(DISTINCT cm.user_id) = 2
			AND (SELECT COUNT(*) FROM conversation_members WHERE conversation_id = c.id) = 2
		LIMIT 1
	`, meID, otherID).Scan(&existingID)

	if err == nil {
		return existingID, nil
	}
	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("cannot lookup direct chat: %w", err)
	}

	// 2) Create new direct in a transaction
	tx, err := db.c.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("cannot begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx, `
		INSERT INTO conversations (type, name)
		VALUES ('direct', NULL)
	`)
	if err != nil {
		return 0, fmt.Errorf("cannot insert direct conversation: %w", err)
	}

	chatID, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("cannot get direct conversation id: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO conversation_members (conversation_id, user_id)
		VALUES (?, ?), (?, ?)
	`, chatID, meID, chatID, otherID); err != nil {
		return 0, fmt.Errorf("cannot add members to direct: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("cannot commit direct creation: %w", err)
	}

	return chatID, nil
}
