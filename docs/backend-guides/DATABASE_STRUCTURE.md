# 📊 SkillMatch Database Structure

## Overview
ระบบใช้ **PostgreSQL 15** แบ่งโครงสร้างเป็น **16 tables** จัดกลุ่มตามหน้าที่การใช้งาน

---

## 🗂️ Tables by Category

### 1️⃣ Core User Tables (ข้อมูลผู้ใช้พื้นฐาน)

#### `users` - ตารางผู้ใช้หลัก
```sql
- user_id (PK)
- username
- email (unique)
- password_hash
- gender_id (FK → genders)
- tier_id (FK → tiers)
- verification_status (unverified/pending/approved/rejected)
- is_admin
- google_id
- google_profile_picture
- registration_date
```

**ตัวอย่างข้อมูล:**
| user_id | username | email | gender_id | tier_id | verification_status | is_admin |
|---------|----------|-------|-----------|---------|-------------------|----------|
| 1 | The BOB Film | audikoratair@gmail.com | 1 | 1 | verified | true |
| 3 | bella_bangkok | bella@example.com | 2 | 2 | approved | false |
| 8 | marco_thai | marco@example.com | 1 | 2 | approved | false |
| 13 | kim_beauty | kim@example.com | 3 | 3 | approved | false |

#### `user_profiles` - โปรไฟล์เพิ่มเติม
```sql
- user_id (PK, FK → users)
- bio
- age, height, weight
- ethnicity
- languages (array)
- working_hours
- is_available
- service_type (incall/outcall/both)
- skills (array)
- province, district, sub_district
- latitude, longitude
- address details
```

**ตัวอย่างข้อมูล:**
| user_id | bio | age | height | service_type | province | is_available |
|---------|-----|-----|--------|--------------|----------|--------------|
| 3 | Professional massage therapist... 🌸 | 25 | 165 | both | กรุงเทพมหานคร | true |
| 8 | Athletic personal trainer... 💪 | 29 | 178 | both | กรุงเทพมหานคร | true |
| 13 | Beautiful ladyboy escort... 💋 | 26 | 172 | outcall | กรุงเทพมหานคร | true |

#### `user_photos` - รูปภาพผู้ใช้
```sql
- photo_id (PK)
- user_id (FK → users)
- photo_url
- sort_order
- uploaded_at
```

**ตัวอย่างข้อมูล:**
| photo_id | user_id | photo_url | sort_order | uploaded_at |
|----------|---------|-----------|------------|-------------|
| 1 | 3 | https://storage.googleapis.com/.../photo1.jpg | 1 | 2025-11-01 |
| 2 | 3 | https://storage.googleapis.com/.../photo2.jpg | 2 | 2025-11-01 |
| 3 | 8 | https://storage.googleapis.com/.../photo3.jpg | 1 | 2025-11-02 |

#### `user_verifications` - KYC Verification
```sql
- verification_id (PK)
- user_id (FK → users)
- national_id_url
- health_cert_url
- face_scan_url
- submitted_at
```

**ตัวอย่างข้อมูล:**
| verification_id | user_id | national_id_url | health_cert_url | submitted_at | status |
|----------------|---------|-----------------|-----------------|--------------|--------|
| 1 | 3 | https://storage.../id_bella.jpg | https://storage.../health_bella.pdf | 2025-10-15 | approved |
| 2 | 8 | https://storage.../id_marco.jpg | https://storage.../health_marco.pdf | 2025-10-20 | approved |
| 3 | 2 | https://storage.../id_test.jpg | https://storage.../health_test.pdf | 2025-11-10 | pending |

#### `genders` - ตารางเพศ
```sql
- gender_id (PK)
- gender_name (Male, Female, LGBTQ+)
```

**ข้อมูลในระบบ:**
| gender_id | gender_name |
|-----------|-------------|
| 1 | Male |
| 2 | Female |
| 3 | LGBTQ+ |

#### `tiers` - แพ็คเกจสมาชิก
```sql
- tier_id (PK)
- name (General, Silver, Gold, Platinum)
- price_monthly
- access_level
```

