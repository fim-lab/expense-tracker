package middleware

import (
	"context"
	"net/http"

	"github.com/fim-lab/expense-tracker/internal/core/ports"
	authutils "github.com/fim-lab/expense-tracker/pkg/auth"
)

type AuthMiddleware struct {
	service ports.SessionService
}

func NewAuthMiddleware(service *ports.SessionService) *AuthMiddleware {
	return &AuthMiddleware{
		service: *service,
	}
}

func (am *AuthMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionToken, err := r.Cookie("session_token")
		if err != nil || sessionToken == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		hashedToken := authutils.HashSessionToken(sessionToken.Value)
		valid, userID := am.service.ValidateSession(hashedToken)
		if !valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "userID", userID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
