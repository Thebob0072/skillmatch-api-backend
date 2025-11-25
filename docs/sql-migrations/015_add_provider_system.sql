-- ============================================================================
-- Migration 015: Provider Document Upload & Verification System
-- ============================================================================
-- Purpose: 
-- 1. Provider Documents Upload (บัตรประชาชน, ใบรับรองสุขภาพ, ฯลฯ)
-- 2. Admin Verification Workflow
-- 3. Provider Categories (ผู้ให้บริการแต่ละคนระบุประเภทบริการ)
-- 4. Provider Tier System (จัดอันดับอัตโนมัติ + Manual)
-- ============================================================================

-- ============================================================================
-- 1. Provider Documents Table (เอกสารยืนยันตัวตน)
-- ============================================================================
CREATE TABLE IF NOT EXISTS provider_documents (
    document_id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    document_type VARCHAR(50) NOT NULL,
    -- document_type: 'national_id', 'health_certificate', 'business_license', 'portfolio', 'other'
    file_url TEXT NOT NULL,
    file_name VARCHAR(255),
    verification_status VARCHAR(20) NOT NULL DEFAULT 'pending',
    -- verification_status: 'pending', 'approved', 'rejected'
    verified_by INT REFERENCES users(user_id),
    verified_at TIMESTAMPTZ,
    rejection_reason TEXT,
    uploaded_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT chk_verification_status CHECK (verification_status IN ('pending', 'approved', 'rejected'))
);

CREATE INDEX idx_provider_documents_user ON provider_documents(user_id);
CREATE INDEX idx_provider_documents_status ON provider_documents(verification_status);

COMMENT ON TABLE provider_documents IS 'เก็บเอกสารที่ provider ต้องส่งเพื่อยืนยันตัวตน';
COMMENT ON COLUMN provider_documents.document_type IS 'ประเภทเอกสาร: national_id, health_certificate, business_license, portfolio, other';
COMMENT ON COLUMN provider_documents.verification_status IS 'สถานะการตรวจสอบ: pending, approved, rejected';

-- ============================================================================
-- 2. Update Provider Categories Table (อัปเดตตารางที่มีอยู่แล้ว)
-- ============================================================================
-- Note: provider_categories table already exists from migration 012
-- We'll add is_primary column

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'provider_categories' 
                   AND column_name = 'is_primary') THEN
        ALTER TABLE provider_categories ADD COLUMN is_primary BOOLEAN DEFAULT false;
    END IF;
END $$;

COMMENT ON TABLE provider_categories IS 'ผู้ให้บริการระบุหมวดหมู่บริการที่ตัวเองให้บริการได้';
COMMENT ON COLUMN provider_categories.is_primary IS 'หมวดหมู่หลักของผู้ให้บริการ (ใช้สำหรับแสดงผลหลัก)';

