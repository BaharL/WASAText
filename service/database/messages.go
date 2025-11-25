package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// Message rappresenta un messaggio come nello schema OpenAPI.
type Message struct {
	ID                     int64   `json:"id"`
	ChatID                 int64   `json:"chatId"`
	SenderID               int64   `json:"senderId"`
	Kind                   string  `json:"kind"`
	Text                   *string `json:"text,omitempty"`
	MediaURL               *string `json:"mediaUrl,omitempty"`
	ReplyToMessageID       *int64  `json:"replyToMessageId,omitempty"`
	ForwardedFromMessageID *int64  `json:"forwardedFromMessageId,omitempty"`
	Status                 string  `json:"status"`
	CreatedAt              string  `json:"createdAt"`
}

// Reaction rappresenta una reazione a un messaggio.
type Reaction struct {
	Emoji     string `json:"emoji"`
	UserID    int64  `json:"userId"`
	CreatedAt string `json:"createdAt"`
}

// initMessageTables crea le tabelle messages e message_reactions se non esistono.
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

// getUserIDByIdentifier converte l'identifier (UUID stringa) nell'id numerico interno.
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

// helper per fare SELECT di un singolo messaggio per id.
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

// SendMessage crea un nuovo messaggio (POST /messages).
func (db *appdbimpl) SendMessage(
	ctx context.Context,
	senderIdentifier string,
	chatID int64,
	kind string,
	text *string,
	mediaURL *string,
) (Message, error) {
	// 1. Trovo l'id numerico dell'utente.
	senderID, err := db.getUserIDByIdentifier(ctx, senderIdentifier)
	if err != nil {
		return Message{}, err
	}

	// 2. Inserisco il messaggio.
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

	// 3. Rileggo il messaggio completo.
	return db.getMessageByID(ctx, msgID)
}

// ForwardMessage inoltra un messaggio esistente in un'altra chat.
func (db *appdbimpl) ForwardMessage(
	ctx context.Context,
	senderIdentifier string,
	messageID int64,
	toChatID int64,
) (Message, error) {
	// 1. Recupero il messaggio originale.
	orig, err := db.getMessageByID(ctx, messageID)
	if err != nil {
		return Message{}, err
	}

	// 2. Campo forwarded_from = id del messaggio originale.
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

// AddReaction aggiunge una reazione a un messaggio.
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

	// Inserisco la reazione; se esiste già, la sovrascrivo.
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

// RemoveReaction rimuove una reazione da un messaggio.
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

// DeleteMessage cancella un messaggio inviato dall'utente corrente.
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

// textOrNil aiuta a passare valori NULL al DB quando text/mediaUrl sono nil.
func textOrNil(s *string) interface{} {
	if s == nil {
		return nil
	}
	return *s
}
