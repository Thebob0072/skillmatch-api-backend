# 🚀 SkillMatch API - Frontend Integration Guide

## 🗺️ ระบบที่อยู่แบบละเอียด + ระยะทาง

### ⚠️ สำคัญ! ข้อมูลที่อยู่มี 2 แบบ:

1. **ข้อมูลสาธารณะ** (แสดงทุกคน):
   - `province` (จังหวัด)
   - `district` (เขต/อำเภอ)
   - `sub_district` (แขวง/ตำบล)
   - `latitude`, `longitude` (สำหรับคำนวณระยะทาง)

2. **ข้อมูลละเอียด** (แสดงเฉพาะหลัง booking confirmed):
   - `address_line1` (บ้านเลขที่, ถนน, ซอย)
   - `postal_code` (รหัสไปรษณีย์)
   - พิกัดแม่นยำ

---

## ✅ ฟีเจอร์ใหม่ที่เพิ่มเข้ามา

### 1. 📦 Service Packages (แพ็คเกจบริการ)

#### GET `/packages/:providerId` - ดูแพ็คเกจของ Provider
```typescript
const getProviderPackages = async (providerId: number) => {
  const response = await api.get(`/packages/${providerId}`);
  return response.data; // ServicePackage[]
};

interface ServicePackage {
  package_id: number;
  provider_id: number;
  package_name: string;
  description: string | null;
  duration: number; // นาที
  price: number;
  is_active: boolean;
  created_at: string;
}
```

#### POST `/packages` - สร้างแพ็คเกจ (Provider เท่านั้น)
```typescript
const createPackage = async (data: {
  package_name: string;
  description?: string;
  duration: number; // นาที
  price: number;
}) => {
  const response = await api.post('/packages', data);
  return response.data;
};
```

---

### 2. 📅 Booking System (ระบบจองบริการ)

#### POST `/bookings` - จองบริการ
```typescript
const createBooking = async (data: {
  provider_id: number;
  package_id: number;
  booking_date: string; // "YYYY-MM-DD"
  start_time: string;   // "HH:MM"
  location?: string;
  special_notes?: string;
}) => {
  const response = await api.post('/bookings', data);
  return response.data;
};
```

#### GET `/bookings/my` - ดูการจองของตัวเอง (Client)
```typescript
const getMyBookings = async () => {
  const response = await api.get('/bookings/my');
  return response.data; // BookingWithDetails[]
};

interface BookingWithDetails {
  booking_id: number;
  client_id: number;
  client_username: string;
  provider_id: number;
  provider_username: string;
  provider_profile_pic: string | null;
  package_name: string;
  duration: number;
  booking_date: string;
  start_time: string;
  end_time: string;
  total_price: number;
  status: "pending" | "confirmed" | "completed" | "cancelled";
  location: string | null;
  special_notes: string | null;
  created_at: string;
  updated_at: string;
}
```

#### GET `/bookings/provider` - ดูการจองที่เข้ามา (Provider)
```typescript
const getProviderBookings = async () => {
  const response = await api.get('/bookings/provider');
  return response.data; // BookingWithDetails[]
};
```

#### PATCH `/bookings/:id/status` - อัพเดทสถานะการจอง
```typescript
const updateBookingStatus = async (
  bookingId: number,
  data: {
    status: "confirmed" | "completed" | "cancelled";
    cancellation_reason?: string;
  }
) => {
  const response = await api.patch(`/bookings/${bookingId}/status`, data);
  return response.data;
};
```

---

### 3. ⭐ Review & Rating System

#### POST `/reviews` - สร้างรีวิว (หลังใช้บริการเสร็จ)
```typescript
const createReview = async (data: {
  booking_id: number;
  rating: number; // 1-5
  comment?: string;
}) => {
  const response = await api.post('/reviews', data);
  return response.data;
};
```

#### GET `/reviews/:providerId` - ดูรีวิวของ Provider
```typescript
const getProviderReviews = async (providerId: number) => {
  const response = await api.get(`/reviews/${providerId}`);
  return response.data; // ReviewWithDetails[]
};

interface ReviewWithDetails {
  review_id: number;
  client_username: string;
  rating: number;
  comment: string | null;
  is_verified: boolean;
  created_at: string;
}
```

#### GET `/reviews/stats/:providerId` - สถิติรีวิวของ Provider
```typescript
const getProviderReviewStats = async (providerId: number) => {
  const response = await api.get(`/reviews/stats/${providerId}`);
  return response.data;
};

interface ReviewStats {
  total_reviews: number;
  average_rating: number;
  rating_5: number;
  rating_4: number;
  rating_3: number;
  rating_2: number;
  rating_1: number;
}
```

---

### 4. ❤️ Favorites System (รายการโปรด)

#### POST `/favorites` - เพิ่มรายการโปรด
```typescript
const addFavorite = async (providerId: number) => {
  const response = await api.post('/favorites', { provider_id: providerId });
  return response.data;
};
```

