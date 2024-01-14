package utils

import (
	"log/slog"
	"net/http"
)

const MimeTypeJSON = "application/json"
const MimeTypeText = "text/plain"

func WriteJSONString(w http.ResponseWriter, data string) bool {
	w.Header().Set("Content-Type", MimeTypeJSON)
	w.WriteHeader(http.StatusOK)

	_, err := w.Write([]byte(data))
	if err != nil {
		slog.Error("failed to write JSON", Tag("failed_write_json"), Error(err), slog.String("data", data))
		return false
	}
	return true
}

func GetIP(r *http.Request) string {
	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		ip = r.Header.Get("CF-Connecting-IP")
	}
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}
	return ip
}
