# 🚀 SkillMatch API - Complete Frontend Reference

**Base URL:** `http://localhost:8080`  
**Production URL:** `https://api.skillmatch.com` (เมื่อ deploy แล้ว)

**Last Updated:** November 13, 2025  
**API Version:** 1.0.0

---

## 📋 Table of Contents

1. [Authentication](#authentication)
2. [Profile Management](#profile-management)
3. [Photos](#photos)
4. [KYC Verification](#kyc-verification)
5. [Subscription & Payment](#subscription--payment)
6. [Browse & Search](#browse--search)
7. [Service Packages](#service-packages)
8. [Bookings](#bookings)
9. [Reviews & Ratings](#reviews--ratings)
10. [Favorites](#favorites)
11. [Messaging (Real-time)](#messaging)
12. [Notifications](#notifications)
13. [Reports](#reports)
14. [Analytics](#analytics)
15. [Block Users](#block-users)
16. [Admin Endpoints](#admin-endpoints)
17. [TypeScript Types](#typescript-types)
18. [Error Handling](#error-handling)
19. [WebSocket](#websocket)

---

## 🔐 Authentication

### Register
```typescript
POST /auth/register
Content-Type: application/json

{
  "username": "john_doe",
  "email": "john@example.com",
  "password": "SecurePass123!",
  "gender_id": 1,
  "first_name": "John",
  "last_name": "Doe",
  "phone_number": "+66812345678"
}

Response 201:
{
  "message": "User registered successfully",
  "user_id": 42,
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

### Login
```typescript
POST /auth/login
Content-Type: application/json

{
  "email": "john@example.com",
  "password": "SecurePass123!"
}

Response 200:
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "user_id": 42,
    "username": "john_doe",
    "email": "john@example.com",
    "is_admin": false
  }
}
```

### Google OAuth
```typescript
POST /auth/google
Content-Type: application/json

{
  "token": "google_oauth_token_here"
}

Response 200:
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": { ... }
}
```

**Store token in localStorage:**
```typescript
localStorage.setItem('token', response.token);

// Use in subsequent requests:
headers: {
  'Authorization': `Bearer ${localStorage.getItem('token')}`
}
```

---

## 👤 Profile Management

### Get My Profile
```typescript
GET /profile/me
Authorization: Bearer {token}

Response 200:
{
  "user_id": 42,
  "username": "john_doe",
  "email": "john@example.com",
  "bio": "Professional service provider",
  "age": 25,
  "height": 170,
  "weight": 60,
  "ethnicity": "Asian",
  "languages": ["English", "Thai", "Chinese"],
  "working_hours": "10:00-22:00",
  "is_available": true,
  "service_type": "both",
  "province": "กรุงเทพมหานคร",
  "district": "บางรัก",
  "sub_district": "สีลม",
  "latitude": 13.7563,
  "longitude": 100.5018,
  "profile_image_url": "https://storage.googleapis.com/...",
  "google_profile_picture": "https://lh3.googleusercontent.com/..."
}
```

### Update My Profile
```typescript
PUT /profile/me
Authorization: Bearer {token}
Content-Type: application/json

{
  "bio": "Updated bio",
  "age": 26,
  "height": 170,
  "weight": 60,
  "ethnicity": "Asian",
  "languages": ["English", "Thai"],
  "working_hours": "10:00-22:00",
  "is_available": true,
  "service_type": "both",
  "skills": ["Massage", "Yoga"],
  "province": "กรุงเทพมหานคร",
  "district": "บางรัก",
  "sub_district": "สีลม",
  "postal_code": "10500",
  "address_line1": "123 Silom Road",
  "latitude": 13.7563,
  "longitude": 100.5018
}

Response 200:
{
  "message": "Profile updated successfully"
}
```

### Get Public Profile (Provider)
```typescript
GET /provider/:userId
Authorization: Bearer {token}

Response 200:
{
  "user_id": 42,
  "username": "john_doe",
  "bio": "...",
  "age": 25,
  "height": 170,
  "service_type": "both",
  "province": "กรุงเทพมหานคร",
  "district": "บางรัก",
  "is_available": true,
  "average_rating": 4.7,
  "review_count": 68,
  "profile_image_url": "...",
  "photos": [
    {
      "photo_id": 1,
      "photo_url": "...",
      "sort_order": 1
    }
  ]
}
```

---

## 📸 Photos

### Start Photo Upload
```typescript
POST /photos/start
Authorization: Bearer {token}
Content-Type: application/json

{
  "filename": "profile-pic.jpg",
  "content_type": "image/jpeg"
}

Response 200:
{
  "upload_url": "https://storage.googleapis.com/...",
  "photo_url": "https://storage.googleapis.com/...",
  "photo_id": 123
}

// Then upload to GCS:
PUT {upload_url}
Content-Type: image/jpeg
Body: <file binary data>
```

### Submit Photo Upload
```typescript
POST /photos/submit
Authorization: Bearer {token}
Content-Type: application/json

{
  "photo_id": 123
}

Response 200:
{
  "message": "Photo added to profile"
}
```

### Delete Photo
```typescript
DELETE /photos/:photoId
Authorization: Bearer {token}

Response 200:
{
  "message": "Photo deleted successfully"
}
```

### Get Provider Photos
```typescript
GET /provider/:userId/photos
Authorization: Bearer {token}

Response 200:
{
  "photos": [
    {
      "photo_id": 1,
      "photo_url": "https://...",
      "sort_order": 1,
      "uploaded_at": "2025-11-13T10:00:00Z"
    }
  ]
}
```

---

## ✅ KYC Verification

### Submit KYC
```typescript
POST /verification/submit
Authorization: Bearer {token}
Content-Type: application/json

{
  "national_id_url": "https://storage.googleapis.com/...",
  "health_cert_url": "https://storage.googleapis.com/...",
  "face_scan_url": "https://storage.googleapis.com/..."
}

Response 200:
{
  "message": "Verification submitted successfully"
}
```

### Check KYC Status
```typescript
GET /verification/status
Authorization: Bearer {token}

Response 200:
{
  "verification_status": "pending" | "approved" | "rejected" | "not_submitted"
}
```

---

## 💳 Subscription & Payment

### Get Tiers
```typescript
GET /tiers

Response 200:
{
  "tiers": [
    {
      "tier_id": 1,
      "name": "General",
      "access_level": 1,
      "price_monthly": 0
    },
    {
      "tier_id": 2,
      "name": "Silver",
      "access_level": 2,
      "price_monthly": 299
    }
  ]
}
```

### Create Checkout Session
```typescript
POST /subscription/create-checkout
Authorization: Bearer {token}
Content-Type: application/json

{
  "tier_id": 2
}

Response 200:
{
  "checkout_url": "https://checkout.stripe.com/..."
}
```

---

## 🔍 Browse & Search

### Browse Providers (Advanced)
```typescript
GET /browse/v2?gender=1&min_age=20&max_age=30&province=กรุงเทพมหานคร&service_type=incall&available=true&min_rating=4.0&limit=20&offset=0
Authorization: Bearer {token}

Query Parameters:
- gender: 1, 2, 3 (optional)
- min_age, max_age: number (optional)
- province: string (optional)
- district: string (optional)
- sub_district: string (optional)
- max_distance: number in km (requires lat & lng)
- lat, lng: GPS coordinates (optional)
- service_type: "incall" | "outcall" | "both" (optional)
- available: boolean (optional)
- min_price, max_price: number (optional)
- min_rating: number 1-5 (optional)
- ethnicity: string (optional)
- limit: number (default 50)
- offset: number (default 0)

Response 200:
{
  "users": [
    {
      "user_id": 42,
      "username": "john_doe",
      "age": 25,
      "height": 170,
      "ethnicity": "Asian",
      "service_type": "both",
      "province": "กรุงเทพมหานคร",
      "district": "บางรัก",
      "is_available": true,
      "average_rating": 4.7,
      "review_count": 68,
      "profile_image_url": "...",
      "distance_km": 2.5
    }
  ],
  "total": 150
}
```

---

## 📦 Service Packages

### Get Provider Packages
```typescript
GET /packages/:providerId
Authorization: Bearer {token}

Response 200:
{
  "packages": [
    {
      "package_id": 1,
      "provider_id": 42,
      "package_name": "1 Hour Standard",
      "description": "Standard massage service",
      "duration": 60,
      "price": 2000,
      "created_at": "2025-11-13T10:00:00Z"
    }
  ]
}
```

### Create Package (Provider Only)
```typescript
POST /packages
Authorization: Bearer {token}
Content-Type: application/json

{
  "package_name": "2 Hours Premium",
  "description": "Premium service with extras",
  "duration": 120,
  "price": 4000
}

Response 201:
{
  "message": "Package created",
  "package_id": 5
}
```

---

## 📅 Bookings

### Create Booking
```typescript
POST /bookings
Authorization: Bearer {token}
Content-Type: application/json

{
  "package_id": 1,
  "booking_date": "2025-11-20",
  "start_time": "14:00",
  "end_time": "15:00",
  "location": "Client's hotel",
  "special_notes": "Please call 30 min before"
}

Response 201:
{
  "message": "Booking created",
  "booking_id": 123
}
```

### Get My Bookings (Client)
```typescript
GET /bookings/my
Authorization: Bearer {token}

Response 200:
{
  "bookings": [
    {
      "booking_id": 123,
      "provider_id": 42,
      "provider_username": "john_doe",
      "provider_profile_pic": "...",
      "package_name": "1 Hour Standard",
      "duration": 60,
      "booking_date": "2025-11-20",
      "start_time": "14:00",
      "end_time": "15:00",
      "total_price": 2000,
      "status": "pending",
      "location": "...",
      "created_at": "2025-11-13T10:00:00Z"
    }
  ]
}
```

### Get Provider Bookings
```typescript
GET /bookings/provider
Authorization: Bearer {token}

Response 200:
{
  "bookings": [
    {
      "booking_id": 123,
      "client_id": 50,
      "client_username": "client_user",
      "package_name": "1 Hour Standard",
      "status": "pending",
      ...
    }
  ]
}
```

### Update Booking Status
```typescript
PATCH /bookings/:id/status
Authorization: Bearer {token}
Content-Type: application/json

{
  "status": "confirmed" | "completed" | "cancelled",
  "cancellation_reason": "Optional reason if cancelled"
}

Response 200:
{
  "message": "Booking status updated"
}
```

---

## ⭐ Reviews & Ratings

### Create Review
```typescript
POST /reviews
Authorization: Bearer {token}
Content-Type: application/json

{
  "booking_id": 123,
  "rating": 5,
  "comment": "Excellent service! Highly recommended."
}

Response 201:
{
  "message": "Review submitted",
  "review_id": 45
}
```

### Get Provider Reviews
```typescript
GET /reviews/:providerId?limit=20&offset=0
Authorization: Bearer {token}

Response 200:
{
  "reviews": [
    {
      "review_id": 45,
      "client_username": "client_user",
      "rating": 5,
      "comment": "Excellent service!",
      "is_verified": true,
      "created_at": "2025-11-13T10:00:00Z"
    }
  ],
  "total": 68
}
```

### Get Review Stats
```typescript
GET /reviews/:providerId/stats
Authorization: Bearer {token}

Response 200:
{
  "average_rating": 4.7,
  "total_reviews": 68,
  "rating_breakdown": {
    "5": 52,
    "4": 12,
    "3": 3,
    "2": 1,
    "1": 0
  }
}
```

---

## ❤️ Favorites

### Add to Favorites
```typescript
POST /favorites
Authorization: Bearer {token}
Content-Type: application/json

{
  "provider_id": 42
}

Response 201:
{
  "message": "Provider added to favorites",
  "favorite_id": 10
}
```

### Remove from Favorites
```typescript
DELETE /favorites/:providerId
Authorization: Bearer {token}

Response 200:
{
  "message": "Provider removed from favorites"
}
```

### Get My Favorites
```typescript
GET /favorites
Authorization: Bearer {token}

Response 200:
{
  "favorites": [
    {
      "favorite_id": 10,
      "provider_id": 42,
      "provider_username": "john_doe",
      "profile_image_url": "...",
      "average_rating": 4.7,
      "added_at": "2025-11-13T10:00:00Z"
    }
  ]
}
```

### Check if Favorite
```typescript
GET /favorites/check/:providerId
Authorization: Bearer {token}

Response 200:
{
  "is_favorite": true
}
```

---

## 💬 Messaging

### Get Conversations
```typescript
GET /conversations
Authorization: Bearer {token}

Response 200:
{
  "conversations": [
    {
      "id": 1,
      "user1_id": 5,
      "user2_id": 42,
      "other_user_id": 42,
      "other_user_name": "john_doe",
      "other_user_picture": "...",
      "last_message": "See you tomorrow!",
      "last_message_at": "2025-11-13T10:00:00Z",
      "unread_count": 2
    }
  ]
}
```

### Get Conversation Messages
```typescript
GET /conversations/:id/messages?limit=50&offset=0
Authorization: Bearer {token}

Response 200:
{
  "messages": [
    {
      "id": 100,
      "conversation_id": 1,
      "sender_id": 42,
      "content": "Hello!",
      "is_read": true,
      "created_at": "2025-11-13T09:00:00Z"
    }
  ]
}
```

### Send Message
```typescript
POST /messages
Authorization: Bearer {token}
Content-Type: application/json

{
  "participant_id": 42,
  "content": "Hello! Are you available tomorrow?"
}

Response 201:
{
  "message": {
    "id": 101,
    "conversation_id": 1,
    "sender_id": 5,
    "content": "Hello! Are you available tomorrow?",
    "created_at": "2025-11-13T10:00:00Z"
  }
}
```

### Mark Messages as Read
```typescript
PATCH /messages/read
Authorization: Bearer {token}
Content-Type: application/json

{
  "conversation_id": 1
}

Response 200:
{
  "message": "Messages marked as read"
}
```

### Delete Message
```typescript
DELETE /messages/:id
Authorization: Bearer {token}

Response 200:
{
  "message": "Message deleted"
}
```

---

## 🔔 Notifications

### Get Notifications
```typescript
GET /notifications?limit=50&offset=0&type=new_message
Authorization: Bearer {token}

Response 200:
{
  "notifications": [
    {
      "id": 10,
      "type": "booking_request",
      "title": "New Booking Request",
      "message": "You have a new booking request",
      "metadata": {
        "booking_id": 123,
        "client_id": 50
      },
      "is_read": false,
      "created_at": "2025-11-13T10:00:00Z"
    }
  ],
  "total": 25,
  "unread_count": 5
}
```

### Get Unread Count
```typescript
GET /notifications/unread/count
Authorization: Bearer {token}

Response 200:
{
  "unread_count": 5
}
```

### Mark as Read
```typescript
PATCH /notifications/:id/read
Authorization: Bearer {token}

Response 200:
{
  "message": "Notification marked as read"
}
```

### Mark All as Read
```typescript
PATCH /notifications/read-all
Authorization: Bearer {token}

Response 200:
{
  "message": "All notifications marked as read",
  "updated_count": 12
}
```

### Delete Notification
```typescript
DELETE /notifications/:id
Authorization: Bearer {token}

Response 200:
{
  "message": "Notification deleted"
}
```

### Delete All Notifications
```typescript
DELETE /notifications
Authorization: Bearer {token}

Response 200:
{
  "message": "All notifications deleted",
  "deleted_count": 25
}
```

---

## 🚨 Reports

### Create Report
```typescript
POST /reports
Authorization: Bearer {token}
Content-Type: application/json

{
  "reported_user_id": 42,
  "reason": "harassment",
  "description": "This user sent threatening messages"
}

Reasons: "harassment" | "inappropriate_content" | "fake_profile" | "scam" | "violence_threat" | "underage" | "spam" | "other"

Response 201:
{
  "message": "Report submitted successfully",
  "report": {
    "id": 15,
    "status": "pending"
  }
}
```

### Get My Reports
```typescript
GET /reports/my
Authorization: Bearer {token}

Response 200:
{
  "reports": [
    {
      "id": 15,
      "reported_user_id": 42,
      "reported_user_name": "john_doe",
      "reason": "harassment",
      "status": "pending",
      "created_at": "2025-11-13T10:00:00Z"
    }
  ]
}
```

---

## 📊 Analytics

### Get Provider Dashboard
```typescript
GET /analytics/provider/dashboard
Authorization: Bearer {token}

Response 200:
{
  "profile_views": 1250,
  "total_bookings": 87,
  "completed_bookings": 72,
  "cancelled_bookings": 5,
  "pending_bookings": 10,
  "total_revenue": 215000.00,
  "average_rating": 4.7,
  "total_reviews": 68,
  "favorite_count": 142,
  "response_rate": 95.5,
  "average_response_time": 15
}
```

### Get Booking Stats
```typescript
GET /analytics/provider/bookings?period=30
Authorization: Bearer {token}

Response 200:
{
  "stats": [
    {
      "date": "2025-11-13",
      "booking_count": 3,
      "revenue": 9000.00,
      "completed_count": 2,
      "cancelled_count": 0
    }
  ],
  "period": "30"
}
```

### Get Revenue Breakdown
```typescript
GET /analytics/provider/revenue
Authorization: Bearer {token}

Response 200:
{
  "revenue_breakdown": [
    {
      "package_name": "2 Hours Premium",
      "booking_count": 45,
      "total_revenue": 135000.00,
      "avg_price": 3000.00
    }
  ]
}
```

### Get Rating Breakdown
```typescript
GET /analytics/provider/ratings
Authorization: Bearer {token}

Response 200:
{
  "breakdown": {
    "rating_5": 52,
    "rating_4": 12,
    "rating_3": 3,
    "rating_2": 1,
    "rating_1": 0
  },
  "total_reviews": 68
}
```

### Get Monthly Stats
```typescript
GET /analytics/provider/monthly
Authorization: Bearer {token}

Response 200:
{
  "monthly_stats": [
    {
      "month": "2025-11",
      "booking_count": 15,
      "completed_count": 12,
      "revenue": 36000.00,
      "new_reviews": 10,
      "average_rating": 4.8
    }
  ]
}
```

### Track Profile View
```typescript
POST /analytics/profile-view
Authorization: Bearer {token}
Content-Type: application/json

{
  "provider_id": 42
}

Response 200:
{
  "message": "Profile view tracked"
}
```

---

## 🚫 Block Users

### Block User
```typescript
POST /blocks
Authorization: Bearer {token}
Content-Type: application/json

{
  "blocked_user_id": 42,
  "reason": "Spam messages"
}

Response 201:
{
  "message": "User blocked successfully",
  "block_id": 15
}
```

### Unblock User
```typescript
DELETE /blocks/:userId
Authorization: Bearer {token}

Response 200:
{
  "message": "User unblocked successfully"
}
```

### Get Blocked Users
```typescript
GET /blocks
Authorization: Bearer {token}

Response 200:
{
  "blocked_users": [
    {
      "user_id": 42,
      "username": "john_doe",
      "profile_image_url": "...",
      "reason": "Spam messages",
      "blocked_at": "2025-11-13T10:00:00Z"
    }
  ],
  "total": 1
}
```

### Check Block Status
```typescript
GET /blocks/check/:userId
Authorization: Bearer {token}

Response 200:
{
  "is_blocked": true,      // You blocked this user
  "is_blocked_by": false   // This user blocked you
}
```

---

## 👨‍💼 Admin Endpoints

### Get Pending KYC
```typescript
GET /admin/pending-users
Authorization: Bearer {admin_token}

Response 200:
{
  "users": [
    {
      "user_id": 42,
      "username": "john_doe",
      "email": "john@example.com",
      "submitted_at": "2025-11-13T10:00:00Z"
    }
  ]
}
```

### Approve KYC
```typescript
POST /admin/approve/:userId
Authorization: Bearer {admin_token}

Response 200:
{
  "message": "User approved"
}
```

### Reject KYC
```typescript
POST /admin/reject/:userId
Authorization: Bearer {admin_token}

Response 200:
{
  "message": "User rejected"
}
```

### Get All Reports
```typescript
GET /admin/reports?status=pending
Authorization: Bearer {admin_token}

Response 200:
{
  "reports": [
    {
      "id": 15,
      "reporter_id": 5,
      "reporter_name": "alice",
      "reported_user_id": 42,
      "reported_user_name": "john_doe",
      "reason": "harassment",
      "status": "pending",
      "created_at": "2025-11-13T10:00:00Z"
    }
  ]
}
```

### Update Report Status
```typescript
PATCH /admin/reports/:id
Authorization: Bearer {admin_token}
Content-Type: application/json

{
  "status": "resolved",
  "admin_notes": "User warned and content removed"
}

Response 200:
{
  "message": "Report updated successfully"
}
```

---

## 🔌 WebSocket

### Connect to WebSocket

**SECURE METHOD** - Token ไม่ถูกแสดงใน URL:

```typescript
// ✅ CORRECT & SECURE - เชื่อมต่อก่อน แล้วส่ง token ผ่าน message
const token = localStorage.getItem('token');
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onopen = () => {
  console.log('WebSocket connected, authenticating...');
  
  // ส่ง auth message หลังเชื่อมต่อสำเร็จ
  ws.send(JSON.stringify({
    type: 'auth',
    payload: {
      token: token
    }
  }));
};

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  
  switch(data.type) {
    case 'auth_success':
      console.log('✅ Authenticated:', data.payload);
      // เริ่มใช้งาน WebSocket ได้เลย
      break;
      
    case 'auth_error':
      console.error('❌ Authentication failed:', data.payload.error);
      ws.close();
      break;
      
    case 'message':
      handleNewMessage(data.payload);
      break;
      
    case 'notification':
      handleNotification(data.payload);
      break;
      
    case 'typing':
      handleTyping(data.payload);
      break;
  }
};

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};

ws.onclose = () => {
  console.log('WebSocket disconnected');
  // Implement reconnection logic here
};
```

**ทำไมต้องทำแบบนี้?**
- 🔒 **ปลอดภัย**: Token ไม่ปรากฏใน URL, browser history, หรือ server logs
- ✅ **มาตรฐาน**: ใช้ WebSocket message protocol ตามที่ควรจะเป็น
- 🚫 **หลีกเลี่ยงปัญหา**: Token ไม่ leak ผ่าน Referer headers

**❌ วิธีเก่าที่ไม่ปลอดภัย (อย่าใช้):**
```typescript
// ❌ ไม่ปลอดภัย - Token แสดงใน URL
const ws = new WebSocket(`ws://localhost:8080/ws?token=${token}`);

// ❌ ไม่ work - WebSocket ไม่รองรับ custom headers
const ws = new WebSocket('ws://localhost:8080/ws', {
  headers: { 'Authorization': `Bearer ${token}` }
});
```

### WebSocket Message Types

**Incoming Messages:**
```typescript
// New chat message
{
  "type": "message",
  "payload": {
    "id": 101,
    "conversation_id": 1,
    "sender_id": 42,
    "content": "Hello!",
    "created_at": "2025-11-13T10:00:00Z"
  }
}

// New notification
{
  "type": "notification",
  "payload": {
    "id": 15,
    "type": "booking_request",
    "title": "New Booking",
    "message": "You have a new booking request"
  }
}

// Typing indicator
{
  "type": "typing",
  "payload": {
    "conversation_id": 1,
    "user_id": 42,
    "is_typing": true
  }
}
```

**Outgoing Messages:**
```typescript
// Send typing indicator
ws.send(JSON.stringify({
  type: 'typing',
  conversation_id: 1,
  is_typing: true
}));
```

---

## 📝 TypeScript Types

```typescript
// User
interface User {
  user_id: number;
  username: string;
  email: string;
  gender_id: number;
  subscription_tier_id: number;
  verification_status: 'not_submitted' | 'pending' | 'approved' | 'rejected';
  is_admin: boolean;
}

// Profile
interface UserProfile {
  user_id: number;
  bio?: string;
  age?: number;
  height?: number;
  weight?: number;
  ethnicity?: string;
  languages?: string[];
  working_hours?: string;
  is_available: boolean;
  service_type?: 'incall' | 'outcall' | 'both';
  province?: string;
  district?: string;
  sub_district?: string;
  latitude?: number;
  longitude?: number;
  profile_image_url?: string;
  average_rating?: number;
  review_count?: number;
}

// Booking
interface Booking {
  booking_id: number;
  client_id: number;
  provider_id: number;
  package_id: number;
  booking_date: string;
  start_time: string;
  end_time: string;
  total_price: number;
  status: 'pending' | 'confirmed' | 'completed' | 'cancelled';
  location?: string;
  special_notes?: string;
  created_at: string;
}

// Message
interface Message {
  id: number;
  conversation_id: number;
  sender_id: number;
  content: string;
  is_read: boolean;
  created_at: string;
}

// Notification
interface Notification {
  id: number;
  user_id: number;
  type: string;
  title: string;
  message: string;
  metadata?: Record<string, any>;
  is_read: boolean;
  created_at: string;
}

// Review
interface Review {
  review_id: number;
  provider_id: number;
  client_id: number;
  booking_id: number;
  rating: number;
  comment: string;
  is_verified: boolean;
  created_at: string;
}
```

---

## ⚠️ Error Handling

### Error Response Format
```typescript
{
  "error": "Error message here"
}
```

### Common HTTP Status Codes
- `200 OK` - Success
- `201 Created` - Resource created
- `400 Bad Request` - Invalid input
- `401 Unauthorized` - Missing/invalid token
- `403 Forbidden` - No permission
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error

### Error Handling Example
```typescript
try {
  const response = await fetch('/api/bookings', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(bookingData)
  });
  
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Request failed');
  }
  
  const data = await response.json();
  return data;
} catch (error) {
  console.error('Booking failed:', error);
  alert(error.message);
}
```

---

## 🛡️ Security Notes

1. **Always include Authorization header** for protected endpoints
2. **Store JWT token securely** (localStorage or httpOnly cookie)
3. **Validate user input** before sending to API
4. **Handle token expiration** (401 errors → redirect to login)
5. **Use HTTPS** in production
6. **Never store sensitive data** in localStorage
7. **Implement CSRF protection** if using cookies
8. **Rate limit** API calls on frontend

---

## 🚀 Quick Start Example

```typescript
// api.ts - API Client
import axios from 'axios';

const API_BASE_URL = 'http://localhost:8080';

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json'
  }
});

// Add token to requests
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Handle errors
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

export default api;

// Usage
import api from './api';

// Login
const login = async (email: string, password: string) => {
  const response = await api.post('/auth/login', { email, password });
  localStorage.setItem('token', response.data.token);
  return response.data;
};

// Get profile
const getProfile = async () => {
  const response = await api.get('/profile/me');
  return response.data;
};

// Browse providers
const browseProviders = async (filters: any) => {
  const response = await api.get('/browse/v2', { params: filters });
  return response.data;
};

// Create booking
const createBooking = async (bookingData: any) => {
  const response = await api.post('/bookings', bookingData);
  return response.data;
};
```

---

## 📞 Support

**Documentation:**
- `MESSAGING_GUIDE.md` - Messaging system
- `NOTIFICATION_GUIDE.md` - Notifications
- `REPORT_GUIDE.md` - Report system
- `ANALYTICS_GUIDE.md` - Analytics dashboard
- `BLOCK_GUIDE.md` - Block user system
- `SECURITY.md` - Security best practices

**Contact:** API Team  
**Version:** 1.0.0  
**Last Updated:** November 13, 2025

---

Made with ❤️ for SkillMatch Frontend Team
