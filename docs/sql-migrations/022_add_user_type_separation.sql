-- Migration 022: แยกประเภทผู้ใช้ให้ชัดเจน (User Type Separation)
-- สร้างวันที่: 2025-11-24
-- วัตถุประสงค์: แยกประเภทผู้ใช้ 4 กลุ่ม - Regular Users, Providers, Admins, GOD

BEGIN;

-- 1. เพิ่ม user_type enum
DO $$ BEGIN
    CREATE TYPE user_type_enum AS ENUM ('client', 'provider', 'admin', 'god');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- 2. เพิ่ม column user_type ใน users table
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS user_type user_type_enum DEFAULT 'client';

-- 3. อัพเดทข้อมูลเดิมตาม logic
UPDATE users SET user_type = 
    CASE 
        WHEN tier_id = 5 THEN 'god'::user_type_enum
        WHEN is_admin = true AND tier_id != 5 THEN 'admin'::user_type_enum
        WHEN verification_status IN ('verified', 'approved') AND 
             EXISTS (SELECT 1 FROM service_packages WHERE provider_id = users.user_id)
        THEN 'provider'::user_type_enum
        ELSE 'client'::user_type_enum
    END;

-- 4. เพิ่ม NOT NULL constraint
ALTER TABLE users ALTER COLUMN user_type SET NOT NULL;

-- 5. สร้าง indexes สำหรับ query performance
CREATE INDEX IF NOT EXISTS idx_users_user_type ON users(user_type);
CREATE INDEX IF NOT EXISTS idx_users_type_status ON users(user_type, verification_status);

-- 6. สร้าง Views สำหรับแต่ละประเภท (ง่ายต่อการ query)

-- View: Regular Clients
CREATE OR REPLACE VIEW view_clients AS
SELECT 
    u.user_id,
    u.username,
    u.email,
    u.gender_id,
    u.tier_id,
    t.name as tier_name,
    u.registration_date,
    u.google_id,
    u.google_profile_picture,
    p.bio,
    p.age,
    COUNT(DISTINCT b.booking_id) as total_bookings,
    COUNT(DISTINCT f.favorite_id) as total_favorites
FROM users u
LEFT JOIN tiers t ON u.tier_id = t.tier_id
LEFT JOIN user_profiles p ON u.user_id = p.user_id
LEFT JOIN bookings b ON u.user_id = b.client_id
LEFT JOIN favorites f ON u.user_id = f.client_id
WHERE u.user_type = 'client'
GROUP BY u.user_id, u.username, u.email, u.gender_id, u.tier_id, t.name, 
         u.registration_date, u.google_id, u.google_profile_picture, p.bio, p.age;

-- View: Service Providers
CREATE OR REPLACE VIEW view_providers AS
SELECT 
    u.user_id,
    u.username,
    u.email,
    u.gender_id,
    u.tier_id,
    t.name as tier_name,
    u.verification_status,
    u.provider_level_id,
    pt.name as provider_tier_name,
    p.bio,
    p.age,
    p.height,
    p.service_type,
    p.province,
    p.is_available,
    COUNT(DISTINCT sp.package_id) as total_packages,
    COUNT(DISTINCT b.booking_id) as total_bookings,
    COALESCE(AVG(r.rating), 0) as avg_rating,
    COUNT(DISTINCT r.review_id) as total_reviews
FROM users u
LEFT JOIN tiers t ON u.tier_id = t.tier_id
LEFT JOIN tiers pt ON u.provider_level_id = pt.tier_id
LEFT JOIN user_profiles p ON u.user_id = p.user_id
LEFT JOIN service_packages sp ON u.user_id = sp.provider_id
LEFT JOIN bookings b ON u.user_id = b.provider_id
LEFT JOIN reviews r ON u.user_id = r.provider_id
WHERE u.user_type = 'provider'
GROUP BY u.user_id, u.username, u.email, u.gender_id, u.tier_id, t.name,
         u.verification_status, u.provider_level_id, pt.name, p.bio, p.age, 
         p.height, p.service_type, p.province, p.is_available;

-- View: Admins
CREATE OR REPLACE VIEW view_admins AS
SELECT 
    u.user_id,
    u.username,
    u.email,
    u.tier_id,
    t.name as tier_name,
    u.is_admin,
    u.registration_date,
    COUNT(DISTINCT verified_docs.document_id) as documents_verified,
    COUNT(DISTINCT approved_users.user_id) as users_approved
FROM users u
LEFT JOIN tiers t ON u.tier_id = t.tier_id
LEFT JOIN provider_documents verified_docs ON verified_docs.verified_at IS NOT NULL
LEFT JOIN users approved_users ON approved_users.verification_status = 'approved'
WHERE u.user_type = 'admin'
GROUP BY u.user_id, u.username, u.email, u.tier_id, t.name, u.is_admin, u.registration_date;

