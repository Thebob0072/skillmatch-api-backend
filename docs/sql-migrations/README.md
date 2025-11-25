# 🗄️ Database Migrations

โฟลเดอร์นี้เก็บ migration files ทั้งหมดที่ใช้สร้างและอัปเดต database schema

---

## 📋 Migration Files (16 files)

### Core System Migrations

| File | Description | Tables Created |
|------|-------------|----------------|
| `005_add_location_details.sql` | Location fields | provinces, districts, service_areas |
| `006_add_service_type.sql` | Service type enum | users.service_type |
| `007_add_messaging_system.sql` | Real-time chat | conversations, messages |
| `008_add_notifications_system.sql` | Push notifications | notifications |
| `009_add_reports_system.sql` | User/content reports | reports |
| `010_add_profile_views.sql` | Profile view tracking | profile_views |
| `011_add_blocks_system.sql` | Block system | blocks |

### Service & Category System

| File | Description | Tables Created |
|------|-------------|----------------|
| `012_add_service_categories.sql` | Service categories | service_categories, user_service_categories |

### Financial System

| File | Description | Tables Created |
|------|-------------|----------------|
| `013_add_financial_system.sql` | Wallet & transactions | transactions, withdrawal_requests |
| `016_add_platform_bank_tracking.sql` | Bank account tracking | platform_bank_accounts |
| `017_add_god_commission_tracking.sql` | GOD commission system | users.god_commission_balance |
| `018_update_fee_structure.sql` | Fee calculation updates | transactions fee fields |

### Authentication & Verification

| File | Description | Tables Created |
|------|-------------|----------------|
| `014_add_email_verification.sql` | Email verification | email_verification_codes |
| `020_add_face_verification.sql` | Face biometric verification | face_verification_requests |
| `021_add_passport_support_face_verification.sql` | Passport support | document_type, document_id columns |

### Provider System

| File | Description | Tables Created |
|------|-------------|----------------|
| `015_add_provider_system.sql` | Provider features | packages, schedules, provider tier fields |

---

## 🔢 Migration Numbering

Migrations are numbered sequentially starting from `005`:
- `001-004`: Initial schema (in migrations.go)
- `005-021`: SQL migration files (this folder)

**Next migration number:** `022`

---

## 🚀 How Migrations Work

### Auto-execution
All migrations in this folder are automatically executed by `runMigrations()` in `migrations.go` when the server starts:

```go
func runMigrations(db *sql.DB) error {
    // Execute migration files from migrations/ folder
    files, _ := ioutil.ReadDir("./migrations")
    for _, file := range files {
        // Run each .sql file
    }
}
```

### Manual Execution
If you need to run migrations manually:

```bash
# Connect to PostgreSQL
psql -U admin -d skillmatch_db -h localhost -p 5432

# Run specific migration
\i docs/sql-migrations/005_add_location_details.sql

# Check migration status
SELECT * FROM schema_migrations;
```

---

## 📝 Creating New Migrations

### Step 1: Create SQL File
```bash
# Create new migration file
touch docs/sql-migrations/022_add_new_feature.sql
```

### Step 2: Write Migration SQL
```sql
-- 022_add_new_feature.sql
BEGIN;

CREATE TABLE IF NOT EXISTS new_table (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add indexes
CREATE INDEX idx_new_table_name ON new_table(name);

COMMIT;
```

### Step 3: Server Auto-runs
Restart server → migrations auto-execute

### Migration Best Practices
- ✅ Use `BEGIN;` and `COMMIT;` for transactions
- ✅ Always use `IF NOT EXISTS` for safety
- ✅ Add indexes for frequently queried columns
- ✅ Include rollback plan in comments
- ✅ Test on dev database first
- ❌ Never edit executed migrations (create new ones)

---

## 🗂️ Migration Categories

### **Location & Geography** (1 file)
- `005_add_location_details.sql`

### **Communication** (3 files)
- `007_add_messaging_system.sql`
- `008_add_notifications_system.sql`
- `009_add_reports_system.sql`

### **User Management** (3 files)
- `010_add_profile_views.sql`
- `011_add_blocks_system.sql`
- `014_add_email_verification.sql`

### **Provider System** (3 files)
- `006_add_service_type.sql`
- `012_add_service_categories.sql`
- `015_add_provider_system.sql`

### **Financial System** (4 files)
- `013_add_financial_system.sql`
- `016_add_platform_bank_tracking.sql`
- `017_add_god_commission_tracking.sql`
- `018_update_fee_structure.sql`

### **Security & Verification** (2 files)
- `020_add_face_verification.sql`
- `021_add_passport_support_face_verification.sql`

---

## 🔍 Find Migration by Feature

### Need to add/modify messaging?
→ `007_add_messaging_system.sql`

### Need to add/modify wallet/transactions?
→ `013_add_financial_system.sql`

### Need to add/modify face verification?
→ `020_add_face_verification.sql` + `021_add_passport_support_face_verification.sql`

### Need to add/modify provider features?
→ `015_add_provider_system.sql`

### Need to add/modify service categories?
→ `012_add_service_categories.sql`

---

## ⚠️ Important Notes

### DO NOT Edit Executed Migrations
Once a migration has been run in production:
- ❌ **NEVER** edit the existing file
- ✅ **CREATE** a new migration to modify the schema

Example:
```
Bad:  Edit 015_add_provider_system.sql
Good: Create 022_modify_provider_system.sql
```

### Rollback Strategy
Migrations don't auto-rollback. To revert:

1. **Create rollback migration**:
```sql
-- 023_rollback_feature.sql
BEGIN;

DROP TABLE IF EXISTS new_table;

COMMIT;
```

2. **Or restore from backup**:
```bash
pg_restore -U admin -d skillmatch_db backup.sql
```

---

## 📊 Migration Statistics

```
Total Migrations: 16 files
Numbering: 005-021
Total Size: ~50KB
Execution: Auto on server start
Location: docs/sql-migrations/
```

---

## 🔗 Related Documentation

- **Database Schema**: [`docs/backend-guides/DATABASE_STRUCTURE.md`](../backend-guides/DATABASE_STRUCTURE.md)
- **Migration Code**: `migrations.go`
- **Docker Database**: `docker-compose.yml`

---

**Last Updated:** November 21, 2025  
**Total Migrations:** 16 files  
**Next Migration:** 022
