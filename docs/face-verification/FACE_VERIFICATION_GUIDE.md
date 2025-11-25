# 🎭 Face Verification System - API Documentation

## 📋 Overview

ระบบยืนยันใบหน้าสำหรับ Provider KYC ที่ใช้ Face Recognition และ Liveness Detection เพื่อป้องกันการใช้รูปถ่ายปลอม

### Features
- ✅ Face Matching: เปรียบเทียบ selfie กับรูปบัตรประชาชน
- ✅ Liveness Detection: ตรวจสอบว่าเป็นคนจริง (ไม่ใช่รูปถ่าย)
- ✅ Confidence Score: แสดงค่าความแม่นยำเป็น %
- ✅ Admin Review: Admin สามารถอนุมัติ/ปฏิเสธเพิ่มเติม
- ⚠️ Retry System: Provider สามารถลองใหม่ได้ถ้าไม่ผ่าน

---

## 🔐 Authentication

ทุก endpoint ต้องใช้ JWT token:
```
Authorization: Bearer <your_token>
```

---

## 📤 Provider APIs

### 1. Submit Face Verification

**Endpoint:** `POST /provider/face-verification`

**Description:** Provider อัปโหลด selfie เพื่อทำ face matching กับรูปเอกสารตัวตน (บัตรประชาชนหรือพาสปอร์ต)

**Request Body:**
```json
{
  "selfie_url": "https://storage.googleapis.com/.../selfie.jpg",
  "liveness_video_url": "https://storage.googleapis.com/.../liveness.mp4",  // Optional
  "document_id": 123,  // ID ของเอกสารที่อัปโหลดไว้แล้วใน provider_documents
  "document_type": "national_id"  // "national_id" (บัตรประชาชนไทย) หรือ "passport" (พาสปอร์ต)
}
```

**Document Types:**
- `"national_id"` - บัตรประชาชนไทย (Thai National ID Card)
- `"passport"` - พาสปอร์ต (Foreign Passport)

**Example for Thai Provider:**
```json
{
  "selfie_url": "https://storage.googleapis.com/skillmatch/selfies/user123.jpg",
  "document_id": 456,
  "document_type": "national_id"
}
```

**Example for Foreign Provider:**
```json
{
  "selfie_url": "https://storage.googleapis.com/skillmatch/selfies/user789.jpg",
  "liveness_video_url": "https://storage.googleapis.com/skillmatch/liveness/user789.mp4",
  "document_id": 789,
  "document_type": "passport"
}
```

**Response (201 Created):**
```json
{
  "message": "Face verification submitted successfully",
  "verification_id": 456,
  "status": "pending",
  "next_step": "Admin will review your face verification"
}
```

**Flow:**
1. Provider อัปโหลดเอกสารตัวตนก่อน (ผ่าน `POST /provider/documents`)
   - บัตรประชาชนไทย (`document_type: "national_id"`) สำหรับคนไทย
   - พาสปอร์ต (`document_type: "passport"`) สำหรับชาวต่างชาติ
2. Provider ถ่าย selfie หรืออัดวิดีโอ liveness check
3. อัปโหลดไฟล์ไปยัง GCS (Google Cloud Storage)
4. เรียก API นี้พร้อม URL ของไฟล์ + document_id + document_type
5. ระบบจะบันทึกและรอ Admin ตรวจสอบ

---

### 2. Get My Face Verification Status

**Endpoint:** `GET /provider/face-verification`

**Description:** ตรวจสอบสถานะ face verification ล่าสุดของตัวเอง

**Response (200 OK):**
```json
{
  "verification_id": 456,
  "user_id": 123,
  "selfie_url": "https://storage.googleapis.com/.../selfie.jpg",
  "liveness_video_url": "https://storage.googleapis.com/.../liveness.mp4",
  "match_confidence": 85.5,  // % ความแม่นยำ (0-100)
  "is_match": true,  // ตรงกับบัตรประชาชนหรือไม่
  "national_id_photo_url": "https://storage.googleapis.com/.../id_card.jpg",
  "liveness_passed": true,  // ผ่าน liveness detection หรือไม่
  "liveness_confidence": 92.3,  // % ความมั่นใจว่าเป็นคนจริง
  "verification_status": "approved",  // "pending", "approved", "rejected", "needs_retry"
  "api_provider": "mock_api",  // "aws_rekognition", "azure_face", etc.
  "created_at": "2025-11-21T10:00:00Z",
  "verified_at": "2025-11-21T10:30:00Z",
  "verified_by": 1,  // Admin user_id
  "rejection_reason": null,
  "retry_count": 0,
  "document_type": "national_id",  // "national_id" หรือ "passport"
  "document_id": 123  // ID ของเอกสารที่ใช้ยืนยัน
}
```

**Response (404 Not Found):**
```json
{
  "error": "No face verification found"
}
```

---

## 👨‍💼 Admin APIs

### 3. List All Face Verifications

**Endpoint:** `GET /admin/face-verifications?status=pending`

