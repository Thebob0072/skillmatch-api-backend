package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ================================
// Trusted Contact Handlers
// ================================

// POST /safety/trusted-contacts
func addTrustedContactHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")

		var input AddTrustedContactRequest
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var contactID int
		err := dbPool.QueryRow(ctx, `
			INSERT INTO trusted_contacts (user_id, name, phone_number, relationship)
			VALUES ($1, $2, $3, $4)
			RETURNING contact_id
		`, userID, input.Name, input.PhoneNumber, input.Relationship).Scan(&contactID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add trusted contact"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"contact_id": contactID,
			"message":    "Trusted contact added successfully",
		})
	}
}

// GET /safety/trusted-contacts
func getTrustedContactsHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")

		rows, err := dbPool.Query(ctx, `
			SELECT contact_id, user_id, name, phone_number, relationship, is_active, created_at, updated_at, last_notified
			FROM trusted_contacts
			WHERE user_id = $1 AND is_active = true
			ORDER BY created_at DESC
		`, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch contacts"})
			return
		}
		defer rows.Close()

		contacts := make([]TrustedContact, 0)
		for rows.Next() {
			var contact TrustedContact
			if err := rows.Scan(&contact.ContactID, &contact.UserID, &contact.Name,
				&contact.PhoneNumber, &contact.Relationship, &contact.IsActive,
				&contact.CreatedAt, &contact.UpdatedAt, &contact.LastNotified); err != nil {
				continue
			}
			contacts = append(contacts, contact)
		}

		c.JSON(http.StatusOK, contacts)
	}
}

// DELETE /safety/trusted-contacts/:id
func deleteTrustedContactHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")
		contactID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid contact ID"})
			return
		}

		result, err := dbPool.Exec(ctx, `
			UPDATE trusted_contacts SET is_active = false, updated_at = NOW()
			WHERE contact_id = $1 AND user_id = $2
		`, contactID, userID)

		if err != nil || result.RowsAffected() == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Contact not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Contact removed successfully"})
	}
}

// ================================
// Emergency SOS Handlers
// ================================

// POST /safety/sos - Trigger SOS Alert
func triggerSOSHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")

		var input TriggerSOSRequest
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Create SOS Alert
		var alertID int
		err := dbPool.QueryRow(ctx, `
			INSERT INTO sos_alerts (user_id, booking_id, latitude, longitude, location_text, status)
			VALUES ($1, $2, $3, $4, $5, 'active')
			RETURNING alert_id
		`, userID, input.BookingID, input.Latitude, input.Longitude, input.LocationText).Scan(&alertID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create SOS alert"})
			return
		}

		// Get user info
		var userName string
		var userPhone *string
		dbPool.QueryRow(ctx, `SELECT username, phone_number FROM users WHERE user_id = $1`, userID).Scan(&userName, &userPhone)

		// Notify all trusted contacts
		rows, _ := dbPool.Query(ctx, `
			SELECT contact_id, phone_number, name FROM trusted_contacts
			WHERE user_id = $1 AND is_active = true
		`, userID)
		defer rows.Close()

		contactsNotified := 0
		for rows.Next() {
			var contactID int
			var phone, name string
			rows.Scan(&contactID, &phone, &name)

			// TODO: Send SMS via Twilio or other provider
			// For now, just log and update last_notified
			log.Printf("🚨 SOS Alert! Notifying %s at %s for user %s (ID: %d)", name, phone, userName, userID)

			dbPool.Exec(ctx, `UPDATE trusted_contacts SET last_notified = NOW() WHERE contact_id = $1`, contactID)
			contactsNotified++
		}

		// Notify all admins
		CreateNotification(0, "sos_alert", fmt.Sprintf("🚨 SOS Alert from %s", userName), map[string]interface{}{
			"alert_id":  alertID,
			"user_id":   userID,
			"latitude":  input.Latitude,
			"longitude": input.Longitude,
		})

		c.JSON(http.StatusCreated, gin.H{
			"alert_id":          alertID,
			"message":           "SOS Alert triggered successfully",
			"contacts_notified": contactsNotified,
			"location": gin.H{
				"latitude":  input.Latitude,
				"longitude": input.Longitude,
				"text":      input.LocationText,
			},
		})
	}
}

