package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aelhady03/twerlo-chat-app/internal/auth"
	"github.com/aelhady03/twerlo-chat-app/internal/models"
)

// writeSuccessResponse writes a successful JSON response
func writeSuccessResponse(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := models.NewSuccessResponse(message, data)
	json.NewEncoder(w).Encode(response)
}

// writeErrorResponse writes an error JSON response
func writeErrorResponse(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := models.NewErrorResponse(code, message, "")
	json.NewEncoder(w).Encode(response)
}

// getUserFromContext extracts user claims from request context
func getUserFromContext(ctx context.Context) (*auth.Claims, error) {
	claims, ok := auth.GetUserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("user not found in context")
	}
	return claims, nil
}

// enableCORS sets CORS headers for the response
func enableCORS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
}

// corsMiddleware is a middleware that handles CORS
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}
		next.ServeHTTP(w, r)
	})
}