**Description:** Admin ดู face verifications ทั้งหมดตามสถานะ

**Query Parameters:**
- `status` (optional): `pending`, `approved`, `rejected`, `needs_retry` (default: `pending`)

**Response (200 OK):**
```json
{
  "verifications": [
    {
      "verification_id": 456,
      "user_id": 123,
      "username": "provider123",
      "email": "provider@example.com",
      "selfie_url": "https://storage.googleapis.com/.../selfie.jpg",
      "national_id_photo_url": "https://storage.googleapis.com/.../id_card.jpg",
      "match_confidence": 85.5,
      "is_match": true,
      "liveness_passed": true,
      "liveness_confidence": 92.3,
      "verification_status": "pending",
      "created_at": "2025-11-21T10:00:00Z",
      "retry_count": 0
    }
  ],
  "total": 1
}
```

---

### 4. Review Face Verification (Approve/Reject)

**Endpoint:** `PATCH /admin/face-verification/:verificationId`

**Description:** Admin อนุมัติ หรือ ปฏิเสธ face verification

**Request Body:**
```json
{
  "action": "approve",  // "approve", "reject", "retry"
  "rejection_reason": "รูปไม่ชัดเจน กรุณาถ่ายใหม่",  // Required if action = "reject"
  "match_confidence": 85.5,  // Optional: Admin กำหนดเอง
  "is_match": true  // Optional: Admin กำหนดเอง
}
```

**Response (200 OK):**
```json
{
  "message": "Face verification approved successfully",
  "status": "approved"
}
```

**Effect:**
- ถ้า `action = "approve"` → `verification_status = "approved"` และ `users.face_verified = true`
- ถ้า `action = "reject"` → `verification_status = "rejected"` และบันทึก `rejection_reason`
- ถ้า `action = "retry"` → `verification_status = "needs_retry"` และ `retry_count++`

---

### 5. Trigger Face Matching API (Manual)

**Endpoint:** `POST /admin/face-verification/:verificationId/trigger-matching`

**Description:** เรียก Face Matching API แบบ manual (สำหรับทดสอบหรือ re-process)

**Response (200 OK):**
```json
{
  "message": "Face matching completed",
  "match_confidence": 85.5,
  "is_match": true,
  "liveness_passed": true,
  "liveness_confidence": 92.3
}
```

**Note:** ตอนนี้ใช้ Mock API อยู่ เมื่อเชื่อมต่อ AWS Rekognition หรือ Azure Face API จะได้ผลลัพธ์จริง

---

## 🔄 Provider Registration Flow with Face Verification

### Complete Flow:

```
1. Register as Provider
   POST /register/provider
   ↓
2. Upload National ID Document
   POST /provider/documents (document_type: "national_id")
   ↓
3. Upload Health Certificate
   POST /provider/documents (document_type: "health_certificate")
   ↓
4. Submit Face Verification  ← NEW STEP
   POST /provider/face-verification
   ↓
5. Admin Reviews Documents
   PATCH /admin/verify-document/:documentId
   ↓
6. Admin Reviews Face Verification  ← NEW STEP
   PATCH /admin/face-verification/:verificationId
   ↓
7. Approve Provider (if all pass)
   PATCH /admin/approve-provider/:userId
   ↓
8. Provider can start creating packages and accepting bookings
```

### Verification Requirements:

Provider จะถูก approve เมื่อ:
- ✅ National ID document = approved
- ✅ Health Certificate document = approved
- ✅ Face Verification = approved (NEW)
- ✅ Admin manually approves provider

---

## 🎨 Frontend Implementation Example

### React Component: Face Verification Upload

```tsx
import { useState } from 'react';
import Webcam from 'react-webcam';

export function FaceVerificationUpload() {
  const [selfieURL, setSelfieURL] = useState<string | null>(null);
  const [status, setStatus] = useState<string>('');
  const webcamRef = useRef<Webcam>(null);

  // ถ่าย Selfie
  const captureSelfie = async () => {
    const imageSrc = webcamRef.current?.getScreenshot();
    if (!imageSrc) return;

    // อัปโหลดไป GCS (ใช้ signed URL pattern)
    const uploadedURL = await uploadToGCS(imageSrc);
    setSelfieURL(uploadedURL);
  };

  // ส่ง Face Verification
  const submitFaceVerification = async () => {
    const token = localStorage.getItem('auth_token');
    
    const response = await fetch('http://localhost:8080/provider/face-verification', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        selfie_url: selfieURL,
        national_id_doc_id: 123  // ID ของเอกสารบัตรประชาชน
      })
    });

    const data = await response.json();
    setStatus(data.status);
    alert(data.message);
  };

  return (
    <div>
      <h2>ยืนยันใบหน้า (Face Verification)</h2>
      
      {/* Webcam */}
      <Webcam
        ref={webcamRef}
        screenshotFormat="image/jpeg"
        width={640}
        height={480}
      />
      
      <button onClick={captureSelfie}>ถ่ายภาพ Selfie</button>
      
      {selfieURL && (
        <>
          <img src={selfieURL} alt="Selfie Preview" />
          <button onClick={submitFaceVerification}>
            ส่งยืนยันใบหน้า
          </button>
        </>
      )}
      
      {status && <p>Status: {status}</p>}
    </div>
  );
}
```

