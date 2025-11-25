# 🔐 Provider Routes - Authentication-Based Access

## ⚠️ การเปลี่ยนแปลงสำคัญ: ซ่อนข้อมูลผู้ให้บริการจากคนที่ไม่ได้ login

เพื่อความเป็นส่วนตัวและปลอดภัย ระบบแยก API endpoints เป็น 2 ระดับ:

### 📋 Public Endpoints (ไม่ต้อง Login - ข้อมูลจำกัด):

```
GET /provider/:userId/public       → ดู profile แบบจำกัด (ไม่แสดง age, height, service_type)
GET /provider/:userId/photos       → ดูรูปภาพผู้ให้บริการ
GET /packages/:providerId          → ดูแพ็คเกจของผู้ให้บริการ
GET /reviews/:providerId           → ดูรีวิวของผู้ให้บริการ
GET /reviews/stats/:providerId     → ดูสถิติรีวิว
GET /service-categories            → ดูหมวดหมู่บริการทั้งหมด
GET /categories/:category_id/providers → ดูผู้ให้บริการในหมวดหมู่
```

### 🔐 Protected Endpoints (ต้อง Login - ข้อมูลเต็ม):

```
GET /provider/:userId              → ดู profile เต็มรูปแบบ (แสดงทุกอย่างรวม age, height, service_type)
GET /browse/v2                     → Browse ผู้ให้บริการ (with filters)
```

---

## 🔧 วิธีใช้ใน Frontend Code

### 1️⃣ **สำหรับผู้ใช้ที่ไม่ได้ Login (ข้อมูลจำกัด):**

```typescript
// ✅ ดูข้อมูลพื้นฐาน - ไม่แสดง age, height, weight, service_type
const response = await fetch(`http://localhost:8080/provider/${userId}/public`);
const profile = await response.json();
// ได้: username, bio, skills, rating, province
// ไม่ได้: age, height, weight, ethnicity, languages, working_hours, service_type
```

### 2️⃣ **สำหรับผู้ใช้ที่ Login แล้ว (ข้อมูลเต็ม):**

```typescript
// ✅ ดูข้อมูลเต็มรูปแบบ - แสดงทุกอย่าง
const token = localStorage.getItem('auth_token');
const response = await fetch(`http://localhost:8080/provider/${userId}`, {
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  }
});
const profile = await response.json();
// ได้ทุกอย่าง: age, height, weight, service_type, working_hours, languages, etc.
```

---

## 📝 ตัวอย่างการใช้งานใน Frontend

### 1. **ดู Provider Profile (แบบมี Login / ไม่มี Login)**

```typescript
// components/ProviderProfile.tsx
async function fetchProviderProfile(userId: number, isAuthenticated: boolean) {
  try {
    let response;
    
    if (isAuthenticated) {
      // ✅ ผู้ใช้ที่ login - ดูข้อมูลเต็ม
      const token = localStorage.getItem('auth_token');
      response = await fetch(`http://localhost:8080/provider/${userId}`, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });
    } else {
      // ✅ ผู้ใช้ที่ไม่ login - ดูข้อมูลจำกัด
      response = await fetch(`http://localhost:8080/provider/${userId}/public`);
    }
    
    if (!response.ok) {
      throw new Error('Provider not found');
    }
    
    const profile = await response.json();
    return profile;
  } catch (error) {
    console.error('Error fetching provider:', error);
    throw error;
  }
}

// Usage
const { isAuthenticated } = useAuth(); // from your auth context
const profile = await fetchProviderProfile(5, isAuthenticated);

// ถ้า isAuthenticated = false (ไม่ login):
// {
//   "user_id": 5,
//   "username": "maya_massage",
//   "tier_name": "General",
//   "skills": ["Oil Massage", "Body Scrub", "Facial"],
//   "bio": "Certified spa therapist",
//   "rating": 4.5
//   // ไม่มี: age, height, weight, service_type
// }

// ถ้า isAuthenticated = true (login แล้ว):
// {
//   "user_id": 5,
//   "username": "maya_massage",
//   "age": 28,
//   "height": 165,
//   "weight": 52,
//   "service_type": "Incall & Outcall",
//   "working_hours": "10:00-22:00",
//   // ... ข้อมูลครบทุกอย่าง
// }
```

### 2. **ดู Provider Photos (Public)**

```typescript
async function fetchProviderPhotos(userId: number) {
  const response = await fetch(`http://localhost:8080/provider/${userId}/photos`);
  const photos = await response.json();
  return photos;
}

