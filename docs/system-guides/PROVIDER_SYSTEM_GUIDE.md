# 🎯 PROVIDER SYSTEM GUIDE - Frontend Developer Documentation

## 📋 Table of Contents
1. [Overview](#overview)
2. [User vs Provider Registration](#user-vs-provider-registration)
3. [Provider Document Upload](#provider-document-upload)
4. [Provider Tier System](#provider-tier-system)
5. [Admin Provider Management](#admin-provider-management)
6. [API Reference](#api-reference)
7. [Frontend Integration Examples](#frontend-integration-examples)

---

## 🎯 Overview

### ความแตกต่างระหว่าง User และ Provider

#### 👥 **User (ผู้ใช้บริการทั่วไป)**
- ✅ ลงทะเบียนง่ายๆ ด้วย Email OTP
- ❌ **ไม่ต้องส่งเอกสารใดๆ**
- ✅ เลือก Subscription Tier: General (ฟรี) หรือ Premium (เสียเงิน)
- ✅ ดูข้อมูล provider, จองบริการ, เขียนรีวิว
- 🔒 ไม่สามารถสร้างแพ็คเกจหรือขายบริการได้

#### 💼 **Provider (ผู้ให้บริการ)**
- ✅ ลงทะเบียนเป็น Provider พร้อมระบุหมวดหมู่บริการ
- ⚠️  **ต้องส่งเอกสารยืนยันตัวตน** (บัตรประชาชน, ใบรับรองสุขภาพ, ฯลฯ)
- ⏳ รอ Admin ตรวจสอบและอนุมัติเอกสาร
- 📊 ระบบจัดอันดับ Provider Tier อัตโนมัติ (ตาม Rating, Reviews, Performance)
- 💰 สามารถสร้างแพ็คเกจ, รับการจอง, ได้รับเงิน
- 🎖️  มี Provider Tier แยกจาก Subscription Tier

---

## 🔐 User vs Provider Registration

### 1️⃣ User Registration Flow (สำหรับผู้ใช้บริการทั่วไป)

```typescript
// Step 1: Send OTP
const sendOTP = async (email: string) => {
  const response = await fetch('http://localhost:8080/auth/send-verification', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email })
  });
  return response.json(); // { message, expires_in: "10 minutes" }
};

// Step 2: Verify OTP (optional - can skip to registration)
const verifyOTP = async (email: string, otp: string) => {
  const response = await fetch('http://localhost:8080/auth/verify-email', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, otp })
  });
  return response.json(); // { verified: true, message }
};

// Step 3: Register User
const registerUser = async (userData: {
  username: string;
  email: string;
  password: string;
  gender_id: number;
  first_name: string;
  last_name: string;
  phone: string;
  otp: string; // OTP จากอีเมล
}) => {
  const response = await fetch('http://localhost:8080/register', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(userData)
  });
  const data = await response.json();
  // { message: "Registration successful", user_id, token }
  
  // เก็บ token สำหรับใช้ต่อ
  localStorage.setItem('auth_token', data.token);
  return data;
};
```

### 2️⃣ Provider Registration Flow (สำหรับผู้ให้บริการ)

```typescript
// Step 1: Send OTP (เหมือน User)
await sendOTP('provider@example.com');

// Step 2: Register as Provider
const registerProvider = async (providerData: {
  // ข้อมูลพื้นฐาน (เหมือน User)
  username: string;
  email: string;
  password: string;
  gender_id: number;
  first_name: string;
  last_name: string;
  phone: string;
  otp: string;

  // ข้อมูล Provider เพิ่มเติม
  category_ids: number[]; // [1, 2, 3] - หมวดหมู่บริการที่ให้บริการ
  service_type: string; // "Incall", "Outcall", "Both"
  bio: string;
  province: string;
  district: string;
}) => {
  const response = await fetch('http://localhost:8080/register/provider', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(providerData)
  });
  const data = await response.json();
  // {
  //   message: "Provider registration successful. Please upload required documents...",
  //   user_id: 31,
  //   token: "eyJhbGci...",
  //   next_step: "Upload documents: National ID, Health Certificate"
  // }

  localStorage.setItem('auth_token', data.token);
  return data;
};

// ตัวอย่างการเรียกใช้
const result = await registerProvider({
  username: 'provider1',
  email: 'provider@example.com',
  password: 'securepassword123',
  gender_id: 2,
  first_name: 'Provider',
  last_name: 'Name',
  phone: '0812345678',
  otp: '123456',
  category_ids: [1, 2], // Massage, Spa
  service_type: 'Both',
  bio: 'Professional massage therapist with 5 years experience',
  province: 'Bangkok',
  district: 'Sukhumvit'
});
```

---

## 📄 Provider Document Upload

### Required Documents (เอกสารที่ต้องส่ง)

| Document Type | Display Name | Required | Description |
|--------------|--------------|----------|-------------|
| `national_id` | National ID Card | ✅ Yes | บัตรประชาชน / บัตรประจำตัวประชาชน |
| `health_certificate` | Health Certificate | ✅ Yes | ใบรับรองสุขภาพ (ไม่เกิน 6 เดือน) |
| `business_license` | Business License | ⚪ Optional | ใบอนุญาตประกอบธุรกิจ (ถ้ามี) |
| `portfolio` | Portfolio | ⚪ Optional | ผลงาน / รูปตัวอย่าง |
| `certification` | Certification | ⚪ Optional | ใบประกาศนียบัตร / ใบรับรองมาตรฐาน |
| `other` | Other Documents | ⚪ Optional | เอกสารอื่นๆ |

### Document Upload Flow

```typescript
// POST /provider/documents - Upload document (ต้อง login)
const uploadProviderDocument = async (documentData: {
  document_type: string; // 'national_id', 'health_certificate', etc.
  file_url: string; // URL ของไฟล์ที่อัปโหลดไปยัง storage
  file_name?: string; // ชื่อไฟล์
}) => {
  const token = localStorage.getItem('auth_token');
  
  const response = await fetch('http://localhost:8080/provider/documents', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(documentData)
  });
  
  const data = await response.json();
  // {
  //   message: "Document uploaded successfully",
  //   document_id: 1,
  //   status: "pending"
  // }
  return data;
};

// GET /provider/documents - Get my documents
const getMyDocuments = async () => {
  const token = localStorage.getItem('auth_token');
  
  const response = await fetch('http://localhost:8080/provider/documents', {
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    }
  });
  
  const data = await response.json();
  // {
  //   documents: [
  //     {
  //       document_id: 1,
  //       document_type: "national_id",
  //       file_url: "https://...",
  //       file_name: "id_card.jpg",
  //       verification_status: "pending", // "pending", "approved", "rejected"
  //       uploaded_at: "2025-11-14T12:00:00Z",
  //       verified_at: null,
  //       rejection_reason: null
  //     }
  //   ],
  //   total: 1
  // }
  return data;
};
```

### React Component Example: Document Upload

```tsx
import { useState, useEffect } from 'react';

interface Document {
  document_id: number;
  document_type: string;
  file_url: string;
  file_name?: string;
  verification_status: 'pending' | 'approved' | 'rejected';
  uploaded_at: string;
  rejection_reason?: string;
}

export function ProviderDocumentUpload() {
  const [documents, setDocuments] = useState<Document[]>([]);
  const [uploading, setUploading] = useState(false);

  // โหลดเอกสารที่มีอยู่
  useEffect(() => {
    loadDocuments();
  }, []);

  const loadDocuments = async () => {
    const data = await getMyDocuments();
    setDocuments(data.documents);
  };

  const handleFileUpload = async (documentType: string, file: File) => {
    setUploading(true);
    
    try {
      // 1. Upload file to storage (GCS, S3, etc.)
      const fileUrl = await uploadFileToStorage(file);
      
      // 2. Submit document metadata to API
      await uploadProviderDocument({
        document_type: documentType,
        file_url: fileUrl,
        file_name: file.name
      });
      
      // 3. Reload documents
      await loadDocuments();
      alert('Document uploaded successfully!');
    } catch (error) {
      alert('Failed to upload document');
    } finally {
      setUploading(false);
    }
  };

  const getStatusBadge = (status: string) => {
    const colors = {
      pending: 'yellow',
      approved: 'green',
      rejected: 'red'
    };
    return <span className={`badge badge-${colors[status]}`}>{status}</span>;
  };

  return (
    <div className="provider-documents">
      <h2>Document Verification</h2>
      
      {/* แสดงเอกสารที่อัปโหลดแล้ว */}
      <div className="documents-list">
        {documents.map(doc => (
          <div key={doc.document_id} className="document-item">
            <span>{doc.document_type}</span>
            {getStatusBadge(doc.verification_status)}
            {doc.rejection_reason && (
              <p className="text-red">{doc.rejection_reason}</p>
            )}
          </div>
        ))}
      </div>

      {/* Upload forms */}
      <div className="upload-section">
        <h3>Upload Documents</h3>
        <DocumentUploadForm 
          type="national_id" 
          label="National ID Card *"
          onUpload={handleFileUpload}
          disabled={uploading}
        />
        <DocumentUploadForm 
          type="health_certificate" 
          label="Health Certificate *"
          onUpload={handleFileUpload}
          disabled={uploading}
        />
      </div>
    </div>
  );
}
```

---

## 📊 Provider Tier System

### Tier Calculation Algorithm

Provider Tier คำนวณจาก **Tier Points** (คะแนน):

```
Total Points (max 600) = 
  + Rating Points (0-100)         = average_rating * 20
  + Completed Bookings (0-250)    = completed_bookings * 5 (max 50 bookings)
  + Total Reviews (0-150)         = total_reviews * 3 (max 50 reviews)
  + Response Rate (0-50)          = response_rate * 0.5
  + Acceptance Rate (0-50)        = acceptance_rate * 0.5
```

### Tier Assignment

| Tier | Points Required | Benefits |
|------|----------------|----------|
| **General** | 0-99 points | Basic visibility |
| **Silver** | 100-249 points | Higher ranking in search |
| **Diamond** | 250-399 points | Premium badge, priority support |
| **Premium** | 400+ points | Top ranking, featured listings |

### API: Get My Provider Tier

```typescript
// GET /provider/my-tier - ดู Tier ปัจจุบัน (ต้อง login)
const getMyProviderTier = async () => {
  const token = localStorage.getItem('auth_token');
  
  const response = await fetch('http://localhost:8080/provider/my-tier', {
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    }
  });
  
  const data = await response.json();
  // {
  //   current_tier_id: 2,
  //   current_tier_name: "Silver",
  //   tier_points: 150,
  //   average_rating: 4.5,
  //   total_reviews: 10,
  //   completed_bookings: 20,
  //   response_rate: 95.0,
  //   acceptance_rate: 85.0,
  //   next_tier_id: 3,
  //   next_tier_name: "Diamond",
  //   points_to_next_tier: 100
  // }
  return data;
};

// GET /provider/tier-history - ดูประวัติการเปลี่ยน Tier
const getMyTierHistory = async () => {
  const token = localStorage.getItem('auth_token');
  
  const response = await fetch('http://localhost:8080/provider/tier-history', {
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    }
  });
  
  const data = await response.json();
  // {
  //   history: [
  //     {
  //       history_id: 1,
  //       old_tier_name: "General",
  //       new_tier_name: "Silver",
  //       change_type: "auto", // "auto", "manual", "subscription"
  //       reason: "Auto tier update based on points: 150",
  //       changed_at: "2025-11-14T12:00:00Z"
  //     }
  //   ],
  //   total: 1
  // }
  return data;
};
```

### React Component: Provider Tier Dashboard

```tsx
interface TierInfo {
  current_tier_name: string;
  tier_points: number;
  next_tier_name?: string;
  points_to_next_tier?: number;
}

export function ProviderTierCard() {
  const [tierInfo, setTierInfo] = useState<TierInfo | null>(null);

  useEffect(() => {
    loadTierInfo();
  }, []);

  const loadTierInfo = async () => {
    const data = await getMyProviderTier();
    setTierInfo(data);
  };

  if (!tierInfo) return <div>Loading...</div>;

  return (
    <div className="tier-card">
      <h3>Current Tier: {tierInfo.current_tier_name}</h3>
      <div className="tier-points">
        <span>{tierInfo.tier_points} points</span>
      </div>

      {tierInfo.next_tier_name && (
        <div className="next-tier">
          <p>Next Tier: {tierInfo.next_tier_name}</p>
          <p>Points needed: {tierInfo.points_to_next_tier}</p>
          <ProgressBar 
            current={tierInfo.tier_points}
            target={tierInfo.tier_points + (tierInfo.points_to_next_tier || 0)}
          />
        </div>
      )}
    </div>
  );
}
```

---

## 👮 Admin Provider Management

### Admin Endpoints

```typescript
// GET /admin/providers/pending - ดู providers ที่รอตรวจสอบ
const getAdminPendingProviders = async () => {
  const token = localStorage.getItem('auth_token');
  
  const response = await fetch('http://localhost:8080/admin/providers/pending', {
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    }
  });
  
  return response.json();
  // {
  //   providers: [
  //     {
  //       user_id: 31,
  //       username: "provider1",
  //       email: "provider@example.com",
  //       provider_verification_status: "documents_submitted",
  //       registration_date: "2025-11-14T12:00:00Z",
  //       total_documents: 2,
  //       approved_documents: 0,
  //       pending_documents: 2
  //     }
  //   ],
  //   total: 1
  // }
};

// PATCH /admin/verify-document/:documentId - อนุมัติ/ปฏิเสธเอกสาร
const adminVerifyDocument = async (documentId: number, status: 'approved' | 'rejected', rejection_reason?: string) => {
  const token = localStorage.getItem('auth_token');
  
  const response = await fetch(`http://localhost:8080/admin/verify-document/${documentId}`, {
    method: 'PATCH',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ status, rejection_reason })
  });
  
  return response.json();
  // { message: "Document approved successfully" }
};

// PATCH /admin/approve-provider/:userId - อนุมัติ provider (เมื่อเอกสารครบถ้วน)
const adminApproveProvider = async (userId: number, approve: boolean, reason?: string) => {
  const token = localStorage.getItem('auth_token');
  
  const response = await fetch(`http://localhost:8080/admin/approve-provider/${userId}`, {
    method: 'PATCH',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ approve, reason })
  });
  
  return response.json();
  // { message: "Provider approved successfully", user_id, status: "approved" }
};

