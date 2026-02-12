package api

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// getMyConversations handles GET /conversations.
func (rt *_router) getMyConversations(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if ctx.UserIdentifier == "" {
		writeJSON(w, http.StatusUnauthorized, errorMsg("missing or invalid Authorization header"))
		return
	}

	convs, err := rt.db.ListUserConversations(r.Context(), ctx.UserIdentifier)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Pragmatic stale-session detection
			if strings.Contains(err.Error(), "user not found") {
				ctx.Logger.WithError(err).Warn("stale session while listing conversations")
				writeJSON(w, http.StatusUnauthorized, errorMsg("invalid or expired session"))
				return
			}

			ctx.Logger.WithError(err).Warn("conversation not found")
			writeJSON(w, http.StatusNotFound, errorMsg("conversation not found"))
			return
		}

		ctx.Logger.WithError(err).Error("cannot list conversations")
		writeJSON(w, http.StatusInternalServerError, errorMsg("internal server error"))
		return
	}

	// NB: convs contiene già unreadCount, title, isGroup, photoUrl, lastMessage
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"conversations": convs,
	})
}

// getConversation handles GET /conversations/{chatId}.
//
// IMPORTANT: ritorna anche metadata (title/isGroup/photoUrl) + messages,
// così il frontend può mostrare Leave group, rename, ecc.
func (rt *_router) getConversation(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if ctx.UserIdentifier == "" {
		writeJSON(w, http.StatusUnauthorized, errorMsg("missing or invalid Authorization header"))
		return
	}

	chatIDStr := ps.ByName("chatId")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil || chatID <= 0 {
		writeJSON(w, http.StatusBadRequest, errorMsg("invalid chatId"))
		return
	}

	// 1) Info conversazione (controlla anche membership)
	info, err := rt.db.GetConversationInfo(r.Context(), ctx.UserIdentifier, chatID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.Logger.WithError(err).Warn("chat not found or user not a member (info)")
			writeJSON(w, http.StatusNotFound, errorMsg("conversation not found"))
			return
		}
		ctx.Logger.WithError(err).Error("cannot load conversation info")
		writeJSON(w, http.StatusInternalServerError, errorMsg("internal server error"))
		return
	}

	// 2) Messaggi
	msgs, err := rt.db.ListConversationMessages(r.Context(), ctx.UserIdentifier, chatID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.Logger.WithError(err).Warn("chat not found or user not a member (messages)")
			writeJSON(w, http.StatusNotFound, errorMsg("conversation not found"))
			return
		}
		ctx.Logger.WithError(err).Error("cannot list conversation messages")
		writeJSON(w, http.StatusInternalServerError, errorMsg("internal server error"))
		return
	}

	// 3) Reply completo per il frontend
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":       info.ID,
		"title":    info.Title,
		"isGroup":  info.IsGroup,
		"photoUrl": info.PhotoURL,
		"messages": msgs,
	})
}

// markConversationReceived handles POST /conversations/{chatId}/received.
// It upgrades incoming messages from "sent" to "received".
func (rt *_router) markConversationReceived(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if ctx.UserIdentifier == "" {
		writeJSON(w, http.StatusUnauthorized, errorMsg("missing or invalid Authorization header"))
		return
	}

	chatIDStr := ps.ByName("chatId")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil || chatID <= 0 {
		writeJSON(w, http.StatusBadRequest, errorMsg("invalid chatId"))
		return
	}

	if err := rt.db.MarkConversationReceived(r.Context(), ctx.UserIdentifier, chatID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, errorMsg("conversation not found"))
			return
		}
		ctx.Logger.WithError(err).Error("cannot mark conversation received")
		writeJSON(w, http.StatusInternalServerError, errorMsg("internal server error"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// markConversationRead handles POST /conversations/{chatId}/read.
// It upgrades incoming messages to "read".
func (rt *_router) markConversationRead(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if ctx.UserIdentifier == "" {
		writeJSON(w, http.StatusUnauthorized, errorMsg("missing or invalid Authorization header"))
		return
	}

	chatIDStr := ps.ByName("chatId")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil || chatID <= 0 {
		writeJSON(w, http.StatusBadRequest, errorMsg("invalid chatId"))
		return
	}

	if err := rt.db.MarkConversationRead(r.Context(), ctx.UserIdentifier, chatID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, errorMsg("conversation not found"))
			return
		}
		ctx.Logger.WithError(err).Error("cannot mark conversation read")
		writeJSON(w, http.StatusInternalServerError, errorMsg("internal server error"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
