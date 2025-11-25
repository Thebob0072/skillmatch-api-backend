# 📚 SkillMatch API Documentation

Complete documentation for SkillMatch marketplace platform - connecting service providers with clients.

---

## 📁 Documentation Structure

```
docs/
├── api-reference/          # Complete API documentation
├── frontend-guides/        # Frontend integration guides
├── backend-guides/         # Backend development guides
├── system-guides/          # Feature-specific system guides
├── implementation/         # Implementation summaries
├── face-verification/      # Face verification system docs
├── sql-migrations/         # Database migration files (16 files)
└── sql-scripts/            # Maintenance & seed scripts (2 files)
```

---

## 🚀 Quick Start Guides

### For Frontend Developers
1. **Start here:** [`frontend-guides/FRONTEND_GUIDE.md`](./frontend-guides/FRONTEND_GUIDE.md)
2. **API Reference:** [`api-reference/API_REFERENCE_FOR_FRONTEND.md`](./api-reference/API_REFERENCE_FOR_FRONTEND.md)
3. **Provider Routes:** [`frontend-guides/FRONTEND_PROVIDER_ROUTES.md`](./frontend-guides/FRONTEND_PROVIDER_ROUTES.md)
4. **Face Verification:** [`face-verification/FRONTEND_FACE_VERIFICATION_PASSPORT_GUIDE.md`](./face-verification/FRONTEND_FACE_VERIFICATION_PASSPORT_GUIDE.md) ⭐

### For Backend Developers
1. **Database Schema:** [`backend-guides/DATABASE_STRUCTURE.md`](./backend-guides/DATABASE_STRUCTURE.md)
2. **Security Guide:** [`backend-guides/SECURITY.md`](./backend-guides/SECURITY.md)
3. **Implementation Notes:** [`backend-guides/BACKEND_CHECKLIST_ERRORS.md`](./backend-guides/BACKEND_CHECKLIST_ERRORS.md)

### For System Integration
1. **Complete API:** [`api-reference/COMPLETE_API_DOCUMENTATION.md`](./api-reference/COMPLETE_API_DOCUMENTATION.md)
2. **Payment System:** [`system-guides/PAYMENT_SYSTEM_GUIDE.md`](./system-guides/PAYMENT_SYSTEM_GUIDE.md)
3. **Financial System:** [`system-guides/FINANCIAL_SYSTEM_GUIDE.md`](./system-guides/FINANCIAL_SYSTEM_GUIDE.md)

### For Database Management
1. **Database Schema:** [`backend-guides/DATABASE_STRUCTURE.md`](./backend-guides/DATABASE_STRUCTURE.md)
2. **Migrations:** [`sql-migrations/README.md`](./sql-migrations/README.md) - 16 migration files
3. **Scripts:** [`sql-scripts/README.md`](./sql-scripts/README.md) - Maintenance & seed data

---

## 📂 Documentation Categories

### 🔌 API Reference (`api-reference/`)

Complete API documentation with all endpoints, request/response formats.

| File | Description | For |
|------|-------------|-----|
| **API_REFERENCE_FOR_FRONTEND.md** | Complete API documentation for frontend | Frontend Developers |
| **COMPLETE_API_DOCUMENTATION.md** | Full API specification with all endpoints | All Developers |
| **SERVICE_CATEGORY_API.md** | Service category management APIs | Frontend & Backend |

**Total Files:** 3

---

### 🎨 Frontend Guides (`frontend-guides/`)

Integration guides, React examples, and frontend-specific documentation.

| File | Description | Topics |
|------|-------------|--------|
| **FRONTEND_GUIDE.md** | Main frontend integration guide | Setup, Authentication, State Management |
| **FRONTEND_PROVIDER_ROUTES.md** | Provider-specific routes and components | Provider Dashboard, Profile, Bookings |
| **FRONTEND_SERVICE_CATEGORY_GUIDE.md** | Service category UI/UX guide | Category Browse, Filters, Display |
| **FINANCIAL_FRONTEND_GUIDE.md** | Financial system frontend guide | Wallet, Withdrawals, Transactions |
| **FRONTEND_API_PAYLOADS.md** | Request/Response payload examples | API Integration Examples |

**Total Files:** 5

**Quick Links:**
- Authentication Flow → `FRONTEND_GUIDE.md#authentication`
- Provider Dashboard → `FRONTEND_PROVIDER_ROUTES.md#dashboard`
- Wallet Integration → `FINANCIAL_FRONTEND_GUIDE.md#wallet-ui`

---

