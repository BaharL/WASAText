package api

import (
	"encoding/json"
	"net/http"
	"regexp"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

//
// Request and response models
//

// changeUsernameRequest represents the JSON body used to update a username.
type changeUsernameRequest struct {
	Username string `json:"username"`
}

// userResponse represents a minimal user object returned after the update.
type userResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

// usernameRegexp validates allowed usernames (3–16 chars, letters/numbers/_/-)
var usernameRegexp = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,16}$`)

//
// Handlers
//

// setMyUserName handles PUT /users/username.
// This updates the authenticated user's username.
func (rt *_router) setMyUserName(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	// Parse JSON body
	var req changeUsernameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"message":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}

	// Validate username format
	if !usernameRegexp.MatchString(req.Username) {
		http.Error(w, `{"message":"invalid username"}`, http.StatusBadRequest)
		return
	}

	// TODO: replace with database operation using ctx.UserID
	updatedUser := userResponse{
		ID:       0, // placeholder until HW2 database implementation
		Username: req.Username,
	}

	// Return updated user object
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(updatedUser)
}

// listUsers handles GET /users.
// This returns the list of users matching ?search= query.
func (rt *_router) listUsers(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	w.WriteHeader(http.StatusNotImplemented)
}

// setMyPhoto handles PUT /users/photo.
// This uploads or replaces the authenticated user's profile picture.
func (rt *_router) setMyPhoto(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	w.WriteHeader(http.StatusNotImplemented)
}
