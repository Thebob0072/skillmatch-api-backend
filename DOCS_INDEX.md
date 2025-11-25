# 📚 SkillMatch API - Documentation Quick Reference

เอกสารทั้งหมดถูกจัดระเบียบไว้ใน **`docs/`** folder แล้ว

---

## 📁 โครงสร้างเอกสาร (31 ไฟล์ / 692KB)

```
docs/
├── README.md                    # 📖 Main documentation index
│
├── api-reference/               # 🔌 API Documentation (3 files)
│   ├── API_REFERENCE_FOR_FRONTEND.md
│   ├── COMPLETE_API_DOCUMENTATION.md
│   └── SERVICE_CATEGORY_API.md
│
├── frontend-guides/             # 🎨 Frontend Integration (5 files)
│   ├── FRONTEND_GUIDE.md
│   ├── FRONTEND_PROVIDER_ROUTES.md
│   ├── FRONTEND_SERVICE_CATEGORY_GUIDE.md
│   ├── FINANCIAL_FRONTEND_GUIDE.md
│   └── FRONTEND_API_PAYLOADS.md
│
├── backend-guides/              # ⚙️ Backend Development (3 files)
│   ├── DATABASE_STRUCTURE.md
│   ├── SECURITY.md
│   └── BACKEND_CHECKLIST_ERRORS.md
│
├── system-guides/               # 🛠️ Feature Systems (12 files)
│   ├── ADMIN_ROLE_GUIDE.md
│   ├── ANALYTICS_GUIDE.md
│   ├── BLOCK_GUIDE.md
│   ├── FINANCIAL_SYSTEM_GUIDE.md
│   ├── LOCATION_GUIDE.md
│   ├── MESSAGING_GUIDE.md
│   ├── NOTIFICATION_GUIDE.md
│   ├── PAYMENT_SYSTEM_GUIDE.md
│   ├── PROVIDER_SYSTEM_GUIDE.md
│   ├── REPORT_GUIDE.md
│   ├── SCHEDULE_SYSTEM_GUIDE.md
│   └── SERVICE_TYPE_GUIDE.md
│
├── implementation/              # 📋 Implementation Notes (2 files)
│   ├── IMPLEMENTATION_SUMMARY.md
│   └── IMPLEMENTATION_SUMMARY_PROVIDER.md
│
└── face-verification/           # 🔐 Face Verification (5 files)
    ├── README.md
    ├── FACE_VERIFICATION_GUIDE.md
    ├── FACE_VERIFICATION_IMPLEMENTATION_SUMMARY.md
    ├── PASSPORT_SUPPORT_IMPLEMENTATION_SUMMARY.md
    └── FRONTEND_FACE_VERIFICATION_PASSPORT_GUIDE.md ⭐
```

---

## 🚀 เริ่มต้นอย่างรวดเร็ว

### 👨‍💻 Frontend Developer
```bash
📖 เริ่มที่: docs/frontend-guides/FRONTEND_GUIDE.md
🔌 API: docs/api-reference/API_REFERENCE_FOR_FRONTEND.md
🔐 Face Verification: docs/face-verification/FRONTEND_FACE_VERIFICATION_PASSPORT_GUIDE.md
```

### 🔧 Backend Developer
```bash
🗄️ Database: docs/backend-guides/DATABASE_STRUCTURE.md
🔒 Security: docs/backend-guides/SECURITY.md
🛠️ Systems: docs/system-guides/
```

### 🗄️ Database Developer / DevOps
```bash
📊 Schema: docs/backend-guides/DATABASE_STRUCTURE.md
🔄 Migrations: docs/sql-migrations/README.md (16 files)
🛠️ Scripts: docs/sql-scripts/README.md (2 files)
```

### 🔌 API Integration
```bash
📚 Complete API: docs/api-reference/COMPLETE_API_DOCUMENTATION.md
💰 Payment: docs/system-guides/PAYMENT_SYSTEM_GUIDE.md
💸 Financial: docs/system-guides/FINANCIAL_SYSTEM_GUIDE.md
```

