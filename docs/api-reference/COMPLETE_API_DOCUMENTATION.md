# 📘 SkillMatch API - Complete System Documentation for Frontend

## 🎯 System Overview

SkillMatch API เป็นระบบ Marketplace สำหรับ Service Providers มี 3 ประเภทผู้ใช้หลัก:

### 👥 User Types

#### 1. **User (ผู้ใช้บริการทั่วไป)**
- ✅ ลงทะเบียนง่ายด้วย Email OTP
- ❌ **ไม่ต้องส่งเอกสาร**
- ✅ Browse providers, จองบริการ, รีวิว
- 💳 เลือก Subscription Tier: Free หรือ Premium (เสียเงิน)
- 🔒 **ไม่สามารถขายบริการได้**

#### 2. **Provider (ผู้ให้บริการ)**
- ✅ ลงทะเบียนเป็น Provider + ระบุหมวดหมู่บริการ
- ⚠️  **ต้องส่งเอกสารยืนยันตัวตน** (บัตรประชาชน, ใบรับรองสุขภาพ)
- ⏳ รอ Admin ตรวจสอบเอกสาร
- 📊 มี **Provider Tier** (General/Silver/Diamond/Premium) คำนวณจาก Rating + Performance
- 💰 สร้างแพ็คเกจ, รับการจอง, รับเงินผ่าน Wallet
- 🎖️  **2 Tiers**: Subscription Tier (ซื้อได้) + Provider Tier (ระบบคำนวณอัตโนมัติ)

#### 3. **Admin**
- ✅ ตรวจสอบและอนุมัติ Provider documents
- ✅ จัดการผู้ใช้ทั้งหมด
- ✅ อนุมัติการถอนเงิน (Withdrawals)
- ✅ เปลี่ยน Provider Tier แบบ Manual (ถ้าจำเป็น)
- ✅ ดูรายงานทางการเงิน

---

## 🔐 Authentication System

### Email OTP Verification
```
POST /auth/send-verification      → ส่ง OTP 6 หลักไปทาง email (หมดอายุ 10 นาที)
POST /auth/verify-email           → ยืนยัน OTP (optional step)
POST /register                    → User registration (ต้องมี OTP)
POST /register/provider           → Provider registration (ต้องมี OTP + ข้อมูล Provider)
POST /login                       → Login ด้วย email/password
POST /auth/google                 → Google OAuth login
```

### User Registration vs Provider Registration

| Field | User | Provider |
|-------|------|----------|
| username, email, password | ✅ | ✅ |
| gender_id, first_name, last_name | ✅ | ✅ |
| phone | ✅ | ✅ |
| otp (6-digit) | ✅ | ✅ |
| **category_ids** (array) | ❌ | ✅ Required |
| **service_type** ("Incall"/"Outcall"/"Both") | ❌ | ✅ |
| **bio** | ❌ | ✅ |
| **province, district** | ❌ | ✅ |

---

## 💼 Provider System

### 1. Provider Lifecycle

```
1. Register as Provider
   ↓
2. Upload Documents (บัตรประชาชน + ใบรับรองสุขภาพ)
   ↓
3. Admin Review & Approve
   ↓
4. Provider Approved ✅
   ↓
5. Start Creating Packages & Receiving Bookings
   ↓
6. Auto Tier Assignment (ตามคะแนน Performance)
```

### 2. Required Documents

| Document Type | Required | Description |
|--------------|----------|-------------|
| `national_id` | ✅ Yes | บัตรประชาชน |
| `health_certificate` | ✅ Yes | ใบรับรองสุขภาพ (ไม่เกิน 6 เดือน) |
| `business_license` | ⚪ Optional | ใบอนุญาตธุรกิจ |
| `portfolio` | ⚪ Optional | ผลงาน/รูปตัวอย่าง |
| `certification` | ⚪ Optional | ใบประกาศนียบัตร |

**API:**
```
POST /provider/documents          → Upload document (ต้อง login as Provider)
GET /provider/documents           → Get my documents with status
```

### 3. Provider Tier System

#### Tier Calculation Algorithm
```
Total Points (max 600) = 
  + (average_rating * 20)              = 0-100 points
  + (completed_bookings * 5)           = 0-250 points (max 50 bookings)
  + (total_reviews * 3)                = 0-150 points (max 50 reviews)
  + (response_rate * 0.5)              = 0-50 points
  + (acceptance_rate * 0.5)            = 0-50 points
```

#### Tier Levels

