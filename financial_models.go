package main

import (
	"time"
)

// ================================
// Bank Account Models
// ================================

type BankAccount struct {
	BankAccountID int        `json:"bank_account_id"`
	UserID        int        `json:"user_id"`
	BankName      string     `json:"bank_name"`
	BankCode      *string    `json:"bank_code"`
	AccountNumber string     `json:"account_number"`
	AccountName   string     `json:"account_name"`
	AccountType   string     `json:"account_type"`
	BranchName    *string    `json:"branch_name"`
	IsVerified    bool       `json:"is_verified"`
	VerifiedAt    *time.Time `json:"verified_at"`
	IsDefault     bool       `json:"is_default"`
	IsActive      bool       `json:"is_active"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// ================================
// Wallet Models
// ================================

type Wallet struct {
	WalletID            int       `json:"wallet_id"`
	UserID              int       `json:"user_id"`
	AvailableBalance    float64   `json:"available_balance"`
	PendingBalance      float64   `json:"pending_balance"`
	TotalEarned         float64   `json:"total_earned"`
	TotalWithdrawn      float64   `json:"total_withdrawn"`
	TotalCommissionPaid float64   `json:"total_commission_paid"`
	LastUpdated         time.Time `json:"last_updated"`
	CreatedAt           time.Time `json:"created_at"`
}

// ================================
// Transaction Models
// ================================

type TransactionType string

const (
	TxTypeBookingPayment  TransactionType = "booking_payment"
	TxTypeBookingRefund   TransactionType = "booking_refund"
	TxTypeCommission      TransactionType = "commission"
	TxTypeProviderEarning TransactionType = "provider_earning"
	TxTypeWithdrawal      TransactionType = "withdrawal"
	TxTypeSubscription    TransactionType = "subscription_fee"
	TxTypeAdjustment      TransactionType = "admin_adjustment"
	TxTypeBonus           TransactionType = "bonus"
	TxTypePenalty         TransactionType = "penalty"
)

type TransactionStatus string

const (
	TxStatusPending    TransactionStatus = "pending"
	TxStatusProcessing TransactionStatus = "processing"
	TxStatusCompleted  TransactionStatus = "completed"
	TxStatusFailed     TransactionStatus = "failed"
	TxStatusCancelled  TransactionStatus = "cancelled"
	TxStatusRefunded   TransactionStatus = "refunded"
)

type Transaction struct {
	TransactionID    int               `json:"transaction_id"`
	TransactionUUID  string            `json:"transaction_uuid"`
	UserID           int               `json:"user_id"`
	RelatedUserID    *int              `json:"related_user_id"`
	Type             TransactionType   `json:"type"`
	Status           TransactionStatus `json:"status"`
	Amount           float64           `json:"amount"`
	CommissionAmount float64           `json:"commission_amount"`
	NetAmount        float64           `json:"net_amount"`
	BookingID        *int              `json:"booking_id"`
	WithdrawalID     *int              `json:"withdrawal_id"`
	PaymentMethod    *string           `json:"payment_method"`
	PaymentIntentID  *string           `json:"payment_intent_id"`
	Description      *string           `json:"description"`
	Notes            *string           `json:"notes"`
	Metadata         *string           `json:"metadata"` // JSONB as string
	ProcessedAt      *time.Time        `json:"processed_at"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

// ================================
// Withdrawal Models
// ================================

type WithdrawalStatus string

const (
	WithdrawalPending    WithdrawalStatus = "pending"
	WithdrawalApproved   WithdrawalStatus = "approved"
	WithdrawalProcessing WithdrawalStatus = "processing"
	WithdrawalCompleted  WithdrawalStatus = "completed"
	WithdrawalRejected   WithdrawalStatus = "rejected"
	WithdrawalFailed     WithdrawalStatus = "failed"
)

type Withdrawal struct {
	WithdrawalID      int              `json:"withdrawal_id"`
	WithdrawalUUID    string           `json:"withdrawal_uuid"`
	UserID            int              `json:"user_id"`
	BankAccountID     int              `json:"bank_account_id"`
	RequestedAmount   float64          `json:"requested_amount"`
	Fee               float64          `json:"fee"`
	NetAmount         float64          `json:"net_amount"`
	Status            WithdrawalStatus `json:"status"`
	RequestedAt       time.Time        `json:"requested_at"`
	ApprovedAt        *time.Time       `json:"approved_at"`
	ApprovedBy        *int             `json:"approved_by"`
	ProcessedAt       *time.Time       `json:"processed_at"`
	CompletedAt       *time.Time       `json:"completed_at"`
	TransferReference *string          `json:"transfer_reference"`
	TransferSlipURL   *string          `json:"transfer_slip_url"`
	RejectionReason   *string          `json:"rejection_reason"`
	RejectedAt        *time.Time       `json:"rejected_at"`
	RejectedBy        *int             `json:"rejected_by"`
	Notes             *string          `json:"notes"`
	Metadata          *string          `json:"metadata"`
	CreatedAt         time.Time        `json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`

	// Joined data
	BankAccount *BankAccount `json:"bank_account,omitempty"`
	UserName    *string      `json:"user_name,omitempty"`
}

// ================================
// Commission Rule Models
// ================================

type CommissionRule struct {
	RuleID             int        `json:"rule_id"`
	Name               string     `json:"name"`
	Description        *string    `json:"description"`
	PlatformRate       float64    `json:"platform_rate"`        // 0.1000 = 10%
	PaymentGatewayRate float64    `json:"payment_gateway_rate"` // 0.0275 = 2.75%
	TierID             *int       `json:"tier_id"`
	EffectiveFrom      time.Time  `json:"effective_from"`
	EffectiveUntil     *time.Time `json:"effective_until"`
	IsActive           bool       `json:"is_active"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// ================================
// Financial Report Models
// ================================

type FinancialReport struct {
	ReportID              int       `json:"report_id"`
	ReportType            string    `json:"report_type"`
	PeriodStart           time.Time `json:"period_start"`
	PeriodEnd             time.Time `json:"period_end"`
	TotalBookings         int       `json:"total_bookings"`
	TotalRevenue          float64   `json:"total_revenue"`
	TotalCommission       float64   `json:"total_commission"`
	TotalProviderEarnings float64   `json:"total_provider_earnings"`
	TotalWithdrawals      float64   `json:"total_withdrawals"`
	TotalSubscriptions    float64   `json:"total_subscriptions"`
	Breakdown             *string   `json:"breakdown"` // JSONB as string
	GeneratedAt           time.Time `json:"generated_at"`
	GeneratedBy           *int      `json:"generated_by"`
	CreatedAt             time.Time `json:"created_at"`
}

// ================================
// Request/Response DTOs
// ================================

type AddBankAccountRequest struct {
	BankName      string  `json:"bank_name" binding:"required"`
	BankCode      *string `json:"bank_code"`
	AccountNumber string  `json:"account_number" binding:"required"`
	AccountName   string  `json:"account_name" binding:"required"`
	AccountType   string  `json:"account_type"`
	BranchName    *string `json:"branch_name"`
	IsDefault     bool    `json:"is_default"`
}

type WithdrawRequest struct {
	BankAccountID   int     `json:"bank_account_id" binding:"required"`
	RequestedAmount float64 `json:"requested_amount" binding:"required,gt=0"`
}

type ProcessWithdrawalRequest struct {
	WithdrawalID      int     `json:"withdrawal_id" binding:"required"`
	Action            string  `json:"action" binding:"required,oneof=approve reject complete"`
	TransferReference *string `json:"transfer_reference"`
	TransferSlipURL   *string `json:"transfer_slip_url"`
	RejectionReason   *string `json:"rejection_reason"`
	Notes             *string `json:"notes"`
}

type WalletSummary struct {
	Wallet          Wallet        `json:"wallet"`
	Transactions    []Transaction `json:"recent_transactions"`
	PendingBookings int           `json:"pending_bookings_count"`
}

type FinancialSummary struct {
	TodayRevenue       float64 `json:"today_revenue"`
	MonthRevenue       float64 `json:"month_revenue"`
	TodayCommission    float64 `json:"today_commission"`
	MonthCommission    float64 `json:"month_commission"`
	PendingWithdrawals int     `json:"pending_withdrawals"`
	TotalUsers         int     `json:"total_users"`
}
