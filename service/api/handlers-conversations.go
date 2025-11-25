package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// listMyConversations handles GET /conversations.
// It returns the list of conversations for the authenticated user.
func (rt *_router) listMyConversations(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	// 1) User must be authenticated.
	if ctx.UserIdentifier == "" {
		http.Error(w, `{"message":"missing or invalid Authorization header"}`, http.StatusUnauthorized)
		return
	}

	// 2) Ask the database for the conversations.
	convs, err := rt.db.ListUserConversations(r.Context(), ctx.UserIdentifier)
	if err != nil {
		http.Error(w, `{"message":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	// 3) Return JSON response.
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(convs); err != nil {
		http.Error(w, `{"message":"internal server error"}`, http.StatusInternalServerError)
		return
	}
}

// listConversationMessages handles GET /conversations/{id}/messages.
// It returns all messages of a specific conversation.
func (rt *_router) listConversationMessages(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	// 1) User must be authenticated.
	if ctx.UserIdentifier == "" {
		http.Error(w, `{"message":"missing or invalid Authorization header"}`, http.StatusUnauthorized)
		return
	}

	// 2) Read conversation id from URL path.
	//    IMPORTANT: adjust "conversationId" to match your router/YAML parameter name (e.g. "chatId").
	chatIDStr := ps.ByName("conversationId")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		http.Error(w, `{"message":"invalid conversation id"}`, http.StatusBadRequest)
		return
	}

	// 3) Ask the database for the messages of this conversation.
	msgs, err := rt.db.ListConversationMessages(r.Context(), ctx.UserIdentifier, chatID)
	if err != nil {
		http.Error(w, `{"message":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	// 4) Return JSON response.
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(msgs); err != nil {
		http.Error(w, `{"message":"internal server error"}`, http.StatusInternalServerError)
		return
	}
}
