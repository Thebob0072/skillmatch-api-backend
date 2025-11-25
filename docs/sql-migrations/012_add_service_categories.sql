-- Migration 012: Add Service Categories System
-- Providers can offer multiple types of services

-- 1. Create service_categories table
CREATE TABLE IF NOT EXISTS service_categories (
    category_id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    name_thai VARCHAR(100) NOT NULL,
    description TEXT,
    icon VARCHAR(50), -- emoji or icon name
    is_adult BOOLEAN DEFAULT false, -- requires age verification
    display_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 2. Create provider_categories junction table (many-to-many)
CREATE TABLE IF NOT EXISTS provider_categories (
    provider_category_id SERIAL PRIMARY KEY,
    provider_id INTEGER NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    category_id INTEGER NOT NULL REFERENCES service_categories(category_id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(provider_id, category_id) -- prevent duplicates
);

-- 3. Add indexes
CREATE INDEX IF NOT EXISTS idx_provider_categories_provider ON provider_categories(provider_id);
CREATE INDEX IF NOT EXISTS idx_provider_categories_category ON provider_categories(category_id);
CREATE INDEX IF NOT EXISTS idx_service_categories_active ON service_categories(is_active);

-- 4. Insert default service categories
INSERT INTO service_categories (name, name_thai, description, icon, is_adult, display_order) VALUES
-- Adult Services (18+)
('adult_entertainment', 'บริการผู้ใหญ่', 'Adult companionship and entertainment services', '🔞', true, 1),
('escort', 'แอสคอร์ท', 'Professional escort services', '💋', true, 2),

-- Healthcare & Wellness
('massage_therapy', 'นวดบำบัด', 'Professional massage and therapy services', '💆', false, 3),
('spa_wellness', 'สปาและเวลเนส', 'Spa treatments and wellness services', '🧖', false, 4),
('personal_care', 'ดูแลส่วนตัว', 'Personal care and assistance', '🤲', false, 5),
('healthcare_companion', 'เพื่อนดูแลสุขภาพ', 'Healthcare companion and elderly care', '👩‍⚕️', false, 6),

-- Entertainment & Bar
('bartender', 'บาร์เทนเดอร์', 'Bartending and drink service', '🍷', false, 7),
('party_host', 'พิธีกร/โฮสต์งานเลี้ยง', 'Party hosting and entertainment', '🎉', false, 8),
('karaoke_companion', 'เพื่อนร้องเพลง', 'Karaoke and singing companion', '🎤', false, 9),

-- Social Activities
('dining_companion', 'เพื่อนทานอาหาร', 'Dining companion services', '🍽️', false, 10),
('movie_companion', 'เพื่อนดูหนัง', 'Movie and entertainment companion', '🎬', false, 11),
('shopping_companion', 'เพื่อนช็อปปิ้ง', 'Shopping companion services', '🛍️', false, 12),
('travel_companion', 'เพื่อนเดินทาง', 'Travel and tour companion', '✈️', false, 13),

-- Professional Services
('personal_assistant', 'ผู้ช่วยส่วนตัว', 'Personal assistant services', '📋', false, 14),
('event_companion', 'เพื่อนงานอีเว้นท์', 'Event and social gathering companion', '🎊', false, 15),
('language_practice', 'ฝึกภาษา', 'Language practice and conversation partner', '💬', false, 16),

-- Fitness & Sports
('fitness_trainer', 'เทรนเนอร์ส่วนตัว', 'Personal fitness trainer', '💪', false, 17),
('sports_companion', 'เพื่อนเล่นกีฬา', 'Sports and exercise companion', '⚽', false, 18),

-- Creative & Arts
('photography_model', 'โมเดลถ่ายภาพ', 'Photography and modeling services', '📸', false, 19),
('art_companion', 'เพื่อนชมศิลปะ', 'Art gallery and museum companion', '🎨', false, 20)

ON CONFLICT (name) DO NOTHING;

-- 5. Add comments
COMMENT ON TABLE service_categories IS 'Lookup table for all available service categories';
COMMENT ON TABLE provider_categories IS 'Junction table mapping providers to their offered service categories';
COMMENT ON COLUMN service_categories.is_adult IS 'Requires 18+ age verification to view/book';
COMMENT ON COLUMN service_categories.display_order IS 'Order for displaying categories in UI';
