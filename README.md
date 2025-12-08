# 🎯 SkillMatch API

> **Marketplace platform connecting service providers with clients**

Go-based REST API with real-time messaging, payment processing, and face verification.

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-336791?logo=postgresql)](https://postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7-DC382D?logo=redis)](https://redis.io/)
[![Stripe](https://img.shields.io/badge/Stripe-Payment-008CDD?logo=stripe)](https://stripe.com/)

---

## 📚 Documentation

### 👉 For Frontend Team
**[FRONTEND_COMPLETE_GUIDE.md](FRONTEND_COMPLETE_GUIDE.md)** - Complete integration guide:
- ✅ **119 API Endpoints** with examples
- ✅ **Filter System** (location, rating, tier, category, service_type, languages)
- ✅ **Translation Guide** (Thai/English) - ข้อมูลไหนที่ Backend ส่งมา vs ต้องแปลเอง
- ✅ **Authentication & Authorization**
- ✅ **Real-time WebSocket**
- ✅ **React Components Examples**

### For Backend Team
- [.github/copilot-instructions.md](.github/copilot-instructions.md) - Development guidelines
- [docs/backend-guides/DATABASE_STRUCTURE.md](docs/backend-guides/DATABASE_STRUCTURE.md) - Database schema
- [docs/system-guides/](docs/system-guides/) - System guides

---

## 🚀 Quick Start

### Prerequisites

- **Go 1.21+** ([Download](https://go.dev/dl/))
- **Docker & Docker Compose** ([Download](https://www.docker.com/products/docker-desktop))
- **Google Cloud Storage credentials** (for file uploads)
- **Stripe account** (for payments)

### Installation

```bash
# Clone repository
git clone https://github.com/your-org/skillmatch-api.git
cd skillmatch-api

# Install dependencies
go mod download

# Start databases (PostgreSQL + Redis)
docker-compose up -d

# Create environment file
cp .env.example .env
# Edit .env with your credentials

# Run server (migrations auto-execute on startup)
go run .
```

Server starts on `http://localhost:8080`

### Test API

```bash
# Health check
curl http://localhost:8080/ping

# Expected response:
# {"message":"pong!","postgres_time":"2025-11-25T10:30:00+07:00"}
```

---

## 📚 Documentation

**Complete documentation:** [`docs/README.md`](./docs/README.md)

### Quick Links

| For | Start Here |
|-----|-----------|
| 🎨 **Frontend Developers** | [`docs/frontend-guides/FRONTEND_GUIDE.md`](./docs/frontend-guides/FRONTEND_GUIDE.md) |
| ⚙️ **Backend Developers** | [`docs/backend-guides/DATABASE_STRUCTURE.md`](./docs/backend-guides/DATABASE_STRUCTURE.md) |
| 🔌 **API Integration** | [`docs/api-reference/API_REFERENCE_FOR_FRONTEND.md`](./docs/api-reference/API_REFERENCE_FOR_FRONTEND.md) |
| 🔐 **Face Verification** | [`docs/face-verification/FRONTEND_FACE_VERIFICATION_PASSPORT_GUIDE.md`](./docs/face-verification/FRONTEND_FACE_VERIFICATION_PASSPORT_GUIDE.md) |

---

## ✨ Key Features

### Core Features
- ✅ **Authentication**: JWT + Google OAuth
- ✅ **Service Booking**: Full booking lifecycle management
- ✅ **Real-time Messaging**: WebSocket-based chat
- ✅ **Payment Processing**: Stripe integration (subscriptions + bookings)
- ✅ **Reviews & Ratings**: 5-star rating system
- ✅ **Financial System**: Wallet, transactions, withdrawals

### Advanced Features
- ✅ **Face Verification**: Thai National ID + Passport support
- ✅ **Provider Tiers**: Auto-calculated based on performance
- ✅ **Schedule Management**: Provider availability system
- ✅ **Location-based Search**: Geolocation with distance calculation
- ✅ **Analytics & Reporting**: Provider dashboard metrics
- ✅ **Block & Report System**: Content moderation

### Admin Features
- ✅ **GOD Tier**: Super admin with full permissions
- ✅ **KYC Verification**: Document approval workflow
- ✅ **Financial Management**: Withdrawal approvals, commission tracking
- ✅ **User Management**: Create/delete admins, manage roles

---

## 🏗️ Tech Stack

### Backend
- **Language**: Go 1.21+
- **Framework**: Gin (HTTP router)
- **Database**: PostgreSQL 15 (pgx/v5 driver)
- **Cache**: Redis 7
- **WebSocket**: Gorilla WebSocket

### External Services
- **Payment**: Stripe
- **Storage**: Google Cloud Storage
- **OAuth**: Google Sign-In
- **Face Verification**: AWS Rekognition (planned) / Azure Face API (planned)

### DevOps
- **Containerization**: Docker + Docker Compose
- **Reverse Proxy**: Nginx (production)
- **Deployment**: Docker containers

---

## 📂 Project Structure

```
skillmatch-api/
├── main.go                         # Server entry point + all routes
├── models.go                       # Database models
├── middleware.go                   # JWT auth middleware
├── migrations.go                   # Auto-execute database migrations
├── database.go                     # Database connection setup
├── websocket_manager.go            # WebSocket connection management
│
├── *_handlers.go                   # Handler files (30+ files)
│   ├── auth_handlers.go            # Login, register, Google OAuth
│   ├── booking_handlers.go         # Booking CRUD operations
│   ├── financial_handlers.go       # Wallet, withdrawals
│   ├── god_handlers.go             # GOD tier admin operations
│   ├── message_handlers.go         # Real-time messaging
│   ├── payment_handlers.go         # Stripe integration
│   ├── provider_system_handlers.go # Provider registration & verification
│   └── ...
│
├── docker-compose.yml              # Local development (PostgreSQL + Redis)
├── docker-compose.prod.yml         # Production setup
├── Dockerfile                      # Multi-stage build
├── nginx.conf                      # Production reverse proxy config
│
├── docs/                           # Complete documentation (70+ files)
│   ├── README.md                   # Documentation index
│   ├── api-reference/              # API documentation (3 files)
│   ├── frontend-guides/            # Frontend integration (5 files)
│   ├── backend-guides/             # Backend development (3 files)
│   ├── system-guides/              # Feature-specific guides (12 files)
│   ├── face-verification/          # Face verification docs (5 files)
│   ├── sql-migrations/             # Database migrations (16 files)
│   └── sql-scripts/                # Maintenance scripts (2 files)
│
├── key/                            # GCS credentials (gitignored)
│   └── gcs-key.json
│
├── .gitignore                      # Ignored files
├── .env                            # Environment variables (gitignored)
└── go.mod / go.sum                 # Go dependencies
```

---

## 🔧 Configuration

### Environment Variables

Create `.env` file in project root:

```bash
# Authentication
JWT_SECRET_KEY=your-secret-key-min-32-chars
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret

# Stripe Payment
STRIPE_SECRET_KEY=sk_test_...
STRIPE_WEBHOOK_SECRET=whsec_...

# Google Cloud Storage
GOOGLE_APPLICATION_CREDENTIALS=key/gcs-key.json

# Database (matches docker-compose.yml)
# Connection string format:
# postgres://admin:mysecretpassword@localhost:5432/skillmatch_db?sslmode=disable
```

### Docker Compose

```yaml
services:
  postgres-db:
    image: postgres:15-alpine
    ports:
      - "5432:5432"
    environment:
      POSTGRES_USER: admin
      POSTGRES_PASSWORD: mysecretpassword
      POSTGRES_DB: skillmatch_db

  redis-cache:
    image: redis:7-alpine
    ports:
      - "6379:6379"
```

---

## 🛠️ Development

### Running the Server

```bash
# Start databases
docker-compose up -d

# Run server with auto-reload
go run .

# Build binary
go build -o skillmatch-api .

# Run binary
./skillmatch-api
```

### Database Migrations

Migrations auto-execute on server startup. Files in `docs/sql-migrations/`:

```
005_add_location_details.sql
007_add_messaging_system.sql
013_add_financial_system.sql
015_add_provider_system.sql
020_add_face_verification.sql
021_add_passport_support_face_verification.sql
...and 10 more
```

To create new migration:

```bash
# Create new file (next number: 022)
touch docs/sql-migrations/022_add_new_feature.sql

# Write SQL with BEGIN/COMMIT
# Server will auto-execute on restart
```

### Testing Endpoints

```bash
# Login
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'

# Get profile (protected)
TOKEN="your-jwt-token"
curl http://localhost:8080/profile/me \
  -H "Authorization: Bearer $TOKEN"
```

---

## 📊 API Overview

### Public Endpoints
- `POST /register` - User registration
- `POST /register/provider` - Provider registration
- `POST /login` - Email/password login
- `POST /auth/google` - Google OAuth login
- `GET /service-categories` - List service categories
- `GET /provider/:userId/public` - Public provider profile

### Protected Endpoints (JWT Required)
- `GET /profile/me` - Get my profile
- `GET /bookings/my` - Get my bookings
- `GET /conversations` - Get my conversations
- `POST /messages` - Send message
- `GET /wallet` - Get wallet balance
- `POST /withdrawals` - Request withdrawal

### Admin Endpoints (Admin Only)
- `GET /admin/stats/god` - GOD dashboard statistics
- `GET /admin/users` - List all users
- `POST /admin/admins` - Create new admin
- `GET /admin/withdrawals` - Pending withdrawals

### GOD Endpoints (Tier 5 Only)
- `POST /god/update-user` - Update any user's role/tier
- `DELETE /god/users/:userId` - Delete any user
- `POST /god/view-mode` - Switch UI view mode

**Complete API Reference:** [`docs/api-reference/API_REFERENCE_FOR_FRONTEND.md`](./docs/api-reference/API_REFERENCE_FOR_FRONTEND.md)

---

## 🔐 Security

### Authentication
- **JWT Tokens**: 7-day expiration (configurable)
- **Password Hashing**: bcrypt (cost 10)
- **OAuth 2.0**: Google Sign-In integration

### Authorization
- **Role-based Access**: User, Provider, Admin, GOD
- **Tier-based Permissions**: 5 tiers (General, Silver, Gold, Platinum, GOD)
- **GOD Protection**: user_id = 1 cannot be modified by others

### Data Protection
- **PII Encryption**: Sensitive data encrypted at rest
- **PDPA Compliance**: Data retention policies
- **API Rate Limiting**: Prevents abuse
- **CORS**: Configured for production

**Full Security Guide:** [`docs/backend-guides/SECURITY.md`](./docs/backend-guides/SECURITY.md)

---

## 🚀 Deployment

### Docker Production

```bash
# Build image
docker build -t skillmatch-api .

# Run with production compose
docker-compose -f docker-compose.prod.yml up -d
```

### Binary Deployment

```bash
# Build for Linux
GOOS=linux GOARCH=amd64 go build -o skillmatch-api

# Upload to server and run
./skillmatch-api
```

### Environment Setup

```bash
# Production environment variables
export GIN_MODE=release
export JWT_SECRET_KEY=<production-secret>
export STRIPE_SECRET_KEY=<production-key>
# ... other production credentials
```

**Deployment Guide:** [`DEPLOYMENT.md`](./DEPLOYMENT.md)

---

## 📈 Status & Roadmap

### Current Status

| Component | Status | Notes |
|-----------|--------|-------|
| Authentication | ✅ Production | JWT + Google OAuth |
| Booking System | ✅ Production | Full lifecycle |
| Messaging | ✅ Production | WebSocket real-time |
| Payment | ✅ Production | Stripe integration |
| Face Verification | ⚠️ Mock Mode | Needs AWS/Azure |
| Financial System | ✅ Production | Wallet + Withdrawals |

### Roadmap

**High Priority:**
- [ ] AWS Rekognition integration (real face matching)
- [ ] Liveness detection for face verification
- [ ] Automated testing suite
- [ ] Payment reconciliation system

**Medium Priority:**
- [ ] Multi-language support (EN/TH/CN)
- [ ] Email notification system
- [ ] Advanced analytics dashboard
- [ ] Mobile app optimization

**See:** [`TODO.md`](./TODO.md) for complete task list

---

## 🤝 Contributing

### Development Workflow

1. **Fork repository**
2. **Create feature branch**: `git checkout -b feature/amazing-feature`
3. **Commit changes**: `git commit -m 'Add amazing feature'`
4. **Push to branch**: `git push origin feature/amazing-feature`
5. **Open Pull Request**

### Code Style

- Follow Go conventions (`gofmt`)
- Document all exported functions
- Add comments for complex logic
- Update relevant documentation

### Pull Request Guidelines

- ✅ All tests pass
- ✅ Documentation updated
- ✅ No breaking changes (or clearly marked)
- ✅ Database migrations included (if needed)

---

## 📝 License

This project is proprietary software. All rights reserved.

---

## 👥 Team

- **Backend Team**: Go API development
- **Frontend Team**: React/TypeScript UI
- **DevOps Team**: Infrastructure & deployment

---

## 📞 Support

### Documentation
- **Index**: [`docs/README.md`](./docs/README.md)
- **API Reference**: [`docs/api-reference/`](./docs/api-reference/)
- **System Guides**: [`docs/system-guides/`](./docs/system-guides/)

### Issues & Questions
- **GitHub Issues**: [Report bugs](https://github.com/your-org/skillmatch-api/issues)
- **Documentation Updates**: Submit PR to relevant guide

---

**Version:** 1.1.0  
**Last Updated:** November 25, 2025  
**Database Version:** Migration 021