#### DELETE `/favorites/:providerId` - ลบรายการโปรด
```typescript
const removeFavorite = async (providerId: number) => {
  await api.delete(`/favorites/${providerId}`);
};
```

#### GET `/favorites` - ดูรายการโปรด
```typescript
const getMyFavorites = async () => {
  const response = await api.get('/favorites');
  return response.data; // FavoriteProvider[]
};

interface FavoriteProvider {
  user_id: number;
  username: string;
  tier_name: string;
  gender_id: number;
  profile_image_url: string | null;
  google_profile_picture: string | null;
  average_rating: number;
  review_count: number;
  // ข้อมูลที่อยู่
  province: string | null;
  district: string | null;
  sub_district: string | null;
  latitude: number | null;
  longitude: number | null;
  distance_km: number | null; // ระยะทางจากตำแหน่งของคุณ
}
```

#### GET `/favorites/check/:providerId` - เช็คว่าอยู่ในรายการโปรดหรือไม่
```typescript
const checkFavorite = async (providerId: number) => {
  const response = await api.get(`/favorites/check/${providerId}`);
  return response.data.is_favorite; // boolean
};
```

---

### 5. 🔍 Browse/Search System - ค้นหาผู้ให้บริการแบบละเอียด

#### GET `/browse/v2` - ค้นหาผู้ให้บริการ (รองรับทุก filter)

**Query Parameters ทั้งหมด:**

```typescript
interface BrowseFilters {
  // ฟิลเตอร์พื้นฐาน
  gender?: number;          // 0 = All, 1 = Male, 2 = Female, 3 = Other
  available?: boolean;      // true = เฉพาะที่ว่าง
  min_age?: number;
  max_age?: number;
  min_price?: number;
  max_price?: number;
  min_rating?: number;      // 0-5
  ethnicity?: string;       // "thai", "chinese", "japanese", etc.
  service_type?: string;    // "incall", "outcall" (ไม่มี "both")
  
  // ฟิลเตอร์ที่อยู่แบบละเอียด
  province?: string;        // "กรุงเทพมหานคร", "เชียงใหม่"
  district?: string;        // "บางรัก", "เมือง"
  sub_district?: string;    // "สีลม", "ช้างคลาน"
  
  // ฟิลเตอร์ระยะทาง
  max_distance?: number;    // กิโลเมตร (ต้องมีพิกัด GPS ของ user)
}

// ตัวอย่างการใช้งาน
const searchProviders = async (filters: BrowseFilters) => {
  const params = new URLSearchParams();
  
  if (filters.gender) params.append('gender', filters.gender.toString());
  if (filters.available) params.append('available', 'true');
  if (filters.province) params.append('province', filters.province);
  if (filters.district) params.append('district', filters.district);
  if (filters.sub_district) params.append('sub_district', filters.sub_district);
  if (filters.max_distance) params.append('max_distance', filters.max_distance.toString());
  // ... เพิ่ม params อื่นๆ
  
  const response = await api.get(`/browse/v2?${params.toString()}`);
  return response.data; // BrowsableUser[]
};

interface BrowsableUser {
  user_id: number;
  username: string;
  tier_name: string;
  gender_id: number;
  profile_image_url: string | null;
  google_profile_picture: string | null;
  age: number | null;
  location: string | null; // Legacy field (deprecated)
  is_available: boolean;
  average_rating: number;
  review_count: number;
  min_price: number | null;
  
  // ข้อมูลที่อยู่แบบละเอียด
  province: string | null;     // จังหวัด
  district: string | null;     // เขต/อำเภอ
  sub_district: string | null; // แขวง/ตำบล
  latitude: number | null;     // พิกัด GPS
  longitude: number | null;    // พิกัด GPS
  distance_km: number | null;  // ระยะทางจากตำแหน่งของคุณ (กิโลเมตร)
  
  // ประเภทการให้บริการ
  service_type: "incall" | "outcall" | null; // incall = มีสถานที่, outcall = ไปหาลูกค้า
}
```

**ตัวอย่างการใช้งาน:**

```typescript
// 1. ค้นหาทั้งหมด
const allProviders = await searchProviders({});

// 2. ค้นหาใน กรุงเทพฯ เขตบางรัก
const bangrakProviders = await searchProviders({
  province: "กรุงเทพมหานคร",
  district: "บางรัก"
});

// 3. ค้นหาในรัศมี 5 กม. (ต้องมีพิกัด GPS ของ user)
const nearbyProviders = await searchProviders({
  max_distance: 5
});

// 4. ค้นหาแบบละเอียด
const detailedSearch = await searchProviders({
  province: "กรุงเทพมหานคร",
  district: "บางรัก",
  sub_district: "สีลม",
  gender: 2,
  min_age: 25,
  max_age: 35,
  min_rating: 4.0,
  available: true,
  max_distance: 3,
  service_type: "both"
});
```

---

### 6. 👤 Profile Management - จัดการข้อมูลส่วนตัว

#### PUT `/profile/me` - อัพเดทข้อมูลโปรไฟล์