### ⚙️ Backend Guides (`backend-guides/`)

Database schema, security, and backend development guidelines.

| File | Description | Topics |
|------|-------------|--------|
| **DATABASE_STRUCTURE.md** | Complete database schema | Tables, Relationships, Indexes |
| **SECURITY.md** | Security best practices | JWT, PDPA, Data Protection |
| **BACKEND_CHECKLIST_ERRORS.md** | Common errors and fixes | Troubleshooting, Solutions |

**Total Files:** 3

**Quick Links:**
- Database Schema → `DATABASE_STRUCTURE.md`
- JWT Implementation → `SECURITY.md#jwt-authentication`
- Common Errors → `BACKEND_CHECKLIST_ERRORS.md`

---

### 🛠️ System Guides (`system-guides/`)

Feature-specific documentation for all major systems.

#### User Management & Admin
| File | Description |
|------|-------------|
| **ADMIN_ROLE_GUIDE.md** | Admin role management, GOD tier, permissions |
| **PROVIDER_SYSTEM_GUIDE.md** | Provider verification, tiers, documents |

#### Communication & Social
| File | Description |
|------|-------------|
| **MESSAGING_GUIDE.md** | Real-time chat, WebSocket, conversations |
| **NOTIFICATION_GUIDE.md** | Push notifications, in-app notifications |
| **BLOCK_GUIDE.md** | User blocking system |
| **REPORT_GUIDE.md** | Report system for abuse/violations |

#### Financial & Payments
| File | Description |
|------|-------------|
| **FINANCIAL_SYSTEM_GUIDE.md** | Wallet, transactions, withdrawals |
| **PAYMENT_SYSTEM_GUIDE.md** | Stripe integration, subscription, booking payments |

#### Business Logic
| File | Description |
|------|-------------|
| **ANALYTICS_GUIDE.md** | Analytics, metrics, reporting |
| **SCHEDULE_SYSTEM_GUIDE.md** | Provider availability scheduling |
| **LOCATION_GUIDE.md** | Location-based services, geolocation |
| **SERVICE_TYPE_GUIDE.md** | Service types (Incall/Outcall/Both) |

**Total Files:** 12

---

### 📋 Implementation (`implementation/`)

Implementation summaries and change logs.

| File | Description |
|------|-------------|
| **IMPLEMENTATION_SUMMARY.md** | Overall system implementation summary |
| **IMPLEMENTATION_SUMMARY_PROVIDER.md** | Provider system implementation details |

**Total Files:** 2

---

### 🔐 Face Verification (`face-verification/`)

Complete documentation for biometric face verification system.

| File | Description | For |
|------|-------------|-----|
| **README.md** | Face verification overview | All Developers |
| **FRONTEND_FACE_VERIFICATION_PASSPORT_GUIDE.md** ⭐ | Frontend integration guide | Frontend (React/TypeScript) |
| **FACE_VERIFICATION_GUIDE.md** | API documentation | Backend & API Integration |
| **FACE_VERIFICATION_IMPLEMENTATION_SUMMARY.md** | Backend implementation | Backend Developers |
| **PASSPORT_SUPPORT_IMPLEMENTATION_SUMMARY.md** | Passport support update | Backend & DevOps |

**Total Files:** 5

**Features:**
- ✅ Thai National ID verification
- ✅ Passport verification (foreign providers)
- ⚠️ Face matching (mock API - needs AWS/Azure integration)
- ⚠️ Liveness detection (TODO)

---

### 🗄️ SQL Migrations (`sql-migrations/`)

Database migration files for schema creation and updates.

| File Pattern | Description |
|-------------|-------------|
| `005_add_location_details.sql` | Location system (provinces, districts) |
| `007_add_messaging_system.sql` | Real-time chat system |
| `013_add_financial_system.sql` | Wallet & transactions |
| `015_add_provider_system.sql` | Provider features (packages, schedules) |
| `020_add_face_verification.sql` | Face verification system |
| `021_add_passport_support_face_verification.sql` | Passport support |
| **...and 10 more** | See [`sql-migrations/README.md`](./sql-migrations/README.md) |

**Total Files:** 16 migrations (005-021)

**Next migration number:** 022

---

### 🛠️ SQL Scripts (`sql-scripts/`)

Maintenance scripts and seed data for database management.

| File | Description | Use Case |
|------|-------------|----------|
| **fix_god_profile.sql** | Fix GOD account data | When GOD profile shows mock data |
| **seed_providers.sql** | Seed demo provider data | Dev/test environment setup |

