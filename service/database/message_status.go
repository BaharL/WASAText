package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// MarkConversationReceived updates incoming messages to "received".
// This is a simplified per-message status model (works best for direct chats).
func (db *appdbimpl) MarkConversationReceived(
	ctx context.Context,
	userIdentifier string,
	conversationID int64,
) error {

	userID, err := db.getUserIDByIdentifier(ctx, userIdentifier)
	if err != nil {
		return err
	}

	if ok, err := db.isConversationMember(ctx, conversationID, userID); err != nil {
		return err
	} else if !ok {
		return sql.ErrNoRows
	}

	_, err = db.c.ExecContext(ctx, `
		UPDATE messages
		SET status = 'received'
		WHERE chat_id = ?
		  AND sender_id <> ?
		  AND status = 'sent'
	`, conversationID, userID)
	if err != nil {
		return fmt.Errorf("mark conversation received: %w", err)
	}

	return nil
}

// MarkConversationRead updates incoming messages to "read".
// This upgrades both "sent" and "received" to "read".
func (db *appdbimpl) MarkConversationRead(
	ctx context.Context,
	userIdentifier string,
	conversationID int64,
) error {

	userID, err := db.getUserIDByIdentifier(ctx, userIdentifier)
	if err != nil {
		return err
	}

	if ok, err := db.isConversationMember(ctx, conversationID, userID); err != nil {
		return err
	} else if !ok {
		return sql.ErrNoRows
	}

	_, err = db.c.ExecContext(ctx, `
		UPDATE messages
		SET status = 'read'
		WHERE chat_id = ?
		  AND sender_id <> ?
		  AND status IN ('sent','received')
	`, conversationID, userID)
	if err != nil {
		return fmt.Errorf("mark conversation read: %w", err)
	}

	return nil
}

func (db *appdbimpl) isConversationMember(ctx context.Context, conversationID int64, userID int64) (bool, error) {
	var exists int
	err := db.c.QueryRowContext(ctx, `
		SELECT 1
		FROM conversation_members
		WHERE conversation_id = ? AND user_id = ?
	`, conversationID, userID).Scan(&exists)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("check conversation membership: %w", err)
	}
	return true, nil
}
