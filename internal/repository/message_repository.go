package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/aelhady03/twerlo-chat-app/internal/database"
	"github.com/aelhady03/twerlo-chat-app/internal/models"

	"github.com/google/uuid"
)

type MessageRepository struct {
	db *database.DB
}

func NewMessageRepository(db *database.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

// Create creates a new message in the database
func (r *MessageRepository) Create(message *models.Message) error {
	query := `
		INSERT INTO messages (id, sender_id, recipient_id, content, message_type, media_url, media_filename, media_size, delivery_status, created_at, updated_at, is_broadcast)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.db.Exec(query,
		message.ID,
		message.SenderID,
		message.RecipientID,
		message.Content,
		message.MessageType,
		message.MediaURL,
		message.MediaFilename,
		message.MediaSize,
		message.DeliveryStatus,
		message.CreatedAt,
		message.UpdatedAt,
		message.IsBroadcast,
	)

	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	return nil
}

// GetByID retrieves a message by its ID
func (r *MessageRepository) GetByID(id uuid.UUID) (*models.Message, error) {
	query := `
		SELECT id, sender_id, recipient_id, content, message_type, media_url, media_filename, media_size, delivery_status, created_at, updated_at, is_broadcast
		FROM messages WHERE id = $1
	`

	message := &models.Message{}
	err := r.db.QueryRow(query, id).Scan(
		&message.ID,
		&message.SenderID,
		&message.RecipientID,
		&message.Content,
		&message.MessageType,
		&message.MediaURL,
		&message.MediaFilename,
		&message.MediaSize,
		&message.DeliveryStatus,
		&message.CreatedAt,
		&message.UpdatedAt,
		&message.IsBroadcast,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("message not found")
		}
		return nil, fmt.Errorf("failed to get message by ID: %w", err)
	}

	return message, nil
}

// GetChatHistory retrieves chat history between two users with pagination
func (r *MessageRepository) GetChatHistory(userID1, userID2 uuid.UUID, page, limit int) ([]models.MessageResponse, int64, error) {
	offset := (page - 1) * limit

	// Get total count
	countQuery := `
		SELECT COUNT(*) FROM messages m
		JOIN users u ON m.sender_id = u.id
		WHERE ((m.sender_id = $1 AND m.recipient_id = $2) OR (m.sender_id = $2 AND m.recipient_id = $1))
		AND m.is_broadcast = false
	`

	var total int64
	err := r.db.QueryRow(countQuery, userID1, userID2).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get message count: %w", err)
	}

	// Get messages
	query := `
		SELECT m.id, m.sender_id, u.username, m.recipient_id, m.content, m.message_type, 
		       m.media_url, m.media_filename, m.media_size, m.delivery_status, m.created_at, m.is_broadcast
		FROM messages m
		JOIN users u ON m.sender_id = u.id
		WHERE ((m.sender_id = $1 AND m.recipient_id = $2) OR (m.sender_id = $2 AND m.recipient_id = $1))
		AND m.is_broadcast = false
		ORDER BY m.created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.Query(query, userID1, userID2, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get chat history: %w", err)
	}
	defer rows.Close()

	var messages []models.MessageResponse
	for rows.Next() {
		var msg models.MessageResponse
		err := rows.Scan(
			&msg.ID,
			&msg.SenderID,
			&msg.SenderUsername,
			&msg.RecipientID,
			&msg.Content,
			&msg.MessageType,
			&msg.MediaURL,
			&msg.MediaFilename,
			&msg.MediaSize,
			&msg.DeliveryStatus,
			&msg.CreatedAt,
			&msg.IsBroadcast,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, msg)
	}

	return messages, total, nil
}

// GetUserMessages retrieves all messages for a user (sent and received) with pagination
func (r *MessageRepository) GetUserMessages(userID uuid.UUID, page, limit int) ([]models.MessageResponse, int64, error) {
	offset := (page - 1) * limit

	// Get total count
	countQuery := `
		SELECT COUNT(*) FROM messages m
		WHERE (m.sender_id = $1 OR m.recipient_id = $1 OR 
		       (m.is_broadcast = true AND EXISTS(SELECT 1 FROM broadcast_messages bm WHERE bm.message_id = m.id AND bm.recipient_id = $1)))
	`

	var total int64
	err := r.db.QueryRow(countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get message count: %w", err)
	}

	// Get messages
	query := `
		SELECT DISTINCT m.id, m.sender_id, u.username, m.recipient_id, m.content, m.message_type, 
		       m.media_url, m.media_filename, m.media_size, m.delivery_status, m.created_at, m.is_broadcast
		FROM messages m
		JOIN users u ON m.sender_id = u.id
		LEFT JOIN broadcast_messages bm ON m.id = bm.message_id
		WHERE (m.sender_id = $1 OR m.recipient_id = $1 OR (m.is_broadcast = true AND bm.recipient_id = $1))
		ORDER BY m.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user messages: %w", err)
	}
	defer rows.Close()

	var messages []models.MessageResponse
	for rows.Next() {
		var msg models.MessageResponse
		err := rows.Scan(
			&msg.ID,
			&msg.SenderID,
			&msg.SenderUsername,
			&msg.RecipientID,
			&msg.Content,
			&msg.MessageType,
			&msg.MediaURL,
			&msg.MediaFilename,
			&msg.MediaSize,
			&msg.DeliveryStatus,
			&msg.CreatedAt,
			&msg.IsBroadcast,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, msg)
	}

	return messages, total, nil
}

// UpdateDeliveryStatus updates the delivery status of a message
func (r *MessageRepository) UpdateDeliveryStatus(messageID uuid.UUID, status models.DeliveryStatus) error {
	query := `
		UPDATE messages 
		SET delivery_status = $1, updated_at = $2
		WHERE id = $3
	`

	now := time.Now()
	_, err := r.db.Exec(query, status, now, messageID)
	if err != nil {
		return fmt.Errorf("failed to update message delivery status: %w", err)
	}

	return nil
}

// CreateBroadcastMessage creates a broadcast message entry
func (r *MessageRepository) CreateBroadcastMessage(broadcastMsg *models.BroadcastMessage) error {
	query := `
		INSERT INTO broadcast_messages (message_id, recipient_id, delivered_at, read_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.Exec(query,
		broadcastMsg.MessageID,
		broadcastMsg.RecipientID,
		broadcastMsg.DeliveredAt,
		broadcastMsg.ReadAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create broadcast message: %w", err)
	}

	return nil
}

// GetBroadcastRecipients retrieves all recipients of a broadcast message
func (r *MessageRepository) GetBroadcastRecipients(messageID uuid.UUID) ([]models.BroadcastMessage, error) {
	query := `
		SELECT message_id, recipient_id, delivered_at, read_at
		FROM broadcast_messages
		WHERE message_id = $1
	`

	rows, err := r.db.Query(query, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get broadcast recipients: %w", err)
	}
	defer rows.Close()

	var recipients []models.BroadcastMessage
	for rows.Next() {
		var recipient models.BroadcastMessage
		err := rows.Scan(
			&recipient.MessageID,
			&recipient.RecipientID,
			&recipient.DeliveredAt,
			&recipient.ReadAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan broadcast recipient: %w", err)
		}
		recipients = append(recipients, recipient)
	}

	return recipients, nil
}

// MarkBroadcastMessageAsRead marks a broadcast message as read for a specific recipient
func (r *MessageRepository) MarkBroadcastMessageAsRead(messageID, recipientID uuid.UUID) error {
	query := `
		UPDATE broadcast_messages 
		SET read_at = $1
		WHERE message_id = $2 AND recipient_id = $3
	`

	_, err := r.db.Exec(query, time.Now(), messageID, recipientID)
	if err != nil {
		return fmt.Errorf("failed to mark broadcast message as read: %w", err)
	}

	return nil
}
