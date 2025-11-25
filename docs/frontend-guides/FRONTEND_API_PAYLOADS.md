# 📡 Frontend API Payloads - Complete Reference

เอกสารรวมข้อมูลทั้งหมดที่ Frontend ส่งไปยัง Backend

**Date:** November 14, 2025  
**Version:** 1.1.0  
**API Base URL:** `http://localhost:8080`

---

## 🔐 Authentication APIs

### 1. User Registration (Standard)

**Endpoint:** `POST /register`

**Request Body:**
```typescript
{
  username: string;        // Required, unique
  email: string;          // Required, unique, valid email
  password: string;       // Required, min 8 characters
  gender_id: number;      // Required, 1=Male, 2=Female, 3=Other, 4=Prefer not to say
  is_provider?: boolean;  // Optional, default: false
}
```

**Example:**
```json
{
  "username": "john_doe",
  "email": "john@example.com",
  "password": "SecurePass123",
  "gender_id": 1,
  "is_provider": false
}
```

**Expected Response:**
```json
{
  "message": "Registration successful",
  "user_id": 123
}
```

---

### 2. Provider Registration (Enhanced)

**Endpoint:** `POST /register/provider`

**Request Body:**
```typescript
{
  // Basic Info (Same as User)
  username: string;              // Required, unique
  email: string;                // Required, unique
  password: string;             // Required, min 8 characters
  gender_id: number;            // Required, 1-4
  
  // Personal Info
  first_name?: string;          // Optional
  last_name?: string;           // Optional
  phone: string;                // Required for provider
  otp: string;                  // Required, 6-digit verification code
  
  // Provider-Specific
  category_ids: number[];       // Required, 1-5 categories, array of category IDs
  service_type: "Incall" | "Outcall" | "Both";  // Required
  bio?: string;                 // Optional, max 1000 chars
  
  // Location
  province?: string;            // Optional
  district?: string;            // Optional
  sub_district?: string;        // Optional
  postal_code?: string;         // Optional
  address_line1?: string;       // Optional
  latitude?: number;            // Optional
  longitude?: number;           // Optional
}
```

**Example:**
```json
{
  "username": "massage_pro",
  "email": "provider@example.com",
  "password": "SecurePass123",
  "gender_id": 2,
  "first_name": "Sarah",
  "last_name": "Johnson",
  "phone": "0812345678",
  "otp": "123456",
  "category_ids": [1, 2, 5],
  "service_type": "Both",
  "bio": "Professional massage therapist with 10 years experience. Specialized in Thai massage and aromatherapy.",
  "province": "Bangkok",
  "district": "Sukhumvit"
}
```

**Expected Response:**
```json
{
  "message": "Provider registration successful. Please upload required documents (National ID & Health Certificate).",
  "user_id": 456,
  "token": "eyJhbGci...",
  "next_step": "Upload documents: National ID, Health Certificate",
  "fee_structure": {
    "total_fee_percentage": 12.75,
    "stripe_fee_percentage": 2.75,
    "platform_commission_percentage": 10.0,
    "provider_earnings_percentage": 87.25
  }
}
```

**Frontend Action After Registration:**
Display one-time modal explaining fee structure:
```
⚠️ Platform Fee Notice

As a service provider, you'll receive 87.25% of each booking.

Fee Breakdown:
• Payment Processing (Stripe): 2.75%
• Platform Commission: 10%
• Total Fees: 12.75%

Example Calculation:
Booking Price: ฿1,000
Fees Deducted: ฿127.50
You Receive: ฿872.50

[I Understand] [View Full Terms]
```

---

### 3. Login

**Endpoint:** `POST /auth/login` or `POST /login`

**Request Body:**
```typescript
{
  email: string;     // Required
  password: string;  // Required
}
```

**Example:**
```json
{
  "email": "user@example.com",
  "password": "SecurePass123"
}
```

**Expected Response:**
```json
{
  "message": "Login successful",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "user_id": 123,
    "username": "john_doe",
    "email": "user@example.com",
    "subscription_tier_id": 1,
    "is_admin": false,
    "verification_status": "verified"
  }
}
```

---

### 4. Google OAuth Login

**Endpoint:** `POST /auth/google`

**Request Body:**
```typescript
{
  code: string;  // Google authorization code from OAuth flow
}
```

**Example:**
```json
{
  "code": "4/0AY0e-g7X..."
}
```

**Expected Response:**
```json
{
  "message": "Login successful",
  "token": "eyJhbGci...",
  "user": {
    "user_id": 789,
    "username": "google_user",
    "email": "user@gmail.com",
    "google_profile_picture": "https://lh3.googleusercontent.com/...",
    "subscription_tier_id": 1,
    "is_admin": false
  }
}
```

---

### 5. Send OTP

**Endpoint:** `POST /auth/send-verification`

**Request Body:**
```typescript
{
  email: string;  // Required, email to receive OTP
}
```

**Example:**
```json
{
  "email": "provider@example.com"
}
```

**Expected Response:**
```json
{
  "message": "Verification code sent to email"
}
```

---

## 📄 Provider Document Management

### 6. Upload Document

**Endpoint:** `POST /provider/documents`

**Headers:**
```
Authorization: Bearer {token}
```

**Request Body:**
```typescript
{
  document_type: "national_id" | "health_certificate" | "business_license" | "portfolio" | "certification";
  file_url: string;   // URL of uploaded file (from cloud storage)
  file_name: string;  // Original filename
}
```

**Example:**
```json
{
  "document_type": "national_id",
  "file_url": "https://storage.googleapis.com/skillmatch/documents/user123_national_id.jpg",
  "file_name": "national_id.jpg"
}
```

**Expected Response:**
```json
{
  "message": "Document uploaded successfully. Pending admin verification.",
  "document_id": 456
}
```

---

### 7. Get My Documents

**Endpoint:** `GET /provider/documents`

**Headers:**
```
Authorization: Bearer {token}
```

**Query Parameters:** None

**Expected Response:**
```json
{
  "documents": [
    {
      "document_id": 456,
      "user_id": 123,
      "document_type": "national_id",
      "file_url": "https://storage.googleapis.com/.../national_id.jpg",
      "file_name": "national_id.jpg",
      "verification_status": "pending",
      "uploaded_at": "2025-11-14T10:30:00Z"
    },
    {
      "document_id": 457,
      "user_id": 123,
      "document_type": "health_certificate",
      "file_url": "https://storage.googleapis.com/.../health_cert.pdf",
      "file_name": "health_certificate.pdf",
      "verification_status": "approved",
      "verified_by": 1,
      "verified_at": "2025-11-14T15:00:00Z",
      "uploaded_at": "2025-11-14T10:35:00Z"
    }
  ]
}
```