```typescript
const updateMyProfile = async (data: {
  bio?: string;
  location?: string; // Legacy (deprecated)
  skills?: string[];
  
  // ข้อมูลที่อยู่แบบละเอียด
  province?: string;      // จังหวัด
  district?: string;      // เขต/อำเภอ
  sub_district?: string;  // แขวง/ตำบล
  postal_code?: string;   // รหัสไปรษณีย์
  address_line1?: string; // บ้านเลขที่ ถนน ซอย
  latitude?: number;      // พิกัด GPS
  longitude?: number;     // พิกัด GPS
}) => {
  const response = await api.put('/profile/me', data);
  return response.data;
};

// ตัวอย่างการใช้งาน
await updateMyProfile({
  bio: "Professional service provider",
  province: "กรุงเทพมหานคร",
  district: "บางรัก",
  sub_district: "สีลม",
  postal_code: "10500",
  address_line1: "123 ถนนสีลม",
  latitude: 13.7278,
  longitude: 100.5318
});
```

---

### 7. 🗺️ วิธีขอ GPS Location จาก Browser

```typescript
// ขอพิกัด GPS จาก browser
const getUserLocation = (): Promise<{ latitude: number; longitude: number }> => {
  return new Promise((resolve, reject) => {
    if (!navigator.geolocation) {
      reject(new Error('Geolocation is not supported'));
      return;
    }
    
    navigator.geolocation.getCurrentPosition(
      (position) => {
        resolve({
          latitude: position.coords.latitude,
          longitude: position.coords.longitude
        });
      },
      (error) => reject(error),
      {
        enableHighAccuracy: true,
        timeout: 10000,
        maximumAge: 0
      }
    );
  });
};

// วิธีใช้งาน
const handleGetLocation = async () => {
  try {
    const location = await getUserLocation();
    console.log('User location:', location);
    
    // บันทึกพิกัดใน profile
    await updateMyProfile({
      latitude: location.latitude,
      longitude: location.longitude
    });
    
    // ค้นหา providers ใกล้เคียง
    const nearbyProviders = await searchProviders({
      max_distance: 5 // 5 กม.
    });
  } catch (error) {
    console.error('Cannot get location:', error);
    // แสดง error แจ้งผู้ใช้
  }
};
```

---

### 8. 🌍 แสดงระยะทางให้ User เห็น

```typescript
// Component แสดงระยะทาง
const DistanceBadge = ({ distance }: { distance: number | null }) => {
  if (!distance) return null;
  
  return (
    <div className="distance-badge">
      <svg>📍</svg>
      <span>{distance.toFixed(1)} กม.</span>
    </div>
  );
};

// ใน Provider Card
const ProviderCard = ({ provider }: { provider: BrowsableUser }) => {
  return (
    <div className="provider-card">
      <img src={provider.profile_image_url || provider.google_profile_picture} />
      <h3>{provider.username}</h3>
      
      {/* แสดงที่อยู่ */}
      <p className="location">
        📍 {provider.sub_district && `${provider.sub_district}, `}
        {provider.district && `${provider.district}, `}
        {provider.province}
      </p>
      
      {/* แสดงระยะทาง */}
      <DistanceBadge distance={provider.distance_km} />
      
      <p>⭐ {provider.average_rating.toFixed(1)} ({provider.review_count})</p>
    </div>
  );
};
```

---

### 9. 🔍 ตัวอย่าง Search Filter UI

```typescript
const SearchFilters = () => {
  const [filters, setFilters] = useState<BrowseFilters>({});
  const [provinces] = useState([
    "กรุงเทพมหานคร",
    "เชียงใหม่",
    "ภูเก็ต",
    "ชลบุรี",
    // ... เพิ่มจังหวัดอื่นๆ
  ]);
  
  const handleSearch = async () => {
    const results = await searchProviders(filters);
    // แสดงผลลัพธ์
  };
  
  return (
    <div className="search-filters">
      {/* จังหวัด */}
      <select 
        value={filters.province || ''} 
        onChange={(e) => setFilters({...filters, province: e.target.value})}
      >
        <option value="">ทุกจังหวัด</option>
        {provinces.map(p => <option key={p} value={p}>{p}</option>)}
      </select>
      
      {/* เขต/อำเภอ */}
      <input
        type="text"
        placeholder="เขต/อำเภอ"
        value={filters.district || ''}
        onChange={(e) => setFilters({...filters, district: e.target.value})}
      />
      
      {/* แขวง/ตำบล */}
      <input
        type="text"
        placeholder="แขวง/ตำบล"
        value={filters.sub_district || ''}
        onChange={(e) => setFilters({...filters, sub_district: e.target.value})}
      />
      
      {/* ระยะทาง */}
      <input
        type="number"
        placeholder="ระยะทางสูงสุด (กม.)"
        value={filters.max_distance || ''}
        onChange={(e) => setFilters({...filters, max_distance: parseFloat(e.target.value)})}
      />
      
      {/* ช่วงราคา */}
      <input
        type="number"
        placeholder="ราคาต่ำสุด"
        value={filters.min_price || ''}
        onChange={(e) => setFilters({...filters, min_price: parseFloat(e.target.value)})}
      />
      <input
        type="number"
        placeholder="ราคาสูงสุด"
        value={filters.max_price || ''}
        onChange={(e) => setFilters({...filters, max_price: parseFloat(e.target.value)})}
      />
      
      {/* คะแนนขั้นต่ำ */}
      <input
        type="number"
        step="0.1"
        min="0"
        max="5"
        placeholder="คะแนนขั้นต่ำ"
        value={filters.min_rating || ''}
        onChange={(e) => setFilters({...filters, min_rating: parseFloat(e.target.value)})}
      />
      
      {/* ปุ่มค้นหา */}
      <button onClick={handleSearch}>🔍 ค้นหา</button>
    </div>
  );
};
```

