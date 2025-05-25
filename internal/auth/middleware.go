package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/aelhady03/twerlo-chat-app/internal/models"
)

type contextKey string

const UserContextKey contextKey = "user"

// AuthMiddleware creates a middleware that validates JWT tokens
func AuthMiddleware(jwtManager *JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeErrorResponse(w, http.StatusUnauthorized, "MISSING_TOKEN", "Authorization header is required")
				return
			}

			// Check if header starts with "Bearer "
			if !strings.HasPrefix(authHeader, "Bearer ") {
				writeErrorResponse(w, http.StatusUnauthorized, "INVALID_TOKEN_FORMAT", "Authorization header must start with 'Bearer '")
				return
			}

			// Extract token
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == "" {
				writeErrorResponse(w, http.StatusUnauthorized, "EMPTY_TOKEN", "Token cannot be empty")
				return
			}

			// Validate token
			claims, err := jwtManager.ValidateToken(tokenString)
			if err != nil {
				writeErrorResponse(w, http.StatusUnauthorized, "INVALID_TOKEN", err.Error())
				return
			}

			// Add user info to request context
			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuthMiddleware creates a middleware that optionally validates JWT tokens
// If token is present and valid, user info is added to context
// If token is missing or invalid, request continues without user info
func OptionalAuthMiddleware(jwtManager *JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				tokenString := strings.TrimPrefix(authHeader, "Bearer ")
				if claims, err := jwtManager.ValidateToken(tokenString); err == nil {
					ctx := context.WithValue(r.Context(), UserContextKey, claims)
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// GetUserFromContext extracts user claims from request context
func GetUserFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(UserContextKey).(*Claims)
	return claims, ok
}

// RequireUser is a helper that gets user from context or returns error
func RequireUser(ctx context.Context) (*Claims, error) {
	claims, ok := GetUserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("user not found in context")
	}
	return claims, nil
}

// writeErrorResponse writes a JSON error response
func writeErrorResponse(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := models.NewErrorResponse(code, message, "")
	json.NewEncoder(w).Encode(response)
}
