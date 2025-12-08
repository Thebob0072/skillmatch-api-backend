package main

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ================================
// Deposit Handlers
// ================================

// GET /provider/deposit-settings
func getDepositSettingsHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")

		var requireDeposit bool
		var depositPercentage float64

		err := dbPool.QueryRow(ctx, `
			SELECT require_deposit, deposit_percentage 
			FROM provider_deposit_settings 
			WHERE provider_id = $1
		`, userID).Scan(&requireDeposit, &depositPercentage)

		if err != nil {
			// Return defaults
			c.JSON(http.StatusOK, gin.H{
				"require_deposit":    false,
				"deposit_percentage": 0.30,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"require_deposit":    requireDeposit,
			"deposit_percentage": depositPercentage,
		})
	}
}

// PUT /provider/deposit-settings
func updateDepositSettingsHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")

		var input DepositSettingsRequest
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate percentage (10-50%)
		if input.DepositPercentage < 0.10 || input.DepositPercentage > 0.50 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Deposit percentage must be between 10% and 50%"})
			return
		}

		_, err := dbPool.Exec(ctx, `
			INSERT INTO provider_deposit_settings (provider_id, require_deposit, deposit_percentage, updated_at)
			VALUES ($1, $2, $3, NOW())
			ON CONFLICT (provider_id) DO UPDATE SET
				require_deposit = $2, deposit_percentage = $3, updated_at = NOW()
		`, userID, input.RequireDeposit, input.DepositPercentage)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Deposit settings updated"})
	}
}

// POST /bookings/:id/deposit/pay - Pay deposit for booking
func payDepositHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")
		bookingID, _ := strconv.Atoi(c.Param("id"))

		// Verify booking and get deposit info
		var clientID, providerID int
		var totalPrice, depositPercentage float64
		err := dbPool.QueryRow(ctx, `
			SELECT b.client_id, b.provider_id, b.total_price, 
				   COALESCE(pds.deposit_percentage, 0.30) as deposit_pct
			FROM bookings b
			LEFT JOIN provider_deposit_settings pds ON b.provider_id = pds.provider_id
			WHERE b.booking_id = $1 AND b.status = 'pending'
		`, bookingID).Scan(&clientID, &providerID, &totalPrice, &depositPercentage)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
			return
		}

		if userID != clientID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only client can pay deposit"})
			return
		}

		depositAmount := totalPrice * depositPercentage

		// Check if deposit already exists
		var existingID int
		err = dbPool.QueryRow(ctx, `SELECT deposit_id FROM booking_deposits WHERE booking_id = $1`, bookingID).Scan(&existingID)
		if err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Deposit already paid"})
			return
		}

		// TODO: Process payment via Stripe
		// For now, just create deposit record

		var depositID int
		err = dbPool.QueryRow(ctx, `
			INSERT INTO booking_deposits (booking_id, client_id, provider_id, amount, percentage, status, paid_at)
			VALUES ($1, $2, $3, $4, $5, 'paid', NOW())
			RETURNING deposit_id
		`, bookingID, clientID, providerID, depositAmount, depositPercentage).Scan(&depositID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process deposit"})
			return
		}

		// Update booking status
		dbPool.Exec(ctx, `UPDATE bookings SET status = 'deposit_paid' WHERE booking_id = $1`, bookingID)

		// Notify provider
		CreateNotification(providerID, "deposit_paid", "Client has paid the deposit", map[string]interface{}{
			"booking_id": bookingID,
			"amount":     depositAmount,
		})

		c.JSON(http.StatusCreated, gin.H{
			"deposit_id": depositID,
			"amount":     depositAmount,
			"percentage": depositPercentage,
			"message":    "Deposit paid successfully",
			"remaining":  totalPrice - depositAmount,
		})
	}
}

// ================================
// Cancellation Fee Handlers
// ================================

// GET /provider/cancellation-policy
func getCancellationPolicyHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		providerID, _ := strconv.Atoi(c.Param("providerId"))
		if providerID == 0 {
			uid, _ := c.Get("userID")
			providerID = uid.(int)
		}

		rows, err := dbPool.Query(ctx, `
			SELECT policy_id, hours_before_booking, fee_percentage
			FROM cancellation_policies
			WHERE provider_id = $1 AND is_active = true
			ORDER BY hours_before_booking DESC
		`, providerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch policies"})
			return
		}
		defer rows.Close()

		policies := make([]gin.H, 0)
		for rows.Next() {
			var policyID, hours int
			var fee float64
			rows.Scan(&policyID, &hours, &fee)
			policies = append(policies, gin.H{
				"policy_id":            policyID,
				"hours_before_booking": hours,
				"fee_percentage":       fee,
			})
		}

		c.JSON(http.StatusOK, policies)
	}
}

