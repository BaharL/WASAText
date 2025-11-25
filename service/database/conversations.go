package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// ConversationSummary represents a single conversation item
// returned by GET /conversations.
type ConversationSummary struct {
	ID          int64    `json:"id"`
	Title       string   `json:"title"`
	LastMessage *Message `json:"lastMessage,omitempty"`
}

// ListUserConversations returns the list of conversations
// in which the given user has sent at least one message.
func (db *appdbimpl) ListUserConversations(
	ctx context.Context,
	userIdentifier string,
) ([]ConversationSummary, error) {

	// 1) Resolve the numeric user ID from the identifier.
	userID, err := db.getUserIDByIdentifier(ctx, userIdentifier)
	if err != nil {
		return nil, err
	}

	// 2) For now, we define a "conversation" as a chat_id where
	// this user has sent at least one message. We get the last
	// message id for each chat.
	rows, err := db.c.QueryContext(ctx, `
		SELECT
			m.chat_id,
			MAX(m.id) AS last_message_id
		FROM messages m
		WHERE m.sender_id = ?
		GROUP BY m.chat_id
		ORDER BY last_message_id DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("query ListUserConversations: %w", err)
	}
	defer rows.Close()

	var result []ConversationSummary

	for rows.Next() {
		var chatID, lastMsgID int64
		if err := rows.Scan(&chatID, &lastMsgID); err != nil {
			return nil, fmt.Errorf("scan ListUserConversations: %w", err)
		}

		// Load the last message using the existing helper.
		lastMsg, err := db.getMessageByID(ctx, lastMsgID)
		if err != nil {
			// If the message disappeared, skip this conversation.
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			return nil, fmt.Errorf("load last message for chat %d: %w", chatID, err)
		}

		// Simple title for now. You can improve it later (e.g. participants, group name, etc.).
		title := fmt.Sprintf("Chat %d", chatID)

		result = append(result, ConversationSummary{
			ID:          chatID,
			Title:       title,
			LastMessage: &lastMsg,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows ListUserConversations: %w", err)
	}

	return result, nil
}

// ListConversationMessages returns all messages for a given conversation.
//
// It currently assumes that the user is allowed to see the conversation.
// You can later add an authorization check if you introduce a participants table.
func (db *appdbimpl) ListConversationMessages(
	ctx context.Context,
	userIdentifier string,
	conversationID int64,
) ([]Message, error) {

	// Optional: check that the user exists (and potentially that it belongs to the chat).
	if _, err := db.getUserIDByIdentifier(ctx, userIdentifier); err != nil {
		return nil, err
	}

	rows, err := db.c.QueryContext(ctx, `
		SELECT
			id, chat_id, sender_id, kind,
			text, media_url,
			reply_to_message_id, forwarded_from_message_id,
			status, datetime(created_at) as created_at
		FROM messages
		WHERE chat_id = ?
		ORDER BY created_at ASC, id ASC
	`, conversationID)
	if err != nil {
		return nil, fmt.Errorf("query ListConversationMessages: %w", err)
	}
	defer rows.Close()

	var messages []Message

	for rows.Next() {
		var m Message
		var text, mediaURL sql.NullString
		var replyID, fwdID sql.NullInt64

		if err := rows.Scan(
			&m.ID,
			&m.ChatID,
			&m.SenderID,
			&m.Kind,
			&text,
			&mediaURL,
			&replyID,
			&fwdID,
			&m.Status,
			&m.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan ListConversationMessages: %w", err)
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

		messages = append(messages, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows ListConversationMessages: %w", err)
	}

	return messages, nil
}
