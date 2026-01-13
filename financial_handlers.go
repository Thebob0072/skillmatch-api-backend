package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ================================
// Bank Account Handlers
// ================================

// Add Bank Account
func addBankAccountHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")

		var req AddBankAccountRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Check if user already has this account number
		var exists bool
		err := dbPool.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM bank_accounts 
				WHERE user_id = $1 AND account_number = $2
			)
		`, userID, req.AccountNumber).Scan(&exists)

		if err == nil && exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "This account number is already added"})
			return
		}

		// If this is set as default, unset other defaults
		if req.IsDefault {
			_, _ = dbPool.Exec(ctx, `
				UPDATE bank_accounts SET is_default = false WHERE user_id = $1
			`, userID)
		}

		// Insert bank account (using actual table columns)
		var bankAccountID int
		err = dbPool.QueryRow(ctx, `
			INSERT INTO bank_accounts (
				user_id, bank_name, account_number, account_holder_name, is_default
			) VALUES ($1, $2, $3, $4, $5)
			RETURNING bank_account_id
		`, userID, req.BankName, req.AccountNumber, req.AccountName, req.IsDefault).Scan(&bankAccountID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add bank account"})
			return
		}

		// Create wallet if doesn't exist
		_, _ = dbPool.Exec(ctx, `
			INSERT INTO wallets (user_id) VALUES ($1)
			ON CONFLICT (user_id) DO NOTHING
		`, userID)

		c.JSON(http.StatusCreated, gin.H{
			"message":         "Bank account added successfully",
			"bank_account_id": bankAccountID,
		})
	}
}

// Get My Bank Accounts
func getMyBankAccountsHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")

		rows, err := dbPool.Query(ctx, `
			SELECT bank_account_id, bank_name, account_number, account_holder_name,
			       is_default, created_at
			FROM bank_accounts
			WHERE user_id = $1
			ORDER BY is_default DESC, created_at DESC
		`, userID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch bank accounts"})
			return
		}
		defer rows.Close()

		accounts := make([]BankAccount, 0)
		for rows.Next() {
			var acc BankAccount
			err := rows.Scan(
				&acc.BankAccountID, &acc.BankName, &acc.AccountNumber,
				&acc.AccountName, &acc.IsDefault, &acc.CreatedAt,
			)
			if err != nil {
				continue
			}
			acc.UserID = userID.(int)
			// Set default values for fields not in database
			acc.IsVerified = false // Default to false until admin verifies
			acc.IsActive = true
			accounts = append(accounts, acc)
		}

		c.JSON(http.StatusOK, gin.H{
			"bank_accounts": accounts,
			"total":         len(accounts),
		})
	}
}

// Delete Bank Account
func deleteBankAccountHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")
		bankAccountID := c.Param("bank_account_id")

		result, err := dbPool.Exec(ctx, `
			DELETE FROM bank_accounts
			WHERE bank_account_id = $1 AND user_id = $2
		`, bankAccountID, userID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete bank account"})
			return
		}

		if result.RowsAffected() == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Bank account not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Bank account deleted successfully"})
	}
}

// ================================
// Wallet Handlers
// ================================

// Get My Wallet
func getMyWalletHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")

		// Get or create wallet
		var wallet Wallet
		err := dbPool.QueryRow(ctx, `
			INSERT INTO wallets (user_id) VALUES ($1)
			ON CONFLICT (user_id) DO UPDATE SET user_id = EXCLUDED.user_id
			RETURNING wallet_id, user_id, available_balance, pending_balance,
			          total_earned, total_withdrawn, updated_at
		`, userID).Scan(
			&wallet.WalletID, &wallet.UserID, &wallet.AvailableBalance,
			&wallet.PendingBalance, &wallet.TotalEarned, &wallet.TotalWithdrawn,
			&wallet.LastUpdated,
		)

		if err != nil {
			log.Printf("❌ Failed to fetch/create wallet for user %v: %v", userID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch wallet"})
			return
		}

		// Get recent transactions
		rows, err := dbPool.Query(ctx, `
			SELECT transaction_id, transaction_uuid, type, status, amount,
			       commission_amount, net_amount, description, created_at
			FROM transactions
			WHERE user_id = $1
			ORDER BY created_at DESC
			LIMIT 10
		`, userID)

		transactions := make([]Transaction, 0)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var tx Transaction
				rows.Scan(&tx.TransactionID, &tx.TransactionUUID, &tx.Type, &tx.Status,
					&tx.Amount, &tx.CommissionAmount, &tx.NetAmount, &tx.Description, &tx.CreatedAt)
				transactions = append(transactions, tx)
			}
		}

		// Get booking statistics
		var stats struct {
			TotalBookings     int     `json:"total_bookings"`
			CompletedBookings int     `json:"completed_bookings"`
			PendingAmount     float64 `json:"pending_amount"`
			AvailableAmount   float64 `json:"available_amount"`
		}

		dbPool.QueryRow(ctx, `
			SELECT 
				COUNT(DISTINCT b.booking_id) FILTER (WHERE b.provider_id = $1),
				COUNT(DISTINCT b.booking_id) FILTER (WHERE b.provider_id = $1 AND b.status = 'completed'),
				COALESCE((SELECT pending_balance FROM wallets WHERE user_id = $1), 0),
				COALESCE((SELECT available_balance FROM wallets WHERE user_id = $1), 0)
			FROM bookings b
		`, userID).Scan(&stats.TotalBookings, &stats.CompletedBookings, &stats.PendingAmount, &stats.AvailableAmount)

		if err != nil {
			log.Printf("❌ Failed to fetch booking stats for user %v: %v", userID, err)
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"wallet":              wallet,
				"recent_transactions": transactions,
				"stats":               stats,
			},
		})
	}
}

// ================================
// Withdrawal Handlers
// ================================

// Request Withdrawal
func requestWithdrawalHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")

		var req WithdrawRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate minimum withdrawal (e.g., 100 baht)
		if req.RequestedAmount < 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Minimum withdrawal is 100 baht"})
			return
		}

		// Check wallet balance
		var availableBalance float64
		err := dbPool.QueryRow(ctx, `
			SELECT available_balance FROM wallets WHERE user_id = $1
		`, userID).Scan(&availableBalance)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check wallet balance"})
			return
		}

		if availableBalance < req.RequestedAmount {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":             "Insufficient balance",
				"available_balance": availableBalance,
				"requested_amount":  req.RequestedAmount,
			})
			return
		}

		// Verify bank account exists
		var exists bool
		err = dbPool.QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM bank_accounts WHERE bank_account_id = $1 AND user_id = $2)
		`, req.BankAccountID, userID).Scan(&exists)

		if err != nil || !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bank account"})
			return
		}

		// Calculate fee (e.g., 10 baht fixed fee)
		const withdrawalFee = 10.0
		netAmount := req.RequestedAmount - withdrawalFee

		// Start transaction
		tx, err := dbPool.Begin(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
			return
		}
		defer tx.Rollback(ctx)

		// Get wallet_id
		var walletID int
		err = tx.QueryRow(ctx, `
			SELECT wallet_id FROM wallets WHERE user_id = $1
		`, userID).Scan(&walletID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Wallet not found"})
			return
		}

		// Create withdrawal request
		var withdrawalID int
		err = tx.QueryRow(ctx, `
			INSERT INTO withdrawal_requests (
				wallet_id, bank_account_id, amount, status
			) VALUES ($1, $2, $3, 'pending')
			RETURNING withdrawal_id
		`, walletID, req.BankAccountID, req.RequestedAmount).Scan(&withdrawalID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create withdrawal"})
			return
		}

		// Deduct from available balance
		_, err = tx.Exec(ctx, `
			UPDATE wallets
			SET available_balance = available_balance - $1
			WHERE user_id = $2
		`, req.RequestedAmount, userID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update wallet"})
			return
		}

		// Create transaction record
		_, err = tx.Exec(ctx, `
			INSERT INTO transactions (
				wallet_id, type, status, amount, platform_fee, net_amount, booking_id
			) VALUES ($1, 'withdrawal', 'pending', $2, $3, $4, NULL)
		`, walletID, req.RequestedAmount, withdrawalFee, netAmount)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transaction"})
			return
		}

		// Commit transaction
		if err = tx.Commit(ctx); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":          "Withdrawal request created successfully",
			"withdrawal_id":    withdrawalID,
			"requested_amount": req.RequestedAmount,
			"fee":              withdrawalFee,
			"net_amount":       netAmount,
			"status":           "pending",
		})
	}
}

