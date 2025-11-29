package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// liveness is an HTTP handler that checks the API server status.
// It returns 200 OK if the service (and the DB) are up, otherwise 500.
func (rt *_router) liveness(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Check DB connection. If Ping fails, the service is not "healthy".
	if err := rt.db.Ping(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