| Tier | Points | Benefits |
|------|--------|----------|
| **General** | 0-99 | Basic visibility |
| **Silver** | 100-249 | Higher search ranking |
| **Diamond** | 250-399 | Premium badge + Priority support |
| **Premium** | 400+ | Top ranking + Featured listings |

**API:**
```
GET /provider/my-tier              → Get current tier + points + next tier
GET /provider/tier-history         → Tier change history
```

### 4. Provider Categories

Providers เลือกได้ว่าให้บริการในหมวดหมู่ไหนบ้าง (เช่น Massage, Spa, Beauty, etc.)

**API:**
```
GET /service-categories                    → List all categories (Public)
GET /provider/categories/me                → My provider categories
PUT /provider/me/categories                → Update my categories
GET /categories/:category_id/providers     → Browse providers by category
```

---

## 🛒 Booking & Package System

### 1. Service Packages

Providers สร้างแพ็คเกจบริการ (เช่น "60-Minute Massage - 500 THB")

**API:**
```
POST /packages                     → Create package (Provider only)
GET /packages/:providerId          → Get provider's packages (Public)
```

### 2. Booking Flow

```
Client → Select Provider → Choose Package → Book → Pay → Receive Service → Review
```

**API:**
```
POST /bookings                     → Create booking (Client)
GET /bookings/my                   → My bookings as client
GET /bookings/provider             → Bookings received (Provider)
PATCH /bookings/:id/status         → Update booking status (Provider: accept/reject/complete)
```

**Booking Statuses:**
- `pending` → รอ Provider ตอบรับ
- `confirmed` → Provider ยอมรับแล้ว
- `completed` → บริการเสร็จสิ้น
- `cancelled` → ถูกยกเลิก

### 3. Reviews

Clients สามารถรีวิวหลังจากการจองเสร็จสิ้น

**API:**
```
POST /reviews                      → Create review (Client - after booking completed)
GET /reviews/:providerId           → Get provider reviews (Public)
GET /reviews/stats/:providerId     → Get rating stats (Public)
```

---

## 💰 Financial System

### 1. Wallet System

Providers มี Wallet สำหรับเก็บเงินจากการให้บริการ

**Wallet Structure:**
- `available_balance`: เงินที่ถอนได้ (หลังผ่าน 7 วันจากการจองเสร็จสิ้น)
- `pending_balance`: เงินที่รอ (ยังถอนไม่ได้)
- `total_earnings`: รายได้รวมทั้งหมด

**API:**
```
GET /wallet                        → Get my wallet (Provider)
GET /transactions                  → Get my transaction history (Provider)
```

### 2. Commission System

ระบบหัก **10% commission** จากแต่ละการจอง

**Example:**
```
Booking Price: 1000 THB
Commission (10%): 100 THB
Provider Earning: 900 THB
```

**Transactions Created:**
1. `booking_payment`: +1000 THB (จาก Client)
2. `commission`: -100 THB (ค่าคอมมิชชั่นระบบ)
3. `provider_earning`: +900 THB (รายได้ของ Provider)

### 3. Bank Accounts & Withdrawals

Providers เพิ่มบัญชีธนาคาร → ขอถอนเงิน → Admin อนุมัติ

**API:**
```
POST /bank-accounts                → Add bank account (Provider)
GET /bank-accounts                 → Get my bank accounts (Provider)
DELETE /bank-accounts/:id          → Delete bank account (Provider)

POST /withdrawals                  → Request withdrawal (Provider)
GET /withdrawals                   → Get my withdrawal requests (Provider)
```

**Admin Approval:**
```
GET /admin/withdrawals                    → Get all pending withdrawals
POST /admin/withdrawals/:id/process       → Approve/Reject withdrawal
```

**Withdrawal Statuses:**
- `pending` → รอ Admin ตรวจสอบ
- `approved` → Admin อนุมัติ (รอโอนเงิน)
- `completed` → โอนเงินเสร็จแล้ว
- `rejected` → ถูกปฏิเสธ

---

## 🔔 Messaging & Notifications

### 1. Real-Time Messaging (WebSocket)

**WebSocket Connection:**
```
ws://localhost:8080/ws
```

**Message Format:**
```json
{
  "type": "authenticate",
  "token": "Bearer eyJhbGci..."
}
```

**API:**
```
GET /conversations                 → List conversations
GET /conversations/:id/messages    → Get messages in conversation
POST /messages                     → Send message
PATCH /messages/read               → Mark messages as read
DELETE /messages/:id               → Delete message
```

### 2. Notifications

**API:**
```
GET /notifications                 → Get my notifications
GET /notifications/unread/count    → Count unread notifications
PATCH /notifications/:id/read      → Mark as read
PATCH /notifications/read-all      → Mark all as read
DELETE /notifications/:id          → Delete notification
```

