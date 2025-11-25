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
				WHERE user_id = $1 AND account_number = $2 AND is_active = true
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

		// Insert bank account
		var bankAccountID int
		err = dbPool.QueryRow(ctx, `
			INSERT INTO bank_accounts (
				user_id, bank_name, bank_code, account_number, account_name,
				account_type, branch_name, is_default, is_active
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, true)
			RETURNING bank_account_id
		`, userID, req.BankName, req.BankCode, req.AccountNumber, req.AccountName,
			req.AccountType, req.BranchName, req.IsDefault).Scan(&bankAccountID)

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
			SELECT bank_account_id, bank_name, bank_code, account_number, account_name,
			       account_type, branch_name, is_verified, verified_at, is_default, is_active,
			       created_at, updated_at
			FROM bank_accounts
			WHERE user_id = $1 AND is_active = true
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
				&acc.BankAccountID, &acc.BankName, &acc.BankCode, &acc.AccountNumber,
				&acc.AccountName, &acc.AccountType, &acc.BranchName, &acc.IsVerified,
				&acc.VerifiedAt, &acc.IsDefault, &acc.IsActive, &acc.CreatedAt, &acc.UpdatedAt,
			)
			if err != nil {
				continue
			}
			acc.UserID = userID.(int)
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
			UPDATE bank_accounts SET is_active = false
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
			          total_earned, total_withdrawn, total_commission_paid,
			          last_updated, created_at
		`, userID).Scan(
			&wallet.WalletID, &wallet.UserID, &wallet.AvailableBalance,
			&wallet.PendingBalance, &wallet.TotalEarned, &wallet.TotalWithdrawn,
			&wallet.TotalCommissionPaid, &wallet.LastUpdated, &wallet.CreatedAt,
		)

		if err != nil {
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

		c.JSON(http.StatusOK, gin.H{
			"wallet":              wallet,
			"recent_transactions": transactions,
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

		// Verify bank account
		var isVerified bool
		err = dbPool.QueryRow(ctx, `
			SELECT is_verified FROM bank_accounts
			WHERE bank_account_id = $1 AND user_id = $2 AND is_active = true
		`, req.BankAccountID, userID).Scan(&isVerified)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bank account"})
			return
		}

		if !isVerified {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Bank account not verified yet"})
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

		// Create withdrawal request
		var withdrawalID int
		var withdrawalUUID string
		err = tx.QueryRow(ctx, `
			INSERT INTO withdrawals (
				user_id, bank_account_id, requested_amount, fee, net_amount, status
			) VALUES ($1, $2, $3, $4, $5, 'pending')
			RETURNING withdrawal_id, withdrawal_uuid
		`, userID, req.BankAccountID, req.RequestedAmount, withdrawalFee, netAmount).Scan(&withdrawalID, &withdrawalUUID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create withdrawal"})
			return
		}

		// Deduct from available balance
		_, err = tx.Exec(ctx, `
			UPDATE wallets
			SET available_balance = available_balance - $1,
			    last_updated = CURRENT_TIMESTAMP
			WHERE user_id = $2
		`, req.RequestedAmount, userID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update wallet"})
			return
		}

		// Create transaction record
		_, err = tx.Exec(ctx, `
			INSERT INTO transactions (
				user_id, type, status, amount, commission_amount, net_amount,
				withdrawal_id, description
			) VALUES ($1, 'withdrawal', 'pending', $2, $3, $4, $5, $6)
		`, userID, req.RequestedAmount, withdrawalFee, netAmount, withdrawalID,
			fmt.Sprintf("Withdrawal request #%d", withdrawalID))

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
			"withdrawal_uuid":  withdrawalUUID,
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
			SELECT w.withdrawal_id, w.withdrawal_uuid, w.requested_amount, w.fee, w.net_amount,
			       w.status, w.requested_at, w.approved_at, w.processed_at, w.completed_at,
			       w.rejection_reason, w.transfer_reference,
			       ba.bank_name, ba.account_number, ba.account_name
			FROM withdrawals w
			JOIN bank_accounts ba ON w.bank_account_id = ba.bank_account_id
			WHERE w.user_id = $1
			ORDER BY w.requested_at DESC
		`, userID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch withdrawals"})
			return
		}
		defer rows.Close()

		withdrawals := make([]map[string]interface{}, 0)
		for rows.Next() {
			var (
				id, uuid                             string
				requested, fee, net                  float64
				status                               string
				requestedAt                          time.Time
				approvedAt, processedAt, completedAt sql.NullTime
				rejectionReason, transferRef         sql.NullString
				bankName, accountNumber, accountName string
			)

			rows.Scan(&id, &uuid, &requested, &fee, &net, &status, &requestedAt,
				&approvedAt, &processedAt, &completedAt, &rejectionReason, &transferRef,
				&bankName, &accountNumber, &accountName)

			withdrawal := map[string]interface{}{
				"withdrawal_id":    id,
				"withdrawal_uuid":  uuid,
				"requested_amount": requested,
				"fee":              fee,
				"net_amount":       net,
				"status":           status,
				"requested_at":     requestedAt,
				"bank_name":        bankName,
				"account_number":   accountNumber,
				"account_name":     accountName,
			}

			if approvedAt.Valid {
				withdrawal["approved_at"] = approvedAt.Time
			}
			if completedAt.Valid {
				withdrawal["completed_at"] = completedAt.Time
			}
			if rejectionReason.Valid {
				withdrawal["rejection_reason"] = rejectionReason.String
			}
			if transferRef.Valid {
				withdrawal["transfer_reference"] = transferRef.String
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
			SELECT transaction_id, transaction_uuid, type, status, amount,
			       commission_amount, net_amount, description, payment_method,
			       booking_id, created_at
			FROM transactions
			WHERE user_id = $1
		`
		args := []interface{}{userID}

		if txType != "" {
			query += " AND type = $2"
			args = append(args, txType)
			query += " ORDER BY created_at DESC LIMIT $3 OFFSET $4"
			args = append(args, limit, offset)
		} else {
			query += " ORDER BY created_at DESC LIMIT $2 OFFSET $3"
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
			rows.Scan(&tx.TransactionID, &tx.TransactionUUID, &tx.Type, &tx.Status,
				&tx.Amount, &tx.CommissionAmount, &tx.NetAmount, &tx.Description,
				&tx.PaymentMethod, &tx.BookingID, &tx.CreatedAt)
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
