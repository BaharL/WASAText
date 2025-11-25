package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"git.sapienzapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// listMyConversations gestisce GET /conversations
// Restituisce la lista delle conversazioni dell'utente autenticato.
func (rt *_router) listMyConversations(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	// 1) Utente deve essere autenticato
	if ctx.UserIdentifier == "" {
		http.Error(w, `{"message":"missing or invalid Authorization header"}`, http.StatusUnauthorized)
		return
	}

	// 2) Chiamo il database
	convs, err := rt.db.ListUserConversations(r.Context(), ctx.UserIdentifier)
	if err != nil {
		http.Error(w, `{"message":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	// 3) Rispondo in JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(convs); err != nil {
		http.Error(w, `{"message":"internal server error"}`, http.StatusInternalServerError)
		return
	}
}

// listConversationMessages gestisce GET /conversations/{id}/messages
// Restituisce i messaggi di una specifica conversazione.
func (rt *_router) listConversationMessages(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	// 1) Utente deve essere autenticato
	if ctx.UserIdentifier == "" {
		http.Error(w, `{"message":"missing or invalid Authorization header"}`, http.StatusUnauthorized)
		return
	}

	// 2) Estraggo l'id della conversazione dall'URL
	chatIDStr := ps.ByName("conversationId") // <-- adatta al nome usato nel tuo router/YAML
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		http.Error(w, `{"message":"invalid conversation id"}`, http.StatusBadRequest)
		return
	}

	// (opzionale) potresti leggere query param tipo ?limit=50, ?offset=0

	// 3) Chiamo il database
	msgs, err := rt.db.ListConversationMessages(r.Context(), ctx.UserIdentifier, chatID)
	if err != nil {
		// puoi distinguere 404/403 se nel db controlli accesso/partecipazione
		http.Error(w, `{"message":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	// 4) Risposta JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(msgs); err != nil {
		http.Error(w, `{"message":"internal server error"}`, http.StatusInternalServerError)
		return
	}
}
