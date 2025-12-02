# 🚨 Backend Requirements - สิ่งที่ Backend ต้องแก้ไขด่วน

> **วันที่**: 2 ธันวาคม 2025  
> **สถานะ Frontend**: ✅ พร้อมรับข้อมูลแล้ว  
> **รอ Backend**: แก้ไข Google OAuth และ Profile endpoints

---

## 📋 สารบัญ
1. [ปัญหาหลัก (Critical)](#-ปัญหาหลัก-critical-issues)
2. [ปัญหารอง (Secondary)](#-ปัญหารอง-secondary-issues)
3. [Checklist](#-checklist-สำหรับ-backend-team)
4. [Testing Guide](#-testing-guide)
5. [Code Examples](#-code-examples)

---

## 🔴 ปัญหาหลัก (Critical Issues)

### 1. ❌ Google OAuth ไม่บันทึก Profile Picture

**ปัญหา**:
- Frontend login ผ่าน Google OAuth สำเร็จ ✅
- Backend ส่ง JWT token กลับมา (200 OK) ✅
- แต่ไม่มี `profile_picture_url` ใน database ❌
- GET /profile/me ไม่ return รูปภาพ ❌
- Navbar แสดงแค่ตัวอักษรแรกแทนรูป

**Root Cause**:
Backend ไม่ได้ดึง `picture` field จาก Google User Info API และไม่ได้บันทึกลง database

**วิธีแก้ (Step-by-Step)**:

#### Step 1: ตรวจสอบ Database Schema
```bash
# เช็คว่ามี column หรือยัง
docker exec -i postgres_db psql -U admin -d skillmatch_db -c "
SELECT column_name, data_type, character_maximum_length
FROM information_schema.columns 
WHERE table_name = 'users' AND column_name = 'profile_picture_url';
"
```

**Expected Output**:
```
     column_name      | data_type | character_maximum_length 
---------------------+-----------+--------------------------
 profile_picture_url | text      |
```

**ถ้าไม่มี column ให้เพิ่ม**:
```sql
ALTER TABLE users ADD COLUMN IF NOT EXISTS profile_picture_url TEXT;
```

#### Step 2: แก้ไข Google OAuth Handler

**ไฟล์**: `auth_handlers.go`

**ที่ต้องแก้**:
```go
// Line ~250-300 ใน handleGoogleCallback function

func handleGoogleCallback(c *gin.Context) {
    var req struct {
        Code string `json:"code" binding:"required"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization code required"})
        return
    }
    
    // 1. Exchange code for token
    ctx := context.Background()
    token, err := googleOauthConfig.Exchange(ctx, req.Code)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "error": "Invalid authorization code",
            "details": err.Error(),
        })
        return
    }
    
    // 2. Get user info from Google API
    client := googleOauthConfig.Client(ctx, token)
    resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to fetch user info from Google",
        })
        return
    }
    defer resp.Body.Close()
    
    // 3. Parse Google user data
    var googleUser struct {
        ID            string `json:"id"`
        Email         string `json:"email"`
        VerifiedEmail bool   `json:"verified_email"`
        Name          string `json:"name"`
        GivenName     string `json:"given_name"`
        FamilyName    string `json:"family_name"`
        Picture       string `json:"picture"` // ⬅️ นี่คือที่สำคัญ!
        Locale        string `json:"locale"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to parse Google user data",
        })
        return
    }
    
    // 4. Find or create user in database
    var user User
    result := db.Where("email = ?", googleUser.Email).First(&user)
    
    if result.Error == gorm.ErrRecordNotFound {
        // Create new user
        user = User{
            Email:                googleUser.Email,
            Username:             googleUser.Name,
            ProfilePictureURL:    &googleUser.Picture, // ⬅️ บันทึกรูปภาพ
            IsEmailVerified:      true,
            TierID:               1, // Default General tier
            VerificationStatus:   "unverified",
            CreatedAt:            time.Now(),
            UpdatedAt:            time.Now(),
        }
        
        if err := db.Create(&user).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": "Failed to create user",
            })
            return
        }
    } else if result.Error == nil {
        // User exists - update profile picture
        if googleUser.Picture != "" {
            user.ProfilePictureURL = &googleUser.Picture // ⬅️ อัพเดทรูปภาพ
            user.UpdatedAt = time.Now()
            
            if err := db.Save(&user).Error; err != nil {
                log.Printf("Warning: Failed to update profile picture: %v", err)
            }
        }
    } else {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Database error",
        })
        return
    }
    
    // 5. Generate JWT token
    jwtToken, err := createJWT(user.UserID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to generate authentication token",
        })
        return
    }
    
    // 6. Fetch complete user data with tier name
    var userResponse struct {
        UserID             int     `json:"user_id"`
        Username           string  `json:"username"`
        Email              string  `json:"email"`
        TierID             int     `json:"tier_id"`
        TierName           string  `json:"tier_name"`
        IsAdmin            bool    `json:"is_admin"`
        ProfilePictureURL  *string `json:"profile_picture_url"` // ⬅️ ส่งกลับไปด้วย
        VerificationStatus string  `json:"verification_status"`
    }
    
    err = db.Raw(`
        SELECT 
            u.user_id,
            u.username,
            u.email,
            u.tier_id,
            COALESCE(t.name, 'General') as tier_name,
            u.is_admin,
            u.profile_picture_url,
            u.verification_status
        FROM users u
        LEFT JOIN tiers t ON u.tier_id = t.tier_id
        WHERE u.user_id = ?
    `, user.UserID).Scan(&userResponse).Error
    
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to fetch user data",
        })
        return
    }
    
    // 7. Return token and user data
    c.JSON(http.StatusOK, gin.H{
        "message": "Login successful",
        "token":   jwtToken,
        "user":    userResponse, // ⬅️ ส่ง user object กลับไปด้วย
    })
}
```

**สิ่งที่เปลี่ยน**:
1. ✅ เพิ่ม `Picture` field ใน struct ของ Google user data
2. ✅ บันทึก `profile_picture_url` เวลาสร้าง user ใหม่
3. ✅ อัพเดท `profile_picture_url` เวลา user login ซ้ำ
4. ✅ ส่ง `profile_picture_url` กลับใน response

---

### 2. ✅ GET /profile/me - Must Return Profile Picture

**Endpoint**: `GET /profile/me` (alias: `/users/me`)

**Current Status**: ✅ Endpoint มีอยู่แล้ว

**ที่ต้องตรวจสอบ**: Response ต้องมี `profile_picture_url`

**ไฟล์**: `user_handlers.go`

```go
// Line ~20-80 ใน getMeHandler function

func getMeHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID, exists := c.Get("userID")
        if !exists {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
            return
        }
        
        var user struct {
            UserID             int     `json:"user_id"`
            Username           string  `json:"username"`
            Email              string  `json:"email"`
            TierID             int     `json:"tier_id"`
            TierName           string  `json:"tier_name"`
            IsAdmin            bool    `json:"is_admin"`
            ProfilePictureURL  *string `json:"profile_picture_url"` // ⬅️ ต้องมี!
            Bio                *string `json:"bio"`
            Phone              *string `json:"phone"`
            VerificationStatus string  `json:"verification_status"`
        }
        
        // Query with LEFT JOIN to get tier name
        sqlStatement := `
            SELECT 
                u.user_id,
                u.username,
                u.email,
                u.tier_id,
                COALESCE(t.name, 'General') as tier_name,
                u.is_admin,
                u.profile_picture_url,  -- ⬅️ เพิ่มบรรทัดนี้
                u.bio,
                u.phone,
                u.verification_status
            FROM users u
            LEFT JOIN tiers t ON u.tier_id = t.tier_id
            WHERE u.user_id = $1
        `
        
        err := dbPool.QueryRow(ctx, sqlStatement, userID).Scan(
            &user.UserID,
            &user.Username,
            &user.Email,
            &user.TierID,
            &user.TierName,
            &user.IsAdmin,
            &user.ProfilePictureURL, // ⬅️ เพิ่มบรรทัดนี้
            &user.Bio,
            &user.Phone,
            &user.VerificationStatus,
        )
        
        if err != nil {
            c.JSON(http.StatusNotFound, gin.H{
                "error": "User not found",
                "details": err.Error(),
            })
            return
        }
        
        c.JSON(http.StatusOK, user)
    }
}
```

**Expected Response**:
```json
{
  "user_id": 1,
  "username": "The BOB Film",
  "email": "audikoratair@gmail.com",
  "tier_id": 5,
  "tier_name": "GOD",
  "is_admin": true,
  "profile_picture_url": "https://lh3.googleusercontent.com/a/ACg8ocK...",
  "bio": null,
  "phone": null,
  "verification_status": "unverified"
}
```

---

## ⚠️ ปัญหารอง (Secondary Issues)

### 3. Browse Filters - ต้องรองรับ Query Parameters ทั้งหมด

**Endpoint**: `GET /browse/search` หรือ `GET /categories/:category_id/providers`

**Query Parameters ที่ Frontend ส่งมา**:
```
?location=Bangkok
&rating=4
&tier=3
&category=1
&service_type=Both
&sort=rating
&page=1
&limit=20
```

**ไฟล์**: `browse_handlers_v2.go` (ถ้ามี) หรือสร้างใหม่

```go
func browseProvidersHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Parse query parameters
        location := c.Query("location")           // Province/District
        ratingStr := c.Query("rating")            // Min rating: "3", "4", "4.5"
        tierStr := c.Query("tier")                // Provider level: "1"-"4"
        categoryStr := c.Query("category")        // Category ID
        serviceType := c.Query("service_type")    // "Incall", "Outcall", "Both"
        sortBy := c.DefaultQuery("sort", "rating") // "rating", "reviews", "price"
        
        // Pagination
        pageStr := c.DefaultQuery("page", "1")
        limitStr := c.DefaultQuery("limit", "20")
        
        page, _ := strconv.Atoi(pageStr)
        limit, _ := strconv.Atoi(limitStr)
        offset := (page - 1) * limit
        
        // Build base query
        query := `
            SELECT DISTINCT
                u.user_id,
                u.username,
                u.profile_image_url,
                u.bio,
                u.provider_level_id,
                pl.name as provider_level_name,
                u.rating_avg,
                u.review_count,
                u.service_type,
                u.province,
                u.district
            FROM users u
            LEFT JOIN tiers pl ON u.provider_level_id = pl.tier_id
            WHERE u.verification_status IN ('approved', 'verified')
        `
        
        args := []interface{}{}
        argPos := 1
        
        // Apply filters
        if location != "" {
            query += fmt.Sprintf(" AND (u.province ILIKE $%d OR u.district ILIKE $%d)", argPos, argPos+1)
            args = append(args, "%"+location+"%", "%"+location+"%")
            argPos += 2
        }
        
        if ratingStr != "" {
            minRating, _ := strconv.ParseFloat(ratingStr, 64)
            query += fmt.Sprintf(" AND u.rating_avg >= $%d", argPos)
            args = append(args, minRating)
            argPos++
        }
        
        if tierStr != "" {
            tierID, _ := strconv.Atoi(tierStr)
            query += fmt.Sprintf(" AND u.provider_level_id = $%d", argPos)
            args = append(args, tierID)
            argPos++
        }
        
        if categoryStr != "" {
            categoryID, _ := strconv.Atoi(categoryStr)
            query += fmt.Sprintf(` AND EXISTS (
                SELECT 1 FROM provider_categories pc 
                WHERE pc.provider_id = u.user_id 
                AND pc.category_id = $%d
            )`, argPos)
            args = append(args, categoryID)
            argPos++
        }
        
        if serviceType != "" && serviceType != "All" {
            query += fmt.Sprintf(" AND (u.service_type = $%d OR u.service_type = 'Both')", argPos)
            args = append(args, serviceType)
            argPos++
        }
        
        // Apply sorting
        switch sortBy {
        case "reviews":
            query += " ORDER BY u.review_count DESC, u.rating_avg DESC"
        case "price":
            query += " ORDER BY u.user_id" // TODO: Join with packages table
        default: // rating
            query += " ORDER BY u.rating_avg DESC, u.review_count DESC"
        }
        
        // Add pagination
        query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argPos, argPos+1)
        args = append(args, limit, offset)
        
        // Execute query
        rows, err := dbPool.Query(ctx, query, args...)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": "Failed to fetch providers",
                "details": err.Error(),
            })
            return
        }
        defer rows.Close()
        
        providers := []map[string]interface{}{}
        for rows.Next() {
            var p struct {
                UserID            int
                Username          string
                ProfileImageURL   *string
                Bio               *string
                ProviderLevelID   int
                ProviderLevelName string
                RatingAvg         float64
                ReviewCount       int
                ServiceType       string
                Province          *string
                District          *string
            }
            
            err := rows.Scan(
                &p.UserID, &p.Username, &p.ProfileImageURL, &p.Bio,
                &p.ProviderLevelID, &p.ProviderLevelName,
                &p.RatingAvg, &p.ReviewCount, &p.ServiceType,
                &p.Province, &p.District,
            )
            
            if err != nil {
                continue
            }
            
            providers = append(providers, map[string]interface{}{
                "user_id":              p.UserID,
                "username":             p.Username,
                "profile_image_url":    p.ProfileImageURL,
                "bio":                  p.Bio,
                "provider_level_id":    p.ProviderLevelID,
                "provider_level_name":  p.ProviderLevelName,
                "rating_avg":           p.RatingAvg,
                "review_count":         p.ReviewCount,
                "service_type":         p.ServiceType,
                "province":             p.Province,
                "district":             p.District,
            })
        }
        
        // Get total count (without pagination)
        countQuery := `SELECT COUNT(DISTINCT u.user_id) FROM users u WHERE u.verification_status IN ('approved', 'verified')`
        // TODO: Add same filters to count query
        
        var total int
        err = dbPool.QueryRow(ctx, countQuery).Scan(&total)
        if err != nil {
            total = len(providers)
        }
        
        c.JSON(http.StatusOK, gin.H{
            "providers": providers,
            "pagination": gin.H{
                "page":  page,
                "limit": limit,
                "total": total,
            },
            "filters_applied": gin.H{
                "location":     location,
                "rating":       ratingStr,
                "tier":         tierStr,
                "category":     categoryStr,
                "service_type": serviceType,
                "sort":         sortBy,
            },
        })
    }
}
```

**Register Route in main.go**:
```go
// Public routes
public.GET("/browse/search", browseProvidersHandler(dbPool, ctx))
```

---

### 4. Service Categories - ต้องมี Thai Names และ Icons

**Endpoint**: `GET /service-categories`

**Current Status**: ✅ Endpoint มีแล้ว แต่ขาด Thai names

**ที่ต้องแก้**:

#### Database Migration (ทำแล้ว ✅)
```bash
# Check current data
docker exec -i postgres_db psql -U admin -d skillmatch_db -c "SELECT * FROM service_categories;"
```

**Expected Output**:
```
 category_id |   name   | name_thai | icon |      description       
-------------+----------+-----------+------+------------------------
           1 | Massage  | นวด       | 💆   | Professional massage...
           2 | Spa      | สปา       | 🧖   | Spa and wellness...
```

**ถ้ายังไม่มี Thai names ให้รัน**:
```sql
UPDATE service_categories SET name_thai = 'นวด', icon = '💆' WHERE name = 'Massage';
UPDATE service_categories SET name_thai = 'สปา', icon = '🧖' WHERE name = 'Spa';
UPDATE service_categories SET name_thai = 'ความงาม', icon = '💄' WHERE name = 'Beauty';
UPDATE service_categories SET name_thai = 'สุขภาพ', icon = '🧘' WHERE name = 'Wellness';
UPDATE service_categories SET name_thai = 'บำบัด', icon = '🩺' WHERE name = 'Therapy';
```

#### Handler Check
**ไฟล์**: `category_handlers.go`

```go
func listServiceCategoriesHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
    return func(c *gin.Context) {
        includeAdult := c.DefaultQuery("include_adult", "true") == "true"

        sqlStatement := `
            SELECT 
                category_id, 
                name, 
                name_thai,    -- ⬅️ ต้องมี
                description, 
                icon,         -- ⬅️ ต้องมี
                is_adult, 
                display_order, 
                is_active
            FROM service_categories
            WHERE is_active = true
        `

        if !includeAdult {
            sqlStatement += " AND is_adult = false"
        }

        sqlStatement += " ORDER BY display_order ASC"

        rows, err := dbPool.Query(ctx, sqlStatement)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error":   "Failed to fetch categories",
                "details": err.Error(),
            })
            return
        }
        defer rows.Close()

        categories := []ServiceCategory{}
        for rows.Next() {
            var cat ServiceCategory
            err := rows.Scan(
                &cat.CategoryID, 
                &cat.Name, 
                &cat.NameThai,    // ⬅️ ต้อง scan
                &cat.Description,
                &cat.Icon,        // ⬅️ ต้อง scan
                &cat.IsAdult, 
                &cat.DisplayOrder, 
                &cat.IsActive,
            )
            if err != nil {
                continue
            }
            categories = append(categories, cat)
        }

        c.JSON(http.StatusOK, gin.H{
            "categories": categories,
            "total":      len(categories),
        })
    }
}
```

**ServiceCategory struct** ใน `models.go`:
```go
type ServiceCategory struct {
    CategoryID   int     `json:"category_id"`
    Name         string  `json:"name"`
    NameThai     string  `json:"name_thai"`    // ⬅️ ต้องมี
    Description  *string `json:"description"`
    Icon         *string `json:"icon"`         // ⬅️ ต้องมี
    IsAdult      bool    `json:"is_adult"`
    DisplayOrder int     `json:"display_order"`
    IsActive     bool    `json:"is_active"`
}
```

---

## 📝 Checklist สำหรับ Backend Team

### 🔴 Priority 1 (ทำก่อน - Critical)

#### Google OAuth Profile Picture
- [ ] **Database Schema**
  - [ ] ตรวจสอบว่ามี `profile_picture_url TEXT` column ใน `users` table
  - [ ] ถ้าไม่มีให้ `ALTER TABLE users ADD COLUMN profile_picture_url TEXT;`

- [ ] **auth_handlers.go**
  - [ ] เพิ่ม `Picture string` ใน Google user struct
  - [ ] บันทึก `profile_picture_url` เวลาสร้าง user ใหม่ (line ~280)
  - [ ] อัพเดท `profile_picture_url` เวลา user login ซ้ำ (line ~295)
  - [ ] ส่ง `profile_picture_url` กลับใน response (line ~320)

- [ ] **Testing**
  - [ ] Login ด้วย Google ผ่าน Frontend
  - [ ] เช็ค database: `SELECT email, profile_picture_url FROM users LIMIT 1;`
  - [ ] ต้องเห็น URL จาก `lh3.googleusercontent.com`

#### Profile Endpoint
- [ ] **user_handlers.go**
  - [ ] เพิ่ม `ProfilePictureURL *string` ใน response struct
  - [ ] เพิ่ม `u.profile_picture_url` ใน SELECT query
  - [ ] เพิ่ม `&user.ProfilePictureURL` ใน Scan()

- [ ] **Testing**
  - [ ] `curl -H "Authorization: Bearer TOKEN" http://localhost:8080/profile/me`
  - [ ] Response ต้องมี `"profile_picture_url": "https://..."`

### 🟡 Priority 2 (Important)

#### Browse Search Filters
- [ ] **browse_handlers_v2.go** (สร้างใหม่)
  - [ ] สร้าง `browseProvidersHandler` function
  - [ ] รองรับ `location` query param (ILIKE search)
  - [ ] รองรับ `rating` query param (>= filter)
  - [ ] รองรับ `tier` query param (provider_level_id)
  - [ ] รองรับ `category` query param (JOIN provider_categories)
  - [ ] รองรับ `service_type` query param
  - [ ] รองรับ `sort` param (rating/reviews/price)
  - [ ] Pagination (page, limit, offset)

- [ ] **main.go**
  - [ ] Register route: `public.GET("/browse/search", browseProvidersHandler(dbPool, ctx))`

- [ ] **Testing**
  - [ ] Test แต่ละ filter แยก
  - [ ] Test combined filters
  - [ ] Test pagination
  - [ ] Test sorting

#### Service Categories Thai Names
- [ ] **Database**
  - [ ] Run UPDATE statements สำหรับ Thai names
  - [ ] Verify: `SELECT name, name_thai, icon FROM service_categories;`

- [ ] **category_handlers.go**
  - [ ] เพิ่ม `NameThai` และ `Icon` ใน SELECT
  - [ ] เพิ่ม scan ใน loop

- [ ] **models.go**
  - [ ] เพิ่ม fields ใน `ServiceCategory` struct

### 🟢 Priority 3 (Nice to have)

- [ ] Provider Photos Endpoint
  - [ ] `GET /provider/:userId/photos`
  - [ ] Sort by `sort_order ASC`
  - [ ] Include `caption` and `uploaded_at`

- [ ] Favorites Check for Guests
  - [ ] `GET /favorites/check/:providerId`
  - [ ] Return `false` ถ้าไม่มี token

- [ ] Notifications Unread Count
  - [ ] `GET /notifications/unread/count`
  - [ ] Return `{ "unread_count": 5 }`

---

## 🧪 Testing Guide

### Test 1: Google OAuth Profile Picture

```bash
# 1. Login ผ่าน Frontend
# เปิด http://localhost:3000 (หรือ 5173)
# คลิก "Sign in with Google"
# Login สำเร็จ

# 2. เช็ค Database
docker exec -i postgres_db psql -U admin -d skillmatch_db -c "
SELECT 
    user_id, 
    email, 
    username, 
    profile_picture_url,
    LENGTH(profile_picture_url) as url_length
FROM users 
WHERE email = 'audikoratair@gmail.com';
"

# Expected Output:
# user_id | email | username | profile_picture_url | url_length
# --------|-------|----------|---------------------|------------
# 1 | audikoratair@gmail.com | The BOB Film | https://lh3.googleusercontent.com/a/ACg8ocK... | 120+

# 3. Test API Endpoint
curl -s -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  http://localhost:8080/profile/me | jq .

# Expected Response:
{
  "user_id": 1,
  "username": "The BOB Film",
  "email": "audikoratair@gmail.com",
  "tier_id": 5,
  "tier_name": "GOD",
  "is_admin": true,
  "profile_picture_url": "https://lh3.googleusercontent.com/a/ACg8ocK...",
  "bio": null,
  "phone": null,
  "verification_status": "unverified"
}
```

### Test 2: Browse Filters

```bash
# Test Location Filter
curl -s "http://localhost:8080/browse/search?location=Bangkok" | jq '.providers | length'

# Test Rating Filter
curl -s "http://localhost:8080/browse/search?rating=4" | jq '.providers[] | {username, rating_avg}'

# Test Tier Filter
curl -s "http://localhost:8080/browse/search?tier=3" | jq '.providers[] | {username, provider_level_name}'

# Test Category Filter
curl -s "http://localhost:8080/browse/search?category=1" | jq '.providers | length'

# Test Combined Filters
curl -s "http://localhost:8080/browse/search?location=Bangkok&rating=4&tier=3&category=1&sort=rating&page=1&limit=10" | jq .

# Expected Response Structure:
{
  "providers": [...],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 50
  },
  "filters_applied": {
    "location": "Bangkok",
    "rating": "4",
    "tier": "3",
    "category": "1",
    "service_type": "",
    "sort": "rating"
  }
}
```

### Test 3: Categories Thai Names

```bash
# Test Categories Endpoint
curl -s http://localhost:8080/service-categories | jq .

# Expected Response:
{
  "categories": [
    {
      "category_id": 1,
      "name": "Massage",
      "name_thai": "นวด",
      "icon": "💆",
      "description": "Professional massage services",
      "is_adult": false,
      "display_order": 1,
      "is_active": true
    },
    {
      "category_id": 2,
      "name": "Spa",
      "name_thai": "สปา",
      "icon": "🧖",
      "description": "Spa and wellness treatments",
      "is_adult": false,
      "display_order": 2,
      "is_active": true
    }
  ],
  "total": 5
}
```

---

## 💡 Code Examples

### Complete Google OAuth Handler

**ไฟล์**: `auth_handlers.go`

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
)

