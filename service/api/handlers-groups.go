package api

import (
	"net/http"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// createGroup handles POST /groups.
// It should create a new group conversation with the specified name and members.
func (rt *_router) createGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.WriteHeader(http.StatusNotImplemented)
}

// addToGroup handles POST /groups/:chatId/members.
// It should add one or more members to the target group conversation.
func (rt *_router) addToGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.WriteHeader(http.StatusNotImplemented)
}

// leaveGroup handles DELETE /groups/:chatId/members/me.
// It should remove the authenticated user from the specified group.
func (rt *_router) leaveGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.WriteHeader(http.StatusNotImplemented)
}

// setGroupName handles PUT /groups/:chatId/name.
// It should update the name of an existing group conversation.
func (rt *_router) setGroupName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.WriteHeader(http.StatusNotImplemented)
}

// setGroupPhoto handles PUT /groups/:chatId/photo.
// It should upload or change the photo associated with a group conversation.
func (rt *_router) setGroupPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.WriteHeader(http.StatusNotImplemented)
}
