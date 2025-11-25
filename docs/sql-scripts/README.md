# 🛠️ SQL Scripts

โฟลเดอร์นี้เก็บ SQL scripts สำหรับ maintenance, fixes, และ seed data

---

## 📋 Available Scripts (2 files)

### 1. `fix_god_profile.sql`
**Purpose:** Fix GOD account profile data  
**Use Case:** Reset GOD account to correct state  
**Run When:** GOD profile shows incorrect/mock data

**Description:**
อัปเดต GOD account (user_id = 1) ให้แสดงข้อมูลจริงจาก database แทนที่จะเป็น mock data

**How to Run:**
```bash
# Connect to database
psql -U admin -d skillmatch_db -h localhost -p 5432

# Run script
\i docs/sql-scripts/fix_god_profile.sql

# Verify
SELECT user_id, username, email, tier_id FROM users WHERE user_id = 1;
```

**What it does:**
- Ensures GOD account (user_id = 1) has correct profile data
- Verifies tier_id = 5 (GOD tier)
- Fixes any corrupted profile fields

---

### 2. `seed_providers.sql`
**Purpose:** Seed demo provider data for testing  
**Use Case:** Populate database with sample providers  
**Run When:** Setting up dev/test environment

**Description:**
สร้างข้อมูล providers ตัวอย่างสำหรับทดสอบระบบ

**How to Run:**
```bash
# Connect to database
psql -U admin -d skillmatch_db -h localhost -p 5432

# Run script
\i docs/sql-scripts/seed_providers.sql

# Verify
SELECT COUNT(*) FROM users WHERE verification_status = 'verified';
```

**What it creates:**
- Demo provider accounts with various tiers
- Sample service categories
- Test packages and schedules
- Mock reviews and ratings

**⚠️ Warning:**
Only run in development/test environments - NOT in production!

---

## 🚀 How to Use Scripts

### Option 1: Direct psql
```bash
# Connect to database
psql -U admin -d skillmatch_db -h localhost -p 5432

# Run script
\i docs/sql-scripts/fix_god_profile.sql

# Exit
\q
```

### Option 2: Command Line
```bash
# Run script directly
psql -U admin -d skillmatch_db -h localhost -p 5432 -f docs/sql-scripts/fix_god_profile.sql
```

### Option 3: Docker Exec
```bash
# Copy script to container
docker cp docs/sql-scripts/fix_god_profile.sql postgres:/tmp/

# Execute inside container
docker exec -it postgres psql -U admin -d skillmatch_db -f /tmp/fix_god_profile.sql
```

---

## 📝 Creating New Scripts

### Step 1: Create SQL File
```bash
touch docs/sql-scripts/my_maintenance_script.sql
```

### Step 2: Write Script with Safety Checks
```sql
-- my_maintenance_script.sql
-- Description: What this script does
-- Author: Your name
-- Date: 2025-11-21

BEGIN;

-- Add safety check
DO $$
BEGIN
    -- Your logic here
    IF (SELECT COUNT(*) FROM users WHERE tier_id = 5) > 1 THEN
        RAISE EXCEPTION 'Multiple GOD accounts detected! Aborting.';
    END IF;
END $$;

-- Main script logic
UPDATE users SET some_field = 'value' WHERE condition;

COMMIT;
```

### Script Best Practices
- ✅ Use `BEGIN;` and `COMMIT;` for transactions
- ✅ Add comments explaining what script does
- ✅ Include safety checks (prevent data corruption)
- ✅ Test on dev database first
- ✅ Add rollback instructions in comments
- ✅ Log what was changed
- ❌ Never run unverified scripts in production

---

## 🗂️ Script Categories

### **Maintenance Scripts**
- `fix_god_profile.sql` - Fix GOD account data

### **Seed Data Scripts**
- `seed_providers.sql` - Populate demo providers

### **Future Scripts** (add as needed)
- Data cleanup scripts
- Performance optimization queries
- Batch update scripts
- Report generation queries

---

## 🔍 Common Use Cases

### Fix Data Corruption
```bash
# Example: Fix corrupted user profiles
psql -U admin -d skillmatch_db -f docs/sql-scripts/fix_user_profiles.sql
```

### Seed Test Data
```bash
# Populate database with demo data
psql -U admin -d skillmatch_db -f docs/sql-scripts/seed_providers.sql
```

### Batch Updates
```bash
# Update multiple records at once
psql -U admin -d skillmatch_db -f docs/sql-scripts/batch_update_tiers.sql
```

### Data Cleanup
```bash
# Remove old/stale data
psql -U admin -d skillmatch_db -f docs/sql-scripts/cleanup_old_data.sql
```

---

## ⚠️ Safety Guidelines

### Before Running Scripts

1. **Backup Database**
```bash
# Create backup
docker exec -t postgres pg_dump -U admin skillmatch_db > backup_$(date +%Y%m%d).sql
```

2. **Test on Dev First**
```bash
# Run on test database
psql -U admin -d skillmatch_db_test -f docs/sql-scripts/my_script.sql
```

3. **Review Script**
- Read entire script before running
- Understand what it does
- Check for potential data loss

4. **Verify Results**
```sql
-- After running script, verify changes
SELECT * FROM affected_table WHERE condition;
```

### Rollback Strategy

If script causes issues:

```bash
# Restore from backup
psql -U admin -d skillmatch_db < backup_20251121.sql

# Or use transaction rollback (if script failed mid-execution)
# psql automatically rolls back failed transactions
```

---

## 📊 Script Statistics

```
Total Scripts: 2 files
Categories: Maintenance (1), Seed Data (1)
Total Size: ~5KB
Location: docs/sql-scripts/
```

---

## 🆚 Scripts vs Migrations

### When to use **Migrations**
- ✅ Schema changes (CREATE TABLE, ALTER TABLE)
- ✅ Adding indexes, constraints
- ✅ Auto-executed on server start
- ✅ Versioned and sequential

### When to use **Scripts**
- ✅ Data fixes/updates
- ✅ Seed data
- ✅ Maintenance tasks
- ✅ Manual execution only
- ✅ One-time operations

---

## 🔗 Related Documentation

- **Database Migrations**: [`docs/sql-migrations/README.md`](../sql-migrations/README.md)
- **Database Schema**: [`docs/backend-guides/DATABASE_STRUCTURE.md`](../backend-guides/DATABASE_STRUCTURE.md)
- **Migration Code**: `migrations.go`

---

**Last Updated:** November 21, 2025  
**Total Scripts:** 2 files  
**Location:** docs/sql-scripts/
