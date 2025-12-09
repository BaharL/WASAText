package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// Message represents a message as defined in the OpenAPI schema.
type Message struct {
	ID                     int64   `json:"id"`
	ChatID                 int64   `json:"chatId"`
	SenderID               int64   `json:"senderId"`
	SenderName            string  `json:"senderName"`
	Kind                   string  `json:"kind"`
	Text                   *string `json:"text,omitempty"`
	MediaURL               *string `json:"mediaUrl,omitempty"`
	ReplyToMessageID       *int64  `json:"replyToMessageId,omitempty"`
	ForwardedFromMessageID *int64  `json:"forwardedFromMessageId,omitempty"`
	Status                 string  `json:"status"`
	CreatedAt              string  `json:"createdAt"`
	Mine                  bool    `json:"mine"` 
}

// Reaction represents a reaction to a message.
type Reaction struct {
	Emoji     string `json:"emoji"`
	UserID    int64  `json:"userId"`
	CreatedAt string `json:"createdAt"`
}

// initMessageTables creates the messages and message_reactions
// tables if they do not already exist.
func initMessageTables(db *sql.DB) error {
	msgStmt := `
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		chat_id INTEGER NOT NULL,
		sender_id INTEGER NOT NULL,
		kind TEXT NOT NULL,
		text TEXT,
		media_url TEXT,
		reply_to_message_id INTEGER,
		forwarded_from_message_id INTEGER,
		status TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := db.Exec(msgStmt); err != nil {
		return fmt.Errorf("error creating messages table: %w", err)
	}

	reactionsStmt := `
	CREATE TABLE IF NOT EXISTS message_reactions (
		message_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		emoji TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (message_id, user_id, emoji)
	);`
	if _, err := db.Exec(reactionsStmt); err != nil {
		return fmt.Errorf("error creating message_reactions table: %w", err)
	}

	return nil
}

// getUserIDByIdentifier converts a user identifier (UUID string)
// into the internal numeric user ID.
func (db *appdbimpl) getUserIDByIdentifier(ctx context.Context, identifier string) (int64, error) {
	var id int64
	err := db.c.QueryRowContext(ctx, `
		SELECT id FROM users WHERE identifier = ?
	`, identifier).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("user not found for identifier %q: %w", identifier, err)
		}
		return 0, fmt.Errorf("cannot lookup user id: %w", err)
	}
	return id, nil
}

// getMessageByID loads a single message by its ID.
func (db *appdbimpl) getMessageByID(ctx context.Context, id int64) (Message, error) {
	var m Message
	var text, mediaURL sql.NullString
	var replyID, fwdID sql.NullInt64

	err := db.c.QueryRowContext(ctx, `
		SELECT
			id, chat_id, sender_id, kind,
			text, media_url,
			reply_to_message_id, forwarded_from_message_id,
			status, datetime(created_at) as created_at
		FROM messages
		WHERE id = ?
	`, id).Scan(
		&m.ID, &m.ChatID, &m.SenderID, &m.Kind,
		&text, &mediaURL, &replyID, &fwdID,
		&m.Status, &m.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Message{}, err
		}
		return Message{}, fmt.Errorf("cannot load message %d: %w", id, err)
	}

	if text.Valid {
		s := text.String
		m.Text = &s
	}
	if mediaURL.Valid {
		s := mediaURL.String
		m.MediaURL = &s
	}
	if replyID.Valid {
		v := replyID.Int64
		m.ReplyToMessageID = &v
	}
	if fwdID.Valid {
		v := fwdID.Int64
		m.ForwardedFromMessageID = &v
	}

	return m, nil
}

// SendMessage creates a new message (POST /messages).
func (db *appdbimpl) SendMessage(
	ctx context.Context,
	senderIdentifier string,
	chatID int64,
	kind string,
	text *string,
	mediaURL *string,
) (Message, error) {

	// 1. Retrieve the numeric user ID.
	senderID, err := db.getUserIDByIdentifier(ctx, senderIdentifier)
	if err != nil {
		return Message{}, err
	}

	// 2. Insert the message.
	res, err := db.c.ExecContext(ctx, `
		INSERT INTO messages (
			chat_id, sender_id, kind, text, media_url, status
		) VALUES (?, ?, ?, ?, ?, 'sent')
	`,
		chatID,
		senderID,
		kind,
		textOrNil(text),
		textOrNil(mediaURL),
	)
	if err != nil {
		return Message{}, fmt.Errorf("cannot insert message: %w", err)
	}

	msgID, err := res.LastInsertId()
	if err != nil {
		return Message{}, fmt.Errorf("cannot get last insert id: %w", err)
	}

	// 3. Reload and return the full message.
	return db.getMessageByID(ctx, msgID)
}

// ForwardMessage forwards an existing message into another chat.
func (db *appdbimpl) ForwardMessage(
	ctx context.Context,
	senderIdentifier string,
	messageID int64,
	toChatID int64,
) (Message, error) {

	// 1. Load the original message.
	orig, err := db.getMessageByID(ctx, messageID)
	if err != nil {
		return Message{}, err
	}

	// 2. Insert the forwarded message.
	senderID, err := db.getUserIDByIdentifier(ctx, senderIdentifier)
	if err != nil {
		return Message{}, err
	}

	res, err := db.c.ExecContext(ctx, `
		INSERT INTO messages (
			chat_id, sender_id, kind, text, media_url,
			forwarded_from_message_id, status
		) VALUES (?, ?, ?, ?, ?, ?, 'sent')
	`,
		toChatID,
		senderID,
		orig.Kind,
		textOrNil(orig.Text),
		textOrNil(orig.MediaURL),
		orig.ID,
	)
	if err != nil {
		return Message{}, fmt.Errorf("cannot forward message: %w", err)
	}

	newID, err := res.LastInsertId()
	if err != nil {
		return Message{}, fmt.Errorf("cannot get forwarded message id: %w", err)
	}

	return db.getMessageByID(ctx, newID)
}

// AddReaction registers a reaction on a message.
func (db *appdbimpl) AddReaction(
	ctx context.Context,
	senderIdentifier string,
	messageID int64,
	emoji string,
) (Reaction, error) {

	userID, err := db.getUserIDByIdentifier(ctx, senderIdentifier)
	if err != nil {
		return Reaction{}, err
	}

	// Insert or overwrite the reaction.
	_, err = db.c.ExecContext(ctx, `
		INSERT OR REPLACE INTO message_reactions (
			message_id, user_id, emoji, created_at
		) VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	`, messageID, userID, emoji)
	if err != nil {
		return Reaction{}, fmt.Errorf("cannot add reaction: %w", err)
	}

	var r Reaction
	err = db.c.QueryRowContext(ctx, `
		SELECT emoji, user_id, datetime(created_at)
		FROM message_reactions
		WHERE message_id = ? AND user_id = ? AND emoji = ?
	`, messageID, userID, emoji).Scan(&r.Emoji, &r.UserID, &r.CreatedAt)
	if err != nil {
		return Reaction{}, fmt.Errorf("cannot reload reaction: %w", err)
	}

	return r, nil
}

// RemoveReaction deletes a reaction from a message.
func (db *appdbimpl) RemoveReaction(
	ctx context.Context,
	senderIdentifier string,
	messageID int64,
	emoji string,
) error {

	userID, err := db.getUserIDByIdentifier(ctx, senderIdentifier)
	if err != nil {
		return err
	}

	res, err := db.c.ExecContext(ctx, `
		DELETE FROM message_reactions
		WHERE message_id = ? AND user_id = ? AND emoji = ?
	`, messageID, userID, emoji)
	if err != nil {
		return fmt.Errorf("cannot remove reaction: %w", err)
	}

	affected, err := res.RowsAffected()
	if err == nil && affected == 0 {
		return sql.ErrNoRows
	}

	return err
}

// DeleteMessage deletes a message sent by the authenticated user.
func (db *appdbimpl) DeleteMessage(
	ctx context.Context,
	senderIdentifier string,
	messageID int64,
) error {

	userID, err := db.getUserIDByIdentifier(ctx, senderIdentifier)
	if err != nil {
		return err
	}

	res, err := db.c.ExecContext(ctx, `
		DELETE FROM messages
		WHERE id = ? AND sender_id = ?
	`, messageID, userID)
	if err != nil {
		return fmt.Errorf("cannot delete message: %w", err)
	}

	affected, err := res.RowsAffected()
	if err == nil && affected == 0 {
		return sql.ErrNoRows
	}

	return err
}

// textOrNil helps passing NULL to the DB when text/mediaURL are nil.
func textOrNil(s *string) interface{} {
	if s == nil {
		return nil
	}
	return *s
}
