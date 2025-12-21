package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// -----------------------------------------------------------------------------
// Request / response models
// -----------------------------------------------------------------------------

type changeUsernameRequest struct {
	Username string `json:"username"`
}

type userSummary struct {
	Identifier string `json:"identifier"`
	Username   string `json:"username"`
}

// Username: 3–16 chars, letters/numbers/_/-
var usernameRegexp = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,16}$`)

// -----------------------------------------------------------------------------
// PUT /users/username — Update my username
// -----------------------------------------------------------------------------

func (rt *_router) setMyUserName(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if ctx.UserIdentifier == "" {
		writeJSON(w, http.StatusUnauthorized, errorMsg("missing or invalid Authorization header"))
		return
	}

	var req changeUsernameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorMsg("invalid JSON body"))
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	if !usernameRegexp.MatchString(req.Username) {
		writeJSON(w, http.StatusBadRequest, errorMsg("invalid username"))
		return
	}

	if err := rt.db.SetUserName(r.Context(), ctx.UserIdentifier, req.Username); err != nil {
		ctx.Logger.WithError(err).Error("cannot update username")
		writeJSON(w, http.StatusInternalServerError, errorMsg("internal server error"))
		return
	}

	// Manteniamo 204 (No Content) come nella versione originale / nei test
	w.WriteHeader(http.StatusNoContent)
}

// -----------------------------------------------------------------------------
// GET /users — Search users by ?search=
// -----------------------------------------------------------------------------

func (rt *_router) listUsers(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	// Siccome in OpenAPI c'è il bearer, proteggo anche qui
	if ctx.UserIdentifier == "" {
		writeJSON(w, http.StatusUnauthorized, errorMsg("missing or invalid Authorization header"))
		return
	}

	search := strings.TrimSpace(r.URL.Query().Get("search"))
	if search == "" {
		// Stesso comportamento logico di prima: se non c'è search → lista vuota
		writeJSON(w, http.StatusOK, []userSummary{})
		return
	}

	users, err := rt.db.SearchUsers(r.Context(), search)
	if err != nil {
		ctx.Logger.WithError(err).Error("cannot search users")
		writeJSON(w, http.StatusInternalServerError, errorMsg("internal server error"))
		return
	}

	resp := make([]userSummary, 0, len(users))
	for _, u := range users {
		resp = append(resp, userSummary{
			Identifier: u.Identifier,
			Username:   u.Name,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

// -----------------------------------------------------------------------------
// PUT /users/photo — Upload real profile photo
// -----------------------------------------------------------------------------

func (rt *_router) setMyPhoto(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	// Verifica che l'utente sia autenticato
	if ctx.UserIdentifier == "" {
		writeJSON(w, http.StatusUnauthorized, errorMsg("missing or invalid Authorization header"))
		return
	}

	// Limite dimensione (es. 10 MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeJSON(w, http.StatusBadRequest, errorMsg("invalid multipart form"))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorMsg("missing file field"))
		return
	}
	defer file.Close()

	// Tipo MIME dal header
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		// Prova a rilevarlo leggendo qualche byte
		buf := make([]byte, 512)
		n, _ := file.Read(buf)
		contentType = http.DetectContentType(buf[:n])
		// Riavvolgi lo stream
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			writeJSON(w, http.StatusInternalServerError, errorMsg("cannot reset file reader"))
			return
		}
	}

	// Riutilizziamo i controlli di /media
	if !isValidMediaType(contentType) {
		writeJSON(w, http.StatusBadRequest, errorMsg("unsupported media type"))
		return
	}

	ext := getExtensionFromContentType(contentType)
	if ext == "" {
		ext = filepath.Ext(header.Filename)
	}
	if ext == "" {
		ext = ".bin"
	}

	// Crea cartella uploads/profiles se non esiste
	if err := os.MkdirAll("./uploads/profiles", 0o755); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorMsg("cannot create profiles directory"))
		return
	}

	// Salviamo la foto con nome basato sull'identifier dell'utente
	filename := fmt.Sprintf("%s%s", ctx.UserIdentifier, ext)
	dstPath := filepath.Join("./uploads/profiles", filename)

	dst, err := os.Create(dstPath)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorMsg("cannot create destination file"))
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorMsg("cannot save file"))
		return
	}

	photoURL := "/uploads/profiles/" + filename

	// Risposta: 200 + URL della foto
	writeJSON(w, http.StatusOK, map[string]string{
		"photoUrl": photoURL,
	})
}
