# 🚀 SkillMatch API - คู่มือเชื่อมต่อสำหรับ Frontend

> **วันที่อัพเดท**: 2 ธันวาคม 2025  
> **สถานะ**: ✅ Backend พร้อมใช้งาน 100%  
> **API Version**: v1.0

---

## 📋 สารบัญ
1. [ภาพรวมระบบ](#ภาพรวมระบบ)
2. [การเริ่มต้นใช้งาน](#การเริ่มต้นใช้งาน)
3. [Authentication](#authentication)
4. [API Endpoints หลัก](#api-endpoints-หลัก)
5. [WebSocket Real-time](#websocket-real-time)
6. [Payment Integration](#payment-integration)
7. [Error Handling](#error-handling)

---

## 🎯 ภาพรวมระบบ

### ✅ สถานะ Backend (พร้อมใช้งาน)

| Component | Status | URL/Details |
|-----------|--------|-------------|
| **API Server** | ✅ Running | `http://localhost:8080` |
| **Database** | ✅ Ready | 30 tables, all migrations completed |
| **Redis Cache** | ✅ Connected | localhost:6379 |
| **WebSocket** | ✅ Ready | `ws://localhost:8080/ws` |
| **Google OAuth** | ✅ Configured | Client ID provided |
| **Stripe Payment** | ✅ Test Mode | Webhook configured |
| **Total Endpoints** | 118 | Public: 18, Protected: 85, Admin: 15 |

### 📊 ข้อมูลในระบบ

#### Service Categories (5 หมวดหมู่)
```json
[
  {"id": 1, "name": "Massage", "name_thai": "นวด", "icon": "💆"},
  {"id": 2, "name": "Spa", "name_thai": "สปา", "icon": "🧖"},
  {"id": 3, "name": "Beauty", "name_thai": "ความงาม", "icon": "💄"},
  {"id": 4, "name": "Wellness", "name_thai": "สุขภาพ", "icon": "🧘"},
  {"id": 5, "name": "Therapy", "name_thai": "บำบัด", "icon": "🩺"}
]
```

#### Subscription Tiers (5 ระดับ)
```json
[
  {"tier_id": 1, "name": "General", "price": 0, "access_level": 0},
  {"tier_id": 2, "name": "Silver", "price": 9.99, "access_level": 1},
  {"tier_id": 3, "name": "Diamond", "price": 29.99, "access_level": 2},
  {"tier_id": 4, "name": "Premium", "price": 99.99, "access_level": 3},
  {"tier_id": 5, "name": "GOD", "price": 9999.99, "access_level": 999}
]
```

### 🔑 Test Account (GOD Admin)

```json
{
  "user_id": 1,
  "username": "The BOB Film",
  "email": "audikoratair@gmail.com",
  "tier_id": 5,
  "tier_name": "GOD",
  "is_admin": true,
  "verification_status": "unverified"
}
```

**JWT Token (ใช้ได้ 7 วัน)**:
```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxIiwiZXhwIjoxNzY0NzQ3MjU5LCJpYXQiOjE3NjQ2NjA4NTl9.Sdu1pra-ADzEAeakCwPI1hfm5906CSM25qYD0U3cFmk
```

---

## 🚀 การเริ่มต้นใช้งาน

### 1. Setup ใน Frontend

```javascript
// config.js
export const API_CONFIG = {
  BASE_URL: 'http://localhost:8080',
  WS_URL: 'ws://localhost:8080/ws',
  GOOGLE_CLIENT_ID: '171089417301-each0gvj9d5l38bgkklu0n36p5eo5eau.apps.googleusercontent.com'
};

// api.js - Helper function
export async function apiCall(endpoint, options = {}) {
  const token = localStorage.getItem('jwt_token');
  
  const config = {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
      ...(token && { 'Authorization': `Bearer ${token}` })
    }
  };
  
  try {
    const response = await fetch(`${API_CONFIG.BASE_URL}${endpoint}`, config);
    const data = await response.json();
    
    if (!response.ok) {
      throw new Error(data.error || data.message || 'API Error');
    }
    
    return data;
  } catch (error) {
    console.error('API Error:', error);
    throw error;
  }
}
```

### 2. ทดสอบการเชื่อมต่อ

```javascript
// ทดสอบว่า API ทำงานหรือไม่
async function testConnection() {
  const result = await apiCall('/ping');
  console.log(result); // { message: "pong", postgres_time: "..." }
}

// ดึงข้อมูล Service Categories
async function getCategories() {
  const result = await apiCall('/service-categories');
  console.log(result); // { categories: [...], total: 5 }
}

// ดึงข้อมูล Tiers
async function getTiers() {
  const result = await apiCall('/tiers');
  console.log(result); // [{ tier_id: 1, name: "General", ... }, ...]
}
```

---

## 🔐 Authentication

### 1. Register with Email Verification

**Flow**:
1. ส่งอีเมลเพื่อรับ OTP (6 หลัก)
2. ยืนยัน OTP
3. สร้างบัญชี
4. Login และรับ JWT token

```javascript
// Step 1: Send Verification Email
async function sendVerification(email) {
  const result = await apiCall('/auth/send-verification', {
    method: 'POST',
    body: JSON.stringify({ email })
  });
  console.log(result); // { message: "Verification code sent to email" }
}

// Step 2: Verify OTP
async function verifyEmail(email, otp) {
  const result = await apiCall('/auth/verify-email', {
    method: 'POST',
    body: JSON.stringify({ email, otp })
  });
  console.log(result); // { message: "Email verified", verified: true }
}

// Step 3: Register
async function register(userData) {
  const result = await apiCall('/register', {
    method: 'POST',
    body: JSON.stringify({
      username: "JohnDoe",
      email: "john@example.com",
      password: "Password123!",
      gender_id: 1, // 1=Male, 2=Female, 3=Other, 4=Prefer not to say
      otp: "123456" // From verification step
    })
  });
  
  // Save token
  localStorage.setItem('jwt_token', result.token);
  return result; // { message: "...", token: "...", user: {...} }
}
```

### 2. Login

```javascript
async function login(email, password) {
  const result = await apiCall('/login', {
    method: 'POST',
    body: JSON.stringify({ email, password })
  });
  
  // Save token
  localStorage.setItem('jwt_token', result.token);
  localStorage.setItem('user', JSON.stringify(result.user));
  
  return result;
}
```

### 3. Google OAuth

```html
<!-- Add Google Sign-In Button -->
<script src="https://accounts.google.com/gsi/client" async defer></script>

<div id="g_id_onload"
     data-client_id="171089417301-each0gvj9d5l38bgkklu0n36p5eo5eau.apps.googleusercontent.com"
     data-callback="handleGoogleCallback">
</div>
<div class="g_id_signin" data-type="standard"></div>
```

```javascript
// Handle Google OAuth callback
async function handleGoogleCallback(response) {
  try {
    const result = await apiCall('/auth/google', {
      method: 'POST',
      body: JSON.stringify({
        code: response.credential // Google returns 'credential' not 'code'
      })
    });
    
    // Save token and user data
    localStorage.setItem('jwt_token', result.token);
    localStorage.setItem('user', JSON.stringify(result.user));
    
    console.log('Logged in:', result.user);
    // Redirect to dashboard
  } catch (error) {
    console.error('Google login failed:', error);
  }
}
```

### 4. Get Current User

```javascript
async function getCurrentUser() {
  const user = await apiCall('/profile'); // or /users/me
  return user;
}
```

### 5. Logout

```javascript
function logout() {
  localStorage.removeItem('jwt_token');
  localStorage.removeItem('user');
  // Redirect to login page
}
```

---

## 📡 API Endpoints หลัก

### 🔓 Public Endpoints (ไม่ต้อง Authentication)

#### 1. Get Service Categories
```javascript
GET /service-categories

// Response
{
  "categories": [
    {
      "category_id": 1,
      "name": "Massage",
      "name_thai": "นวด",
      "description": "Professional massage services",
      "icon": "💆",
      "is_adult": false,
      "display_order": 1,
      "is_active": true
    }
  ],
  "total": 5
}
```

#### 2. Browse Providers by Category
```javascript
GET /categories/:category_id/providers?page=1&limit=20

// Example
const providers = await apiCall('/categories/1/providers?page=1&limit=20');

// Response
{
  "providers": [
    {
      "user_id": 123,
      "username": "Provider Name",
      "profile_image_url": "/uploads/...",
      "bio": "...",
      "rating_avg": 4.5,
      "review_count": 25,
      "service_type": "Both"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 50
  }
}
```

#### 3. Get Provider Public Profile
```javascript
GET /provider/:userId/public

const profile = await apiCall('/provider/123/public');

// Response
{
  "user_id": 123,
  "username": "Provider Name",
  "bio": "...",
  "profile_image_url": "...",
  "rating_avg": 4.5,
  "review_count": 25,
  "service_type": "Both",
  "categories": ["Massage", "Spa"]
}
```

#### 4. Get Provider Photos
```javascript
GET /provider/:userId/photos

const photos = await apiCall('/provider/123/photos');

// Response
{
  "photos": [
    {
      "photo_id": 1,
      "photo_url": "/uploads/...",
      "sort_order": 1,
      "caption": "My workspace",
      "uploaded_at": "2025-01-01T..."
    }
  ]
}
```

#### 5. Get Provider Packages
```javascript
GET /packages/:providerId

const packages = await apiCall('/packages/123');

// Response
{
  "packages": [
    {
      "package_id": 1,
      "name": "1 Hour Massage",
      "description": "...",
      "price": 500,
      "duration_minutes": 60
    }
  ]
}
```

#### 6. Get Provider Reviews
```javascript
GET /reviews/:providerId?page=1&limit=10

const reviews = await apiCall('/reviews/123?page=1&limit=10');

// Response
{
  "reviews": [
    {
      "review_id": 1,
      "rating": 5,
      "comment": "Great service!",
      "client_username": "John",
      "created_at": "..."
    }
  ],
  "pagination": {...}
}
```

#### 7. Get Provider Review Stats
```javascript
GET /reviews/stats/:providerId

const stats = await apiCall('/reviews/stats/123');

// Response
{
  "provider_id": 123,
  "total_reviews": 25,
  "average_rating": 4.5,
  "rating_breakdown": {
    "5": 15,
    "4": 8,
    "3": 2,
    "2": 0,
    "1": 0
  }
}
```

#### 8. Check if Provider is Favorited
```javascript
GET /favorites/check/:providerId

// ถ้าไม่มี token จะ return false
// ถ้ามี token จะ return สถานะจริง
const result = await apiCall('/favorites/check/123');

// Response
{
  "is_favorite": true
}
```

---

### 🔐 Protected Endpoints (ต้องมี JWT Token)

#### 1. Get Current User Profile
```javascript
GET /profile  // or /users/me

const profile = await apiCall('/profile');

// Response
{
  "user_id": 1,
  "username": "The BOB Film",
  "email": "audikoratair@gmail.com",
  "tier_id": 5,
  "tier_name": "GOD",
  "is_admin": true,
  "profile_image_url": null,
  "bio": null,
  "phone": null,
  "verification_status": "unverified"
}
```

#### 2. Update Profile
```javascript
PUT /profile/me

await apiCall('/profile/me', {
  method: 'PUT',
  body: JSON.stringify({
    bio: "New bio",
    phone: "0812345678"
  })
});
```

#### 3. Add to Favorites
```javascript
POST /favorites

await apiCall('/favorites', {
  method: 'POST',
  body: JSON.stringify({
    provider_id: 123
  })
});
```

#### 4. Remove from Favorites
```javascript
DELETE /favorites/:providerId

await apiCall('/favorites/123', {
  method: 'DELETE'
});
```

#### 5. Get My Favorites
```javascript
GET /favorites

const favorites = await apiCall('/favorites');

// Response
{
  "favorites": [
    {
      "provider_id": 123,
      "username": "Provider Name",
      "profile_image_url": "...",
      "rating_avg": 4.5
    }
  ]
}
```

#### 6. Create Booking with Payment
```javascript
POST /bookings/create-with-payment

const result = await apiCall('/bookings/create-with-payment', {
  method: 'POST',
  body: JSON.stringify({
    provider_id: 123,
    package_id: 1,
    booking_date: "2025-12-10",
    booking_time: "14:00:00",
    notes: "Please bring essential oils"
  })
});

// Response
{
  "booking_id": 456,
  "checkout_url": "https://checkout.stripe.com/...",
  "message": "Redirect to checkout"
}

// Redirect user to checkout_url
window.location.href = result.checkout_url;
```

#### 7. Get My Bookings
```javascript
GET /bookings/my?status=all

const bookings = await apiCall('/bookings/my?status=paid');

// Response
{
  "bookings": [
    {
      "booking_id": 456,
      "provider_username": "Provider Name",
      "package_name": "1 Hour Massage",
      "booking_date": "2025-12-10",
      "booking_time": "14:00:00",
      "status": "paid",
      "total_price": 500
    }
  ]
}
```

#### 8. Update Booking Status (Provider Only)
```javascript
PATCH /bookings/:id/status

await apiCall('/bookings/456/status', {
  method: 'PATCH',
  body: JSON.stringify({
    status: "confirmed" // confirmed, completed, cancelled
  })
});
```

#### 9. Create Review
```javascript
POST /reviews

await apiCall('/reviews', {
  method: 'POST',
  body: JSON.stringify({
    booking_id: 456,
    provider_id: 123,
    rating: 5,
    comment: "Excellent service!"
  })
});
```

---

## 💬 WebSocket Real-time

### 1. Connect to WebSocket

```javascript
class ChatWebSocket {
  constructor(token) {
    this.ws = null;
    this.token = token;
    this.reconnectDelay = 1000;
  }
  
  connect() {
    this.ws = new WebSocket('ws://localhost:8080/ws');
    
    this.ws.onopen = () => {
      console.log('✅ WebSocket Connected');
      
      // Authenticate after connection
      this.send({
        type: 'auth',
        payload: { token: this.token }
      });
    };
    
    this.ws.onmessage = (event) => {
      const message = JSON.parse(event.data);
      this.handleMessage(message);
    };
    
    this.ws.onerror = (error) => {
      console.error('❌ WebSocket Error:', error);
    };
    
    this.ws.onclose = () => {
      console.log('🔌 WebSocket Disconnected');
      // Auto reconnect
      setTimeout(() => this.connect(), this.reconnectDelay);
    };
  }
  
  handleMessage(message) {
    switch(message.type) {
      case 'new_message':
        // Handle new chat message
        console.log('New message:', message.payload);
        this.onNewMessage(message.payload);
        break;
        
      case 'notification':
        // Handle notification
        console.log('New notification:', message.payload);
        this.onNotification(message.payload);
        break;
        
      case 'booking_update':
        // Handle booking status change
        console.log('Booking updated:', message.payload);
        this.onBookingUpdate(message.payload);
        break;
        
      case 'typing':
        // Handle typing indicator
        this.onTyping(message.payload);
        break;
    }
  }
  
  send(message) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    }
  }
  
  sendTyping(conversationId) {
    this.send({
      type: 'typing',
      payload: { conversation_id: conversationId }
    });
  }
  
  disconnect() {
    if (this.ws) {
      this.ws.close();
    }
  }
  
  // Callbacks (override these)
  onNewMessage(message) {}
  onNotification(notification) {}
  onBookingUpdate(booking) {}
  onTyping(data) {}
}

// Usage
const token = localStorage.getItem('jwt_token');
const chatWs = new ChatWebSocket(token);

chatWs.onNewMessage = (message) => {
  // Update UI with new message
  console.log('Received:', message);
};

chatWs.connect();
```

### 2. Messaging Endpoints

```javascript
// Get Conversations List
GET /conversations

const conversations = await apiCall('/conversations');

// Response
{
  "conversations": [
    {
      "conversation_id": 1,
      "other_user_id": 123,
      "other_user_username": "Provider Name",
      "last_message": "Hello!",
      "last_message_time": "...",
      "unread_count": 3
    }
  ]
}

// Get Messages in Conversation
GET /conversations/:id/messages?limit=50&offset=0

const messages = await apiCall('/conversations/1/messages?limit=50');

// Response
{
  "messages": [
    {
      "message_id": 1,
      "sender_id": 123,
      "content": "Hello!",
      "is_read": false,
      "created_at": "..."
    }
  ]
}

// Send Message
POST /messages

await apiCall('/messages', {
  method: 'POST',
  body: JSON.stringify({
    receiver_id: 123,
    content: "Hi there!"
  })
});

// Mark Messages as Read
PATCH /messages/read

await apiCall('/messages/read', {
  method: 'PATCH',
  body: JSON.stringify({
    conversation_id: 1
  })
});
```

### 3. Notifications

```javascript
// Get Notifications
GET /notifications?limit=20

const notifications = await apiCall('/notifications?limit=20');

// Get Unread Count
GET /notifications/unread/count

const count = await apiCall('/notifications/unread/count');
// Response: { unread_count: 5 }

// Mark as Read
PATCH /notifications/:id/read

await apiCall('/notifications/123/read', {
  method: 'PATCH'
});

// Mark All as Read
PATCH /notifications/read-all

await apiCall('/notifications/read-all', {
  method: 'PATCH'
});
```

---

## 💳 Payment Integration (Stripe)

### Booking Payment Flow

```javascript
// 1. Create Booking (returns Stripe Checkout URL)
const result = await apiCall('/bookings/create-with-payment', {
  method: 'POST',
  body: JSON.stringify({
    provider_id: 123,
    package_id: 1,
    booking_date: "2025-12-10",
    booking_time: "14:00:00"
  })
});

// 2. Redirect to Stripe Checkout
window.location.href = result.checkout_url;

// 3. After payment, Stripe redirects back to your success page
// The webhook will update booking status to "paid"

// 4. Check booking status
const booking = await apiCall('/bookings/my');
// status will be "paid" after successful payment
```

### Subscription Payment Flow

```javascript
// 1. Create Stripe Checkout Session
const result = await apiCall('/subscription/create-checkout', {
  method: 'POST',
  body: JSON.stringify({
    tier_id: 2 // Silver tier
  })
});

// 2. Redirect to Stripe
window.location.href = result.checkout_url;

// 3. After payment, user's tier_id will be updated automatically
```

---

## ⚠️ Error Handling

### Standard Error Response

```javascript
{
  "error": "Error message",
  "details": "More detailed error information"
}
```

### Common HTTP Status Codes

- **200**: Success
- **201**: Created successfully
- **400**: Bad request (validation error)
- **401**: Unauthorized (invalid/missing token)
- **403**: Forbidden (insufficient permissions)
- **404**: Not found
- **500**: Internal server error

### Error Handling Example

```javascript
async function safeApiCall(endpoint, options) {
  try {
    const result = await apiCall(endpoint, options);
    return { success: true, data: result };
  } catch (error) {
    // Handle specific errors
    if (error.message.includes('401') || error.message.includes('Unauthorized')) {
      // Token expired or invalid
      localStorage.removeItem('jwt_token');
      window.location.href = '/login';
    }
    
    return { success: false, error: error.message };
  }
}
```

---

## 🔧 Provider Registration

```javascript
// Flow: Send OTP → Verify → Register Provider
async function registerProvider(data) {
  // 1. Send OTP
  await apiCall('/auth/send-verification', {
    method: 'POST',
    body: JSON.stringify({ email: data.email })
  });
  
  // 2. User enters OTP (wait for user input)
  
  // 3. Register as Provider
  const result = await apiCall('/register/provider', {
    method: 'POST',
    body: JSON.stringify({
      username: data.username,
      email: data.email,
      password: data.password,
      gender_id: data.gender_id,
      phone: data.phone, // 10 digits
      otp: data.otp, // 6 digits
      category_ids: [1, 2], // 1-5 categories
      service_type: "Both", // "Incall", "Outcall", or "Both"
      bio: data.bio, // Optional
      province: data.province, // Optional
      district: data.district // Optional
    })
  });
  
  // Save token
  localStorage.setItem('jwt_token', result.token);
  return result;
}
```

---

## 📝 Important Notes

### 1. JWT Token
- **Expiration**: 7 days
- **Storage**: localStorage
- **Format**: `Authorization: Bearer <token>`
- **Refresh**: Re-login when expired

### 2. Fee Structure
- **Total Fee**: 12.75%
  - Stripe: 2.75%
  - Platform: 10%
- **Provider receives**: 87.25% of booking price
- **Only providers see fee breakdown**

### 3. Booking Flow
```
Create Booking → Pay (Stripe) → Status: paid → Provider Confirms → Status: confirmed → Complete → Status: completed → Can Review
```

### 4. Provider Verification
```
Register → Upload Documents → Status: pending → Admin Reviews → Status: approved → Visible to clients
```

### 5. Message Restrictions
- Users can **ONLY** send templated/automated messages
- Direct contact exchange is **NOT ALLOWED**
- All communication must be through the platform

### 6. Wallet System
- **Pending Balance**: Held for 7 days after booking completion
- **Available Balance**: Can request withdrawal
- **Withdrawal**: Provider requests → Admin approves → Transfer via platform bank

### 7. Tier System - ระบบระดับที่ต้องเข้าใจ

**⚠️ สำคัญมาก: มี 2 ระบบ Tier แยกกัน**

#### 🎫 Client Tier (tier_id) - ระดับสมาชิกผู้ใช้บริการ
ใช้สำหรับ **ผู้ใช้บริการ (Client)** ที่จ่ายค่า Subscription เพื่อใช้แพลตฟอร์ม

| Tier | Name | Price/Month | Access Level | Features |
|------|------|-------------|--------------|----------|
| 1 | General | ฟรี | 0 | ใช้งานพื้นฐาน, ดู provider จำกัด |
| 2 | Silver | 9.99 | 1 | ดู provider เพิ่มเติม, จอง priority |
| 3 | Diamond | 29.99 | 2 | ดูทุก provider, ส่วนลดพิเศษ |
| 4 | Premium | 99.99 | 3 | VIP access, ส่วนลดสูงสุด |
| 5 | GOD | 9999.99 | 999 | Admin full access |

**การแสดงผล**: แสดงที่มุมขวาบนของ profile client หรือใน dashboard

```javascript
// ตรวจสอบ tier ของ client
const user = await apiCall('/profile');
console.log(`User tier: ${user.tier_name} (Level ${user.tier_id})`);

// แสดงใน UI
if (user.tier_id === 1) {
  showBadge('GENERAL', 'gray');
} else if (user.tier_id === 2) {
  showBadge('SILVER', 'silver');
} else if (user.tier_id === 3) {
  showBadge('DIAMOND', 'blue');
} else if (user.tier_id === 4) {
  showBadge('PREMIUM', 'gold');
}
```

#### 👔 Provider Level (provider_level_id) - ระดับผู้ให้บริการ
ใช้สำหรับ **ผู้ให้บริการ (Provider)** คำนวณอัตโนมัติจากประสิทธิภาพการทำงาน

| Level | Name | Points | Criteria | Badge Color |
|-------|------|--------|----------|-------------|
| 1 | General | 0-99 | เริ่มต้น | Gray |
| 2 | Silver | 100-249 | มีรีวิว, rating ดี | Silver |
| 3 | Diamond | 250-399 | booking มาก, rating สูง | Blue |
| 4 | Premium | 400+ | Top performer | Gold |

**การคำนวณ Provider Level**:
```
Points = (avg_rating × 20) + (total_bookings × 5) + (total_reviews × 3) 
         + (response_rate × 0.5) + (acceptance_rate × 0.5)
Max Points = 600
```

**การแสดงผล**: แสดงที่โปรไฟล์ provider แบบ Badge หรือ Star Rating

```javascript
// ดึงข้อมูล provider พร้อม level
const provider = await apiCall('/provider/123/public');

// Response จะมี
{
  "user_id": 123,
  "username": "Provider Name",
  "provider_level_id": 3,  // ⚠️ คนละตัวกับ tier_id
  "provider_level_name": "Diamond",
  "tier_id": 1,  // tier สมาชิกของตัวเอง (ใช้ไม่ค่อยเกี่ยวกับการแสดงผล)
  "rating_avg": 4.8,
  "total_bookings": 120,
  "total_reviews": 95
}

// แสดงใน UI
function showProviderBadge(provider) {
  const badges = {
    1: { name: 'General', color: '#gray', icon: '⭐' },
    2: { name: 'Silver', color: '#C0C0C0', icon: '⭐⭐' },
    3: { name: 'Diamond', color: '#4A90E2', icon: '⭐⭐⭐' },
    4: { name: 'Premium', color: '#FFD700', icon: '⭐⭐⭐⭐' }
  };
  
  const badge = badges[provider.provider_level_id];
  return `<span style="color: ${badge.color}">${badge.icon} ${badge.name}</span>`;
}
```

---

### 8. Package Display Logic - การแสดง Package ตามระดับ

**กฎการแสดง Package**:

#### สำหรับ Client ที่ยังไม่ Login
```javascript
// แสดงเฉพาะ provider ระดับ General และ Silver เท่านั้น
// ไม่แสดง Diamond และ Premium

const providers = await apiCall('/categories/1/providers');
// Backend จะ filter ให้อัตโนมัติ (ถ้าไม่มี token)

// หรือถ้ามี token แต่เป็น General tier
// จะเห็นเฉพาะ General และ Silver providers
```

#### สำหรับ Client ที่ Login แล้ว
```javascript
const user = await apiCall('/profile');

// ตาราง: Client Tier vs Provider Level ที่เห็น
// Client Tier 1 (General)   → เห็น Provider: General, Silver
// Client Tier 2 (Silver)    → เห็น Provider: General, Silver, Diamond
// Client Tier 3 (Diamond)   → เห็น Provider: General, Silver, Diamond, Premium
// Client Tier 4+ (Premium+) → เห็นทุก Provider Level

// ตัวอย่าง UI Logic
async function canViewProvider(clientTierId, providerLevelId) {
  const accessMatrix = {
    1: [1, 2],           // General client sees: General, Silver providers
    2: [1, 2, 3],        // Silver client sees: General, Silver, Diamond
    3: [1, 2, 3, 4],     // Diamond client sees: All
    4: [1, 2, 3, 4],     // Premium sees: All
    5: [1, 2, 3, 4]      // GOD sees: All
  };
  
  return accessMatrix[clientTierId]?.includes(providerLevelId) || false;
}

// ใช้ใน UI
if (!canViewProvider(currentUser.tier_id, provider.provider_level_id)) {
  showUpgradePrompt('คุณต้อง upgrade เป็น Silver tier เพื่อดู provider ระดับนี้');
}
```

#### การแสดง Package ของ Provider
```javascript
// ดึง packages จาก provider
const packages = await apiCall('/packages/123');

// Response
{
  "packages": [
    {
      "package_id": 1,
      "name": "1 Hour Massage",
      "description": "Traditional Thai massage",
      "price": 500,
      "duration_minutes": 60,
      "is_active": true
    },
    {
      "package_id": 2,
      "name": "2 Hour Spa Package",
      "price": 1200,
      "duration_minutes": 120
    }
  ],
  "provider_info": {
    "provider_level_id": 3,
    "provider_level_name": "Diamond"
  }
}

// UI Display Logic
function displayPackages(packages, providerLevel, clientTier) {
  // 1. แสดง Provider Level Badge
  showProviderBadge(providerLevel);
  
  // 2. แสดง Packages
  packages.forEach(pkg => {
    // แสดงราคา + duration
    showPackageCard(pkg);
  });
  
  // 3. ถ้า client ไม่สามารถจองได้ (tier ต่ำเกิน)
  if (!canViewProvider(clientTier, providerLevel)) {
    disableBookingButton();
    showUpgradeMessage();
  }
}
```

---

### 9. UI Display Examples - ตัวอย่างการแสดงผลจริง

#### Provider Card ในหน้า Browse
```html
<div class="provider-card">
  <!-- Provider Level Badge (ด้านบน) -->
  <div class="provider-badge diamond">
    ⭐⭐⭐ Diamond Provider
  </div>
  
  <!-- Provider Image -->
  <img src="/uploads/provider-123.jpg" />
  
  <!-- Provider Info -->
  <h3>Provider Name</h3>
  <p>⭐ 4.8 (120 reviews)</p>
  
  <!-- Service Type -->
  <span class="service-type">📍 Both (Incall/Outcall)</span>
  
  <!-- Categories -->
  <div class="categories">
    <span>💆 Massage</span>
    <span>🧖 Spa</span>
  </div>
  
  <!-- Pricing -->
  <p class="price-range">฿500 - ฿2,000</p>
  
  <!-- ถ้า client tier ต่ำเกิน -->
  <div class="locked-overlay" v-if="!canView">
    🔒 Upgrade to Silver to view this provider
    <button>Upgrade Now</button>
  </div>
</div>
```

#### Provider Profile Page
```html
<div class="provider-profile">
  <!-- Header with Badge -->
  <div class="profile-header">
    <img src="profile.jpg" class="profile-image" />
    <div class="info">
      <h1>Provider Name</h1>
      
      <!-- Provider Level Badge (ใหญ่) -->
      <div class="level-badge diamond">
        <span class="stars">⭐⭐⭐</span>
        <span class="level-name">Diamond Provider</span>
        <span class="level-points">350 points</span>
      </div>
      
      <!-- Stats -->
      <div class="stats">
        <span>⭐ 4.8 average</span>
        <span>📦 120 bookings</span>
        <span>💬 95 reviews</span>
        <span>📍 Bangkok</span>
      </div>
    </div>
  </div>
  
  <!-- Service Packages Section -->
  <div class="packages-section">
    <h2>Service Packages</h2>
    
    <div class="package-card" v-for="pkg in packages">
      <h3>{{ pkg.name }}</h3>
      <p>{{ pkg.description }}</p>
      <div class="package-details">
        <span class="duration">⏱ {{ pkg.duration_minutes }} minutes</span>
        <span class="price">฿{{ pkg.price }}</span>
      </div>
      
      <!-- ปุ่มจอง -->
      <button 
        :disabled="!canBook" 
        @click="bookPackage(pkg.package_id)">
        {{ canBook ? 'Book Now' : '🔒 Upgrade to Book' }}
      </button>
    </div>
  </div>
</div>
```

#### Client Dashboard (แสดง Client Tier ของตัวเอง)
```html
<div class="user-dashboard">
  <!-- Client Tier Badge (มุมขวาบน) -->
  <div class="user-tier-badge silver">
    <span class="tier-icon">🎫</span>
    <span class="tier-name">Silver Member</span>
    <button @click="showUpgradeOptions">Upgrade</button>
  </div>
  
  <!-- Benefits Based on Tier -->
  <div class="tier-benefits">
    <h3>Your Benefits (Silver Tier)</h3>
    <ul>
      <li>✅ Access to General & Silver providers</li>
      <li>✅ Priority booking</li>
      <li>✅ 5% discount on all bookings</li>
      <li>❌ Access to Diamond providers (Upgrade to Diamond)</li>
      <li>❌ Access to Premium providers (Upgrade to Diamond)</li>
    </ul>
  </div>
</div>
```

---

### 10. API Response Examples - ตัวอย่าง Response จริง

#### Browse Providers (filtered by client tier)
```javascript
// Client: Silver tier (tier_id = 2)
GET /categories/1/providers?page=1&limit=20

// Response: จะเห็นเฉพาะ provider_level_id 1, 2, 3
{
  "providers": [
    {
      "user_id": 101,
      "username": "Massage Pro",
      "provider_level_id": 2,  // Silver
      "provider_level_name": "Silver",
      "rating_avg": 4.5,
      "review_count": 50,
      "profile_image_url": "/uploads/...",
      "service_type": "Both"
    },
    {
      "user_id": 102,
      "username": "Spa Expert",
      "provider_level_id": 3,  // Diamond
      "provider_level_name": "Diamond",
      "rating_avg": 4.8,
      "review_count": 120
    }
    // ❌ ไม่มี provider_level_id = 4 (Premium) เพราะ client เป็น Silver
  ],
  "your_tier": {
    "tier_id": 2,
    "tier_name": "Silver",
    "can_view_levels": [1, 2, 3]
  }
}
```

#### Provider Full Profile
```javascript
GET /provider/123

// Response
{
  "user_id": 123,
  "username": "Premium Spa",
  
  // Provider Level (แสดงระดับความเป็นมืออาชีพ)
  "provider_level_id": 4,
  "provider_level_name": "Premium",
  "provider_level_points": 450,
  
  // Client Tier ของตัวเอง (ไม่ค่อยแสดงใน UI)
  "tier_id": 1,
  "tier_name": "General",
  
  // Stats
  "rating_avg": 4.9,
  "total_bookings": 200,
  "total_reviews": 180,
  "response_rate": 98,
  "acceptance_rate": 95,
  
  // Service Info
  "service_type": "Both",
  "categories": [
    {"category_id": 1, "name": "Massage", "name_thai": "นวด"},
    {"category_id": 2, "name": "Spa", "name_thai": "สปา"}
  ],
  
  // Packages
  "packages": [
    {
      "package_id": 1,
      "name": "Premium Massage 90min",
      "price": 1500,
      "duration_minutes": 90
    }
  ]
}
```

---

### 11. Frontend State Management - จัดการ State

```javascript
// store/user.js - User Store
export const useUserStore = {
  state: {
    user: null,
    isLoggedIn: false
  },
  
  getters: {
    clientTier: (state) => state.user?.tier_id || 1,
    clientTierName: (state) => state.user?.tier_name || 'General',
    
    canViewProviderLevel: (state) => (providerLevelId) => {
      const tier = state.user?.tier_id || 1;
      const accessMatrix = {
        1: [1, 2],
        2: [1, 2, 3],
        3: [1, 2, 3, 4],
        4: [1, 2, 3, 4],
        5: [1, 2, 3, 4]
      };
      return accessMatrix[tier]?.includes(providerLevelId) || false;
    },
    
    nextTierBenefits: (state) => {
      const currentTier = state.user?.tier_id || 1;
      const benefits = {
        2: 'Access Diamond providers + 5% discount',
        3: 'Access Premium providers + 10% discount',
        4: 'VIP access + 15% discount'
      };
      return benefits[currentTier + 1] || 'Max tier reached';
    }
  }
};

// components/ProviderCard.vue
export default {
  computed: {
    canViewProvider() {
      return this.$store.getters.canViewProviderLevel(this.provider.provider_level_id);
    },
    
    needsUpgrade() {
      return !this.canViewProvider;
    },
    
    upgradeMessage() {
      const nextTier = this.$store.state.user.tier_id + 1;
      const tierNames = ['', 'General', 'Silver', 'Diamond', 'Premium'];
      return `Upgrade to ${tierNames[nextTier]} to view this provider`;
    }
  }
};
```

---

## 🧪 Testing Checklist

### Basic Tests
```bash
# 1. Test API is running
curl http://localhost:8080/ping

# 2. Test categories
curl http://localhost:8080/service-categories

# 3. Test tiers
curl http://localhost:8080/tiers

# 4. Test profile (with token)
curl -H "Authorization: Bearer YOUR_TOKEN" http://localhost:8080/profile
```

### Frontend Integration Tests
- [ ] Register new user
- [ ] Login with email/password
- [ ] Login with Google OAuth
- [ ] Browse providers by category
- [ ] View provider profile
- [ ] Add/remove favorites
- [ ] Create booking with payment
- [ ] Send message via WebSocket
- [ ] Receive notifications
- [ ] Create review

---

## 📞 Support

### Quick Commands
```bash
# Check if server is running
curl http://localhost:8080/ping

# Get categories
curl http://localhost:8080/service-categories

# Test with GOD token
curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxIiwiZXhwIjoxNzY0NzQ3MjU5LCJpYXQiOjE3NjQ2NjA4NTl9.Sdu1pra-ADzEAeakCwPI1hfm5906CSM25qYD0U3cFmk" http://localhost:8080/profile
```

### Database Info
- **Total Tables**: 30
- **Migrations**: 32 (all completed ✅)
- **Test Data**: 5 service categories, 5 tiers, 1 GOD user

### Common Issues
1. **401 Unauthorized**: Check token format and expiration
2. **CORS errors**: Make sure frontend runs on allowed origins
3. **WebSocket disconnects**: Implement auto-reconnect
4. **Payment fails**: Check Stripe test mode and webhook

---

## ✨ Summary

### ✅ พร้อมใช้งาน
- ✅ 118 API Endpoints
- ✅ Authentication (JWT + Google OAuth)
- ✅ Real-time Messaging (WebSocket)
- ✅ Payment Integration (Stripe)
- ✅ File Uploads (Photos, Documents)
- ✅ Provider System (Categories, Packages, Reviews)
- ✅ Booking System with Payment
- ✅ Notification System
- ✅ Financial System (Wallets, Withdrawals)

### 🎯 Next Steps for Frontend
1. Setup API helper functions
2. Implement authentication pages
3. Create provider browsing UI
4. Build booking flow with Stripe
5. Implement WebSocket chat
6. Add notification system
7. Test all flows end-to-end

---

**🚀 Backend API พร้อมใช้งาน 100% - เริ่มพัฒนา Frontend ได้เลย!**

สำหรับเอกสาร API ฉบับเต็ม (118 endpoints) ดูที่: `FRONTEND_SETUP.md`
