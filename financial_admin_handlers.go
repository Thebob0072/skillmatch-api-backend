package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ================================
// Admin: Withdrawal Management
// ================================

// Get All Pending Withdrawals (Admin)
func adminGetPendingWithdrawalsHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := c.DefaultQuery("status", "pending")

		rows, err := dbPool.Query(ctx, `
			SELECT w.withdrawal_id, w.withdrawal_uuid, w.user_id, w.requested_amount,
			       w.fee, w.net_amount, w.status, w.requested_at,
			       u.username, u.email,
			       ba.bank_name, ba.account_number, ba.account_name, ba.branch_name
			FROM withdrawals w
			JOIN users u ON w.user_id = u.user_id
			JOIN bank_accounts ba ON w.bank_account_id = ba.bank_account_id
			WHERE w.status = $1
			ORDER BY w.requested_at ASC
		`, status)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch withdrawals"})
			return
		}
		defer rows.Close()

		withdrawals := make([]map[string]interface{}, 0)
		for rows.Next() {
			var (
				id, uuid, username, email            string
				userID                               int
				requested, fee, net                  float64
				statusStr                            string
				requestedAt                          time.Time
				bankName, accountNumber, accountName string
				branchName                           sql.NullString
			)

			rows.Scan(&id, &uuid, &userID, &requested, &fee, &net, &statusStr, &requestedAt,
				&username, &email, &bankName, &accountNumber, &accountName, &branchName)

			withdrawal := map[string]interface{}{
				"withdrawal_id":    id,
				"withdrawal_uuid":  uuid,
				"user_id":          userID,
				"username":         username,
				"email":            email,
				"requested_amount": requested,
				"fee":              fee,
				"net_amount":       net,
				"status":           statusStr,
				"requested_at":     requestedAt,
				"bank_account": map[string]interface{}{
					"bank_name":      bankName,
					"account_number": accountNumber,
					"account_name":   accountName,
					"branch_name":    branchName.String,
				},
			}

			withdrawals = append(withdrawals, withdrawal)
		}

		c.JSON(http.StatusOK, gin.H{
			"withdrawals": withdrawals,
			"total":       len(withdrawals),
			"status":      status,
		})
	}
}