---

### 10. 📍 ข้อควรระวัง

1. **พิกัด GPS**:
   - ต้องขอ permission จาก browser ก่อน
   - ควรเก็บ cache พิกัดของ user (ไม่ต้องถามทุกครั้ง)
   - Provider ควรตั้งพิกัดแม่นยำเพื่อการคำนวณระยะทาง

2. **ที่อยู่ละเอียด**:
   - `address_line1` แสดงเฉพาะหลัง booking confirmed
   - ไม่ควรแสดงที่อยู่บ้านเลขที่แบบเต็มใน browse page

3. **ระยะทาง**:
   - คำนวณจากพิกัด GPS (Haversine formula)
   - แม่นยำพอสมควร แต่ไม่ใช่ระยะทางเดินทางจริง
   - ควรแสดงคำเตือนว่าเป็นระยะทางโดยตรง

4. **Performance**:
   - การคำนวณระยะทางทำฝั่ง backend
   - ใช้ index สำหรับค้นหาจังหวัด/เขต/แขวง
   - พิจารณาใช้ pagination ถ้าผลลัพธ์มากเกินไป

---

### 11. 🗄️ Database Migration

ต้อง run migration ก่อนใช้งาน:

```bash
psql -U your_user -d skillmatch -f migrations/005_add_location_details.sql
```

หรือถ้าใช้ Docker:

```bash
docker exec -i skillmatch-postgres psql -U postgres -d skillmatch < migrations/005_add_location_details.sql
```

---

## 🚨 Breaking Changes

### ⚠️ เปลี่ยนแปลงจากเดิม:

1. **BrowsableUser** เพิ่มฟิลด์:
   - `province`, `district`, `sub_district`
   - `latitude`, `longitude`
   - `distance_km`

2. **Query Parameters ใหม่ใน `/browse/v2`**:
   - `province`, `district`, `sub_district`
   - `max_distance`

3. **ProfileUpdate API** รับฟิลด์เพิ่ม:
   - `province`, `district`, `sub_district`
   - `postal_code`, `address_line1`
   - `latitude`, `longitude`

---

## 📊 ตัวอย่าง Use Cases

### Use Case 1: ค้นหา Provider ใกล้ฉัน
```typescript
// 1. ขอพิกัด GPS
const myLocation = await getUserLocation();

// 2. บันทึกพิกัดของตัวเอง
await updateMyProfile({
  latitude: myLocation.latitude,
  longitude: myLocation.longitude
});

// 3. ค้นหาใน รัศมี 3 กม.
const nearby = await searchProviders({ max_distance: 3 });
```

### Use Case 2: ค้นหาแบบละเอียด
```typescript
const results = await searchProviders({
  province: "กรุงเทพมหานคร",
  district: "วัฒนา",
  sub_district: "คลองเตย",
  gender: 2,
  min_age: 25,
  max_age: 35,
  min_rating: 4.5,
  available: true,
  service_type: "both"
});
```

### Use Case 3: แสดงระยะทางในรายการโปรด
```typescript
const favorites = await getMyFavorites();

favorites.forEach(provider => {
  console.log(`${provider.username}: ${provider.distance_km} กม.`);
});
```

---

## 🔢 ตัวอย่างข้อมูล JSON Response

### GET `/browse/v2?province=กรุงเทพมหานคร&max_distance=5`

```json
[
  {
    "user_id": 10,
    "username": "alice_pro",
    "tier_name": "VIP",
    "gender_id": 2,
    "profile_image_url": "https://storage.googleapis.com/...",
    "google_profile_picture": null,
    "age": 28,
    "location": "Bangkok",
    "is_available": true,
    "average_rating": 4.8,
    "review_count": 45,
    "min_price": 1500,
    "province": "กรุงเทพมหานคร",
    "district": "บางรัก",
    "sub_district": "สีลม",
    "latitude": 13.7278,
    "longitude": 100.5318,
    "distance_km": 2.34
  },
  {
    "user_id": 15,
    "username": "bob_premium",
    "tier_name": "Premium",
    "gender_id": 1,
    "profile_image_url": null,
    "google_profile_picture": "https://lh3.googleusercontent.com/...",
    "age": 32,
    "location": "Bangkok",
    "is_available": true,
    "average_rating": 4.5,
    "review_count": 28,
    "min_price": 2000,
    "province": "กรุงเทพมหานคร",
    "district": "วัฒนา",
    "sub_district": "คลองเตย",
    "latitude": 13.7307,
    "longitude": 100.5418,
    "distance_km": 4.12
  }
]
```

