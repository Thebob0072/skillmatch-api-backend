# 🏷️ Service Category API Documentation

## ภาพรวม
ระบบหมวดหมู่บริการ (Service Categories) ช่วยให้ผู้ให้บริการสามารถเลือกประเภทบริการที่ให้ได้หลายหมวดหมู่ และผู้ใช้งานสามารถค้นหาผู้ให้บริการตามหมวดหมู่ได้

## 📊 ข้อมูลพื้นฐาน

### หมวดหมู่ที่มีอยู่ (20 หมวดหมู่)

#### Adult Services (18+) - ต้องยืนยันอายุ
- 🔞 **adult_entertainment** - บริการผู้ใหญ่
- 💋 **escort** - แอสคอร์ท

#### Healthcare & Wellness
- 💆 **massage_therapy** - นวดบำบัด
- 🧖 **spa_wellness** - สปาและเวลเนส
- 🤲 **personal_care** - ดูแลส่วนตัว
- 🏥 **healthcare_companion** - เพื่อนดูแลสุขภาพ

#### Entertainment & Events
- 🍷 **bartender** - บาร์เทนเดอร์
- 🎉 **party_host** - พิธีกรงานปาร์ตี้
- 🎤 **karaoke_companion** - เพื่อนร้องเพลง

#### Social Companionship
- 🍽️ **dining_companion** - เพื่อนทานอาหาร
- 🎬 **movie_companion** - เพื่อนดูหนัง
- 🛍️ **shopping_companion** - เพื่อนช็อปปิ้ง
- ✈️ **travel_companion** - เพื่อนเดินทาง

#### Professional Services
- 👔 **personal_assistant** - ผู้ช่วยส่วนตัว
- 🎪 **event_companion** - เพื่อนร่วมงานอีเว้นท์
- 📚 **language_practice** - ฝึกภาษา
- 💪 **fitness_trainer** - เทรนเนอร์ส่วนตัว
- 🎨 **hobby_companion** - เพื่อนทำกิจกรรมงานอดิเรก
- 📷 **photo_model** - นางแบบถ่ายภาพ
- 🎵 **music_companion** - เพื่อนฟังเพลง/คอนเสิร์ต

### กฎและข้อจำกัด
- ผู้ให้บริการเลือกได้ **สูงสุด 5 หมวดหมู่**
- หมวดหมู่ Adult (is_adult=true) ต้องยืนยันอายุ 18+ ก่อนใช้
- หมวดหมู่มี display_order สำหรับเรียงลำดับแสดงผล
- ผู้ใช้สามารถกรองไม่แสดงหมวดหมู่ Adult ได้

---

## 📡 API Endpoints

### 1️⃣ ดูหมวดหมู่ทั้งหมด (Public)

```http
GET /service-categories
```

**Query Parameters:**
- `include_adult` (boolean, default: true) - แสดงหมวดหมู่ Adult หรือไม่

**Response 200:**
```json
{
  "categories": [
    {
      "category_id": 3,
      "name": "massage_therapy",
      "name_thai": "นวดบำบัด",
      "description": "Traditional and modern massage services",
      "icon": "💆",
      "is_adult": false,
      "display_order": 3,
      "is_active": true
    },
    ...
  ],
  "total": 20
}
```

**ตัวอย่างการใช้:**
```bash
# ดูทั้งหมด
curl http://localhost:8080/service-categories

# กรอง Adult categories ออก
curl 'http://localhost:8080/service-categories?include_adult=false'
```

---

### 2️⃣ ดูหมวดหมู่ของผู้ให้บริการ (Protected)

```http
GET /providers/:userId/categories
```

**Headers:**
```
Authorization: Bearer <token>
```

**Response 200:**
```json
{
  "provider_id": 26,
  "categories": [
    {
      "category_id": 3,
      "name": "massage_therapy",
      "name_thai": "นวดบำบัด",
      "description": "Traditional and modern massage services",
      "icon": "💆",
      "is_adult": false,
      "display_order": 3
    },
    {
      "category_id": 7,
      "name": "bartender",
      "name_thai": "บาร์เทนเดอร์",
      "description": "Professional bartending services",
      "icon": "🍷",
      "is_adult": false,
      "display_order": 7
    }
  ],
  "total": 2
}
```

**ตัวอย่างการใช้:**
```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/providers/26/categories
```

---

### 3️⃣ อัพเดทหมวดหมู่ของตัวเอง (Provider Only)

```http
PUT /provider/me/categories
```

**Headers:**
```
Authorization: Bearer <token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "category_ids": [1, 3, 7, 10]
}
```

**Validation Rules:**
- `category_ids` required, array of integers
- สูงสุด 5 หมวดหมู่
- category_id ต้องมีอยู่ในระบบ

