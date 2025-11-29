package api

import (
	"net/http"
	"strconv"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// listMyConversations handles GET /conversations.
func (rt *_router) listMyConversations(
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
		ctx.Logger.WithError(err).Error("cannot list conversations")
		writeJSON(w, http.StatusInternalServerError, errorMsg("internal server error"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"conversations": convs,
	})
}

// listConversationMessages handles GET /conversations/{chatId}.
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
		ctx.Logger.WithError(err).Error("cannot list conversation messages")
		writeJSON(w, http.StatusInternalServerError, errorMsg("internal server error"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"messages": msgs,
	})
}
