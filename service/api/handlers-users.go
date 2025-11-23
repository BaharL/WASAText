package api

import (
	"net/http"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// setMyUserName handles PUT /users/username.
// This will update the authenticated user's username.
func (rt *_router) setMyUserName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.WriteHeader(http.StatusNotImplemented)
}

// listUsers handles GET /users.
// This will return the list of users matching the 'search' query parameter.
func (rt *_router) listUsers(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.WriteHeader(http.StatusNotImplemented)
}

// setMyPhoto handles PUT /users/photo.
// This will upload or replace the authenticated user's profile picture.
func (rt *_router) setMyPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.WriteHeader(http.StatusNotImplemented)
}
