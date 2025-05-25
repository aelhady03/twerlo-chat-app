package service

import (
	"fmt"
	"time"

	"github.com/aelhady03/twerlo-chat-app/internal/models"
	"github.com/aelhady03/twerlo-chat-app/internal/repository"

	"github.com/google/uuid"
)

type MessageService struct {
	messageRepo *repository.MessageRepository
	userRepo    *repository.UserRepository
}

func NewMessageService(messageRepo *repository.MessageRepository, userRepo *repository.UserRepository) *MessageService {
	return &MessageService{
		messageRepo: messageRepo,
		userRepo:    userRepo,
	}
}

// SendMessage sends a direct message to a specific user
func (s *MessageService) SendMessage(senderID uuid.UUID, req *models.MessageRequest) (*models.MessageResponse, error) {
	// Validate recipient exists
	if req.RecipientID != nil {
		_, err := s.userRepo.GetByID(*req.RecipientID)
		if err != nil {
			return nil, fmt.Errorf("recipient not found: %w", err)
		}
	}

	// Get sender info
	sender, err := s.userRepo.GetByID(senderID)
	if err != nil {
		return nil, fmt.Errorf("sender not found: %w", err)
	}

	// Create message
	message := &models.Message{
		ID:             uuid.New(),
		SenderID:       senderID,
		RecipientID:    req.RecipientID,
		Content:        req.Content,
		MessageType:    req.MessageType,
		MediaURL:       req.MediaURL,
		DeliveryStatus: models.DeliveryStatusSent,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		IsBroadcast:    false,
	}

	err = s.messageRepo.Create(message)
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Return message response
	return &models.MessageResponse{
		ID:             message.ID,
		SenderID:       message.SenderID,
		SenderUsername: sender.Username,
		RecipientID:    message.RecipientID,
		Content:        message.Content,
		MessageType:    message.MessageType,
		MediaURL:       message.MediaURL,
		MediaFilename:  message.MediaFilename,
		MediaSize:      message.MediaSize,
		DeliveryStatus: message.DeliveryStatus,
		CreatedAt:      message.CreatedAt,
		IsBroadcast:    message.IsBroadcast,
	}, nil
}

// BroadcastMessage sends a message to multiple users
func (s *MessageService) BroadcastMessage(senderID uuid.UUID, req *models.MessageRequest) (*models.MessageResponse, error) {
	if len(req.RecipientIDs) == 0 {
		return nil, fmt.Errorf("no recipients specified for broadcast")
	}

	// Validate all recipients exist
	for _, recipientID := range req.RecipientIDs {
		_, err := s.userRepo.GetByID(recipientID)
		if err != nil {
			return nil, fmt.Errorf("recipient %s not found: %w", recipientID, err)
		}
	}

	// Get sender info
	sender, err := s.userRepo.GetByID(senderID)
	if err != nil {
		return nil, fmt.Errorf("sender not found: %w", err)
	}

	// Create broadcast message
	message := &models.Message{
		ID:             uuid.New(),
		SenderID:       senderID,
		RecipientID:    nil, // nil for broadcast messages
		Content:        req.Content,
		MessageType:    req.MessageType,
		MediaURL:       req.MediaURL,
		DeliveryStatus: models.DeliveryStatusSent,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		IsBroadcast:    true,
	}

	err = s.messageRepo.Create(message)
	if err != nil {
		return nil, fmt.Errorf("failed to create broadcast message: %w", err)
	}

	// Create broadcast message entries for each recipient
	for _, recipientID := range req.RecipientIDs {
		broadcastMsg := &models.BroadcastMessage{
			MessageID:   message.ID,
			RecipientID: recipientID,
			DeliveredAt: time.Now(),
			ReadAt:      nil,
		}

		err = s.messageRepo.CreateBroadcastMessage(broadcastMsg)
		if err != nil {
			// Log error but continue with other recipients
			fmt.Printf("Failed to create broadcast message for recipient %s: %v\n", recipientID, err)
		}
	}

	// Return message response
	return &models.MessageResponse{
		ID:             message.ID,
		SenderID:       message.SenderID,
		SenderUsername: sender.Username,
		RecipientID:    nil,
		Content:        message.Content,
		MessageType:    message.MessageType,
		MediaURL:       message.MediaURL,
		MediaFilename:  message.MediaFilename,
		MediaSize:      message.MediaSize,
		DeliveryStatus: message.DeliveryStatus,
		CreatedAt:      message.CreatedAt,
		IsBroadcast:    message.IsBroadcast,
	}, nil
}

// GetChatHistory retrieves chat history between two users
func (s *MessageService) GetChatHistory(userID1, userID2 uuid.UUID, page, limit int) (*models.ChatHistory, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	messages, total, err := s.messageRepo.GetChatHistory(userID1, userID2, page, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat history: %w", err)
	}

	hasMore := int64((page-1)*limit)+int64(len(messages)) < total

	return &models.ChatHistory{
		Messages: messages,
		Page:     page,
		Limit:    limit,
		Total:    total,
		HasMore:  hasMore,
	}, nil
}

// GetUserMessages retrieves all messages for a user (sent and received)
func (s *MessageService) GetUserMessages(userID uuid.UUID, page, limit int) (*models.ChatHistory, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	messages, total, err := s.messageRepo.GetUserMessages(userID, page, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get user messages: %w", err)
	}

	hasMore := int64((page-1)*limit)+int64(len(messages)) < total

	return &models.ChatHistory{
		Messages: messages,
		Page:     page,
		Limit:    limit,
		Total:    total,
		HasMore:  hasMore,
	}, nil
}

// UpdateDeliveryStatus updates the delivery status of a message
func (s *MessageService) UpdateDeliveryStatus(messageID uuid.UUID, status models.DeliveryStatus) error {
	err := s.messageRepo.UpdateDeliveryStatus(messageID, status)
	if err != nil {
		return fmt.Errorf("failed to update delivery status: %w", err)
	}

	return nil
}

// MarkBroadcastMessageAsRead marks a broadcast message as read for a specific recipient
func (s *MessageService) MarkBroadcastMessageAsRead(messageID, recipientID uuid.UUID) error {
	err := s.messageRepo.MarkBroadcastMessageAsRead(messageID, recipientID)
	if err != nil {
		return fmt.Errorf("failed to mark broadcast message as read: %w", err)
	}

	return nil
}

// GetBroadcastRecipients retrieves all recipients of a broadcast message
func (s *MessageService) GetBroadcastRecipients(messageID uuid.UUID) ([]models.BroadcastMessage, error) {
	recipients, err := s.messageRepo.GetBroadcastRecipients(messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get broadcast recipients: %w", err)
	}

	return recipients, nil
}