---

## 🛡️ Admin Panel

### 1. Provider Management

**API:**
```
GET /admin/providers/pending             → Providers waiting for approval
PATCH /admin/verify-document/:docId      → Approve/Reject document
PATCH /admin/approve-provider/:userId    → Approve/Reject provider
GET /admin/provider-stats                → Provider statistics
```

### 2. Provider Tier Management

**API:**
```
POST /admin/recalculate-provider-tiers   → Recalculate all provider tiers
PATCH /admin/set-provider-tier/:userId   → Manually change provider tier
GET /admin/provider/:userId/tier-details → View provider tier details
```

### 3. Financial Management

**API:**
```
GET /admin/withdrawals                   → Pending withdrawals
POST /admin/withdrawals/:id/process      → Approve/Reject/Complete withdrawal
POST /admin/bank-accounts/:id/verify     → Verify bank account
GET /admin/financial/summary             → Financial summary
POST /admin/financial/reports            → Generate financial report
GET /admin/wallets/:user_id              → View user wallet
POST /admin/wallets/:user_id/adjust      → Adjust wallet (bonus/penalty)
```

### 4. User Management (GOD Tier Only)

**API:**
```
GET /god/view-mode                       → Get GOD view mode
POST /god/view-mode                      → Set GOD view mode (user/provider/admin)
POST /god/update-user                    → Update any user's role/tier

GET /admin/users                         → List all users
GET /admin/admins                        → List all admins (GOD only)
POST /admin/admins                       → Create admin (GOD only)
DELETE /admin/admins/:user_id            → Delete admin (GOD only)
GET /admin/stats/god                     → GOD statistics
```

---

## 📊 Analytics & Reports

### 1. Provider Analytics

**API:**
```
GET /analytics/provider/dashboard        → Provider overview dashboard
GET /analytics/provider/bookings         → Booking stats by date
GET /analytics/provider/revenue          → Revenue breakdown by package
GET /analytics/provider/ratings          → Rating distribution
GET /analytics/provider/monthly          → Monthly summary
POST /analytics/profile-view             → Track profile view
```

### 2. Profile Views

ระบบติดตามจำนวน profile views สำหรับ Providers

**API:**
```
POST /analytics/profile-view             → Track view (with user_id or null for anonymous)
```

### 3. Reports

Users สามารถรายงานผู้ใช้อื่นที่ผิดกฎ

**API:**
```
POST /reports                            → Create report
GET /reports/my                          → My reports

GET /admin/reports                       → All reports (Admin)
PATCH /admin/reports/:id                 → Update report status (Admin)
DELETE /admin/reports/:id                → Delete report (Admin)
```

---

## 🚫 Block User System

Users สามารถ Block ผู้ใช้อื่น

**API:**
```
POST /blocks                             → Block user
DELETE /blocks/:userId                   → Unblock user
GET /blocks                              → Get blocked users
GET /blocks/check/:userId                → Check if user is blocked
```

---

## 🔍 Browse & Search

### 1. Browse Providers (ต้อง Login)

**API:**
```
GET /browse/v2                           → Browse providers with filters
  Query Parameters:
    - category: string
    - province: string
    - district: string
    - min_rating: number
    - min_price: number
    - max_price: number
    - service_type: "Incall" | "Outcall" | "Both"
    - page: number
    - limit: number
```

### 2. Public Profile (ไม่ต้อง Login)

**API:**
```
GET /provider/:userId/public             → Limited profile data (no age/height/service_type)
GET /provider/:userId/photos             → Provider photos (Public)
GET /packages/:providerId                → Provider packages (Public)
GET /reviews/:providerId                 → Provider reviews (Public)
```

### 3. Full Profile (ต้อง Login)

**API:**
```
GET /provider/:userId                    → Full profile data (including age/height/service_type)
```

**Comparison:**

| Field | Public (`/provider/:userId/public`) | Authenticated (`/provider/:userId`) |
|-------|-------------------------------------|-------------------------------------|
| username, bio, skills | ✅ | ✅ |
| profile_image, rating, reviews | ✅ | ✅ |
| province, district, sub_district | ✅ | ✅ |
| **age, height, weight** | ❌ | ✅ |
| **service_type, working_hours** | ❌ | ✅ |
| **address_line1, lat/lng** | ❌ | ✅ |

---

## 🎖️ Tier System Summary

### User Subscription Tiers (ซื้อได้)

