package main

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// createScheduleHandler - Provider สร้างตารางงาน/ลงคิว
func createScheduleHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		// ดึง userID จาก JWT middleware
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var req struct {
			StartTime        string   `json:"start_time" binding:"required"` // ISO 8601
			EndTime          string   `json:"end_time" binding:"required"`   // ISO 8601
			Status           string   `json:"status"`                        // available, blocked
			LocationType     *string  `json:"location_type"`                 // Incall, Outcall, Both
			LocationAddress  *string  `json:"location_address"`
			LocationProvince *string  `json:"location_province"`
			LocationDistrict *string  `json:"location_district"`
			Latitude         *float64 `json:"latitude"`
			Longitude        *float64 `json:"longitude"`
			Notes            *string  `json:"notes"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Parse time
		startTime, err := time.Parse(time.RFC3339, req.StartTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_time format (use ISO 8601)"})
			return
		}

		endTime, err := time.Parse(time.RFC3339, req.EndTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_time format (use ISO 8601)"})
			return
		}

		// Validate times
		if endTime.Before(startTime) || endTime.Equal(startTime) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "end_time must be after start_time"})
			return
		}

		// Default status
		status := "available"
		if req.Status != "" {
			if req.Status != "available" && req.Status != "blocked" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Status must be 'available' or 'blocked'"})
				return
			}
			status = req.Status
		}

		// ตรวจสอบ overlap กับ schedule ที่มีอยู่แล้ว
		var overlapping int
		err = dbPool.QueryRow(ctx, `
			SELECT COUNT(*) FROM provider_schedules
			WHERE provider_id = $1
			AND (
				(start_time <= $2 AND end_time > $2) OR -- New start falls within existing
				(start_time < $3 AND end_time >= $3) OR -- New end falls within existing
				(start_time >= $2 AND end_time <= $3)   -- New schedule contains existing
			)
		`, userID, startTime, endTime).Scan(&overlapping)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error checking overlaps"})
			return
		}

		if overlapping > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Schedule overlaps with existing entry"})
			return
		}

		// Insert schedule
		var scheduleID int
		err = dbPool.QueryRow(ctx, `
			INSERT INTO provider_schedules (
				provider_id, start_time, end_time, status,
				location_type, location_address, location_province, location_district,
				latitude, longitude, notes, is_visible_to_admin
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, true)
			RETURNING schedule_id
		`, userID, startTime, endTime, status,
			req.LocationType, req.LocationAddress, req.LocationProvince, req.LocationDistrict,
			req.Latitude, req.Longitude, req.Notes).Scan(&scheduleID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create schedule", "details": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":     "Schedule created successfully",
			"schedule_id": scheduleID,
		})
	}
}

// getMySchedulesHandler - Provider ดูตารางงานของตัวเอง
func getMySchedulesHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Query parameters
		startDate := c.Query("start_date") // YYYY-MM-DD
		endDate := c.Query("end_date")     // YYYY-MM-DD
		status := c.Query("status")        // available, booked, blocked

		query := `
			SELECT 
				schedule_id, provider_id, booking_id, start_time, end_time, status,
				location_type, location_address, location_province, location_district,
				latitude, longitude, notes, is_visible_to_admin, created_at, updated_at
			FROM provider_schedules
			WHERE provider_id = $1
		`
		args := []interface{}{userID}
		argCounter := 2

		// Filter by date range
		if startDate != "" {
			query += " AND start_time >= $" + strconv.Itoa(argCounter)
			args = append(args, startDate)
			argCounter++
		}
		if endDate != "" {
			query += " AND end_time <= $" + strconv.Itoa(argCounter)
			args = append(args, endDate)
			argCounter++
		}

		// Filter by status
		if status != "" {
			query += " AND status = $" + strconv.Itoa(argCounter)
			args = append(args, status)
		}

		query += " ORDER BY start_time ASC"

		rows, err := dbPool.Query(ctx, query, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch schedules"})
			return
		}
		defer rows.Close()

		var schedules []ProviderSchedule
		for rows.Next() {
			var s ProviderSchedule
			err := rows.Scan(
				&s.ScheduleID, &s.ProviderID, &s.BookingID, &s.StartTime, &s.EndTime, &s.Status,
				&s.LocationType, &s.LocationAddress, &s.LocationProvince, &s.LocationDistrict,
				&s.Latitude, &s.Longitude, &s.Notes, &s.IsVisibleToAdmin, &s.CreatedAt, &s.UpdatedAt,
			)
			if err != nil {
				continue
			}
			schedules = append(schedules, s)
		}

		c.JSON(http.StatusOK, gin.H{
			"schedules": schedules,
			"total":     len(schedules),
		})
	}
}

// updateScheduleHandler - Provider แก้ไขตารางงาน
func updateScheduleHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		scheduleID := c.Param("scheduleId")
		if scheduleID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "schedule_id is required"})
			return
		}

		var req struct {
			StartTime        *string  `json:"start_time"` // ISO 8601
			EndTime          *string  `json:"end_time"`   // ISO 8601
			Status           *string  `json:"status"`     // available, blocked
			LocationType     *string  `json:"location_type"`
			LocationAddress  *string  `json:"location_address"`
			LocationProvince *string  `json:"location_province"`
			LocationDistrict *string  `json:"location_district"`
			Latitude         *float64 `json:"latitude"`
			Longitude        *float64 `json:"longitude"`
			Notes            *string  `json:"notes"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Verify ownership
		var ownerID int
		var currentStatus string
		err := dbPool.QueryRow(ctx, `
			SELECT provider_id, status FROM provider_schedules WHERE schedule_id = $1
		`, scheduleID).Scan(&ownerID, &currentStatus)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Schedule not found"})
			return
		}

		if ownerID != userID.(int) {
			c.JSON(http.StatusForbidden, gin.H{"error": "You can only update your own schedule"})
			return
		}

		// Cannot update booked schedules
		if currentStatus == "booked" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot update booked schedule. Please cancel booking first."})
			return
		}

		// Build dynamic update query
		query := "UPDATE provider_schedules SET "
		args := []interface{}{}
		argCounter := 1
		updates := []string{}

		if req.StartTime != nil {
			startTime, err := time.Parse(time.RFC3339, *req.StartTime)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_time format"})
				return
			}
			updates = append(updates, "start_time = $"+strconv.Itoa(argCounter))
			args = append(args, startTime)
			argCounter++
		}

		if req.EndTime != nil {
			endTime, err := time.Parse(time.RFC3339, *req.EndTime)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_time format"})
				return
			}
			updates = append(updates, "end_time = $"+strconv.Itoa(argCounter))
			args = append(args, endTime)
			argCounter++
		}

		if req.Status != nil {
			if *req.Status != "available" && *req.Status != "blocked" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Status must be 'available' or 'blocked'"})
				return
			}
			updates = append(updates, "status = $"+strconv.Itoa(argCounter))
			args = append(args, *req.Status)
			argCounter++
		}

		if req.LocationType != nil {
			updates = append(updates, "location_type = $"+strconv.Itoa(argCounter))
			args = append(args, req.LocationType)
			argCounter++
		}

		if req.LocationAddress != nil {
			updates = append(updates, "location_address = $"+strconv.Itoa(argCounter))
			args = append(args, req.LocationAddress)
			argCounter++
		}

		if req.LocationProvince != nil {
			updates = append(updates, "location_province = $"+strconv.Itoa(argCounter))
			args = append(args, req.LocationProvince)
			argCounter++
		}

		if req.LocationDistrict != nil {
			updates = append(updates, "location_district = $"+strconv.Itoa(argCounter))
			args = append(args, req.LocationDistrict)
			argCounter++
		}

		if req.Latitude != nil {
			updates = append(updates, "latitude = $"+strconv.Itoa(argCounter))
			args = append(args, req.Latitude)
			argCounter++
		}

		if req.Longitude != nil {
			updates = append(updates, "longitude = $"+strconv.Itoa(argCounter))
			args = append(args, req.Longitude)
			argCounter++
		}

		if req.Notes != nil {
			updates = append(updates, "notes = $"+strconv.Itoa(argCounter))
			args = append(args, req.Notes)
			argCounter++
		}

		if len(updates) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
			return
		}

		// ถ้ามีการเปลี่ยน start_time หรือ end_time ต้องตรวจสอบ overlap
		if req.StartTime != nil || req.EndTime != nil {
			// ดึงข้อมูลปัจจุบัน
			var currentStart, currentEnd time.Time
			err := dbPool.QueryRow(ctx, `
				SELECT start_time, end_time FROM provider_schedules WHERE schedule_id = $1
			`, scheduleID).Scan(&currentStart, &currentEnd)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch current schedule"})
				return
			}

			// ใช้ค่าใหม่ถ้ามี ไม่งั้นใช้ค่าเก่า
			newStart := currentStart
			newEnd := currentEnd

			if req.StartTime != nil {
				newStart, _ = time.Parse(time.RFC3339, *req.StartTime)
			}
			if req.EndTime != nil {
				newEnd, _ = time.Parse(time.RFC3339, *req.EndTime)
			}

			// ตรวจสอบ overlap (ยกเว้น schedule ปัจจุบัน)
			var overlapping int
			err = dbPool.QueryRow(ctx, `
				SELECT COUNT(*) FROM provider_schedules
				WHERE provider_id = $1
				AND schedule_id != $2
				AND (
					(start_time <= $3 AND end_time > $3) OR
					(start_time < $4 AND end_time >= $4) OR
					(start_time >= $3 AND end_time <= $4)
				)
			`, userID, scheduleID, newStart, newEnd).Scan(&overlapping)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error checking overlaps"})
				return
			}

			if overlapping > 0 {
				c.JSON(http.StatusConflict, gin.H{"error": "Updated schedule would overlap with existing entry"})
				return
			}
		}

		// Complete query
		for i, update := range updates {
			query += update
			if i < len(updates)-1 {
				query += ", "
			}
		}
		query += " WHERE schedule_id = $" + strconv.Itoa(argCounter)
		args = append(args, scheduleID)

		_, err = dbPool.Exec(ctx, query, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update schedule", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Schedule updated successfully"})
	}
}

