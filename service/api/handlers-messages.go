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

// Request models aligned with OpenAPI

type sendMessageRequest struct {
	ChatID           int64   `json:"chatId"`
	Kind             string  `json:"kind"`
	Text             *string `json:"text,omitempty"`
	MediaURL         *string `json:"mediaUrl,omitempty"`
	ReplyToMessageID *int64  `json:"replyToMessageId,omitempty"`
}

type forwardMessageRequest struct {
	ToChatID int64 `json:"toChatId"`
}

type addReactionRequest struct {
	Emoji string `json:"emoji"`
}

// -----------------------------------------------------------------------------
// POST /messages — Send a new message
// -----------------------------------------------------------------------------

func (rt *_router) sendMessage(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if ctx.UserIdentifier == "" {
		writeJSON(w, http.StatusUnauthorized, errorMsg("missing or invalid Authorization header"))
		return
	}

	var req sendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorMsg("invalid JSON body"))
		return
	}

	req.Kind = strings.ToLower(strings.TrimSpace(req.Kind))

	if req.ChatID <= 0 {
		writeJSON(w, http.StatusBadRequest, errorMsg("chatId must be positive"))
		return
	}
	if req.Kind != "text" && req.Kind != "gif" && req.Kind != "image" {
		writeJSON(w, http.StatusBadRequest, errorMsg("invalid kind"))
		return
	}

	// Validation according to kind
	if req.Kind == "text" {
		if req.Text == nil || strings.TrimSpace(*req.Text) == "" {
			writeJSON(w, http.StatusBadRequest, errorMsg("text is required for text messages"))
			return
		}
		req.MediaURL = nil
	} else if req.MediaURL == nil || strings.TrimSpace(*req.MediaURL) == "" {
		writeJSON(w, http.StatusBadRequest, errorMsg("mediaUrl is required for gif/image messages"))
		return
	}

	// Reply validation: if present, must exist and belong to the same chat
	if req.ReplyToMessageID != nil {
		if *req.ReplyToMessageID <= 0 {
			writeJSON(w, http.StatusBadRequest, errorMsg("replyToMessageId must be positive"))
			return
		}

		orig, err := rt.db.GetMessageByID(r.Context(), *req.ReplyToMessageID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeJSON(w, http.StatusNotFound, errorMsg("reply target message not found"))
				return
			}
			ctx.Logger.WithError(err).Error("cannot load reply target message")
			writeJSON(w, http.StatusInternalServerError, errorMsg("internal server error"))
			return
		}
		if orig.ChatID != req.ChatID {
			writeJSON(w, http.StatusBadRequest, errorMsg("reply target must belong to the same chat"))
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
		req.ReplyToMessageID,
	)
	if err != nil {
		ctx.Logger.WithError(err).Error("cannot send message")
		writeJSON(w, http.StatusInternalServerError, errorMsg("internal server error"))
		return
	}

	writeJSON(w, http.StatusCreated, msg)
}

// -----------------------------------------------------------------------------
// POST /messages/:messageId/forward — Forward message
// -----------------------------------------------------------------------------

func (rt *_router) forwardMessage(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if ctx.UserIdentifier == "" {
		writeJSON(w, http.StatusUnauthorized, errorMsg("missing or invalid Authorization header"))
		return
	}

	messageID, err := parseID(ps.ByName("messageId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorMsg("invalid messageId"))
		return
	}

	var req forwardMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorMsg("invalid JSON body"))
		return
	}
	if req.ToChatID <= 0 {
		writeJSON(w, http.StatusBadRequest, errorMsg("toChatId must be positive"))
		return
	}

	msg, err := rt.db.ForwardMessage(r.Context(), ctx.UserIdentifier, messageID, req.ToChatID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, errorMsg("message not found"))
			return
		}
		ctx.Logger.WithError(err).Error("cannot forward message")
		writeJSON(w, http.StatusInternalServerError, errorMsg("internal server error"))
		return
	}

	writeJSON(w, http.StatusCreated, msg)
}

// -----------------------------------------------------------------------------
// POST /messages/:messageId/reactions — Add reaction
// -----------------------------------------------------------------------------

func (rt *_router) commentMessage(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if ctx.UserIdentifier == "" {
		writeJSON(w, http.StatusUnauthorized, errorMsg("missing or invalid Authorization header"))
		return
	}

	messageID, err := parseID(ps.ByName("messageId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorMsg("invalid messageId"))
		return
	}

	var req addReactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorMsg("invalid JSON body"))
		return
	}

	req.Emoji = strings.TrimSpace(req.Emoji)
	if req.Emoji == "" {
		writeJSON(w, http.StatusBadRequest, errorMsg("emoji is required"))
		return
	}

	reaction, err := rt.db.AddReaction(r.Context(), ctx.UserIdentifier, messageID, req.Emoji)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, errorMsg("message not found"))
			return
		}
		ctx.Logger.WithError(err).Error("cannot add reaction")
		writeJSON(w, http.StatusInternalServerError, errorMsg("internal server error"))
		return
	}

	writeJSON(w, http.StatusCreated, reaction)
}

// -----------------------------------------------------------------------------
// DELETE /messages/:messageId/reactions/:emoji — Remove reaction
// -----------------------------------------------------------------------------

func (rt *_router) uncommentMessage(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if ctx.UserIdentifier == "" {
		writeJSON(w, http.StatusUnauthorized, errorMsg("missing or invalid Authorization header"))
		return
	}

	messageID, err := parseID(ps.ByName("messageId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorMsg("invalid messageId"))
		return
	}

	emoji := strings.TrimSpace(ps.ByName("emoji"))
	if emoji == "" {
		writeJSON(w, http.StatusBadRequest, errorMsg("emoji is required"))
		return
	}

	err = rt.db.RemoveReaction(r.Context(), ctx.UserIdentifier, messageID, emoji)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, errorMsg("reaction not found"))
			return
		}
		ctx.Logger.WithError(err).Error("cannot remove reaction")
		writeJSON(w, http.StatusInternalServerError, errorMsg("internal server error"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// -----------------------------------------------------------------------------
// DELETE /messages/:messageId — Delete message
// -----------------------------------------------------------------------------

func (rt *_router) deleteMessage(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if ctx.UserIdentifier == "" {
		writeJSON(w, http.StatusUnauthorized, errorMsg("missing or invalid Authorization header"))
		return
	}

	messageID, err := parseID(ps.ByName("messageId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorMsg("invalid messageId"))
		return
	}

	err = rt.db.DeleteMessage(r.Context(), ctx.UserIdentifier, messageID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, errorMsg("message not found or not owned by user"))
			return
		}
		ctx.Logger.WithError(err).Error("cannot delete message")
		writeJSON(w, http.StatusInternalServerError, errorMsg("internal server error"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

func parseID(s string) (int64, error) {
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid id")
	}
	return id, nil
}

func errorMsg(msg string) map[string]string {
	return map[string]string{"message": msg}
}
