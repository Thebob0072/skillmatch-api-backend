package main

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetConversations returns all conversations for the current user
// GET /conversations
func GetConversations(c *gin.Context) {
	userID := c.GetInt("userID")

	query := `
		SELECT 
			c.id,
			c.user1_id,
			c.user2_id,
			c.last_message_at,
			c.created_at,
			c.updated_at,
			COALESCE(u.username, u.email) as other_user_name,
			COALESCE(up.profile_image_url, u.google_profile_picture) as profile_photo_url,
			(SELECT COUNT(*) FROM messages m 
			 WHERE m.conversation_id = c.id 
			 AND m.receiver_id = $1 
			 AND m.is_read = FALSE) as unread_count,
			(SELECT content FROM messages m 
			 WHERE m.conversation_id = c.id 
			 ORDER BY m.created_at DESC LIMIT 1) as last_message_content,
			(SELECT created_at FROM messages m 
			 WHERE m.conversation_id = c.id 
			 ORDER BY m.created_at DESC LIMIT 1) as last_message_time
		FROM conversations c
		LEFT JOIN users u ON (
			CASE 
				WHEN c.user1_id = $1 THEN c.user2_id
				ELSE c.user1_id
			END = u.user_id
		)
		LEFT JOIN user_profiles up ON u.user_id = up.user_id
		WHERE c.user1_id = $1 OR c.user2_id = $1
		ORDER BY c.last_message_at DESC
	`

	rows, err := db.Query(query, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch conversations"})
		return
	}
	defer rows.Close()

	conversations := []Conversation{}
	totalUnread := 0

	for rows.Next() {
		var conv Conversation
		var otherUser ConversationUser
		var lastMsg Message
		var profilePhotoURL sql.NullString
		var lastMsgContent sql.NullString
		var lastMsgTime sql.NullTime

		err := rows.Scan(
			&conv.ID,
			&conv.User1ID,
			&conv.User2ID,
			&conv.LastMessageAt,
			&conv.CreatedAt,
			&conv.UpdatedAt,
			&otherUser.DisplayName,
			&profilePhotoURL,
			&conv.UnreadCount,
			&lastMsgContent,
			&lastMsgTime,
		)
		if err != nil {
			continue
		}

		// Set other user info
		if conv.User1ID == userID {
			otherUser.ID = conv.User2ID
		} else {
			otherUser.ID = conv.User1ID
		}
		if profilePhotoURL.Valid {
			otherUser.ProfilePhotoURL = profilePhotoURL.String
		}
		conv.OtherUser = &otherUser

		// Set last message
		if lastMsgContent.Valid {
			lastMsg.Content = lastMsgContent.String
			if lastMsgTime.Valid {
				lastMsg.CreatedAt = lastMsgTime.Time
			}
			conv.LastMessage = &lastMsg
		}

		totalUnread += conv.UnreadCount
		conversations = append(conversations, conv)
	}

	c.JSON(http.StatusOK, ConversationListResponse{
		Conversations: conversations,
		Total:         len(conversations),
		UnreadTotal:   totalUnread,
	})
}

// GetConversationMessages returns all messages in a specific conversation
// GET /conversations/:id/messages
func GetConversationMessages(c *gin.Context) {
	userID := c.GetInt("userID")
	conversationID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	// Verify user is part of this conversation
	var count int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM conversations 
		WHERE id = $1 AND (user1_id = $2 OR user2_id = $2)
	`, conversationID, userID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Pagination parameters
	limit := 50
	offset := 0
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Get messages
	query := `
		SELECT 
			m.id,
			m.conversation_id,
			m.sender_id,
			m.receiver_id,
			m.content,
			m.message_type,
			m.is_read,
			m.read_at,
			m.created_at,
			COALESCE(u.display_name, u.email) as sender_name,
			up.profile_photo_url
		FROM messages m
		LEFT JOIN users u ON m.sender_id = u.id
		LEFT JOIN user_profiles up ON u.id = up.user_id
		WHERE m.conversation_id = $1
		ORDER BY m.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := db.Query(query, conversationID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
		return
	}
	defer rows.Close()

	messages := []Message{}
	for rows.Next() {
		var msg Message
		var senderName string
		var profilePhotoURL sql.NullString

		err := rows.Scan(
			&msg.ID,
			&msg.ConversationID,
			&msg.SenderID,
			&msg.ReceiverID,
			&msg.Content,
			&msg.MessageType,
			&msg.IsRead,
			&msg.ReadAt,
			&msg.CreatedAt,
			&senderName,
			&profilePhotoURL,
		)
		if err != nil {
			continue
		}

		msg.SenderName = senderName
		if profilePhotoURL.Valid {
			msg.SenderPhotoURL = profilePhotoURL.String
		}

		messages = append(messages, msg)
	}

	// Get total count
	var total int
	db.QueryRow("SELECT COUNT(*) FROM messages WHERE conversation_id = $1", conversationID).Scan(&total)

	c.JSON(http.StatusOK, MessageListResponse{
		Messages:       messages,
		Total:          total,
		ConversationID: conversationID,
		HasMore:        offset+limit < total,
	})
}

