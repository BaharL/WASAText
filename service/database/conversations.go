package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// ConversationSummary è il modello restituito da GET /conversations.
type ConversationSummary struct {
	ID          int64    `json:"id"`
	Title       string   `json:"title"`
	IsGroup     bool     `json:"isGroup"`
	LastMessage *Message `json:"lastMessage,omitempty"`
}

// ListUserConversations restituisce tutte le conversazioni a cui partecipa l’utente,
// con l’ultimo messaggio (se presente).
func (db *appdbimpl) ListUserConversations(
	ctx context.Context,
	userIdentifier string,
) ([]ConversationSummary, error) {
	userID, err := db.getUserIDByIdentifier(ctx, userIdentifier)
	if err != nil {
		return nil, err
	}

	rows, err := db.c.QueryContext(ctx, `
		SELECT
			c.id,
			COALESCE(c.name, printf('Chat %d', c.id)) AS title,
			c.type,
			MAX(m.id) AS last_message_id
		FROM conversations c
		JOIN conversation_members cm ON cm.conversation_id = c.id
		LEFT JOIN messages m ON m.chat_id = c.id
		WHERE cm.user_id = ?
		GROUP BY c.id, title, c.type
		ORDER BY last_message_id DESC, c.id DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list user conversations: %w", err)
	}
	defer rows.Close()

	var convs []ConversationSummary

	for rows.Next() {
		var conv ConversationSummary
		var lastMsgID sql.NullInt64
		var convType sql.NullString

		if err := rows.Scan(&conv.ID, &conv.Title, &convType, &lastMsgID); err != nil {
			return nil, fmt.Errorf("scan user conversations: %w", err)
		}

		// NEW: isGroup basato su conversations.type
		if convType.Valid && convType.String == "group" {
			conv.IsGroup = true
		} else {
			conv.IsGroup = false
		}

		if lastMsgID.Valid {
			lastMsg, err := db.getMessageByID(ctx, lastMsgID.Int64)
			if err != nil {
				if !errors.Is(err, sql.ErrNoRows) {
					return nil, fmt.Errorf("load last message for conversation %d: %w", conv.ID, err)
				}
			} else {
				conv.LastMessage = &lastMsg
			}
		}

		convs = append(convs, conv)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows user conversations: %w", err)
	}

	return convs, nil
}

// ListConversationMessages restituisce tutti i messaggi di una conversazione,
// se l’utente è membro di quella conversazione.
func (db *appdbimpl) ListConversationMessages(
	ctx context.Context,
	userIdentifier string,
	conversationID int64,
) ([]Message, error) {
	// Trovo l'ID interno dell'utente corrente
	userID, err := db.getUserIDByIdentifier(ctx, userIdentifier)
	if err != nil {
		return nil, err
	}

	// Controllo che l’utente faccia parte della conversazione
	var exists int
	err = db.c.QueryRowContext(ctx, `
		SELECT 1
		FROM conversation_members
		WHERE conversation_id = ? AND user_id = ?
	`, conversationID, userID).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("check conversation membership: %w", err)
	}

	// Carico i messaggi + nome del mittente
	rows, err := db.c.QueryContext(ctx, `
		SELECT
			m.id,
			m.chat_id,
			m.sender_id,
			u.name AS sender_name,
			m.kind,
			m.text,
			m.media_url,
			m.reply_to_message_id,
			m.forwarded_from_message_id,
			m.status,
			datetime(m.created_at) AS created_at
		FROM messages m
		JOIN users u ON u.id = m.sender_id
		WHERE m.chat_id = ?
		ORDER BY m.created_at ASC, m.id ASC
	`, conversationID)
	if err != nil {
		return nil, fmt.Errorf("list conversation messages: %w", err)
	}
	defer rows.Close()

	var messages []Message

	for rows.Next() {
		var m Message
		var text, mediaURL sql.NullString
		var replyID, fwdID sql.NullInt64
		var senderName sql.NullString

		if err := rows.Scan(
			&m.ID,
			&m.ChatID,
			&m.SenderID,
			&senderName,
			&m.Kind,
			&text,
			&mediaURL,
			&replyID,
			&fwdID,
			&m.Status,
			&m.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan conversation messages: %w", err)
		}

		if senderName.Valid {
			m.SenderName = senderName.String
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

		// NEW: segno se è un messaggio dell'utente loggato
		m.Mine = (m.SenderID == userID)

		messages = append(messages, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows conversation messages: %w", err)
	}

	return messages, nil
}
