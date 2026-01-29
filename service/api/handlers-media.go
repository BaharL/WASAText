package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

const (
	mediaUploadDir  = "./uploads/media"
	mediaURLPrefix  = "/v1/uploads/media/"
)

// uploadMedia gestisce POST /v1/media (AUTENTICATO)
// Carica un file media (immagine/GIF/WebP) e ritorna un URL pubblico
func (rt *_router) uploadMedia(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if ctx.UserIdentifier == "" {
		writeErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Limite hard: 10MB (utile per evitare body enormi)
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	// Parse multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "File too large or invalid multipart form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "Missing file field")
		return
	}
	defer file.Close()

	// Sniff content-type
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	detectedType := http.DetectContentType(buf[:n])

	if !isValidMediaType(detectedType) {
		writeErrorJSON(w, http.StatusBadRequest, "Invalid file type. Only images and GIFs allowed")
		return
	}

	// Rimetti i bytes letti davanti allo stream originale
	reader := io.MultiReader(bytes.NewReader(buf[:n]), file)

	// Estensione coerente
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == "" {
		ext = getExtensionFromContentType(detectedType)
	}
	if ext == "" {
		ext = defaultImageExt
	}

	// Crea directory
	if err := os.MkdirAll(mediaUploadDir, 0o755); err != nil {
		ctx.Logger.WithError(err).Error("cannot create upload directory")
		writeErrorJSON(w, http.StatusInternalServerError, "Failed to save file")
		return
	}

	// Nome file unico
	safeUser := strings.ReplaceAll(ctx.UserIdentifier, "/", "_")
	filename := fmt.Sprintf("media_%s_%d%s", safeUser, time.Now().UnixNano(), ext)
	filePath := filepath.Join(mediaUploadDir, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		ctx.Logger.WithError(err).Error("cannot create destination file")
		writeErrorJSON(w, http.StatusInternalServerError, "Failed to save file")
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, reader); err != nil {
		ctx.Logger.WithError(err).Error("cannot write file")
		writeErrorJSON(w, http.StatusInternalServerError, "Failed to save file")
		return
	}

	publicURL := mediaURLPrefix + filename

	writeJSON(w, http.StatusCreated, map[string]string{
		"url": publicURL,
	})
}