### 📊 Project Manager / Overview
```bash
📖 Main Index: docs/README.md
✅ TODO: TODO.md (root folder)
📝 Implementation: docs/implementation/
```

---

## 📂 เอกสารตาม Role

| Role | เริ่มต้นที่ | จำนวนไฟล์ที่เกี่ยวข้อง |
|------|------------|----------------------|
| **Frontend Dev** | `docs/frontend-guides/` | 5 + 3 API + 5 Face Verification = 13 files |
| **Backend Dev** | `docs/backend-guides/` | 3 + 12 System + 16 Migrations = 31 files |
| **Database Admin** | `docs/sql-migrations/` | 16 migrations + 2 scripts = 18 SQL files |
| **Full Stack** | `docs/README.md` | อ่านทั้งหมด 49 files (31 .md + 18 .sql) |
| **DevOps** | `docs/backend-guides/SECURITY.md` | 3 backend + 2 implementation + 18 SQL = 23 files |
| **API Integrator** | `docs/api-reference/` | 3 files |

---

## 🎯 เอกสารยอดนิยม (Top 12)

1. **`docs/README.md`** - Main documentation index ⭐
2. **`docs/api-reference/API_REFERENCE_FOR_FRONTEND.md`** - Complete API for frontend
3. **`docs/frontend-guides/FRONTEND_GUIDE.md`** - Frontend integration guide
4. **`docs/backend-guides/DATABASE_STRUCTURE.md`** - Database schema
5. **`docs/sql-migrations/README.md`** - Database migrations guide (NEW) ⭐
6. **`docs/face-verification/FRONTEND_FACE_VERIFICATION_PASSPORT_GUIDE.md`** - Face verification
7. **`docs/system-guides/PAYMENT_SYSTEM_GUIDE.md`** - Stripe payment integration
8. **`docs/system-guides/MESSAGING_GUIDE.md`** - Real-time chat (WebSocket)
9. **`docs/system-guides/FINANCIAL_SYSTEM_GUIDE.md`** - Wallet & withdrawals
10. **`docs/backend-guides/SECURITY.md`** - Security & JWT
11. **`docs/sql-scripts/README.md`** - Maintenance scripts (NEW) ⭐
12. **`TODO.md`** - Project tasks (root folder)

---

## 🔍 ค้นหาเอกสารเฉพาะเรื่อง

### Authentication & Security
- JWT: `docs/backend-guides/SECURITY.md#jwt-authentication`
- Google OAuth: `docs/frontend-guides/FRONTEND_GUIDE.md#google-oauth`
- Face Verification: `docs/face-verification/`

### Payment & Financial
- Stripe Integration: `docs/system-guides/PAYMENT_SYSTEM_GUIDE.md`
- Wallet System: `docs/system-guides/FINANCIAL_SYSTEM_GUIDE.md`
- Withdrawals: `docs/frontend-guides/FINANCIAL_FRONTEND_GUIDE.md`

### Communication
- Real-time Chat: `docs/system-guides/MESSAGING_GUIDE.md`
- WebSocket: `docs/system-guides/MESSAGING_GUIDE.md#websocket`
- Notifications: `docs/system-guides/NOTIFICATION_GUIDE.md`

### Provider Features
- Registration: `docs/system-guides/PROVIDER_SYSTEM_GUIDE.md`
- KYC Verification: `docs/system-guides/PROVIDER_SYSTEM_GUIDE.md#kyc`
- Tier System: `docs/system-guides/PROVIDER_SYSTEM_GUIDE.md#tiers`
- Schedule: `docs/system-guides/SCHEDULE_SYSTEM_GUIDE.md`

### Admin Features
- GOD Account: `docs/system-guides/ADMIN_ROLE_GUIDE.md#god-tier`
- User Management: `docs/system-guides/ADMIN_ROLE_GUIDE.md`
- Financial Admin: `docs/system-guides/FINANCIAL_SYSTEM_GUIDE.md#admin-features`

