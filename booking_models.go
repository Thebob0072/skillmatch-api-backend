package main

import "time"

// ServicePackage (แพ็คเกจบริการของ Provider)
type ServicePackage struct {
	PackageID   int       `json:"package_id"`
	ProviderID  int       `json:"provider_id"`
	PackageName string    `json:"package_name"` // เช่น "1 Hour", "2 Hours", "Overnight"
	Description *string   `json:"description"`  // รายละเอียดบริการ
	Duration    int       `json:"duration"`     // ระยะเวลา (นาที)
	Price       float64   `json:"price"`        // ราคา
	IsActive    bool      `json:"is_active"`    // เปิด/ปิดใช้งาน
	CreatedAt   time.Time `json:"created_at"`
}

// Booking (การจองของ Client)
type Booking struct {
	BookingID          int        `json:"booking_id"`
	ClientID           int        `json:"client_id"`     // ผู้จอง
	ProviderID         int        `json:"provider_id"`   // ผู้ให้บริการ
	PackageID          int        `json:"package_id"`    // แพ็คเกจที่เลือก
	BookingDate        time.Time  `json:"booking_date"`  // วันที่จอง
	StartTime          time.Time  `json:"start_time"`    // เวลาเริ่ม
	EndTime            time.Time  `json:"end_time"`      // เวลาสิ้นสุด
	TotalPrice         float64    `json:"total_price"`   // ราคารวม
	Status             string     `json:"status"`        // pending, confirmed, completed, cancelled
	Location           *string    `json:"location"`      // สถานที่ (Incall/Outcall)
	SpecialNotes       *string    `json:"special_notes"` // หมายเหตุพิเศษ
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	CompletedAt        *time.Time `json:"completed_at"`
	CancelledAt        *time.Time `json:"cancelled_at"`
	CancellationReason *string    `json:"cancellation_reason"`
}

// BookingWithDetails (สำหรับแสดงข้อมูลเต็ม)
type BookingWithDetails struct {
	BookingID          int       `json:"booking_id"`
	ClientID           int       `json:"client_id"`
	ClientUsername     string    `json:"client_username"`
	ProviderID         int       `json:"provider_id"`
	ProviderUsername   string    `json:"provider_username"`
	ProviderProfilePic *string   `json:"provider_profile_pic"`
	PackageName        string    `json:"package_name"`
	Duration           int       `json:"duration"`
	BookingDate        time.Time `json:"booking_date"`
	StartTime          time.Time `json:"start_time"`
	EndTime            time.Time `json:"end_time"`
	TotalPrice         float64   `json:"total_price"`
	Status             string    `json:"status"`
	Location           *string   `json:"location"`
	SpecialNotes       *string   `json:"special_notes"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// Review (รีวิวหลังใช้บริการ)
type Review struct {
	ReviewID   int       `json:"review_id"`
	BookingID  int       `json:"booking_id"`
	ClientID   int       `json:"client_id"`
	ProviderID int       `json:"provider_id"`
	Rating     int       `json:"rating"`      // 1-5
	Comment    *string   `json:"comment"`     // ความคิดเห็น
	IsVerified bool      `json:"is_verified"` // ยืนยันว่าใช้บริการจริง
	CreatedAt  time.Time `json:"created_at"`
}

// ReviewWithDetails (สำหรับแสดงรีวิว)
type ReviewWithDetails struct {
	ReviewID       int       `json:"review_id"`
	ClientUsername string    `json:"client_username"`
	Rating         int       `json:"rating"`
	Comment        *string   `json:"comment"`
	IsVerified     bool      `json:"is_verified"`
	CreatedAt      time.Time `json:"created_at"`
}

// ProviderAvailability (กำหนดช่วงเวลาว่าง)
type ProviderAvailability struct {
	AvailabilityID int    `json:"availability_id"`
	ProviderID     int    `json:"provider_id"`
	DayOfWeek      int    `json:"day_of_week"` // 0=Sunday, 6=Saturday
	StartTime      string `json:"start_time"`  // HH:MM format
	EndTime        string `json:"end_time"`    // HH:MM format
	IsActive       bool   `json:"is_active"`
}

// Favorite (รายการโปรด)
type Favorite struct {
	FavoriteID int       `json:"favorite_id"`
	ClientID   int       `json:"client_id"`
	ProviderID int       `json:"provider_id"`
	CreatedAt  time.Time `json:"created_at"`
}

// ProviderSchedule (ตารางงาน/คิวของ Provider)
type ProviderSchedule struct {
	ScheduleID       int       `json:"schedule_id"`
	ProviderID       int       `json:"provider_id"`
	BookingID        *int      `json:"booking_id"`          // ถ้ามีการจองจะมี booking_id
	StartTime        time.Time `json:"start_time"`          // เวลาเริ่มต้น
	EndTime          time.Time `json:"end_time"`            // เวลาสิ้นสุด
	Status           string    `json:"status"`              // available, booked, blocked
	LocationType     *string   `json:"location_type"`       // Incall, Outcall, Both
	LocationAddress  *string   `json:"location_address"`    // ที่อยู่
	LocationProvince *string   `json:"location_province"`   // จังหวัด
	LocationDistrict *string   `json:"location_district"`   // เขต/อำเภอ
	Latitude         *float64  `json:"latitude"`            // พิกัด
	Longitude        *float64  `json:"longitude"`           // พิกัด
	Notes            *string   `json:"notes"`               // หมายเหตุ (เช่น "At spa", "Available for outcall")
	IsVisibleToAdmin bool      `json:"is_visible_to_admin"` // Admin/GOD มองเห็นได้
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// ProviderScheduleWithDetails (สำหรับ Admin/GOD view พร้อมข้อมูล Provider)
type ProviderScheduleWithDetails struct {
	ScheduleID       int       `json:"schedule_id"`
	ProviderID       int       `json:"provider_id"`
	ProviderUsername string    `json:"provider_username"`
	ProviderPhone    *string   `json:"provider_phone"`
	BookingID        *int      `json:"booking_id"`
	ClientID         *int      `json:"client_id"`       // ถ้ามี booking
	ClientUsername   *string   `json:"client_username"` // ถ้ามี booking
	StartTime        time.Time `json:"start_time"`
	EndTime          time.Time `json:"end_time"`
	Status           string    `json:"status"`
	LocationType     *string   `json:"location_type"`
	LocationAddress  *string   `json:"location_address"`
	LocationProvince *string   `json:"location_province"`
	LocationDistrict *string   `json:"location_district"`
	Latitude         *float64  `json:"latitude"`
	Longitude        *float64  `json:"longitude"`
	Notes            *string   `json:"notes"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
