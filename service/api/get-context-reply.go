package api

import (
    "encoding/json"
    "net/http"
    "os"
    "path/filepath"

    "git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
    "github.com/julienschmidt/httprouter"
)

type contextReply struct {
    UserIdentifier string `json:"userIdentifier"`
    Username       string `json:"username"`
    PhotoURL       string `json:"photoUrl,omitempty"`
}

func (rt *_router) getContextReply(
    w http.ResponseWriter,
    r *http.Request,
    _ httprouter.Params,
    ctx reqcontext.RequestContext,
) {
    // 1) Controllo autenticazione
    if ctx.UserIdentifier == "" {
        w.WriteHeader(http.StatusUnauthorized)
        return
    }

    // 2) Prendo username dal DB
    username, err := rt.db.GetName(r.Context(), ctx.UserIdentifier)
    if err != nil {
        ctx.Logger.WithError(err).Error("cannot load username")
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    // 3) Cerco la foto profilo se esiste
    possibleExtensions := []string{".png", ".jpg", ".jpeg", ".webp"}
    var photoURL string

    for _, ext := range possibleExtensions {
        p := filepath.Join("./uploads/profiles", ctx.UserIdentifier+ext)
        if _, err := os.Stat(p); err == nil {
            photoURL = "/uploads/profiles/" + ctx.UserIdentifier + ext
            break
        }
    }

    // 4) Costruisco risposta JSON
    reply := contextReply{
        UserIdentifier: ctx.UserIdentifier,
        Username:       username,
        PhotoURL:       photoURL, // vuoto se nessuna foto presente
    }

    w.Header().Set("Content-Type", "application/json")
    
    if err := json.NewEncoder(w).Encode(reply); err != nil {
    	ctx.Logger.WithError(err).Error("cannot encode context reply")
    	http.Error(w, "failed to encode response", http.StatusInternalServerError)
    	return
    }

}