// PUT /provider/cancellation-policy
func updateCancellationPolicyHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")

		var input CancellationPolicyRequest
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Deactivate old policies
		dbPool.Exec(ctx, `UPDATE cancellation_policies SET is_active = false WHERE provider_id = $1`, userID)

		// Insert new policies
		for _, policy := range input.Policies {
			if policy.FeePercentage < 0 || policy.FeePercentage > 1 {
				continue
			}
			dbPool.Exec(ctx, `
				INSERT INTO cancellation_policies (provider_id, hours_before_booking, fee_percentage)
				VALUES ($1, $2, $3)
			`, userID, policy.HoursBeforeBooking, policy.FeePercentage)
		}

		c.JSON(http.StatusOK, gin.H{"message": "Cancellation policy updated"})
	}
}

// POST /bookings/:id/cancel - Cancel booking with fee calculation
func cancelBookingWithFeeHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")
		bookingID, _ := strconv.Atoi(c.Param("id"))

		var input struct {
			Reason string `json:"reason"`
		}
		c.ShouldBindJSON(&input)

		// Get booking details
		var clientID, providerID int
		var totalPrice float64
		var startTime time.Time
		err := dbPool.QueryRow(ctx, `
			SELECT client_id, provider_id, total_price, start_time
			FROM bookings
			WHERE booking_id = $1 AND status IN ('pending', 'confirmed', 'deposit_paid')
		`, bookingID).Scan(&clientID, &providerID, &totalPrice, &startTime)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found or cannot be cancelled"})
			return
		}

		if userID != clientID && userID != providerID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
			return
		}

		// Calculate hours before booking
		hoursUntilBooking := time.Until(startTime).Hours()

		// Get applicable fee
		var feePercentage float64 = 0
		err = dbPool.QueryRow(ctx, `
			SELECT fee_percentage FROM cancellation_policies
			WHERE provider_id = $1 AND is_active = true AND hours_before_booking >= $2
			ORDER BY hours_before_booking ASC
			LIMIT 1
		`, providerID, hoursUntilBooking).Scan(&feePercentage)

		feeAmount := totalPrice * feePercentage

		// Create cancellation fee record (only if client cancels)
		if userID == clientID && feeAmount > 0 {
			dbPool.Exec(ctx, `
				INSERT INTO cancellation_fees (booking_id, cancelled_by, fee_amount, fee_percentage, status)
				VALUES ($1, $2, $3, $4, 'pending')
			`, bookingID, userID, feeAmount, feePercentage)
		}

		// Handle deposit (refund or forfeit)
		if userID == clientID {
			// Client cancels - forfeit deposit to provider
			dbPool.Exec(ctx, `
				UPDATE booking_deposits SET status = 'forfeited', forfeited_at = NOW()
				WHERE booking_id = $1 AND status = 'paid'
			`, bookingID)
		} else {
			// Provider cancels - refund deposit to client
			dbPool.Exec(ctx, `
				UPDATE booking_deposits SET status = 'refunded', refunded_at = NOW()
				WHERE booking_id = $1 AND status = 'paid'
			`, bookingID)
		}

		// Update booking
		dbPool.Exec(ctx, `
			UPDATE bookings SET status = 'cancelled', cancelled_at = NOW(), cancellation_reason = $1
			WHERE booking_id = $2
		`, input.Reason, bookingID)

		// Notify other party
		notifyUserID := clientID
		if userID == clientID {
			notifyUserID = providerID
		}
		CreateNotification(notifyUserID, "booking_cancelled", "Booking has been cancelled", map[string]interface{}{
			"booking_id":       bookingID,
			"cancelled_by":     userID,
			"cancellation_fee": feeAmount,
		})

		c.JSON(http.StatusOK, gin.H{
			"message":             "Booking cancelled",
			"cancellation_fee":    feeAmount,
			"fee_percentage":      feePercentage,
			"hours_until_booking": hoursUntilBooking,
		})
	}
}

// ================================
// Featured/Boost Profile Handlers
// ================================

