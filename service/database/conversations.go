package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// ConversationSummary è il modello restituito da GET /conversations.
type ConversationSummary struct {
	ID          int64    `json:"id"`
	Title       string   `json:"title"`
	IsGroup     bool     `json:"isGroup"`
	PhotoURL    *string  `json:"photoUrl,omitempty"`
	LastMessage *Message `json:"lastMessage,omitempty"`
}

// ConversationInfo contiene i metadati di una conversazione
type ConversationInfo struct {
	ID       int64   `json:"id"`
	Title    string  `json:"title"`
	IsGroup  bool    `json:"isGroup"`
	PhotoURL *string `json:"photoUrl,omitempty"`
}

// ensureMessageStatusTable crea la tabella message_status se non esiste.
// Serve per le spunte (received/read).
func (db *appdbimpl) ensureMessageStatusTable(ctx context.Context) {
	_, _ = db.c.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS message_status (
			message_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			status TEXT NOT NULL,
			PRIMARY KEY (message_id, user_id)
		);
	`)
}

// ListUserConversations restituisce tutte le conversazioni a cui partecipa l'utente,
// con l'ultimo messaggio (se presente) + title/photo "giusti" per direct e group.
func (db *appdbimpl) ListUserConversations(
	ctx context.Context,
	userIdentifier string,
) ([]ConversationSummary, error) {

	userID, err := db.getUserIDByIdentifier(ctx, userIdentifier)
	if err != nil {
		return nil, err
	}

	// Safe: per evitare "no such table: message_status" in altri endpoint
	db.ensureMessageStatusTable(ctx)

	rows, err := db.c.QueryContext(ctx, `
		SELECT
			c.id,
			c.type,
			c.name,
			c.photo_url,
			MAX(m.id) AS last_message_id,

			-- Direct: nome dell'altro utente (rispetto a me)
			(
				SELECT u2.name
				FROM conversation_members cm2
				JOIN users u2 ON u2.id = cm2.user_id
				WHERE cm2.conversation_id = c.id AND cm2.user_id <> ?
				LIMIT 1
			) AS direct_other_name,

			(
				SELECT u2.photo_url
				FROM conversation_members cm2
				JOIN users u2 ON u2.id = cm2.user_id
				WHERE cm2.conversation_id = c.id AND cm2.user_id <> ?
				LIMIT 1
			) AS direct_other_photo

		FROM conversations c
		JOIN conversation_members cm ON cm.conversation_id = c.id
		LEFT JOIN messages m ON m.chat_id = c.id
		WHERE cm.user_id = ?
		GROUP BY c.id, c.type, c.name, c.photo_url
		ORDER BY last_message_id DESC, c.id DESC
	`, userID, userID, userID)
	if err != nil {
		return nil, fmt.Errorf("list user conversations: %w", err)
	}
	defer rows.Close()

	convs := make([]ConversationSummary, 0)

	for rows.Next() {
		var conv ConversationSummary

		var convType sql.NullString
		var groupName sql.NullString
		var groupPhoto sql.NullString
		var lastMsgID sql.NullInt64

		var directOtherName sql.NullString
		var directOtherPhoto sql.NullString

		if err := rows.Scan(
			&conv.ID,
			&convType,
			&groupName,
			&groupPhoto,
			&lastMsgID,
			&directOtherName,
			&directOtherPhoto,
		); err != nil {
			return nil, fmt.Errorf("scan user conversations: %w", err)
		}

		isGroup := convType.Valid && strings.ToLower(convType.String) == "group"
		conv.IsGroup = isGroup

		// Title + PhotoURL
		if isGroup {
			if groupName.Valid && groupName.String != "" {
				conv.Title = groupName.String
			} else {
				conv.Title = fmt.Sprintf("Chat %d", conv.ID)
			}
			if groupPhoto.Valid && groupPhoto.String != "" {
				s := groupPhoto.String
				conv.PhotoURL = &s
			}
		} else {
			if directOtherName.Valid && directOtherName.String != "" {
				conv.Title = directOtherName.String
			} else {
				conv.Title = fmt.Sprintf("Chat %d", conv.ID)
			}
			if directOtherPhoto.Valid && directOtherPhoto.String != "" {
				s := directOtherPhoto.String
				conv.PhotoURL = &s
			}
		}

		// Last message (se c'è)
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
// se l'utente è membro di quella conversazione.
func (db *appdbimpl) ListConversationMessages(
	ctx context.Context,
	userIdentifier string,
	conversationID int64,
) ([]Message, error) {

	userID, err := db.getUserIDByIdentifier(ctx, userIdentifier)
	if err != nil {
		return nil, err
	}

	// ensure status table exists
	db.ensureMessageStatusTable(ctx)

	// membership check
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

	// NOTE:
	// - Per i messaggi "mine", lo status dipende da cosa hanno fatto GLI ALTRI membri (received/read).
	// - Per i messaggi non miei, lato UI non mostriamo spunte, quindi possiamo mettere 'sent' fisso.
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

			CASE
				WHEN m.sender_id = ? THEN
					CASE
						-- READ se almeno un altro membro ha read
						WHEN EXISTS (
							SELECT 1
							FROM conversation_members cm
							JOIN message_status ms
							  ON ms.user_id = cm.user_id AND ms.message_id = m.id
							WHERE cm.conversation_id = m.chat_id
							  AND cm.user_id <> ?
							  AND ms.status = 'read'
						) THEN 'read'
						-- RECEIVED se almeno un altro membro ha received
						WHEN EXISTS (
							SELECT 1
							FROM conversation_members cm
							JOIN message_status ms
							  ON ms.user_id = cm.user_id AND ms.message_id = m.id
							WHERE cm.conversation_id = m.chat_id
							  AND cm.user_id <> ?
							  AND ms.status = 'received'
						) THEN 'received'
						ELSE 'sent'
					END
				ELSE 'sent'
			END AS status,

			datetime(m.created_at) AS created_at
		FROM messages m
		JOIN users u ON u.id = m.sender_id
		WHERE m.chat_id = ?
		ORDER BY m.created_at ASC, m.id ASC
	`, userID, userID, userID, conversationID)
	if err != nil {
		return nil, fmt.Errorf("list conversation messages: %w", err)
	}
	defer rows.Close()

	messages := make([]Message, 0)

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

		m.Mine = (m.SenderID == userID)
		messages = append(messages, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows conversation messages: %w", err)
	}

	// Reactions in 1 query
	ids := make([]int64, 0, len(messages))
	for i := range messages {
		ids = append(ids, messages[i].ID)
	}

	reactionMap, err := db.loadReactionSummaries(ctx, userID, ids)
	if err != nil {
		return nil, err
	}

	for i := range messages {
		if rs, ok := reactionMap[messages[i].ID]; ok {
			messages[i].Reactions = rs
		}
	}

	return messages, nil
}