-- View: GOD Account
CREATE OR REPLACE VIEW view_god AS
SELECT 
    u.user_id,
    u.username,
    u.email,
    u.tier_id,
    t.name as tier_name,
    u.is_admin,
    u.registration_date,
    (SELECT COUNT(*) FROM users WHERE user_type = 'client') as total_clients,
    (SELECT COUNT(*) FROM users WHERE user_type = 'provider') as total_providers,
    (SELECT COUNT(*) FROM users WHERE user_type = 'admin') as total_admins,
    (SELECT COUNT(*) FROM transactions) as total_transactions,
    (SELECT SUM(amount) FROM transactions WHERE transaction_type = 'platform_commission') as total_commission
FROM users u
LEFT JOIN tiers t ON u.tier_id = t.tier_id
WHERE u.user_type = 'god';

-- 7. สร้าง Functions สำหรับเปลี่ยนประเภทผู้ใช้

-- Function: Promote User to Provider
CREATE OR REPLACE FUNCTION promote_to_provider(target_user_id INT)
RETURNS BOOLEAN AS $$
BEGIN
    UPDATE users 
    SET user_type = 'provider'::user_type_enum,
        verification_status = 'pending'
    WHERE user_id = target_user_id AND user_type = 'client';
    
    RETURN FOUND;
END;
$$ LANGUAGE plpgsql;

-- Function: Promote User to Admin (GOD only)
CREATE OR REPLACE FUNCTION promote_to_admin(target_user_id INT, requester_id INT)
RETURNS BOOLEAN AS $$
DECLARE
    requester_type user_type_enum;
BEGIN
    -- Check if requester is GOD
    SELECT user_type INTO requester_type FROM users WHERE user_id = requester_id;
    
    IF requester_type != 'god' THEN
        RAISE EXCEPTION 'Only GOD can promote users to admin';
    END IF;
    
    -- Promote user
    UPDATE users 
    SET user_type = 'admin'::user_type_enum,
        is_admin = true
    WHERE user_id = target_user_id AND user_type != 'god';
    
    RETURN FOUND;
END;
$$ LANGUAGE plpgsql;

-- Function: Demote User (GOD only)
CREATE OR REPLACE FUNCTION demote_user(target_user_id INT, requester_id INT)
RETURNS BOOLEAN AS $$
DECLARE
    requester_type user_type_enum;
BEGIN
    -- Check if requester is GOD
    SELECT user_type INTO requester_type FROM users WHERE user_id = requester_id;
    
    IF requester_type != 'god' THEN
        RAISE EXCEPTION 'Only GOD can demote users';
    END IF;
    
    -- Cannot demote GOD
    IF target_user_id = 1 THEN
        RAISE EXCEPTION 'Cannot demote GOD account';
    END IF;
    
    -- Demote to client
    UPDATE users 
    SET user_type = 'client'::user_type_enum,
        is_admin = false
    WHERE user_id = target_user_id;
    
    RETURN FOUND;
END;
$$ LANGUAGE plpgsql;

-- 8. สร้าง Triggers เพื่อป้องกันการแก้ไข GOD account

-- Trigger: ป้องกันการเปลี่ยน user_type ของ GOD
CREATE OR REPLACE FUNCTION protect_god_user_type()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.user_id = 1 AND NEW.user_type != 'god' THEN
        RAISE EXCEPTION 'Cannot change GOD user type';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_protect_god_user_type
BEFORE UPDATE ON users
FOR EACH ROW
WHEN (OLD.user_id = 1)
EXECUTE FUNCTION protect_god_user_type();

-- 9. เพิ่ม Check Constraint
ALTER TABLE users
ADD CONSTRAINT check_god_tier 
CHECK (
    (user_type = 'god' AND tier_id = 5 AND is_admin = true) OR
    (user_type != 'god')
);

-- 10. Comment เพิ่มเติม
COMMENT ON COLUMN users.user_type IS 'ประเภทผู้ใช้: client (ลูกค้า), provider (ผู้ให้บริการ), admin (ผู้ดูแลระบบ), god (พระเจ้า)';
COMMENT ON VIEW view_clients IS 'มุมมองข้อมูลลูกค้าทั่วไป';
COMMENT ON VIEW view_providers IS 'มุมมองข้อมูลผู้ให้บริการ';
COMMENT ON VIEW view_admins IS 'มุมมองข้อมูล Admin';
COMMENT ON VIEW view_god IS 'มุมมองข้อมูล GOD พร้อม statistics';

COMMIT;

-- การใช้งาน:
-- 1. Query clients: SELECT * FROM view_clients;
-- 2. Query providers: SELECT * FROM view_providers;
-- 3. Query admins: SELECT * FROM view_admins;
-- 4. Query GOD: SELECT * FROM view_god;
-- 5. Promote to provider: SELECT promote_to_provider(user_id);
-- 6. Promote to admin (GOD only): SELECT promote_to_admin(target_user_id, god_user_id);
-- 7. Demote user (GOD only): SELECT demote_user(target_user_id, god_user_id);
