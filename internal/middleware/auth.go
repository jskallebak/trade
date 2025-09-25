package middleware

import (
	"context"
	"net/http"
	"strings"
	"trade/internal/auth"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for login routes
		if r.URL.Path == "/login" ||
			r.URL.Path == "/api/login" ||
			r.URL.Path == "/logout" ||
			r.URL.Path == "/api/logout" ||
			r.URL.Path == "/api/webhook" ||
			r.URL.Path == "api/bbb" ||
			strings.HasPrefix(r.URL.Path, "/test/") ||
			strings.HasPrefix(r.URL.Path, "/static/") {
			next.ServeHTTP(w, r)
			return
		}

		var token string
		isAPIRoute := strings.HasPrefix(r.URL.Path, "/api/")

		// Try to get token from Authorization header first (for API calls)
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			if strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimPrefix(authHeader, "Bearer ")
			}
		} else {
			// Try to get token from cookie (for web page navigation)
			cookie, err := r.Cookie("auth_token")
			if err == nil {
				token = cookie.Value
			}
		}

		if token == "" {
			if isAPIRoute {
				http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			} else {
				// Redirect ALL web pages to login
				http.Redirect(w, r, "/login", http.StatusSeeOther)
			}
			return
		}

		userID, _, err := auth.ValidateJWT(token)
		if err != nil {
			if isAPIRoute {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
			} else {
				// Clear invalid cookie and redirect to login
				http.SetCookie(w, &http.Cookie{
					Name:   "auth_token",
					Value:  "",
					MaxAge: -1,
					Path:   "/",
				})
				http.Redirect(w, r, "/login", http.StatusSeeOther)
			}
			return
		}

		ctx := context.WithValue(r.Context(), "userID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
