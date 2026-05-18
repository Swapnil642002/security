package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"firewall-manager/internal/auth"
)

type ctxKey string

const UserIDContextKey ctxKey = "user_id"

func RequireAuth(jwtManager *auth.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := strings.TrimSpace(r.Header.Get("Authorization"))
			if header == "" || !strings.HasPrefix(header, "Bearer ") {
				http.Error(w, "missing or invalid authorization header", http.StatusUnauthorized)
				return
			}

			token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
			claims, err := jwtManager.Parse(token)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDContextKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) (int64, bool) {
	v := ctx.Value(UserIDContextKey)
	if v == nil {
		return 0, false
	}
	id, ok := v.(int64)
	if ok {
		return id, true
	}
	strID, ok := v.(string)
	if !ok {
		return 0, false
	}
	parsed, err := strconv.ParseInt(strID, 10, 64)
	if err != nil {
		return 0, false
	}
	return parsed, true
}
