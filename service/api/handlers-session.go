package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

// LoginRequest mirrors the request body defined in api.yaml.
type LoginRequest struct {
	Name string `json:"name"`
}

// LoginResponse mirrors the response body defined in api.yaml.
type LoginResponse struct {
	Identifier string `json:"identifier"`
}

// doLogin handles POST /session.
// It logs in an existing user or creates a new one.
// This is the only public endpoint (no authentication required).
func (rt *_router) doLogin(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorMsg("invalid JSON body"))
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if len(req.Name) < 3 || len(req.Name) > 16 {
		writeJSON(w, http.StatusBadRequest, errorMsg("name must be between 3 and 16 characters"))
		return
	}

	identifier, err := rt.db.LoginOrCreateUser(r.Context(), req.Name)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorMsg("internal server error"))
		return
	}

	resp := LoginResponse{
		Identifier: identifier,
	}

	// According to the OpenAPI spec this endpoint returns HTTP 201.
	writeJSON(w, http.StatusCreated, resp)
}
