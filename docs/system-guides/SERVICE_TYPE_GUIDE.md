# 🏪 Service Type System - ระบบประเภทการให้บริการ

## 📋 ภาพรวม

ระบบ Service Type กำหนดว่า Provider ให้บริการแบบไหน:
- **Incall**: Provider มีโรงแรม/สถานที่ให้บริการ (Client ไปหา)
- **Outcall**: Provider ไปหาลูกค้าถึงที่ (Provider ไปหา Client)

⚠️ **ไม่มี "Both"** - Provider ต้องเลือกอย่างใดอย่างหนึ่ง

---

## 🗄️ Database Schema

### ฟิลด์ที่เพิ่ม

```sql
-- ใน user_profiles table
service_type VARCHAR(20) CHECK (service_type IN ('incall', 'outcall'))
```

| ฟิลด์        | Type    | Nullable | ค่าที่เป็นไปได้    | Description                        |
|--------------|---------|----------|-------------------|------------------------------------|
| service_type | VARCHAR | YES      | 'incall', 'outcall' | ประเภทการให้บริการ                 |

### Migration

```bash
# Run migration
docker exec -i postgres_db psql -U admin -d skillmatch_db < migrations/006_add_service_type.sql
```

---

## 🔍 API Changes

### 1. GET `/browse/v2` - เพิ่มฟิลด์ service_type

**Response:**

```json
{
  "user_id": 10,
  "username": "alice_pro",
  "tier_name": "VIP",
  "province": "กรุงเทพมหานคร",
  "district": "บางรัก",
  "service_type": "incall",
  "average_rating": 4.8,
  "min_price": 1500
}
```

### 2. GET `/browse/v2?service_type=incall` - ฟิลเตอร์ตาม service_type

**Query Parameters:**

```typescript
interface BrowseFilters {
  service_type?: "incall" | "outcall"; // ฟิลเตอร์ตามประเภท
  // ... ฟิลเตอร์อื่นๆ
}

// ตัวอย่าง
GET /browse/v2?service_type=incall     // Provider ที่มีสถานที่
GET /browse/v2?service_type=outcall    // Provider ที่ไปหาลูกค้า
GET /browse/v2                         // ทั้งหมด (ไม่กรอง)
```

### 3. PUT `/profile/me` - อัพเดท service_type

**Request:**

```json
{
  "service_type": "incall"
}
```

**Validation:**

- ✅ `"incall"` - ผ่าน
- ✅ `"outcall"` - ผ่าน
- ❌ `"both"` - Error 400
- ❌ `"other"` - Error 400
- ✅ `null` - ผ่าน (optional field)

**Error Response:**

```json
{
  "error": "Invalid service_type",
  "message": "service_type must be 'incall' or 'outcall'"
}
```

### 4. POST `/bookings` - Validation ตาม service_type

#### กรณี Outcall (Provider ไปหาลูกค้า)

**ต้องมี `location` จาก Client:**

```json
{
  "provider_id": 10,
  "package_id": 5,
  "booking_date": "2025-11-20",
  "start_time": "14:00",
  "location": "123 ถนนสุขุมวิท แขวงคลองเตย เขตคลองเตย กรุงเทพฯ 10110"
}
```

**ถ้าไม่มี location → Error:**

```json
{
  "error": "Location is required for outcall services",
  "message": "ผู้ให้บริการรายนี้ให้บริการแบบไปหาลูกค้า กรุณาระบุที่อยู่ของคุณ"
}
```

#### กรณี Incall (Client ไปหา Provider)

**ไม่ต้องมี `location`:**

```json
{
  "provider_id": 10,
  "package_id": 5,
  "booking_date": "2025-11-20",
  "start_time": "14:00"
  // location ไม่จำเป็น
}
```

Provider's location จะแสดงหลัง booking confirmed

---

## 🎯 Business Logic

### 1. Provider ตั้งค่า Service Type

```typescript
// Provider เลือก service type ตอน setup profile
await api.put('/profile/me', {
  service_type: "incall", // หรือ "outcall"
  province: "กรุงเทพมหานคร",
  district: "บางรัก",
  latitude: 13.7278,
  longitude: 100.5318
});
```

### 2. Client ค้นหาตาม Service Type

```typescript
// ค้นหา Provider ที่มีสถานที่
const incallProviders = await api.get('/browse/v2?service_type=incall');

// ค้นหา Provider ที่ไปหาลูกค้า
const outcallProviders = await api.get('/browse/v2?service_type=outcall');
```

### 3. Client จองบริการ

