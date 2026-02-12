package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// =====================
// MODELS
// =====================

type ConversationSummary struct {
	ID          int64    `json:"id"`
	Title       string   `json:"title"`
	IsGroup     bool     `json:"isGroup"`
	PhotoURL    *string  `json:"photoUrl,omitempty"`
	LastMessage *Message `json:"lastMessage,omitempty"`
	UnreadCount int      `json:"unreadCount"`
}

type ConversationInfo struct {
	ID       int64   `json:"id"`
	Title    string  `json:"title"`
	IsGroup  bool    `json:"isGroup"`
	PhotoURL *string `json:"photoUrl,omitempty"`
}

// =====================
// LIST USER CONVERSATIONS
// =====================

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
			c.type,
			c.name,
			c.photo_url,
			MAX(m.id) AS last_message_id,

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

		isGroup := convType.Valid && convType.String == "group"
		conv.IsGroup = isGroup

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

		// =====================
		// UNREAD COUNT (PER USER)
		// =====================

		var unread int
		err = db.c.QueryRowContext(ctx, `
			SELECT COUNT(*)
			FROM messages m
			LEFT JOIN message_status ms
			  ON ms.message_id = m.id AND ms.user_id = ?
			WHERE m.chat_id = ?
			  AND m.sender_id <> ?
			  AND (ms.status IS NULL OR ms.status <> 'read')
		`, userID, conv.ID, userID).Scan(&unread)
		if err != nil {
			return nil, fmt.Errorf("count unread messages: %w", err)
		}
		conv.UnreadCount = unread

		// =====================
		// LAST MESSAGE
		// =====================

		if lastMsgID.Valid {
			lastMsg, err := db.getMessageByID(ctx, lastMsgID.Int64)
			if err == nil {
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

// =====================
// LIST CONVERSATION MESSAGES
// =====================

func (db *appdbimpl) ListConversationMessages(
	ctx context.Context,
	userIdentifier string,
	conversationID int64,
) ([]Message, error) {

	userID, err := db.getUserIDByIdentifier(ctx, userIdentifier)
	if err != nil {
		return nil, err
	}

	if ok, err := db.isConversationMember(ctx, conversationID, userID); err != nil {
		return nil, err
	} else if !ok {
		return nil, sql.ErrNoRows
	}

	rows, err := db.c.QueryContext(ctx, `
		SELECT
			m.id,
			m.chat_id,
			m.sender_id,
			u.name,
			m.kind,
			m.text,
			m.media_url,
			m.reply_to_message_id,
			m.forwarded_from_message_id,
			datetime(m.created_at),
			COALESCE(ms.status, 'sent') as user_status
		FROM messages m
		JOIN users u ON u.id = m.sender_id
		LEFT JOIN message_status ms
		  ON ms.message_id = m.id AND ms.user_id = ?
		WHERE m.chat_id = ?
		ORDER BY m.created_at ASC, m.id ASC
	`, userID, conversationID)
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
			&m.CreatedAt,
			&m.Status,
		); err != nil {
			return nil, err
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

	return messages, nil
}

// =====================
// GET CONVERSATION INFO
// =====================

func (db *appdbimpl) GetConversationInfo(
	ctx context.Context,
	userIdentifier string,
	conversationID int64,
) (ConversationInfo, error) {

	userID, err := db.getUserIDByIdentifier(ctx, userIdentifier)
	if err != nil {
		return ConversationInfo{}, err
	}

	if ok, err := db.isConversationMember(ctx, conversationID, userID); err != nil {
		return ConversationInfo{}, err
	} else if !ok {
		return ConversationInfo{}, sql.ErrNoRows
	}

	row := db.c.QueryRowContext(ctx, `
		SELECT
			c.id,
			c.type,
			c.name,
			c.photo_url
		FROM conversations c
		WHERE c.id = ?
	`, conversationID)

	var info ConversationInfo
	var convType sql.NullString
	var name sql.NullString
	var photo sql.NullString

	if err := row.Scan(&info.ID, &convType, &name, &photo); err != nil {
		return ConversationInfo{}, err
	}

	info.IsGroup = convType.Valid && convType.String == "group"

	if name.Valid && name.String != "" {
		info.Title = name.String
	} else {
		info.Title = fmt.Sprintf("Chat %d", info.ID)
	}

	if photo.Valid && photo.String != "" {
		s := photo.String
		info.PhotoURL = &s
	}

	return info, nil
}