---

### 12. 🔍 Advanced Browse with Filters (Legacy - Deprecated)

#### GET `/browse/v2` - Browse ด้วย Advanced Filters
```typescript
const browseProviders = async (filters?: {
  gender?: number;          // 1=Male, 2=Female, 3=Other
  location?: string;        // จังหวัด
  available?: boolean;      // เฉพาะที่ว่าง
  min_age?: number;
  max_age?: number;
  min_price?: number;
  max_price?: number;
  min_rating?: number;      // 1-5
  ethnicity?: string;
  service_type?: string;    // "incall" | "outcall" | "both"
}) => {
  const params = new URLSearchParams();
  if (filters) {
    Object.entries(filters).forEach(([key, value]) => {
      if (value !== undefined) params.append(key, String(value));
    });
  }
  const response = await api.get(`/browse/v2?${params}`);
  return response.data; // BrowsableUser[]
};

interface BrowsableUser {
  user_id: number;
  username: string;
  tier_name: string;
  gender_id: number;
  profile_image_url: string | null;
  google_profile_picture: string | null;
  age: number | null;
  location: string | null;
  is_available: boolean;
  average_rating: number;
  review_count: number;
  min_price: number | null;
}
```

---

### 6. 👤 Advanced Profile Fields

#### อัพเดท UserProfile interface เพิ่มฟิลด์ใหม่
```typescript
interface UserProfile {
  user_id: number;
  bio: string | null;
  location: string | null;
  skills: string[];
  profile_image_url: string | null;
  updated_at: string;
  // ฟิลด์ใหม่
  age: number | null;
  height: number | null;        // cm
  weight: number | null;        // kg
  ethnicity: string | null;
  languages: string[];
  working_hours: string | null; // "9:00-22:00"
  is_available: boolean;        // ว่างหรือไม่
  service_type: string | null;  // "incall" | "outcall" | "both"
}
```

#### PUT `/profile/me` - อัพเดท Profile ด้วยฟิลด์ใหม่
```typescript
const updateMyProfile = async (data: {
  bio?: string;
  location?: string;
  skills?: string[];
  age?: number;
  height?: number;
  weight?: number;
  ethnicity?: string;
  languages?: string[];
  working_hours?: string;
  is_available?: boolean;
  service_type?: string;
}) => {
  const response = await api.put('/profile/me', data);
  return response.data;
};
```

---

## 📋 Use Cases สำหรับหน้า Frontend

### 1. หน้า Provider Profile
```typescript
// แสดงแพ็คเกจ
const packages = await getProviderPackages(providerId);

// แสดงรีวิวและสถิติ
const reviews = await getProviderReviews(providerId);
const stats = await getProviderReviewStats(providerId);

// เช็คว่าอยู่ในรายการโปรดไหม
const isFavorite = await checkFavorite(providerId);

// ปุ่มเพิ่ม/ลบรายการโปรด
if (isFavorite) {
  await removeFavorite(providerId);
} else {
  await addFavorite(providerId);
}
```

### 2. หน้า Booking
```typescript
// 1. เลือกแพ็คเกจ
const packages = await getProviderPackages(providerId);

// 2. จองบริการ
await createBooking({
  provider_id: providerId,
  package_id: selectedPackage.package_id,
  booking_date: "2025-11-15",
  start_time: "14:00",
  location: "Bangkok",
  special_notes: "Please arrive on time"
});

// 3. ดูสถานะการจอง
const myBookings = await getMyBookings();
```

### 3. หน้า My Bookings (Client)
```typescript
const bookings = await getMyBookings();

// ยกเลิกการจอง
await updateBookingStatus(bookingId, {
  status: "cancelled",
  cancellation_reason: "Change of plans"
});

// รีวิวหลังใช้บริการ
await createReview({
  booking_id: bookingId,
  rating: 5,
  comment: "Excellent service!"
});
```

### 4. หน้า Provider Dashboard
```typescript
// ดูการจองที่เข้ามา
const bookings = await getProviderBookingsHandler();

// ยืนยันการจอง
await updateBookingStatus(bookingId, { status: "confirmed" });

// เสร็จสิ้นบริการ
await updateBookingStatus(bookingId, { status: "completed" });

// สร้างแพ็คเกจใหม่
await createPackage({
  package_name: "2 Hours Premium",
  description: "2 hours of premium service",
  duration: 120,
  price: 3000
});
```

