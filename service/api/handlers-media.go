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

// uploadMedia gestisce POST /media (AUTENTICATO)
// Carica un file media (immagine/GIF) e ritorna un URL pubblico
func (rt *_router) uploadMedia(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	// 1) Auth: serve utente loggato
	if ctx.UserIdentifier == "" {
		writeErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// 2) Parse multipart form (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "File too large or invalid")
		return
	}

	// 3) Ottieni il file
	file, header, err := r.FormFile("file")
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "Missing file")
		return
	}
	defer file.Close()

	// 4) Sniff content-type dai primi bytes (più affidabile del header)
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	detectedType := http.DetectContentType(buf[:n])

	if !isValidMediaType(detectedType) {
		writeErrorJSON(w, http.StatusBadRequest, "Invalid file type. Only images and GIFs allowed")
		return
	}

	// Rimetti i bytes letti davanti allo stream originale
	reader := io.MultiReader(bytes.NewReader(buf[:n]), file)

	// 5) Estensione coerente col tipo rilevato
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = getExtensionFromContentType(detectedType)
	} else {
		ext = strings.ToLower(ext)
	}

	// 6) Nome file unico (uso UserIdentifier così è coerente col progetto)
	filename := fmt.Sprintf("media_%s_%d%s", ctx.UserIdentifier, time.Now().UnixNano(), ext)

	// 7) Percorso dove salvare il file
	uploadDir := "./uploads/media"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		rt.baseLogger.WithError(err).Error("Failed to create upload directory")
		writeErrorJSON(w, http.StatusInternalServerError, "Failed to save file")
		return
	}

	filePath := filepath.Join(uploadDir, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Failed to create file")
		writeErrorJSON(w, http.StatusInternalServerError, "Failed to save file")
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, reader); err != nil {
		rt.baseLogger.WithError(err).Error("Failed to write file")
		writeErrorJSON(w, http.StatusInternalServerError, "Failed to save file")
		return
	}

	// 8) Ritorna URL pubblico *assoluto* (coerente con OpenAPI format: url)
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	publicURL := fmt.Sprintf("%s://%s/uploads/media/%s", scheme, r.Host, filename)

	writeJSON(w, http.StatusCreated, map[string]string{
		"url": publicURL,
	})
}

// isValidMediaType controlla se il content type è valido per media messages
func isValidMediaType(contentType string) bool {
	validTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
		"image/webp",
	}
	for _, vt := range validTypes {
		if strings.HasPrefix(contentType, vt) {
			return true
		}
	}
	return false
}

// getExtensionFromContentType ritorna l'estensione file dal content type
func getExtensionFromContentType(contentType string) string {
	switch {
	case strings.Contains(contentType, "jpeg"), strings.Contains(contentType, "jpg"):
		return ".jpg"
	case strings.Contains(contentType, "png"):
		return ".png"
	case strings.Contains(contentType, "gif"):
		return ".gif"
	case strings.Contains(contentType, "webp"):
		return ".webp"
	default:
		return ".jpg"
	}
}
