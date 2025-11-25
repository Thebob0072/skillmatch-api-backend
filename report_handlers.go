package main

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Report represents a user report
type Report struct {
	ID                int        `json:"id"`
	ReporterID        int        `json:"reporter_id"`
	ReportedUserID    *int       `json:"reported_user_id,omitempty"`
	ReportType        string     `json:"report_type"`
	Reason            string     `json:"reason"`
	AdditionalDetails *string    `json:"additional_details,omitempty"`
	Status            string     `json:"status"`
	AdminNotes        *string    `json:"admin_notes,omitempty"`
	ResolvedBy        *int       `json:"resolved_by,omitempty"`
	ResolvedAt        *time.Time `json:"resolved_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`

	// Additional fields for API responses
	ReporterName     string `json:"reporter_name,omitempty"`
	ReportedUserName string `json:"reported_user_name,omitempty"`
	ResolvedByName   string `json:"resolved_by_name,omitempty"`
}

// CreateReportRequest is the payload for creating a new report
type CreateReportRequest struct {
	ReportedUserID    int     `json:"reported_user_id" binding:"required"`
	ReportType        string  `json:"report_type" binding:"required"`
	Reason            string  `json:"reason" binding:"required,min=10,max=1000"`
	AdditionalDetails *string `json:"additional_details"`
}

// UpdateReportStatusRequest is the payload for updating a report status (admin only)
type UpdateReportStatusRequest struct {
	Status     string  `json:"status" binding:"required"`
	AdminNotes *string `json:"admin_notes"`
}

// ReportListResponse wraps the report list with metadata
type ReportListResponse struct {
	Reports      []Report `json:"reports"`
	Total        int      `json:"total"`
	PendingCount int      `json:"pending_count"`
}

// CreateReport creates a new report
// POST /reports
func CreateReport(c *gin.Context) {
	userID := c.GetInt("userID")

	var req CreateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate report type
	validTypes := map[string]bool{
		"harassment":            true,
		"inappropriate_content": true,
		"fake_profile":          true,
		"scam":                  true,
		"violence_threat":       true,
		"underage":              true,
		"spam":                  true,
		"other":                 true,
	}
	if !validTypes[req.ReportType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report type"})
		return
	}

	// Cannot report yourself
	if userID == req.ReportedUserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot report yourself"})
		return
	}

	// Check if reported user exists
	var userExists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", req.ReportedUserID).Scan(&userExists)
	if err != nil || !userExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Reported user not found"})
		return
	}

	// Check if user already reported this user for the same reason (prevent spam)
	var existingCount int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM reports 
		WHERE reporter_id = $1 
		AND reported_user_id = $2 
		AND report_type = $3 
		AND created_at > NOW() - INTERVAL '24 hours'
	`, userID, req.ReportedUserID, req.ReportType).Scan(&existingCount)

	if err == nil && existingCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You have already reported this user for this reason recently"})
		return
	}

	// Create report
	var reportID int
	err = db.QueryRow(`
		INSERT INTO reports (reporter_id, reported_user_id, report_type, reason, additional_details)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, userID, req.ReportedUserID, req.ReportType, req.Reason, req.AdditionalDetails).Scan(&reportID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create report"})
		return
	}

	// TODO: Notify admins about new report
	// CreateNotification for admin users

	c.JSON(http.StatusCreated, gin.H{
		"report_id": reportID,
		"message":   "Report submitted successfully. Our team will review it shortly.",
	})
}

// GetMyReports returns all reports submitted by the current user
// GET /reports/my
func GetMyReports(c *gin.Context) {
	userID := c.GetInt("userID")

	query := `
		SELECT 
			r.id,
			r.reporter_id,
			r.reported_user_id,
			r.report_type,
			r.reason,
			r.additional_details,
			r.status,
			r.created_at,
			r.updated_at,
			COALESCE(u.display_name, u.email) as reported_user_name
		FROM reports r
		LEFT JOIN users u ON r.reported_user_id = u.id
		WHERE r.reporter_id = $1
		ORDER BY r.created_at DESC
	`

	rows, err := db.Query(query, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reports"})
		return
	}
	defer rows.Close()

	reports := []Report{}
	for rows.Next() {
		var report Report
		var reportedUserName sql.NullString

		err := rows.Scan(
			&report.ID,
			&report.ReporterID,
			&report.ReportedUserID,
			&report.ReportType,
			&report.Reason,
			&report.AdditionalDetails,
			&report.Status,
			&report.CreatedAt,
			&report.UpdatedAt,
			&reportedUserName,
		)
		if err != nil {
			continue
		}

		if reportedUserName.Valid {
			report.ReportedUserName = reportedUserName.String
		}

		reports = append(reports, report)
	}

	c.JSON(http.StatusOK, gin.H{
		"reports": reports,
		"total":   len(reports),
	})
}

