# 📁 Face Verification Documentation

เอกสารทั้งหมดเกี่ยวกับระบบ Face Verification with Passport Support

---

## 📚 เอกสารในโฟลเดอร์นี้

### 1. **FACE_VERIFICATION_GUIDE.md**
**สำหรับ:** Backend Developers & Frontend Developers  
**เนื้อหา:** คู่มือ API ครบถ้วนสำหรับระบบ Face Verification  
**ประกอบด้วย:**
- API Endpoint specifications (5 endpoints)
- Request/Response examples
- Face matching & Liveness detection concepts
- AWS Rekognition integration code
- Azure Face API integration code
- Security considerations (PDPA compliance)

**ใช้เมื่อ:** ต้องการดู API documentation หรือ integrate face matching API

---

### 2. **FACE_VERIFICATION_IMPLEMENTATION_SUMMARY.md**
**สำหรับ:** Backend Developers & DevOps  
**เนื้อหา:** สรุปการ implement ระบบ Face Verification  
**ประกอบด้วย:**
- Migration 020 details (database schema)
- Handler implementations (6 functions)
- Route registrations (5 endpoints)
- Known limitations & TODOs
- Performance metrics
- Deployment status

**ใช้เมื่อ:** ต้องการเข้าใจ architecture หรือ troubleshoot ระบบ

---

### 3. **PASSPORT_SUPPORT_IMPLEMENTATION_SUMMARY.md**
**สำหรับ:** Backend Developers & Project Managers  
**เนื้อหา:** สรุปการเพิ่มการรองรับพาสปอร์ต  
**ประกอบด้วย:**
- Migration 021 details (document_type, document_id columns)
- Breaking changes documentation
- Updated API request/response formats
- Database verification results
- Security considerations for passport data
- PDPA compliance notes

**ใช้เมื่อ:** ต้องการดูรายละเอียดการ update ล่าสุด (Nov 21, 2025)

---

### 4. **FRONTEND_FACE_VERIFICATION_PASSPORT_GUIDE.md** ⭐ **RECOMMENDED FOR FRONTEND**
**สำหรับ:** Frontend Developers (React/TypeScript)  
**เนื้อหา:** คู่มือครบถ้วนสำหรับ implement UI/UX  
**ประกอบด้วย:**
- **Breaking Changes Alert** - ต้องเปลี่ยน API calls
- Complete API documentation with TypeScript interfaces
- **3 Ready-to-use React Components:**
  - FaceVerificationForm (with Webcam)
  - VerificationStatusCard
  - AdminReviewPanel
- Frontend validation functions
- Error handling patterns
- Testing checklist (30+ items)
- UI/UX recommendations with CSS
- Complete TypeScript type definitions

**ใช้เมื่อ:** เริ่มทำ Frontend integration (เอกสารหลัก)

---

## 🚀 Quick Start

### สำหรับ Frontend Developers

1. **อ่านเอกสารหลัก:**
   ```bash
   FRONTEND_FACE_VERIFICATION_PASSPORT_GUIDE.md
   ```

2. **ดู Breaking Changes:**
   - ❌ เดิม: `national_id_doc_id`
   - ✅ ใหม่: `document_id` + `document_type`

3. **Copy React Components:**
   - เอา 3 components จากเอกสารไปใช้เลย
   - มี TypeScript types พร้อมแล้ว

4. **ทดสอบ:**
   - ทดสอบทั้ง `document_type: "national_id"` และ `"passport"`

### สำหรับ Backend Developers

1. **ดู Implementation Summary:**
   ```bash
   FACE_VERIFICATION_IMPLEMENTATION_SUMMARY.md
   PASSPORT_SUPPORT_IMPLEMENTATION_SUMMARY.md
   ```

2. **ตรวจสอบ Migration:**
   ```sql
   -- Migration 020: Face Verification System
   -- Migration 021: Passport Support
   ```

3. **ดู API Guide:**
   ```bash
   FACE_VERIFICATION_GUIDE.md
   ```

4. **TODO Items:**
   - Replace mock face matching API with AWS/Azure
   - Add real liveness detection
   - Create GCS upload endpoint for selfies

---

## 📊 System Overview

```
┌─────────────────────────────────────────────────────────┐
│                Face Verification System                  │
├─────────────────────────────────────────────────────────┤
│                                                           │
│  Provider Flow:                                          │
│  1. เลือกเอกสาร (บัตรประชาชน/พาสปอร์ต)                │
│  2. อัปโหลดเอกสาร → provider_documents                  │
│  3. ถ่ายเซลฟี่ → GCS                                     │
│  4. POST /provider/face-verification                     │
│     {                                                     │
│       document_id: 123,                                  │
│       document_type: "national_id" | "passport",        │
│       selfie_url: "https://..."                          │
│     }                                                     │
│  5. GET /provider/face-verification (check status)       │
│                                                           │
│  Admin Flow:                                             │
│  1. GET /admin/face-verifications?status=pending         │
│  2. ดูรูปเซลฟี่ vs รูปเอกสาร (side-by-side)             │
│  3. PATCH /admin/face-verification/:id                   │
│     { action: "approve" | "reject" | "retry" }          │
│  4. Trigger updates users.face_verified = true           │
│                                                           │
└─────────────────────────────────────────────────────────┘
```