// Usage
const photos = await fetchProviderPhotos(5);
// [
//   {
//     "photo_id": 1,
//     "user_id": 5,
//     "photo_url": "https://...",
//     "sort_order": 1
//   }
// ]
```

### 3. **Browse Providers with Filters (ต้อง Login)**

```typescript
async function browseProviders(filters?: {
  category?: string;
  province?: string;
  min_rating?: number;
  page?: number;
}) {
  // ⚠️ Browse ต้อง login เท่านั้น
  const token = localStorage.getItem('auth_token');
  
  if (!token) {
    throw new Error('Please login to browse providers');
  }
  
  const params = new URLSearchParams();
  if (filters?.category) params.append('category', filters.category);
  if (filters?.province) params.append('province', filters.province);
  if (filters?.min_rating) params.append('min_rating', filters.min_rating.toString());
  if (filters?.page) params.append('page', filters.page.toString());
  
  const response = await fetch(`http://localhost:8080/browse/v2?${params}`, {
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    }
  });
  
  if (response.status === 401) {
    throw new Error('Please login to browse providers');
  }
  
  const providers = await response.json();
  return providers;
}

// Usage
const providers = await browseProviders({
  category: 'massage',
  province: 'Bangkok',
  min_rating: 4
});
```

### 4. **ดู Provider Packages (Public)**

```typescript
async function fetchProviderPackages(providerId: number) {
  const response = await fetch(`http://localhost:8080/packages/${providerId}`);
  const packages = await response.json();
  return packages;
}
```

### 5. **ดู Provider Reviews (Public)**

```typescript
async function fetchProviderReviews(providerId: number, limit = 20, offset = 0) {
  const response = await fetch(
    `http://localhost:8080/reviews/${providerId}?limit=${limit}&offset=${offset}`
  );
  const reviews = await response.json();
  return reviews;
}

async function fetchProviderReviewStats(providerId: number) {
  const response = await fetch(`http://localhost:8080/reviews/stats/${providerId}`);
  const stats = await response.json();
  return stats;
  // {
  //   "average_rating": 4.5,
  //   "total_reviews": 10,
  //   "rating_breakdown": { "5": 6, "4": 3, "3": 1 }
  // }
}
```

---

## 🔐 Endpoints อื่นๆ ที่ต้อง Login

```
GET /provider/:userId              → ดู profile เต็มรูปแบบ (ต้อง login)
GET /browse/v2                     → Browse ผู้ให้บริการ (ต้อง login)
POST /reviews                      → สร้างรีวิว (ต้อง login)
POST /bookings                     → จองบริการ (ต้อง login)
POST /favorites                    → เพิ่มรายการโปรด (ต้อง login)
GET /favorites                     → ดูรายการโปรดของตัวเอง (ต้อง login)
GET /bookings/my                   → ดูการจองของตัวเอง (ต้อง login)
GET /wallet                        → ดู wallet (ต้อง login)
POST /packages                     → สร้างแพ็คเกจ (provider only)
```

### ✅ **สำหรับ Protected Endpoints:**

```typescript
// ✅ ส่ง Authorization header
async function createBooking(data: BookingData) {
  const token = localStorage.getItem('auth_token');
  
  const response = await fetch('http://localhost:8080/bookings', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(data)
  });
  
  return response.json();
}
```

---

## 🎯 React Component Example

```tsx
// pages/ProviderDetailPage.tsx
import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { useAuth } from '../context/AuthContext'; // Your auth context

interface PublicProvider {
  user_id: number;
  username: string;
  tier_name: string;
  bio: string;
  skills: string[];
  average_rating: number;
  review_count: number;
  province: string;
}

interface FullProvider extends PublicProvider {
  age: number;
  height: number;
  weight: number;
  ethnicity: string;
  languages: string[];
  working_hours: string;
  service_type: string;
  address_line1: string;
}

