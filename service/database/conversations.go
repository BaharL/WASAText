package database

// ConversationSummary represents a single conversation item
// returned by GET /conversations.
type ConversationSummary struct {
	ID          int64    `json:"id"`
	Title       string   `json:"title"`
	LastMessage *Message `json:"lastMessage,omitempty"`
}

// ConversationDetails represents a full conversation with all messages.
// (You may or may not need this depending on your YAML.)
type ConversationDetails struct {
	ID       int64     `json:"id"`
	Title    string    `json:"title"`
	Messages []Message `json:"messages"`
}
