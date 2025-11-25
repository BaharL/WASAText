package api

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

//
// Request / response models
//

// changeUsernameRequest represents the JSON body used to update a username.
type changeUsernameRequest struct {
	Username string `json:"username"`
}

type userSummary struct {
	Identifier string `json:"identifier"`
	Username   string `json:"username"`
}

// usernameRegexp validates allowed usernames (3–16 chars, letters/numbers/_/-).
var usernameRegexp = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,16}$`)

//
// Handlers
//

// setMyUserName handles PUT /users/username.
// It updates the authenticated user's username in the database.
func (rt *_router) setMyUserName(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	// 1) Utente deve essere autenticato
	if ctx.UserIdentifier == "" {
		http.Error(w, `{"message":"missing or invalid Authorization header"}`, http.StatusUnauthorized)
		return
	}

	// 2) Leggo il body JSON
	var req changeUsernameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"message":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}

	// 3) Valido il formato dello username
	if !usernameRegexp.MatchString(req.Username) {
		http.Error(w, `{"message":"invalid username"}`, http.StatusBadRequest)
		return
	}

	// 4) Aggiorno nel database
	if err := rt.db.SetUserName(r.Context(), ctx.UserIdentifier, req.Username); err != nil {
		ctx.Logger.WithError(err).Error("cannot update username")
		http.Error(w, `{"message":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	// 5) Tutto ok → risposta 204, niente body
	w.WriteHeader(http.StatusNoContent)
}

// listUsers handles GET /users.
// It returns the list of users matching the 'search' query parameter.
func (rt *_router) listUsers(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	// Leggo il parametro ?search=...
	search := strings.TrimSpace(r.URL.Query().Get("search"))
	if search == "" {
		// Scelta semplice: se non c'è search, ritorniamo lista vuota.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode([]userSummary{})
		return
	}

	// Chiamo il database
	users, err := rt.db.SearchUsers(r.Context(), search)
	if err != nil {
		ctx.Logger.WithError(err).Error("cannot search users")
		http.Error(w, `{"message":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Mappo i risultati in userSummary (solo id + username)
	resp := make([]userSummary, 0, len(users))
	for _, u := range users {
		resp = append(resp, userSummary{
			Identifier: u.Identifier,
			Username:   u.Name,
		})
	}

	// Rispondo con JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// setMyPhoto handles PUT /users/photo.
// This uploads or replaces the authenticated user's profile picture.
// (Da implementare in HW3)
func (rt *_router) setMyPhoto(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	_ = ctx
	w.WriteHeader(http.StatusNotImplemented)
}
