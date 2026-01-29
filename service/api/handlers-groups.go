package api

import (
	"net/http"
	"strconv"
	"strings"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// --- Request Models ---------------------------------------------------------

type createGroupRequest struct {
	Name    string   `json:"name"`
	Members []string `json:"members"`
}

type addMembersRequest struct {
	Members []string `json:"members"`
}

type setGroupNameRequest struct {
	Name string `json:"name"`
}

type createGroupResponse struct {
	ChatID int64 `json:"chatId"`
}

// ---------------------------------------------------------------------------
// createGroup  → POST /groups
// ---------------------------------------------------------------------------

func (rt *_router) createGroup(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if ctx.UserIdentifier == "" {
		writeErrorJSON(w, http.StatusUnauthorized, "Invalid or missing token")
		return
	}

	var req createGroupRequest
	if err := readJSON(r, &req); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeErrorJSON(w, http.StatusBadRequest, "Group name is required")
		return
	}

	if req.Members == nil {
		req.Members = []string{}
	}

	chatID, err := rt.db.CreateGroup(r.Context(), ctx.UserIdentifier, req.Name, req.Members)
	if err != nil {
		ctx.Logger.WithError(err).Error("cannot create group")
		writeErrorJSON(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, createGroupResponse{ChatID: chatID})
}

// ---------------------------------------------------------------------------
// addToGroup  → POST /groups/:chatId/members
// ---------------------------------------------------------------------------

func (rt *_router) addToGroup(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if ctx.UserIdentifier == "" {
		writeErrorJSON(w, http.StatusUnauthorized, "Invalid or missing token")
		return
	}

	chatID, err := strconv.ParseInt(ps.ByName("chatId"), 10, 64)
	if err != nil || chatID <= 0 {
		writeErrorJSON(w, http.StatusBadRequest, "Invalid chatId")
		return
	}

	var req addMembersRequest
	if err := readJSON(r, &req); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	if len(req.Members) == 0 {
		writeErrorJSON(w, http.StatusBadRequest, "Members list cannot be empty")
		return
	}

	if err := rt.db.AddMembersToGroup(r.Context(), ctx.UserIdentifier, chatID, req.Members); err != nil {
		ctx.Logger.WithError(err).Error("cannot add members to group")
		writeErrorJSON(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ---------------------------------------------------------------------------
// leaveGroup  → DELETE /groups/:chatId/members/me
// ---------------------------------------------------------------------------

func (rt *_router) leaveGroup(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if ctx.UserIdentifier == "" {
		writeErrorJSON(w, http.StatusUnauthorized, "Invalid or missing token")
		return
	}

	chatID, err := strconv.ParseInt(ps.ByName("chatId"), 10, 64)
	if err != nil || chatID <= 0 {
		writeErrorJSON(w, http.StatusBadRequest, "Invalid chatId")
		return
	}

	if err := rt.db.LeaveGroup(r.Context(), ctx.UserIdentifier, chatID); err != nil {
		ctx.Logger.WithError(err).Error("cannot leave group")
		writeErrorJSON(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ---------------------------------------------------------------------------
// setGroupName  → PUT /groups/:chatId/name
// ---------------------------------------------------------------------------

func (rt *_router) setGroupName(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if ctx.UserIdentifier == "" {
		writeErrorJSON(w, http.StatusUnauthorized, "Invalid or missing token")
		return
	}

	chatID, err := strconv.ParseInt(ps.ByName("chatId"), 10, 64)
	if err != nil || chatID <= 0 {
		writeErrorJSON(w, http.StatusBadRequest, "Invalid chatId")
		return
	}

	var req setGroupNameRequest
	if err := readJSON(r, &req); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeErrorJSON(w, http.StatusBadRequest, "Group name is required")
		return
	}

	if err := rt.db.SetGroupName(r.Context(), ctx.UserIdentifier, chatID, req.Name); err != nil {
		ctx.Logger.WithError(err).Error("cannot set group name")
		writeErrorJSON(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Group name updated",
	})
}
