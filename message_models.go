package main

import (
	"time"
)

// Conversation represents a chat conversation between two users
type Conversation struct {
	ID            int       `json:"id"`
	User1ID       int       `json:"user1_id"`
	User2ID       int       `json:"user2_id"`
	LastMessageAt time.Time `json:"last_message_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// Additional fields for API responses
	OtherUser   *ConversationUser `json:"other_user,omitempty"`
	LastMessage *Message          `json:"last_message,omitempty"`
	UnreadCount int               `json:"unread_count,omitempty"`
}

// ConversationUser contains basic user info for conversation lists
type ConversationUser struct {
	ID              int    `json:"id"`
	DisplayName     string `json:"display_name"`
	ProfilePhotoURL string `json:"profile_photo_url"`
	IsOnline        bool   `json:"is_online"`
}

// Message represents a single message in a conversation
type Message struct {
	ID             int        `json:"id"`
	ConversationID int        `json:"conversation_id"`
	SenderID       int        `json:"sender_id"`
	ReceiverID     int        `json:"receiver_id"`
	Content        string     `json:"content"`
	MessageType    string     `json:"message_type"` // "text", "image", "system"
	IsRead         bool       `json:"is_read"`
	ReadAt         *time.Time `json:"read_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`

	// Additional fields for API responses
	SenderName     string `json:"sender_name,omitempty"`
	SenderPhotoURL string `json:"sender_photo_url,omitempty"`
}

// CreateMessageRequest is the payload for sending a new message
type CreateMessageRequest struct {
	ReceiverID  int    `json:"receiver_id" binding:"required"`
	Content     string `json:"content" binding:"required,min=1,max=5000"`
	MessageType string `json:"message_type"` // Optional, defaults to "text"
}

// MarkReadRequest is the payload for marking messages as read
type MarkReadRequest struct {
	MessageIDs []int `json:"message_ids" binding:"required,min=1"`
}

// ConversationListResponse wraps the conversation list with metadata
type ConversationListResponse struct {
	Conversations []Conversation `json:"conversations"`
	Total         int            `json:"total"`
	UnreadTotal   int            `json:"unread_total"`
}

// MessageListResponse wraps the message list with pagination
type MessageListResponse struct {
	Messages       []Message `json:"messages"`
	Total          int       `json:"total"`
	ConversationID int       `json:"conversation_id"`
	HasMore        bool      `json:"has_more"`
}

// WebSocketMessage represents a message sent over WebSocket
type WebSocketMessage struct {
	Type    string      `json:"type"` // "message", "read_receipt", "typing", "error"
	Payload interface{} `json:"payload"`
}

// TypingIndicator represents a typing notification
type TypingIndicator struct {
	ConversationID int  `json:"conversation_id"`
	UserID         int  `json:"user_id"`
	IsTyping       bool `json:"is_typing"`
}

// ReadReceipt represents a read receipt notification
type ReadReceipt struct {
	ConversationID int       `json:"conversation_id"`
	MessageIDs     []int     `json:"message_ids"`
	ReadAt         time.Time `json:"read_at"`
}