// GET /admin/provider-stats - สถิติ providers ทั้งหมด
const getAdminProviderStats = async () => {
  const token = localStorage.getItem('auth_token');
  
  const response = await fetch('http://localhost:8080/admin/provider-stats', {
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    }
  });
  
  return response.json();
  // {
  //   total_providers: 50,
  //   approved_providers: 40,
  //   pending_providers: 8,
  //   rejected_providers: 2
  // }
};

// POST /admin/recalculate-provider-tiers - คำนวณ Tier ใหม่ทั้งหมด
const adminRecalculateProviderTiers = async () => {
  const token = localStorage.getItem('auth_token');
  
  const response = await fetch('http://localhost:8080/admin/recalculate-provider-tiers', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    }
  });
  
  return response.json();
  // {
  //   message: "Provider tiers recalculated successfully",
  //   total_providers: 40,
  //   updates: [...]
  // }
};

// PATCH /admin/set-provider-tier/:userId - เปลี่ยน Tier แบบ Manual
const adminSetProviderTier = async (userId: number, newTierId: number, reason: string) => {
  const token = localStorage.getItem('auth_token');
  
  const response = await fetch(`http://localhost:8080/admin/set-provider-tier/${userId}`, {
    method: 'PATCH',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ new_tier_id: newTierId, reason })
  });
  
  return response.json();
  // {
  //   message: "Provider tier updated successfully",
  //   user_id, old_tier_id, new_tier_id
  // }
};