// PATCH /safety/sos/:id/resolve - Resolve SOS Alert (Admin only)
func resolveSOSHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")
		alertID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
			return
		}

		var input struct {
			ResolutionNote string `json:"resolution_note"`
		}
		c.ShouldBindJSON(&input)

		result, err := dbPool.Exec(ctx, `
			UPDATE sos_alerts 
			SET status = 'resolved', resolved_at = NOW(), resolved_by = $1, resolution_note = $2, updated_at = NOW()
			WHERE alert_id = $3
		`, userID, input.ResolutionNote, alertID)

		if err != nil || result.RowsAffected() == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Alert not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "SOS Alert resolved successfully"})
	}
}

// GET /safety/sos/active - Get active SOS alerts (Admin only)
func getActiveSOSAlertsHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := dbPool.Query(ctx, `
			SELECT s.alert_id, s.user_id, u.username, u.phone_number, 
				   s.booking_id, s.latitude, s.longitude, s.location_text, 
				   s.status, s.created_at
			FROM sos_alerts s
			JOIN users u ON s.user_id = u.user_id
			WHERE s.status = 'active'
			ORDER BY s.created_at DESC
		`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch SOS alerts"})
			return
		}
		defer rows.Close()

		alerts := make([]gin.H, 0)
		for rows.Next() {
			var alertID, userID int
			var username string
			var phone, locationText *string
			var bookingID *int
			var lat, lng float64
			var status string
			var createdAt time.Time

			rows.Scan(&alertID, &userID, &username, &phone, &bookingID, &lat, &lng, &locationText, &status, &createdAt)

			alerts = append(alerts, gin.H{
				"alert_id":      alertID,
				"user_id":       userID,
				"username":      username,
				"phone":         phone,
				"booking_id":    bookingID,
				"latitude":      lat,
				"longitude":     lng,
				"location_text": locationText,
				"status":        status,
				"created_at":    createdAt,
			})
		}

		c.JSON(http.StatusOK, alerts)
	}
}

// ================================
// Check-in/Check-out Handlers
// ================================

// POST /safety/check-in
func checkInHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")

		var input CheckInRequest
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Verify booking exists and user is provider
		var providerID, clientID int
		var duration int
		err := dbPool.QueryRow(ctx, `
			SELECT b.provider_id, b.client_id, sp.duration
			FROM bookings b
			JOIN service_packages sp ON b.package_id = sp.package_id
			WHERE b.booking_id = $1 AND b.status = 'confirmed'
		`, input.BookingID).Scan(&providerID, &clientID, &duration)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found or not confirmed"})
			return
		}

		if userID != providerID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only provider can check-in"})
			return
		}

		// Create check-in record
		now := time.Now()
		expectedEnd := now.Add(time.Duration(duration) * time.Minute)

		var checkInID int
		err = dbPool.QueryRow(ctx, `
			INSERT INTO booking_check_ins (booking_id, provider_id, client_id, checked_in_at, expected_end_time, latitude, longitude, notes, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'active')
			RETURNING check_in_id
		`, input.BookingID, providerID, clientID, now, expectedEnd, input.Latitude, input.Longitude, input.Notes).Scan(&checkInID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check-in"})
			return
		}

		// Notify client
		CreateNotification(clientID, "booking_checkin", "Provider has checked in for your booking", map[string]interface{}{
			"booking_id":  input.BookingID,
			"check_in_id": checkInID,
		})

		c.JSON(http.StatusCreated, gin.H{
			"check_in_id":       checkInID,
			"checked_in_at":     now,
			"expected_end_time": expectedEnd,
			"duration_minutes":  duration,
			"message":           "Check-in successful",
		})
	}
}