// deleteScheduleHandler - Provider ลบตารางงาน
func deleteScheduleHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		scheduleID := c.Param("scheduleId")
		if scheduleID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "schedule_id is required"})
			return
		}

		// Verify ownership and status
		var ownerID int
		var status string
		err := dbPool.QueryRow(ctx, `
			SELECT provider_id, status FROM provider_schedules WHERE schedule_id = $1
		`, scheduleID).Scan(&ownerID, &status)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Schedule not found"})
			return
		}

		if ownerID != userID.(int) {
			c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete your own schedule"})
			return
		}

		// Cannot delete booked schedules
		if status == "booked" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete booked schedule. Please cancel booking first."})
			return
		}

		// Delete schedule
		_, err = dbPool.Exec(ctx, `DELETE FROM provider_schedules WHERE schedule_id = $1`, scheduleID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete schedule"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Schedule deleted successfully"})
	}
}

// getProviderScheduleAdminHandler - Admin/GOD ดูตารางงานของ Provider
func getProviderScheduleAdminHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check admin permission
		isAdmin, _ := c.Get("isAdmin")
		if !isAdmin.(bool) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			return
		}

		providerID := c.Param("providerId")
		if providerID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "provider_id is required"})
			return
		}

		// Query parameters
		startDate := c.Query("start_date")
		endDate := c.Query("end_date")
		status := c.Query("status")

		query := `
			SELECT 
				s.schedule_id, s.provider_id, u.username as provider_username, u.phone as provider_phone,
				s.booking_id, b.client_id, uc.username as client_username,
				s.start_time, s.end_time, s.status,
				s.location_type, s.location_address, s.location_province, s.location_district,
				s.latitude, s.longitude, s.notes, s.created_at, s.updated_at
			FROM provider_schedules s
			JOIN users u ON s.provider_id = u.user_id
			LEFT JOIN bookings b ON s.booking_id = b.booking_id
			LEFT JOIN users uc ON b.client_id = uc.user_id
			WHERE s.provider_id = $1 AND s.is_visible_to_admin = true
		`
		args := []interface{}{providerID}
		argCounter := 2

		if startDate != "" {
			query += " AND s.start_time >= $" + strconv.Itoa(argCounter)
			args = append(args, startDate)
			argCounter++
		}
		if endDate != "" {
			query += " AND s.end_time <= $" + strconv.Itoa(argCounter)
			args = append(args, endDate)
			argCounter++
		}
		if status != "" {
			query += " AND s.status = $" + strconv.Itoa(argCounter)
			args = append(args, status)
		}

		query += " ORDER BY s.start_time ASC"

		rows, err := dbPool.Query(ctx, query, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch schedules", "details": err.Error()})
			return
		}
		defer rows.Close()

		var schedules []ProviderScheduleWithDetails
		for rows.Next() {
			var s ProviderScheduleWithDetails
			err := rows.Scan(
				&s.ScheduleID, &s.ProviderID, &s.ProviderUsername, &s.ProviderPhone,
				&s.BookingID, &s.ClientID, &s.ClientUsername,
				&s.StartTime, &s.EndTime, &s.Status,
				&s.LocationType, &s.LocationAddress, &s.LocationProvince, &s.LocationDistrict,
				&s.Latitude, &s.Longitude, &s.Notes, &s.CreatedAt, &s.UpdatedAt,
			)
			if err != nil {
				continue
			}
			schedules = append(schedules, s)
		}

		c.JSON(http.StatusOK, gin.H{
			"provider_id": providerID,
			"schedules":   schedules,
			"total":       len(schedules),
		})
	}
}

