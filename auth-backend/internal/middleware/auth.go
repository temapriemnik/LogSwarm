package middleware

import (
	"context"
	"net/http"
	"strings"

	"authbackend/internal/domain"
	"authbackend/internal/usecase"
)

func Auth(authService usecase.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "authorization required", http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, "invalid authorization header", http.StatusUnauthorized)
				return
			}

			user, err := authService.ValidateToken(r.Context(), parts[1])
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			ctx := r.Context()
			ctx = contextWithUser(ctx, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type contextKey string

const userKey contextKey = "user"

func contextWithUser(ctx context.Context, user *domain.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func UserFromContext(ctx context.Context) (*domain.User, bool) {
	user, ok := ctx.Value(userKey).(*domain.User)
	return user, ok
}

func Protected(handler http.Handler, authService usecase.AuthService) http.Handler {
	return Auth(authService)(handler)
}