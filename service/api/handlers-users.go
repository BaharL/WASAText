package api

import (
	"bytes"
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

		// ✅ username già esistente → 409
		if isUniqueUsernameError(err) {
			writeJSON(w, http.StatusConflict, errorMsg("username already exists"))
			return
		}

		ctx.Logger.WithError(err).Error("cannot update username")
		writeJSON(w, http.StatusInternalServerError, errorMsg("internal server error"))
		return
	}

	// 204 No Content
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
	if ctx.UserIdentifier == "" {
		writeJSON(w, http.StatusUnauthorized, errorMsg("missing or invalid Authorization header"))
		return
	}

	search := strings.TrimSpace(r.URL.Query().Get("search"))
	if search == "" {
		// lista vuota coerente
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
	if ctx.UserIdentifier == "" {
		writeJSON(w, http.StatusUnauthorized, errorMsg("missing or invalid Authorization header"))
		return
	}

	const maxUpload = 10 << 20 // 10MB

	// HARD limit sul body (questo evita hangup e upload enormi)
	r.Body = http.MaxBytesReader(w, r.Body, maxUpload)

	// Parse multipart form (maxMemory)
	if err := r.ParseMultipartForm(maxUpload); err != nil {
		// se supera MaxBytesReader spesso è "http: request body too large"
		if strings.Contains(strings.ToLower(err.Error()), "request body too large") {
			writeJSON(w, http.StatusRequestEntityTooLarge, errorMsg("file too large (max 10MB)"))
			return
		}
		writeJSON(w, http.StatusBadRequest, errorMsg("invalid multipart form"))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorMsg("missing file field"))
		return
	}
	defer file.Close()

	// Sniff sicuro senza Seek
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	detectedType := http.DetectContentType(buf[:n])
	reader := io.MultiReader(bytes.NewReader(buf[:n]), file)

	// Usa detectedType, non fidarti del Content-Type del client
	if !isValidMediaType(detectedType) {
		// se vuoi supportare HEIC/HEIF, aggiungili in isValidMediaType + getExtension...
		writeJSON(w, http.StatusBadRequest, errorMsg("unsupported media type"))
		return
	}

	ext := getExtensionFromContentType(detectedType)
	if ext == "" {
		ext = strings.ToLower(filepath.Ext(header.Filename))
	}
	if ext == "" {
		ext = defaultImageExt
	}

	if err := os.MkdirAll("./uploads/profiles", 0o755); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorMsg("cannot create profiles directory"))
		return
	}

	filename := fmt.Sprintf("%s%s", ctx.UserIdentifier, ext)
	dstPath := filepath.Join("./uploads/profiles", filename)

	dst, err := os.Create(dstPath)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorMsg("cannot create destination file"))
		return
	}
	defer dst.Close()

	// copia usando reader (che include i bytes sniffati)
	// io.Copy(dst, reader)
	if _, err := io.Copy(dst, reader); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorMsg("cannot save file"))
		return
	}

	photoURL := "/v1/uploads/profiles/" + filename

	if err := rt.db.SetUserPhoto(r.Context(), ctx.UserIdentifier, photoURL); err != nil {
		ctx.Logger.WithError(err).Error("cannot persist user photo url")
		writeJSON(w, http.StatusInternalServerError, errorMsg("internal server error"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message":  "Photo updated",
		"photoUrl": photoURL,
	})

}