export default function ProviderDetailPage() {
  const { userId } = useParams<{ userId: string }>();
  const { isAuthenticated, token } = useAuth();
  const [provider, setProvider] = useState<PublicProvider | FullProvider | null>(null);
  const [photos, setPhotos] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function loadProvider() {
      try {
        setLoading(true);
        
        let profileRes;
        if (isAuthenticated && token) {
          // ✅ Login แล้ว - ดูข้อมูลเต็ม
          profileRes = await fetch(`http://localhost:8080/provider/${userId}`, {
            headers: {
              'Authorization': `Bearer ${token}`,
              'Content-Type': 'application/json'
            }
          });
        } else {
          // ✅ ไม่ login - ดูข้อมูลจำกัด
          profileRes = await fetch(`http://localhost:8080/provider/${userId}/public`);
        }

        const photosRes = await fetch(`http://localhost:8080/provider/${userId}/photos`);

        if (!profileRes.ok) {
          throw new Error('Provider not found');
        }

        const profileData = await profileRes.json();
        const photosData = await photosRes.json();

        setProvider(profileData);
        setPhotos(photosData);
      } catch (err: any) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    }

    if (userId) {
      loadProvider();
    }
  }, [userId, isAuthenticated, token]);

  if (loading) return <div>Loading...</div>;
  if (error) return <div>Error: {error}</div>;
  if (!provider) return <div>Provider not found</div>;

  const isFullProfile = (p: any): p is FullProvider => 'age' in p;

  return (
    <div className="provider-detail">
      <h1>{provider.username}</h1>
      <p className="tier">{provider.tier_name}</p>
      <p className="bio">{provider.bio}</p>
      
      <div className="skills">
        {provider.skills.map((skill, idx) => (
          <span key={idx} className="skill-badge">{skill}</span>
        ))}
      </div>

      <div className="rating">
        ⭐ {provider.average_rating} ({provider.review_count} reviews)
      </div>

      {/* แสดงข้อมูลเต็มเฉพาะเมื่อ login */}
      {isAuthenticated && isFullProfile(provider) && (
        <div className="detailed-info">
          <h3>Detailed Information (Members Only)</h3>
          <p>Age: {provider.age}</p>
          <p>Height: {provider.height} cm</p>
          <p>Weight: {provider.weight} kg</p>
          <p>Service Type: {provider.service_type}</p>
          <p>Working Hours: {provider.working_hours}</p>
          <p>Languages: {provider.languages.join(', ')}</p>
        </div>
      )}

      {/* แสดงข้อความชักชวนให้ login */}
      {!isAuthenticated && (
        <div className="login-prompt">
          <p>🔒 Login to see more details about this provider</p>
          <button onClick={() => navigate('/login')}>Login Now</button>
        </div>
      )}

      <div className="photos">
        {photos.map(photo => (
          <img key={photo.photo_id} src={photo.photo_url} alt="Provider" />
        ))}
      </div>
    </div>
  );
}
```

---

## 🚨 Important Notes

### 1. **เช็ค Authentication Status ก่อนเรียก API**

```typescript
const { isAuthenticated, token } = useAuth();

if (isAuthenticated && token) {
  // ✅ เรียก authenticated endpoint - ได้ข้อมูลเต็ม
  const response = await fetch(`/provider/${id}`, {
    headers: { 'Authorization': `Bearer ${token}` }
  });
} else {
  // ✅ เรียก public endpoint - ได้ข้อมูลจำกัด
  const response = await fetch(`/provider/${id}/public`);
}

// จัดการ errors
if (response.status === 401) {
  // Token หมดอายุ - ให้ logout
  logout();
  navigate('/login');
}
if (response.status === 404) {
  throw new Error('Provider not found');
}
```

### 2. **SEO-Friendly กับ Public Endpoint**

เนื่องจากมี public endpoint (`/provider/:userId/public`):
- Google bot สามารถ crawl ได้
- ไม่ต้อง JavaScript ก็ดูได้ (ถ้าใช้ SSR)
- Share link บน social media จะแสดง preview ได้
- แต่ไม่เห็นข้อมูลละเอียด (age, height, service_type)

### 3. **Error Handling**

```typescript
// Provider ที่ไม่ verified จะได้ 404
// Provider ที่ถูก block จะได้ 404
// Provider ที่ไม่มีจริงจะได้ 404

