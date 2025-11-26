package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// Request models for group operations.

type createGroupRequest struct {
	Name    string   `json:"name"`
	Members []string `json:"members"` // user identifiers (same as in /session response)
}

type addMembersRequest struct {
	Members []string `json:"members"`
}

type setGroupNameRequest struct {
	Name string `json:"name"`
}

type setGroupPhotoRequest struct {
	PhotoURL string `json:"photoUrl"`
}

type createGroupResponse struct {
	ChatID int64 `json:"chatId"`
}

// createGroup handles POST /groups.
// It creates a new group conversation with the specified name and members.
func (rt *_router) createGroup(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	// 1) User must be authenticated.
	if ctx.UserIdentifier == "" {
		http.Error(w, `{"message":"missing or invalid Authorization header"}`, http.StatusUnauthorized)
		return
	}

	// 2) Parse JSON body.
	var req createGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"message":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		http.Error(w, `{"message":"group name is required"}`, http.StatusBadRequest)
		return
	}

	// If members is nil, use empty slice to avoid nil handling in DB.
	if req.Members == nil {
		req.Members = []string{}
	}

	// 3) Call database to create the group.
	chatID, err := rt.db.CreateGroup(r.Context(), ctx.UserIdentifier, req.Name, req.Members)
	if err != nil {
		ctx.Logger.WithError(err).Error("cannot create group")
		http.Error(w, `{"message":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	// 4) Respond with created group id.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(createGroupResponse{
		ChatID: chatID,
	})
}

// addToGroup handles POST /groups/:chatId/members.
// It adds one or more members to the target group conversation.
func (rt *_router) addToGroup(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if ctx.UserIdentifier == "" {
		http.Error(w, `{"message":"missing or invalid Authorization header"}`, http.StatusUnauthorized)
		return
	}

	chatIDStr := ps.ByName("chatId")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil || chatID <= 0 {
		http.Error(w, `{"message":"invalid chatId"}`, http.StatusBadRequest)
		return
	}

	var req addMembersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"message":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}

	if len(req.Members) == 0 {
		http.Error(w, `{"message":"members list cannot be empty"}`, http.StatusBadRequest)
		return
	}

	if err := rt.db.AddMembersToGroup(r.Context(), ctx.UserIdentifier, chatID, req.Members); err != nil {
		ctx.Logger.WithError(err).Error("cannot add members to group")
		http.Error(w, `{"message":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// leaveGroup handles DELETE /groups/:chatId/members/me.
// It removes the authenticated user from the specified group.
func (rt *_router) leaveGroup(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if ctx.UserIdentifier == "" {
		http.Error(w, `{"message":"missing or invalid Authorization header"}`, http.StatusUnauthorized)
		return
	}

	chatIDStr := ps.ByName("chatId")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil || chatID <= 0 {
		http.Error(w, `{"message":"invalid chatId"}`, http.StatusBadRequest)
		return
	}

	if err := rt.db.LeaveGroup(r.Context(), ctx.UserIdentifier, chatID); err != nil {
		// Per semplicità, non distinguiamo 404/403, ma puoi farlo in seguito.
		ctx.Logger.WithError(err).Error("cannot leave group")
		http.Error(w, `{"message":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// setGroupName handles PUT /groups/:chatId/name.
// It updates the name of an existing group conversation.
func (rt *_router) setGroupName(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if ctx.UserIdentifier == "" {
		http.Error(w, `{"message":"missing or invalid Authorization header"}`, http.StatusUnauthorized)
		return
	}

	chatIDStr := ps.ByName("chatId")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil || chatID <= 0 {
		http.Error(w, `{"message":"invalid chatId"}`, http.StatusBadRequest)
		return
	}

	var req setGroupNameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"message":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		http.Error(w, `{"message":"group name is required"}`, http.StatusBadRequest)
		return
	}

	if err := rt.db.SetGroupName(r.Context(), ctx.UserIdentifier, chatID, req.Name); err != nil {
		ctx.Logger.WithError(err).Error("cannot set group name")
		http.Error(w, `{"message":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Group name updated", // o "Group photo updated"
	})

}

// setGroupPhoto handles PUT /groups/:chatId/photo.
// It uploads or changes the photo associated with a group conversation.
func (rt *_router) setGroupPhoto(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if ctx.UserIdentifier == "" {
		http.Error(w, `{"message":"missing or invalid Authorization header"}`, http.StatusUnauthorized)
		return
	}

	chatIDStr := ps.ByName("chatId")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil || chatID <= 0 {
		http.Error(w, `{"message":"invalid chatId"}`, http.StatusBadRequest)
		return
	}

	var req setGroupPhotoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"message":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}
	req.PhotoURL = strings.TrimSpace(req.PhotoURL)
	if req.PhotoURL == "" {
		http.Error(w, `{"message":"photoUrl is required"}`, http.StatusBadRequest)
		return
	}

	if err := rt.db.SetGroupPhoto(r.Context(), ctx.UserIdentifier, chatID, req.PhotoURL); err != nil {
		ctx.Logger.WithError(err).Error("cannot set group photo")
		http.Error(w, `{"message":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