### Database Management
- Migrations: `docs/sql-migrations/README.md`
- Scripts: `docs/sql-scripts/README.md`
- Schema: `docs/backend-guides/DATABASE_STRUCTURE.md`

---

## 📊 สถิติเอกสาร

```
📚 Total Files: 49 files (31 .md + 18 .sql)
💾 Total Size: ~750KB
📁 Categories: 8

By Category:
- System Guides: 12 files (248KB) - .md ที่มากที่สุด
- SQL Migrations: 16 files (~50KB) - Migrations 005-021
- Frontend Guides: 5 files (192KB)
- Face Verification: 5 files (96KB)
- API Reference: 3 files (64KB)
- Backend Guides: 3 files (44KB)
- Implementation: 2 files (32KB)
- SQL Scripts: 2 files (~5KB)
- Main README: 1 file
```

---

## 🆕 อัปเดตล่าสุด (Nov 21, 2025)

### ✅ เพิ่มใหม่
- 🔐 Face Verification with Passport Support (5 documents)
- 📁 Documentation restructuring (จัดเก็บใน docs/)
- 🗄️ SQL organization (migrations + scripts)
- 📖 Main README.md สำหรับ navigation
- 📝 SQL-specific READMEs (migrations/scripts)

### 📂 การจัดเก็บไฟล์
- ✅ All .md files → `docs/` (8 categories)
- ✅ Migration files → `docs/sql-migrations/` (16 files)
- ✅ SQL scripts → `docs/sql-scripts/` (2 files)
- ✅ Only `TODO.md` remains in root

### ⚠️ Breaking Changes
- Face Verification API: `national_id_doc_id` → `document_id` + `document_type`
- Migration paths: `migrations/*.sql` → `docs/sql-migrations/*.sql`
- Scripts paths: `*.sql` → `docs/sql-scripts/*.sql`
- ดูรายละเอียด: `docs/face-verification/PASSPORT_SUPPORT_IMPLEMENTATION_SUMMARY.md`

---

## 🔗 Quick Links

| Category | Link |
|----------|------|
| 📚 Main Docs Index | [`docs/README.md`](./docs/README.md) |
| 🔌 API Reference | [`docs/api-reference/`](./docs/api-reference/) |
| 🎨 Frontend Guides | [`docs/frontend-guides/`](./docs/frontend-guides/) |
| ⚙️ Backend Guides | [`docs/backend-guides/`](./docs/backend-guides/) |
| 🛠️ System Guides | [`docs/system-guides/`](./docs/system-guides/) |
| 🔐 Face Verification | [`docs/face-verification/`](./docs/face-verification/) |
| 🗄️ SQL Migrations | [`docs/sql-migrations/`](./docs/sql-migrations/) |
| 🛠️ SQL Scripts | [`docs/sql-scripts/`](./docs/sql-scripts/) |
| ✅ Project TODO | [`TODO.md`](./TODO.md) |

---

## 💡 Tips

### ค้นหาเอกสารในโปรเจค
```bash
# ค้นหาคำใน docs
grep -r "keyword" docs/

# หาไฟล์ที่มีชื่อเฉพาะ
find docs/ -name "*payment*"

# ดูขนาดแต่ละ folder
du -sh docs/*/
```

### อ่านเอกสารแบบ offline
```bash
# ใช้ VS Code
code docs/README.md

# หรือเปิดใน browser (ถ้ามี markdown viewer)
open docs/README.md
```

---

## 📞 ติดต่อ

- **Documentation Issues:** สร้าง issue ใน GitHub
- **Missing Docs:** ตรวจสอบ `TODO.md`
- **Update Request:** Submit PR

---

**Last Updated:** November 21, 2025  
**Total Documentation Files:** 31  
**Project:** SkillMatch API
