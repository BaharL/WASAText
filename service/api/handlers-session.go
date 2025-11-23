package api

import (
	"encoding/json"
	"net/http"

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
func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Enforce POST method (safety check).
	if r.Method != http.MethodPost {
		http.Error(w, `{"message":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Parse JSON body into LoginRequest.
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"message":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}

	// Basic validation for the "name" field.
	if len(req.Name) < 3 || len(req.Name) > 16 {
		http.Error(w, `{"message":"name must be between 3 and 16 characters"}`, http.StatusBadRequest)
		return
	}

	// Interact with the database:
	// - if the user exists → return existing identifier
	// - otherwise → create a new user and return the new identifier.
	identifier, err := rt.db.LoginOrCreateUser(r.Context(), req.Name)
	if err != nil {
		http.Error(w, `{"message":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Build response.
	resp := LoginResponse{
		Identifier: identifier,
	}

	// According to the OpenAPI spec this endpoint returns HTTP 201.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Encode the response as JSON.
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, `{"message":"internal server error"}`, http.StatusInternalServerError)
		return
	}
}
