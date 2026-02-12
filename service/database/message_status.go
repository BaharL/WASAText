package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// MarkConversationReceived: set incoming messages to received for THIS user only.
func (db *appdbimpl) MarkConversationReceived(ctx context.Context, userIdentifier string, conversationID int64) error {
	userID, err := db.getUserIDByIdentifier(ctx, userIdentifier)
	if err != nil {
		return err
	}
	if ok, err := db.isConversationMember(ctx, conversationID, userID); err != nil {
		return err
	} else if !ok {
		return sql.ErrNoRows
	}

	// Ensure table exists (if you already create it elsewhere, ok)
	_, _ = db.c.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS message_status (
			message_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			status TEXT NOT NULL,
			PRIMARY KEY (message_id, user_id)
		);
	`)

	// Insert/upgrade to received only if not already read
	_, err = db.c.ExecContext(ctx, `
		INSERT INTO message_status (message_id, user_id, status)
		SELECT m.id, ?, 'received'
		FROM messages m
		WHERE m.chat_id = ?
		  AND m.sender_id <> ?
		  AND NOT EXISTS (
		    SELECT 1 FROM message_status ms
		    WHERE ms.message_id = m.id AND ms.user_id = ?
		      AND ms.status IN ('received', 'read')
		  )
	`, userID, conversationID, userID, userID)
	if err != nil {
		return fmt.Errorf("mark conversation received: %w", err)
	}
	return nil
}

// MarkConversationRead: set incoming messages to read for THIS user only.
func (db *appdbimpl) MarkConversationRead(ctx context.Context, userIdentifier string, conversationID int64) error {
	userID, err := db.getUserIDByIdentifier(ctx, userIdentifier)
	if err != nil {
		return err
	}
	if ok, err := db.isConversationMember(ctx, conversationID, userID); err != nil {
		return err
	} else if !ok {
		return sql.ErrNoRows
	}

	_, _ = db.c.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS message_status (
			message_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			status TEXT NOT NULL,
			PRIMARY KEY (message_id, user_id)
		);
	`)

	// Upsert to read for all incoming messages
	_, err = db.c.ExecContext(ctx, `
		INSERT INTO message_status (message_id, user_id, status)
		SELECT m.id, ?, 'read'
		FROM messages m
		WHERE m.chat_id = ?
		  AND m.sender_id <> ?
		ON CONFLICT(message_id, user_id) DO UPDATE SET status='read'
	`, userID, conversationID, userID)
	if err != nil {
		return fmt.Errorf("mark conversation read: %w", err)
	}
	return nil
}

func (db *appdbimpl) isConversationMember(ctx context.Context, conversationID int64, userID int64) (bool, error) {
	var exists int
	err := db.c.QueryRowContext(ctx, `
		SELECT 1 FROM conversation_members
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
