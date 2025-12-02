-- ============================================================================
-- DATABASE CLEANUP SCRIPT - SkillMatch API Backend
-- ============================================================================
-- Purpose: Remove redundant columns, merge duplicate tables, optimize indexes
-- Date: 2025-12-02
-- WARNING: BACKUP DATABASE BEFORE RUNNING!
-- ============================================================================

BEGIN;

-- ============================================================================
-- SECTION 1: MERGE PROFILE PICTURES (3 columns → 1 column)
-- ============================================================================
-- Strategy: Keep users.profile_picture_url, drop others
-- Timeline: Google OAuth already uses profile_picture_url

-- Step 1.1: Migrate data from google_profile_picture to profile_picture_url (if any exists)
UPDATE users 
SET profile_picture_url = google_profile_picture 
WHERE profile_picture_url IS NULL 
  AND google_profile_picture IS NOT NULL;

-- Step 1.2: Migrate data from user_profiles.profile_image_url to users.profile_picture_url
UPDATE users u
SET profile_picture_url = COALESCE(u.profile_picture_url, p.profile_image_url)
FROM user_profiles p
WHERE u.user_id = p.user_id 
  AND u.profile_picture_url IS NULL 
  AND p.profile_image_url IS NOT NULL;

-- Step 1.3: Drop redundant columns
ALTER TABLE users DROP COLUMN IF EXISTS google_profile_picture;
ALTER TABLE user_profiles DROP COLUMN IF EXISTS profile_image_url;

COMMENT ON COLUMN users.profile_picture_url IS 'Single source of truth for profile pictures (Google OAuth or uploaded)';

-- ============================================================================
-- SECTION 2: CLEAN UP DUPLICATE INDEXES
-- ============================================================================
-- PostgreSQL UNIQUE constraints automatically create indexes
-- Drop manual indexes if constraint already exists

DROP INDEX IF EXISTS email_idx; -- users_email_key constraint already exists
DROP INDEX IF EXISTS google_id_idx; -- users_google_id_key constraint already exists

-- ============================================================================
-- SECTION 3: ANALYZE provider_availability vs provider_schedules
-- ============================================================================
-- Decision: KEEP BOTH (different purposes)
-- - provider_availability: Recurring weekly schedule (Mon-Sun, time slots)
-- - provider_schedules: Specific bookings + calendar events
-- Both are empty now, but serve different purposes

COMMENT ON TABLE provider_availability IS 'Provider recurring weekly schedule (e.g., Mon 9AM-5PM, Tue 10AM-2PM)';
COMMENT ON TABLE provider_schedules IS 'Provider specific calendar events: bookings, blocked times, custom availability';

-- ============================================================================
-- SECTION 4: CONSOLIDATE LOCATION DATA
-- ============================================================================
-- Current state:
-- - user_profiles.location (VARCHAR 255) - simple text field ✅ Currently used
-- - provider_schedules has location_province, location_district - structured ⚠️ Not used yet
-- - bookings.location (TEXT) - per-booking location ✅ Currently used

-- Decision: Keep all for now, but standardize format
-- Future: Consider adding separate address table for structured data

-- Add index for location search in user_profiles
CREATE INDEX IF NOT EXISTS idx_user_profiles_location 
ON user_profiles USING gin (location gin_trgm_ops);

COMMENT ON COLUMN user_profiles.location IS 'Free-text location (e.g., "Bangkok, Sukhumvit", "Chiang Mai"). Searchable with ILIKE/trigram.';
COMMENT ON COLUMN provider_schedules.location_province IS 'Structured province for specific booking location';
COMMENT ON COLUMN provider_schedules.location_district IS 'Structured district for specific booking location';

-- ============================================================================
-- SECTION 5: OPTIMIZE user_profiles TABLE
-- ============================================================================
-- Remove unused/redundant columns if any

-- Check if skills column is actually used (currently TEXT[] array)
-- If skills are never used for providers (only categories), can drop
-- For sex workers, categories are more important than general skills
-- Decision: KEEP skills as optional field for additional talents

-- Optimize working_hours - should match provider_schedules
COMMENT ON COLUMN user_profiles.working_hours IS 'Deprecated: Use provider_schedules or provider_availability instead';