// GetConversationInfo restituisce solo le info di base di una conversazione
func (db *appdbimpl) GetConversationInfo(
	ctx context.Context,
	userIdentifier string,
	conversationID int64,
) (ConversationInfo, error) {

	userID, err := db.getUserIDByIdentifier(ctx, userIdentifier)
	if err != nil {
		return ConversationInfo{}, err
	}

	// membership check
	var exists int
	err = db.c.QueryRowContext(ctx, `
		SELECT 1
		FROM conversation_members
		WHERE conversation_id = ? AND user_id = ?
	`, conversationID, userID).Scan(&exists)
	if err != nil {
		return ConversationInfo{}, err
	}

	row := db.c.QueryRowContext(ctx, `
		SELECT
			c.id,
			c.type,
			c.name,
			c.photo_url,

			-- direct other name
			(
				SELECT u2.name
				FROM conversation_members cm2
				JOIN users u2 ON u2.id = cm2.user_id
				WHERE cm2.conversation_id = c.id AND cm2.user_id <> ?
				LIMIT 1
			) AS direct_other_name,

			(
				SELECT u2.photo_url
				FROM conversation_members cm2
				JOIN users u2 ON u2.id = cm2.user_id
				WHERE cm2.conversation_id = c.id AND cm2.user_id <> ?
				LIMIT 1
			) AS direct_other_photo
		FROM conversations c
		WHERE c.id = ?
	`, userID, userID, conversationID)

	var info ConversationInfo
	var convType sql.NullString
	var groupName sql.NullString
	var groupPhoto sql.NullString
	var directOtherName sql.NullString
	var directOtherPhoto sql.NullString

	if err := row.Scan(
		&info.ID,
		&convType,
		&groupName,
		&groupPhoto,
		&directOtherName,
		&directOtherPhoto,
	); err != nil {
		return ConversationInfo{}, err
	}

	isGroup := convType.Valid && strings.ToLower(convType.String) == "group"
	info.IsGroup = isGroup

	if isGroup {
		if groupName.Valid && groupName.String != "" {
			info.Title = groupName.String
		} else {
			info.Title = fmt.Sprintf("Chat %d", info.ID)
		}
		if groupPhoto.Valid && groupPhoto.String != "" {
			s := groupPhoto.String
			info.PhotoURL = &s
		}
	} else {
		if directOtherName.Valid && directOtherName.String != "" {
			info.Title = directOtherName.String
		} else {
			info.Title = fmt.Sprintf("Chat %d", info.ID)
		}
		if directOtherPhoto.Valid && directOtherPhoto.String != "" {
			s := directOtherPhoto.String
			info.PhotoURL = &s
		}
	}

	return info, nil
}