// Get My Withdrawals
func getMyWithdrawalsHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")

		rows, err := dbPool.Query(ctx, `
			SELECT wr.withdrawal_id, wr.amount, wr.status, wr.requested_at,
			       wr.approved_at, wr.rejected_at, wr.completed_at, wr.admin_notes,
			       ba.bank_name, ba.account_number, ba.account_holder_name
			FROM withdrawal_requests wr
			JOIN wallets w ON wr.wallet_id = w.wallet_id
			JOIN bank_accounts ba ON wr.bank_account_id = ba.bank_account_id
			WHERE w.user_id = $1
			ORDER BY wr.requested_at DESC
		`, userID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch withdrawals"})
			return
		}
		defer rows.Close()

		withdrawals := make([]map[string]interface{}, 0)
		for rows.Next() {
			var (
				id                int
				amount            float64
				status            string
				requestedAt       time.Time
				approvedAt        sql.NullTime
				rejectedAt        sql.NullTime
				completedAt       sql.NullTime
				adminNotes        sql.NullString
				bankName          string
				accountNumber     string
				accountHolderName string
			)

			rows.Scan(&id, &amount, &status, &requestedAt,
				&approvedAt, &rejectedAt, &completedAt, &adminNotes,
				&bankName, &accountNumber, &accountHolderName)

			// Calculate fee (10 THB fixed fee)
			fee := 10.0
			netAmount := amount - fee

			withdrawal := map[string]interface{}{
				"withdrawal_id":    id,
				"requested_amount": amount,
				"fee":              fee,
				"net_amount":       netAmount,
				"status":           status,
				"requested_at":     requestedAt,
				"bank_name":        bankName,
				"account_number":   accountNumber,
				"account_name":     accountHolderName,
			}

			if approvedAt.Valid {
				withdrawal["approved_at"] = approvedAt.Time
			}
			if rejectedAt.Valid {
				withdrawal["rejected_at"] = rejectedAt.Time
			}
			if completedAt.Valid {
				withdrawal["completed_at"] = completedAt.Time
			}
			if adminNotes.Valid {
				withdrawal["admin_notes"] = adminNotes.String
			}

			withdrawals = append(withdrawals, withdrawal)
		}

		c.JSON(http.StatusOK, gin.H{
			"withdrawals": withdrawals,
			"total":       len(withdrawals),
		})
	}
}

