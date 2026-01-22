package api

import "strings"

func isValidMediaType(contentType string) bool {
	ct := strings.ToLower(contentType)

	validPrefixes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
		"image/webp",
	}

	for _, p := range validPrefixes {
		if strings.HasPrefix(ct, p) {
			return true
		}
	}
	return false
}

func getExtensionFromContentType(contentType string) string {
	ct := strings.ToLower(contentType)

	switch {
	case strings.Contains(ct, "jpeg"), strings.Contains(ct, "jpg"):
		return ".jpg"
	case strings.Contains(ct, "png"):
		return ".png"
	case strings.Contains(ct, "gif"):
		return ".gif"
	case strings.Contains(ct, "webp"):
		return ".webp"
	default:
		return ".jpg"
	}
}
