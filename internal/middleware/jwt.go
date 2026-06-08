package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/eminel9311/freeapi-hub/internal/auth"
	"github.com/eminel9311/freeapi-hub/internal/httputil"
)

// Key dùng để lưu user info vào request context.
// Phải dùng type riêng (không phải string) để tránh collision với key của package khác.
type ctxKey string

const userClaimsKey ctxKey = "userClaims"

// JWTAuth tạo middleware verify JWT.
func JWTAuth(jwt *auth.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Lấy header "Authorization: Bearer <token>"
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				httputil.Error(w, http.StatusUnauthorized, "missing Authorization header")
				return
			}

			// 2. Tách prefix "Bearer "
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				httputil.Error(w, http.StatusUnauthorized, "invalid Authorization format")
				return
			}
			token := parts[1]

			// 3. Verify token
			claims, err := jwt.Verify(token)
			if err != nil {
				httputil.Error(w, http.StatusUnauthorized, "invalid token: "+err.Error())
				return
			}

			// 4. Gắn claims vào context để handler dùng được
			ctx := context.WithValue(r.Context(), userClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ClaimsFromContext lấy claims từ request context (handler dùng).
func ClaimsFromContext(ctx context.Context) (*auth.Claims, bool) {
	claims, ok := ctx.Value(userClaimsKey).(*auth.Claims)
	return claims, ok
}