### 5. หน้า Browse with Filters
```typescript
const providers = await browseProviders({
  location: "Bangkok",
  available: true,
  min_rating: 4,
  min_age: 20,
  max_age: 35,
  service_type: "both",
  min_price: 1000,
  max_price: 5000
});
```

### 6. หน้า Favorites
```typescript
const favorites = await getMyFavorites();
```

---

## 🎨 UI Components ที่ควรมี

### 1. PackageCard Component
```tsx
<PackageCard
  name={pkg.package_name}
  duration={pkg.duration}
  price={pkg.price}
  description={pkg.description}
  onBook={() => handleBooking(pkg.package_id)}
/>
```

### 2. ReviewCard Component
```tsx
<ReviewCard
  username={review.client_username}
  rating={review.rating}
  comment={review.comment}
  isVerified={review.is_verified}
  createdAt={review.created_at}
/>
```

### 3. BookingCard Component
```tsx
<BookingCard
  provider={booking.provider_username}
  package={booking.package_name}
  date={booking.booking_date}
  time={booking.start_time}
  status={booking.status}
  price={booking.total_price}
  onCancel={() => handleCancel(booking.booking_id)}
  onReview={() => handleReview(booking.booking_id)}
/>
```

### 4. ProviderCard Component (Updated)
```tsx
<ProviderCard
  username={user.username}
  age={user.age}
  location={user.location}
  isAvailable={user.is_available}
  averageRating={user.average_rating}
  reviewCount={user.review_count}
  minPrice={user.min_price}
  isFavorite={isFavorite}
  onToggleFavorite={() => handleToggleFavorite(user.user_id)}
/>
```

### 5. FilterPanel Component
```tsx
<FilterPanel
  onFilter={(filters) => handleFilter(filters)}
  filters={{
    location: "Bangkok",
    available: true,
    minAge: 20,
    maxAge: 35,
    minRating: 4,
    minPrice: 1000,
    maxPrice: 5000
  }}
/>
```

### 6. KYCUploadForm Component (ใหม่!)
```tsx
<KYCUploadForm
  onSubmit={async (data) => {
    // 1. Upload ไฟล์ไปยัง GCS ด้วย signed URLs
    await uploadToGCS(data.nationalIdFile, nationalIdUrl);
    await uploadToGCS(data.healthCertFile, healthCertUrl);
    await uploadToGCS(data.faceSelfieFile, faceSelfieUrl);
    
    // 2. ส่ง keys พร้อมวันเกิดไปยัง backend
    await submitVerification({
      national_id_key: nationalIdKey,
      health_cert_key: healthCertKey,
      face_scan_key: faceSelfieKey,
      birth_date: data.birthDate
    });
  }}
>
  <ImageUploader
    label="บัตรประชาชน"
    helpText="ถ่ายให้เห็นหน้าชัดเจน ไม่มีแสงสะท้อน"
    accept="image/*"
    required
  />
  <ImageUploader
    label="รูปถ่ายใบหน้า (Selfie)"
    helpText="ถ่ายหน้าตรง ไม่สวมแว่นดำหรือหมวก"
    accept="image/*"
    capture="user" // เปิดกล้องหน้า
    required
  />
  <ImageUploader
    label="ใบรับรองสุขภาพ"
    helpText="ออกโดยโรงพยาบาล ไม่เกิน 6 เดือน"
    accept="image/*,application/pdf"
    required
  />
  <DatePicker
    label="วันเกิด"
    maxDate={new Date()} // ไม่เกินวันนี้
    minDate={new Date('1900-01-01')}
    required
    onChange={(date) => {
      const age = calculateAge(date);
      if (age < 20) {
        showError('คุณต้องมีอายุ 20 ปีขึ้นไปเพื่อใช้บริการ');
      }
    }}
  />
  <Button type="submit">ส่งเอกสารยืนยันตัวตน</Button>
</KYCUploadForm>
```

---

## 🔐 Authorization & Security

### การยืนยันตัวตน (KYC)
- **ทุกคน** (Client และ Provider) ต้องผ่านการยืนยันตัวตน
- ต้องส่ง **3 เอกสาร**: บัตรประชาชน + ใบรับรองสุขภาพ + รูป Selfie
- Admin จะตรวจสอบด้วยตนเองว่า **ใบหน้าในบัตร** ตรงกับ **รูป Selfie** หรือไม่
- ต้องมีอายุ **20 ปีขึ้นไป**
- ข้อมูลส่วนบุคคลเก็บเป็นความลับ เข้าถึงได้เฉพาะ Admin

### สิทธิ์การใช้งาน
- **Bookings**: Client สามารถจอง, Provider สามารถอนุมัติ/ปฏิเสธ
- **Reviews**: เฉพาะ Client ที่ใช้บริการเสร็จแล้วถึงจะรีวิวได้
- **Packages**: เฉพาะ Provider ถึงจะสร้าง/แก้ไขแพ็คเกจได้
- **Favorites**: เฉพาะ Client ถึงจะมีรายการโปรดได้

