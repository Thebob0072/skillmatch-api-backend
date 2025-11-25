package main

import "time"

// Gender (ตาราง Lookup)
type Gender struct {
	GenderID   int    `json:"gender_id"`
	GenderName string `json:"gender_name"`
}

// Tiers (ตารางระดับสมาชิก)
type Tier struct {
	TierID       int     `json:"tier_id"`
	Name         string  `json:"name"`
	AccessLevel  int     `json:"access_level"`
	PriceMonthly float64 `json:"price_monthly"`
}

// User (ตารางหลัก - ฉบับสมบูรณ์)
type User struct {
	UserID             int       `json:"user_id"`
	Username           string    `json:"username"`
	Email              string    `json:"email"`
	PasswordHash       *string   `json:"-"` // (ซ่อน)
	GenderID           int       `json:"gender_id"`
	TierID             int       `json:"subscription_tier_id"` // Subscription Tier (จากตาราง users)
	TierName           string    `json:"tier_name"`            // Tier Name (จาก JOIN กับ tiers)
	ProviderLevelID    int       `json:"provider_level_id"`    // Provider Level (จากตาราง users)
	RegistrationDate   time.Time `json:"registration_date"`
	GoogleID           *string   `json:"-"` // (ซ่อน)
	PhoneNumber        *string   `json:"phone_number"`
	VerificationStatus string    `json:"verification_status"`
	IsAdmin            bool      `json:"is_admin"`
	FirstName          *string   `json:"first_name"`
	LastName           *string   `json:"last_name"`

	// (ข้อมูลจาก JOIN)
	GoogleProfilePicture *string  `json:"google_profile_picture"` // รูปจาก Google
	Bio                  *string  `json:"bio"`
	Location             *string  `json:"location"`
	Skills               []string `json:"skills"`
	ProfileImageUrl      *string  `json:"profile_image_url"` // รูปที่อัปโหลดเอง
	Age                  *int     `json:"age"`               // อายุ (จาก user_profiles)
}

// UserPhoto (ตารางรูปโปรไฟล์แกลเลอรี)
type UserPhoto struct {
	PhotoID    int       `json:"photo_id"`
	UserID     int       `json:"user_id"`
	PhotoURL   string    `json:"photo_url"`
	SortOrder  int       `json:"sort_order"`
	UploadedAt time.Time `json:"uploaded_at"`
}

// UserVerification (ตาราง KYC)
type UserVerification struct {
	VerificationID int        `json:"verification_id"`
	UserID         int        `json:"user_id"`
	NationalIDUrl  *string    `json:"national_id_url"`
	HealthCertUrl  *string    `json:"health_cert_url"`
	FaceScanUrl    *string    `json:"face_scan_url"`
	SubmittedAt    *time.Time `json:"submitted_at"`
}

// UserProfile (ตารางเก็บข้อมูลสาธารณะที่ผู้ใช้กรอกเอง)
type UserProfile struct {
	UserID               int       `json:"user_id"`
	Bio                  *string   `json:"bio"`
	Location             *string   `json:"location"` // Legacy field (deprecated)
	Skills               []string  `json:"skills"`
	ProfileImageUrl      *string   `json:"profile_image_url"`
	GoogleProfilePicture *string   `json:"google_profile_picture"` // รูปจาก Google
	UpdatedAt            time.Time `json:"updated_at"`
	// ข้อมูลเพิ่มเติม
	Age          *int     `json:"age"`
	Height       *int     `json:"height"` // cm
	Weight       *int     `json:"weight"` // kg
	Ethnicity    *string  `json:"ethnicity"`
	Languages    []string `json:"languages"`
	WorkingHours *string  `json:"working_hours"` // เช่น "9:00-22:00"
	IsAvailable  bool     `json:"is_available"`  // ว่างหรือไม่
	ServiceType  *string  `json:"service_type"`  // incall, outcall, both

	// ข้อมูลที่อยู่แบบละเอียด (Address Details)
	Province     *string  `json:"province"`      // จังหวัด เช่น "กรุงเทพมหานคร"
	District     *string  `json:"district"`      // เขต/อำเภอ เช่น "บางรัก"
	SubDistrict  *string  `json:"sub_district"`  // แขวง/ตำบล เช่น "สีลม"
	PostalCode   *string  `json:"postal_code"`   // รหัสไปรษณีย์ เช่น "10500"
	AddressLine1 *string  `json:"address_line1"` // บ้านเลขที่ ถนน ซอย
	Latitude     *float64 `json:"latitude"`      // พิกัด GPS (latitude)
	Longitude    *float64 `json:"longitude"`     // พิกัด GPS (longitude)
}

