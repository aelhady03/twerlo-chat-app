package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/aelhady03/twerlo-chat-app/internal/auth"
	"github.com/aelhady03/twerlo-chat-app/internal/models"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

// ServeWS handles websocket requests from the peer
func ServeWS(hub *Hub, jwtManager *auth.JWTManager, w http.ResponseWriter, r *http.Request) {
	// Get token from query parameter
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		return
	}

	// Validate token
	claims, err := jwtManager.ValidateToken(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Create client
	client := &Client{
		ID:       uuid.New(),
		UserID:   claims.UserID,
		Username: claims.Username,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Hub:      hub,
	}

	// Register client with hub
	client.Hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in new goroutines
	go client.writePump()
	go client.readPump()
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle incoming message
		c.handleMessage(message)
	}
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming WebSocket messages from the client
func (c *Client) handleMessage(data []byte) {
	var wsMessage models.WebSocketMessage
	if err := json.Unmarshal(data, &wsMessage); err != nil {
		log.Printf("Error unmarshaling WebSocket message: %v", err)
		c.sendError("Invalid message format")
		return
	}

	switch wsMessage.Type {
	case models.WSMessageTypePing:
		c.sendPong()
	case models.WSMessageTypeDeliveryUpdate:
		c.handleDeliveryUpdate(wsMessage.Data)
	default:
		log.Printf("Unknown message type: %s", wsMessage.Type)
		c.sendError("Unknown message type")
	}
}

// sendPong sends a pong response to the client
func (c *Client) sendPong() {
	pongMessage := models.WebSocketMessage{
		Type:      models.WSMessageTypePong,
		Data:      nil,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(pongMessage)
	if err != nil {
		log.Printf("Error marshaling pong message: %v", err)
		return
	}

	select {
	case c.Send <- data:
	default:
		close(c.Send)
	}
}

// sendError sends an error message to the client
func (c *Client) sendError(message string) {
	errorMessage := models.WebSocketMessage{
		Type: models.WSMessageTypeError,
		Data: map[string]string{
			"error": message,
		},
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(errorMessage)
	if err != nil {
		log.Printf("Error marshaling error message: %v", err)
		return
	}

	select {
	case c.Send <- data:
	default:
		close(c.Send)
	}
}

// handleDeliveryUpdate processes delivery status updates from the client
func (c *Client) handleDeliveryUpdate(data interface{}) {
	// Convert data to delivery update
	dataBytes, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling delivery update data: %v", err)
		return
	}

	var deliveryUpdate models.MessageDeliveryUpdate
	if err := json.Unmarshal(dataBytes, &deliveryUpdate); err != nil {
		log.Printf("Error unmarshaling delivery update: %v", err)
		return
	}

	// Here you would typically update the message delivery status in the database
	// For now, we'll just log it
	log.Printf("Delivery update from %s: Message %s status %s",
		c.Username, deliveryUpdate.MessageID, deliveryUpdate.Status)
}
