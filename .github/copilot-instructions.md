# SkillMatch API - Copilot Instructions

## Project Architecture

**SkillMatch** is a Go-based marketplace API for service providers and clients, built with Gin framework, PostgreSQL, Redis, and real-time WebSocket communication. The system manages user profiles, bookings, messaging, payments (Stripe), financial transactions, and provider verification.

### Core Stack
- **Framework**: Gin (HTTP), pgx/v5 (PostgreSQL), go-redis, Gorilla WebSocket
- **Storage**: PostgreSQL (primary), Redis (caching), Google Cloud Storage (files)
- **External Services**: Stripe (payments), Google OAuth
- **Deployment**: Docker Compose (postgres:15-alpine, redis:7-alpine)

### Key Files
- `main.go` - Server bootstrap, all route definitions, dependency injection
- `models.go` - Go structs for database entities (User, Booking, Review, etc.)
- `database.go` - Global `db *sql.DB` connection (used by message/notification handlers)
- `migrations.go` - Database schema setup, runs on startup
- `middleware.go` - `authMiddleware()` (JWT validation) and `adminAuthMiddleware()`
- `websocket_manager.go` - WebSocket connection management and broadcasting

## Handler Pattern (Critical Convention)

**Every handler returns `gin.HandlerFunc`** and accepts dependencies as parameters. This pattern enables dependency injection from `main.go`:

```go
// Definition (in handler file)
func createBookingHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Access userID from middleware
        userID := c.GetInt("userID")
        
        // Business logic here
    }
}

// Usage (in main.go)
protected.POST("/bookings", createBookingHandler(dbPool, ctx))
```

**Always follow this pattern** - never create handlers that accept `*gin.Context` directly as the function signature.

## Authentication & Authorization

### JWT Authentication
- Middleware sets `c.Set("userID", userID)` - retrieve with `c.GetInt("userID")` or `c.Get("userID")`
- JWT secret stored in `jwtKey` (from `JWT_SECRET_KEY` env var)
- Token format: `Authorization: Bearer <token>`
- `createJWT(userID)` helper in `auth_handlers.go`

### User Roles & Tiers
- **Subscription Tier** (`tier_id` in users): General (free), Silver, Gold, Platinum - client subscription level
- **Provider Level** (`provider_level_id` in users): Separate tier system for providers (auto-calculated by performance)
- **Admin Flag** (`is_admin`): Boolean for admin access
- **GOD Tier**: tier_id = 5, access_level = 999 (super admin)

### GOD Account Protection (Critical Security Rules)
- **Hard-coded Protection**: user_id = 1 is the GOD account - **NEVER** allow modification/deletion by anyone except self
- **Double Verification**: GOD endpoints require BOTH `is_admin = true` AND `tier_id = 5` (must check both fields)
- **Self-Protection**: In `updateUserHandler` and `deleteAdminHandler`, check `if req.UserID == 1 && requesterID.(int) != 1` → return 403
- **View Mode System**: GOD can preview UI as different roles (user/provider/admin) without changing actual role (stored in `godViewModes` map)
- **Endpoint Protection**: All GOD endpoints must verify `requesterTierID == 5` before allowing operations
- **Database Prevention**: Never expose GOD credentials in logs, never allow password reset via public endpoints
- **Admin Creation**: Only GOD (tier_id = 5) can create new admins via POST `/admin/god/create-admin`
- **Critical Rule**: When querying for calculations (provider tiers, stats), always exclude GOD tier with `WHERE tier_id < 5`

### Verification Status Flow
- `unverified` → `pending` (documents submitted) → `approved`/`rejected`
- Only `approved` or `verified` users visible as providers

## Database Patterns

### Connection Management
- **Two connection pools exist**:
  1. `dbPool *pgxpool.Pool` - Modern pgx driver (most handlers)
  2. `db *sql.DB` (global) - Legacy sql.DB (message_handlers.go, notification_handlers.go, report_handlers.go)
- Initialize global `db` with `InitDatabase(connString)` in `main.go`
- Prefer `dbPool` for new code

### Query Conventions
- Use `$1, $2...` placeholders (PostgreSQL style)
- Arrays in Go: `[]string` maps to PostgreSQL `TEXT[]`
- Always use `COALESCE(array_column, '{}')` when scanning arrays to avoid null pointer errors
- Geographic data: `DECIMAL(10,8)` for lat/long

### Common JOINs
```sql
-- Provider profile with tier and stats
SELECT u.*, t.name as tier_name, AVG(r.rating), COUNT(r.review_id)
FROM users u
LEFT JOIN tiers t ON u.provider_level_id = t.tier_id  -- Note: provider_level_id, not tier_id
LEFT JOIN reviews r ON u.user_id = r.provider_id
WHERE u.verification_status IN ('verified', 'approved')
GROUP BY u.user_id, t.name
```

## Real-Time Communication

