package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/aelhady03/twerlo-chat-app/internal/auth"
	"github.com/aelhady03/twerlo-chat-app/internal/models"
	"github.com/aelhady03/twerlo-chat-app/internal/service"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin in development
		// In production, you should check the origin properly
		return true
	},
}

// Client represents a WebSocket client
type Client struct {
	ID       uuid.UUID
	UserID   uuid.UUID
	Username string
	Conn     *websocket.Conn
	Send     chan []byte
	Hub      *Hub
}

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Inbound messages from the clients
	broadcast chan []byte

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// User ID to client mapping for direct messaging
	userClients map[uuid.UUID]*Client

	// Mutex for thread-safe operations
	mutex sync.RWMutex

	// JWT manager for authentication
	jwtManager *auth.JWTManager

	// User service for database operations
	userService *service.UserService
}

// NewHub creates a new WebSocket hub
func NewHub(jwtManager *auth.JWTManager, userService *service.UserService) *Hub {
	return &Hub{
		clients:     make(map[*Client]bool),
		broadcast:   make(chan []byte),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		userClients: make(map[uuid.UUID]*Client),
		jwtManager:  jwtManager,
		userService: userService,
	}
}

// Run starts the hub and handles client registration/unregistration and message broadcasting
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.userClients[client.UserID] = client
			h.mutex.Unlock()

			log.Printf("Client %s (%s) connected", client.Username, client.UserID)

			// Update user online status in database
			if err := h.userService.UpdateOnlineStatus(client.UserID, true); err != nil {
				log.Printf("Failed to update online status for user %s: %v", client.UserID, err)
			}

			// Send user status update to all clients
			h.broadcastUserStatus(client.UserID, client.Username, true)

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				delete(h.userClients, client.UserID)
				close(client.Send)
			}
			h.mutex.Unlock()

			log.Printf("Client %s (%s) disconnected", client.Username, client.UserID)

			// Update user online status in database
			if err := h.userService.UpdateOnlineStatus(client.UserID, false); err != nil {
				log.Printf("Failed to update offline status for user %s: %v", client.UserID, err)
			}

			// Send user status update to all clients
			h.broadcastUserStatus(client.UserID, client.Username, false)

		case message := <-h.broadcast:
			h.mutex.RLock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
					delete(h.userClients, client.UserID)
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// BroadcastMessage broadcasts a message to all connected clients
func (h *Hub) BroadcastMessage(message *models.MessageResponse) {
	wsMessage := models.WebSocketMessage{
		Type:      models.WSMessageTypeNewMessage,
		Data:      message,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(wsMessage)
	if err != nil {
		log.Printf("Error marshaling broadcast message: %v", err)
		return
	}

	h.broadcast <- data
}

// SendDirectMessage sends a message to a specific user
func (h *Hub) SendDirectMessage(userID uuid.UUID, message *models.MessageResponse) {
	h.mutex.RLock()
	client, exists := h.userClients[userID]
	h.mutex.RUnlock()

	if !exists {
		log.Printf("User %s is not connected", userID)
		return
	}

	wsMessage := models.WebSocketMessage{
		Type:      models.WSMessageTypeNewMessage,
		Data:      message,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(wsMessage)
	if err != nil {
		log.Printf("Error marshaling direct message: %v", err)
		return
	}

	select {
	case client.Send <- data:
	default:
		close(client.Send)
		h.mutex.Lock()
		delete(h.clients, client)
		delete(h.userClients, client.UserID)
		h.mutex.Unlock()
	}
}

// SendToMultipleUsers sends a message to multiple specific users
func (h *Hub) SendToMultipleUsers(userIDs []uuid.UUID, message *models.MessageResponse) {
	wsMessage := models.WebSocketMessage{
		Type:      models.WSMessageTypeNewMessage,
		Data:      message,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(wsMessage)
	if err != nil {
		log.Printf("Error marshaling multi-user message: %v", err)
		return
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, userID := range userIDs {
		if client, exists := h.userClients[userID]; exists {
			select {
			case client.Send <- data:
			default:
				close(client.Send)
				delete(h.clients, client)
				delete(h.userClients, client.UserID)
			}
		}
	}
}

// broadcastUserStatus broadcasts user online/offline status to all clients
func (h *Hub) broadcastUserStatus(userID uuid.UUID, username string, isOnline bool) {
	status := models.UserStatus{
		UserID:   userID,
		IsOnline: isOnline,
		LastSeen: time.Now(),
	}

	wsMessage := models.WebSocketMessage{
		Type:      models.WSMessageTypeUserStatus,
		Data:      status,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(wsMessage)
	if err != nil {
		log.Printf("Error marshaling user status: %v", err)
		return
	}

	h.broadcast <- data
}

// GetConnectedUsers returns a list of currently connected users
func (h *Hub) GetConnectedUsers() []models.UserStatus {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	var users []models.UserStatus
	for client := range h.clients {
		users = append(users, models.UserStatus{
			UserID:   client.UserID,
			IsOnline: true,
			LastSeen: time.Now(),
		})
	}

	return users
}

// IsUserOnline checks if a user is currently connected
func (h *Hub) IsUserOnline(userID uuid.UUID) bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	_, exists := h.userClients[userID]
	return exists
}