// Process Withdrawal (Approve/Reject/Complete)
func adminProcessWithdrawalHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		adminID, _ := c.Get("userID")
		withdrawalID := c.Param("withdrawal_id")

		var req struct {
			Action            string  `json:"action" binding:"required,oneof=approve reject complete"`
			TransferReference *string `json:"transfer_reference"`
			TransferSlipURL   *string `json:"transfer_slip_url"`
			RejectionReason   *string `json:"rejection_reason"`
			Notes             *string `json:"notes"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Start transaction
		tx, err := dbPool.Begin(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
			return
		}
		defer tx.Rollback(ctx)

		// Get withdrawal details
		var (
			userID          int
			requestedAmount float64
			currentStatus   string
		)
		err = tx.QueryRow(ctx, `
			SELECT user_id, requested_amount, status FROM withdrawals WHERE withdrawal_id = $1
		`, withdrawalID).Scan(&userID, &requestedAmount, &currentStatus)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Withdrawal not found"})
			return
		}

		now := time.Now()

		switch req.Action {
		case "approve":
			if currentStatus != "pending" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Can only approve pending withdrawals"})
				return
			}

			// Update withdrawal status
			_, err = tx.Exec(ctx, `
			UPDATE withdrawals
			SET status = 'approved',
			    approved_at = $1,
			    approved_by = $2,
			    notes = $3
			WHERE withdrawal_id = $4
		`, now, adminID, req.Notes, withdrawalID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve withdrawal", "details": err.Error()})
				return
			}

			// Update transaction status
			_, err = tx.Exec(ctx, `
			UPDATE transactions
			SET status = 'processing'
			WHERE withdrawal_id = $1 AND type = 'withdrawal'
		`, withdrawalID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update transaction", "details": err.Error()})
				return
			}

		case "reject":
			if currentStatus != "pending" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Can only reject pending withdrawals"})
				return
			}

			// Update withdrawal status
			_, err = tx.Exec(ctx, `
			UPDATE withdrawals
			SET status = 'rejected',
			    rejected_at = $1,
			    rejected_by = $2,
			    rejection_reason = $3,
			    notes = $4
			WHERE withdrawal_id = $5
		`, now, adminID, req.RejectionReason, req.Notes, withdrawalID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject withdrawal", "details": err.Error()})
				return
			}

			// Refund to wallet
			_, err = tx.Exec(ctx, `
			UPDATE wallets
			SET available_balance = available_balance + $1,
			    last_updated = CURRENT_TIMESTAMP
			WHERE user_id = $2
		`, requestedAmount, userID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refund wallet", "details": err.Error()})
				return
			}

			// Update transaction status
			_, err = tx.Exec(ctx, `
			UPDATE transactions
			SET status = 'cancelled'
			WHERE withdrawal_id = $1 AND type = 'withdrawal'
		`, withdrawalID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update transaction", "details": err.Error()})
				return
			}

		case "complete":
			if currentStatus != "approved" && currentStatus != "processing" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Can only complete approved withdrawals"})
				return
			}

			// Update withdrawal status
			_, err = tx.Exec(ctx, `
			UPDATE withdrawals
			SET status = 'completed',
			    completed_at = $1,
			    transfer_reference = $2,
			    transfer_slip_url = $3,
			    notes = $4
			WHERE withdrawal_id = $5
		`, now, req.TransferReference, req.TransferSlipURL, req.Notes, withdrawalID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete withdrawal", "details": err.Error()})
				return
			}

			// Update wallet
			_, err = tx.Exec(ctx, `
			UPDATE wallets
			SET total_withdrawn = total_withdrawn + $1,
			    last_updated = CURRENT_TIMESTAMP
			WHERE user_id = $2
		`, requestedAmount, userID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update wallet", "details": err.Error()})
				return
			}

			// Update transaction status
			_, err = tx.Exec(ctx, `
			UPDATE transactions
			SET status = 'completed', processed_at = $1
			WHERE withdrawal_id = $2 AND type = 'withdrawal'
		`, now, withdrawalID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update transaction", "details": err.Error()})
				return
			}
		}

		// Commit transaction
		if err = tx.Commit(ctx); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":       "Withdrawal processed successfully",
			"withdrawal_id": withdrawalID,
			"action":        req.Action,
		})
	}
}

// ================================
// Admin: Bank Account Verification
// ================================

func adminVerifyBankAccountHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		adminID, _ := c.Get("userID")
		bankAccountID := c.Param("bank_account_id")

		var req struct {
			Verified bool    `json:"verified" binding:"required"`
			Notes    *string `json:"notes"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		now := time.Now()
		_, err := dbPool.Exec(ctx, `
			UPDATE bank_accounts
			SET is_verified = $1,
			    verified_at = $2,
			    verified_by = $3
			WHERE bank_account_id = $4
		`, req.Verified, now, adminID, bankAccountID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify bank account"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":         "Bank account verification updated",
			"bank_account_id": bankAccountID,
			"verified":        req.Verified,
		})
	}
}

// ================================
// Admin: Financial Reports
// ================================

func adminGetFinancialSummaryHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		today := time.Now().Format("2006-01-02")
		thisMonth := time.Now().Format("2006-01")

		var summary FinancialSummary

		// Today's revenue & commission
		dbPool.QueryRow(ctx, `
			SELECT COALESCE(SUM(amount), 0), COALESCE(SUM(commission_amount), 0)
			FROM transactions
			WHERE type = 'booking_payment' 
			  AND status = 'completed'
			  AND DATE(created_at) = $1
		`, today).Scan(&summary.TodayRevenue, &summary.TodayCommission)

		// This month's revenue & commission
		dbPool.QueryRow(ctx, `
			SELECT COALESCE(SUM(amount), 0), COALESCE(SUM(commission_amount), 0)
			FROM transactions
			WHERE type = 'booking_payment' 
			  AND status = 'completed'
			  AND TO_CHAR(created_at, 'YYYY-MM') = $1
		`, thisMonth).Scan(&summary.MonthRevenue, &summary.MonthCommission)

		// Pending withdrawals count
		dbPool.QueryRow(ctx, `
			SELECT COUNT(*) FROM withdrawals WHERE status = 'pending'
		`).Scan(&summary.PendingWithdrawals)

		// Total users
		dbPool.QueryRow(ctx, `
			SELECT COUNT(*) FROM users
		`).Scan(&summary.TotalUsers)

		c.JSON(http.StatusOK, summary)
	}
}

func adminGenerateFinancialReportHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		adminID, _ := c.Get("userID")

		var req struct {
			ReportType  string    `json:"report_type" binding:"required,oneof=daily weekly monthly yearly"`
			PeriodStart time.Time `json:"period_start" binding:"required"`
			PeriodEnd   time.Time `json:"period_end" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Calculate metrics
		var (
			totalBookings, totalSubscriptions                                      int
			totalRevenue, totalCommission, totalProviderEarnings, totalWithdrawals float64
		)

		// Total bookings & revenue
		dbPool.QueryRow(ctx, `
			SELECT COUNT(*), COALESCE(SUM(amount), 0), COALESCE(SUM(commission_amount), 0)
			FROM transactions
			WHERE type = 'booking_payment'
			  AND status = 'completed'
			  AND created_at BETWEEN $1 AND $2
		`, req.PeriodStart, req.PeriodEnd).Scan(&totalBookings, &totalRevenue, &totalCommission)

		// Provider earnings
		dbPool.QueryRow(ctx, `
			SELECT COALESCE(SUM(net_amount), 0)
			FROM transactions
			WHERE type = 'provider_earning'
			  AND status = 'completed'
			  AND created_at BETWEEN $1 AND $2
		`, req.PeriodStart, req.PeriodEnd).Scan(&totalProviderEarnings)

		// Withdrawals
		dbPool.QueryRow(ctx, `
			SELECT COALESCE(SUM(net_amount), 0)
			FROM withdrawals
			WHERE status = 'completed'
			  AND completed_at BETWEEN $1 AND $2
		`, req.PeriodStart, req.PeriodEnd).Scan(&totalWithdrawals)

		// Subscription fees
		dbPool.QueryRow(ctx, `
			SELECT COUNT(*)
			FROM transactions
			WHERE type = 'subscription_fee'
			  AND status = 'completed'
			  AND created_at BETWEEN $1 AND $2
		`, req.PeriodStart, req.PeriodEnd).Scan(&totalSubscriptions)

		// Insert report
		var reportID int
		err := dbPool.QueryRow(ctx, `
			INSERT INTO financial_reports (
				report_type, period_start, period_end,
				total_bookings, total_revenue, total_commission,
				total_provider_earnings, total_withdrawals, total_subscriptions,
				generated_by
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			RETURNING report_id
		`, req.ReportType, req.PeriodStart, req.PeriodEnd,
			totalBookings, totalRevenue, totalCommission,
			totalProviderEarnings, totalWithdrawals, totalSubscriptions,
			adminID).Scan(&reportID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate report"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":   "Financial report generated successfully",
			"report_id": reportID,
			"summary": gin.H{
				"total_bookings":          totalBookings,
				"total_revenue":           totalRevenue,
				"total_commission":        totalCommission,
				"total_provider_earnings": totalProviderEarnings,
				"total_withdrawals":       totalWithdrawals,
				"total_subscriptions":     totalSubscriptions,
			},
		})
	}
}

// ================================
// Admin: Commission Rules
// ================================

func adminGetCommissionRulesHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := dbPool.Query(ctx, `
			SELECT rule_id, name, description, platform_rate, payment_gateway_rate,
			       tier_id, effective_from, effective_until, is_active,
			       created_at, updated_at
			FROM commission_rules
			ORDER BY effective_from DESC
		`)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch commission rules"})
			return
		}
		defer rows.Close()

		rules := make([]CommissionRule, 0)
		for rows.Next() {
			var rule CommissionRule
			rows.Scan(&rule.RuleID, &rule.Name, &rule.Description, &rule.PlatformRate,
				&rule.PaymentGatewayRate, &rule.TierID, &rule.EffectiveFrom,
				&rule.EffectiveUntil, &rule.IsActive, &rule.CreatedAt, &rule.UpdatedAt)
			rules = append(rules, rule)
		}

		c.JSON(http.StatusOK, gin.H{
			"commission_rules": rules,
			"total":            len(rules),
		})
	}
}

func adminUpdateCommissionRuleHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		ruleID := c.Param("rule_id")

		var req struct {
			PlatformRate       float64    `json:"platform_rate" binding:"required,min=0,max=1"`
			PaymentGatewayRate float64    `json:"payment_gateway_rate" binding:"required,min=0,max=1"`
			EffectiveFrom      *time.Time `json:"effective_from"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, err := dbPool.Exec(ctx, `
			UPDATE commission_rules
			SET platform_rate = $1,
			    payment_gateway_rate = $2,
			    effective_from = COALESCE($3, effective_from),
			    updated_at = CURRENT_TIMESTAMP
			WHERE rule_id = $4
		`, req.PlatformRate, req.PaymentGatewayRate, req.EffectiveFrom, ruleID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update commission rule"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Commission rule updated successfully",
			"rule_id": ruleID,
		})
	}
}