-- ============================================================================
-- 3. Provider Stats Table (สถิติผู้ให้บริการ - สำหรับจัดอันดับ)
-- ============================================================================
CREATE TABLE IF NOT EXISTS provider_stats (
    user_id INT PRIMARY KEY REFERENCES users(user_id) ON DELETE CASCADE,
    total_bookings INT DEFAULT 0,
    completed_bookings INT DEFAULT 0,
    cancelled_bookings INT DEFAULT 0,
    average_rating DECIMAL(3,2) DEFAULT 0.00,
    total_reviews INT DEFAULT 0,
    response_rate DECIMAL(5,2) DEFAULT 0.00, -- อัตราการตอบกลับ (%)
    acceptance_rate DECIMAL(5,2) DEFAULT 0.00, -- อัตราการรับงาน (%)
    total_earnings DECIMAL(12,2) DEFAULT 0.00,
    last_active_at TIMESTAMPTZ,
    tier_points INT DEFAULT 0, -- คะแนนสำหรับจัดอันดับ Tier
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_provider_stats_rating ON provider_stats(average_rating);
CREATE INDEX idx_provider_stats_points ON provider_stats(tier_points);

COMMENT ON TABLE provider_stats IS 'สถิติผู้ให้บริการ - ใช้สำหรับจัดอันดับ Tier อัตโนมัติ';
COMMENT ON COLUMN provider_stats.tier_points IS 'คะแนนสำหรับจัดอันดับ: (rating * 20) + (completed_bookings * 5) + (total_reviews * 3)';

-- ============================================================================
-- 4. Provider Tier History (ประวัติการเปลี่ยน Tier)
-- ============================================================================
CREATE TABLE IF NOT EXISTS provider_tier_history (
    history_id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    old_tier_id INT REFERENCES tiers(tier_id),
    new_tier_id INT NOT NULL REFERENCES tiers(tier_id),
    change_type VARCHAR(20) NOT NULL,
    -- change_type: 'auto' (ระบบคำนวณอัตโนมัติ), 'manual' (admin เปลี่ยน), 'subscription' (ซื้อ tier)
    reason TEXT,
    changed_by INT REFERENCES users(user_id), -- ถ้าเป็น manual ให้ระบุ admin ที่เปลี่ยน
    changed_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_provider_tier_history_user ON provider_tier_history(user_id);
CREATE INDEX idx_provider_tier_history_date ON provider_tier_history(changed_at);

COMMENT ON TABLE provider_tier_history IS 'เก็บประวัติการเปลี่ยน Provider Tier (อัตโนมัติ + Manual)';

-- ============================================================================
-- 5. Update Users Table - Add Provider Flags
-- ============================================================================
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_provider BOOLEAN DEFAULT false;
ALTER TABLE users ADD COLUMN IF NOT EXISTS provider_verified_at TIMESTAMPTZ;
ALTER TABLE users ADD COLUMN IF NOT EXISTS provider_verification_status VARCHAR(20) DEFAULT 'pending';
-- provider_verification_status: 'pending', 'documents_submitted', 'approved', 'rejected'

ALTER TABLE users ADD CONSTRAINT chk_provider_verification_status 
    CHECK (provider_verification_status IN ('pending', 'documents_submitted', 'approved', 'rejected'))
    NOT VALID;

CREATE INDEX IF NOT EXISTS idx_users_is_provider ON users(is_provider);
CREATE INDEX IF NOT EXISTS idx_users_provider_status ON users(provider_verification_status);

COMMENT ON COLUMN users.is_provider IS 'ผู้ใช้เป็น provider หรือไม่ (provider ต้องส่งเอกสาร + ผ่านการตรวจสอบ)';
COMMENT ON COLUMN users.provider_verified_at IS 'เวลาที่ admin อนุมัติให้เป็น provider';
COMMENT ON COLUMN users.provider_verification_status IS 'สถานะการตรวจสอบ provider: pending, documents_submitted, approved, rejected';

-- ============================================================================
-- 6. Initialize Provider Stats for Existing Users
-- ============================================================================
-- (Skip initialization if is_provider column doesn't exist yet)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns 
               WHERE table_name = 'users' AND column_name = 'is_provider') THEN
        INSERT INTO provider_stats (user_id, total_bookings, completed_bookings, average_rating, total_reviews)
        SELECT 
            u.user_id,
            COALESCE(COUNT(DISTINCT b.booking_id), 0) as total_bookings,
            COALESCE(COUNT(DISTINCT CASE WHEN b.status = 'completed' THEN b.booking_id END), 0) as completed_bookings,
            COALESCE(AVG(r.rating), 0.00) as average_rating,
            COALESCE(COUNT(DISTINCT r.review_id), 0) as total_reviews
        FROM users u
        LEFT JOIN bookings b ON u.user_id = b.provider_id
        LEFT JOIN reviews r ON u.user_id = r.provider_id
        WHERE u.is_provider = true
        GROUP BY u.user_id
        ON CONFLICT (user_id) DO NOTHING;
    END IF;
END $$;

-- ============================================================================
-- 7. Provider Document Type Reference (ตารางอ้างอิงประเภทเอกสาร)
-- ============================================================================
CREATE TABLE IF NOT EXISTS document_types (
    type_code VARCHAR(50) PRIMARY KEY,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    is_required BOOLEAN DEFAULT false,
    display_order INT DEFAULT 0
);

INSERT INTO document_types (type_code, display_name, description, is_required, display_order) VALUES
('national_id', 'National ID Card', 'บัตรประชาชน / บัตรประจำตัวประชาชน', true, 1),
('health_certificate', 'Health Certificate', 'ใบรับรองสุขภาพ (ไม่เกิน 6 เดือน)', true, 2),
('business_license', 'Business License', 'ใบอนุญาตประกอบธุรกิจ (ถ้ามี)', false, 3),
('portfolio', 'Portfolio / Work Examples', 'ผลงาน / รูปตัวอย่าง', false, 4),
('certification', 'Professional Certification', 'ใบประกาศนียบัตร / ใบรับรองมาตรฐาน', false, 5),
('other', 'Other Documents', 'เอกสารอื่นๆ', false, 6)
ON CONFLICT (type_code) DO NOTHING;

COMMENT ON TABLE document_types IS 'ประเภทเอกสารที่ provider ต้องส่ง (บางอันบังคับ บางอันไม่บังคับ)';

-- ============================================================================
-- 8. Create View: Provider Dashboard Summary
-- ============================================================================
CREATE OR REPLACE VIEW provider_dashboard AS
SELECT 
    u.user_id,
    u.username,
    u.email,
    u.is_provider,
    u.provider_verification_status,
    u.provider_verified_at,
    pt.name as provider_tier_name,
    pt.access_level as provider_tier_level,
    ps.average_rating,
    ps.total_reviews,
    ps.completed_bookings,
    ps.total_bookings,
    ps.response_rate,
    ps.acceptance_rate,
    ps.tier_points,
    ps.total_earnings,
    ps.last_active_at,
    -- Document verification count
    (SELECT COUNT(*) FROM provider_documents pd WHERE pd.user_id = u.user_id AND pd.verification_status = 'approved') as approved_documents,
    (SELECT COUNT(*) FROM provider_documents pd WHERE pd.user_id = u.user_id AND pd.verification_status = 'pending') as pending_documents,
    -- Categories count
    (SELECT COUNT(*) FROM provider_categories pc WHERE pc.provider_id = u.user_id) as total_categories
FROM users u
LEFT JOIN tiers pt ON u.provider_level_id = pt.tier_id
LEFT JOIN provider_stats ps ON u.user_id = ps.user_id
WHERE u.is_provider = true;

COMMENT ON VIEW provider_dashboard IS 'สรุปข้อมูล Provider สำหรับ Dashboard (รวมสถิติ + เอกสาร + หมวดหมู่)';

-- ============================================================================
-- 9. Trigger: Auto-update Provider Stats
-- ============================================================================
CREATE OR REPLACE FUNCTION update_provider_stats()
RETURNS TRIGGER AS $$
BEGIN
    -- Update provider stats when booking status changes
    IF TG_TABLE_NAME = 'bookings' THEN
        UPDATE provider_stats
        SET 
            total_bookings = (SELECT COUNT(*) FROM bookings WHERE provider_id = NEW.provider_id),
            completed_bookings = (SELECT COUNT(*) FROM bookings WHERE provider_id = NEW.provider_id AND status = 'completed'),
            cancelled_bookings = (SELECT COUNT(*) FROM bookings WHERE provider_id = NEW.provider_id AND status = 'cancelled'),
            updated_at = NOW()
        WHERE user_id = NEW.provider_id;
    END IF;
    
    -- Update provider stats when new review is added
    IF TG_TABLE_NAME = 'reviews' THEN
        UPDATE provider_stats
        SET 
            average_rating = (SELECT COALESCE(AVG(rating), 0) FROM reviews WHERE provider_id = NEW.provider_id),
            total_reviews = (SELECT COUNT(*) FROM reviews WHERE provider_id = NEW.provider_id),
            updated_at = NOW()
        WHERE user_id = NEW.provider_id;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers
DROP TRIGGER IF EXISTS trigger_update_provider_stats_booking ON bookings;
CREATE TRIGGER trigger_update_provider_stats_booking
AFTER INSERT OR UPDATE ON bookings
FOR EACH ROW
EXECUTE FUNCTION update_provider_stats();

DROP TRIGGER IF EXISTS trigger_update_provider_stats_review ON reviews;
CREATE TRIGGER trigger_update_provider_stats_review
AFTER INSERT OR UPDATE ON reviews
FOR EACH ROW
EXECUTE FUNCTION update_provider_stats();

-- ============================================================================
-- 10. Function: Calculate Provider Tier Points
-- ============================================================================
CREATE OR REPLACE FUNCTION calculate_provider_tier_points(p_user_id INT)
RETURNS INT AS $$
DECLARE
    v_points INT := 0;
    v_rating DECIMAL(3,2);
    v_completed INT;
    v_reviews INT;
    v_response_rate DECIMAL(5,2);
    v_acceptance_rate DECIMAL(5,2);
BEGIN
    SELECT 
        average_rating,
        completed_bookings,
        total_reviews,
        response_rate,
        acceptance_rate
    INTO 
        v_rating,
        v_completed,
        v_reviews,
        v_response_rate,
        v_acceptance_rate
    FROM provider_stats
    WHERE user_id = p_user_id;
    
    -- Rating Points (0-100 points)
    v_points := v_points + (v_rating * 20);
    
    -- Completed Bookings (5 points each, max 250 points)
    v_points := v_points + LEAST(v_completed * 5, 250);
    
    -- Total Reviews (3 points each, max 150 points)
    v_points := v_points + LEAST(v_reviews * 3, 150);
    
    -- Response Rate (max 50 points)
    v_points := v_points + (v_response_rate * 0.5);
    
    -- Acceptance Rate (max 50 points)
    v_points := v_points + (v_acceptance_rate * 0.5);
    
    RETURN v_points;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION calculate_provider_tier_points IS 'คำนวณคะแนน Tier ของ Provider (max 600 points)';

-- ============================================================================
-- END OF MIGRATION 015
-- ============================================================================
