package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

// !!! เปลี่ยนเป็นชื่อ GCS Bucket จริงของคุณ !!!
const GCS_BUCKET_NAME = "sex-worker-bucket"

func main() {
	ctx := context.Background()

	// --- 0. Load Environment Variables ---
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, using system environment variables")
	}

	// --- 1. Setup Stripe ---
	// (ต้องตั้งค่า ENV VARS: STRIPE_SECRET_KEY, STRIPE_WEBHOOK_SECRET)
	setupStripe()
	fmt.Println("✅ Stripe client initialized.")

	// --- 2. Connect to Databases ---
	// (ใช้การตั้งค่าจาก docker-compose.yml)
	dbPool, err := pgxpool.New(ctx, "postgres://admin:mysecretpassword@localhost:5432/skillmatch_db?sslmode=disable")
	if err != nil {
		log.Fatalf("ไม่สามารถเชื่อมต่อ PostgreSQL ได้: %v\n", err)
	}
	defer dbPool.Close()
	if err = dbPool.Ping(ctx); err != nil {
		log.Fatalf("ไม่สามารถ Ping PostgreSQL ได้: %v\n", err)
	}
	fmt.Println("✅ เชื่อมต่อ PostgreSQL สำเร็จ!")

	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379", Password: "", DB: 0})
	if _, err = rdb.Ping(ctx).Result(); err != nil {
		log.Fatalf("ไม่สามารถเชื่อมต่อ Redis ได้: %v\n", err)
	}
	fmt.Println("✅ เชื่อมต่อ Redis สำเร็จ!")

	// --- 3. Connect to Google Cloud Storage (Optional for Development) ---
	// (ต้องตั้งค่า ENV VAR: GOOGLE_APPLICATION_CREDENTIALS)
	var storageClient *storage.Client
	storageClient, err = storage.NewClient(ctx)
	if err != nil {
		log.Printf("⚠️  Failed to create Google Storage client: %v\n", err)
		log.Println("⚠️  Running in DEVELOPMENT MODE without GCS (file uploads will fail)")
		log.Println("⚠️  To enable GCS, set GOOGLE_APPLICATION_CREDENTIALS environment variable")
		storageClient = nil // Set to nil to indicate GCS is unavailable
	} else {
		defer storageClient.Close()
		fmt.Println("✅ เชื่อมต่อ Google Cloud Storage สำเร็จ!")
	}

	// --- 4. Initialize Global Database Connection ---
	// (for message, notification, report handlers)
	if err := InitDatabase("postgres://admin:mysecretpassword@localhost:5432/skillmatch_db?sslmode=disable"); err != nil {
		log.Fatalf("Failed to initialize database: %v\n", err)
	}
	defer db.Close()

	// --- 5. Initialize WebSocket Manager ---
	InitWebSocketManager()
	fmt.Println("✅ WebSocket manager initialized")

	// --- 6. Run Migrations (from migrations.go) ---
	runMigrations(dbPool, ctx)

	// --- 7. Setup Gin Router ---
	router := gin.Default()

	// --- 8. Apply CORS Middleware (Allow React App) ---
	// (นี่คือการตั้งค่าที่ถูกต้องสำหรับ development)
	router.Use(cors.New(cors.Config{
		AllowAllOrigins:  true, // (สำหรับ Development)
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"}, // (อนุญาต Authorization header)
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Add Content-Type middleware
	router.Use(func(c *gin.Context) {
		c.Header("Content-Type", "application/json; charset=utf-8")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Next()
	})

	// --- 9. Define Routes ---

	// Public Routes
	router.GET("/ping", func(c *gin.Context) {
		var pgTime time.Time
		dbPool.QueryRow(ctx, "SELECT NOW()").Scan(&pgTime)
		c.JSON(http.StatusOK, gin.H{"message": "pong!", "postgres_time": pgTime})
	})
	router.GET("/tiers", getTiersHandler(dbPool, ctx)) // (from tier_handlers.go)

	// Authentication & Registration Routes
	router.POST("/auth/send-verification", sendVerificationHandler(dbPool, ctx)) // ส่ง OTP ไปทาง email (from email_verification.go)
	router.POST("/auth/verify-email", verifyEmailHandler(dbPool, ctx))           // ยืนยัน OTP (from email_verification.go)
	router.POST("/register", registerWithVerificationHandler(dbPool, ctx))       // สมัครสมาชิก (User - ต้องยืนยัน OTP ก่อน) (from email_verification.go)
	router.POST("/register/provider", registerProviderHandler(dbPool, ctx))      // สมัครเป็น Provider (ต้องส่งเอกสาร) (from provider_system_handlers.go)
	router.POST("/register/basic", createUserHandler(dbPool, ctx))               // สมัครแบบไม่ต้องยืนยัน email (เก่า - deprecated)
	router.POST("/login", loginHandler(dbPool, ctx))                             // (from auth_handlers.go)
	router.POST("/auth/google", handleGoogleCallback(dbPool, ctx))               // (from auth_handlers.go)

	router.POST("/payment/webhook", paymentWebhookHandler(dbPool, ctx)) // (from payment_handlers.go)

	// WebSocket endpoint (public - authenticates via message)
	router.GET("/ws", HandleWebSocket) // WebSocket connection with message-based auth

	// Protected Routes (ต้อง Login)
	protected := router.Group("/")
	protected.Use(authMiddleware()) // (from middleware.go)
	{
		// User Routes
		protected.GET("/users/me", getMeHandler(dbPool, ctx))    // (from user_handlers.go)
		protected.GET("/profile", getMeHandler(dbPool, ctx))     // Alias for /users/me (Frontend compatibility)
		protected.GET("/users/:id", getUserHandler(dbPool, ctx)) // (from user_handlers.go)

		// Browse Routes
		protected.GET("/browse/users", browseUsersHandler(dbPool, ctx)) // (from browse_handlers.go)

		// Verification (KYC) Routes
		protected.POST("/verification/start", startVerificationHandler(dbPool, storageClient, GCS_BUCKET_NAME, ctx)) // (from verification_handlers.go)
		protected.POST("/verification/submit", submitVerificationHandler(dbPool, ctx))                               // (from verification_handlers.go)

		// Photo Gallery Routes
		protected.GET("/photos/me", getMyPhotosHandler(dbPool, ctx))                                          // (from photo_handlers.go)
		protected.POST("/photos/start", startPhotoUploadHandler(dbPool, storageClient, GCS_BUCKET_NAME, ctx)) // (from photo_handlers.go)
		protected.POST("/photos/submit", submitPhotoUploadHandler(dbPool, ctx))                               // (from photo_handlers.go)
		protected.DELETE("/photos/:photoId", deletePhotoHandler(dbPool, ctx))                                 // (from photo_handlers.go)

		// Subscription Routes
		protected.POST("/subscription/create-checkout", createCheckoutSessionHandler(dbPool, ctx)) // (from payment_handlers.go)

		// Profile Routes (Edit/View)
		protected.GET("/profile/me", getMyProfileHandler(dbPool, ctx))    // (from profile_handlers.go)
		protected.PUT("/profile/me", updateMyProfileHandler(dbPool, ctx)) // (from profile_handlers.go)

		// 🆕 Service Category Routes - MUST BE BEFORE /provider/:userId
		protected.PUT("/provider/me/categories", updateProviderCategoriesHandler(dbPool, ctx)) // อัพเดทหมวดหมู่ของตัวเอง

		// 🆕 Booking Routes
		protected.POST("/packages", createPackageHandler(dbPool, ctx))                                // สร้างแพ็คเกจ (provider)
		protected.POST("/bookings", createBookingHandler(dbPool, ctx))                                // จองบริการ (ไม่มีการชำระเงิน)
		protected.POST("/bookings/create-with-payment", createBookingWithPaymentHandler(dbPool, ctx)) // 🆕 จองบริการพร้อมชำระเงิน
		protected.GET("/bookings/my", getMyBookingsHandler(dbPool, ctx))                              // ดูการจองของตัวเอง (client)
		protected.GET("/bookings/provider", getProviderBookingsHandler(dbPool, ctx))                  // ดูการจองที่เข้ามา (provider)
		protected.PATCH("/bookings/:id/status", updateBookingStatusHandler(dbPool, ctx))              // อัพเดทสถานะการจอง

		// 🆕 Review Routes
		protected.POST("/reviews", createReviewHandler(dbPool, ctx)) // สร้างรีวิว

		// 🆕 Favorite Routes
		protected.POST("/favorites", addFavoriteHandler(dbPool, ctx))                  // เพิ่มรายการโปรด
		protected.DELETE("/favorites/:providerId", removeFavoriteHandler(dbPool, ctx)) // ลบรายการโปรด
		protected.GET("/favorites", getMyFavoritesHandler(dbPool, ctx))                // ดูรายการโปรด

		// 🆕 Messaging Routes
		protected.GET("/conversations", GetConversations)                     // รายการ conversations
		protected.GET("/conversations/:id/messages", GetConversationMessages) // ข้อความใน conversation
		protected.POST("/messages", SendMessage)                              // ส่งข้อความ
		protected.PATCH("/messages/read", MarkMessagesAsRead)                 // อ่านข้อความแล้ว
		protected.DELETE("/messages/:id", DeleteMessage)                      // ลบข้อความ

		// 🆕 Notification Routes
		protected.GET("/notifications", GetNotifications)                        // รายการแจ้งเตือน
		protected.GET("/notifications/unread/count", GetUnreadNotificationCount) // จำนวนที่ยังไม่อ่าน
		protected.PATCH("/notifications/:id/read", MarkNotificationAsRead)       // อ่านแจ้งเตือนแล้ว
		protected.PATCH("/notifications/read-all", MarkAllNotificationsAsRead)   // อ่านทั้งหมดแล้ว
		protected.DELETE("/notifications/:id", DeleteNotification)               // ลบแจ้งเตือน
		protected.DELETE("/notifications", DeleteAllNotifications)               // ลบทั้งหมด

		// 🆕 Report Routes
		protected.POST("/reports", CreateReport)   // สร้างรายงาน
		protected.GET("/reports/my", GetMyReports) // ดูรายงานของตัวเอง

		// 🆕 Analytics Routes (Provider)
		protected.GET("/analytics/provider/dashboard", getProviderDashboardHandler(dbPool, ctx)) // Overview dashboard
		protected.GET("/analytics/provider/bookings", getBookingStatsHandler(dbPool, ctx))       // Booking stats by date
		protected.GET("/analytics/provider/revenue", getRevenueBreakdownHandler(dbPool, ctx))    // Revenue by package
		protected.GET("/analytics/provider/ratings", getRatingBreakdownHandler(dbPool, ctx))     // Rating distribution
		protected.GET("/analytics/provider/monthly", getMonthlyStatsHandler(dbPool, ctx))        // Monthly summary
		protected.POST("/analytics/profile-view", trackProfileViewHandler(dbPool, ctx))          // Track profile view

		// 🆕 Block User Routes
		protected.POST("/blocks", blockUserHandler(dbPool, ctx))                     // Block a user
		protected.DELETE("/blocks/:userId", unblockUserHandler(dbPool, ctx))         // Unblock a user
		protected.GET("/blocks", getBlockedUsersHandler(dbPool, ctx))                // Get blocked users list
		protected.GET("/blocks/check/:userId", checkBlockStatusHandler(dbPool, ctx)) // Check if user is blocked

		// 🆕 Financial System Routes - User (Provider)
		protected.POST("/bank-accounts", addBankAccountHandler(dbPool, ctx))                       // เพิ่มบัญชีธนาคาร
		protected.GET("/bank-accounts", getMyBankAccountsHandler(dbPool, ctx))                     // ดูบัญชีธนาคารของตัวเอง
		protected.DELETE("/bank-accounts/:bank_account_id", deleteBankAccountHandler(dbPool, ctx)) // ลบบัญชีธนาคาร
		protected.GET("/wallet", getMyWalletHandler(dbPool, ctx))                                  // ดู wallet ของตัวเอง
		protected.POST("/withdrawals", requestWithdrawalHandler(dbPool, ctx))                      // ขอถอนเงิน
		protected.GET("/withdrawals", getMyWithdrawalsHandler(dbPool, ctx))                        // ดูประวัติการถอนเงิน
		protected.GET("/transactions", getMyTransactionsHandler(dbPool, ctx))                      // ดูประวัติธุรกรรม

		// 🆕 Provider Document & Verification System
		protected.POST("/provider/documents", uploadProviderDocumentHandler(dbPool, ctx))     // อัปโหลดเอกสาร (from provider_system_handlers.go)
		protected.GET("/provider/documents", getMyDocumentsHandler(dbPool, ctx))              // ดูเอกสารของตัวเอง (from provider_system_handlers.go)
		protected.GET("/provider/categories/me", getMyProviderCategoriesHandler(dbPool, ctx)) // ดูหมวดหมู่บริการของตัวเอง (from provider_system_handlers.go)

		// 🆕 Face Verification System (from face_verification_handlers.go)
		protected.POST("/provider/face-verification", submitFaceVerificationHandler(dbPool, ctx)) // อัปโหลด selfie สำหรับ face matching
		protected.GET("/provider/face-verification", getMyFaceVerificationHandler(dbPool, ctx))   // ดูสถานะ face verification

		// 🆕 Provider Tier Management
		protected.GET("/provider/my-tier", getMyProviderTierHandler(dbPool, ctx))     // ดู Tier ปัจจุบันของตัวเอง (from provider_tier_handlers.go)
		protected.GET("/provider/tier-history", getMyTierHistoryHandler(dbPool, ctx)) // ดูประวัติการเปลี่ยน Tier (from provider_tier_handlers.go)

		// 🆕 Provider Schedule Management (from schedule_handlers.go)
		protected.POST("/provider/schedule", createScheduleHandler(dbPool, ctx))               // สร้างตารางงาน
		protected.GET("/provider/schedule/me", getMySchedulesHandler(dbPool, ctx))             // ดูตารางงานของตัวเอง
		protected.PATCH("/provider/schedule/:scheduleId", updateScheduleHandler(dbPool, ctx))  // แก้ไขตารางงาน
		protected.DELETE("/provider/schedule/:scheduleId", deleteScheduleHandler(dbPool, ctx)) // ลบตารางงาน
	}

	// Admin Routes (ต้อง Login และเป็น Admin)
	admin := router.Group("/admin")
	admin.Use(authMiddleware())
	admin.Use(adminAuthMiddleware(dbPool, ctx)) // (from admin_middleware.go)
	{
		admin.GET("/pending-users", getPendingUsersHandler(dbPool, ctx))
		admin.GET("/kyc-details/:userId", getKycDetailsHandler(dbPool, ctx))
		admin.POST("/approve/:userId", approveUserHandler(dbPool, ctx))
		admin.POST("/reject/:userId", rejectUserHandler(dbPool, ctx))
		admin.GET("/kyc-file-url", getKycFileUrlHandler(storageClient, GCS_BUCKET_NAME, ctx))
		admin.POST("/users", adminCreateUserHandler(dbPool, ctx))

		// 🆕 Admin Report Management
		admin.GET("/reports", GetAllReports)            // ดูรายงานทั้งหมด
		admin.PATCH("/reports/:id", UpdateReportStatus) // อัพเดทสถานะรายงาน
		admin.DELETE("/reports/:id", DeleteReport)      // ลบรายงาน

		// 🆕 GOD Dashboard & User Management
		admin.GET("/stats/god", getGodStatsHandler(dbPool, ctx))          // GOD statistics
		admin.GET("/users", listAllUsersHandler(dbPool, ctx))             // List all users
		admin.GET("/admins", listAdminsHandler(dbPool, ctx))              // List all admins (GOD only)
		admin.POST("/admins", createAdminHandler(dbPool, ctx))            // Create admin (GOD only)
		admin.DELETE("/admins/:user_id", deleteAdminHandler(dbPool, ctx)) // Delete admin (GOD only)
		admin.DELETE("/users/:user_id", deleteUserHandler(dbPool, ctx))   // Delete any user (GOD only)

		// 🆕 Financial System Routes - Admin
		admin.GET("/withdrawals", adminGetPendingWithdrawalsHandler(dbPool, ctx))                        // ดูคำขอถอนเงินทั้งหมด
		admin.POST("/withdrawals/:withdrawal_id/process", adminProcessWithdrawalHandler(dbPool, ctx))    // อนุมัติ/ปฏิเสธ/complete การถอน
		admin.POST("/bank-accounts/:bank_account_id/verify", adminVerifyBankAccountHandler(dbPool, ctx)) // ยืนยันบัญชีธนาคาร
		admin.GET("/financial/summary", adminGetFinancialSummaryHandler(dbPool, ctx))                    // สรุปรายได้/ค่าคอมฯ
		admin.POST("/financial/reports", adminGenerateFinancialReportHandler(dbPool, ctx))               // สร้างรายงานทางการเงิน
		admin.GET("/commission-rules", adminGetCommissionRulesHandler(dbPool, ctx))                      // ดูกฎค่าคอมมิชชั่น
		admin.PUT("/commission-rules/:rule_id", adminUpdateCommissionRuleHandler(dbPool, ctx))           // แก้ไขกฎค่าคอมมิชชั่น
		admin.GET("/wallets/:user_id", adminGetUserWalletHandler(dbPool, ctx))                           // ดู wallet ของ user
		admin.POST("/wallets/:user_id/adjust", adminAdjustWalletHandler(dbPool, ctx))                    // ปรับยอด wallet (bonus/penalty)

		// 🆕 Admin Provider Management
		admin.GET("/providers/pending", getAdminPendingProvidersHandler(dbPool, ctx))        // ดู providers ที่รอตรวจสอบ (from provider_system_handlers.go)
		admin.PATCH("/verify-document/:documentId", adminVerifyDocumentHandler(dbPool, ctx)) // อนุมัติ/ปฏิเสธเอกสาร (from provider_system_handlers.go)
		admin.PATCH("/approve-provider/:userId", adminApproveProviderHandler(dbPool, ctx))   // อนุมัติ provider (from provider_system_handlers.go)
		admin.GET("/provider-stats", getAdminProviderStatsHandler(dbPool, ctx))              // สถิติ providers (from provider_system_handlers.go)

		// 🆕 Admin Provider Tier Management
		admin.POST("/recalculate-provider-tiers", adminRecalculateProviderTiersHandler(dbPool, ctx)) // คำนวณ Tier อัตโนมัติทั้งหมด (from provider_tier_handlers.go)
		admin.PATCH("/set-provider-tier/:userId", adminSetProviderTierHandler(dbPool, ctx))          // เปลี่ยน Tier แบบ Manual (from provider_tier_handlers.go)
		admin.GET("/provider/:userId/tier-details", adminGetProviderTierDetailsHandler(dbPool, ctx)) // ดูรายละเอียด Tier (from provider_tier_handlers.go)

		// 🆕 Admin Face Verification Management (from face_verification_handlers.go)
		admin.GET("/face-verifications", adminListFaceVerificationsHandler(dbPool, ctx))                           // ดู face verifications ทั้งหมด
		admin.PATCH("/face-verification/:verificationId", adminReviewFaceVerificationHandler(dbPool, ctx))         // อนุมัติ/ปฏิเสธ face verification
		admin.POST("/face-verification/:verificationId/trigger-matching", triggerFaceMatchingHandler(dbPool, ctx)) // เรียก Face Matching API

		// 🆕 Admin Schedule Viewing (from schedule_handlers.go)
		admin.GET("/schedules/provider/:providerId", getProviderScheduleAdminHandler(dbPool, ctx)) // ดูตารางงานของ Provider คนใดคนหนึ่ง
		admin.GET("/schedules/all", getAllProvidersScheduleAdminHandler(dbPool, ctx))              // ดูตารางงานของ Providers ทั้งหมด
	}

	// GOD Routes (ต้อง Login และเป็น GOD tier 5)
	god := router.Group("/god")
	god.Use(authMiddleware())
	{
		// View Mode Switching (UI simulation - doesn't modify DB)
		god.POST("/view-mode", setGodViewModeHandler(dbPool, ctx)) // Set GOD view mode (user/provider/admin)
		god.GET("/view-mode", getGodViewModeHandler(dbPool, ctx))  // Get current view mode

		// User Management (modifies actual user data in DB)
		god.POST("/update-user", updateUserHandler(dbPool, ctx))      // Update any user's role/tier
		god.DELETE("/users/:user_id", deleteUserHandler(dbPool, ctx)) // Delete any user (except GOD)
	}

	// 🆕 Service Category Public Routes
	router.GET("/service-categories", listServiceCategoriesHandler(dbPool, ctx))                    // ดูหมวดหมู่ทั้งหมด (Public)
	router.GET("/categories/:category_id/providers", browseProvidersByCategoryHandler(dbPool, ctx)) // ดูผู้ให้บริการในหมวดหมู่

	// 🆕 Browse Search with Filters (Public)
	router.GET("/browse/search", browseSearchHandler(dbPool, ctx)) // ⬅️ NEW: Advanced search with all filters

	// 🆕 Provider Public Profile Routes (No auth required - anyone can view)
	// Public routes - ข้อมูลจำกัด (ไม่แสดง Age, Height, Weight, ServiceType, etc.)
	router.GET("/provider/:userId/public", getPublicProfileHandler(dbPool, ctx))         // ดู profile แบบจำกัด (ไม่ต้อง login)
	router.GET("/provider/:userId/photos", getProviderPhotosHandler(dbPool, ctx))        // ดูรูปภาพของผู้ให้บริการ (Public)
	router.GET("/packages/:providerId", getProviderPackagesHandler(dbPool, ctx))         // ดูแพ็คเกจของ provider (Public)
	router.GET("/reviews/:providerId", getProviderReviewsHandler(dbPool, ctx))           // ดูรีวิวของ provider (Public)
	router.GET("/reviews/stats/:providerId", getProviderReviewStatsHandler(dbPool, ctx)) // สถิติรีวิว (Public)
	router.GET("/favorites/check/:providerId", checkFavoriteHandler(dbPool, ctx))        // เช็ค favorite (Public - optional auth)

	// Protected routes - ข้อมูลเต็มรูปแบบ (ต้อง login)
	protected.GET("/provider/:userId", getAuthenticatedProfileHandler(dbPool, ctx))           // ดู profile เต็มรูปแบบ (ต้อง login)
	protected.GET("/browse/v2", browseUsersHandlerV2(dbPool, ctx))                            // Browse providers (ต้อง login)
	protected.GET("/providers/:userId/categories", getProviderCategoriesHandler(dbPool, ctx)) // ดูหมวดหมู่ของผู้ให้บริการ

	// --- 10. Start Server ---
	fmt.Println("🚀 เซิร์ฟเวอร์กำลังทำงานที่ http://localhost:8080")
	router.Run(":8080")
}