### ระดับความปลอดภัย
1. **JWT Authentication** - ทุก API ต้องมี token
2. **Email Verification** - ยืนยัน email ก่อนใช้งาน
3. **KYC Verification** - ยืนยันตัวตนด้วย 3 เอกสาร
4. **Face Matching** - Admin เปรียบเทียบใบหน้ากับบัตรประชาชน
5. **Age Verification** - ต้องมีอายุ 20+ ปี
6. **Manual Review** - Admin ตรวจสอบทุกคำขอด้วยตนเอง

---

## 👮 Admin - KYC Review Tools

### GET `/admin/pending-users` - ดูรายการรอการอนุมัติ

```typescript
const getPendingUsers = async () => {
  const response = await api.get('/admin/pending-users');
  return response.data; // User[] with age field
};

interface PendingUser {
  user_id: number;
  username: string;
  email: string;
  age: number | null;              // อายุที่คำนวณจาก birth_date
  registration_date: string;
  verification_status: "pending";
  // ... other fields
}
```

### GET `/admin/kyc-details/:userId` - ดูเอกสาร KYC

```typescript
const getKycDetails = async (userId: number) => {
  const response = await api.get(`/admin/kyc-details/${userId}`);
  return response.data;
};

interface KycDetails {
  verification_id: number;
  user_id: number;
  national_id_url: string;    // Key สำหรับดาวน์โหลดบัตรประชาชน
  health_cert_url: string;    // Key สำหรับดาวน์โหลดใบรับรองสุขภาพ
  face_scan_url: string;      // Key สำหรับดาวน์โหลดรูป Selfie
  submitted_at: string;
}
```

### GET `/admin/kyc-file-url?key=xxx` - ดาวน์โหลดไฟล์ KYC

```typescript
const getKycFileUrl = async (fileKey: string) => {
  const response = await api.get('/admin/kyc-file-url', {
    params: { key: fileKey }
  });
  return response.data.url; // Signed URL (valid 10 minutes)
};

// ตัวอย่างการใช้งาน
const viewKycDocuments = async (userId: number) => {
  const kyc = await getKycDetails(userId);
  
  // ดาวน์โหลด URLs สำหรับแสดงรูป
  const idCardUrl = await getKycFileUrl(kyc.national_id_url);
  const faceUrl = await getKycFileUrl(kyc.face_scan_url);
  const healthUrl = await getKycFileUrl(kyc.health_cert_url);
  
  // แสดงรูปเพื่อเปรียบเทียบ
  showComparisonView(idCardUrl, faceUrl);
};
```

### POST `/admin/approve/:userId` - อนุมัติการยืนยันตัวตน

```typescript
const approveUser = async (userId: number) => {
  const response = await api.post(`/admin/approve/${userId}`);
  return response.data;
};
```

### POST `/admin/reject/:userId` - ปฏิเสธการยืนยันตัวตน

```typescript
const rejectUser = async (userId: number, reason?: string) => {
  const response = await api.post(`/admin/reject/${userId}`, { reason });
  return response.data;
};
```

### 🔍 Admin Review Checklist

**ขั้นตอนการตรวจสอบ:**
1. ✅ เปิดรูปบัตรประชาชนและรูป Selfie เคียงข้างกัน
2. ✅ ตรวจสอบว่าใบหน้าเป็นบุคคลเดียวกันหรือไม่
3. ✅ ตรวจสอบว่าบัตรไม่หมดอายุ
4. ✅ ตรวจสอบอายุจากบัตรตรงกับที่ระบบคำนวณหรือไม่
5. ✅ ตรวจสอบใบรับรองสุขภาพว่าถูกต้องและไม่เกิน 6 เดือน
6. ✅ หากพบความผิดปกติ → ปฏิเสธพร้อมระบุเหตุผล
7. ✅ หากทุกอย่างถูกต้อง → อนุมัติ

**เหตุผลที่ควรปฏิเสธ:**
- ❌ ใบหน้าในบัตรไม่ตรงกับรูป Selfie
- ❌ บัตรประชาชนหมดอายุ
- ❌ รูปภาพเบลอหรือไม่ชัดเจน
- ❌ อายุไม่ถึง 20 ปี
- ❌ ใบรับรองสุขภาพหมดอายุหรือไม่ถูกต้อง
- ❌ มีร่องรอยการแก้ไขหรือปลอมแปลง

---

## 🎂 KYC Verification - ยืนยันตัวตนแบบเข้มงวด (ใหม่!)

### 🔐 ระบบป้องกันอาชญากรรม

**ทั้งผู้ใช้บริการและผู้ให้บริการต้องผ่านการยืนยันตัวตน 3 ขั้นตอน:**

1. **บัตรประชาชน** - ถ่ายรูปบัตรประชาชนที่ชัดเจน
2. **ใบรับรองสุขภาพ** - เอกสารแสดงสุขภาพจากโรงพยาบาล
3. **รูปถ่ายใบหน้า (Selfie)** - ถ่ายใบหน้าชัดเจนเพื่อเปรียบเทียบกับบัตรประชาชน

