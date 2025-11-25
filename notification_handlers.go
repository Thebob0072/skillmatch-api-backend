package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Notification represents a user notification
type Notification struct {
	ID        int                    `json:"id"`
	UserID    int                    `json:"user_id"`
	Type      string                 `json:"type"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	IsRead    bool                   `json:"is_read"`
	ReadAt    *time.Time             `json:"read_at,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

// NotificationListResponse wraps the notification list with metadata
type NotificationListResponse struct {
	Notifications []Notification `json:"notifications"`
	Total         int            `json:"total"`
	UnreadCount   int            `json:"unread_count"`
}

// CreateNotification creates a new notification for a user
// This is an internal function, not an HTTP handler
func CreateNotification(userID int, notifType, message string, metadata map[string]interface{}) error {
	// Generate title based on type
	title := getNotificationTitle(notifType)

	// Convert metadata to JSONB
	var metadataJSON interface{}
	if metadata != nil {
		metadataJSON = metadata
	}

	var notifID int
	err := db.QueryRow(`
		INSERT INTO notifications (user_id, type, title, message, metadata)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, userID, notifType, title, message, metadataJSON).Scan(&notifID)

	if err != nil {
		return err
	}

	// Send real-time notification via WebSocket
	if wsManager != nil {
		wsManager.BroadcastToUser(userID, WebSocketMessage{
			Type: "notification",
			Payload: map[string]interface{}{
				"id":       notifID,
				"type":     notifType,
				"title":    title,
				"message":  message,
				"metadata": metadata,
			},
		})
	}

	return nil
}

// getNotificationTitle returns a default title based on notification type
func getNotificationTitle(notifType string) string {
	titles := map[string]string{
		"new_message":       "New Message",
		"booking_request":   "New Booking Request",
		"booking_confirmed": "Booking Confirmed",
		"booking_cancelled": "Booking Cancelled",
		"booking_completed": "Booking Completed",
		"kyc_approved":      "KYC Approved",
		"kyc_rejected":      "KYC Rejected",
		"new_review":        "New Review",
		"payment_success":   "Payment Successful",
		"payment_failed":    "Payment Failed",
		"tier_upgraded":     "Tier Upgraded",
	}
	if title, ok := titles[notifType]; ok {
		return title
	}
	return "Notification"
}

// GetNotifications returns all notifications for the current user
// GET /notifications
func GetNotifications(c *gin.Context) {
	userID := c.GetInt("user_id")

	// Pagination
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

	// Filter by type (optional)
	notifType := c.Query("type")

	// Build query
	query := `
		SELECT id, user_id, type, title, message, metadata, is_read, read_at, created_at
		FROM notifications
		WHERE user_id = $1
	`
	args := []interface{}{userID}

	if notifType != "" {
		query += " AND type = $2"
		args = append(args, notifType)
		query += " ORDER BY created_at DESC LIMIT $3 OFFSET $4"
		args = append(args, limit, offset)
	} else {
		query += " ORDER BY created_at DESC LIMIT $2 OFFSET $3"
		args = append(args, limit, offset)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notifications"})
		return
	}
	defer rows.Close()

	notifications := []Notification{}
	for rows.Next() {
		var notif Notification
		var metadata sql.NullString

		err := rows.Scan(
			&notif.ID,
			&notif.UserID,
			&notif.Type,
			&notif.Title,
			&notif.Message,
			&metadata,
			&notif.IsRead,
			&notif.ReadAt,
			&notif.CreatedAt,
		)
		if err != nil {
			continue
		}

		// Parse metadata JSON
		if metadata.Valid {
			var meta map[string]interface{}
			if err := json.Unmarshal([]byte(metadata.String), &meta); err == nil {
				notif.Metadata = meta
			}
		}

		notifications = append(notifications, notif)
	}

	// Get total and unread count
	var total, unreadCount int
	countQuery := "SELECT COUNT(*), COUNT(*) FILTER (WHERE is_read = FALSE) FROM notifications WHERE user_id = $1"
	if notifType != "" {
		countQuery += " AND type = $2"
		db.QueryRow(countQuery, userID, notifType).Scan(&total, &unreadCount)
	} else {
		db.QueryRow(countQuery, userID).Scan(&total, &unreadCount)
	}

	c.JSON(http.StatusOK, NotificationListResponse{
		Notifications: notifications,
		Total:         total,
		UnreadCount:   unreadCount,
	})
}

// GetUnreadNotificationCount returns the count of unread notifications
// GET /notifications/unread/count
func GetUnreadNotificationCount(c *gin.Context) {
	userID := c.GetInt("userID")

	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM notifications 
		WHERE user_id = $1 AND is_read = FALSE
	`, userID).Scan(&count)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"unread_count": count})
}

// MarkNotificationAsRead marks a notification as read
// PATCH /notifications/:id/read
func MarkNotificationAsRead(c *gin.Context) {
	userID := c.GetInt("userID")
	notifID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	// Verify ownership and mark as read
	result, err := db.Exec(`
		UPDATE notifications
		SET is_read = TRUE, read_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND user_id = $2
	`, notifID, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notification"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}

// MarkAllNotificationsAsRead marks all notifications as read for the current user
// PATCH /notifications/read-all
func MarkAllNotificationsAsRead(c *gin.Context) {
	userID := c.GetInt("userID")

	result, err := db.Exec(`
		UPDATE notifications
		SET is_read = TRUE, read_at = CURRENT_TIMESTAMP
		WHERE user_id = $1 AND is_read = FALSE
	`, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notifications"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	c.JSON(http.StatusOK, gin.H{
		"message":       "All notifications marked as read",
		"updated_count": rowsAffected,
	})
}

// DeleteNotification deletes a notification
// DELETE /notifications/:id
func DeleteNotification(c *gin.Context) {
	userID := c.GetInt("userID")
	notifID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	result, err := db.Exec(`
		DELETE FROM notifications
		WHERE id = $1 AND user_id = $2
	`, notifID, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete notification"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification deleted"})
}

// DeleteAllNotifications deletes all notifications for the current user
// DELETE /notifications
func DeleteAllNotifications(c *gin.Context) {
	userID := c.GetInt("userID")

	result, err := db.Exec(`
		DELETE FROM notifications
		WHERE user_id = $1
	`, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete notifications"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	c.JSON(http.StatusOK, gin.H{
		"message":       "All notifications deleted",
		"deleted_count": rowsAffected,
	})
}