// SendMessage sends a new message and creates a conversation if needed
// POST /messages
func SendMessage(c *gin.Context) {
	userID := c.GetInt("userID")

	var req CreateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Default message type
	if req.MessageType == "" {
		req.MessageType = "text"
	}

	// Validate message type
	if req.MessageType != "text" && req.MessageType != "image" && req.MessageType != "system" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message type"})
		return
	}

	// Cannot send message to yourself
	if userID == req.ReceiverID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot send message to yourself"})
		return
	}

	// Check if receiver exists
	var receiverExists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", req.ReceiverID).Scan(&receiverExists)
	if err != nil || !receiverExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Receiver not found"})
		return
	}

	// Get or create conversation
	user1ID := userID
	user2ID := req.ReceiverID
	if user1ID > user2ID {
		user1ID, user2ID = user2ID, user1ID
	}

	var conversationID int
	err = db.QueryRow(`
		INSERT INTO conversations (user1_id, user2_id)
		VALUES ($1, $2)
		ON CONFLICT (user1_id, user2_id) 
		DO UPDATE SET updated_at = CURRENT_TIMESTAMP
		RETURNING id
	`, user1ID, user2ID).Scan(&conversationID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create conversation"})
		return
	}

	// Insert message
	var message Message
	err = db.QueryRow(`
		INSERT INTO messages (conversation_id, sender_id, receiver_id, content, message_type)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, conversation_id, sender_id, receiver_id, content, message_type, is_read, created_at
	`, conversationID, userID, req.ReceiverID, req.Content, req.MessageType).Scan(
		&message.ID,
		&message.ConversationID,
		&message.SenderID,
		&message.ReceiverID,
		&message.Content,
		&message.MessageType,
		&message.IsRead,
		&message.CreatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message"})
		return
	}

	// Send via WebSocket if receiver is online
	wsManager.BroadcastToUser(req.ReceiverID, WebSocketMessage{
		Type:    "message",
		Payload: message,
	})

	// Create notification
	CreateNotification(req.ReceiverID, "new_message", "You have a new message", map[string]interface{}{
		"conversation_id": conversationID,
		"sender_id":       userID,
		"message_id":      message.ID,
	})

	c.JSON(http.StatusCreated, message)
}

// MarkMessagesAsRead marks one or more messages as read
// PATCH /messages/read
func MarkMessagesAsRead(c *gin.Context) {
	userID := c.GetInt("userID")

	var req MarkReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.MessageIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No message IDs provided"})
		return
	}

	// Build placeholders for IN clause
	query := `
		UPDATE messages
		SET is_read = TRUE, read_at = CURRENT_TIMESTAMP
		WHERE receiver_id = $1 AND id = ANY($2) AND is_read = FALSE
		RETURNING id, conversation_id
	`

	rows, err := db.Query(query, userID, req.MessageIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark messages as read"})
		return
	}
	defer rows.Close()

	updatedIDs := []int{}
	conversationIDs := map[int]bool{}
	for rows.Next() {
		var msgID, convID int
		rows.Scan(&msgID, &convID)
		updatedIDs = append(updatedIDs, msgID)
		conversationIDs[convID] = true
	}

	// Send read receipts via WebSocket
	for convID := range conversationIDs {
		var senderID int
		db.QueryRow(`
			SELECT sender_id FROM messages 
			WHERE id = ANY($1) AND conversation_id = $2 
			LIMIT 1
		`, req.MessageIDs, convID).Scan(&senderID)

		if senderID > 0 {
			wsManager.BroadcastToUser(senderID, WebSocketMessage{
				Type: "read_receipt",
				Payload: ReadReceipt{
					ConversationID: convID,
					MessageIDs:     updatedIDs,
				},
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"updated_count": len(updatedIDs),
		"message_ids":   updatedIDs,
	})
}

// DeleteMessage soft deletes a message (only sender can delete)
// DELETE /messages/:id
func DeleteMessage(c *gin.Context) {
	userID := c.GetInt("userID")
	messageID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	// Verify ownership
	var senderID int
	err = db.QueryRow("SELECT sender_id FROM messages WHERE id = $1", messageID).Scan(&senderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}

	if senderID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete your own messages"})
		return
	}

	// Update message content to indicate deletion
	_, err = db.Exec(`
		UPDATE messages 
		SET content = '[Message deleted]', message_type = 'system'
		WHERE id = $1
	`, messageID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message deleted"})
}