---

### 8. Get My Provider Categories

**Endpoint:** `GET /provider/categories/me`

**Headers:**
```
Authorization: Bearer {token}
```

**Expected Response:**
```json
{
  "provider_id": 123,
  "category_ids": [1, 2, 5],
  "categories": [
    {
      "category_id": 1,
      "name": "Massage",
      "name_thai": "นวดแผนไทย",
      "is_primary": true
    },
    {
      "category_id": 2,
      "name": "Spa",
      "name_thai": "สปา",
      "is_primary": false
    }
  ]
}
```

---

## 🏆 Provider Tier System

### 9. Get My Provider Tier

**Endpoint:** `GET /provider/my-tier`

**Headers:**
```
Authorization: Bearer {token}
```

**Expected Response:**
```json
{
  "user_id": 123,
  "username": "massage_pro",
  "current_tier_id": 2,
  "current_tier_name": "Silver",
  "tier_points": 150,
  "next_tier_id": 3,
  "next_tier_name": "Diamond",
  "points_to_next_tier": 100,
  "points_breakdown": {
    "rating_points": 85,
    "booking_points": 40,
    "review_points": 15,
    "response_rate_points": 8,
    "acceptance_rate_points": 2
  },
  "stats": {
    "average_rating": 4.25,
    "completed_bookings": 8,
    "total_reviews": 5,
    "response_rate": 80.0,
    "acceptance_rate": 40.0
  }
}
```

---

### 10. Get Tier History

**Endpoint:** `GET /provider/tier-history`

**Headers:**
```
Authorization: Bearer {token}
```

**Expected Response:**
```json
{
  "history": [
    {
      "history_id": 1,
      "user_id": 123,
      "old_tier_id": 1,
      "old_tier_name": "General",
      "new_tier_id": 2,
      "new_tier_name": "Silver",
      "change_type": "auto",
      "reason": "Reached 100 tier points",
      "changed_at": "2025-11-10T14:30:00Z"
    }
  ]
}
```

---

## 👮 Admin APIs

### 11. Get Pending Providers

**Endpoint:** `GET /admin/providers/pending`

**Headers:**
```
Authorization: Bearer {token}
```

**Required:** User must be Admin (`is_admin: true`)

**Expected Response:**
```json
{
  "providers": [
    {
      "user_id": 456,
      "username": "new_provider",
      "email": "provider@example.com",
      "provider_verification_status": "documents_submitted",
      "registration_date": "2025-11-14T08:00:00Z",
      "pending_documents": [
        {
          "document_id": 789,
          "document_type": "national_id",
          "file_url": "https://storage.googleapis.com/.../id.jpg",
          "file_name": "national_id.jpg",
          "verification_status": "pending",
          "uploaded_at": "2025-11-14T08:30:00Z"
        }
      ]
    }
  ]
}
```

---

### 12. Verify Document (Admin)

**Endpoint:** `PATCH /admin/verify-document/:documentId`

**Headers:**
```
Authorization: Bearer {token}
```

**Required:** Admin only

**Request Body:**
```typescript
{
  verification_status: "approved" | "rejected";
  rejection_reason?: string;  // Required if rejected
}
```

**Example (Approve):**
```json
{
  "verification_status": "approved"
}
```

**Example (Reject):**
```json
{
  "verification_status": "rejected",
  "rejection_reason": "ID photo is blurry. Please upload a clearer image."
}
```

**Expected Response:**
```json
{
  "message": "Document verification updated successfully"
}
```

---

### 13. Approve/Reject Provider (Admin)

**Endpoint:** `PATCH /admin/approve-provider/:userId`

**Headers:**
```
Authorization: Bearer {token}
```

**Required:** Admin only

**Request Body:**
```typescript
{
  verification_status: "approved" | "rejected";
  rejection_reason?: string;  // Required if rejected
}
```

**Example (Approve):**
```json
{
  "verification_status": "approved"
}
```

**Example (Reject):**
```json
{
  "verification_status": "rejected",
  "rejection_reason": "Missing required health certificate"
}
```

**Expected Response:**
```json
{
  "message": "Provider verification status updated successfully"
}
```

---

### 14. Get Provider Statistics (Admin)

**Endpoint:** `GET /admin/provider-stats`

**Headers:**
```
Authorization: Bearer {token}
```

**Required:** Admin only

**Expected Response:**
```json
{
  "total_providers": 150,
  "pending_approval": 12,
  "approved_providers": 120,
  "rejected_providers": 18,
  "pending_documents": 25,
  "tier_distribution": {
    "general": 80,
    "silver": 45,
    "diamond": 20,
    "premium": 5
  }
}
```

---

### 15. Recalculate All Provider Tiers (Admin)

**Endpoint:** `POST /admin/recalculate-provider-tiers`

**Headers:**
```
Authorization: Bearer {token}
```

**Required:** Admin only

**Request Body:** None

**Expected Response:**
```json
{
  "message": "Provider tiers recalculated successfully",
  "updated_count": 85
}
```

---

### 16. Manually Set Provider Tier (Admin)

**Endpoint:** `PATCH /admin/set-provider-tier/:userId`

**Headers:**
```
Authorization: Bearer {token}
```

**Required:** Admin only

**Request Body:**
```typescript
{
  provider_level_id: number;  // 1=General, 2=Silver, 3=Diamond, 4=Premium
  reason: string;             // Required, reason for manual change
}
```

**Example:**
```json
{
  "provider_level_id": 4,
  "reason": "Top performer promotion for exceptional customer service"
}
```

**Expected Response:**
```json
{
  "message": "Provider tier updated successfully"
}
```

---

### 17. Get Provider Tier Details (Admin)

**Endpoint:** `GET /admin/provider/:userId/tier-details`

**Headers:**
```
Authorization: Bearer {token}
```

**Required:** Admin only

**Expected Response:**
```json
{
  "user_id": 123,
  "username": "provider_name",
  "current_tier_id": 3,
  "current_tier_name": "Diamond",
  "tier_points": 350,
  "points_breakdown": {
    "rating_points": 95,
    "booking_points": 150,
    "review_points": 75,
    "response_rate_points": 20,
    "acceptance_rate_points": 10
  },
  "stats": {
    "average_rating": 4.75,
    "completed_bookings": 30,
    "total_reviews": 25,
    "response_rate": 95.0,
    "acceptance_rate": 85.0
  }
}
```

