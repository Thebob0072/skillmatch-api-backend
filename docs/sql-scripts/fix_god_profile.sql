-- ============================================================================
-- FIX GOD ACCOUNT PROFILE DATA
-- ============================================================================
-- Issue: Google OAuth login doesn't populate first_name, last_name, username
-- Solution: Manually update user_id = 1 with correct Google profile data
-- ============================================================================

-- Check current data
SELECT 
    user_id, 
    username, 
    email, 
    first_name, 
    last_name, 
    google_profile_picture,
    google_id,
    tier_id,
    is_admin
FROM users 
WHERE user_id = 1;

-- Update GOD account with Google profile data
UPDATE users 
SET 
    username = COALESCE(NULLIF(username, ''), 'The BOB Film'),  -- Set if empty
    first_name = COALESCE(NULLIF(first_name, ''), 'The BOB'),   -- Set if empty
    last_name = COALESCE(NULLIF(last_name, ''), 'Film')         -- Set if empty
WHERE user_id = 1;

-- Verify the update
SELECT 
    user_id, 
    username, 
    email, 
    first_name, 
    last_name, 
    google_profile_picture,
    tier_id,
    is_admin
FROM users 
WHERE user_id = 1;

-- ============================================================================
-- OPTIONAL: Update google_profile_picture if needed
-- ============================================================================
-- If you have the Google profile picture URL, uncomment and run:
-- UPDATE users 
-- SET google_profile_picture = 'https://lh3.googleusercontent.com/...'
-- WHERE user_id = 1;
