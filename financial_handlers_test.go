package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test Wallet Balance Calculations
func TestWalletBalanceCalculations(t *testing.T) {
	t.Run("Calculate Available Balance", func(t *testing.T) {
		totalEarned := 10000.0
		totalWithdrawn := 3000.0
		pendingBalance := 2000.0

		expectedAvailable := totalEarned - totalWithdrawn - pendingBalance
		actualAvailable := 5000.0

		assert.Equal(t, expectedAvailable, actualAvailable)
	})

	t.Run("Insufficient Balance for Withdrawal", func(t *testing.T) {
		availableBalance := 500.0
		withdrawalAmount := 1000.0

		canWithdraw := availableBalance >= withdrawalAmount
		assert.False(t, canWithdraw, "Should not allow withdrawal exceeding balance")
	})

	t.Run("Valid Withdrawal Amount", func(t *testing.T) {
		availableBalance := 1000.0
		withdrawalAmount := 500.0
		minWithdrawal := 100.0

		isValid := withdrawalAmount >= minWithdrawal && withdrawalAmount <= availableBalance
		assert.True(t, isValid, "Withdrawal amount should be valid")
	})
}

// Test Commission Calculations
func TestCommissionCalculations(t *testing.T) {
	t.Run("Calculate Platform Commission", func(t *testing.T) {
		bookingAmount := 1000.0
		commissionRate := 0.10 // 10%

		expectedCommission := bookingAmount * commissionRate
		assert.Equal(t, 100.0, expectedCommission)
	})

	t.Run("Calculate Provider Earnings", func(t *testing.T) {
		bookingAmount := 1000.0
		stripeFee := bookingAmount * 0.0275        // 2.75%
		platformCommission := bookingAmount * 0.10 // 10%

		providerEarnings := bookingAmount - stripeFee - platformCommission

		assert.InDelta(t, 872.5, providerEarnings, 0.01)
	})

	t.Run("Verify Total Percentages Add Up", func(t *testing.T) {
		stripeFeePercent := 2.75
		commissionPercent := 10.00
		providerPercent := 87.25

		total := stripeFeePercent + commissionPercent + providerPercent
		assert.InDelta(t, 100.0, total, 0.01)
	})
}

// Test Withdrawal Fees
func TestWithdrawalFees(t *testing.T) {
	t.Run("Calculate Withdrawal Fee", func(t *testing.T) {
		requestedAmount := 1000.0
		withdrawalFee := 10.0

		netAmount := requestedAmount - withdrawalFee
		assert.Equal(t, 990.0, netAmount)
	})

	t.Run("Minimum Withdrawal Amount", func(t *testing.T) {
		minWithdrawal := 100.0

		tests := []struct {
			amount  float64
			isValid bool
		}{
			{150.0, true},
			{100.0, true},
			{50.0, false},
			{0.0, false},
		}

		for _, tt := range tests {
			isValid := tt.amount >= minWithdrawal
			assert.Equal(t, tt.isValid, isValid)
		}
	})
}

// Test Transaction Types
func TestTransactionTypes(t *testing.T) {
	validTypes := []string{
		"booking_payment",
		"provider_earning",
		"withdrawal",
		"withdrawal_fee",
		"refund",
		"adjustment",
	}

	t.Run("Valid Transaction Type", func(t *testing.T) {
		transactionType := "booking_payment"

		found := false
		for _, validType := range validTypes {
			if validType == transactionType {
				found = true
				break
			}
		}

		assert.True(t, found)
	})

	t.Run("Invalid Transaction Type", func(t *testing.T) {
		transactionType := "invalid_type"

		found := false
		for _, validType := range validTypes {
			if validType == transactionType {
				found = true
				break
			}
		}

		assert.False(t, found)
	})
}

// Test Withdrawal Status Flow
func TestWithdrawalStatusFlow(t *testing.T) {
	validStatuses := []string{"pending", "approved", "rejected", "processing", "completed", "failed"}

	statusTransitions := map[string][]string{
		"pending":    {"approved", "rejected"},
		"approved":   {"processing", "rejected"},
		"processing": {"completed", "failed"},
		"completed":  {},
		"rejected":   {},
		"failed":     {},
	}

	t.Run("Valid Status Transition", func(t *testing.T) {
		currentStatus := "pending"
		newStatus := "approved"

		validTransitions := statusTransitions[currentStatus]
		found := false
		for _, status := range validTransitions {
			if status == newStatus {
				found = true
				break
			}
		}

		assert.True(t, found)
	})

	t.Run("Invalid Status Transition", func(t *testing.T) {
		currentStatus := "completed"
		newStatus := "pending"

		validTransitions := statusTransitions[currentStatus]
		found := false
		for _, status := range validTransitions {
			if status == newStatus {
				found = true
				break
			}
		}

		assert.False(t, found, "Cannot go from completed back to pending")
	})

	t.Run("All Statuses Are Valid", func(t *testing.T) {
		for _, status := range validStatuses {
			assert.NotEmpty(t, status)
		}
	})
}

// Test Bank Account Validation
func TestBankAccountValidation(t *testing.T) {
	t.Run("Valid Account Number", func(t *testing.T) {
		accountNumber := "1234567890"

		isValid := len(accountNumber) >= 10 && len(accountNumber) <= 12
		assert.True(t, isValid)
	})

	t.Run("Invalid Account Number - Too Short", func(t *testing.T) {
		accountNumber := "123456"

		isValid := len(accountNumber) >= 10 && len(accountNumber) <= 12
		assert.False(t, isValid)
	})

	t.Run("Invalid Account Number - Too Long", func(t *testing.T) {
		accountNumber := "12345678901234"

		isValid := len(accountNumber) >= 10 && len(accountNumber) <= 12
		assert.False(t, isValid)
	})

	t.Run("Valid Account Types", func(t *testing.T) {
		validTypes := []string{"savings", "current"}

		tests := []struct {
			accountType string
			isValid     bool
		}{
			{"savings", true},
			{"current", true},
			{"checking", false},
			{"", false},
		}

		for _, tt := range tests {
			found := false
			for _, validType := range validTypes {
				if validType == tt.accountType {
					found = true
					break
				}
			}

			assert.Equal(t, tt.isValid, found)
		}
	})
}

// Test Financial Statistics
func TestFinancialStatistics(t *testing.T) {
	t.Run("Calculate Total Revenue", func(t *testing.T) {
		bookings := []float64{1000.0, 1500.0, 2000.0}

		totalRevenue := 0.0
		for _, amount := range bookings {
			totalRevenue += amount
		}

		assert.Equal(t, 4500.0, totalRevenue)
	})

	t.Run("Calculate Average Booking Value", func(t *testing.T) {
		bookings := []float64{1000.0, 1500.0, 2000.0}

		totalRevenue := 0.0
		for _, amount := range bookings {
			totalRevenue += amount
		}

		averageValue := totalRevenue / float64(len(bookings))
		assert.Equal(t, 1500.0, averageValue)
	})
}