#### Incall (ไปหา Provider)

```typescript
await api.post('/bookings', {
  provider_id: 10,
  package_id: 5,
  booking_date: "2025-11-20",
  start_time: "14:00"
  // ไม่ต้องระบุ location
});

// หลัง confirmed → จะได้เห็นที่อยู่ของ Provider
```

#### Outcall (Provider มาหา)

```typescript
await api.post('/bookings', {
  provider_id: 10,
  package_id: 5,
  booking_date: "2025-11-20",
  start_time: "14:00",
  location: "บ้านเลขที่ 123 ถนนสุขุมวิท..." // ต้องระบุ
});
```

---

## 🔒 Privacy & Location Visibility

### ก่อน Booking Confirmed

| Service Type | ที่อยู่ที่แสดง                     |
|--------------|------------------------------------|
| Incall       | จังหวัด, เขต, แขวง (ไม่แสดงที่แน่นอน) |
| Outcall      | จังหวัด, เขต, แขวง (service area)      |

### หลัง Booking Confirmed

| Service Type | ที่อยู่ที่แสดง                     |
|--------------|------------------------------------|
| Incall       | ที่อยู่เต็มของ Provider (บ้านเลขที่, โรงแรม) |
| Outcall      | ที่อยู่ที่ Client ระบุ               |

---

## 🎨 Frontend UI Examples

### 1. Service Type Filter

```tsx
const ServiceTypeFilter = ({ value, onChange }: FilterProps) => {
  return (
    <div className="service-type-filter">
      <label>ประเภทบริการ</label>
      <select value={value || ''} onChange={(e) => onChange(e.target.value)}>
        <option value="">ทั้งหมด</option>
        <option value="incall">🏨 Incall - มีสถานที่</option>
        <option value="outcall">🚗 Outcall - ไปหาลูกค้า</option>
      </select>
    </div>
  );
};
```

### 2. Service Type Badge

```tsx
const ServiceTypeBadge = ({ type }: { type: "incall" | "outcall" | null }) => {
  if (!type) return null;
  
  return (
    <span className={`badge badge-${type}`}>
      {type === 'incall' ? '🏨 Incall' : '🚗 Outcall'}
    </span>
  );
};
```

### 3. Provider Card with Service Type

```tsx
const ProviderCard = ({ provider }: { provider: BrowsableUser }) => {
  return (
    <div className="provider-card">
      <img src={provider.profile_image_url} />
      <h3>{provider.username}</h3>
      
      {/* Service Type */}
      <ServiceTypeBadge type={provider.service_type} />
      
      {/* Location */}
      <p>📍 {provider.district}, {provider.province}</p>
      
      {/* Rating */}
      <p>⭐ {provider.average_rating.toFixed(1)}</p>
    </div>
  );
};
```

### 4. Booking Form - Location Input

```tsx
const BookingForm = ({ provider }: { provider: Provider }) => {
  const [location, setLocation] = useState('');
  const requiresLocation = provider.service_type === 'outcall';
  
  return (
    <form onSubmit={handleSubmit}>
      <input type="date" name="booking_date" required />
      <input type="time" name="start_time" required />
      
      {/* แสดง location input เฉพาะ outcall */}
      {requiresLocation && (
        <div className="location-input">
          <label>ที่อยู่ของคุณ *</label>
          <textarea
            value={location}
            onChange={(e) => setLocation(e.target.value)}
            placeholder="บ้านเลขที่, ถนน, แขวง, เขต, จังหวัด, รหัสไปรษณีย์"
            required
          />
          <small className="text-warning">
            ⚠️ ผู้ให้บริการจะไปหาคุณถึงที่ กรุณาระบุที่อยู่ที่ถูกต้อง
          </small>
        </div>
      )}
      
      {!requiresLocation && (
        <div className="info-box">
          ℹ️ คุณจะได้รับที่อยู่ของผู้ให้บริการหลังการจองได้รับการยืนยัน
        </div>
      )}
      
      <button type="submit">จองเลย</button>
    </form>
  );
};
```

---

## ✅ Testing

### Database

```sql
-- ตรวจสอบ service_type ถูกเพิ่มแล้ว
\d user_profiles

-- ทดสอบ constraint
UPDATE user_profiles SET service_type = 'incall' WHERE user_id = 1; -- ✅
UPDATE user_profiles SET service_type = 'outcall' WHERE user_id = 2; -- ✅
UPDATE user_profiles SET service_type = 'both' WHERE user_id = 3; -- ❌ Error
UPDATE user_profiles SET service_type = NULL WHERE user_id = 4; -- ✅
```