**Total Files:** 2

⚠️ **Warning:** Only run seed scripts in dev/test environments, not production!

---

## 📊 System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    SkillMatch Platform                       │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  Frontend (React/TypeScript)                                 │
│  ├── User Dashboard                                          │
│  ├── Provider Dashboard                                      │
│  └── Admin Panel                                             │
│                                                               │
│  Backend (Go + Gin Framework)                                │
│  ├── Authentication (JWT)                                    │
│  ├── Real-time Communication (WebSocket)                     │
│  ├── Payment Processing (Stripe)                             │
│  ├── Face Verification (AWS/Azure)                           │
│  └── File Storage (Google Cloud Storage)                     │
│                                                               │
│  Database (PostgreSQL 15)                                    │
│  ├── Users & Profiles                                        │
│  ├── Bookings & Reviews                                      │
│  ├── Financial Transactions                                  │
│  └── Messages & Notifications                                │
│                                                               │
│  Cache & Sessions (Redis)                                    │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

---

## 🔑 Key Features

### Core Features
- ✅ User Authentication (JWT + Google OAuth)
- ✅ Provider Profiles with KYC verification
- ✅ Service Booking System
- ✅ Real-time Messaging (WebSocket)
- ✅ Payment Processing (Stripe)
- ✅ Review & Rating System
- ✅ Financial System (Wallet, Withdrawals)

### Advanced Features
- ✅ Face Verification (Thai ID + Passport)
- ✅ Provider Tier System (auto-calculated)
- ✅ Schedule Management
- ✅ Analytics & Reporting
- ✅ Block & Report System
- ✅ Location-based Search
- ✅ Multi-service Categories

### Admin Features
- ✅ GOD Tier (super admin)
- ✅ KYC Approval Workflow
- ✅ Financial Management
- ✅ User Management
- ✅ Content Moderation

---

## 🎯 Common Tasks

### Adding a New Feature

1. **Database Migration:**
   ```bash
   # Create new migration file (next number: 022)
   touch docs/sql-migrations/022_add_new_feature.sql
   
   # Write migration SQL with BEGIN/COMMIT
   # Server will auto-execute on restart
   
   # Update DATABASE_STRUCTURE.md
   ```

2. **Create Handlers:**
   ```go
   // Follow pattern: func handlerName(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc
   ```

3. **Register Routes:**
   ```go
   // In main.go, add to appropriate group (public/protected/admin)
   ```

4. **Update Documentation:**
   - Add to relevant guide in `docs/system-guides/`
   - Update `API_REFERENCE_FOR_FRONTEND.md`
   - Update `sql-migrations/README.md`

### Testing New Endpoints

```bash
# Health check
curl http://localhost:8080/ping

# Protected endpoint (requires JWT)
curl -H "Authorization: Bearer <token>" http://localhost:8080/users/me
```

### Running Database Scripts

```bash
# Connect to PostgreSQL
docker exec -it postgres_db psql -U admin -d skillmatch_db

# Run migration manually (if needed)
\i docs/sql-migrations/022_add_new_feature.sql

# Run maintenance script
\i docs/sql-scripts/fix_god_profile.sql

# Seed test data
\i docs/sql-scripts/seed_providers.sql
```

### Checking Database

```bash
# Connect to PostgreSQL
docker exec -it postgres_db psql -U admin -d skillmatch_db

# List tables
\dt

# Describe table
\d table_name
```

---

## ⚠️ Breaking Changes Log

### November 21, 2025 - Face Verification Passport Support
**Breaking Change:** API request format changed

❌ **OLD:**
```json
{
  "national_id_doc_id": 123
}
```

✅ **NEW:**
```json
{
  "document_id": 123,
  "document_type": "national_id"  // or "passport"
}
```

**Impact:** Frontend must update all face verification API calls  
**Documentation:** See [`face-verification/PASSPORT_SUPPORT_IMPLEMENTATION_SUMMARY.md`](./face-verification/PASSPORT_SUPPORT_IMPLEMENTATION_SUMMARY.md)

---

## 🔒 Security Guidelines

### Authentication
- All protected routes require JWT Bearer token
- Token format: `Authorization: Bearer <token>`
- Tokens expire after 7 days (configurable)

### Data Protection
- Passwords: bcrypt hashed
- Biometric data: Encrypted at rest
- PII: PDPA compliant storage

### API Security
- Rate limiting: Implemented
- CORS: Configured for production
- Input validation: All endpoints

**Full Guide:** [`backend-guides/SECURITY.md`](./backend-guides/SECURITY.md)