**ข้อมูลในระบบ:**
| tier_id | name | price_monthly | access_level | features |
|---------|------|---------------|--------------|----------|
| 1 | General | 0.00 | 1 | Basic profile, limited visibility |
| 2 | Silver | 299.00 | 2 | Enhanced profile, priority support |
| 3 | Gold | 599.00 | 3 | Top visibility, analytics dashboard |
| 4 | Platinum | 999.00 | 4 | VIP badge, featured listing, ad-free |

---

### 2️⃣ Service & Booking Tables (บริการและการจอง)

#### `service_packages` - แพ็คเกจบริการ
```sql
- package_id (PK)
- provider_id (FK → users)
- package_name
- description
- duration (minutes)
- price
- is_active
```

**ตัวอย่างข้อมูล:**
| package_id | provider_id | package_name | duration | price | description |
|------------|-------------|--------------|----------|-------|-------------|
| 1 | 3 | 1 Hour Thai Massage | 60 | 1200.00 | Traditional Thai massage with stretching |
| 2 | 3 | 2 Hours Aromatherapy | 120 | 2500.00 | Full body aromatherapy massage |
| 3 | 4 | Dinner Date (3 Hours) | 180 | 8000.00 | Elegant companion for dinner |
| 4 | 13 | Companion Evening (4 Hours) | 240 | 10000.00 | Glamorous companion for events |

#### `bookings` - การจองบริการ
```sql
- booking_id (PK)
- client_id (FK → users)
- provider_id (FK → users)
- package_id (FK → service_packages)
- booking_date
- start_time, end_time
- total_price
- status (pending/confirmed/completed/cancelled)
- location
- special_notes
- cancellation_reason
- created_at
```

**ตัวอย่างข้อมูล:**
| booking_id | client_id | provider_id | package_id | booking_date | status | total_price |
|------------|-----------|-------------|------------|--------------|--------|-------------|
| 1 | 1 | 3 | 1 | 2025-11-20 | confirmed | 1200.00 |
| 2 | 1 | 8 | 5 | 2025-11-21 | pending | 1800.00 |
| 3 | 2 | 13 | 4 | 2025-11-22 | confirmed | 10000.00 |

#### `provider_availability` - ช่วงเวลาว่าง
```sql
- availability_id (PK)
- provider_id (FK → users)
- day_of_week
- start_time, end_time
- is_available
```

**ตัวอย่างข้อมูล:**
| availability_id | provider_id | day_of_week | start_time | end_time | is_available |
|----------------|-------------|-------------|------------|----------|--------------|
| 1 | 3 | 1 (Monday) | 10:00:00 | 22:00:00 | true |
| 2 | 3 | 2 (Tuesday) | 10:00:00 | 22:00:00 | true |
| 3 | 8 | 1 (Monday) | 08:00:00 | 20:00:00 | true |
| 4 | 8 | 0 (Sunday) | 08:00:00 | 20:00:00 | false |

---

### 3️⃣ Social Features Tables (ฟีเจอร์โซเชียล)

#### `favorites` - รายการโปรด
```sql
- favorite_id (PK)
- client_id (FK → users)
- provider_id (FK → users)
- added_at
```

**ตัวอย่างข้อมูล:**
| favorite_id | client_id | provider_id | added_at |
|-------------|-----------|-------------|----------|
| 1 | 1 | 3 | 2025-11-10 |
| 2 | 1 | 8 | 2025-11-11 |
| 3 | 1 | 13 | 2025-11-12 |
| 4 | 2 | 3 | 2025-11-09 |

#### `reviews` - รีวิวและคะแนน
```sql
- review_id (PK)
- booking_id (FK → bookings, unique)
- client_id (FK → users)
- provider_id (FK → users)
- rating (1-5)
- comment
- is_verified
- created_at
```

**ตัวอย่างข้อมูล:**
| review_id | booking_id | provider_id | rating | comment | created_at |
|-----------|------------|-------------|--------|---------|------------|
| 1 | 1 | 3 | 5 | Amazing massage! Highly recommend! 🌟 | 2025-11-03 |
| 2 | 2 | 8 | 5 | Best sports massage I've had! 💪 | 2025-11-05 |
| 3 | 3 | 13 | 5 | Kim is absolutely gorgeous! 10/10! 💋 | 2025-11-07 |

#### `blocks` - บล็อกผู้ใช้
```sql
- block_id (PK)
- blocker_id (FK → users) -- คนที่บล็อก
- blocked_id (FK → users) -- คนที่ถูกบล็อก
- reason
- blocked_at
```

