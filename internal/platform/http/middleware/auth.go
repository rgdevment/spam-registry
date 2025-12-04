package middleware

import (
	"net/http"
)

func APIKeyAuth(validKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientKey := r.Header.Get("X-API-Key")

			if clientKey == "" || clientKey != validKey {
				http.Error(w, "Unauthorized: Invalid or missing API Key", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