---

## 📚 Category Management

### 18. Get All Categories

**Endpoint:** `GET /service-categories`

**Query Parameters:**
```typescript
{
  include_adult?: boolean;  // Optional, default: true
}
```

**Example:** `GET /service-categories?include_adult=true`

**Expected Response:**
```json
{
  "categories": [
    {
      "category_id": 1,
      "name": "Massage",
      "name_thai": "นวดแผนไทย",
      "description": "Traditional and modern massage services",
      "icon": "💆",
      "is_adult": false,
      "display_order": 1,
      "is_active": true
    },
    {
      "category_id": 2,
      "name": "Spa",
      "name_thai": "สปา",
      "description": "Spa and wellness services",
      "icon": "🧖",
      "is_adult": false,
      "display_order": 2,
      "is_active": true
    }
  ],
  "total": 15
}
```

---

### 19. Update My Categories (Provider)

**Endpoint:** `PUT /provider/me/categories`

**Headers:**
```
Authorization: Bearer {token}
```

**Request Body:**
```typescript
{
  category_ids: number[];  // Array of category IDs, max 5
}
```

**Example:**
```json
{
  "category_ids": [1, 2, 3]
}
```

**Expected Response:**
```json
{
  "message": "Categories updated successfully",
  "category_ids": [1, 2, 3],
  "total": 3
}
```

---

## 🔍 Browse & Search

### 20. Browse Providers

**Endpoint:** `GET /browse/v2`

**Headers:**
```
Authorization: Bearer {token}
```

**Query Parameters:**
```typescript
{
  category?: string;           // Optional, category name
  province?: string;           // Optional
  district?: string;           // Optional
  min_rating?: number;         // Optional, 0-5
  min_price?: number;          // Optional
  max_price?: number;          // Optional
  service_type?: "Incall" | "Outcall" | "Both";  // Optional
  page?: number;               // Optional, default: 1
  limit?: number;              // Optional, default: 20
}
```

**Example:** `GET /browse/v2?category=Massage&province=Bangkok&min_rating=4&page=1&limit=20`

**Expected Response:**
```json
{
  "users": [
    {
      "user_id": 123,
      "username": "massage_pro",
      "age": 28,
      "province": "Bangkok",
      "district": "Sukhumvit",
      "is_available": true,
      "average_rating": 4.5,
      "review_count": 25,
      "profile_image_url": "https://...",
      "distance_km": 2.5
    }
  ],
  "total": 45
}
```

---

### 21. Get Provider Public Profile (Logged In)

**Endpoint:** `GET /provider/:userId`

**Headers:**
```
Authorization: Bearer {token}
```

**Required:** Token required for full profile access

**Expected Response:**
```json
{
  "user_id": 123,
  "username": "massage_pro",
  "bio": "Professional massage therapist with 10 years experience",
  "location": "Bangkok, Sukhumvit",
  "province": "Bangkok",
  "district": "Sukhumvit",
  "gender_id": 2,
  "tier_name": "Diamond",
  "average_rating": 4.75,
  "review_count": 45,
  "profile_image_url": "profiles/user123.jpg",
  "google_profile_picture": null,
  "skills": ["Thai Massage", "Aromatherapy", "Spa"],
  "age": 28,
  "height": 165,
  "service_type": "Both",
  "is_available": true
}
```

---

### 22. Get Provider Public Profile (Guest)

**Endpoint:** `GET /provider/:userId/public`

**Headers:** None (Public endpoint)

**Expected Response:**
```json
{
  "user_id": 123,
  "username": "massage_pro",
  "bio": "Professional massage therapist with 10 years experience",
  "location": "Bangkok, Sukhumvit",
  "province": "Bangkok",
  "district": "Sukhumvit",
  "gender_id": 2,
  "tier_name": "Diamond",
  "average_rating": 4.75,
  "review_count": 45,
  "profile_image_url": "profiles/user123.jpg",
  "google_profile_picture": null,
  "skills": ["Thai Massage", "Aromatherapy", "Spa"],
  "is_available": true
}
```

**Note:** Guest view does NOT include: `age`, `height`, `service_type`

---

### 23. Get Provider Photos

**Endpoint:** `GET /provider/:userId/photos`

**Headers:** None (Public endpoint)

**TypeScript Interface:**
```typescript
interface UserPhoto {
  photo_id: number;
  user_id: number;
  photo_url: string;
  caption?: string;      // Optional photo caption/description
  sort_order?: number;   // Optional display order
  uploaded_at: string;   // ISO 8601 timestamp
}
```

**Expected Response:**
```json
[
  {
    "photo_id": 1,
    "user_id": 123,
    "photo_url": "photos/user123_photo1.jpg",
    "caption": "Professional workspace",
    "sort_order": 1,
    "uploaded_at": "2025-11-10T10:00:00Z"
  },
  {
    "photo_id": 2,
    "user_id": 123,
    "photo_url": "photos/user123_photo2.jpg",
    "caption": null,
    "sort_order": 2,
    "uploaded_at": "2025-11-10T10:05:00Z"
  }
]
```

**Frontend Usage:**
- First photo is used as cover photo on provider profile
- Photo gallery displays all photos in grid layout
- Click photo to open full-screen modal with caption (if available)
- Navigate with arrow keys (← →) or click buttons (‹ ›)
- Press ESC to close modal

---

### 24. Check if Provider is Favorite

**Endpoint:** `GET /favorites/check/:providerId`

**Headers:**
```
Authorization: Bearer {token}
```

**Required:** Token required (or return `false` if no token)

**Expected Response:**
```json
{
  "is_favorite": true
}
```

**Note:** If no token provided, should return `{"is_favorite": false}` instead of 401

---

### 25. Get Provider Packages

**Endpoint:** `GET /packages/:userId`

**Headers:** None (Public endpoint)

**Expected Response:**
```json
[
  {
    "package_id": 1,
    "provider_id": 123,
    "name": "1 Hour Thai Massage",
    "description": "Traditional Thai massage with aromatherapy oil",
    "price": 500.00,
    "duration_minutes": 60,
    "is_active": true,
    "created_at": "2025-11-01T10:00:00Z"
  },
  {
    "package_id": 2,
    "provider_id": 123,
    "name": "2 Hour Full Body Massage",
    "description": "Deep tissue massage with hot stone therapy",
    "price": 900.00,
    "duration_minutes": 120,
    "is_active": true,
    "created_at": "2025-11-01T10:15:00Z"
  }
]
```

---

### 26. Get Provider Reviews

**Endpoint:** `GET /reviews/:providerId`

