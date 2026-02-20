package handlers

import (
	"context"
	"net/http"
)

type contextKey string

const (
	// ContextKeyUserID is the context key for the authenticated user's ID.
	ContextKeyUserID contextKey = "userID"
	// ContextKeyUserRole is the context key for the authenticated user's role.
	ContextKeyUserRole contextKey = "userRole"
)

// AuthMiddleware extracts X-User-ID and X-User-Role headers and injects
// them into the request context. Returns 401 if either header is missing.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		userRole := r.Header.Get("X-User-Role")

		if userID == "" || userRole == "" {
			http.Error(w, `{"error": "X-User-ID and X-User-Role headers are required"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ContextKeyUserID, userID)
		ctx = context.WithValue(ctx, ContextKeyUserRole, userRole)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