// POST /safety/check-out
func checkOutHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")

		var input CheckOutRequest
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Verify check-in exists
		var checkInID, providerID, clientID int
		var checkedInAt time.Time
		err := dbPool.QueryRow(ctx, `
			SELECT check_in_id, provider_id, client_id, checked_in_at
			FROM booking_check_ins
			WHERE booking_id = $1 AND status = 'active'
		`, input.BookingID).Scan(&checkInID, &providerID, &clientID, &checkedInAt)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "No active check-in found"})
			return
		}

		if userID != providerID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only provider can check-out"})
			return
		}

		now := time.Now()
		duration := now.Sub(checkedInAt)

		// Update check-in record
		_, err = dbPool.Exec(ctx, `
			UPDATE booking_check_ins 
			SET checked_out_at = $1, status = 'completed', updated_at = NOW()
			WHERE check_in_id = $2
		`, now, checkInID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check-out"})
			return
		}

		// Update booking status to completed
		dbPool.Exec(ctx, `UPDATE bookings SET status = 'completed', completed_at = $1 WHERE booking_id = $2`, now, input.BookingID)

		// === ESCROW RELEASE: Move money from pending to available ===
		// Get booking amount from transactions
		var providerEarnings float64
		err = dbPool.QueryRow(ctx, `
			SELECT net_amount FROM transactions 
			WHERE booking_id = $1 AND type = 'booking_payment' AND status = 'completed'
			ORDER BY created_at DESC LIMIT 1
		`, input.BookingID).Scan(&providerEarnings)

		if err == nil && providerEarnings > 0 {
			// Move from pending_balance to available_balance
			_, err = dbPool.Exec(ctx, `
				UPDATE wallets 
				SET pending_balance = pending_balance - $1,
				    available_balance = available_balance + $1,
				    updated_at = NOW()
				WHERE user_id = $2 AND pending_balance >= $1
			`, providerEarnings, providerID)

			if err != nil {
				fmt.Printf("Warning: Failed to release escrow for booking %d: %v\n", input.BookingID, err)
			} else {
				// Create wallet transaction record
				dbPool.Exec(ctx, `
					INSERT INTO wallet_transactions (user_id, type, amount, balance_before, balance_after, description, booking_id, status)
					SELECT $1, 'escrow_release', $2, 
						   available_balance - $2, available_balance,
						   'Escrow released after service completion', $3, 'completed'
					FROM wallets WHERE user_id = $1
				`, providerID, providerEarnings, input.BookingID)

				fmt.Printf("✅ Escrow released: Provider=%d, Amount=฿%.2f, Booking=%d\n", providerID, providerEarnings, input.BookingID)
			}
		}

		// Notify client
		CreateNotification(clientID, "booking_checkout", "Provider has checked out. Session completed.", map[string]interface{}{
			"booking_id":       input.BookingID,
			"duration_minutes": int(duration.Minutes()),
		})

		// Notify provider about payment release
		CreateNotification(providerID, "payment_released", "Payment has been released to your wallet", map[string]interface{}{
			"booking_id": input.BookingID,
			"amount":     providerEarnings,
		})

		c.JSON(http.StatusOK, gin.H{
			"message":          "Check-out successful",
			"checked_out_at":   now,
			"duration_minutes": int(duration.Minutes()),
			"payment_released": providerEarnings,
		})
	}
}

