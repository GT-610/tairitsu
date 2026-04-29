package httpcode

import (
	"net/http"
	"strings"
)

func DefaultErrorCode(status int) string {
	text := strings.ToLower(http.StatusText(status))
	if text == "" {
		return "error.unknown"
	}
	return "http." + strings.ReplaceAll(text, " ", "_")
}