**Response 200:**
```json
{
  "message": "Categories updated successfully",
  "category_ids": [1, 3, 7, 10],
  "total": 4
}
```

**Error 400:**
```json
{
  "error": "Cannot select more than 5 categories"
}
```

**ตัวอย่างการใช้:**
```bash
curl -X PUT http://localhost:8080/provider/me/categories \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "category_ids": [3, 7, 10]
  }'
```

---

### 4️⃣ ค้นหาผู้ให้บริการตามหมวดหมู่ (Public)

```http
GET /categories/:category_id/providers
```

**Query Parameters:**
- `page` (int, default: 1) - หน้าที่ต้องการ
- `limit` (int, default: 20, max: 50) - จำนวนผล/หน้า

**Response 200:**
```json
{
  "category_id": 7,
  "providers": [
    {
      "user_id": 26,
      "username": "testprovider1",
      "gender_id": 2,
      "age": null,
      "profile_image_url": null,
      "google_profile_picture": null,
      "province": null,
      "district": null,
      "sub_district": null,
      "average_rating": 0,
      "review_count": 0,
      "min_price": null
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 20,
  "total_pages": 1
}
```

**ตัวอย่างการใช้:**
```bash
# หน้าแรก
curl http://localhost:8080/categories/7/providers

# หน้า 2, แสดง 10 รายการ
curl 'http://localhost:8080/categories/7/providers?page=2&limit=10'
```

---

## 🗄️ Database Schema

### service_categories
```sql
CREATE TABLE service_categories (
    category_id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    name_thai VARCHAR(100) NOT NULL,
    description TEXT,
    icon VARCHAR(50),
    is_adult BOOLEAN DEFAULT false,
    display_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### provider_categories (Junction Table)
```sql
CREATE TABLE provider_categories (
    provider_category_id SERIAL PRIMARY KEY,
    provider_id INTEGER NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    category_id INTEGER NOT NULL REFERENCES service_categories(category_id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(provider_id, category_id)
);
```

**Indexes:**
- `idx_provider_categories_provider` ON provider_categories(provider_id)
- `idx_provider_categories_category` ON provider_categories(category_id)
- `idx_service_categories_active` ON service_categories(is_active, display_order)

---

## 🔄 Workflow Examples

### 📝 Provider สร้างโปรไฟล์และเลือกหมวดหมู่

```bash
# 1. Register
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "massage_pro",
    "email": "massage@example.com",
    "password": "securepass123",
    "gender_id": 2,
    "subscription_tier_id": 2
  }'
# Response: {"user_id": 27, "message": "User created successfully"}

# 2. Login
TOKEN=$(curl -s -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "massage@example.com",
    "password": "securepass123"
  }' | jq -r '.token')

# 3. เลือกหมวดหมู่บริการ
curl -X PUT http://localhost:8080/provider/me/categories \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "category_ids": [3, 4, 5]
  }'
# Response: {"message": "Categories updated successfully", "total": 3}

# 4. ดูหมวดหมู่ที่เลือก
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/providers/27/categories
```

### 🔍 User ค้นหาผู้ให้บริการ

```bash
# 1. ดูหมวดหมู่ทั้งหมด (ไม่แสดง Adult)
curl 'http://localhost:8080/service-categories?include_adult=false'

# 2. เลือกหมวดหมู่ "นวดบำบัด" (ID 3)
curl http://localhost:8080/categories/3/providers

# 3. ดูรายละเอียดผู้ให้บริการ
curl -H "Authorization: Bearer $USER_TOKEN" \
  http://localhost:8080/provider/27

# 4. ดูหมวดหมู่ของผู้ให้บริการคนนั้น
curl -H "Authorization: Bearer $USER_TOKEN" \
  http://localhost:8080/providers/27/categories
```

---

## 🎨 Frontend Integration Guide

### React Component Example

```typescript
// types.ts
interface ServiceCategory {
  category_id: number;
  name: string;
  name_thai: string;
  description?: string;
  icon?: string;
  is_adult: boolean;
  display_order: number;
  is_active: boolean;
}

// API calls
const getCategories = async (includeAdult = true) => {
  const response = await fetch(
    `/service-categories?include_adult=${includeAdult}`
  );
  return response.json();
};

const updateProviderCategories = async (categoryIds: number[]) => {
  const response = await fetch('/provider/me/categories', {
    method: 'PUT',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ category_ids: categoryIds })
  });
  return response.json();
};

const getProvidersByCategory = async (categoryId: number, page = 1) => {
  const response = await fetch(
    `/categories/${categoryId}/providers?page=${page}`
  );
  return response.json();
};

