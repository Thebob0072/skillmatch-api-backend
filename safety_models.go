package main

import "time"

// ================================
// Trusted Contact Models
// ================================

type TrustedContact struct {
	ContactID    int        `json:"contact_id"`
	UserID       int        `json:"user_id"`
	Name         string     `json:"name"`
	PhoneNumber  string     `json:"phone_number"`
	Relationship string     `json:"relationship"` // friend, family, other
	IsActive     bool       `json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	LastNotified *time.Time `json:"last_notified"` // วันที่แจ้งเตือนล่าสุด
}

// ================================
// Emergency SOS Models
// ================================

type SOSAlert struct {
	AlertID        int        `json:"alert_id"`
	UserID         int        `json:"user_id"`
	BookingID      *int       `json:"booking_id"` // ถ้ามี booking
	Latitude       float64    `json:"latitude"`
	Longitude      float64    `json:"longitude"`
	LocationText   *string    `json:"location_text"` // ที่อยู่แบบ text
	Status         string     `json:"status"`        // active, resolved, cancelled
	ResolvedAt     *time.Time `json:"resolved_at"`
	ResolvedBy     *int       `json:"resolved_by"` // admin ที่จัดการ
	ResolutionNote *string    `json:"resolution_note"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type SOSAlertWithDetails struct {
	SOSAlert
	UserName         string  `json:"user_name"`
	UserPhone        *string `json:"user_phone"`
	ContactsNotified int     `json:"contacts_notified"`
	BookingDetails   *string `json:"booking_details"` // JSON string
}

// ================================
// Check-in/Check-out Models
// ================================

type BookingCheckIn struct {
	CheckInID       int        `json:"check_in_id"`
	BookingID       int        `json:"booking_id"`
	ProviderID      int        `json:"provider_id"`
	ClientID        int        `json:"client_id"`
	CheckedInAt     time.Time  `json:"checked_in_at"`
	ExpectedEndTime time.Time  `json:"expected_end_time"`
	CheckedOutAt    *time.Time `json:"checked_out_at"`
	Status          string     `json:"status"` // active, completed, overdue, emergency
	Latitude        *float64   `json:"latitude"`
	Longitude       *float64   `json:"longitude"`
	Notes           *string    `json:"notes"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// ================================
// Private Gallery Models
// ================================

type PrivatePhoto struct {
	PhotoID      int       `json:"photo_id"`
	UserID       int       `json:"user_id"`
	PhotoURL     string    `json:"photo_url"`
	ThumbnailURL *string   `json:"thumbnail_url"` // รูปเบลอ
	SortOrder    int       `json:"sort_order"`
	Price        *float64  `json:"price"` // ราคาถ้าขาย per photo (null = แพ็คเกจ)
	IsActive     bool      `json:"is_active"`
	UploadedAt   time.Time `json:"uploaded_at"`
}

type PrivateGalleryAccess struct {
	AccessID       int        `json:"access_id"`
	GalleryOwnerID int        `json:"gallery_owner_id"` // เจ้าของแกลเลอรี่
	ViewerID       int        `json:"viewer_id"`        // คนดู
	AccessType     string     `json:"access_type"`      // subscription, one_time
	ExpiresAt      *time.Time `json:"expires_at"`       // หมดอายุ (null = ถาวร)
	GrantedAt      time.Time  `json:"granted_at"`
	PaymentID      *int       `json:"payment_id"` // อ้างอิงการจ่ายเงิน
}

type PrivateGallerySettings struct {
	SettingID    int       `json:"setting_id"`
	UserID       int       `json:"user_id"`
	IsEnabled    bool      `json:"is_enabled"`     // เปิดใช้งาน private gallery
	MonthlyPrice *float64  `json:"monthly_price"`  // ราคารายเดือน
	OneTimePrice *float64  `json:"one_time_price"` // ราคาครั้งเดียว
	AllowOneTime bool      `json:"allow_one_time"` // อนุญาตจ่ายครั้งเดียว
	UpdatedAt    time.Time `json:"updated_at"`
}

// ================================
// Deposit System Models
// ================================

type BookingDeposit struct {
	DepositID       int        `json:"deposit_id"`
	BookingID       int        `json:"booking_id"`
	ClientID        int        `json:"client_id"`
	ProviderID      int        `json:"provider_id"`
	Amount          float64    `json:"amount"`
	Percentage      float64    `json:"percentage"` // เช่น 0.30 = 30%
	Status          string     `json:"status"`     // pending, paid, refunded, forfeited
	PaidAt          *time.Time `json:"paid_at"`
	RefundedAt      *time.Time `json:"refunded_at"`
	ForfeitedAt     *time.Time `json:"forfeited_at"` // ถูกยึด (client ยกเลิก)
	PaymentIntentID *string    `json:"payment_intent_id"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// ================================
// Cancellation Fee Models
// ================================

type CancellationPolicy struct {
	PolicyID           int       `json:"policy_id"`
	ProviderID         int       `json:"provider_id"`
	HoursBeforeBooking int       `json:"hours_before_booking"` // ชั่วโมงก่อน booking
	FeePercentage      float64   `json:"fee_percentage"`       // ค่าปรับ %
	IsActive           bool      `json:"is_active"`
	CreatedAt          time.Time `json:"created_at"`
}

type CancellationFee struct {
	FeeID         int        `json:"fee_id"`
	BookingID     int        `json:"booking_id"`
	CancelledBy   int        `json:"cancelled_by"` // user_id ที่ยกเลิก
	FeeAmount     float64    `json:"fee_amount"`
	FeePercentage float64    `json:"fee_percentage"`
	Status        string     `json:"status"` // pending, paid, waived
	PaidAt        *time.Time `json:"paid_at"`
	WaivedAt      *time.Time `json:"waived_at"`
	WaivedBy      *int       `json:"waived_by"` // admin ที่ยกเว้น
	WaiverReason  *string    `json:"waiver_reason"`
	CreatedAt     time.Time  `json:"created_at"`
}

// ================================
// Featured/Boost Profile Models
// ================================

type ProfileBoost struct {
	BoostID   int       `json:"boost_id"`
	UserID    int       `json:"user_id"`
	BoostType string    `json:"boost_type"` // featured, spotlight, top_search
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Amount    float64   `json:"amount"` // ราคาที่จ่าย
	Status    string    `json:"status"` // active, expired, cancelled
	PaymentID *int      `json:"payment_id"`
	CreatedAt time.Time `json:"created_at"`
}

type BoostPackage struct {
	PackageID   int       `json:"package_id"`
	Name        string    `json:"name"`
	BoostType   string    `json:"boost_type"`
	Duration    int       `json:"duration"` // ชั่วโมง
	Price       float64   `json:"price"`
	Description *string   `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

// ================================
// Promotion/Coupon Models
// ================================

type Coupon struct {
	CouponID         int       `json:"coupon_id"`
	Code             string    `json:"code"`
	DiscountType     string    `json:"discount_type"` // percentage, fixed
	DiscountValue    float64   `json:"discount_value"`
	MinBookingAmount *float64  `json:"min_booking_amount"` // ขั้นต่ำ
	MaxDiscount      *float64  `json:"max_discount"`       // ลดสูงสุด
	ValidFrom        time.Time `json:"valid_from"`
	ValidUntil       time.Time `json:"valid_until"`
	UsageLimit       *int      `json:"usage_limit"` // ใช้ได้กี่ครั้ง
	UsedCount        int       `json:"used_count"`
	IsActive         bool      `json:"is_active"`
	CreatedBy        int       `json:"created_by"`  // admin/provider
	ProviderID       *int      `json:"provider_id"` // null = platform-wide
	CreatedAt        time.Time `json:"created_at"`
}

type CouponUsage struct {
	UsageID        int       `json:"usage_id"`
	CouponID       int       `json:"coupon_id"`
	UserID         int       `json:"user_id"`
	BookingID      int       `json:"booking_id"`
	DiscountAmount float64   `json:"discount_amount"`
	UsedAt         time.Time `json:"used_at"`
}

// ================================
// Verified Photo Badge Models
// ================================

type PhotoVerification struct {
	VerificationID  int        `json:"verification_id"`
	PhotoID         int        `json:"photo_id"`
	UserID          int        `json:"user_id"`
	Status          string     `json:"status"` // pending, verified, rejected
	VerifiedAt      *time.Time `json:"verified_at"`
	VerifiedBy      *int       `json:"verified_by"` // admin
	RejectionReason *string    `json:"rejection_reason"`
	CreatedAt       time.Time  `json:"created_at"`
}

// ================================
// Request/Response DTOs
// ================================

// Trusted Contact
type AddTrustedContactRequest struct {
	Name         string `json:"name" binding:"required"`
	PhoneNumber  string `json:"phone_number" binding:"required"`
	Relationship string `json:"relationship" binding:"required"`
}

// SOS
type TriggerSOSRequest struct {
	Latitude     float64 `json:"latitude" binding:"required"`
	Longitude    float64 `json:"longitude" binding:"required"`
	LocationText *string `json:"location_text"`
	BookingID    *int    `json:"booking_id"`
}

// Check-in
type CheckInRequest struct {
	BookingID int      `json:"booking_id" binding:"required"`
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
	Notes     *string  `json:"notes"`
}

type CheckOutRequest struct {
	BookingID int     `json:"booking_id" binding:"required"`
	Notes     *string `json:"notes"`
}

// Private Gallery
type PrivateGallerySettingsRequest struct {
	IsEnabled    bool     `json:"is_enabled"`
	MonthlyPrice *float64 `json:"monthly_price"`
	OneTimePrice *float64 `json:"one_time_price"`
	AllowOneTime bool     `json:"allow_one_time"`
}

type PurchaseGalleryRequest struct {
	ProviderID int    `json:"provider_id" binding:"required"`
	AccessType string `json:"access_type" binding:"required"` // subscription, one_time
}

// Deposit
type DepositSettingsRequest struct {
	RequireDeposit    bool    `json:"require_deposit"`
	DepositPercentage float64 `json:"deposit_percentage"` // 0.10 - 0.50 (10-50%)
}

// Cancellation
type CancellationPolicyRequest struct {
	Policies []struct {
		HoursBeforeBooking int     `json:"hours_before_booking"`
		FeePercentage      float64 `json:"fee_percentage"`
	} `json:"policies"`
}

// Boost
type PurchaseBoostRequest struct {
	PackageID int `json:"package_id" binding:"required"`
}

// Coupon
type CreateCouponRequest struct {
	Code             string   `json:"code" binding:"required"`
	DiscountType     string   `json:"discount_type" binding:"required"` // percentage, fixed
	DiscountValue    float64  `json:"discount_value" binding:"required"`
	MinBookingAmount *float64 `json:"min_booking_amount"`
	MaxDiscount      *float64 `json:"max_discount"`
	ValidFrom        string   `json:"valid_from" binding:"required"`
	ValidUntil       string   `json:"valid_until" binding:"required"`
	UsageLimit       *int     `json:"usage_limit"`
}

type ApplyCouponRequest struct {
	Code      string `json:"code" binding:"required"`
	BookingID int    `json:"booking_id" binding:"required"`
}

// Photo Verification
type SubmitPhotoVerificationRequest struct {
	PhotoID int `json:"photo_id" binding:"required"`
}