| Tier | Price/Month | Features |
|------|-------------|----------|
| **General** | Free | Basic access |
| **Silver** | 9.99 THB | Premium features |
| **Diamond** | 29.99 THB | Advanced features |
| **Premium** | 99.99 THB | Full access |
| **GOD** | 9999.99 THB | Admin + Full control |

**API:**
```
POST /subscription/create-checkout       → Create Stripe checkout session
POST /payment/webhook                    → Stripe webhook (auto-upgrade tier)
```

### Provider Tiers (คำนวณอัตโนมัติ)

| Tier | Points | How to Get |
|------|--------|-----------|
| **General** | 0-99 | New providers |
| **Silver** | 100-249 | Good ratings + reviews |
| **Diamond** | 250-399 | Excellent performance |
| **Premium** | 400+ | Top performers |

**คำนวณจาก:** Rating, Completed Bookings, Reviews, Response Rate, Acceptance Rate

---

## 📝 Complete API Endpoints Summary

### Public Endpoints (ไม่ต้อง Login)
```
✅ POST   /auth/send-verification
✅ POST   /auth/verify-email
✅ POST   /register
✅ POST   /register/provider
✅ POST   /login
✅ POST   /auth/google
✅ GET    /service-categories
✅ GET    /categories/:id/providers
✅ GET    /provider/:userId/public
✅ GET    /provider/:userId/photos
✅ GET    /packages/:providerId
✅ GET    /reviews/:providerId
✅ GET    /reviews/stats/:providerId
```

### Protected Endpoints (ต้อง Login)
```
🔒 GET    /users/me
🔒 GET    /provider/:userId
🔒 GET    /browse/v2
🔒 POST   /provider/documents
🔒 GET    /provider/documents
🔒 GET    /provider/my-tier
🔒 GET    /provider/tier-history
🔒 POST   /packages
🔒 POST   /bookings
🔒 GET    /bookings/my
🔒 GET    /bookings/provider
🔒 POST   /reviews
🔒 POST   /bank-accounts
🔒 GET    /wallet
🔒 POST   /withdrawals
🔒 GET    /transactions
🔒 GET    /conversations
🔒 POST   /messages
🔒 GET    /notifications
🔒 POST   /blocks
```

### Admin Endpoints (ต้อง Login + Admin)
```
👮 GET    /admin/providers/pending
👮 PATCH  /admin/verify-document/:id
👮 PATCH  /admin/approve-provider/:userId
👮 GET    /admin/provider-stats
👮 POST   /admin/recalculate-provider-tiers
👮 PATCH  /admin/set-provider-tier/:userId
👮 GET    /admin/withdrawals
👮 POST   /admin/withdrawals/:id/process
👮 GET    /admin/financial/summary
👮 GET    /admin/users
👮 GET    /admin/reports
```

---

## 🚀 Quick Start for Frontend

### 1. User Registration (ง่าย - ไม่ต้องส่งเอกสาร)
```typescript
// ส่ง OTP
await sendOTP('user@example.com');

// Register
await registerUser({
  username: 'john_doe',
  email: 'user@example.com',
  password: 'password123',
  gender_id: 1,
  first_name: 'John',
  last_name: 'Doe',
  phone: '0812345678',
  otp: '123456'
});

// เสร็จแล้ว! User สามารถใช้งานได้ทันที
```

### 2. Provider Registration (ซับซ้อนกว่า - ต้องส่งเอกสาร)
```typescript
// ส่ง OTP
await sendOTP('provider@example.com');

// Register as Provider
await registerProvider({
  username: 'massage_pro',
  email: 'provider@example.com',
  password: 'password123',
  gender_id: 2,
  first_name: 'Provider',
  last_name: 'Name',
  phone: '0812345678',
  otp: '123456',
  category_ids: [1, 2], // Massage, Spa
  service_type: 'Both',
  bio: 'Professional massage therapist',
  province: 'Bangkok',
  district: 'Sukhumvit'
});

// Upload documents
await uploadDocument({
  document_type: 'national_id',
  file_url: 'https://storage/.../id_card.jpg',
  file_name: 'id_card.jpg'
});

await uploadDocument({
  document_type: 'health_certificate',
  file_url: 'https://storage/.../health_cert.pdf',
  file_name: 'health_cert.pdf'
});

// รอ Admin อนุมัติเอกสาร
// หลังจากอนุมัติแล้ว Provider สามารถสร้างแพ็คเกจและรับการจองได้
```

---

**Server:** http://localhost:8080  
**Documentation:** `/PROVIDER_SYSTEM_GUIDE.md`, `/FRONTEND_PROVIDER_ROUTES.md`  
**Status:** ✅ All systems operational  
**Last Updated:** November 14, 2025