// BrowsableUser (สำหรับหน้า Browse)
type BrowsableUser struct {
	UserID               int     `json:"user_id"`
	Username             string  `json:"username"`
	TierName             string  `json:"tier_name"`
	GenderID             int     `json:"gender_id"`
	ProfileImageUrl      *string `json:"profile_image_url"`
	GoogleProfilePicture *string `json:"google_profile_picture"`
	// เพิ่มข้อมูลสำหรับ filtering
	Age           *int     `json:"age"`
	Location      *string  `json:"location"` // Legacy field
	IsAvailable   bool     `json:"is_available"`
	AverageRating float64  `json:"average_rating"`
	ReviewCount   int      `json:"review_count"`
	MinPrice      *float64 `json:"min_price"` // ราคาต่ำสุด

	// ข้อมูลที่อยู่แบบละเอียด
	Province    *string  `json:"province"`     // จังหวัด
	District    *string  `json:"district"`     // เขต/อำเภอ
	SubDistrict *string  `json:"sub_district"` // แขวง/ตำบล
	Latitude    *float64 `json:"latitude"`     // พิกัด GPS
	Longitude   *float64 `json:"longitude"`    // พิกัด GPS
	Distance    *float64 `json:"distance_km"`  // ระยะทาง (กิโลเมตร) - คำนวณจาก user location

	// ประเภทการให้บริการ
	ServiceType *string `json:"service_type"` // incall, outcall
}

// PublicProfile (สำหรับหน้า Profile สาธารณะ - ซ่อนข้อมูลที่บ่งบอกว่าขายบริการ)
type PublicProfile struct {
	UserID               int      `json:"user_id"`
	Username             string   `json:"username"`
	GenderID             int      `json:"gender_id"`
	TierName             string   `json:"tier_name"`
	Bio                  *string  `json:"bio"`
	Location             *string  `json:"location"` // Legacy field
	Skills               []string `json:"skills"`
	ProfileImageUrl      *string  `json:"profile_image_url"`
	GoogleProfilePicture *string  `json:"google_profile_picture"`
	IsAvailable          bool     `json:"is_available"`
	AverageRating        float64  `json:"average_rating"`
	ReviewCount          int      `json:"review_count"`

	// ข้อมูลที่อยู่ (ระดับกว้าง - ไม่เฉพาะเจาะจง)
	Province    *string `json:"province"`
	District    *string `json:"district"`
	SubDistrict *string `json:"sub_district"`
	// Note: ซ่อน Age, Height, Weight, Ethnicity, Languages, WorkingHours, ServiceType, AddressLine1, Latitude, Longitude
}

// FullProfile (สำหรับผู้ใช้ที่ login แล้ว - แสดงข้อมูลเต็มรูปแบบ)
type FullProfile struct {
	UserID               int      `json:"user_id"`
	Username             string   `json:"username"`
	GenderID             int      `json:"gender_id"`
	TierName             string   `json:"tier_name"`
	Bio                  *string  `json:"bio"`
	Location             *string  `json:"location"` // Legacy field
	Skills               []string `json:"skills"`
	ProfileImageUrl      *string  `json:"profile_image_url"`
	GoogleProfilePicture *string  `json:"google_profile_picture"`
	IsAvailable          bool     `json:"is_available"`
	AverageRating        float64  `json:"average_rating"`
	ReviewCount          int      `json:"review_count"`

	// ข้อมูลที่อยู่
	Province     *string  `json:"province"`
	District     *string  `json:"district"`
	SubDistrict  *string  `json:"sub_district"`
	AddressLine1 *string  `json:"address_line1"`
	Latitude     *float64 `json:"latitude"`
	Longitude    *float64 `json:"longitude"`

	// ข้อมูลละเอียด (แสดงเฉพาะผู้ใช้ที่ login)
	Age          *int     `json:"age"`
	Height       *int     `json:"height"`
	Weight       *int     `json:"weight"`
	Ethnicity    *string  `json:"ethnicity"`
	Languages    []string `json:"languages"`
	WorkingHours *string  `json:"working_hours"`
	ServiceType  *string  `json:"service_type"`
}
