# 🎯 SkillMatch API - Frontend Integration Guide

> **สำหรับ Frontend Developer**: คู่มือการใช้ API ทั้งหมด พร้อม Response Examples

---

## 📌 Base URL

```
Development: http://localhost:8080
Production: https://api.skillmatch.com
```

---

## 🔐 Authentication

### ทุก protected endpoint ต้องใส่ header:

```javascript
headers: {
  'Authorization': 'Bearer <JWT_TOKEN>',
  'Content-Type': 'application/json'
}
```

---

## 👤 User Roles & Permissions

| Role | tier_id | is_admin | Permissions |
|------|---------|----------|-------------|
| **GOD** | 5 | true | ทำทุกอย่างได้ รวมถึงลบ user, สร้าง admin |
| **Admin** | 2-4 | true | จัดการ KYC, reports, provider verification |
| **Provider** | 1-4 | false | ให้บริการ, รับ booking, ถอนเงิน |
| **User** | 1-4 | false | จอง, review, favorite |

---

## 🔑 GOD Endpoints (tier_id = 5 only)

### 1. **GET /admin/stats/god** - GOD Dashboard Statistics

**Response:**
```json
{
  "total_users": 150,
  "total_providers": 45,
  "total_admins": 3,
  "pending_verification": 8,
  "total_bookings": 320,
  "total_revenue": 125000.50,
  "active_users_24h": 25,
  "new_users_today": 5
}
```

---

### 2. **GET /admin/users** - List All Users (with filters)

**Query Parameters:**
- `page` (default: 1)
- `limit` (default: 50, max: 100)
- `is_admin` (true/false)
- `verification_status` (unverified/pending/verified/approved/rejected)
- `search` (username or email)

**Example Request:**
```javascript
GET /admin/users?page=1&limit=20&verification_status=verified&search=john
```

**Response:**
```json
{
  "users": [
    {
      "user_id": 2,
      "username": "john_doe",
      "email": "john@example.com",
      "gender_id": 1,
      "subscription_tier_id": 2,
      "provider_level_id": 3,
      "is_admin": false,
      "verification_status": "verified",
      "first_name": "John",
      "last_name": "Doe",
      "phone_number": "0812345678",
      "registration_date": "2025-01-15T10:30:00Z",
      "profile_image_url": "https://storage.googleapis.com/...",
      "age": 28,
      "tier_name": "Silver"
    }
  ],
  "total": 150,
  "page": 1,
  "limit": 20,
  "total_pages": 8
}
```

---

### 3. **POST /god/update-user** - Update User Role/Tier/Status

**Request Body:**
```json
{
  "user_id": 5,
  "is_admin": true,
  "tier_id": 3,
  "provider_level_id": 2,
  "verification_status": "verified"
}
```

**ฟิลด์ทุกตัวเป็น optional** - ส่งแค่ที่ต้องการเปลี่ยน

**Response:**
```json
{
  "message": "User role updated successfully",
  "user_id": 5
}
```

**⚠️ ห้าม**: แก้ไข GOD account (user_id = 1) ยกเว้น GOD เอง

---

### 4. **DELETE /admin/users/:user_id** - Delete Any User

**Example:**
```javascript
DELETE /admin/users/25
```

**Response:**
```json
{
  "message": "User deleted successfully",
  "user_id": 25,
  "username": "deleted_user",
  "email": "user@example.com",
  "was_admin": false
}
```

**⚠️ ห้าม**: ลบ GOD account (user_id = 1)

---

### 5. **POST /admin/admins** - Create New Admin

**Request Body:**
```json
{
  "username": "new_admin",
  "email": "admin@skillmatch.com",
  "password": "SecurePassword123!",
  "gender_id": 1,
  "admin_type": "user_manager",
  "tier_id": 2
}
```

**Response:**
```json
{
  "message": "Admin created successfully",
  "user_id": 30,
  "admin_type": "user_manager"
}
```

---

### 6. **GET /admin/admins** - List All Admins

**Response:**
```json
{
  "admins": [
    {
      "user_id": 1,
      "username": "The BOB Film",
      "email": "audikoratair@gmail.com",
      "is_admin": true,
      "tier_id": 5,
      "tier_name": "GOD",
      "admin_type": "god",
      "created_at": "2025-01-01T00:00:00Z"
    },
    {
      "user_id": 24,
      "username": "admin_john",
      "email": "john@admin.com",
      "is_admin": true,
      "tier_id": 2,
      "tier_name": "Silver",
      "admin_type": "admin",
      "created_at": "2025-02-10T14:20:00Z"
    }
  ],
  "total": 2
}
```

---

### 7. **DELETE /admin/admins/:user_id** - Delete Admin

**Example:**
```javascript
DELETE /admin/admins/24
```

**Response:**
```json
{
  "message": "Admin deleted successfully",
  "user_id": 24
}
```

**⚠️ ห้าม**: ลบ GOD account

---

### 8. **POST /god/view-mode** - Switch GOD View Mode (UI Only)

GOD สามารถเปลี่ยนโหมดการแสดงผล UI โดยไม่เปลี่ยน role จริง

**Request Body:**
```json
{
  "mode": "provider"
}
```

**Modes:**
- `"user"` - ดู UI แบบ user ธรรมดา
- `"provider"` - ดู UI แบบ provider
- `"admin"` - ดู UI แบบ admin
- `"god"` - ดู UI แบบ GOD (default)

**Response:**
```json
{
  "message": "View mode updated successfully",
  "current_mode": "provider",
  "note": "You are still GOD. This only affects UI display.",
  "actual_role": {
    "is_admin": true,
    "tier_id": 5
  }
}
```

