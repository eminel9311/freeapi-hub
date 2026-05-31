package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/eminel9311/freeapi-hub/internal/auth"
)

// contextKey type giúp tránh collision với context keys của package khác.
type contextKey string

const (
	userIDKey contextKey = "user_id"
	emailKey  contextKey = "email"
)

// Auth là middleware xác thực JWT.
// Pattern: middleware nhận handler, trả về handler bọc thêm logic.
//
// TUẦN 4 - BUỔI 6: hiểu pattern này. Đây là pattern phổ biến nhất Go web.
func Auth(jwtManager *auth.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
				return
			}

			// Format: "Bearer <token>"
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, `{"error":"invalid authorization format"}`, http.StatusUnauthorized)
				return
			}

			claims, err := jwtManager.Verify(parts[1])
			if err != nil {
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			// Inject user info vào context để handler downstream dùng.
			ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
			ctx = context.WithValue(ctx, emailKey, claims.Email)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext giúp handler lấy userID đã được middleware inject.
func UserIDFromContext(ctx context.Context) (int64, bool) {
	v, ok := ctx.Value(userIDKey).(int64)
	return v, ok
}

func EmailFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(emailKey).(string)
	return v, ok
}