**Headers:** None (Public endpoint)

**Query Parameters:**
```typescript
{
  limit?: number;   // Optional, default: 20
  offset?: number;  // Optional, default: 0
}
```

**Example:** `GET /reviews/123?limit=20&offset=0`

**Expected Response:**
```json
[
  {
    "review_id": 1,
    "booking_id": 456,
    "client_id": 789,
    "client_username": "john_doe",
    "provider_id": 123,
    "rating": 5,
    "comment": "Excellent service! Very professional.",
    "created_at": "2025-11-10T15:30:00Z"
  },
  {
    "review_id": 2,
    "booking_id": 457,
    "client_id": 790,
    "client_username": "jane_smith",
    "provider_id": 123,
    "rating": 4,
    "comment": "Good massage, would recommend.",
    "created_at": "2025-11-09T14:20:00Z"
  }
]
```

---

### 27. Get Provider Review Statistics

**Endpoint:** `GET /reviews/stats/:providerId`

**Headers:** None (Public endpoint)

**Expected Response:**
```json
{
  "provider_id": 123,
  "total_reviews": 45,
  "average_rating": 4.75,
  "rating_distribution": {
    "5": 35,
    "4": 8,
    "3": 2,
    "2": 0,
    "1": 0
  }
}
```

---

## 💬 Messaging System

⚠️ **Important Policy**: Users can only send **automated/templated messages** for booking-related communication. Direct personal contact exchange (phone numbers, Line ID, email, social media, etc.) is strictly prohibited. System actively monitors and blocks attempts to share contact details.

### 28. Send Message

**Endpoint:** `POST /messages`

**Headers:**
```
Authorization: Bearer {token}
```

**Request Body:**
```typescript
{
  receiver_id: number;  // Required, ID of message recipient
  content: string;      // Required, message text, max 2000 chars
                       // ⚠️ Content is monitored - no contact info allowed
}
```

**Content Restrictions:**
- ❌ Cannot include: phone numbers, email addresses, Line ID, Facebook, Instagram, WhatsApp, etc.
- ✅ Can include: Booking inquiries, service questions, general communication
- Backend may filter or block messages containing contact information

**Example:**
```json
{
  "receiver_id": 123,
  "content": "Hello! I'm interested in your Thai massage service."
}
```

**Expected Response:**
```json
{
  "message": "Message sent successfully",
  "message_id": 789,
  "conversation_id": 456
}
```

**Possible Error Responses:**
```json
{
  "error": "Message contains prohibited contact information"
}
```

---

### 29. Get My Conversations

**Endpoint:** `GET /conversations`

**Headers:**
```
Authorization: Bearer {token}
```

**Expected Response:**
```json
{
  "conversations": [
    {
      "conversation_id": 456,
      "other_user_id": 123,
      "other_username": "massage_pro",
      "last_message": "Hello! I'm interested...",
      "last_message_at": "2025-11-14T10:30:00Z",
      "unread_count": 2
    }
  ]
}
```

---

### 30. Get Conversation Messages

**Endpoint:** `GET /conversations/:conversationId/messages`

**Headers:**
```
Authorization: Bearer {token}
```

**Query Parameters:**
```typescript
{
  limit?: number;   // Optional, default: 50
  offset?: number;  // Optional, default: 0
}
```

**Example:** `GET /conversations/456/messages?limit=50&offset=0`

**Expected Response:**
```json
{
  "messages": [
    {
      "message_id": 789,
      "sender_id": 27,
      "receiver_id": 123,
      "content": "Hello! I'm interested in your services.",
      "is_read": false,
      "sent_at": "2025-11-14T10:30:00Z"
    }
  ]
}
```

---

## 💳 Payment APIs

### 31. Create Subscription Checkout (Client Tier Upgrade)

**Endpoint:** `POST /payment/create-checkout-session`

**Headers:**
```
Authorization: Bearer {token}
```

**Request Body:**
```typescript
{
  tier_id: number;     // Required, 2=Silver, 3=Gold, 4=Platinum
  success_url: string; // Optional, redirect after payment success
  cancel_url: string;  // Optional, redirect on payment cancel
}
```

**Example:**
```json
{
  "tier_id": 2,
  "success_url": "https://skillmatch.com/payment/success",
  "cancel_url": "https://skillmatch.com/payment/cancel"
}
```

**Expected Response:**
```json
{
  "checkout_url": "https://checkout.stripe.com/pay/cs_test_...",
  "session_id": "cs_test_a1B2c3D4..."
}
```

**Frontend Action:**
- Redirect user to `checkout_url`
- User completes payment on Stripe Checkout page
- Stripe redirects back to `success_url` or `cancel_url`
- Backend webhook automatically updates user's `tier_id` after payment

---

### 32. Create Booking Payment Checkout

**Endpoint:** `POST /bookings/create-with-payment`

**Headers:**
```
Authorization: Bearer {token}
```

**Request Body:**
```typescript
{
  provider_id: number;      // Required
  package_id: number;       // Required
  booking_date: string;     // Required, ISO 8601 format
  notes?: string;           // Optional
  success_url?: string;     // Optional, default: "http://localhost:5174/booking/success?session_id={CHECKOUT_SESSION_ID}"
  cancel_url?: string;      // Optional, default: "http://localhost:5174/booking/cancel"
}
```

**Example:**
```json
{
  "provider_id": 123,
  "package_id": 5,
  "booking_date": "2025-11-20T14:00:00Z",
  "notes": "First time booking, please confirm availability",
  "success_url": "https://skillmatch.com/booking/success?session_id={CHECKOUT_SESSION_ID}",
  "cancel_url": "https://skillmatch.com/booking/cancel"
}
```

**Expected Response:**
```json
{
  "message": "Booking created. Please complete payment.",
  "checkout_url": "https://checkout.stripe.com/pay/cs_test_x1Y2z3A4...",
  "session_id": "cs_test_x1Y2z3A4...",
  "booking_id": 789,
  "total_amount": 1000.00
}
```

**Payment Flow:**
1. **Backend creates pending booking**: Status = `pending`, stored in database
2. **Client pays ฿1,000 via Stripe**: Full package price (no fees shown to client)
3. **Stripe webhook processes payment**:
   - Stripe deducts 2.75% (฿27.50) - payment processing fee
   - Platform retains 10% (฿100) - commission
   - Provider receives 87.25% (฿872.50) in `pending_balance`
4. **Booking status updated**: `pending` → `paid`
5. **7-day hold period**: Funds move from `pending_balance` → `available_balance`
6. **Provider can withdraw**: After booking status = `completed`