// ================================
// Admin: User Wallet Management
// ================================

func adminGetUserWalletHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.Param("user_id")
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		var wallet Wallet
		err = dbPool.QueryRow(ctx, `
			SELECT wallet_id, user_id, available_balance, pending_balance,
			       total_earned, total_withdrawn, total_commission_paid,
			       last_updated, created_at
			FROM wallets
			WHERE user_id = $1
		`, userID).Scan(
			&wallet.WalletID, &wallet.UserID, &wallet.AvailableBalance,
			&wallet.PendingBalance, &wallet.TotalEarned, &wallet.TotalWithdrawn,
			&wallet.TotalCommissionPaid, &wallet.LastUpdated, &wallet.CreatedAt,
		)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Wallet not found"})
			return
		}

		c.JSON(http.StatusOK, wallet)
	}
}

func adminAdjustWalletHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		adminID, _ := c.Get("userID")
		userIDStr := c.Param("user_id")
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		var req struct {
			Amount      float64 `json:"amount" binding:"required"`
			Type        string  `json:"type" binding:"required,oneof=bonus penalty adjustment"`
			Description string  `json:"description" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Start transaction
		tx, err := dbPool.Begin(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
			return
		}
		defer tx.Rollback(ctx)

		// Update wallet
		_, err = tx.Exec(ctx, `
			UPDATE wallets
			SET available_balance = available_balance + $1,
			    last_updated = CURRENT_TIMESTAMP
			WHERE user_id = $2
		`, req.Amount, userID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update wallet"})
			return
		}

		// Create transaction record
		_, err = tx.Exec(ctx, `
			INSERT INTO transactions (
				user_id, type, status, amount, commission_amount, net_amount,
				description, notes, related_user_id
			) VALUES ($1, $2, 'completed', $3, 0, $3, $4, $5, $6)
		`, userID, req.Type, req.Amount, req.Description,
			fmt.Sprintf("Admin adjustment by user %d", adminID), adminID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transaction"})
			return
		}

		if err = tx.Commit(ctx); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Wallet adjusted successfully",
			"user_id": userID,
			"amount":  req.Amount,
			"type":    req.Type,
		})
	}
}