-- ============================================================================
-- SECTION 6: ADD MISSING INDEXES FOR PERFORMANCE
-- ============================================================================

-- Bookings performance
CREATE INDEX IF NOT EXISTS idx_bookings_created_at ON bookings(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_bookings_completed_at ON bookings(completed_at DESC) WHERE completed_at IS NOT NULL;

-- Reviews performance
CREATE INDEX IF NOT EXISTS idx_reviews_created_at ON reviews(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_reviews_rating ON reviews(rating);

-- User profiles search
CREATE INDEX IF NOT EXISTS idx_user_profiles_service_type ON user_profiles(service_type) WHERE service_type IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_user_profiles_available ON user_profiles(is_available) WHERE is_available = true;

-- Provider categories search
CREATE INDEX IF NOT EXISTS idx_provider_categories_category ON provider_categories(category_id);

-- Transactions performance
CREATE INDEX IF NOT EXISTS idx_transactions_created_at ON transactions(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_transactions_type ON transactions(transaction_type);

-- ============================================================================
-- SECTION 7: ADD CONSTRAINTS FOR DATA INTEGRITY
-- ============================================================================

-- Ensure provider_level_id and tier_id reference correct tiers
-- provider_level_id: 1-4 (General, Silver, Diamond, Premium) - auto-calculated
-- tier_id: 1-5 (1=General client, 2=Silver client, 3=Gold client, 4=Platinum client, 5=GOD)

-- Add check constraint for verification_status
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_verification_status_check;
ALTER TABLE users ADD CONSTRAINT users_verification_status_check 
  CHECK (verification_status IN ('unverified', 'pending', 'documents_submitted', 'approved', 'verified', 'rejected'));

-- Add check constraint for service_type
ALTER TABLE user_profiles DROP CONSTRAINT IF EXISTS user_profiles_service_type_check;
ALTER TABLE user_profiles ADD CONSTRAINT user_profiles_service_type_check 
  CHECK (service_type IN ('Incall', 'Outcall', 'Both') OR service_type IS NULL);

-- ============================================================================
-- SECTION 8: CLEANUP UNUSED DATA
-- ============================================================================

-- Remove test/invalid users if any (keep GOD account user_id=1)
-- DELETE FROM users WHERE user_id > 1 AND verification_status = 'unverified' AND registration_date < NOW() - INTERVAL '30 days';

-- Remove old notifications (keep last 90 days)
-- DELETE FROM notifications WHERE created_at < NOW() - INTERVAL '90 days' AND is_read = true;

-- Remove old messages (keep last 1 year)
-- DELETE FROM messages WHERE created_at < NOW() - INTERVAL '1 year';

-- ============================================================================
-- SECTION 9: VACUUM AND ANALYZE
-- ============================================================================

VACUUM ANALYZE users;
VACUUM ANALYZE user_profiles;
VACUUM ANALYZE bookings;
VACUUM ANALYZE reviews;
VACUUM ANALYZE service_packages;
VACUUM ANALYZE provider_categories;
VACUUM ANALYZE transactions;
VACUUM ANALYZE wallets;
VACUUM ANALYZE notifications;
VACUUM ANALYZE messages;

COMMIT;

-- ============================================================================
-- POST-CLEANUP VERIFICATION QUERIES
-- ============================================================================

-- Check duplicate indexes (should be clean now)
SELECT 
  schemaname, tablename, indexname
FROM pg_indexes 
WHERE schemaname = 'public' 
  AND indexname LIKE '%email%' OR indexname LIKE '%google_id%'
ORDER BY tablename;

-- Check table sizes after cleanup
SELECT 
  schemaname, tablename, 
  pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size,
  pg_size_pretty(pg_indexes_size(schemaname||'.'||tablename)) AS indexes_size
FROM pg_tables 
WHERE schemaname = 'public' 
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- Verify profile_picture_url migration
SELECT COUNT(*) as users_with_profile_pic 
FROM users 
WHERE profile_picture_url IS NOT NULL;

-- Check for any remaining NULL profile pictures in active providers
SELECT user_id, username, verification_status 
FROM users 
WHERE verification_status IN ('approved', 'verified') 
  AND profile_picture_url IS NULL
LIMIT 10;