**ตัวอย่างข้อมูล:**
| block_id | blocker_id | blocked_id | reason | blocked_at |
|----------|------------|------------|--------|------------|
| 1 | 1 | 2 | Spam messages | 2025-11-10 |
| 2 | 3 | 5 | Inappropriate behavior | 2025-11-11 |

---

### 4️⃣ Communication Tables (การสื่อสาร)

#### `conversations` - ห้องสนทนา
```sql
- id (PK)
- user1_id (FK → users)
- user2_id (FK → users)
- created_at
- last_message_at
UNIQUE(user1_id, user2_id)
```

**ตัวอย่างข้อมูล:**
| id | user1_id | user2_id | last_message_at |
|----|----------|----------|-----------------|
| 1 | 1 | 3 | 2025-11-13 10:30:00 |
| 2 | 1 | 8 | 2025-11-13 09:15:00 |
| 3 | 2 | 13 | 2025-11-12 20:45:00 |

#### `messages` - ข้อความ
```sql
- id (PK)
- conversation_id (FK → conversations)
- sender_id (FK → users)
- receiver_id (FK → users)
- content
- is_read
- created_at
```

**ตัวอย่างข้อมูล:**
| id | conversation_id | sender_id | receiver_id | content | is_read | created_at |
|----|----------------|-----------|-------------|---------|---------|------------|
| 1 | 1 | 1 | 3 | Hello! Are you available tomorrow? | true | 2025-11-13 10:00:00 |
| 2 | 1 | 3 | 1 | Yes, I have slots at 2pm and 5pm | true | 2025-11-13 10:15:00 |
| 3 | 1 | 1 | 3 | Great! I'll book the 2pm slot | false | 2025-11-13 10:30:00 |

#### `notifications` - การแจ้งเตือน
```sql
- id (PK)
- user_id (FK → users)
- type (booking_request/new_message/review_received/etc.)
- title
- message
- metadata (jsonb)
- is_read
- created_at
```

**ตัวอย่างข้อมูล:**
| id | user_id | type | title | message | is_read | created_at |
|----|---------|------|-------|---------|---------|------------|
| 1 | 3 | booking_request | New Booking Request | You have a new booking from The BOB Film | false | 2025-11-13 11:00:00 |
| 2 | 1 | booking_confirmed | Booking Confirmed | Your booking with bella_bangkok is confirmed | true | 2025-11-13 11:05:00 |
| 3 | 3 | new_message | New Message | You have a new message from The BOB Film | false | 2025-11-13 10:30:00 |
| 4 | 3 | review_received | New Review | You received a 5-star review! | false | 2025-11-03 12:00:00 |

---

### 5️⃣ Safety & Moderation Tables (ความปลอดภัย)

#### `reports` - รายงานผู้ใช้
```sql
- id (PK)
- reporter_id (FK → users)
- reported_user_id (FK → users)
- reason (harassment/inappropriate_content/fake_profile/etc.)
- description
- status (pending/under_review/resolved/dismissed)
- admin_notes
- resolved_by (FK → users)
- created_at
- resolved_at
```

**ตัวอย่างข้อมูล:**
| id | reporter_id | reported_user_id | reason | description | status | created_at |
|----|-------------|------------------|--------|-------------|--------|------------|
| 1 | 1 | 5 | harassment | This user sent threatening messages | pending | 2025-11-13 |
| 2 | 3 | 7 | inappropriate_content | Posted inappropriate photos | under_review | 2025-11-12 |
| 3 | 2 | 9 | fake_profile | Using fake photos and information | resolved | 2025-11-10 |

---

## 🔗 Key Relationships

```
users (1) ────→ (1) user_profiles
users (1) ────→ (n) user_photos
users (1) ────→ (n) service_packages [as provider]
users (1) ────→ (n) bookings [as client]
users (1) ────→ (n) bookings [as provider]
users (1) ────→ (n) reviews [as client]
users (1) ────→ (n) reviews [as provider]
users (1) ────→ (n) favorites
users (1) ────→ (n) blocks
users (1) ────→ (n) messages
users (1) ────→ (n) notifications
users (2) ────→ (1) conversations

service_packages (1) ─→ (n) bookings
bookings (1) ─────────→ (0..1) reviews
conversations (1) ────→ (n) messages
```

