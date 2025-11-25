package main

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Block struct
type Block struct {
	ID        int       `json:"id"`
	BlockerID int       `json:"blocker_id"`
	BlockedID int       `json:"blocked_id"`
	Reason    *string   `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
}

type BlockedUser struct {
	UserID               int     `json:"user_id"`
	Username             string  `json:"username"`
	ProfileImageUrl      *string `json:"profile_image_url"`
	GoogleProfilePicture *string `json:"google_profile_picture"`
	Reason               *string `json:"reason"`
	BlockedAt            string  `json:"blocked_at"`
}

// --- POST /blocks (Block a user) ---
func blockUserHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")
		blockerID := userID.(int)

		var input struct {
			BlockedUserID int     `json:"blocked_user_id" binding:"required"`
			Reason        *string `json:"reason"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Prevent self-blocking
		if blockerID == input.BlockedUserID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot block yourself"})
			return
		}

		// Check if already blocked
		var exists bool
		err := dbPool.QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM blocks WHERE blocker_id = $1 AND blocked_id = $2)
		`, blockerID, input.BlockedUserID).Scan(&exists)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		if exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User already blocked"})
			return
		}

		// Insert block
		var blockID int
		err = dbPool.QueryRow(ctx, `
			INSERT INTO blocks (blocker_id, blocked_id, reason, created_at)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`, blockerID, input.BlockedUserID, input.Reason, time.Now()).Scan(&blockID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to block user"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":  "User blocked successfully",
			"block_id": blockID,
		})
	}
}

// --- DELETE /blocks/:userId (Unblock a user) ---
func unblockUserHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")
		blockerID := userID.(int)

		blockedUserID, err := strconv.Atoi(c.Param("userId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		// Delete block
		result, err := dbPool.Exec(ctx, `
			DELETE FROM blocks 
			WHERE blocker_id = $1 AND blocked_id = $2
		`, blockerID, blockedUserID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unblock user"})
			return
		}

		rowsAffected := result.RowsAffected()
		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Block not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "User unblocked successfully",
		})
	}
}

// --- GET /blocks (Get blocked users list) ---
func getBlockedUsersHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")
		blockerID := userID.(int)

		rows, err := dbPool.Query(ctx, `
			SELECT 
				u.user_id, 
				u.username,
				p.profile_image_url,
				p.google_profile_picture,
				b.reason,
				b.created_at
			FROM blocks b
			JOIN users u ON b.blocked_id = u.user_id
			LEFT JOIN user_profiles p ON u.user_id = p.user_id
			WHERE b.blocker_id = $1
			ORDER BY b.created_at DESC
		`, blockerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch blocked users"})
			return
		}
		defer rows.Close()

		blockedUsers := make([]BlockedUser, 0)
		for rows.Next() {
			var user BlockedUser
			if err := rows.Scan(&user.UserID, &user.Username, &user.ProfileImageUrl,
				&user.GoogleProfilePicture, &user.Reason, &user.BlockedAt); err != nil {
				continue
			}
			blockedUsers = append(blockedUsers, user)
		}

		c.JSON(http.StatusOK, gin.H{
			"blocked_users": blockedUsers,
			"total":         len(blockedUsers),
		})
	}
}

// --- GET /blocks/check/:userId (Check if user is blocked) ---
func checkBlockStatusHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")
		blockerID := userID.(int)

		targetUserID, err := strconv.Atoi(c.Param("userId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		var isBlocked bool
		err = dbPool.QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM blocks WHERE blocker_id = $1 AND blocked_id = $2)
		`, blockerID, targetUserID).Scan(&isBlocked)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// Also check if the target user has blocked you
		var isBlockedBy bool
		err = dbPool.QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM blocks WHERE blocker_id = $1 AND blocked_id = $2)
		`, targetUserID, blockerID).Scan(&isBlockedBy)

		if err != nil {
			isBlockedBy = false
		}

		c.JSON(http.StatusOK, gin.H{
			"is_blocked":    isBlocked,   // You blocked this user
			"is_blocked_by": isBlockedBy, // This user blocked you
		})
	}
}

// Helper function: Check if user A has blocked user B
func IsUserBlocked(dbPool *pgxpool.Pool, ctx context.Context, blockerID, blockedID int) bool {
	var exists bool
	err := dbPool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM blocks WHERE blocker_id = $1 AND blocked_id = $2)
	`, blockerID, blockedID).Scan(&exists)

	return err == nil && exists
}
