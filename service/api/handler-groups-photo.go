package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

const (
	groupsUploadDir       = "./uploads/groups"
	groupsUploadURLPrefix = "/v1/uploads/groups/"
)

func (rt *_router) setGroupPhoto(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if ctx.UserIdentifier == "" {
		writeJSON(w, http.StatusUnauthorized, errorMsg("missing or invalid Authorization header"))
		return
	}

	chatID, err := strconv.ParseInt(ps.ByName("chatId"), 10, 64)
	if err != nil || chatID <= 0 {
		writeJSON(w, http.StatusBadRequest, errorMsg("invalid chatId"))
		return
	}

	// Max 10MB
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

	// Detect MIME
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		buf := make([]byte, 512)
		n, _ := file.Read(buf)
		contentType = http.DetectContentType(buf[:n])
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			writeJSON(w, http.StatusInternalServerError, errorMsg("cannot reset file reader"))
			return
		}
	}

	if !isValidMediaType(contentType) {
		writeJSON(w, http.StatusBadRequest, errorMsg("unsupported media type"))
		return
	}

	ext := getExtensionFromContentType(contentType)
	if ext == "" {
		ext = strings.ToLower(filepath.Ext(header.Filename))
	}
	if ext == "" {
		ext = defaultImageExt
	}

	// Ensure folder
	if err := os.MkdirAll(groupsUploadDir, 0o755); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorMsg("cannot create groups directory"))
		return
	}

	filename := fmt.Sprintf("group-%d%s", chatID, ext)
	dstPath := filepath.Join(groupsUploadDir, filename)

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

	// Served by ServeFiles: /v1/uploads/...
	photoURL := groupsUploadURLPrefix + filename

	if err := rt.db.SetGroupPhoto(r.Context(), ctx.UserIdentifier, chatID, photoURL); err != nil {
		ctx.Logger.WithError(err).Error("cannot persist group photo url")
		writeJSON(w, http.StatusInternalServerError, errorMsg("internal server error"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"photoUrl": photoURL})
}
