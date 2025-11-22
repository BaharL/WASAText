package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// LoginRequest matches the structure defined in api.yaml
type LoginRequest struct {
	Name string `json:"name"`
}

// LoginResponse matches the structure defined in api.yaml
type LoginResponse struct {
	Identifier string `json:"identifier"`
}

// doLogin implements POST /session
func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	// Only POST method is allowed
	if r.Method != http.MethodPost {
		http.Error(w, `{"message":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Parse JSON body into LoginRequest
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"message":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}

	// Simple validation for the "name" field
	if len(req.Name) < 3 || len(req.Name) > 16 {
		http.Error(w, `{"message":"name must be between 3 and 16 characters"}`, http.StatusBadRequest)
		return
	}

	// Here we must interact with the database.
	// Idea:
	//   - If the user already exists → return the existing identifier
	//   - If the user does NOT exist → create a new one and return the new identifier
	//
	// This method (LoginOrCreateUser) will be implemented later inside service/database.
	identifier, err := rt.db.LoginOrCreateUser(r.Context(), req.Name)
	if err != nil {
		// Later we can define specific error types (e.g., ErrConflict, ErrBadRequest)
		http.Error(w, `{"message":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	// According to the OpenAPI spec (api.yaml), this endpoint must return HTTP 201
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	resp := LoginResponse{
		Identifier: identifier,
	}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		// In case encoding fails, return 500 (also log later)
		http.Error(w, `{"message":"internal server error"}`, http.StatusInternalServerErr