---

### 9. **GET /god/view-mode** - Get Current View Mode

**Response:**
```json
{
  "current_mode": "provider",
  "actual_role": {
    "is_admin": true,
    "tier_id": 5
  },
  "available_modes": ["user", "provider", "admin", "god"]
}
```

---

## 🎨 Frontend Implementation Example

### React/Next.js Example:

```typescript
// api/god.ts
import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL;

// Get JWT token from localStorage/cookies
const getAuthHeader = () => ({
  headers: { Authorization: `Bearer ${localStorage.getItem('token')}` }
});

// 1. Get GOD Stats
export const getGodStats = async () => {
  const res = await axios.get(`${API_URL}/admin/stats/god`, getAuthHeader());
  return res.data;
};

// 2. List All Users with filters
export const listUsers = async (params: {
  page?: number;
  limit?: number;
  is_admin?: boolean;
  verification_status?: string;
  search?: string;
}) => {
  const res = await axios.get(`${API_URL}/admin/users`, {
    ...getAuthHeader(),
    params
  });
  return res.data;
};

// 3. Update User Role
export const updateUser = async (data: {
  user_id: number;
  is_admin?: boolean;
  tier_id?: number;
  provider_level_id?: number;
  verification_status?: string;
}) => {
  const res = await axios.post(`${API_URL}/god/update-user`, data, getAuthHeader());
  return res.data;
};

// 4. Delete User
export const deleteUser = async (userId: number) => {
  const res = await axios.delete(`${API_URL}/admin/users/${userId}`, getAuthHeader());
  return res.data;
};

// 5. Create Admin
export const createAdmin = async (data: {
  username: string;
  email: string;
  password: string;
  gender_id: number;
  admin_type?: string;
  tier_id?: number;
}) => {
  const res = await axios.post(`${API_URL}/admin/admins`, data, getAuthHeader());
  return res.data;
};

// 6. List Admins
export const listAdmins = async () => {
  const res = await axios.get(`${API_URL}/admin/admins`, getAuthHeader());
  return res.data;
};

// 7. Switch View Mode
export const setViewMode = async (mode: 'user' | 'provider' | 'admin' | 'god') => {
  const res = await axios.post(`${API_URL}/god/view-mode`, { mode }, getAuthHeader());
  return res.data;
};
```

---

### React Component Example:

```tsx
// components/GODDashboard.tsx
import { useState, useEffect } from 'react';
import { getGodStats, listUsers, deleteUser } from '@/api/god';

export default function GODDashboard() {
  const [stats, setStats] = useState(null);
  const [users, setUsers] = useState([]);
  const [page, setPage] = useState(1);

  useEffect(() => {
    // Fetch stats
    getGodStats().then(setStats);
    
    // Fetch users
    listUsers({ page, limit: 20 }).then(data => {
      setUsers(data.users);
    });
  }, [page]);

  const handleDelete = async (userId: number) => {
    if (confirm('Are you sure?')) {
      await deleteUser(userId);
      // Refresh list
      listUsers({ page, limit: 20 }).then(data => setUsers(data.users));
    }
  };

  return (
    <div>
      <h1>GOD Dashboard</h1>
      
      {/* Stats Cards */}
      <div className="grid grid-cols-4 gap-4">
        <div className="card">
          <h3>Total Users</h3>
          <p>{stats?.total_users}</p>
        </div>
        <div className="card">
          <h3>Providers</h3>
          <p>{stats?.total_providers}</p>
        </div>
        <div className="card">
          <h3>Revenue</h3>
          <p>฿{stats?.total_revenue}</p>
        </div>
        <div className="card">
          <h3>Active 24h</h3>
          <p>{stats?.active_users_24h}</p>
        </div>
      </div>

      {/* Users Table */}
      <table>
        <thead>
          <tr>
            <th>ID</th>
            <th>Username</th>
            <th>Email</th>
            <th>Role</th>
            <th>Status</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          {users.map(user => (
            <tr key={user.user_id}>
              <td>{user.user_id}</td>
              <td>{user.username}</td>
              <td>{user.email}</td>
              <td>{user.is_admin ? 'Admin' : 'User'}</td>
              <td>{user.verification_status}</td>
              <td>
                <button onClick={() => handleDelete(user.user_id)}>
                  Delete
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
```

---

## 🚨 Error Responses

### 403 Forbidden (Not GOD)
```json
{
  "error": "Only GOD can delete users"
}
```

### 404 Not Found
```json
{
  "error": "User not found"
}
```

### 400 Bad Request
```json
{
  "error": "Invalid user ID"
}
```

### 500 Internal Server Error
```json
{
  "error": "Failed to delete user",
  "details": "pq: duplicate key value..."
}
```

---

## 📚 Full API Reference

สำหรับ endpoints อื่นๆ ดูที่:
- `API_REFERENCE_FOR_FRONTEND.md` - Complete API docs
- `DATABASE_STRUCTURE.md` - Database schema
- `PROVIDER_SYSTEM_GUIDE.md` - Provider features
- `FINANCIAL_SYSTEM_GUIDE.md` - Payment & wallet

---

## ✅ Testing

```bash
# Test GOD endpoints
curl -X GET http://localhost:8080/admin/stats/god \
  -H "Authorization: Bearer <GOD_JWT_TOKEN>"

curl -X DELETE http://localhost:8080/admin/users/25 \
  -H "Authorization: Bearer <GOD_JWT_TOKEN>"
```

---

**Updated:** November 24, 2025  
**API Version:** 1.0.0