**Frontend Implementation:**

```typescript
// Step 1: Call booking payment endpoint
async function createBookingWithPayment(bookingData: {
  provider_id: number;
  package_id: number;
  booking_date: string;
  notes?: string;
}) {
  const response = await fetch('http://localhost:8080/bookings/create-with-payment', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      ...bookingData,
      success_url: `${window.location.origin}/booking/success?session_id={CHECKOUT_SESSION_ID}`,
      cancel_url: `${window.location.origin}/booking/cancel`
    })
  });

  const data = await response.json();
  
  if (response.ok) {
    // Step 2: Redirect to Stripe Checkout
    window.location.href = data.checkout_url;
  } else {
    throw new Error(data.error || 'Failed to create booking');
  }
}

// Step 3: Handle success page
// URL: /booking/success?session_id=cs_test_...
function BookingSuccessPage() {
  const searchParams = new URLSearchParams(window.location.search);
  const sessionId = searchParams.get('session_id');
  
  return (
    <div>
      <h1>Payment Successful! 🎉</h1>
      <p>Your booking has been confirmed.</p>
      <p>Session ID: {sessionId}</p>
      <p>Please wait for provider confirmation.</p>
      <Link to="/bookings/my">View My Bookings</Link>
    </div>
  );
}

// Step 4: Handle cancel page
function BookingCancelPage() {
  return (
    <div>
      <h1>Payment Cancelled</h1>
      <p>Your booking was not completed.</p>
      <Link to="/browse">Browse Providers</Link>
    </div>
  );
}
```

**WebSocket Notification (Provider receives):**
```json
{
  "type": "booking_payment",
  "payload": {
    "booking_id": "789",
    "amount": 1000.00,
    "provider_earnings": 872.50,
    "message": "New booking payment received!"
  }
}
```

**Error Responses:**
```json
// Invalid package
{
  "error": "Package not found or inactive"
}

// Package doesn't belong to provider
{
  "error": "Package does not belong to specified provider"
}

// Stripe error
{
  "error": "Failed to create payment session",
  "details": "Stripe error message"
}
```

**Important Notes:**
- ✅ Booking created with status `pending` before payment
- ✅ If Stripe Checkout fails, booking is automatically deleted
- ✅ Payment amount = package price (client pays full price, no fee markup)
- ✅ Provider sees fee breakdown (12.75%) in transaction history
- ✅ `{CHECKOUT_SESSION_ID}` in success_url is replaced by Stripe automatically
- ⚠️ Do NOT show fee breakdown to clients (only to providers)

---

### 33. Payment Webhook

**Endpoint:** `POST /payment/webhook`

**Headers:**
```
Stripe-Signature: {signature}
```

**Request Body:** Stripe Event Object (automatically sent by Stripe)

**Event Types:**
- `checkout.session.completed` - Payment successful
- `invoice.payment_succeeded` - Subscription payment received
- `customer.subscription.deleted` - Subscription cancelled

**Backend Actions:**
- Verify webhook signature
- Check payment type via metadata
- **Subscription Payment**: Update `users.tier_id`
- **Booking Payment**: 
  - Update booking status to `paid`
  - Calculate fees (12.75% total)
  - Create transaction record
  - Update provider `pending_balance`
  - Send notification to provider

**Expected Response:**
```json
{
  "received": true
}
```

---

## 💰 Financial APIs (Provider Wallet)

### 34. Get My Wallet Balance (Provider)

**Endpoint:** `GET /wallet/balance`

**Headers:**
```
Authorization: Bearer {token}
```

**Expected Response:**
```json
{
  "user_id": 123,
  "pending_balance": 2500.00,
  "available_balance": 8725.00,
  "total_earned": 15437.50,
  "total_withdrawn": 4212.50
}
```

**Balance Explanation:**
- `pending_balance`: Funds from completed bookings (7-day hold)
- `available_balance`: Ready for withdrawal
- `total_earned`: Lifetime earnings (87.25% of all bookings)
- `total_withdrawn`: Total amount withdrawn to bank

---

### 35. Get Transaction History (Provider)

**Endpoint:** `GET /wallet/transactions`

**Headers:**
```
Authorization: Bearer {token}
```

**Query Parameters:**
```typescript
{
  limit?: number;   // Optional, default: 20
  offset?: number;  // Optional, default: 0
  type?: "earning" | "withdrawal" | "refund" | "adjustment";  // Optional
}
```

**Example:** `GET /wallet/transactions?limit=20&offset=0&type=earning`

**Expected Response:**
```json
{
  "transactions": [
    {
      "transaction_id": 456,
      "user_id": 123,
      "transaction_type": "earning",
      "amount": 872.50,
      "booking_id": 789,
      "description": "Payment from booking #789",
      "fee_breakdown": {
        "original_amount": 1000.00,
        "stripe_fee": 27.50,
        "platform_commission": 100.00,
        "total_fee_percentage": 12.75,
        "net_amount": 872.50
      },
      "status": "completed",
      "created_at": "2025-11-14T15:00:00Z"
    }
  ],
  "total": 45
}
```

**Fee Display (Provider Only):**
```
Booking Price: ฿1,000.00
Payment Processing: -฿27.50 (2.75%)
Platform Fee: -฿100.00 (10%)
Total Fees: -฿127.50 (12.75%)
Your Earnings: ฿872.50 (87.25%)
```

---

### 36. Request Withdrawal (Provider)

**Endpoint:** `POST /wallet/withdraw`

**Headers:**
```
Authorization: Bearer {token}
```

**Request Body:**
```typescript
{
  amount: number;          // Required, must be <= available_balance
  bank_name: string;       // Required
  bank_account_number: string;  // Required
  account_holder_name: string;  // Required
}
```

**Example:**
```json
{
  "amount": 5000.00,
  "bank_name": "Kasikorn Bank",
  "bank_account_number": "1234567890",
  "account_holder_name": "Sarah Johnson"
}
```

**Expected Response:**
```json
{
  "message": "Withdrawal request submitted successfully. Admin will review within 24-48 hours.",
  "withdrawal_id": 789,
  "amount": 5000.00,
  "status": "pending"
}
```

**Withdrawal Flow:**
1. Provider requests withdrawal from `available_balance`
2. Admin reviews and approves
3. **GOD transfers funds via platform bank account** (tracked centrally)
4. Transfer slip masked (GOD account details hidden)
5. Provider receives notification via WebSocket + Email with masked slip

---

### 37. Get Withdrawal History (Provider)

**Endpoint:** `GET /wallet/withdrawals`