// GET /boost/packages
func getBoostPackagesHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := dbPool.Query(ctx, `
			SELECT package_id, name, boost_type, duration, price, description
			FROM boost_packages
			WHERE is_active = true
			ORDER BY price ASC
		`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch packages"})
			return
		}
		defer rows.Close()

		packages := make([]BoostPackage, 0)
		for rows.Next() {
			var pkg BoostPackage
			rows.Scan(&pkg.PackageID, &pkg.Name, &pkg.BoostType, &pkg.Duration, &pkg.Price, &pkg.Description)
			packages = append(packages, pkg)
		}

		c.JSON(http.StatusOK, packages)
	}
}

// POST /boost/purchase
func purchaseBoostHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")

		var input PurchaseBoostRequest
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get package details
		var pkg BoostPackage
		err := dbPool.QueryRow(ctx, `
			SELECT package_id, name, boost_type, duration, price
			FROM boost_packages
			WHERE package_id = $1 AND is_active = true
		`, input.PackageID).Scan(&pkg.PackageID, &pkg.Name, &pkg.BoostType, &pkg.Duration, &pkg.Price)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Package not found"})
			return
		}

		// Check if user already has active boost of same type
		var existingBoostID int
		err = dbPool.QueryRow(ctx, `
			SELECT boost_id FROM profile_boosts
			WHERE user_id = $1 AND boost_type = $2 AND status = 'active' AND end_time > NOW()
		`, userID, pkg.BoostType).Scan(&existingBoostID)

		if err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "You already have an active boost of this type"})
			return
		}

		// TODO: Process payment via Stripe

		startTime := time.Now()
		endTime := startTime.Add(time.Duration(pkg.Duration) * time.Hour)

		var boostID int
		err = dbPool.QueryRow(ctx, `
			INSERT INTO profile_boosts (user_id, boost_type, start_time, end_time, amount, status)
			VALUES ($1, $2, $3, $4, $5, 'active')
			RETURNING boost_id
		`, userID, pkg.BoostType, startTime, endTime, pkg.Price).Scan(&boostID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to activate boost"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"boost_id":   boostID,
			"boost_type": pkg.BoostType,
			"start_time": startTime,
			"end_time":   endTime,
			"price":      pkg.Price,
			"message":    "Boost activated successfully",
		})
	}
}

// GET /boost/active
func getActiveBoostsHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")

		rows, err := dbPool.Query(ctx, `
			SELECT boost_id, boost_type, start_time, end_time, amount, status
			FROM profile_boosts
			WHERE user_id = $1 AND status = 'active' AND end_time > NOW()
			ORDER BY end_time ASC
		`, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch boosts"})
			return
		}
		defer rows.Close()

		boosts := make([]gin.H, 0)
		for rows.Next() {
			var boostID int
			var boostType, status string
			var startTime, endTime time.Time
			var amount float64

			rows.Scan(&boostID, &boostType, &startTime, &endTime, &amount, &status)

			boosts = append(boosts, gin.H{
				"boost_id":        boostID,
				"boost_type":      boostType,
				"start_time":      startTime,
				"end_time":        endTime,
				"amount":          amount,
				"remaining_hours": time.Until(endTime).Hours(),
			})
		}

		c.JSON(http.StatusOK, boosts)
	}
}

// ================================
// Coupon Handlers
// ================================

// POST /coupons (Admin/Provider)
func createCouponHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")

		var input CreateCouponRequest
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Parse dates
		validFrom, _ := time.Parse("2006-01-02", input.ValidFrom)
		validUntil, _ := time.Parse("2006-01-02", input.ValidUntil)

		// Check if provider-specific or platform-wide
		var providerID *int
		isAdmin := false
		dbPool.QueryRow(ctx, `SELECT is_admin FROM users WHERE user_id = $1`, userID).Scan(&isAdmin)

		if !isAdmin {
			uid := userID.(int)
			providerID = &uid
		}

		var couponID int
		err := dbPool.QueryRow(ctx, `
			INSERT INTO coupons (code, discount_type, discount_value, min_booking_amount, max_discount, 
								 valid_from, valid_until, usage_limit, created_by, provider_id)
			VALUES (UPPER($1), $2, $3, $4, $5, $6, $7, $8, $9, $10)
			RETURNING coupon_id
		`, input.Code, input.DiscountType, input.DiscountValue, input.MinBookingAmount,
			input.MaxDiscount, validFrom, validUntil, input.UsageLimit, userID, providerID).Scan(&couponID)

		if err != nil {
			if strings.Contains(err.Error(), "duplicate") {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Coupon code already exists"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create coupon"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"coupon_id": couponID,
			"code":      strings.ToUpper(input.Code),
			"message":   "Coupon created successfully",
		})
	}
}

