package api

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// -----------------------------------------------------------------------------
// Request / response models
// -----------------------------------------------------------------------------

type changeUsernameRequest struct {
	Username string `json:"username"`
}

type userSummary struct {
	Identifier string `json:"identifier"`
	Username   string `json:"username"`
}

// Username: 3–16 chars, letters/numbers/_/-
var usernameRegexp = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,16}$`)

// -----------------------------------------------------------------------------
// PUT /users/username — Update my username
// -----------------------------------------------------------------------------

func (rt *_router) setMyUserName(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if ctx.UserIdentifier == "" {
		writeJSON(w, http.StatusUnauthorized, errorMsg("missing or invalid Authorization header"))
		return
	}

	var req changeUsernameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorMsg("invalid JSON body"))
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	if !usernameRegexp.MatchString(req.Username) {
		writeJSON(w, http.StatusBadRequest, errorMsg("invalid username"))
		return
	}

	if err := rt.db.SetUserName(r.Context(), ctx.UserIdentifier, req.Username); err != nil {
		ctx.Logger.WithError(err).Error("cannot update username")
		writeJSON(w, http.StatusInternalServerError, errorMsg("internal server error"))
		return
	}

	// Manteniamo 204 (No Content) come nella versione originale / nei test
	w.WriteHeader(http.StatusNoContent)
}

// -----------------------------------------------------------------------------
// GET /users — Search users by ?search=
// -----------------------------------------------------------------------------

func (rt *_router) listUsers(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	// Siccome in OpenAPI c'è il bearer, proteggo anche qui
	if ctx.UserIdentifier == "" {
		writeJSON(w, http.StatusUnauthorized, errorMsg("missing or invalid Authorization header"))
		return
	}

	search := strings.TrimSpace(r.URL.Query().Get("search"))
	if search == "" {
		// Stesso comportamento logico di prima: se non c'è search → lista vuota
		writeJSON(w, http.StatusOK, []userSummary{})
		return
	}

	users, err := rt.db.SearchUsers(r.Context(), search)
	if err != nil {
		ctx.Logger.WithError(err).Error("cannot search users")
		writeJSON(w, http.StatusInternalServerError, errorMsg("internal server error"))
		return
	}

	resp := make([]userSummary, 0, len(users))
	for _, u := range users {
		resp = append(resp, userSummary{
			Identifier: u.Identifier,
			Username:   u.Name,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

// -----------------------------------------------------------------------------
// PUT /users/photo — Stub (HW3)
// -----------------------------------------------------------------------------

func (rt *_router) setMyPhoto(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	_ = ctx
	w.WriteHeader(http.StatusNotImplemented)
}
