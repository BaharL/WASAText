package api

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// listMyConversations handles GET /conversations.
//
// Rules:
//   - If Authorization is missing/invalid => 401
//   - If the token is valid but references a user that does not exist anymore
//     (stale session / DB reset / deleted user) => 401 (NOT 500)
//   - Only real internal errors => 500
func (rt *_router) listMyConversations(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	// Middleware should set ctx.UserIdentifier from the token.
	// If empty => no valid auth.
	if ctx.UserIdentifier == "" {
		writeJSON(w, http.StatusUnauthorized, errorMsg("missing or invalid Authorization header"))
		return
	}

	convs, err := rt.db.ListUserConversations(r.Context(), ctx.UserIdentifier)
	if err != nil {

		// If user does not exist in DB, it is NOT an internal server error:
		// it means the client has an old token (stale session).
		if errors.Is(err, sql.ErrNoRows) {
			ctx.Logger.WithError(err).Warn("user not found while listing conversations (stale session)")
			writeJSON(w, http.StatusUnauthorized, errorMsg("invalid or expired session"))
			return
		}

		// Any other error is a real server-side problem.
		ctx.Logger.WithError(err).Error("cannot list conversations")
		writeJSON(w, http.StatusInternalServerError, errorMsg("internal server error"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"conversations": convs,
	})
}

// listConversationMessages handles GET /conversations/{chatId}.
//
// Rules:
// - If Authorization is missing/invalid => 401
// - If chat does not exist OR user is not a member => 404 (or 403, depending on your design)
// - Only real internal errors => 500
func (rt *_router) listConversationMessages(
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

	msgs, err := rt.db.ListConversationMessages(r.Context(), ctx.UserIdentifier, chatID)
	if err != nil {

		// In your DB layer, sql.ErrNoRows can mean:
		// - chat does not exist
		// - OR user is not a member of that chat
		// Returning 404 avoids leaking information and is common in chat apps.
		if errors.Is(err, sql.ErrNoRows) {
			ctx.Logger.WithError(err).Warn("chat not found or user not a member")
			writeJSON(w, http.StatusNotFound, errorMsg("conversation not found"))
			return
		}

		ctx.Logger.WithError(err).Error("cannot list conversation messages")
		writeJSON(w, http.StatusInternalServerError, errorMsg("internal server error"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"messages": msgs,
	})
}
