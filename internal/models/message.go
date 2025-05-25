package models

import (
	"time"

	"github.com/google/uuid"
)

type MessageType string

const (
	MessageTypeText  MessageType = "text"
	MessageTypeImage MessageType = "image"
	MessageTypeVideo MessageType = "video"
	MessageTypeFile  MessageType = "file"
)

type DeliveryStatus string

const (
	DeliveryStatusSent      DeliveryStatus = "sent"
	DeliveryStatusDelivered DeliveryStatus = "delivered"
	DeliveryStatusRead      DeliveryStatus = "read"
)

type Message struct {
	ID             uuid.UUID      `json:"id" db:"id"`
	SenderID       uuid.UUID      `json:"sender_id" db:"sender_id"`
	RecipientID    *uuid.UUID     `json:"recipient_id,omitempty" db:"recipient_id"` // nil for broadcast messages
	Content        string         `json:"content" db:"content"`
	MessageType    MessageType    `json:"message_type" db:"message_type"`
	MediaURL       *string        `json:"media_url,omitempty" db:"media_url"`
	MediaFilename  *string        `json:"media_filename,omitempty" db:"media_filename"`
	MediaSize      *int64         `json:"media_size,omitempty" db:"media_size"`
	DeliveryStatus DeliveryStatus `json:"delivery_status" db:"delivery_status"`
	CreatedAt      time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at" db:"updated_at"`
	IsBroadcast    bool           `json:"is_broadcast" db:"is_broadcast"`
}

type MessageRequest struct {
	RecipientID  *uuid.UUID  `json:"recipient_id,omitempty"`
	RecipientIDs []uuid.UUID `json:"recipient_ids,omitempty"` // For broadcast messages
	Content      string      `json:"content" validate:"required"`
	MessageType  MessageType `json:"message_type" validate:"required"`
	MediaURL     *string     `json:"media_url,omitempty"`
}

type MessageResponse struct {
	ID             uuid.UUID      `json:"id"`
	SenderID       uuid.UUID      `json:"sender_id"`
	SenderUsername string         `json:"sender_username"`
	RecipientID    *uuid.UUID     `json:"recipient_id,omitempty"`
	Content        string         `json:"content"`
	MessageType    MessageType    `json:"message_type"`
	MediaURL       *string        `json:"media_url,omitempty"`
	MediaFilename  *string        `json:"media_filename,omitempty"`
	MediaSize      *int64         `json:"media_size,omitempty"`
	DeliveryStatus DeliveryStatus `json:"delivery_status"`
	CreatedAt      time.Time      `json:"created_at"`
	IsBroadcast    bool           `json:"is_broadcast"`
}

type BroadcastMessage struct {
	MessageID   uuid.UUID  `json:"message_id" db:"message_id"`
	RecipientID uuid.UUID  `json:"recipient_id" db:"recipient_id"`
	DeliveredAt time.Time  `json:"delivered_at" db:"delivered_at"`
	ReadAt      *time.Time `json:"read_at,omitempty" db:"read_at"`
}

type ChatHistory struct {
	Messages []MessageResponse `json:"messages"`
	Page     int               `json:"page"`
	Limit    int               `json:"limit"`
	Total    int64             `json:"total"`
	HasMore  bool              `json:"has_more"`
}

type MessageDeliveryUpdate struct {
	MessageID uuid.UUID      `json:"message_id"`
	Status    DeliveryStatus `json:"status"`
	UpdatedAt time.Time      `json:"updated_at"`
}
