# 🔐 Admin Role & Tier Management Guide

## 📋 Table of Contents
- [Role Hierarchy](#role-hierarchy)
- [Tier System](#tier-system)
- [Permission Matrix](#permission-matrix)
- [API Endpoints](#api-endpoints)
- [Implementation Details](#implementation-details)

---

## 🎭 Role Hierarchy

### 1. **GOD (Super Admin)** 👑
**Database Fields:**
- `is_admin = true`
- `tier_id = 5` (GOD tier)

**Capabilities:**
- ✅ **Full System Control**
- ✅ Create, modify, delete **any admin** (including other admins)
- ✅ Manage **all users** (regular users + providers)
- ✅ Manage **all providers**
- ✅ Access all admin endpoints
- ✅ Modify tier assignments
- ✅ Override any system restrictions

**Example Account:**
```
User ID: 1
Username: The BOB Film
Email: audikoratair@gmail.com
Password: admin123456
is_admin: true
tier_id: 5
```

---

### 2. **Admin (User Manager)** 👤
**Database Fields:**
- `is_admin = true`
- `tier_id = 1-4` (any non-GOD tier)
- **Admin Type:** User Manager

**Capabilities:**
- ✅ Manage **regular users only** (non-providers)
- ✅ View user profiles
- ✅ Update user verification status
- ✅ Handle user reports
- ❌ **Cannot** manage providers
- ❌ **Cannot** create/delete other admins
- ❌ **Cannot** modify GOD account
- ❌ **Cannot** access provider management endpoints

**Use Case:** Customer support, user moderation

---

### 3. **Admin (Provider Manager)** 🏢
**Database Fields:**
- `is_admin = true`
- `tier_id = 1-4` (any non-GOD tier)
- **Admin Type:** Provider Manager

**Capabilities:**
- ✅ Manage **providers only**
- ✅ Approve/reject provider applications
- ✅ Update provider verification status
- ✅ Manage service packages
- ✅ Handle provider-specific reports
- ❌ **Cannot** manage regular users
- ❌ **Cannot** create/delete other admins
- ❌ **Cannot** modify GOD account

**Use Case:** Business operations, provider quality control

---

### 4. **Regular User** 👥
**Database Fields:**
- `is_admin = false`
- `tier_id = 1-4` (subscription tier)

**Capabilities:**
- ✅ Browse providers
- ✅ Book services
- ✅ Send messages
- ✅ Write reviews
- ✅ Manage own profile
- ❌ No admin access

---

## 💎 Tier System

### Tier Structure
```sql
SELECT * FROM tiers ORDER BY tier_id;
```

| Tier ID | Name    | Access Level | Price (Monthly) | Description |
|---------|---------|--------------|-----------------|-------------|
| 1       | General | 0            | ฿0.00          | Free tier   |
| 2       | Silver  | 1            | ฿9.99          | Basic paid  |
| 3       | Diamond | 2            | ฿29.99         | Premium     |
| 4       | Premium | 3            | ฿99.99         | VIP         |
| 5       | GOD     | 999          | ฿9,999.99      | Super Admin |

### Tier Levels Explained

**For Users (`tier_id`):**
- Controls subscription benefits
- Access to premium features
- Browse filter capabilities
- Priority in search results

**For Providers (`provider_level_id`):**
- Determines visibility in browse results
- Access level filtering: Users can only see providers with `provider_level_id <= their_tier_access_level`
- Higher tiers appear first in search results

---

## 🔒 Permission Matrix

| Action | GOD | Admin (User) | Admin (Provider) | Regular User |
|--------|-----|--------------|------------------|--------------|
| **User Management** |
| View all users | ✅ | ✅ | ❌ | ❌ |
| Create/delete users | ✅ | ✅ | ❌ | ❌ |
| Update user profile | ✅ | ✅ | ❌ | Own only |
| Change user tier | ✅ | ❌ | ❌ | ❌ |
| Verify user identity | ✅ | ✅ | ❌ | ❌ |
| **Provider Management** |
| View all providers | ✅ | ❌ | ✅ | Limited |
| Approve providers | ✅ | ❌ | ✅ | ❌ |
| Update provider status | ✅ | ❌ | ✅ | ❌ |
| Manage service packages | ✅ | ❌ | ✅ | Own only |
| Change provider tier | ✅ | ❌ | ❌ | ❌ |
| **Admin Management** |
| Create admins | ✅ | ❌ | ❌ | ❌ |
| Delete admins | ✅ | ❌ | ❌ | ❌ |
| Modify admin roles | ✅ | ❌ | ❌ | ❌ |
| Change admin type | ✅ | ❌ | ❌ | ❌ |
| **System Management** |
| View all reports | ✅ | ✅ | ✅ | ❌ |
| Resolve reports | ✅ | ✅ | ✅ | ❌ |
| System settings | ✅ | ❌ | ❌ | ❌ |
| Database access | ✅ | ❌ | ❌ | ❌ |

---

## 🛠 API Endpoints

### 1. Admin Management (GOD Only)

#### Create Admin
```http
POST /admin/create-admin
Authorization: Bearer <GOD_TOKEN>
Content-Type: application/json

{
  "username": "admin_user_01",
  "email": "admin@example.com",
  "password": "securepass123",
  "admin_type": "user_manager",  // or "provider_manager"
  "tier_id": 2
}
```

**Response:**
```json
{
  "message": "Admin created successfully",
  "user_id": 26,
  "admin_type": "user_manager"
}
```

---

#### List All Admins
```http
GET /admin/list-admins
Authorization: Bearer <GOD_TOKEN>
```

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
      "user_id": 26,
      "username": "admin_user_01",
      "email": "admin@example.com",
      "is_admin": true,
      "tier_id": 2,
      "tier_name": "Silver",
      "admin_type": "user_manager",
      "created_at": "2025-11-13T10:30:00Z"
    }
  ],
  "total": 2
}
```

---

#### Update Admin Role
```http
PUT /admin/update-admin/:user_id
Authorization: Bearer <GOD_TOKEN>
Content-Type: application/json

{
  "admin_type": "provider_manager",
  "tier_id": 3
}
```

---

#### Delete Admin
```http
DELETE /admin/delete-admin/:user_id
Authorization: Bearer <GOD_TOKEN>
```

**Response:**
```json
{
  "message": "Admin deleted successfully",
  "user_id": 26
}
```

**Note:** Cannot delete GOD account (user_id = 1)

---

### 2. User Management (Admin User Manager)

#### List All Users
```http
GET /admin/users
Authorization: Bearer <ADMIN_TOKEN>
```

**Query Parameters:**
- `is_provider=false` - Only regular users (required for User Manager)
- `page=1`
- `limit=20`
- `verification_status=verified|approved|pending|rejected`

---

#### Update User Status
```http
PUT /admin/users/:user_id/status
Authorization: Bearer <ADMIN_TOKEN>
Content-Type: application/json

{
  "verification_status": "verified",
  "is_active": true
}
```

---

### 3. Provider Management (Admin Provider Manager)

#### List All Providers
```http
GET /admin/providers
Authorization: Bearer <ADMIN_PROVIDER_TOKEN>
```

**Query Parameters:**
- `is_provider=true` - Only providers (required for Provider Manager)
- `page=1`
- `limit=20`
- `verification_status=approved|pending|rejected`

---

#### Approve Provider
```http
PUT /admin/providers/:user_id/approve
Authorization: Bearer <ADMIN_PROVIDER_TOKEN>
Content-Type: application/json

{
  "verification_status": "approved",
  "provider_level_id": 2
}
```

---

#### Reject Provider
```http
PUT /admin/providers/:user_id/reject
Authorization: Bearer <ADMIN_PROVIDER_TOKEN>
Content-Type: application/json

{
  "verification_status": "rejected",
  "reason": "Incomplete documentation"
}
```

---

### 4. Tier Management (GOD Only)

#### Update User Tier
```http
PUT /admin/users/:user_id/tier
Authorization: Bearer <GOD_TOKEN>
Content-Type: application/json

{
  "tier_id": 3
}
```

---

#### Update Provider Tier
```http
PUT /admin/providers/:user_id/tier
Authorization: Bearer <GOD_TOKEN>
Content-Type: application/json

{
  "provider_level_id": 4
}
```

---

## 💻 Implementation Details

### Middleware Structure

#### 1. Authentication Middleware
```go
func authMiddleware(jwtSecret string, dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
    // Validates JWT token
    // Sets userID in context
    // Used for ALL protected routes
}
```

#### 2. Admin Middleware
```go
func adminAuthMiddleware(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
    // Checks if user has is_admin = true
    // Used for ALL admin routes
    // Requires authMiddleware to run first
}
```

#### 3. GOD Middleware (New - To Be Implemented)
```go
func godAuthMiddleware(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetInt("userID")
        
        var tierID int
        var isAdmin bool
        err := dbPool.QueryRow(ctx, 
            "SELECT tier_id, is_admin FROM users WHERE user_id = $1", 
            userID,
        ).Scan(&tierID, &isAdmin)
        
        if err != nil || !isAdmin || tierID != 5 {
            c.JSON(http.StatusForbidden, gin.H{
                "error": "GOD access required",
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

---

### Route Groups

```go
// main.go

// Public routes (no auth)
public := router.Group("/")
{
    public.POST("/register", registerHandler)
    public.POST("/login", loginHandler)
    public.GET("/ws", HandleWebSocket) // WebSocket uses message-based auth
}

// Protected routes (requires auth)
protected := router.Group("/")
protected.Use(authMiddleware(jwtSecret, dbPool, ctx))
{
    protected.GET("/profile", getProfileHandler)
    protected.GET("/browse/v2", browseUsersHandlerV2)
    // ... other user routes
}

// Admin routes (requires admin)
admin := router.Group("/admin")
admin.Use(authMiddleware(jwtSecret, dbPool, ctx))
admin.Use(adminAuthMiddleware(dbPool, ctx))
{
    // User Manager endpoints
    admin.GET("/users", listUsersHandler)
    admin.PUT("/users/:id/status", updateUserStatusHandler)
    
    // Provider Manager endpoints
    admin.GET("/providers", listProvidersHandler)
    admin.PUT("/providers/:id/approve", approveProviderHandler)
    
    // Reports (both admin types)
    admin.GET("/reports", listReportsHandler)
    admin.PUT("/reports/:id/resolve", resolveReportHandler)
}

// GOD routes (requires GOD tier)
god := router.Group("/admin/god")
god.Use(authMiddleware(jwtSecret, dbPool, ctx))
god.Use(godAuthMiddleware(dbPool, ctx))
{
    god.POST("/create-admin", createAdminHandler)
    god.GET("/list-admins", listAdminsHandler)
    god.PUT("/update-admin/:id", updateAdminHandler)
    god.DELETE("/delete-admin/:id", deleteAdminHandler)
    god.PUT("/users/:id/tier", updateUserTierHandler)
    god.PUT("/providers/:id/tier", updateProviderTierHandler)
}
```

---

### Database Schema

```sql
-- Users table structure
CREATE TABLE users (
    user_id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT,
    gender_id INTEGER NOT NULL REFERENCES genders(gender_id),
    
    -- Tier system
    tier_id INTEGER DEFAULT 1 REFERENCES tiers(tier_id),
    provider_level_id INTEGER DEFAULT 1 REFERENCES tiers(tier_id),
    
    -- Admin system
    is_admin BOOLEAN NOT NULL DEFAULT false,
    admin_type VARCHAR(50), -- 'god', 'user_manager', 'provider_manager'
    
    -- Status
    verification_status VARCHAR(20) NOT NULL DEFAULT 'unverified',
    is_active BOOLEAN DEFAULT true,
    
    -- Other fields
    registration_date TIMESTAMPTZ DEFAULT NOW(),
    phone_number VARCHAR(20),
    google_id TEXT UNIQUE,
    google_profile_picture TEXT
);

-- Tiers table
CREATE TABLE tiers (
    tier_id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    access_level INTEGER NOT NULL,
    price_monthly DECIMAL(10,2) NOT NULL
);
```

---

## 🔍 Permission Check Examples

### Check User Permission
```sql
-- Check if user is GOD
SELECT 
    user_id, 
    username, 
    is_admin, 
    tier_id,
    CASE 
        WHEN is_admin = true AND tier_id = 5 THEN 'GOD'
        WHEN is_admin = true AND admin_type = 'user_manager' THEN 'Admin (User)'
        WHEN is_admin = true AND admin_type = 'provider_manager' THEN 'Admin (Provider)'
        ELSE 'Regular User'
    END as role
FROM users 
WHERE user_id = $1;
```

### List Users by Admin Type
```sql
-- User Manager can only see regular users
SELECT u.* 
FROM users u
WHERE u.is_admin = false 
  AND (u.provider_level_id IS NULL OR u.provider_level_id = 1);

-- Provider Manager can only see providers
SELECT u.* 
FROM users u
WHERE u.provider_level_id > 1 
  OR EXISTS (
      SELECT 1 FROM service_packages sp 
      WHERE sp.provider_id = u.user_id
  );
```

---

## 🚀 Quick Start

### 1. Login as GOD
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "audikoratair@gmail.com",
    "password": "admin123456"
  }'
```

**Response:**
```json
{
  "token": "eyJhbGci...",
  "user": {
    "user_id": 1,
    "username": "The BOB Film",
    "is_admin": true,
    "tier_id": 5
  }
}
```

### 2. Create User Manager Admin
```bash
GOD_TOKEN="eyJhbGci..."

curl -X POST http://localhost:8080/admin/god/create-admin \
  -H "Authorization: Bearer $GOD_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "user_admin",
    "email": "useradmin@example.com",
    "password": "admin123",
    "admin_type": "user_manager",
    "tier_id": 2
  }'
```

### 3. Create Provider Manager Admin
```bash
curl -X POST http://localhost:8080/admin/god/create-admin \
  -H "Authorization: Bearer $GOD_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "provider_admin",
    "email": "provideradmin@example.com",
    "password": "admin123",
    "admin_type": "provider_manager",
    "tier_id": 2
  }'
```

---

## ⚠️ Security Considerations

### 1. **GOD Protection**
- Only ONE GOD account should exist (user_id = 1)
- GOD account cannot be deleted via API
- Require additional verification for GOD actions

### 2. **Admin Separation**
- User Manager admins cannot access provider endpoints
- Provider Manager admins cannot access user endpoints
- Enforce separation at middleware level

### 3. **Audit Logging**
- Log all admin actions (create, update, delete)
- Track who made changes and when
- Store in separate `admin_audit_log` table

### 4. **Rate Limiting**
- Implement rate limiting on admin endpoints
- Prevent brute force attacks
- Monitor unusual activity

---

## 📝 Future Enhancements

1. **Admin Permissions Table**
   - Granular permission system
   - Custom role creation
   - Permission inheritance

2. **Admin Dashboard**
   - Real-time statistics
   - User/Provider analytics
   - Report management UI

3. **Multi-Factor Authentication**
   - 2FA for admin accounts
   - SMS/Email verification
   - Authenticator app support

4. **Admin Activity Log**
   - Detailed action history
   - Search and filter capabilities
   - Export functionality

---

## 📞 Support

For questions about admin roles and permissions:
- Technical Documentation: See `API_REFERENCE_FOR_FRONTEND.md`
- Database Structure: See `DATABASE_STRUCTURE.md`
- Security Guide: See `SECURITY.md`

**GOD Account Access:**
- Email: audikoratair@gmail.com
- Password: admin123456
- User ID: 1
- Never share GOD credentials!

---

**Last Updated:** November 13, 2025
**Version:** 1.0.0
**Status:** ✅ Documentation Complete (Implementation Pending)