---

## 🗄️ Database Schema

### face_verifications Table
```sql
- verification_id (PK)
- user_id (FK → users)
- selfie_url
- liveness_video_url
- document_type ← NEW (national_id | passport)
- document_id ← NEW (FK → provider_documents)
- match_confidence (0-100%)
- is_match
- liveness_passed
- verification_status (pending/approved/rejected/needs_retry)
- created_at, verified_at, verified_by
- rejection_reason, retry_count
```

### users Table (Updated)
```sql
- face_verified (BOOLEAN) ← NEW
- face_verification_id (FK) ← NEW
```

---

## 🔑 Key Endpoints

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/provider/face-verification` | POST | Provider | ส่งเซลฟี่ + document_id + document_type |
| `/provider/face-verification` | GET | Provider | เช็คสถานะการยืนยัน |
| `/admin/face-verifications` | GET | Admin | ดูรายการที่รอตรวจสอบ |
| `/admin/face-verification/:id` | PATCH | Admin | อนุมัติ/ปฏิเสธ/ให้ส่งใหม่ |
| `/admin/face-verification/:id/trigger-matching` | POST | Admin | ทริกเกอร์ face matching แบบ manual |

---

## ⚠️ Breaking Changes (Nov 21, 2025)

### Frontend ต้องอัปเดต API Calls

**OLD (ใช้ไม่ได้แล้ว):**
```typescript
{
  selfie_url: "https://...",
  national_id_doc_id: 123  // ❌ REMOVED
}
```

**NEW (ต้องใช้):**
```typescript
{
  selfie_url: "https://...",
  document_id: 123,           // ✅ NEW
  document_type: "national_id" // ✅ NEW (or "passport")
}
```

---

## 🔒 Security & Compliance

### PDPA Requirements
- ✅ User consent required before collecting biometric data
- ✅ Encrypt selfie URLs at rest
- ✅ Admin-only access to verification photos
- ⚠️ TODO: Auto-delete after 90 days

### Anti-Spoofing
- ⚠️ TODO: Implement real liveness detection
- ⚠️ TODO: Replace mock face matching (currently 85.5% hardcoded)

---

## 📞 Support

### Common Questions

**Q: ต้อง update Frontend อย่างไร?**  
A: อ่าน `FRONTEND_FACE_VERIFICATION_PASSPORT_GUIDE.md` มี React components พร้อมใช้

**Q: Migration run แล้วหรือยัง?**  
A: ✅ Migration 020 และ 021 run สำเร็จแล้ว (verified)

**Q: รองรับเอกสารอะไรบ้าง?**  
A: บัตรประชาชนไทย (`national_id`) และพาสปอร์ต (`passport`)

**Q: Face matching ทำงานยังไง?**  
A: ตอนนี้ใช้ mock API (85.5% confidence), ต้อง integrate AWS Rekognition/Azure Face API

---

## 🎯 TODO List

### High Priority
- [ ] Integrate AWS Rekognition CompareFaces API
- [ ] Add real liveness detection service
- [ ] Frontend implementation (React components provided)
- [ ] Create GCS signed URL endpoint for selfie upload

### Medium Priority
- [ ] Update provider registration flow to require face verification
- [ ] Admin dashboard UI for photo comparison
- [ ] End-to-end testing (Thai + Foreign providers)

### Low Priority
- [ ] PDPA consent forms
- [ ] Auto-delete biometric data after 90 days
- [ ] Audit logging for verification photo access
- [ ] Multi-language error messages (English)

---

## 📈 Version History

| Version | Date | Changes | Files Updated |
|---------|------|---------|---------------|
| 1.0 | Nov 21, 2025 | Initial Face Verification System | Migration 020, 6 handlers, 5 endpoints |
| 1.1 | Nov 21, 2025 | Added Passport Support | Migration 021, updated all handlers |

---

## 📁 File Structure

```
docs/face-verification/
├── README.md (this file)
├── FACE_VERIFICATION_GUIDE.md (API docs)
├── FACE_VERIFICATION_IMPLEMENTATION_SUMMARY.md (Backend summary)
├── PASSPORT_SUPPORT_IMPLEMENTATION_SUMMARY.md (Update summary)
└── FRONTEND_FACE_VERIFICATION_PASSPORT_GUIDE.md (Frontend guide) ⭐
```

---

## 🚀 Deployment Status

**Current Build:** skillmatch-api-passport (71MB)  
**Server Status:** ✅ Running (port 8080)  
**Migrations:** ✅ 020 & 021 executed successfully  
**Database:** ✅ face_verifications table with document_type & document_id  
**Endpoints:** ✅ All 5 routes registered  

**Production Ready:** ⚠️ Mock API mode only  
**Real Production Ready:** After AWS/Azure integration

---

**Last Updated:** November 21, 2025  
**Maintained By:** SkillMatch Backend Team
