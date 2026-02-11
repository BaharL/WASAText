//go:build webui
package main

import (
	"fmt"
	"io/fs"
	"net/http"
	"strings"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/webui"
)

func registerWebUI(hdl http.Handler) (http.Handler, error) {
	distDirectory, err := fs.Sub(webui.Dist, "dist")
	if err != nil {
		return nil, fmt.Errorf("error embedding WebUI dist/ directory: %w", err)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.RequestURI, "/dashboard/") {
			http.StripPrefix("/dashboard/", http.FileServer(http.FS(distDirectory))).ServeHTTP(w, r)
			return
		}
		hdl.ServeHTTP(w, r)
	}), nil
}

// applyCORSHandler applica gli header CORS a tutte le richieste
func applyCORSHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ✅ Permetti richieste da qualsiasi origine (necessario per testing su dispositivi esterni)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")
		w.Header().Set("Access-Control-Max-Age", "3600")
		
		// ✅ Gestisci richieste preflight OPTIONS
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		
		h.ServeHTTP(w, r)
	})
}