// CategorySelector Component
const CategorySelector = () => {
  const [categories, setCategories] = useState<ServiceCategory[]>([]);
  const [selected, setSelected] = useState<number[]>([]);

  useEffect(() => {
    getCategories(false).then(data => setCategories(data.categories));
  }, []);

  const handleToggle = (id: number) => {
    if (selected.includes(id)) {
      setSelected(selected.filter(x => x !== id));
    } else if (selected.length < 5) {
      setSelected([...selected, id]);
    } else {
      alert('สามารถเลือกได้สูงสุด 5 หมวดหมู่');
    }
  };

  const handleSave = async () => {
    await updateProviderCategories(selected);
    alert('บันทึกสำเร็จ!');
  };

  return (
    <div>
      <h3>เลือกหมวดหมู่บริการ (เลือกได้สูงสุด 5)</h3>
      <div className="grid grid-cols-3 gap-4">
        {categories.map(cat => (
          <button
            key={cat.category_id}
            onClick={() => handleToggle(cat.category_id)}
            className={selected.includes(cat.category_id) ? 'selected' : ''}
          >
            <span>{cat.icon}</span>
            <span>{cat.name_thai}</span>
          </button>
        ))}
      </div>
      <button onClick={handleSave}>บันทึก ({selected.length}/5)</button>
    </div>
  );
};

// Browse by Category
const BrowseByCategory = ({ categoryId }: { categoryId: number }) => {
  const [providers, setProviders] = useState([]);
  const [page, setPage] = useState(1);

  useEffect(() => {
    getProvidersByCategory(categoryId, page)
      .then(data => setProviders(data.providers));
  }, [categoryId, page]);

  return (
    <div>
      {providers.map(provider => (
        <ProviderCard key={provider.user_id} {...provider} />
      ))}
    </div>
  );
};
```

---

## 🧪 Testing

### Test Data
```sql
-- Insert test provider categories
INSERT INTO provider_categories (provider_id, category_id) VALUES
  (3, 1), (3, 3), (3, 7),   -- bella_bangkok: adult + massage + bartender
  (4, 2), (4, 4),           -- sophia_silom: escort + spa
  (5, 3), (5, 5), (5, 6);   -- maya_massage: massage + care + health
```

### cURL Test Suite
```bash
# Test 1: Get all categories
curl http://localhost:8080/service-categories | jq '.total'
# Expected: 20

# Test 2: Get non-adult categories
curl 'http://localhost:8080/service-categories?include_adult=false' | jq '.total'
# Expected: 18

# Test 3: Login and update categories
TOKEN=$(curl -s -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email": "test@test.com", "password": "testpass"}' \
  | jq -r '.token')

curl -X PUT http://localhost:8080/provider/me/categories \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"category_ids": [3, 7, 10]}'
# Expected: {"message": "Categories updated successfully", "total": 3}

# Test 4: Get provider categories
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/providers/26/categories | jq '.total'
# Expected: 3

# Test 5: Browse by category
curl http://localhost:8080/categories/7/providers | jq '.total'
# Expected: 1 or more

# Test 6: Try to select > 5 categories (should fail)
curl -X PUT http://localhost:8080/provider/me/categories \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"category_ids": [1, 2, 3, 4, 5, 6]}'
# Expected: {"error": "Cannot select more than 5 categories"}
```

---

## 🚀 Deployment Checklist

- [x] Migration 012 executed (service_categories + provider_categories tables)
- [x] 20 service categories seeded
- [x] API handlers created (category_handlers.go)
- [x] Routes registered in main.go
- [x] Indexes created for performance
- [x] Adult category filter working
- [x] Provider category selection working (max 5)
- [x] Browse by category working
- [x] Documentation complete

---

## 📊 Performance Notes

- **Indexes**: 3 indexes on provider_categories and service_categories
- **Query Performance**: Browse by category uses JOIN with aggregation (AVG rating, COUNT reviews)
- **Caching**: Consider caching service_categories (rarely changes)
- **Pagination**: Default 20 items/page, max 50

---

## 🔮 Future Enhancements

1. **Category Search**: เพิ่ม search bar สำหรับค้นหาหมวดหมู่
2. **Multi-Category Filter**: ให้กรองหลายหมวดหมู่พร้อมกันใน browse
3. **Category Analytics**: สถิติว่าหมวดหมู่ไหนได้รับความนิยม
4. **Custom Categories**: ให้ provider สร้างหมวดหมู่เองได้ (pending admin approval)
5. **Category Badges**: แสดง badge บนโปรไฟล์ผู้ให้บริการ
6. **Age Verification**: เพิ่มระบบยืนยันอายุก่อนดูหมวดหมู่ Adult

---

**Last Updated:** November 14, 2025  
**API Version:** 1.0  
**Migration:** 012_add_service_categories.sql
