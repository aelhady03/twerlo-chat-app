package api

import (
	"encoding/json"
	"net/http"

	"github.com/aelhady03/twerlo-chat-app/internal/models"
	"github.com/aelhady03/twerlo-chat-app/internal/service"
)

type AuthHandler struct {
	userService *service.UserService
}

func NewAuthHandler(userService *service.UserService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.UserRegistration
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Basic validation
	if req.Username == "" || req.Email == "" || req.Password == "" {
		writeErrorResponse(w, http.StatusBadRequest, "MISSING_FIELDS", "Username, email, and password are required")
		return
	}

	if len(req.Password) < 6 {
		writeErrorResponse(w, http.StatusBadRequest, "WEAK_PASSWORD", "Password must be at least 6 characters long")
		return
	}

	// Register user
	authResponse, err := h.userService.Register(&req)
	if err != nil {
		if err.Error() == "email already exists" || err.Error() == "username already exists" {
			writeErrorResponse(w, http.StatusConflict, "USER_EXISTS", err.Error())
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, "REGISTRATION_FAILED", "Failed to register user")
		return
	}

	writeSuccessResponse(w, http.StatusCreated, "User registered successfully", authResponse)
}

// Login handles user authentication
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.UserLogin
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Basic validation
	if req.Email == "" || req.Password == "" {
		writeErrorResponse(w, http.StatusBadRequest, "MISSING_FIELDS", "Email and password are required")
		return
	}

	// Authenticate user
	authResponse, err := h.userService.Login(&req)
	if err != nil {
		writeErrorResponse(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password")
		return
	}

	writeSuccessResponse(w, http.StatusOK, "Login successful", authResponse)
}

// Logout handles user logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Get user from context (set by auth middleware)
	claims, err := getUserFromContext(r.Context())
	if err != nil {
		writeErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Update user online status
	err = h.userService.Logout(claims.UserID)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "LOGOUT_FAILED", "Failed to logout user")
		return
	}

	writeSuccessResponse(w, http.StatusOK, "Logout successful", nil)
}