// ================================
// Transaction History
// ================================

// Get My Transactions
func getMyTransactionsHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		txType := c.Query("type") // optional filter

		if page < 1 {
			page = 1
		}
		if limit < 1 || limit > 100 {
			limit = 20
		}
		offset := (page - 1) * limit

		query := `
			SELECT t.transaction_id, t.type, t.status, t.amount,
			       t.platform_fee, t.net_amount, t.booking_id, t.created_at
			FROM transactions t
			INNER JOIN wallets w ON t.wallet_id = w.wallet_id
			WHERE w.user_id = $1
		`
		args := []interface{}{userID}

		if txType != "" {
			query += " AND t.type = $2"
			args = append(args, txType)
			query += " ORDER BY t.created_at DESC LIMIT $3 OFFSET $4"
			args = append(args, limit, offset)
		} else {
			query += " ORDER BY t.created_at DESC LIMIT $2 OFFSET $3"
			args = append(args, limit, offset)
		}

		rows, err := dbPool.Query(ctx, query, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
			return
		}
		defer rows.Close()

		transactions := make([]Transaction, 0)
		for rows.Next() {
			var tx Transaction
			rows.Scan(&tx.TransactionID, &tx.Type, &tx.Status,
				&tx.Amount, &tx.CommissionAmount, &tx.NetAmount,
				&tx.BookingID, &tx.CreatedAt)
			transactions = append(transactions, tx)
		}

		// Get total count
		var totalCount int
		countQuery := "SELECT COUNT(*) FROM transactions WHERE user_id = $1"
		if txType != "" {
			dbPool.QueryRow(ctx, countQuery+" AND type = $2", userID, txType).Scan(&totalCount)
		} else {
			dbPool.QueryRow(ctx, countQuery, userID).Scan(&totalCount)
		}

		c.JSON(http.StatusOK, gin.H{
			"transactions": transactions,
			"total":        totalCount,
			"page":         page,
			"limit":        limit,
			"total_pages":  (totalCount + limit - 1) / limit,
		})
	}
}