// GET /admin/provider/:userId/tier-details - ดูรายละเอียด Tier (Admin)
const adminGetProviderTierDetails = async (userId: number) => {
  const token = localStorage.getItem('auth_token');
  
  const response = await fetch(`http://localhost:8080/admin/provider/${userId}/tier-details`, {
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    }
  });
  
  return response.json();
  // {
  //   user_id, username, email,
  //   current_tier_id, current_tier_name, tier_points,
  //   average_rating, total_reviews, completed_bookings,
  //   recommended_tier_id, recommended_tier_name
  // }
};
```

---

## 📖 API Reference

### Authentication & Registration

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/auth/send-verification` | ส่ง OTP ไปทาง email | ❌ |
| POST | `/auth/verify-email` | ยืนยัน OTP | ❌ |
| POST | `/register` | ลงทะเบียน User | ❌ |
| POST | `/register/provider` | ลงทะเบียน Provider | ❌ |
| POST | `/login` | Login | ❌ |

### Provider Document Management

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/provider/documents` | อัปโหลดเอกสาร | ✅ Provider |
| GET | `/provider/documents` | ดูเอกสารของตัวเอง | ✅ Provider |
| GET | `/provider/categories/me` | ดูหมวดหมู่บริการของตัวเอง | ✅ Provider |

### Provider Tier Management

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/provider/my-tier` | ดู Tier ปัจจุบัน | ✅ Provider |
| GET | `/provider/tier-history` | ดูประวัติการเปลี่ยน Tier | ✅ Provider |