// POST /coupons/apply
func applyCouponHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")

		var input ApplyCouponRequest
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get coupon details
		var coupon Coupon
		err := dbPool.QueryRow(ctx, `
			SELECT coupon_id, code, discount_type, discount_value, min_booking_amount, max_discount, 
				   valid_from, valid_until, usage_limit, used_count, provider_id
			FROM coupons
			WHERE UPPER(code) = UPPER($1) AND is_active = true
		`, input.Code).Scan(&coupon.CouponID, &coupon.Code, &coupon.DiscountType, &coupon.DiscountValue,
			&coupon.MinBookingAmount, &coupon.MaxDiscount, &coupon.ValidFrom, &coupon.ValidUntil,
			&coupon.UsageLimit, &coupon.UsedCount, &coupon.ProviderID)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Coupon not found"})
			return
		}

		// Validate coupon
		now := time.Now()
		if now.Before(coupon.ValidFrom) || now.After(coupon.ValidUntil) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Coupon has expired or not yet valid"})
			return
		}

		if coupon.UsageLimit != nil && coupon.UsedCount >= *coupon.UsageLimit {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Coupon usage limit reached"})
			return
		}

		// Check if user already used this coupon
		var usageID int
		err = dbPool.QueryRow(ctx, `SELECT usage_id FROM coupon_usages WHERE coupon_id = $1 AND user_id = $2`,
			coupon.CouponID, userID).Scan(&usageID)
		if err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "You have already used this coupon"})
			return
		}

		// Get booking details
		var providerID int
		var totalPrice float64
		err = dbPool.QueryRow(ctx, `
			SELECT provider_id, total_price FROM bookings
			WHERE booking_id = $1 AND client_id = $2 AND status IN ('pending', 'deposit_paid')
		`, input.BookingID, userID).Scan(&providerID, &totalPrice)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
			return
		}

		// Check if coupon is for this provider
		if coupon.ProviderID != nil && *coupon.ProviderID != providerID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Coupon not valid for this provider"})
			return
		}

		// Check minimum booking amount
		if coupon.MinBookingAmount != nil && totalPrice < *coupon.MinBookingAmount {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Booking amount below minimum required"})
			return
		}

		// Calculate discount
		var discountAmount float64
		if coupon.DiscountType == "percentage" {
			discountAmount = totalPrice * (coupon.DiscountValue / 100)
		} else {
			discountAmount = coupon.DiscountValue
		}

		// Apply max discount cap
		if coupon.MaxDiscount != nil && discountAmount > *coupon.MaxDiscount {
			discountAmount = *coupon.MaxDiscount
		}

		newTotal := totalPrice - discountAmount

		// Record usage
		dbPool.Exec(ctx, `
			INSERT INTO coupon_usages (coupon_id, user_id, booking_id, discount_amount)
			VALUES ($1, $2, $3, $4)
		`, coupon.CouponID, userID, input.BookingID, discountAmount)

		// Update coupon used count
		dbPool.Exec(ctx, `UPDATE coupons SET used_count = used_count + 1 WHERE coupon_id = $1`, coupon.CouponID)

		// Update booking total
		dbPool.Exec(ctx, `UPDATE bookings SET total_price = $1 WHERE booking_id = $2`, newTotal, input.BookingID)

		c.JSON(http.StatusOK, gin.H{
			"coupon_code":     coupon.Code,
			"discount_type":   coupon.DiscountType,
			"discount_value":  coupon.DiscountValue,
			"discount_amount": discountAmount,
			"original_price":  totalPrice,
			"new_total":       newTotal,
			"message":         "Coupon applied successfully",
		})
	}
}