---

## 🎯 Current Mock Data Summary

### Users (19 total)
- **1** Admin user (audikoratair@gmail.com)
- **1** Test user
- **17** Mock providers:
  - 5 Female providers (bella_bangkok, sophia_silom, maya_massage, luna_therapy, nina_wellness)
  - 5 Male providers (marco_thai, alex_sports, david_fitness, ryan_wellness, jason_therapy)
  - 7 LGBTQ+ providers (kim_beauty, rose_glamour, mimi_style, angel_paradise, tony_pride, kevin_rainbow, sam_fabulous)

### Service Packages (15+)
- Each provider has 1-3 packages
- Price range: ฿1,200 - ฿80,000
- Duration: 60 - 2,880 minutes (1 hour - 2 days)

### User Profiles (17)
- All mock providers have complete profiles
- Locations: Various districts in Bangkok
- Service types: incall, outcall, both

---

## 🔍 Example Queries

### Get provider with stats
```sql
SELECT 
    u.username,
    up.age,
    up.service_type,
    COUNT(DISTINCT sp.package_id) as packages_count,
    COUNT(DISTINCT r.review_id) as reviews_count,
    AVG(r.rating) as avg_rating
FROM users u
LEFT JOIN user_profiles up ON u.user_id = up.user_id
LEFT JOIN service_packages sp ON u.user_id = sp.provider_id
LEFT JOIN reviews r ON u.user_id = r.provider_id
GROUP BY u.user_id, u.username, up.age, up.service_type;
```

### Get conversations with unread count
```sql
SELECT 
    c.id,
    u.username as other_user,
    COUNT(*) FILTER (WHERE m.is_read = false AND m.receiver_id = $1) as unread_count
FROM conversations c
JOIN users u ON (u.user_id = c.user2_id AND c.user1_id = $1) 
             OR (u.user_id = c.user1_id AND c.user2_id = $1)
LEFT JOIN messages m ON m.conversation_id = c.id
WHERE c.user1_id = $1 OR c.user2_id = $1
GROUP BY c.id, u.username;
```

### Browse providers with filters
```sql
SELECT 
    u.user_id,
    u.username,
    up.age,
    up.service_type,
    up.province,
    AVG(r.rating) as avg_rating,
    COUNT(r.review_id) as review_count
FROM users u
JOIN user_profiles up ON u.user_id = up.user_id
LEFT JOIN reviews r ON u.user_id = r.provider_id
WHERE up.is_available = true
  AND u.gender_id = $1  -- filter by gender
  AND up.province = $2   -- filter by province
GROUP BY u.user_id, u.username, up.age, up.service_type, up.province
HAVING AVG(r.rating) >= $3  -- minimum rating
ORDER BY avg_rating DESC;
```

---

## 📈 Database Indexes

### Important indexes for performance:
- `users.email` - UNIQUE index
- `users.google_id` - UNIQUE index
- `bookings.client_id` - для быстрого поиска
- `bookings.provider_id` - для быстрого поиска
- `reviews.provider_id` - для статистики
- `messages.conversation_id` - для чата
- `notifications.user_id` - для уведомлений
- `conversations (user1_id, user2_id)` - UNIQUE index

---

## 🔒 Foreign Key Constraints

All tables use `ON DELETE CASCADE` for dependent records:
- When user is deleted → all their photos, bookings, messages are deleted
- When booking is deleted → review is deleted
- When conversation is deleted → all messages are deleted

---

## 💾 Data Types Used

- **IDs**: `SERIAL` (auto-increment integers)
- **Text**: `VARCHAR`, `TEXT`
- **Numbers**: `INTEGER`, `DECIMAL(10,2)` for prices
- **Dates**: `TIMESTAMP WITH TIME ZONE`
- **Arrays**: `TEXT[]` for languages and skills
- **JSON**: `JSONB` for notification metadata
- **Boolean**: `BOOLEAN` for flags
- **Geographic**: `DECIMAL(10,8)` for latitude/longitude

---

## 🚀 Migrations

Migrations are managed in code (`migrations.go`):
- Migration 001-006: Core tables
- Migration 007: Messaging system
- Migration 008: Notifications
- Migration 009: Reports system
- Migration 010: Profile views analytics
- Migration 011: Block system

All migrations run automatically on server start! ✅
