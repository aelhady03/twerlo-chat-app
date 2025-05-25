package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/aelhady03/twerlo-chat-app/internal/models"
	"github.com/aelhady03/twerlo-chat-app/internal/service"
	"github.com/aelhady03/twerlo-chat-app/internal/websocket"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type MessageHandler struct {
	messageService *service.MessageService
	hub            *websocket.Hub
}

func NewMessageHandler(messageService *service.MessageService, hub *websocket.Hub) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
		hub:            hub,
	}
}

// SendMessage handles sending a direct message
func (h *MessageHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	claims, err := getUserFromContext(r.Context())
	if err != nil {
		writeErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	var req models.MessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Validate request
	if req.Content == "" {
		writeErrorResponse(w, http.StatusBadRequest, "MISSING_CONTENT", "Message content is required")
		return
	}

	if req.RecipientID == nil {
		writeErrorResponse(w, http.StatusBadRequest, "MISSING_RECIPIENT", "Recipient ID is required")
		return
	}

	// Send message
	message, err := h.messageService.SendMessage(claims.UserID, &req)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "SEND_FAILED", "Failed to send message")
		return
	}

	// Send real-time notification to recipient
	if req.RecipientID != nil {
		h.hub.SendDirectMessage(*req.RecipientID, message)
	}

	writeSuccessResponse(w, http.StatusCreated, "Message sent successfully", message)
}

// BroadcastMessage handles sending a broadcast message to multiple users
func (h *MessageHandler) BroadcastMessage(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	claims, err := getUserFromContext(r.Context())
	if err != nil {
		writeErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	var req models.MessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Validate request
	if req.Content == "" {
		writeErrorResponse(w, http.StatusBadRequest, "MISSING_CONTENT", "Message content is required")
		return
	}

	if len(req.RecipientIDs) == 0 {
		writeErrorResponse(w, http.StatusBadRequest, "MISSING_RECIPIENTS", "At least one recipient is required")
		return
	}

	// Send broadcast message
	message, err := h.messageService.BroadcastMessage(claims.UserID, &req)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "BROADCAST_FAILED", "Failed to broadcast message")
		return
	}

	// Send real-time notification to selected recipients
	h.hub.SendToMultipleUsers(req.RecipientIDs, message)

	// Also send to sender for immediate feedback
	h.hub.SendDirectMessage(claims.UserID, message)

	writeSuccessResponse(w, http.StatusCreated, "Message broadcasted successfully", message)
}

// GetChatHistory retrieves chat history between the authenticated user and another user
func (h *MessageHandler) GetChatHistory(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	claims, err := getUserFromContext(r.Context())
	if err != nil {
		writeErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Get other user ID from query parameters
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		writeErrorResponse(w, http.StatusBadRequest, "MISSING_USER_ID", "User ID is required")
		return
	}

	otherUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID format")
		return
	}

	// Get pagination parameters
	page := 1
	limit := 50

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Get chat history
	history, err := h.messageService.GetChatHistory(claims.UserID, otherUserID, page, limit)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "HISTORY_FAILED", "Failed to retrieve chat history")
		return
	}

	writeSuccessResponse(w, http.StatusOK, "Chat history retrieved successfully", history)
}

// GetUserMessages retrieves all messages for the authenticated user
func (h *MessageHandler) GetUserMessages(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	claims, err := getUserFromContext(r.Context())
	if err != nil {
		writeErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Get pagination parameters
	page := 1
	limit := 50

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Get user messages
	messages, err := h.messageService.GetUserMessages(claims.UserID, page, limit)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "MESSAGES_FAILED", "Failed to retrieve messages")
		return
	}

	writeSuccessResponse(w, http.StatusOK, "Messages retrieved successfully", messages)
}

// UpdateDeliveryStatus updates the delivery status of a message
func (h *MessageHandler) UpdateDeliveryStatus(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	_, err := getUserFromContext(r.Context())
	if err != nil {
		writeErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Get message ID from URL
	vars := mux.Vars(r)
	messageIDStr := vars["messageId"]
	messageID, err := uuid.Parse(messageIDStr)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_MESSAGE_ID", "Invalid message ID format")
		return
	}

	var req struct {
		Status models.DeliveryStatus `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Update delivery status
	err = h.messageService.UpdateDeliveryStatus(messageID, req.Status)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update delivery status")
		return
	}

	writeSuccessResponse(w, http.StatusOK, "Delivery status updated successfully", nil)
}