---

## 🐛 Troubleshooting

### Common Issues

| Issue | Solution | Reference |
|-------|----------|-----------|
| PostgreSQL array scan error | Use COALESCE with empty array | `BACKEND_CHECKLIST_ERRORS.md` |
| JWT token expired | Re-login to get new token | `SECURITY.md#jwt` |
| WebSocket not connecting | Check CORS and token format | `MESSAGING_GUIDE.md#websocket` |
| Stripe webhook failing | Verify webhook secret | `PAYMENT_SYSTEM_GUIDE.md#webhooks` |
| Face verification failing | Check document_type matches | `face-verification/README.md` |

---

## 📈 Version & Status

### Current Version
- **API Version:** 1.1
- **Database Version:** Migration 021
- **Documentation Last Updated:** November 21, 2025

### Build Status
- **Server:** ✅ Running (port 8080)
- **Database:** ✅ PostgreSQL 15 (Docker)
- **Cache:** ✅ Redis 7 (Docker)
- **Binary:** skillmatch-api-passport (71MB)

### Feature Status

| Feature | Status | Notes |
|---------|--------|-------|
| Authentication | ✅ Production | JWT + Google OAuth |
| Booking System | ✅ Production | Full booking lifecycle |
| Messaging | ✅ Production | WebSocket real-time |
| Payment | ✅ Production | Stripe integration |
| Face Verification | ⚠️ Mock Mode | Needs AWS/Azure integration |
| Financial System | ✅ Production | Wallet + Withdrawals |
| Admin Panel | ✅ Production | GOD tier + KYC |

---

## 🚀 Deployment

### Environment Variables Required

```bash
# Authentication
JWT_SECRET_KEY=<secret>
GOOGLE_CLIENT_ID=<id>
GOOGLE_CLIENT_SECRET=<secret>

# Stripe
STRIPE_SECRET_KEY=<key>
STRIPE_WEBHOOK_SECRET=<secret>

# Google Cloud
GOOGLE_APPLICATION_CREDENTIALS=key/gcs-key.json

# Database
DATABASE_URL=postgres://admin:password@localhost:5432/skillmatch_db
```

### Start Services

```bash
# Start databases
docker-compose up -d

# Run server
go run .

# Build for production
go build -o skillmatch-api .
```

---

## 📞 Support

### Documentation Issues
- Missing information? Check [`TODO.md`](../TODO.md)
- Found errors? Update relevant guide and submit PR

### Development Help
- **Frontend:** Start with `frontend-guides/FRONTEND_GUIDE.md`
- **Backend:** Check `backend-guides/DATABASE_STRUCTURE.md`
- **API:** See `api-reference/COMPLETE_API_DOCUMENTATION.md`
- **Features:** Browse `system-guides/` for specific systems

---

## 🎯 TODO & Roadmap

See [`TODO.md`](../TODO.md) for complete task list.

### High Priority
- [ ] Integrate AWS Rekognition for real face matching
- [ ] Add liveness detection
- [ ] Complete payment reconciliation system
- [ ] Add automated testing suite

### Medium Priority
- [ ] Multi-language support (EN/TH)
- [ ] Email notification system
- [ ] Advanced analytics dashboard
- [ ] Mobile app API optimization

---

## 📚 Documentation Index

### By Role

**Frontend Developer:**
1. `frontend-guides/FRONTEND_GUIDE.md`
2. `api-reference/API_REFERENCE_FOR_FRONTEND.md`
3. `face-verification/FRONTEND_FACE_VERIFICATION_PASSPORT_GUIDE.md`

**Backend Developer:**
1. `backend-guides/DATABASE_STRUCTURE.md`
2. `backend-guides/SECURITY.md`
3. `system-guides/` (all feature guides)

**DevOps:**
1. `backend-guides/SECURITY.md`
2. `implementation/` (deployment notes)
3. `face-verification/PASSPORT_SUPPORT_IMPLEMENTATION_SUMMARY.md`

**API Integrator:**
1. `api-reference/COMPLETE_API_DOCUMENTATION.md`
2. `api-reference/SERVICE_CATEGORY_API.md`
3. `face-verification/FACE_VERIFICATION_GUIDE.md`

**Project Manager:**
1. `docs/README.md` (this file)
2. `TODO.md`
3. `implementation/IMPLEMENTATION_SUMMARY.md`

---

**Last Updated:** November 21, 2025  
**Maintained By:** SkillMatch Backend Team  
**Repository:** [GitHub](https://github.com/skillmatch/api)
