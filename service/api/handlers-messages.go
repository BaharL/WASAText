package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// Request models for messages (aligned with the OpenAPI schema).

type sendMessageRequest struct {
	ChatID   int64   `json:"chatId"`
	Kind     string  `json:"kind"`
	Text     *string `json:"text,omitempty"`
	MediaURL *string `json:"mediaUrl,omitempty"`
}

type forwardMessageRequest struct {
	ToChatID int64 `json:"toChatId"`
}

type addReactionRequest struct {
	Emoji string `json:"emoji"`
}

// sendMessage handles POST /messages.
func (rt *_router) sendMessage(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if ctx.UserIdentifier == "" {
		http.Error(w, `{"message":"missing or invalid Authorization header"}`, http.StatusUnauthorized)
		return
	}

	var req sendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"message":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}

	// Basic validation
	req.Kind = strings.ToLower(strings.TrimSpace(req.Kind))
	if req.ChatID <= 0 {
		http.Error(w, `{"message":"chatId must be positive"}`, http.StatusBadRequest)
		return
	}
	if req.Kind != "text" && req.Kind != "gif" && req.Kind != "image" {
		http.Error(w, `{"message":"invalid kind"}`, http.StatusBadRequest)
		return
	}

	// Rules on text / mediaUrl
	if req.Kind == "text" {
		if req.Text == nil || strings.TrimSpace(*req.Text) == "" {
			http.Error(w, `{"message":"text is required for text messages"}`, http.StatusBadRequest)
			return
		}
		// For safety, mediaUrl must be empty
		req.MediaURL = nil
	} else {
		// gif / image → mediaUrl is required
		if req.MediaURL == nil || strings.TrimSpace(*req.MediaURL) == "" {
			http.Error(w, `{"message":"mediaUrl is required for gif/image messages"}`, http.StatusBadRequest)
			return
		}
	}

	msg, err := rt.db.SendMessage(
		r.Context(),
		ctx.UserIdentifier,
		req.ChatID,
		req.Kind,
		req.Text,
		req.MediaURL,
	)
	if err != nil {
		ctx.Logger.WithError(err).Error("cannot send message")
		http.Error(w, `{"message":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(msg)
}

// forwardMessage handles POST /messages/:messageId/forward.
func (rt *_router) forwardMessage(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if ctx.UserIdentifier == "" {
		http.Error(w, `{"message":"missing or invalid Authorization header"}`, http.StatusUnauthorized)
		return
	}

	// messageId from path
	msgIDStr := ps.ByName("messageId")
	messageID, err := strconv.ParseInt(msgIDStr, 10, 64)
	if err != nil || messageID <= 0 {
		http.Error(w, `{"message":"invalid messageId"}`, http.StatusBadRequest)
		return
	}

	var req forwardMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"message":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}
	if req.ToChatID <= 0 {
		http.Error(w, `{"message":"toChatId must be positive"}`, http.StatusBadRequest)
		return
	}

	msg, err := rt.db.ForwardMessage(r.Context(), ctx.UserIdentifier, messageID, req.ToChatID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, `{"message":"message not found"}`, http.StatusNotFound)
			return
		}
		ctx.Logger.WithError(err).Error("cannot forward message")
		http.Error(w, `{"message":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(msg)
}

// commentMessage handles POST /messages/:messageId/reactions.
func (rt *_router) commentMessage(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if ctx.UserIdentifier == "" {
		http.Error(w, `{"message":"missing or invalid Authorization header"}`, http.StatusUnauthorized)
		return
	}

	msgIDStr := ps.ByName("messageId")
	messageID, err := strconv.ParseInt(msgIDStr, 10, 64)
	if err != nil || messageID <= 0 {
		http.Error(w, `{"message":"invalid messageId"}`, http.StatusBadRequest)
		return
	}

	var req addReactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"message":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}
	req.Emoji = strings.TrimSpace(req.Emoji)
	if req.Emoji == "" {
		http.Error(w, `{"message":"emoji is required"}`, http.StatusBadRequest)
		return
	}

	reaction, err := rt.db.AddReaction(r.Context(), ctx.UserIdentifier, messageID, req.Emoji)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, `{"message":"message not found"}`, http.StatusNotFound)
			return
		}
		ctx.Logger.WithError(err).Error("cannot add reaction")
		http.Error(w, `{"message":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(reaction)
}

// uncommentMessage handles DELETE /messages/:messageId/reactions/:emoji.
func (rt *_router) uncommentMessage(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if ctx.UserIdentifier == "" {
		http.Error(w, `{"message":"missing or invalid Authorization header"}`, http.StatusUnauthorized)
		return
	}

	msgIDStr := ps.ByName("messageId")
	messageID, err := strconv.ParseInt(msgIDStr, 10, 64)
	if err != nil || messageID <= 0 {
		http.Error(w, `{"message":"invalid messageId"}`, http.StatusBadRequest)
		return
	}

	emoji := ps.ByName("emoji")
	emoji = strings.TrimSpace(emoji)
	if emoji == "" {
		http.Error(w, `{"message":"emoji is required"}`, http.StatusBadRequest)
		return
	}

	err = rt.db.RemoveReaction(r.Context(), ctx.UserIdentifier, messageID, emoji)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, `{"message":"reaction not found"}`, http.StatusNotFound)
			return
		}
		ctx.Logger.WithError(err).Error("cannot remove reaction")
		http.Error(w, `{"message":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// deleteMessage handles DELETE /messages/:messageId.
func (rt *_router) deleteMessage(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if ctx.UserIdentifier == "" {
		http.Error(w, `{"message":"missing or invalid Authorization header"}`, http.StatusUnauthorized)
		return
	}

	msgIDStr := ps.ByName("messageId")
	messageID, err := strconv.ParseInt(msgIDStr, 10, 64)
	if err != nil || messageID <= 0 {
		http.Error(w, `{"message":"invalid messageId"}`, http.StatusBadRequest)
		return
	}

	err = rt.db.DeleteMessage(r.Context(), ctx.UserIdentifier, messageID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, `{"message":"message not found or not owned by user"}`, http.StatusNotFound)
			return
		}
		ctx.Logger.WithError(err).Error("cannot delete message")
		http.Error(w, `{"message":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