// GET /safety/check-ins/active - Get active check-ins (Admin)
func getActiveCheckInsHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := dbPool.Query(ctx, `
			SELECT ci.check_in_id, ci.booking_id, ci.provider_id, ci.client_id,
				   u_p.username as provider_name, u_c.username as client_name,
				   ci.checked_in_at, ci.expected_end_time, ci.latitude, ci.longitude, ci.status,
				   CASE WHEN ci.expected_end_time < NOW() THEN true ELSE false END as is_overdue
			FROM booking_check_ins ci
			JOIN users u_p ON ci.provider_id = u_p.user_id
			JOIN users u_c ON ci.client_id = u_c.user_id
			WHERE ci.status = 'active'
			ORDER BY ci.expected_end_time ASC
		`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch check-ins"})
			return
		}
		defer rows.Close()

		checkIns := make([]gin.H, 0)
		for rows.Next() {
			var checkInID, bookingID, providerID, clientID int
			var providerName, clientName, status string
			var checkedInAt, expectedEnd time.Time
			var lat, lng *float64
			var isOverdue bool

			rows.Scan(&checkInID, &bookingID, &providerID, &clientID,
				&providerName, &clientName, &checkedInAt, &expectedEnd,
				&lat, &lng, &status, &isOverdue)

			checkIns = append(checkIns, gin.H{
				"check_in_id":       checkInID,
				"booking_id":        bookingID,
				"provider_id":       providerID,
				"provider_name":     providerName,
				"client_id":         clientID,
				"client_name":       clientName,
				"checked_in_at":     checkedInAt,
				"expected_end_time": expectedEnd,
				"latitude":          lat,
				"longitude":         lng,
				"status":            status,
				"is_overdue":        isOverdue,
			})
		}

		c.JSON(http.StatusOK, checkIns)
	}
}

// ================================
// Private Gallery Handlers
// ================================

// GET /gallery/private/settings
func getPrivateGallerySettingsHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")

		var settings PrivateGallerySettings
		err := dbPool.QueryRow(ctx, `
			SELECT setting_id, user_id, is_enabled, monthly_price, one_time_price, allow_one_time, updated_at
			FROM private_gallery_settings
			WHERE user_id = $1
		`, userID).Scan(&settings.SettingID, &settings.UserID, &settings.IsEnabled,
			&settings.MonthlyPrice, &settings.OneTimePrice, &settings.AllowOneTime, &settings.UpdatedAt)

		if err == sql.ErrNoRows {
			// Return default settings
			c.JSON(http.StatusOK, PrivateGallerySettings{
				UserID:       userID.(int),
				IsEnabled:    false,
				AllowOneTime: true,
			})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch settings"})
			return
		}

		c.JSON(http.StatusOK, settings)
	}
}

// PUT /gallery/private/settings
func updatePrivateGallerySettingsHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")

		var input PrivateGallerySettingsRequest
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, err := dbPool.Exec(ctx, `
			INSERT INTO private_gallery_settings (user_id, is_enabled, monthly_price, one_time_price, allow_one_time, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW())
			ON CONFLICT (user_id) DO UPDATE SET
				is_enabled = $2, monthly_price = $3, one_time_price = $4, allow_one_time = $5, updated_at = NOW()
		`, userID, input.IsEnabled, input.MonthlyPrice, input.OneTimePrice, input.AllowOneTime)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Settings updated successfully"})
	}
}

// POST /gallery/private/photos
func uploadPrivatePhotoHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")

		file, err := c.FormFile("photo")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No photo provided"})
			return
		}

		// Get max sort order
		var maxOrder int
		dbPool.QueryRow(ctx, `SELECT COALESCE(MAX(sort_order), 0) FROM private_photos WHERE user_id = $1`, userID).Scan(&maxOrder)

		// TODO: Upload to GCS and get URL
		photoURL := fmt.Sprintf("https://storage.example.com/private/%d/%s", userID, file.Filename)
		thumbnailURL := fmt.Sprintf("https://storage.example.com/private/%d/thumb_%s", userID, file.Filename)

		var photoID int
		err = dbPool.QueryRow(ctx, `
			INSERT INTO private_photos (user_id, photo_url, thumbnail_url, sort_order)
			VALUES ($1, $2, $3, $4)
			RETURNING photo_id
		`, userID, photoURL, thumbnailURL, maxOrder+1).Scan(&photoID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save photo"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"photo_id":      photoID,
			"photo_url":     photoURL,
			"thumbnail_url": thumbnailURL,
		})
	}
}