**Headers:**
```
Authorization: Bearer {token}
```

**Query Parameters:**
```typescript
{
  limit?: number;   // Optional, default: 20
  offset?: number;  // Optional, default: 0
  status?: "pending" | "approved" | "rejected" | "completed";  // Optional
}
```

**Example:** `GET /wallet/withdrawals?limit=10&status=completed`

**Expected Response:**
```json
{
  "withdrawals": [
    {
      "withdrawal_id": 789,
      "user_id": 123,
      "amount": 5000.00,
      "bank_name": "Kasikorn Bank",
      "bank_account_number": "***7890",
      "account_holder_name": "Sarah Johnson",
      "status": "completed",
      "transfer_slip_url": "https://storage.googleapis.com/.../masked_slip_789.jpg",
      "requested_at": "2025-11-10T10:00:00Z",
      "approved_at": "2025-11-11T14:30:00Z",
      "completed_at": "2025-11-11T16:45:00Z"
    }
  ],
  "total": 8
}
```

**Transfer Slip:**
- `transfer_slip_url`: Masked slip (GOD account details hidden)
- Shows only: Transfer amount, recipient account (last 4 digits), date/time
- Provider cannot see GOD's platform bank account number

---

### 38. Approve Withdrawal Request (Admin)

**Endpoint:** `POST /admin/withdrawals/:withdrawalId/approve`

**Headers:**
```
Authorization: Bearer {token}
```

**Required:** Admin only

**Request Body:**
```typescript
{
  notes?: string;  // Optional, admin notes
}
```

**Example:**
```json
{
  "notes": "Verified account details, processing transfer"
}
```

**Expected Response:**
```json
{
  "message": "Withdrawal approved successfully. Proceed to transfer funds.",
  "withdrawal_id": 789
}
```

---

### 39. Complete Withdrawal with Transfer Slip (Admin)

**Endpoint:** `POST /admin/withdrawals/:withdrawalId/complete`

**Headers:**
```
Authorization: Bearer {token}
```

**Required:** Admin only

**Request Body:**
```typescript
{
  original_slip_url: string;   // Full transfer slip from GOD's bank
  transfer_slip_url: string;   // Masked slip for provider (GOD details removed)
  platform_bank_account_id: number;  // Which GOD account was used
}
```

**Example:**
```json
{
  "original_slip_url": "https://storage.googleapis.com/.../original_slip_789.jpg",
  "transfer_slip_url": "https://storage.googleapis.com/.../masked_slip_789.jpg",
  "platform_bank_account_id": 1
}
```

**Expected Response:**
```json
{
  "message": "Withdrawal completed. Provider notified via WebSocket and Email.",
  "withdrawal_id": 789
}
```

**Backend Actions:**
1. Store both slip URLs (original for audit, masked for provider)
2. Update withdrawal status to `completed`
3. Record in `withdrawal_transfer_logs`
4. Send WebSocket notification with masked slip
5. Send email notification with masked slip
6. Update GOD commission balance

---

### 40. Get GOD Financial Dashboard (GOD Only)

**Endpoint:** `GET /admin/god/financial-dashboard`

**Headers:**
```
Authorization: Bearer {token}
```

**Required:** GOD tier only (`tier_id = 5`, `access_level = 999`)

**Expected Response:**
```json
{
  "commission_balance": {
    "total_collected": 125000.00,
    "total_transferred": 87250.00,
    "available_balance": 37750.00
  },
  "platform_bank_accounts": [
    {
      "account_id": 1,
      "bank_name": "SCB",
      "account_number": "***3456",
      "account_holder_name": "SkillMatch Platform",
      "is_default": true,
      "total_transactions": 45,
      "total_transferred": 87250.00
    }
  ],
  "recent_transactions": [
    {
      "transaction_id": 456,
      "booking_id": 789,
      "provider_id": 123,
      "original_amount": 1000.00,
      "commission_amount": 100.00,
      "commission_percentage": 10.0,
      "created_at": "2025-11-14T15:00:00Z"
    }
  ],
  "stats": {
    "total_bookings_processed": 1250,
    "total_revenue_processed": 1250000.00,
    "total_commission_earned": 125000.00,
    "average_commission_per_booking": 100.00
  }
}
```

**Dashboard Displays:**
- Total commission collected (10% of all bookings)
- Available balance for transfers
- Platform bank accounts used for withdrawals
- Recent commission transactions
- Overall platform financial statistics

---

## 📊 Summary Statistics

### Total Endpoints: 40
### Categories:
- 🔐 Authentication: 5 endpoints
- 📄 Provider Documents: 3 endpoints
- 🏆 Provider Tier: 2 endpoints
- 👮 Admin Management: 7 endpoints
- 📚 Categories: 2 endpoints
- 🔍 Browse & Provider Profile: 8 endpoints
- 💬 Messaging: 3 endpoints (automated notifications only, contact info prohibited)
- 💳 Payment: 3 endpoints (subscription + booking payments)
- 💰 Financial/Wallet: 7 endpoints (wallet, transactions, withdrawals, GOD dashboard)

---

## 🔑 Authentication Requirements

### Public Endpoints (No token required):
- `POST /register`
- `POST /register/provider`
- `POST /auth/login`
- `POST /auth/google`
- `POST /auth/send-verification`
- `GET /service-categories`
- `GET /provider/:userId/public` (guest profile view)
- `GET /provider/:userId/photos` (public photo gallery)
- `GET /packages/:userId` (provider packages)
- `GET /reviews/:providerId` (provider reviews)
- `GET /reviews/stats/:providerId` (review statistics)
- `POST /payment/webhook` (Stripe webhooks)

### Protected Endpoints (Token required):
- All Provider endpoints (`/provider/*`)
- All Browse endpoints (`/browse/*`)
- `GET /provider/:userId` (full profile access with age, height, service_type)
- `GET /favorites/check/:providerId` (should gracefully return false if no token)
- All Messaging endpoints (`/messages`, `/conversations`)
- All Payment endpoints (`/payment/create-*`)
- All Wallet/Financial endpoints (`/wallet/*`)

### Admin Only Endpoints (Token + is_admin: true):
- All Admin endpoints (`/admin/*`)
- Withdrawal approval/completion (`/admin/withdrawals/*`)

### GOD Only Endpoints (Token + tier_id: 5 + access_level: 999):
- `GET /admin/god/financial-dashboard`
- All GOD-specific financial tracking

---

## 📝 Field Validations

### Username:
- Min: 3 characters
- Max: 50 characters
- Allowed: alphanumeric, underscore, dash
- Unique required