### Check Verification Status

```tsx
useEffect(() => {
  const checkStatus = async () => {
    const token = localStorage.getItem('auth_token');
    
    const response = await fetch('http://localhost:8080/provider/face-verification', {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    });

    if (response.ok) {
      const data = await response.json();
      console.log('Face Verification Status:', data.verification_status);
      console.log('Match Confidence:', data.match_confidence);
      console.log('Liveness Passed:', data.liveness_passed);
    }
  };

  checkStatus();
}, []);
```

---

## 🚀 Integration with Third-Party Services

### Option 1: AWS Rekognition

```go
// ใน face_verification_handlers.go
import "github.com/aws/aws-sdk-go/service/rekognition"

func callAWSRekognition(selfieURL, idPhotoURL string) (float64, bool, error) {
  client := rekognition.New(session.New())
  
  input := &rekognition.CompareFacesInput{
    SourceImage: &rekognition.Image{
      S3Object: &rekognition.S3Object{
        Bucket: aws.String("your-bucket"),
        Name:   aws.String("selfie.jpg"),
      },
    },
    TargetImage: &rekognition.Image{
      S3Object: &rekognition.S3Object{
        Bucket: aws.String("your-bucket"),
        Name:   aws.String("id_card.jpg"),
      },
    },
    SimilarityThreshold: aws.Float64(80.0),
  }
  
  result, err := client.CompareFaces(input)
  if err != nil {
    return 0, false, err
  }
  
  if len(result.FaceMatches) > 0 {
    similarity := *result.FaceMatches[0].Similarity
    return similarity, similarity >= 80.0, nil
  }
  
  return 0, false, nil
}
```

### Option 2: Azure Face API

```go
// ใน face_verification_handlers.go
func callAzureFaceAPI(selfieURL, idPhotoURL string) (float64, bool, error) {
  // POST https://[location].api.cognitive.microsoft.com/face/v1.0/verify
  // Body: { "faceId1": "...", "faceId2": "..." }
  
  // 1. Detect faces
  // 2. Get face IDs
  // 3. Verify
  
  return confidence, isMatch, nil
}
```

---

## 📊 Verification Statuses

| Status | Description | Next Action |
|--------|-------------|-------------|
| `pending` | รอตรวจสอบ | Admin review |
| `approved` | อนุมัติแล้ว | Provider can proceed |
| `rejected` | ปฏิเสธ | Provider can retry |
| `needs_retry` | ต้องลองใหม่ | Provider upload again |

---

## 🔒 Security Considerations

### PDPA Compliance
- เก็บข้อมูล biometric ใน encrypted storage
- ลบรูป selfie หลังจาก verification pass (ตามนโยบาย)
- ขอความยินยอมก่อนเก็บข้อมูล

### Anti-Spoofing
- ✅ Liveness Detection: ป้องกันการใช้รูปถ่าย
- ✅ 3D Face Recognition: ตรวจจับ mask
- ✅ Video Recording: บันทึกกระบวนการยืนยัน

### Data Retention
- Selfie: เก็บไว้ 90 วัน หรือจนกว่า verification pass
- API Responses: เก็บใน `api_response_data` (JSONB) สำหรับ audit trail

---

## ✅ Testing Checklist

### Provider Flow
- [ ] ถ่าย selfie ได้
- [ ] อัปโหลดรูปสำเร็จ
- [ ] Submit face verification ได้
- [ ] เช็คสถานะได้
- [ ] แสดง confidence score
- [ ] Retry ได้เมื่อถูกปฏิเสธ

### Admin Flow
- [ ] ดูรายการ pending verifications
- [ ] เปิดดูรูป selfie และบัตรประชาชนแบบเคียงข้างกัน
- [ ] Approve verification ได้
- [ ] Reject พร้อมระบุเหตุผล
- [ ] Trigger face matching manual ได้

---

## 🐛 Known Issues & TODOs

- [ ] **TODO**: เชื่อมต่อ AWS Rekognition หรือ Azure Face API จริง
- [ ] **TODO**: เพิ่ม Liveness Detection API (ตอนนี้ mock อยู่)
- [ ] **TODO**: Auto-retry mechanism เมื่อ API fail
- [ ] **TODO**: Webhook notification เมื่อ verification approved/rejected
- [ ] **TODO**: Rate limiting สำหรับ face matching API (ป้องกัน spam)

---

## 📞 Support

หากมีปัญหาหรือต้องการความช่วยเหลือ:
- **Backend Developer**: ติดต่อผู้ดูแลระบบ
- **API Keys**: ต้องมี AWS/Azure credentials พร้อมใช้งาน
- **Storage**: ต้องมี GCS bucket สำหรับเก็บรูปภาพ
