package api

import (
	"encoding/json"
	"net/http"
)

// errorResponse è il formato standard per gli errori JSON
// { "message": "..." }
type errorResponse struct {
	Message string `json:"message"`
}

// writeJSON scrive una risposta JSON con status code e body generico.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// writeErrorJSON scrive una risposta di errore con il campo "message".
func writeErrorJSON(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Message: msg})
}

// readJSON legge il body e fa il Decode nel valore passato.
func readJSON(r *http.Request, dst interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(dst)
}