// GetAllReports returns all reports (admin only)
// GET /admin/reports
func GetAllReports(c *gin.Context) {
	// Filter by status
	status := c.Query("status")

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

	// Build query
	query := `
		SELECT 
			r.id,
			r.reporter_id,
			r.reported_user_id,
			r.report_type,
			r.reason,
			r.additional_details,
			r.status,
			r.admin_notes,
			r.resolved_by,
			r.resolved_at,
			r.created_at,
			r.updated_at,
			COALESCE(u1.display_name, u1.email) as reporter_name,
			COALESCE(u2.display_name, u2.email) as reported_user_name,
			COALESCE(u3.display_name, u3.email) as resolved_by_name
		FROM reports r
		LEFT JOIN users u1 ON r.reporter_id = u1.id
		LEFT JOIN users u2 ON r.reported_user_id = u2.id
		LEFT JOIN users u3 ON r.resolved_by = u3.id
	`

	args := []interface{}{}
	if status != "" {
		query += " WHERE r.status = $1"
		args = append(args, status)
		query += " ORDER BY r.created_at DESC LIMIT $2 OFFSET $3"
		args = append(args, limit, offset)
	} else {
		query += " ORDER BY r.created_at DESC LIMIT $1 OFFSET $2"
		args = append(args, limit, offset)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reports"})
		return
	}
	defer rows.Close()

	reports := []Report{}
	for rows.Next() {
		var report Report
		var reporterName, reportedUserName, resolvedByName sql.NullString

		err := rows.Scan(
			&report.ID,
			&report.ReporterID,
			&report.ReportedUserID,
			&report.ReportType,
			&report.Reason,
			&report.AdditionalDetails,
			&report.Status,
			&report.AdminNotes,
			&report.ResolvedBy,
			&report.ResolvedAt,
			&report.CreatedAt,
			&report.UpdatedAt,
			&reporterName,
			&reportedUserName,
			&resolvedByName,
		)
		if err != nil {
			continue
		}

		if reporterName.Valid {
			report.ReporterName = reporterName.String
		}
		if reportedUserName.Valid {
			report.ReportedUserName = reportedUserName.String
		}
		if resolvedByName.Valid {
			report.ResolvedByName = resolvedByName.String
		}

		reports = append(reports, report)
	}

	// Get total and pending count
	var total, pendingCount int
	if status != "" {
		db.QueryRow("SELECT COUNT(*) FROM reports WHERE status = $1", status).Scan(&total)
	} else {
		db.QueryRow("SELECT COUNT(*) FROM reports").Scan(&total)
	}
	db.QueryRow("SELECT COUNT(*) FROM reports WHERE status = 'pending'").Scan(&pendingCount)

	c.JSON(http.StatusOK, ReportListResponse{
		Reports:      reports,
		Total:        total,
		PendingCount: pendingCount,
	})
}

// UpdateReportStatus updates the status of a report (admin only)
// PATCH /admin/reports/:id
func UpdateReportStatus(c *gin.Context) {
	adminID := c.GetInt("userID")
	reportID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report ID"})
		return
	}

	var req UpdateReportStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate status
	validStatuses := map[string]bool{
		"pending":   true,
		"reviewing": true,
		"resolved":  true,
		"dismissed": true,
	}
	if !validStatuses[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
		return
	}

	// Update report
	query := `
		UPDATE reports
		SET status = $1, admin_notes = $2, updated_at = CURRENT_TIMESTAMP
	`
	args := []interface{}{req.Status, req.AdminNotes}

	if req.Status == "resolved" || req.Status == "dismissed" {
		query += ", resolved_by = $3, resolved_at = CURRENT_TIMESTAMP"
		args = append(args, adminID)
		query += " WHERE id = $4"
		args = append(args, reportID)
	} else {
		query += " WHERE id = $3"
		args = append(args, reportID)
	}

	result, err := db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update report"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Report not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Report status updated successfully"})
}

// DeleteReport deletes a report (admin only)
// DELETE /admin/reports/:id
func DeleteReport(c *gin.Context) {
	reportID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report ID"})
		return
	}

	result, err := db.Exec("DELETE FROM reports WHERE id = $1", reportID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete report"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Report not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Report deleted successfully"})
}