// GET /coupons (Provider's coupons)
func getProviderCouponsHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")

		rows, err := dbPool.Query(ctx, `
			SELECT coupon_id, code, discount_type, discount_value, min_booking_amount, max_discount,
				   valid_from, valid_until, usage_limit, used_count, is_active, created_at
			FROM coupons
			WHERE provider_id = $1 OR created_by = $1
			ORDER BY created_at DESC
		`, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch coupons"})
			return
		}
		defer rows.Close()

		coupons := make([]gin.H, 0)
		for rows.Next() {
			var couponID, usedCount int
			var code, discountType string
			var discountValue float64
			var minAmount, maxDiscount *float64
			var validFrom, validUntil, createdAt time.Time
			var usageLimit *int
			var isActive bool

			rows.Scan(&couponID, &code, &discountType, &discountValue, &minAmount, &maxDiscount,
				&validFrom, &validUntil, &usageLimit, &usedCount, &isActive, &createdAt)

			coupons = append(coupons, gin.H{
				"coupon_id":          couponID,
				"code":               code,
				"discount_type":      discountType,
				"discount_value":     discountValue,
				"min_booking_amount": minAmount,
				"max_discount":       maxDiscount,
				"valid_from":         validFrom,
				"valid_until":        validUntil,
				"usage_limit":        usageLimit,
				"used_count":         usedCount,
				"is_active":          isActive,
				"created_at":         createdAt,
			})
		}

		c.JSON(http.StatusOK, coupons)
	}
}

// ================================
// Verified Photo Badge Handlers
// ================================

// POST /photos/:id/verify - Submit photo for verification
func submitPhotoVerificationHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")
		photoID, _ := strconv.Atoi(c.Param("id"))

		// Verify photo belongs to user
		var ownerID int
		err := dbPool.QueryRow(ctx, `SELECT user_id FROM user_photos WHERE photo_id = $1`, photoID).Scan(&ownerID)
		if err != nil || ownerID != userID.(int) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Photo not found"})
			return
		}

		// Check if already submitted
		var existingID int
		err = dbPool.QueryRow(ctx, `
			SELECT verification_id FROM photo_verifications 
			WHERE photo_id = $1 AND status IN ('pending', 'verified')
		`, photoID).Scan(&existingID)
		if err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Photo already submitted for verification"})
			return
		}

		var verificationID int
		err = dbPool.QueryRow(ctx, `
			INSERT INTO photo_verifications (photo_id, user_id, status)
			VALUES ($1, $2, 'pending')
			RETURNING verification_id
		`, photoID, userID).Scan(&verificationID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit verification"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"verification_id": verificationID,
			"status":          "pending",
			"message":         "Photo submitted for verification",
		})
	}
}

// PATCH /admin/photos/:id/verify - Admin approve/reject verification
func adminVerifyPhotoHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		adminID, _ := c.Get("userID")
		photoID, _ := strconv.Atoi(c.Param("id"))

		var input struct {
			Action string  `json:"action" binding:"required"` // approve, reject
			Reason *string `json:"reason"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		status := "verified"
		if input.Action == "reject" {
			status = "rejected"
		}

		result, err := dbPool.Exec(ctx, `
			UPDATE photo_verifications 
			SET status = $1, verified_at = NOW(), verified_by = $2, rejection_reason = $3
			WHERE photo_id = $4 AND status = 'pending'
		`, status, adminID, input.Reason, photoID)

		if err != nil || result.RowsAffected() == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Verification request not found"})
			return
		}

		// If approved, add verified badge to photo
		if status == "verified" {
			dbPool.Exec(ctx, `
				UPDATE user_photos SET is_verified = true WHERE photo_id = $1
			`, photoID)
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Photo verification " + input.Action + "d",
			"status":  status,
		})
	}
}

// GET /admin/photos/pending - Get pending verifications
func getPendingPhotoVerificationsHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := dbPool.Query(ctx, `
			SELECT pv.verification_id, pv.photo_id, pv.user_id, u.username, up.photo_url, pv.created_at
			FROM photo_verifications pv
			JOIN users u ON pv.user_id = u.user_id
			JOIN user_photos up ON pv.photo_id = up.photo_id
			WHERE pv.status = 'pending'
			ORDER BY pv.created_at ASC
		`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch verifications"})
			return
		}
		defer rows.Close()

		verifications := make([]gin.H, 0)
		for rows.Next() {
			var verificationID, photoID, userID int
			var username, photoURL string
			var createdAt time.Time

			rows.Scan(&verificationID, &photoID, &userID, &username, &photoURL, &createdAt)

			verifications = append(verifications, gin.H{
				"verification_id": verificationID,
				"photo_id":        photoID,
				"user_id":         userID,
				"username":        username,
				"photo_url":       photoURL,
				"created_at":      createdAt,
			})
		}

		c.JSON(http.StatusOK, verifications)
	}
}