### POST `/verification/submit` - ส่งเอกสาร KYC พร้อมวันเกิด

```typescript
const submitVerification = async (data: {
  national_id_key: string;    // รูปบัตรประชาชน (ต้องเห็นหน้าชัดเจน)
  health_cert_key: string;    // ใบรับรองสุขภาพ
  face_scan_key: string;      // รูป Selfie (ถ่ายหน้าตรง)
  birth_date: string;         // "YYYY-MM-DD" เช่น "2000-05-15"
}) => {
  const response = await api.post('/verification/submit', data);
  return response.data;
};

// ตัวอย่าง
try {
  await submitVerification({
    national_id_key: "kyc/123/national_id_abc.jpg",
    health_cert_key: "kyc/123/health_cert_xyz.jpg",
    face_scan_key: "kyc/123/face_scan_def.jpg",  // Selfie เพื่อเปรียบเทียบ
    birth_date: "2003-11-13" // ต้องอายุ 20+ ปี
  });
} catch (error) {
  if (error.response?.status === 403) {
    // อายุไม่ถึง 20 ปี
    console.log(error.response.data.error);
    console.log(error.response.data.age);
  }
}
```

### 📸 คำแนะนำในการถ่ายรูป

**บัตรประชาชน:**
- ✅ ถ่ายในที่แสงสว่างพอดี ไม่สะท้อน
- ✅ ให้เห็นรายละเอียดทั้งหมดชัดเจน
- ✅ บัตรต้องไม่หมดอายุ
- ❌ ห้ามใช้รูปที่เบลอหรือมีแสงสะท้อน

**รูปใบหน้า (Selfie):**
- ✅ ถ่ายในที่แสงสว่าง หน้าตรงกล้อง
- ✅ ไม่สวมหมวก แว่นตาดำ หน้ากาก
- ✅ ใบหน้าต้องชัดเจน ไม่มีวัตถุบัง
- ✅ ควรมีสีหน้าเป็นธรรมชาติ
- ❌ ห้ามใช้รูปที่แต่งหน้าจัด หรือใช้ Filter

**ใบรับรองสุขภาพ:**
- ✅ ออกโดยโรงพยาบาลที่ได้รับการรับรอง
- ✅ ออกไม่เกิน 6 เดือน
- ✅ มีตราประทับและลายเซ็นแพทย์

### 🛡️ การตรวจสอบโดย Admin

Admin จะตรวจสอบ:
1. **ใบหน้าในบัตรประชาชน** vs **รูป Selfie** - ต้องเป็นบุคคลเดียวกัน
2. **อายุจากบัตรประชาชน** - ตรงกับที่ระบุหรือไม่
3. **ความถูกต้องของเอกสาร** - ไม่มีการปลอมแปลง
4. **ใบรับรองสุขภาพ** - ถูกต้องตามกฎหมาย

### ⚠️ หมายเหตุสำคัญ

- **ทั้ง Client และ Provider ต้องยืนยันตัวตน** - ไม่มีข้อยกเว้น (ยกเว้น GOD tier)
- **ระบบจะคำนวณอายุ** จาก `birth_date` ที่ส่งมา
- **อายุต้อง 20 ปีขึ้นไป** - มิฉะนั้นจะได้ **403 Forbidden**
- **ข้อมูลจะถูกเก็บเป็นความลับ** - เข้าถึงได้เฉพาะ Admin
- **Admin จะตรวจสอบทุกคำขอด้วยตนเอง** - ป้องกันการปลอมแปลง
- **หากถูกปฏิเสธ** - สามารถส่งเอกสารใหม่ได้อีกครั้ง

---

## 🚨 Error Handling

```typescript
try {
  await createBooking(data);
} catch (error) {
  if (error.response?.status === 404) {
    // Package not found
  } else if (error.response?.status === 409) {
    // Time slot conflict
  } else {
    // General error
  }
}
```

---

## 📊 Database Changes

ตารางใหม่ที่เพิ่มเข้ามา:
- `service_packages` - แพ็คเกจบริการ
- `bookings` - การจอง
- `reviews` - รีวิว
- `provider_availability` - ช่วงเวลาว่าง
- `favorites` - รายการโปรด

คอลัมน์ใหม่ใน `user_profiles`:
- `age`, `height`, `weight`
- `ethnicity`, `languages`
- `working_hours`, `is_available`, `service_type`

---

## ✅ สิ่งที่ต้องทำต่อ (Optional)

1. **Real-time Notifications** - แจ้งเตือนการจองใหม่
2. **Chat System** - แชทระหว่าง Client & Provider
3. **Payment Integration** - ชำระเงินผ่านระบบ
4. **Calendar View** - แสดงช่วงเวลาว่างแบบ calendar
5. **Image Compression** - optimize รูปภาพก่อนแสดง

---

## 🎉 Happy Coding!

ถ้ามีคำถามหรือพบปัญหา ติดต่อทีม Backend ได้เลยครับ!
