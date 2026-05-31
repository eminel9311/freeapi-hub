package httputil

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// JSON helper viết JSON response chuẩn.
// Dùng ở mọi handler để response format đồng nhất.
func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("encode response", "error", err)
	}
}

// ErrorResponse format chuẩn cho lỗi.
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details any    `json:"details,omitempty"`
}

// Error viết error response.
func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, ErrorResponse{Error: message})
}
