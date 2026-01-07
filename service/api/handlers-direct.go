package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

type createDirectRequest struct {
	Members []string `json:"members"`
}

type createDirectResponse struct {
	ChatID int64 `json:"chatId"`
}

func (rt *_router) createDirect(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	var req createDirectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorMsg("invalid JSON body"))
		return
	}

	if len(req.Members) != 1 {
		writeJSON(w, http.StatusBadRequest, errorMsg("members must contain exactly 1 user identifier"))
		return
	}

	other := strings.TrimSpace(req.Members[0])
	if other == "" {
		writeJSON(w, http.StatusBadRequest, errorMsg("member identifier cannot be empty"))
		return
	}

	// ctx.UserIdentifier (o come si chiama nel tuo reqcontext) deve essere l'identifier del bearer
	chatID, err := rt.db.GetOrCreateDirectChat(r.Context(), ctx.UserIdentifier, other)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorMsg("internal server error"))
		return
	}

	writeJSON(w, http.StatusCreated, createDirectResponse{ChatID: chatID})
}
