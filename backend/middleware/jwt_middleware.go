package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/AndreaCasaluci/go-chat-app/utils"
)

func JWTMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the token from the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		tokenString := strings.Split(authHeader, "Bearer ")
		if len(tokenString) != 2 {
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}

		claims, err := utils.ValidateToken(tokenString[1])
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, "user_id", claims.UserID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