try {
  const response = await fetch(`http://localhost:8080/provider/${userId}`);
  
  if (response.status === 404) {
    // แสดงหน้า "Provider not found"
    setError('ไม่พบผู้ให้บริการนี้');
    return;
  }
  
  const data = await response.json();
  setProvider(data);
} catch (error) {
  setError('เกิดข้อผิดพลาดในการโหลดข้อมูล');
}
```

---

## 📊 API Response Examples

### GET /provider/:userId/public (ไม่ต้อง Login)
```json
{
  "user_id": 5,
  "username": "maya_massage",
  "gender_id": 2,
  "tier_name": "General",
  "bio": "Certified spa therapist",
  "location": null,
  "skills": ["Oil Massage", "Body Scrub", "Facial"],
  "profile_image_url": null,
  "google_profile_picture": "https://i.pravatar.cc/300?img=9",
  "is_available": false,
  "average_rating": 0,
  "review_count": 0,
  "province": null,
  "district": null,
  "sub_district": null
}
```

**⚠️ ข้อมูลที่ซ่อนไว้ (ต้อง login ถึงจะเห็น):**
- ❌ Age, Height, Weight (ข้อมูลร่างกาย)
- ❌ Ethnicity (เชื้อชาติ)
- ❌ Languages (ภาษา)
- ❌ WorkingHours (เวลาทำงาน)
- ❌ ServiceType (ประเภทการให้บริการ)
- ❌ AddressLine1, Latitude, Longitude (ที่อยู่เฉพาะเจาะจง)

### GET /provider/:userId (ต้อง Login + Token)
```json
{
  "user_id": 5,
  "username": "maya_massage",
  "gender_id": 2,
  "tier_name": "General",
  "bio": "Certified spa therapist",
  "location": null,
  "skills": ["Oil Massage", "Body Scrub", "Facial"],
  "profile_image_url": null,
  "google_profile_picture": "https://i.pravatar.cc/300?img=9",
  "is_available": false,
  "average_rating": 0,
  "review_count": 0,
  "province": null,
  "district": null,
  "sub_district": null,
  "address_line1": null,
  "latitude": null,
  "longitude": null,
  "age": 28,
  "height": 165,
  "weight": 52,
  "ethnicity": "Thai",
  "languages": ["Thai", "English"],
  "working_hours": "10:00-22:00",
  "service_type": "Incall & Outcall"
}
```

**✅ ได้ข้อมูลเต็มรูปแบบ:**
- ✅ ข้อมูลพื้นฐาน + ข้อมูลละเอียด
- ✅ Age, Height, Weight, Service Type
- ✅ Working Hours, Languages
- ✅ ที่อยู่เฉพาะเจาะจง (สำหรับการจอง)

### GET /browse/v2 (ต้อง Login + Token)
```json
{
  "providers": [
    {
      "user_id": 5,
      "username": "maya_massage",
      "tier_name": "General",
      "average_rating": 4.5,
      "review_count": 10,
      "province": "Bangkok",
      "age": 28,
      "service_type": "Incall & Outcall"
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 20
}
```

---

## ✅ Checklist สำหรับ Frontend Dev

### Phase 1: Update API Calls
- [ ] เพิ่ม logic เช็ค `isAuthenticated` ก่อนเรียก API
- [ ] ใช้ `/provider/:userId/public` สำหรับผู้ใช้ที่ไม่ login
- [ ] ใช้ `/provider/:userId` (with token) สำหรับผู้ใช้ที่ login
- [ ] ส่ง `Authorization: Bearer {token}` สำหรับ `/browse/v2`
- [ ] จัดการ 401 error (token หมดอายุ → redirect to login)

### Phase 2: Update UI Components
- [ ] แสดง "Login to see more details" prompt สำหรับผู้ใช้ที่ไม่ login
- [ ] แสดงข้อมูลละเอียด (age, height, service_type) เฉพาะผู้ใช้ที่ login
- [ ] ซ่อน Browse page จากผู้ใช้ที่ไม่ login (ใช้ ProtectedRoute)
- [ ] เพิ่ม TypeScript interfaces แยก `PublicProvider` และ `FullProvider`

### Phase 3: Testing
- [ ] ทดสอบดู provider profile โดยไม่ login → ควรเห็นข้อมูลจำกัด
- [ ] ทดสอบดู provider profile โดย login แล้ว → ควรเห็นข้อมูลเต็ม
- [ ] ทดสอบ browse โดยไม่ login → ควรได้ 401 หรือ redirect
- [ ] ทดสอบ browse โดย login แล้ว → ควรเห็นรายการผู้ให้บริการ
- [ ] ทดสอบ token หมดอายุ → ควร logout และ redirect to login

### Phase 4: SEO & UX
- [ ] เพิ่ม SEO meta tags สำหรับ provider public pages
- [ ] เพิ่ม loading state สำหรับทั้ง public และ authenticated pages
- [ ] เพิ่ม error handling ที่เหมาะสม (404, 401, 500)

---

**Last Updated:** November 14, 2025, 11:05 AM  
**Backend:** Running on http://localhost:8080  
**Status:** ✅ Authentication-based routes working
**Security:** ✅ Sensitive data hidden from non-members
