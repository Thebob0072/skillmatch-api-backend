package main

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// --- GET /packages/:providerId (ดูแพ็คเกจของ Provider) ---
func getProviderPackagesHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		providerID, err := strconv.Atoi(c.Param("providerId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider ID"})
			return
		}

		rows, err := dbPool.Query(ctx, `
			SELECT package_id, provider_id, package_name, description, duration, price, is_active, created_at
			FROM service_packages
			WHERE provider_id = $1 AND is_active = true
			ORDER BY price ASC
		`, providerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}
		defer rows.Close()

		packages := make([]ServicePackage, 0)
		for rows.Next() {
			var pkg ServicePackage
			if err := rows.Scan(&pkg.PackageID, &pkg.ProviderID, &pkg.PackageName, &pkg.Description,
				&pkg.Duration, &pkg.Price, &pkg.IsActive, &pkg.CreatedAt); err != nil {
				continue
			}
			packages = append(packages, pkg)
		}

		c.JSON(http.StatusOK, packages)
	}
}

// --- POST /packages (สร้างแพ็คเกจ - Provider เท่านั้น) ---
func createPackageHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")

		var input struct {
			PackageName string  `json:"package_name" binding:"required"`
			Description *string `json:"description"`
			Duration    int     `json:"duration" binding:"required"`
			Price       float64 `json:"price" binding:"required"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var packageID int
		err := dbPool.QueryRow(ctx, `
			INSERT INTO service_packages (provider_id, package_name, description, duration, price)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING package_id
		`, userID, input.PackageName, input.Description, input.Duration, input.Price).Scan(&packageID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create package"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"package_id": packageID, "message": "Package created successfully"})
	}
}

// --- POST /bookings (สร้างการจอง) ---
func createBookingHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientID, _ := c.Get("userID")

		var input struct {
			ProviderID   int     `json:"provider_id" binding:"required"`
			PackageID    int     `json:"package_id" binding:"required"`
			BookingDate  string  `json:"booking_date" binding:"required"` // YYYY-MM-DD
			StartTime    string  `json:"start_time" binding:"required"`   // HH:MM
			Location     *string `json:"location"`
			SpecialNotes *string `json:"special_notes"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// ตรวจสอบ service_type ของ provider
		var serviceType *string
		err := dbPool.QueryRow(ctx, `
			SELECT p.service_type 
			FROM user_profiles p
			JOIN users u ON u.user_id = p.user_id
			WHERE u.user_id = $1
		`, input.ProviderID).Scan(&serviceType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check provider service type"})
			return
		}

		// ถ้า provider เป็น outcall → ต้องมี location จาก client
		if serviceType != nil && *serviceType == "outcall" {
			if input.Location == nil || *input.Location == "" {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Location is required for outcall services",
					"message": "ผู้ให้บริการรายนี้ให้บริการแบบไปหาลูกค้า กรุณาระบุที่อยู่ของคุณ",
				})
				return
			}
		}

		// ดึงข้อมูลแพ็คเกจ
		var duration int
		var price float64
		err = dbPool.QueryRow(ctx, `
			SELECT duration, price FROM service_packages WHERE package_id = $1 AND is_active = true
		`, input.PackageID).Scan(&duration, &price)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Package not found"})
			return
		}

		// สร้าง timestamp
		startTime, err := time.Parse("2006-01-02 15:04", input.BookingDate+" "+input.StartTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date/time format"})
			return
		}
		endTime := startTime.Add(time.Duration(duration) * time.Minute)

		// สร้างการจอง
		var bookingID int
		err = dbPool.QueryRow(ctx, `
			INSERT INTO bookings (client_id, provider_id, package_id, booking_date, start_time, end_time, 
								  total_price, status, location, special_notes)
			VALUES ($1, $2, $3, $4, $5, $6, $7, 'pending', $8, $9)
			RETURNING booking_id
		`, clientID, input.ProviderID, input.PackageID, input.BookingDate, startTime, endTime,
			price, input.Location, input.SpecialNotes).Scan(&bookingID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create booking", "details": err.Error()})
			return
		}

		// Send notification to provider
		CreateNotification(input.ProviderID, "booking_request", "You have a new booking request", map[string]interface{}{
			"booking_id": bookingID,
			"client_id":  clientID,
		})

		c.JSON(http.StatusCreated, gin.H{"booking_id": bookingID, "message": "Booking created successfully"})
	}
}

// --- GET /bookings/my (ดูการจองของตัวเอง - Client) ---
func getMyBookingsHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientID, _ := c.Get("userID")

		rows, err := dbPool.Query(ctx, `
			SELECT b.booking_id, b.client_id, u_client.username, b.provider_id, u_provider.username,
				   p.profile_image_url, sp.package_name, sp.duration, b.booking_date, b.start_time, b.end_time,
				   b.total_price, b.status, b.location, b.special_notes, b.created_at, b.updated_at
			FROM bookings b
			JOIN users u_client ON b.client_id = u_client.user_id
			JOIN users u_provider ON b.provider_id = u_provider.user_id
			JOIN service_packages sp ON b.package_id = sp.package_id
			LEFT JOIN user_profiles p ON b.provider_id = p.user_id
			WHERE b.client_id = $1
			ORDER BY b.created_at DESC
		`, clientID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}
		defer rows.Close()

		bookings := make([]BookingWithDetails, 0)
		for rows.Next() {
			var booking BookingWithDetails
			if err := rows.Scan(&booking.BookingID, &booking.ClientID, &booking.ClientUsername,
				&booking.ProviderID, &booking.ProviderUsername, &booking.ProviderProfilePic,
				&booking.PackageName, &booking.Duration, &booking.BookingDate, &booking.StartTime,
				&booking.EndTime, &booking.TotalPrice, &booking.Status, &booking.Location,
				&booking.SpecialNotes, &booking.CreatedAt, &booking.UpdatedAt); err != nil {
				continue
			}
			bookings = append(bookings, booking)
		}

		c.JSON(http.StatusOK, bookings)
	}
}

