package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
)

// uploadMedia gestisce POST /media
// Carica un file media (immagine/GIF) e ritorna un URL pubblico
func (rt *_router) uploadMedia(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Ottieni user ID dal contesto (inserito dal middleware auth)
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		writeErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse multipart form (max 10MB)
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "File too large or invalid")
		return
	}

	// Ottieni il file
	file, header, err := r.FormFile("file")
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "Missing file")
		return
	}
	defer file.Close()

	// Valida il tipo di file
	contentType := header.Header.Get("Content-Type")
	if !isValidMediaType(contentType) {
		writeErrorJSON(w, http.StatusBadRequest, "Invalid file type. Only images and GIFs allowed")
		return
	}

	// Genera nome file unico
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = getExtensionFromContentType(contentType)
	}
	filename := fmt.Sprintf("media_%d_%d%s", userID, time.Now().UnixNano(), ext)

	// Percorso dove salvare il file
	uploadDir := "./uploads/media"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		rt.baseLogger.WithError(err).Error("Failed to create upload directory")
		writeErrorJSON(w, http.StatusInternalServerError, "Failed to save file")
		return
	}

	filePath := filepath.Join(uploadDir, filename)

	// Salva il file
	dst, err := os.Create(filePath)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Failed to create file")
		writeErrorJSON(w, http.StatusInternalServerError, "Failed to save file")
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		rt.baseLogger.WithError(err).Error("Failed to write file")
		writeErrorJSON(w, http.StatusInternalServerError, "Failed to save file")
		return
	}

	// Ritorna URL pubblico
	// In produzione, useresti un URL completo come https://your-domain.com/uploads/media/...
	publicURL := fmt.Sprintf("/uploads/media/%s", filename)

	writeJSON(w, http.StatusCreated, map[string]string{
		"url": publicURL,
	})
}

// setMyPhoto gestisce PUT /users/photo
// Carica la foto profilo dell'utente corrente
func (rt *_router) setMyPhoto(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Ottieni user ID dal contesto
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		writeErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse multipart form (max 10MB)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "File too large or invalid")
		return
	}

	// Ottieni il file
	file, header, err := r.FormFile("file")
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "Missing file")
		return
	}
	defer file.Close()

	// Valida il tipo di file (solo immagini per profilo)
	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		writeErrorJSON(w, http.StatusBadRequest, "Invalid file type. Only images allowed")
		return
	}

	// Genera nome file unico
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = getExtensionFromContentType(contentType)
	}
	filename := fmt.Sprintf("profile_%d%s", userID, ext)

	// Percorso dove salvare il file
	uploadDir := "./uploads/profiles"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		rt.baseLogger.WithError(err).Error("Failed to create upload directory")
		writeErrorJSON(w, http.StatusInternalServerError, "Failed to save file")
		return
	}

	filePath := filepath.Join(uploadDir, filename)

	// Rimuovi vecchia foto se esiste
	_ = os.Remove(filePath)

	// Salva il file
	dst, err := os.Create(filePath)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Failed to create file")
		writeErrorJSON(w, http.StatusInternalServerError, "Failed to save file")
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		rt.baseLogger.WithError(err).Error("Failed to write file")
		writeErrorJSON(w, http.StatusInternalServerError, "Failed to save file")
		return
	}

	// Aggiorna il database con il percorso della foto
	photoURL := fmt.Sprintf("/uploads/profiles/%s", filename)
	err = rt.db.UpdateUserPhoto(userID, photoURL)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Failed to update user photo in database")
		writeErrorJSON(w, http.StatusInternalServerError, "Failed to update photo")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Photo updated",
	})
}

// Helper functions

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
