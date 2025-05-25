package api

import (
	"net/http"

	"github.com/aelhady03/twerlo-chat-app/internal/auth"
	"github.com/aelhady03/twerlo-chat-app/internal/config"
	"github.com/aelhady03/twerlo-chat-app/internal/service"
	"github.com/aelhady03/twerlo-chat-app/internal/websocket"

	"github.com/gorilla/mux"
)

type Router struct {
	authHandler    *AuthHandler
	messageHandler *MessageHandler
	mediaHandler   *MediaHandler
	userService    *service.UserService
	jwtManager     *auth.JWTManager
	hub            *websocket.Hub
	config         *config.Config
}

func NewRouter(
	userService *service.UserService,
	messageService *service.MessageService,
	jwtManager *auth.JWTManager,
	hub *websocket.Hub,
	config *config.Config,
) *Router {
	return &Router{
		authHandler:    NewAuthHandler(userService),
		messageHandler: NewMessageHandler(messageService, hub),
		mediaHandler:   NewMediaHandler(config),
		userService:    userService,
		jwtManager:     jwtManager,
		hub:            hub,
		config:         config,
	}
}

func (r *Router) SetupRoutes() *mux.Router {
	router := mux.NewRouter()

	// Apply CORS middleware to all routes
	router.Use(corsMiddleware)

	// API routes
	api := router.PathPrefix("/api").Subrouter()

	// Public routes (no authentication required)
	api.HandleFunc("/auth/register", r.authHandler.Register).Methods("POST")
	api.HandleFunc("/auth/login", r.authHandler.Login).Methods("POST")

	// Protected routes (authentication required)
	protected := api.PathPrefix("").Subrouter()
	protected.Use(auth.AuthMiddleware(r.jwtManager))

	// Auth routes
	protected.HandleFunc("/auth/logout", r.authHandler.Logout).Methods("POST")

	// Message routes
	protected.HandleFunc("/messages/send", r.messageHandler.SendMessage).Methods("POST")
	protected.HandleFunc("/messages/broadcast", r.messageHandler.BroadcastMessage).Methods("POST")
	protected.HandleFunc("/messages/history", r.messageHandler.GetChatHistory).Methods("GET")
	protected.HandleFunc("/messages", r.messageHandler.GetUserMessages).Methods("GET")
	protected.HandleFunc("/messages/{messageId}/status", r.messageHandler.UpdateDeliveryStatus).Methods("PUT")

	// Media routes
	protected.HandleFunc("/media/upload", r.mediaHandler.UploadMedia).Methods("POST")

	// User routes
	protected.HandleFunc("/users", r.GetUsers).Methods("GET")
	protected.HandleFunc("/users/me", r.GetCurrentUser).Methods("GET")
	protected.HandleFunc("/users/online", r.GetOnlineUsers).Methods("GET")

	// WebSocket route
	router.HandleFunc("/ws", func(w http.ResponseWriter, req *http.Request) {
		websocket.ServeWS(r.hub, r.jwtManager, w, req)
	})

	// Static files
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	router.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads/"))))

	// Serve index.html for the root path
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})

	return router
}

// GetUsers returns all users
func (r *Router) GetUsers(w http.ResponseWriter, req *http.Request) {
	users, err := r.userService.GetAllUsers()
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "USERS_FAILED", "Failed to retrieve users")
		return
	}

	writeSuccessResponse(w, http.StatusOK, "Users retrieved successfully", users)
}

// GetCurrentUser returns the current authenticated user
func (r *Router) GetCurrentUser(w http.ResponseWriter, req *http.Request) {
	claims, err := getUserFromContext(req.Context())
	if err != nil {
		writeErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	user, err := r.userService.GetUserByID(claims.UserID)
	if err != nil {
		writeErrorResponse(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
		return
	}

	writeSuccessResponse(w, http.StatusOK, "User retrieved successfully", user)
}

// GetOnlineUsers returns currently online users
func (r *Router) GetOnlineUsers(w http.ResponseWriter, req *http.Request) {
	onlineUsers := r.hub.GetConnectedUsers()
	writeSuccessResponse(w, http.StatusOK, "Online users retrieved successfully", onlineUsers)
}