### Email:
- Valid email format
- Unique required

### Password:
- Min: 8 characters
- Recommended: uppercase, lowercase, number, special char

### Phone:
- Format: 10 digits (Thai format)
- Example: "0812345678"

### OTP:
- Exactly 6 digits
- Valid for 10 minutes
- Example: "123456"

### Category IDs:
- Min: 1 category
- Max: 5 categories
- Must be valid category IDs from database

### Service Type:
- Allowed values: "Incall", "Outcall", "Both"
- Case-sensitive

### Document Types:
- Required: "national_id", "health_certificate"
- Optional: "business_license", "portfolio", "certification"

### Provider Tier IDs:
- 1 = General (0-99 points)
- 2 = Silver (100-249 points)
- 3 = Diamond (250-399 points)
- 4 = Premium (400+ points)

### Withdrawal Amount:
- Must be positive number
- Cannot exceed `available_balance`
- Min amount: ฿100 (configurable)

---

## 🎯 Tier Points Calculation

```
Total Points (max 600) = 
  + (average_rating * 20)              = 0-100 points
  + (completed_bookings * 5)           = 0-250 points (max 50 bookings)
  + (total_reviews * 3)                = 0-150 points (max 50 reviews)
  + (response_rate * 0.5)              = 0-50 points
  + (acceptance_rate * 0.5)            = 0-50 points
```

---

## 💰 Fee Calculation Formula

### Total Platform Fees: 12.75%

```typescript
// Example: ฿1,000 booking
const bookingPrice = 1000.00;

// Stripe payment processing fee
const stripeFeePercentage = 2.75;
const stripeFee = bookingPrice * (stripeFeePercentage / 100);  // ฿27.50

// Platform commission (calculated on original amount)
const platformCommissionPercentage = 10.0;
const platformCommission = bookingPrice * (platformCommissionPercentage / 100);  // ฿100.00

// Total fees deducted
const totalFees = stripeFee + platformCommission;  // ฿127.50
const totalFeePercentage = 12.75;

// Provider net earnings
const providerEarnings = bookingPrice - totalFees;  // ฿872.50
const providerEarningsPercentage = 87.25;

// Commission distribution
// - Stripe keeps: ฿27.50 (deducted automatically)
// - GOD platform retains: ฿100.00 (stored in god_commission_balance)
// - Provider receives: ฿872.50 (credited to wallet after 7-day hold)
```

### Fee Display Guidelines

**Client View (During Booking):**
```
Package: 1 Hour Thai Massage
Price: ฿1,000.00
─────────────────────────
Total: ฿1,000.00

[Proceed to Payment]
```
*No fee breakdown shown to clients*

**Provider View (Transaction History):**
```
Booking #789 - November 14, 2025

Booking Price:        ฿1,000.00
Payment Processing:      -฿27.50  (2.75%)
Platform Commission:    -฿100.00  (10.00%)
─────────────────────────────────────
Total Fees:             -฿127.50  (12.75%)
Your Earnings:           ฿872.50  (87.25%)

Status: Pending (Available in 7 days)
```

**Provider Dashboard Summary:**
```
💰 Wallet Balance

Pending:    ฿2,500.00  (Available in 3-7 days)
Available:  ฿8,725.00  (Ready to withdraw)
─────────────────────────────────────
Total Earned: ฿15,437.50
Withdrawn:     ฿4,212.50

ℹ️ Platform fees (12.75%) are deducted automatically
   You receive 87.25% of each booking
```

---

## 📝 Backend Implementation Notes

### Provider Profile Photo Display
Frontend implementation includes:
1. **Profile Picture** - Circular avatar (32x32) with online status indicator
   - Source priority: `profile_image_url` → `google_profile_picture` → placeholder
   - Displayed below cover photo on provider profile page
   
2. **Cover Photo** - First photo from gallery used as hero background (h-80)
   - Full GCS URL: `https://storage.googleapis.com/sex-worker-bucket/{photo_url}`
   
3. **Photo Gallery** - Grid display (2-3 columns) with click-to-expand modal
   - Shows count: "📸 แกลเลอรี่ (5 รูป)"
   - Clickable thumbnails with hover effect
   
4. **Full-Screen Modal** - Image viewer with navigation
   - Displays `caption` if available (centered below image)
   - Arrow navigation: Previous (‹) / Next (›)
   - Keyboard shortcuts: ← → ESC
   - Click backdrop to close

**Backend should ensure:**
- `GET /provider/:userId/photos` returns array sorted by `sort_order`
- `caption` field is optional but recommended for better UX
- `photo_url` contains relative path (base URL handled by frontend)

---

### Financial System & Fee Structure

**Total Platform Fees: 12.75%**
- Stripe Payment Gateway: 2.75%
- Platform Commission: 10%
- **Provider Receives: 87.25%** of booking price

**Fee Visibility Strategy:**
- **Clients**: Pay full package price (no fee breakdown shown)
- **Providers**: See detailed fee breakdown (12.75% deduction explained)
- **Provider Registration**: Display one-time notification about 12.75% fee structure
- **Provider Dashboard**: Show earnings breakdown with fees deducted

**Payment & Withdrawal Flow:**
1. **Booking Payment**: Client pays full price via Stripe
2. **Fee Calculation**: 
   - Stripe deducts 2.75% (payment processing)
   - Platform retains 10% (commission)
   - Provider receives 87.25% in wallet
3. **Wallet States**:
   - `pending_balance`: 7-day hold period after booking completion
   - `available_balance`: Ready for withdrawal
4. **Withdrawal Process**:
   - Provider requests withdrawal from available balance
   - Admin reviews and approves
   - **GOD transfers 87.25% via platform bank account** (tracked centrally)
   - Transfer slip masked (hide GOD account details) before sending to provider
   - Provider receives masked slip via:
     * WebSocket (real-time chat notification)
     * Email (permanent record)