### WebSocket Architecture
- **Public endpoint**: `GET /ws` (no auth header required)
- **Auth flow**: Connect first → send `{"type":"auth","payload":{"token":"..."}}`
- Global `wsManager` broadcasts messages to online users
- Use `wsManager.BroadcastToUser(userID, message)` to notify users
- **Always broadcast** when: new message sent, booking status changed, notification created

### WebSocket Message Types
```go
WebSocketMessage{
    Type: "new_message" | "typing" | "notification" | "booking_update" | "auth",
    Payload: interface{} // Type-specific data
}
```

### Conversation Management
- Conversations enforce `user1_id < user2_id` (prevents duplicates)
- Use helper function to normalize IDs before INSERT
- Query pattern: Join messages with `unread_count` aggregation

### Messaging Endpoints
- **POST** `/messages` - Send message (creates conversation if doesn't exist)
- **GET** `/conversations` - List user's conversations with unread counts
- **GET** `/conversations/:id/messages` - Get messages (supports pagination)
- **PATCH** `/messages/read` - Mark messages as read
- **DELETE** `/messages/:id` - Delete message
- All messaging uses global `db *sql.DB` connection (not `dbPool`)
- **Important Restriction**: Users can only send automated/templated messages - direct contact exchange is not allowed

## Feature-Specific Patterns

### Provider System
- **Registration**: POST `/register/provider` requires:
  - Basic auth: `username`, `email`, `password`, `gender_id`
  - Verification: `phone`, `otp` (6 digits, 10-min validity)
  - Provider data: `category_ids` (array, 1-5 categories), `service_type` ("Incall"/"Outcall"/"Both")
  - Optional: `bio`, location fields
- **Documents**: Upload to GCS via `startPhotoUploadHandler` pattern (presigned URL)
  - Required: `national_id`, `health_certificate`
  - Optional: `business_license`, `portfolio`, `certification`
- **Photo Gallery System**:
  - Public endpoint: `GET /provider/:userId/photos` (no auth required)
  - Returns array sorted by `sort_order ASC`
  - Each photo has: `photo_id`, `photo_url` (relative path), `sort_order`, `caption` (nullable), `uploaded_at`
  - Unlimited photos per provider
  - First photo serves as cover photo
  - `profile_image_url` vs `photos[]`: Profile picture is main avatar, photos array is full gallery
- **Tier Calculation**: Admin triggers via POST `/admin/recalculate-provider-tiers`
  - Points formula (max 600): `(avg_rating * 20) + (bookings * 5) + (reviews * 3) + (response_rate * 0.5) + (acceptance_rate * 0.5)`
  - Tiers: General (0-99), Silver (100-249), Diamond (250-399), Premium (400+)
- **Verification Flow**: `unverified` → `documents_submitted` → Admin reviews → `approved`/`rejected`

### Financial System (Stripe Integration)
- **Payment Flow**: Client pays full price → Stripe webhook → calculate fees → update provider wallet
- **Fee Structure (Total 12.75%)**:
  - Stripe Payment Gateway: 2.75%
  - Platform Commission: 10%
  - Provider receives: 87.25% of booking price
- **Fee Visibility**: Only providers see fee breakdown (12.75% total) - clients pay full price without seeing fees
- **Account Creation**: When provider registers, display fee notification (12.75% will be deducted from earnings)
- **Commission Storage**: 10% retained in GOD's platform bank account, 87.25% transferred to provider (after Stripe fee)
- **Wallet States**: `pending_balance` (7-day hold) → `available_balance` (withdrawable)
- **Withdrawals**: Provider requests → Admin approves → **GOD transfers 87.25% via platform bank account** → Mark as `completed`
- **Platform Bank Account**: All withdrawals must go through GOD's verified platform bank account for tracking and security
- **Transfer Flow**: Provider wallet → Platform bank (GOD) → Provider's personal bank (87.25% only)
- **Slip Masking**: Transfer slips must mask GOD account details before sending to provider (via chat/email)
- **Notification**: Send masked slip via WebSocket (real-time chat) and Email after withdrawal completion
- Use `setupStripe()` before any payment operations

### Booking Lifecycle
- **States**: `pending` → `paid` → `confirmed` → `completed` / `cancelled`
- **Payment Flow**: Client creates booking → redirected to Stripe Checkout → payment success → booking status = `paid`
- **Price Calculation**: `package.price` → `booking.total_price` (store at booking time)
- **Payment Release**: Provider receives payment only after booking status = `completed` (job finished and confirmed)
- **Review Creation**: Only allowed after booking status = `completed`

### Payment Types
1. **Subscription Payment** (Client upgrades tier):
   - Flow: Client → Stripe Checkout (subscription mode) → Webhook updates `users.tier_id`
   - Current Status: ✅ Implemented
   - File: `payment_handlers.go`

2. **Booking Payment** (Client books provider):
   - Flow: Client → Stripe Checkout (payment mode) → Webhook creates transaction → Updates wallet
   - Current Status: ⚠️ Needs Implementation
   - Required: Create booking checkout handler + webhook for booking payments

## Environment Variables (Required)

```bash
# Authentication
JWT_SECRET_KEY=<your-secret>
GOOGLE_CLIENT_ID=<google-oauth-client-id>
GOOGLE_CLIENT_SECRET=<google-oauth-secret>

# Stripe
STRIPE_SECRET_KEY=<stripe-secret>
STRIPE_WEBHOOK_SECRET=<webhook-secret>

# Google Cloud Storage
GOOGLE_APPLICATION_CREDENTIALS=key/gcs-key.json

# Database (matches docker-compose.yml)
# postgres://admin:mysecretpassword@localhost:5432/skillmatch_db?sslmode=disable
```

## Development Workflow

### Running the Project
```bash
# Start databases
docker-compose up -d

# Run server (migrations auto-run on startup)
go run .

# Build binary
go build -o skillmatch-api
```

### Adding New Features
1. **Define model** in `models.go` if database-backed
2. **Add migration** to `runMigrations()` in `migrations.go`
3. **Create handler** using the handler pattern (dependency injection)
4. **Register route** in `main.go` under appropriate group (public/protected/admin)
5. **Update documentation** in relevant `*_GUIDE.md` file

### Testing Endpoints
- Health check: `GET /ping`
- Database time: Returns `postgres_time` to verify DB connection
- Use `Authorization: Bearer <token>` for protected routes

## Input Validation Rules

### Authentication Fields
- **Username**: 3-50 chars, alphanumeric + underscore/dash, unique
- **Email**: Valid format, unique
- **Password**: Min 8 chars (recommend: uppercase, lowercase, number, special char)
- **Phone**: Exactly 10 digits (Thai format, e.g., "0812345678")
- **OTP**: Exactly 6 digits, valid for 10 minutes
- **Gender ID**: 1=Male, 2=Female, 3=Other, 4=Prefer not to say

### Provider Fields
- **Category IDs**: Array of 1-5 valid category IDs
- **Service Type**: Must be exactly "Incall", "Outcall", or "Both" (case-sensitive)
- **Bio**: Max 1000 characters
- **Document Types**: 
  - Required: "national_id", "health_certificate"
  - Optional: "business_license", "portfolio", "certification"

### Geographic Data
- **Latitude/Longitude**: DECIMAL(10,8)
- **Province/District**: VARCHAR, matches Thai administrative divisions

## Common Pitfalls

❌ **DON'T** create handlers like `func myHandler(c *gin.Context)`  
✅ **DO** create handlers like `func myHandler(dbPool *pgxpool.Pool, ctx context.Context) gin.HandlerFunc`

❌ **DON'T** forget to scan array columns with `COALESCE(column, '{}')`  
✅ **DO** always use `COALESCE` for nullable arrays

❌ **DON'T** forget WebSocket broadcasts for real-time features  
✅ **DO** call `wsManager.BroadcastToUser()` after state changes

❌ **DON'T** mix up `tier_id` (subscription) and `provider_level_id` (provider tier)  
✅ **DO** join correct tier FK based on context

❌ **DON'T** accept case-insensitive service_type values  
✅ **DO** validate exact case: "Incall", "Outcall", "Both"

❌ **DON'T** allow more than 5 categories per provider  
✅ **DO** validate category_ids array length (1-5)

## Public vs Protected Endpoints

### Truly Public (No auth middleware)
- Authentication endpoints (`/register`, `/login`, `/auth/google`)
- Public profile views (`/provider/:userId/public`)
- Provider resources:
  - `GET /provider/:userId/photos` - Gallery photos sorted by sort_order, includes caption
  - `GET /packages/:providerId` - Service packages
  - `GET /reviews/:providerId` - Reviews with pagination
  - `GET /reviews/stats/:providerId` - Rating statistics
- Service categories (`/service-categories`)
- **Favorites check** (`/favorites/check/:providerId`) - Returns `false` if no token, actual status if logged in

### Protected (Requires JWT token)
- Full profile access (`/provider/:userId`) - includes age, height, service_type
- All user management (`/users/me`, `/profile/me`)
- Bookings, favorites (add/remove), messaging
- Provider-specific endpoints (`/provider/documents`, `/provider/my-tier`)

### Admin Only (Requires JWT + is_admin)
- All `/admin/*` endpoints - KYC approval, financial management, provider verification

## Documentation Files

Refer to these guides for implementation details:
- `DATABASE_STRUCTURE.md` - Complete schema reference
- `MESSAGING_GUIDE.md` - Real-time chat implementation
- `FINANCIAL_SYSTEM_GUIDE.md` - Payment and wallet flows
- `PROVIDER_SYSTEM_GUIDE.md` - Provider verification and tiers
- `API_REFERENCE_FOR_FRONTEND.md` - Complete endpoint documentation
