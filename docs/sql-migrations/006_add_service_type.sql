-- Migration: เพิ่ม service_type สำหรับระบุว่า provider ให้บริการแบบไหน

BEGIN;

-- 1. เพิ่มคอลัมน์ service_type
ALTER TABLE user_profiles 
ADD COLUMN IF NOT EXISTS service_type VARCHAR(20) CHECK (service_type IN ('incall', 'outcall'));

-- 2. เพิ่ม comment
COMMENT ON COLUMN user_profiles.service_type IS 
'Service type: incall (provider has location), outcall (provider goes to client)';

-- 3. เพิ่ม index สำหรับการ filter
CREATE INDEX IF NOT EXISTS idx_user_profiles_service_type ON user_profiles(service_type);

COMMIT;