// GET /gallery/private/:userId - View private gallery (requires access)
func getPrivateGalleryHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		viewerID, _ := c.Get("userID")
		ownerID, err := strconv.Atoi(c.Param("userId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		// Check if viewer has access
		hasAccess := false
		if viewerID == ownerID {
			hasAccess = true // Owner can always view
		} else {
			var accessID int
			err = dbPool.QueryRow(ctx, `
				SELECT access_id FROM private_gallery_access
				WHERE gallery_owner_id = $1 AND viewer_id = $2
				AND (expires_at IS NULL OR expires_at > NOW())
			`, ownerID, viewerID).Scan(&accessID)
			hasAccess = err == nil
		}

		// Get photos (show thumbnails if no access, full if access)
		rows, err := dbPool.Query(ctx, `
			SELECT photo_id, photo_url, thumbnail_url, sort_order, price, uploaded_at
			FROM private_photos
			WHERE user_id = $1 AND is_active = true
			ORDER BY sort_order ASC
		`, ownerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch gallery"})
			return
		}
		defer rows.Close()

		photos := make([]gin.H, 0)
		for rows.Next() {
			var photoID, sortOrder int
			var photoURL, thumbnailURL *string
			var price *float64
			var uploadedAt time.Time

			rows.Scan(&photoID, &photoURL, &thumbnailURL, &sortOrder, &price, &uploadedAt)

			photo := gin.H{
				"photo_id":    photoID,
				"sort_order":  sortOrder,
				"price":       price,
				"uploaded_at": uploadedAt,
			}

			if hasAccess {
				photo["photo_url"] = photoURL
			} else {
				photo["thumbnail_url"] = thumbnailURL // Blurred version
			}

			photos = append(photos, photo)
		}

		// Get settings
		var settings struct {
			MonthlyPrice *float64
			OneTimePrice *float64
			AllowOneTime bool
		}
		dbPool.QueryRow(ctx, `
			SELECT monthly_price, one_time_price, allow_one_time
			FROM private_gallery_settings WHERE user_id = $1
		`, ownerID).Scan(&settings.MonthlyPrice, &settings.OneTimePrice, &settings.AllowOneTime)

		c.JSON(http.StatusOK, gin.H{
			"has_access":     hasAccess,
			"photos":         photos,
			"monthly_price":  settings.MonthlyPrice,
			"one_time_price": settings.OneTimePrice,
			"allow_one_time": settings.AllowOneTime,
		})
	}
}

// POST /gallery/private/purchase
func purchaseGalleryAccessHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		viewerID, _ := c.Get("userID")

		var input PurchaseGalleryRequest
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get pricing
		var monthlyPrice, oneTimePrice *float64
		var allowOneTime bool
		err := dbPool.QueryRow(ctx, `
			SELECT monthly_price, one_time_price, allow_one_time
			FROM private_gallery_settings WHERE user_id = $1 AND is_enabled = true
		`, input.ProviderID).Scan(&monthlyPrice, &oneTimePrice, &allowOneTime)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Private gallery not available"})
			return
		}

		var price float64
		var expiresAt *time.Time

		if input.AccessType == "subscription" && monthlyPrice != nil {
			price = *monthlyPrice
			exp := time.Now().AddDate(0, 1, 0) // 1 month
			expiresAt = &exp
		} else if input.AccessType == "one_time" && allowOneTime && oneTimePrice != nil {
			price = *oneTimePrice
			expiresAt = nil // Permanent
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid access type or pricing not set"})
			return
		}

		// TODO: Process payment via Stripe
		// For now, just grant access

		var accessID int
		err = dbPool.QueryRow(ctx, `
			INSERT INTO private_gallery_access (gallery_owner_id, viewer_id, access_type, expires_at)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (gallery_owner_id, viewer_id) DO UPDATE SET
				access_type = $3, expires_at = $4, granted_at = NOW()
			RETURNING access_id
		`, input.ProviderID, viewerID, input.AccessType, expiresAt).Scan(&accessID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to grant access"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"access_id":   accessID,
			"access_type": input.AccessType,
			"expires_at":  expiresAt,
			"price":       price,
			"message":     "Gallery access granted successfully",
		})
	}
}