### API Endpoints

```bash
# 1. ทดสอบ browse filter
curl "http://localhost:8080/browse/v2?service_type=incall"
curl "http://localhost:8080/browse/v2?service_type=outcall"

# 2. ทดสอบ update profile
curl -X PUT http://localhost:8080/profile/me \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"service_type": "incall"}'

# 3. ทดสอบ validation
curl -X PUT http://localhost:8080/profile/me \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"service_type": "both"}'
# Expected: 400 Bad Request

# 4. ทดสอบ booking outcall without location
curl -X POST http://localhost:8080/bookings \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "provider_id": 10,
    "package_id": 5,
    "booking_date": "2025-11-20",
    "start_time": "14:00"
  }'
# Expected: 400 if provider is outcall

# 5. ทดสอบ booking outcall with location
curl -X POST http://localhost:8080/bookings \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "provider_id": 10,
    "package_id": 5,
    "booking_date": "2025-11-20",
    "start_time": "14:00",
    "location": "123 ถนนสุขุมวิท แขวงคลองเตย"
  }'
# Expected: 201 Created
```

---

## 📊 Data Examples

### Provider Profile with Incall

```json
{
  "user_id": 10,
  "username": "alice_hotel",
  "service_type": "incall",
  "province": "กรุงเทพมหานคร",
  "district": "บางรัก",
  "sub_district": "สีลม",
  "latitude": 13.7278,
  "longitude": 100.5318,
  "address_line1": "โรงแรม XYZ ชั้น 5 ห้อง 501" // แสดงหลัง confirmed
}
```

### Provider Profile with Outcall

```json
{
  "user_id": 15,
  "username": "bob_mobile",
  "service_type": "outcall",
  "province": "กรุงเทพมหานคร",
  "district": "วัฒนา",
  "sub_district": "คลองเตย",
  "latitude": 13.7307,
  "longitude": 100.5418
  // ไม่มี address_line1 เพราะไปหาลูกค้า
}
```

---

## 🚨 Common Errors

### Error 1: Invalid service_type

```json
{
  "error": "Invalid service_type",
  "message": "service_type must be 'incall' or 'outcall'"
}
```

**สาเหตุ:** ส่งค่าที่ไม่ใช่ 'incall' หรือ 'outcall' (เช่น 'both')

### Error 2: Location required for outcall

```json
{
  "error": "Location is required for outcall services",
  "message": "ผู้ให้บริการรายนี้ให้บริการแบบไปหาลูกค้า กรุณาระบุที่อยู่ของคุณ"
}
```

**สาเหตุ:** จอง outcall provider แต่ไม่ได้ระบุ location

---

## 🎯 Best Practices

### สำหรับ Provider

1. **เลือก Service Type ที่เหมาะสม**
   - มีโรงแรม/ห้อง → Incall
   - ต้องการไปหาลูกค้า → Outcall

2. **Incall Provider**
   - ระบุที่อยู่แม่นยำ (โรงแรม, ชั้น, ห้อง)
   - อัพเดท availability ตามห้องว่าง

3. **Outcall Provider**
   - ระบุพื้นที่บริการ (จังหวัด, เขต)
   - คิดค่าเดินทางถ้าไกล

### สำหรับ Client

1. **Incall Booking**
   - เช็คที่อยู่หลัง confirmed
   - ถามรายละเอียดก่อนไป

2. **Outcall Booking**
   - ระบุที่อยู่ชัดเจน ครบถ้วน
   - เช็คว่า Provider รับพื้นที่หรือไม่

---

## 📝 Notes

1. **ไม่มี "Both" option**
   - ถ้าต้องการรองรับทั้งสองแบบ → สร้าง 2 accounts หรือคุยกับ client ก่อน

2. **Location Privacy**
   - Incall: แสดงที่อยู่เต็มหลัง confirmed
   - Outcall: ใช้ที่อยู่ที่ client ระบุ

3. **Service Area (Outcall)**
   - พิจารณาเพิ่มระบบระบุรัศมีพื้นที่บริการในอนาคต
   - ตอนนี้ใช้ province/district filter

4. **Future Enhancements**
   - Service area radius (รัศมี 5 กม., 10 กม.)
   - Travel fee calculator (คำนวณค่าเดินทาง)
   - Multi-location support (หลายสถานที่)

---

## 🆘 Support

หากมีคำถามเกี่ยวกับ Service Type System ติดต่อได้เลย!