### Admin Provider Management

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/admin/providers/pending` | ดู providers ที่รอตรวจสอบ | ✅ Admin |
| PATCH | `/admin/verify-document/:documentId` | อนุมัติ/ปฏิเสธเอกสาร | ✅ Admin |
| PATCH | `/admin/approve-provider/:userId` | อนุมัติ provider | ✅ Admin |
| GET | `/admin/provider-stats` | สถิติ providers | ✅ Admin |
| POST | `/admin/recalculate-provider-tiers` | คำนวณ Tier ใหม่ | ✅ Admin |
| PATCH | `/admin/set-provider-tier/:userId` | เปลี่ยน Tier (Manual) | ✅ Admin |
| GET | `/admin/provider/:userId/tier-details` | ดูรายละเอียด Tier | ✅ Admin |

---

## ✅ Implementation Checklist

### Phase 1: Registration (Frontend)
- [ ] สร้างหน้า "Register as Provider" แยกจาก User registration
- [ ] เพิ่มฟิลด์: `category_ids`, `service_type`, `bio`, `province`, `district`
- [ ] เพิ่ม Multi-select dropdown สำหรับเลือกหมวดหมู่บริการ
- [ ] เชื่อมต่อ API: `POST /register/provider`
- [ ] แสดง next step: "Upload documents" หลัง registration สำเร็จ

### Phase 2: Document Upload (Frontend)
- [ ] สร้างหน้า "Provider Documents" dashboard
- [ ] เพิ่ม file upload component สำหรับแต่ละประเภทเอกสาร
- [ ] อัปโหลดไฟล์ไปยัง cloud storage (GCS, S3)
- [ ] เชื่อมต่อ API: `POST /provider/documents`
- [ ] แสดงสถานะเอกสาร: pending, approved, rejected
- [ ] แสดง rejection_reason ถ้าถูกปฏิเสธ

### Phase 3: Provider Tier Display (Frontend)
- [ ] สร้าง "Provider Dashboard" ที่แสดง Tier ปัจจุบัน
- [ ] เชื่อมต่อ API: `GET /provider/my-tier`
- [ ] แสดง progress bar สู่ Tier ถัดไป
- [ ] แสดงประวัติการเปลี่ยน Tier: `GET /provider/tier-history`

### Phase 4: Admin Panel (Frontend)
- [ ] สร้างหน้า "Pending Providers" สำหรับ Admin
- [ ] เชื่อมต่อ API: `GET /admin/providers/pending`
- [ ] เพิ่มปุ่ม "Approve/Reject Document" สำหรับแต่ละเอกสาร
- [ ] เพิ่มปุ่ม "Approve/Reject Provider" เมื่อเอกสารครบถ้วน
- [ ] แสดงสถิติ: `GET /admin/provider-stats`
- [ ] เพิ่มฟีเจอร์ Manual Tier assignment

### Phase 5: Testing
- [ ] ทดสอบ Provider registration flow แบบ end-to-end
- [ ] ทดสอบ Document upload และ Admin verification
- [ ] ทดสอบ Auto Tier assignment (คำนวณอัตโนมัติ)
- [ ] ทดสอบ Manual Tier change (Admin เปลี่ยนเอง)

---

**Last Updated:** November 14, 2025, 2:00 PM  
**Backend:** Running on http://localhost:8080  
**Status:** ✅ Provider system fully implemented  
**Migration:** 015 executed successfully