// --- GET /bookings/provider (ดูการจองที่เข้ามา - Provider) ---
func getProviderBookingsHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		providerID, _ := c.Get("userID")

		rows, err := dbPool.Query(ctx, `
			SELECT b.booking_id, b.client_id, u_client.username, b.provider_id, u_provider.username,
				   p.profile_image_url, sp.package_name, sp.duration, b.booking_date, b.start_time, b.end_time,
				   b.total_price, b.status, b.location, b.special_notes, b.created_at, b.updated_at
			FROM bookings b
			JOIN users u_client ON b.client_id = u_client.user_id
			JOIN users u_provider ON b.provider_id = u_provider.user_id
			JOIN service_packages sp ON b.package_id = sp.package_id
			LEFT JOIN user_profiles p ON b.provider_id = p.user_id
			WHERE b.provider_id = $1
			ORDER BY b.created_at DESC
		`, providerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}
		defer rows.Close()

		bookings := make([]BookingWithDetails, 0)
		for rows.Next() {
			var booking BookingWithDetails
			if err := rows.Scan(&booking.BookingID, &booking.ClientID, &booking.ClientUsername,
				&booking.ProviderID, &booking.ProviderUsername, &booking.ProviderProfilePic,
				&booking.PackageName, &booking.Duration, &booking.BookingDate, &booking.StartTime,
				&booking.EndTime, &booking.TotalPrice, &booking.Status, &booking.Location,
				&booking.SpecialNotes, &booking.CreatedAt, &booking.UpdatedAt); err != nil {
				continue
			}
			bookings = append(bookings, booking)
		}

		c.JSON(http.StatusOK, bookings)
	}
}

// --- PATCH /bookings/:id/status (อัพเดทสถานะการจอง) ---
func updateBookingStatusHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		bookingID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking ID"})
			return
		}

		userID, _ := c.Get("userID")

		var input struct {
			Status             string  `json:"status" binding:"required"`
			CancellationReason *string `json:"cancellation_reason"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// ตรวจสอบว่าเป็น provider หรือ client ของการจองนี้
		var providerID, clientID int
		err = dbPool.QueryRow(ctx, "SELECT provider_id, client_id FROM bookings WHERE booking_id = $1", bookingID).
			Scan(&providerID, &clientID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
			return
		}

		if userID != providerID && userID != clientID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
			return
		}

		// อัพเดทสถานะ
		now := time.Now()
		var query string
		if input.Status == "completed" {
			query = "UPDATE bookings SET status = $1, updated_at = $2, completed_at = $2 WHERE booking_id = $3"
		} else if input.Status == "cancelled" {
			query = "UPDATE bookings SET status = $1, updated_at = $2, cancelled_at = $2, cancellation_reason = $4 WHERE booking_id = $3"
		} else {
			query = "UPDATE bookings SET status = $1, updated_at = $2 WHERE booking_id = $3"
		}

		if input.Status == "cancelled" {
			_, err = dbPool.Exec(ctx, query, input.Status, now, bookingID, input.CancellationReason)
		} else {
			_, err = dbPool.Exec(ctx, query, input.Status, now, bookingID)
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update booking"})
			return
		}

		// Send notifications based on status change
		switch input.Status {
		case "confirmed":
			// Auto-create schedule entry when booking is confirmed
			var startTime, endTime time.Time
			var location *string
			err = dbPool.QueryRow(ctx, `
				SELECT start_time, end_time, location
				FROM bookings
				WHERE booking_id = $1
			`, bookingID).Scan(&startTime, &endTime, &location)

			if err == nil && !startTime.IsZero() && !endTime.IsZero() {
				// Insert schedule entry with status 'booked'
				_, scheduleErr := dbPool.Exec(ctx, `
					INSERT INTO provider_schedules (
						provider_id, booking_id, start_time, end_time, status,
						location_address, notes, is_visible_to_admin
					) VALUES ($1, $2, $3, $4, 'booked', $5, 'Auto-created from booking confirmation', true)
					ON CONFLICT DO NOTHING
				`, providerID, bookingID, startTime, endTime, location)

				if scheduleErr != nil {
					log.Printf("Warning: Failed to create schedule entry for booking %d: %v", bookingID, scheduleErr)
				}
			}

			CreateNotification(clientID, "booking_confirmed", "Your booking has been confirmed", map[string]interface{}{
				"booking_id":  bookingID,
				"provider_id": providerID,
			})
		case "cancelled":
			// Remove schedule entry if exists
			_, _ = dbPool.Exec(ctx, `DELETE FROM provider_schedules WHERE booking_id = $1`, bookingID)

			// Notify the other party
			if userID == providerID {
				CreateNotification(clientID, "booking_cancelled", "Your booking has been cancelled by the provider", map[string]interface{}{
					"booking_id": bookingID,
					"reason":     input.CancellationReason,
				})
			} else {
				CreateNotification(providerID, "booking_cancelled", "The client has cancelled the booking", map[string]interface{}{
					"booking_id": bookingID,
					"reason":     input.CancellationReason,
				})
			}
		case "completed":
			CreateNotification(clientID, "booking_completed", "Your booking has been completed", map[string]interface{}{
				"booking_id": bookingID,
			})
		}

		c.JSON(http.StatusOK, gin.H{"message": "Booking status updated"})
	}
}
