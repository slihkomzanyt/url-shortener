package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"url-shortener/internal/auth"
)

type ctxKey string

const CtxUser ctxKey = "user"

func JWT(tm *auth.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if h == "" || !strings.HasPrefix(h, "Bearer ") {
				http.Error(w, "missing bearer token", http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(h, "Bearer ")
			tok, claims, err := tm.Parse(tokenStr)
			if err != nil || !tok.Valid {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			sub, _ := claims["sub"].(string)
			if sub == "" {
				http.Error(w, "invalid token claims", http.StatusUnauthorized)
				return
			}

			// Доп. проверка алгоритма (минимальная защита)
			if tok.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				http.Error(w, "unexpected signing method", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), CtxUser, sub)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