**Frontend Requirements:**
- **Provider Registration Page**: Display fee notification modal (see endpoint #2)
- **Provider Dashboard**: Earnings card with breakdown (see Fee Display Guidelines)
- **Withdrawal History**: Show transfer slip with masked account details
- **Never display**: GOD account number, full platform bank details

**GOD Account Protection:**
- User ID = 1 is GOD account (tier_id = 5, access_level = 999)
- Frontend should never display GOD user in public listings
- Admin creation only accessible to GOD tier
- All financial reports exclude GOD transactions from provider statistics

---

### Messaging System Policy

Frontend displays prominent warning banners:
- **Messages List Page**: Yellow warning box explaining policy and restrictions
- **Chat Page**: Compact inline warning above message input
- **Translations**: Available in Thai, English, and Chinese
- **User Education**: Clear communication that only templated/automated messages allowed

**Backend Implementation:**
- **Template System**: Pre-approved message templates for common booking scenarios
- **Content Filtering**: Real-time detection and blocking of contact information:
  - Phone numbers (Thai format: 0812345678, International: +66812345678)
  - Email addresses (any@domain.com)
  - Social media handles (Line ID, Facebook, Instagram, WhatsApp, Twitter, TikTok)
  - URLs and links
  - Attempts to spell out numbers ("zero eight one two...")
- **Enforcement**: 
  - Return HTTP 400 with clear error message when prohibited content detected
  - Log violations for moderation review
  - Consider temporary suspension for repeated violations
- **WebSocket**: Real-time message blocking before delivery

**Contact Info Detection Patterns:**
```regex
Phone: /\b0[0-9]{9}\b|\+66\s?[0-9]{9}\b|(?:zero|one|two|three|four|five|six|seven|eight|nine)\s*(?:zero|one|two|three|four|five|six|seven|eight|nine)/i
Email: /[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}/
Line: /line\s*(?:id|@|:|\s)\s*[a-zA-Z0-9]/i
Social: /(?:facebook|fb\.com|instagram|ig\.com|whatsapp|twitter|tiktok|wechat|telegram|snapchat)[\s:/]?[a-zA-Z0-9]/i
URLs: /(?:https?:\/\/|www\.)[^\s]+/i
```

**Recommended Message Templates:**
- "I'm interested in booking [Package Name]. When are you available?"
- "Can you provide more details about [Service]?"
- "I'd like to confirm my booking for [Date/Time]."
- "Thank you for the service!"

---

### Payment System Architecture

**Two Payment Types:**

1. **Subscription Payment** (Client Tier Upgrade)
   - Stripe Mode: `subscription`
   - Updates: `users.tier_id`
   - Recurring: Monthly billing
   - Status: ✅ Fully Implemented

2. **Booking Payment** (Service Booking)
   - Stripe Mode: `payment`
   - Updates: `bookings.status`, `transactions`, `wallets`
   - One-time: Per booking
   - Status: ⚠️ Needs Implementation

**Webhook Handler Logic:**
```typescript
// Pseudo-code for payment webhook
if (event.type === 'checkout.session.completed') {
  const session = event.data.object;
  const paymentType = session.metadata.payment_type;
  
  if (paymentType === 'subscription') {
    // Update user's tier_id
    updateUserTier(session.client_reference_id, session.metadata.tier_id);
  } 
  else if (paymentType === 'booking') {
    // Calculate fees and update booking
    const bookingId = session.metadata.booking_id;
    const amount = session.amount_total / 100;  // Convert from cents
    
    // Calculate fee breakdown
    const stripeFee = amount * 0.0275;           // 2.75%
    const platformCommission = amount * 0.10;    // 10%
    const providerEarnings = amount * 0.8725;    // 87.25%
    
    // Update booking status
    updateBookingStatus(bookingId, 'paid');
    
    // Create transaction record
    createTransaction({
      booking_id: bookingId,
      amount: amount,
      stripe_fee: stripeFee,
      platform_commission: platformCommission,
      total_fee_percentage: 12.75,
      provider_earnings: providerEarnings
    });
    
    // Update provider wallet (pending balance with 7-day hold)
    updateProviderWallet(providerId, providerEarnings, 'pending');
    
    // Send notifications
    sendWebSocketNotification(providerId, 'New booking payment received');
    sendEmailNotification(providerId, 'Payment Confirmation', emailTemplate);
  }
}
```

**Security Considerations:**
- Always verify Stripe webhook signature
- Implement idempotency (prevent duplicate processing)
- Use `payment_intent_id` for tracking
- Log all webhook events for audit trail

---

## 📋 Change Log

### November 14, 2025 - v1.1.0
- ✅ Added 10 new financial endpoints (#31-40)
- ✅ Documented complete payment system (subscription + booking)
- ✅ Added fee structure: 12.75% total (2.75% Stripe + 10% platform)
- ✅ Provider earnings: 87.25% per booking
- ✅ Withdrawal flow via platform bank account with masked slips
- ✅ GOD financial dashboard for commission tracking
- ✅ Updated messaging policy: Automated/templated messages only
- ✅ Enhanced content filtering patterns for messaging
- ✅ Provider registration fee notification requirements
- ✅ Complete fee calculation formulas and display guidelines

### November 14, 2025 - v1.0.0
- ✅ Initial complete API documentation
- ✅ 30 endpoints documented across 7 categories
- ✅ Provider profile photo display implementation
- ✅ Messaging system with policy warnings

---

**Last Updated:** November 14, 2025 (v1.1.0)  
**Status:** ✅ Production Ready (Frontend Integration Guide)  
**Breaking Changes:** 
- Fee structure updated - requires frontend UI changes for provider dashboard
- New payment endpoints - booking payment needs frontend integration
- Withdrawal system - requires admin dashboard for approval workflow

---

## 🎯 Next Steps for Frontend

### High Priority (Blocking Features)
1. **Implement Booking Payment Flow**
   - Add "Book with Payment" button on provider profile
   - Integrate Stripe Checkout redirect
   - Handle success/cancel callback URLs
   - Display payment status on booking history

2. **Provider Dashboard Financial UI**
   - Display wallet balance (pending + available)
   - Show transaction history with fee breakdown
   - Withdrawal request form
   - Transfer slip viewer (masked)

3. **Provider Registration Fee Notification**
   - Modal popup on successful registration
   - Display 12.75% fee structure
   - "I Understand" acknowledgment required
   - Link to full terms and conditions

### Medium Priority (UX Enhancements)
4. **Admin Withdrawal Management**
   - Pending withdrawals queue
   - Approve/Reject actions
   - Upload transfer slip (original + masked)
   - Notification system integration

5. **GOD Financial Dashboard**
   - Commission balance overview
   - Platform bank account management
   - Transaction logs
   - Financial statistics and reports

### Low Priority (Nice-to-Have)
6. **Enhanced Messaging Templates**
   - Pre-defined message templates
   - Quick reply buttons
   - Contact info blocking UI feedback

7. **Payment Analytics**
   - Revenue charts
   - Fee breakdown reports
   - Provider earnings trends

---

**Document Version:** 1.1.0  
**Total Endpoints:** 40  
**Total Pages:** 35+  
**Ready for Frontend Integration:** ✅