// getAllProvidersScheduleAdminHandler - Admin/GOD ดูตารางงานของ Provider ทั้งหมด
func getAllProvidersScheduleAdminHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check admin permission
		isAdmin, _ := c.Get("isAdmin")
		if !isAdmin.(bool) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			return
		}

		// Query parameters
		startDate := c.Query("start_date")
		endDate := c.Query("end_date")
		status := c.Query("status")
		province := c.Query("province")

		query := `
			SELECT 
				s.schedule_id, s.provider_id, u.username as provider_username, u.phone as provider_phone,
				s.booking_id, b.client_id, uc.username as client_username,
				s.start_time, s.end_time, s.status,
				s.location_type, s.location_address, s.location_province, s.location_district,
				s.latitude, s.longitude, s.notes, s.created_at, s.updated_at
			FROM provider_schedules s
			JOIN users u ON s.provider_id = u.user_id
			LEFT JOIN bookings b ON s.booking_id = b.booking_id
			LEFT JOIN users uc ON b.client_id = uc.user_id
			WHERE s.is_visible_to_admin = true
		`
		args := []interface{}{}
		argCounter := 1

		if startDate != "" {
			query += " AND s.start_time >= $" + strconv.Itoa(argCounter)
			args = append(args, startDate)
			argCounter++
		}
		if endDate != "" {
			query += " AND s.end_time <= $" + strconv.Itoa(argCounter)
			args = append(args, endDate)
			argCounter++
		}
		if status != "" {
			query += " AND s.status = $" + strconv.Itoa(argCounter)
			args = append(args, status)
			argCounter++
		}
		if province != "" {
			query += " AND s.location_province = $" + strconv.Itoa(argCounter)
			args = append(args, province)
		}

		query += " ORDER BY s.start_time ASC"

		rows, err := dbPool.Query(ctx, query, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch schedules", "details": err.Error()})
			return
		}
		defer rows.Close()

		var schedules []ProviderScheduleWithDetails
		for rows.Next() {
			var s ProviderScheduleWithDetails
			err := rows.Scan(
				&s.ScheduleID, &s.ProviderID, &s.ProviderUsername, &s.ProviderPhone,
				&s.BookingID, &s.ClientID, &s.ClientUsername,
				&s.StartTime, &s.EndTime, &s.Status,
				&s.LocationType, &s.LocationAddress, &s.LocationProvince, &s.LocationDistrict,
				&s.Latitude, &s.Longitude, &s.Notes, &s.CreatedAt, &s.UpdatedAt,
			)
			if err != nil {
				continue
			}
			schedules = append(schedules, s)
		}

		c.JSON(http.StatusOK, gin.H{
			"schedules": schedules,
			"total":     len(schedules),
		})
	}
}
