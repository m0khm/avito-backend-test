package middleware

import (
	"context"
	"net/http"
	"strings"

	apperrors "room-booking-service/internal/errors"
	"room-booking-service/pkg/jwtutil"
)

type ctxKey string

const (
	ctxUserIDKey ctxKey = "user_id"
	ctxRoleKey   ctxKey = "role"
)

type AuthMiddleware struct {
	jwtSecret string
}

func NewAuthMiddleware(jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{jwtSecret: jwtSecret}
}

func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			WriteError(w, apperrors.ErrUnauthorized)
			return
		}
		token := strings.TrimPrefix(header, "Bearer ")
		claims, err := jwtutil.ParseToken(m.jwtSecret, token)
		if err != nil {
			WriteError(w, apperrors.ErrUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), ctxUserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, ctxRoleKey, claims.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if RoleFromContext(r.Context()) != role {
				WriteError(w, apperrors.New(apperrors.ErrForbidden.Code, "forbidden", http.StatusForbidden))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func UserIDFromContext(ctx context.Context) string {
	value, _ := ctx.Value(ctxUserIDKey).(string)
	return value
}

func RoleFromContext(ctx context.Context) string {
	value, _ := ctx.Value(ctxRoleKey).(string)
	return value
}