type GoogleUserInfo struct {
    ID            string `json:"id"`
    Email         string `json:"email"`
    VerifiedEmail bool   `json:"verified_email"`
    Name          string `json:"name"`
    GivenName     string `json:"given_name"`
    FamilyName    string `json:"family_name"`
    Picture       string `json:"picture"`
    Locale        string `json:"locale"`
}

func handleGoogleCallback(c *gin.Context) {
    var req struct {
        Code string `json:"code" binding:"required"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Authorization code is required",
        })
        return
    }
    
    // 1. Exchange authorization code for access token
    ctx := context.Background()
    token, err := googleOauthConfig.Exchange(ctx, req.Code)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "error": "Invalid authorization code",
            "details": err.Error(),
        })
        return
    }
    
    // 2. Get user info from Google API
    client := googleOauthConfig.Client(ctx, token)
    resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to fetch user information from Google",
        })
        return
    }
    defer resp.Body.Close()
    
    // 3. Parse Google user data
    var googleUser GoogleUserInfo
    if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to parse Google user data",
        })
        return
    }
    
    // 4. Find or create user in database
    var user User
    err = db.QueryRow(`
        SELECT user_id, email, username, tier_id, is_admin, verification_status
        FROM users WHERE email = $1
    `, googleUser.Email).Scan(
        &user.UserID, &user.Email, &user.Username,
        &user.TierID, &user.IsAdmin, &user.VerificationStatus,
    )
    
    if err == sql.ErrNoRows {
        // Create new user
        err = db.QueryRow(`
            INSERT INTO users (
                email, username, profile_picture_url, 
                tier_id, is_admin, verification_status, created_at, updated_at
            ) VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
            RETURNING user_id
        `, googleUser.Email, googleUser.Name, googleUser.Picture,
           1, false, "unverified").Scan(&user.UserID)
        
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": "Failed to create user account",
            })
            return
        }
        
        user.Email = googleUser.Email
        user.Username = googleUser.Name
        user.TierID = 1
        user.IsAdmin = false
        user.VerificationStatus = "unverified"
        
    } else if err == nil {
        // User exists - update profile picture
        _, err = db.Exec(`
            UPDATE users 
            SET profile_picture_url = $1, updated_at = NOW()
            WHERE user_id = $2
        `, googleUser.Picture, user.UserID)
        
        if err != nil {
            fmt.Printf("Warning: Failed to update profile picture: %v\n", err)
        }
    } else {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Database error",
        })
        return
    }
    
    // 5. Generate JWT token
    jwtToken, err := createJWT(user.UserID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to generate authentication token",
        })
        return
    }
    
    // 6. Fetch complete user data with tier and profile picture
    var userResponse struct {
        UserID             int     `json:"user_id"`
        Username           string  `json:"username"`
        Email              string  `json:"email"`
        TierID             int     `json:"tier_id"`
        TierName           string  `json:"tier_name"`
        IsAdmin            bool    `json:"is_admin"`
        ProfilePictureURL  *string `json:"profile_picture_url"`
        VerificationStatus string  `json:"verification_status"`
    }
    
    err = db.QueryRow(`
        SELECT 
            u.user_id,
            u.username,
            u.email,
            u.tier_id,
            COALESCE(t.name, 'General') as tier_name,
            u.is_admin,
            u.profile_picture_url,
            u.verification_status
        FROM users u
        LEFT JOIN tiers t ON u.tier_id = t.tier_id
        WHERE u.user_id = $1
    `, user.UserID).Scan(
        &userResponse.UserID,
        &userResponse.Username,
        &userResponse.Email,
        &userResponse.TierID,
        &userResponse.TierName,
        &userResponse.IsAdmin,
        &userResponse.ProfilePictureURL,
        &userResponse.VerificationStatus,
    )
    
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to retrieve user data",
        })
        return
    }
    
    // 7. Return token and complete user object
    c.JSON(http.StatusOK, gin.H{
        "message": "Login successful",
        "token":   jwtToken,
        "user":    userResponse,
    })
}
```

---

## 📊 Expected Data Flow

### Flow 1: Google OAuth Login with Profile Picture

```
┌─────────┐                 ┌─────────┐                ┌────────┐               ┌──────────┐
│ Frontend│                 │ Google  │                │ Backend│               │ Database │
└────┬────┘                 └────┬────┘                └───┬────┘               └─────┬────┘
     │                           │                         │                          │
     │ 1. User clicks            │                         │                          │
     │   "Sign in with Google"   │                         │                          │
     ├──────────────────────────>│                         │                          │
     │                           │                         │                          │
     │ 2. Returns auth code      │                         │                          │
     │<──────────────────────────┤                         │                          │
     │                           │                         │                          │
     │ 3. POST /auth/google      │                         │                          │
     │   { code: "..." }         │                         │                          │
     ├────────────────────────────────────────────────────>│                          │
     │                           │                         │                          │
     │                           │ 4. Exchange code        │                          │
     │                           │    for access token     │                          │
     │                           │<────────────────────────┤                          │
     │                           │                         │                          │
     │                           │ 5. Get user info        │                          │
     │                           │    (including picture)  │                          │
     │                           │<────────────────────────┤                          │
     │                           │ { email, name,          │                          │
     │                           │   picture: "https://..." }                        │
     │                           │                         │                          │
     │                           │                         │ 6. Find user by email    │
     │                           │                         ├─────────────────────────>│
     │                           │                         │                          │
     │                           │                         │ 7. If not exists, INSERT │
     │                           │                         │    with profile_picture_url
     │                           │                         ├─────────────────────────>│
     │                           │                         │                          │
     │                           │                         │ 8. If exists, UPDATE     │
     │                           │                         │    profile_picture_url   │
     │                           │                         ├─────────────────────────>│
     │                           │                         │                          │
     │                           │                         │ 9. Return user data      │
     │                           │                         │<─────────────────────────┤
     │                           │                         │                          │
     │ 10. Return JWT + user     │                         │                          │
     │     (with profile_picture_url)                      │                          │
     │<────────────────────────────────────────────────────┤                          │
     │ { token: "...",           │                         │                          │
     │   user: {                 │                         │                          │
     │     profile_picture_url:  │                         │                          │
     │     "https://lh3..."      │                         │                          │
     │   }                       │                         │                          │
     │ }                         │                         │                          │
     │                           │                         │                          │
     │ 11. Save token & display  │                         │                          │
     │     profile picture       │                         │                          │
     │                           │                         │                          │
```

### Flow 2: Browse with Filters

```
┌─────────┐                                    ┌────────┐               ┌──────────┐
│ Frontend│                                    │ Backend│               │ Database │
└────┬────┘                                    └───┬────┘               └─────┬────┘
     │                                             │                          │
     │ User selects filters:                      │                          │
     │ - Location: Bangkok                        │                          │
     │ - Rating: 4+                               │                          │
     │ - Tier: Diamond (3)                        │                          │
     │ - Category: Massage (1)                    │                          │
     │ - Sort: Rating                             │                          │
     │                                             │                          │
     │ GET /browse/search?                        │                          │
     │   location=Bangkok&                        │                          │
     │   rating=4&                                │                          │
     │   tier=3&                                  │                          │
     │   category=1&                              │                          │
     │   sort=rating&                             │                          │
     │   page=1&limit=20                          │                          │
     ├────────────────────────────────────────────>│                          │
     │                                             │                          │
     │                                             │ Build SQL with filters:  │
     │                                             │ WHERE province ILIKE '%Bangkok%'
     │                                             │   AND rating_avg >= 4    │
     │                                             │   AND provider_level_id = 3
     │                                             │   AND EXISTS (SELECT ... │
     │                                             │     FROM provider_categories
     │                                             │     WHERE category_id = 1)
     │                                             │ ORDER BY rating_avg DESC │
     │                                             │ LIMIT 20 OFFSET 0        │
     │                                             ├─────────────────────────>│
     │                                             │                          │
     │                                             │ Return filtered providers│
     │                                             │<─────────────────────────┤
     │                                             │                          │
     │ Response:                                   │                          │
     │ {                                           │                          │
     │   "providers": [                            │                          │
     │     {                                       │                          │
     │       "username": "Diamond Spa",            │                          │
     │       "provider_level_id": 3,               │                          │
     │       "provider_level_name": "Diamond",     │                          │
     │       "rating_avg": 4.8,                    │                          │
     │       "province": "Bangkok"                 │                          │
     │     }                                       │                          │
     │   ],                                        │                          │
     │   "pagination": { ... },                    │                          │
     │   "filters_applied": { ... }                │                          │
     │ }                                           │                          │
     │<────────────────────────────────────────────┤                          │
     │                                             │                          │
     │ Display filtered providers                  │                          │
     │                                             │                          │
```

---

## 🚀 Summary & Timeline

### What Frontend Has Done ✅
- Google OAuth integration (frontend complete)
- Profile picture UI (navbar avatar)
- Browse filters UI (all filter components ready)
- API service layer (ready to call endpoints)

### What Backend Needs to Do ⏳

| Task | Priority | Estimated Time | Status |
|------|----------|---------------|--------|
| Google OAuth save profile picture | 🔴 Critical | 30 min | ⏳ TODO |
| GET /profile/me return picture | 🔴 Critical | 15 min | ⏳ TODO |
| Browse search filters | 🟡 High | 1-2 hours | ⏳ TODO |
| Categories Thai names | 🟡 High | 30 min | ✅ DONE |
| Provider photos endpoint | 🟢 Medium | 30 min | ⏳ TODO |
| Favorites check endpoint | 🟢 Low | 15 min | ⏳ TODO |

**Total Estimated Time**: ~3-4 hours

---

## 📞 Support & Questions

### Quick Debug Commands

```bash
# 1. Check if profile_picture_url column exists
docker exec -i postgres_db psql -U admin -d skillmatch_db -c "\d users"

# 2. Check current data
docker exec -i postgres_db psql -U admin -d skillmatch_db -c "SELECT user_id, email, profile_picture_url FROM users LIMIT 5;"

# 3. Check service categories
docker exec -i postgres_db psql -U admin -d skillmatch_db -c "SELECT * FROM service_categories;"

# 4. Test API endpoints
curl http://localhost:8080/ping
curl http://localhost:8080/service-categories
curl -H "Authorization: Bearer TOKEN" http://localhost:8080/profile/me
```

### Common Errors & Solutions

**Error: `column "profile_picture_url" does not exist`**
```sql
ALTER TABLE users ADD COLUMN profile_picture_url TEXT;
```

**Error: `column "name_thai" does not exist`**
```sql
-- Already fixed in migration 032 ✅
```

**Error: `invalid authorization code`**
- Check Google OAuth credentials in `.env`
- Verify redirect URI in Google Console
- Make sure frontend sends correct `code` (not `credential`)

---

## ✅ Verification Checklist

ก่อนส่งให้ Frontend ให้ตรวจสอบ:

- [ ] **Google OAuth**: Login แล้ว database มี profile_picture_url
- [ ] **GET /profile/me**: Response มี profile_picture_url field
- [ ] **Browse filters**: ทุก query parameter ทำงานถูกต้อง
- [ ] **Categories**: มี name_thai และ icon
- [ ] **Test กับ Frontend**: ทดสอบ end-to-end workflow

---

**Frontend พร้อมแล้ว! รอ Backend update API endpoints 🚀**
