-- Migration 034: Safety & Business Features
-- Trusted Contacts, SOS, Check-in/Check-out, Private Gallery, Deposits, Cancellation, Boost, Coupons

-- ================================
-- Trusted Contacts
-- ================================
CREATE TABLE IF NOT EXISTS trusted_contacts (
    contact_id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    phone_number VARCHAR(20) NOT NULL,
    relationship VARCHAR(50) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    last_notified TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_trusted_contacts_user ON trusted_contacts(user_id);

-- ================================
-- SOS Alerts
-- ================================
CREATE TABLE IF NOT EXISTS sos_alerts (
    alert_id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    booking_id INT REFERENCES bookings(booking_id) ON DELETE SET NULL,
    latitude DECIMAL(10, 8) NOT NULL,
    longitude DECIMAL(11, 8) NOT NULL,
    location_text TEXT,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'resolved', 'cancelled')),
    resolved_at TIMESTAMPTZ,
    resolved_by INT REFERENCES users(user_id),
    resolution_note TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sos_alerts_user ON sos_alerts(user_id);
CREATE INDEX IF NOT EXISTS idx_sos_alerts_status ON sos_alerts(status);

-- ================================
-- Booking Check-ins
-- ================================
CREATE TABLE IF NOT EXISTS booking_check_ins (
    check_in_id SERIAL PRIMARY KEY,
    booking_id INT NOT NULL REFERENCES bookings(booking_id) ON DELETE CASCADE,
    provider_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    client_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    checked_in_at TIMESTAMPTZ NOT NULL,
    expected_end_time TIMESTAMPTZ NOT NULL,
    checked_out_at TIMESTAMPTZ,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'completed', 'overdue', 'emergency')),
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_check_ins_booking ON booking_check_ins(booking_id);
CREATE INDEX IF NOT EXISTS idx_check_ins_status ON booking_check_ins(status);

-- ================================
-- Private Gallery
-- ================================
CREATE TABLE IF NOT EXISTS private_photos (
    photo_id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    photo_url TEXT NOT NULL,
    thumbnail_url TEXT,
    sort_order INT DEFAULT 0,
    price DECIMAL(10, 2),
    is_active BOOLEAN DEFAULT true,
    uploaded_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_private_photos_user ON private_photos(user_id);

CREATE TABLE IF NOT EXISTS private_gallery_settings (
    setting_id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE UNIQUE,
    is_enabled BOOLEAN DEFAULT false,
    monthly_price DECIMAL(10, 2),
    one_time_price DECIMAL(10, 2),
    allow_one_time BOOLEAN DEFAULT true,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS private_gallery_access (
    access_id SERIAL PRIMARY KEY,
    gallery_owner_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    viewer_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    access_type VARCHAR(20) NOT NULL CHECK (access_type IN ('subscription', 'one_time')),
    expires_at TIMESTAMPTZ,
    granted_at TIMESTAMPTZ DEFAULT NOW(),
    payment_id INT,
    UNIQUE(gallery_owner_id, viewer_id)
);

CREATE INDEX IF NOT EXISTS idx_gallery_access_owner ON private_gallery_access(gallery_owner_id);
CREATE INDEX IF NOT EXISTS idx_gallery_access_viewer ON private_gallery_access(viewer_id);

-- ================================
-- Deposit System
-- ================================
CREATE TABLE IF NOT EXISTS provider_deposit_settings (
    setting_id SERIAL PRIMARY KEY,
    provider_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE UNIQUE,
    require_deposit BOOLEAN DEFAULT false,
    deposit_percentage DECIMAL(3, 2) DEFAULT 0.30 CHECK (deposit_percentage BETWEEN 0.10 AND 0.50),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS booking_deposits (
    deposit_id SERIAL PRIMARY KEY,
    booking_id INT NOT NULL REFERENCES bookings(booking_id) ON DELETE CASCADE UNIQUE,
    client_id INT NOT NULL REFERENCES users(user_id),
    provider_id INT NOT NULL REFERENCES users(user_id),
    amount DECIMAL(10, 2) NOT NULL,
    percentage DECIMAL(3, 2) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'paid', 'refunded', 'forfeited')),
    paid_at TIMESTAMPTZ,
    refunded_at TIMESTAMPTZ,
    forfeited_at TIMESTAMPTZ,
    payment_intent_id VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_deposits_booking ON booking_deposits(booking_id);
CREATE INDEX IF NOT EXISTS idx_deposits_status ON booking_deposits(status);

-- ================================
-- Cancellation Policy
-- ================================
CREATE TABLE IF NOT EXISTS cancellation_policies (
    policy_id SERIAL PRIMARY KEY,
    provider_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    hours_before_booking INT NOT NULL,
    fee_percentage DECIMAL(3, 2) NOT NULL CHECK (fee_percentage BETWEEN 0 AND 1),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_cancellation_policies_provider ON cancellation_policies(provider_id);

CREATE TABLE IF NOT EXISTS cancellation_fees (
    fee_id SERIAL PRIMARY KEY,
    booking_id INT NOT NULL REFERENCES bookings(booking_id) ON DELETE CASCADE,
    cancelled_by INT NOT NULL REFERENCES users(user_id),
    fee_amount DECIMAL(10, 2) NOT NULL,
    fee_percentage DECIMAL(3, 2) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'paid', 'waived')),
    paid_at TIMESTAMPTZ,
    waived_at TIMESTAMPTZ,
    waived_by INT REFERENCES users(user_id),
    waiver_reason TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_cancellation_fees_booking ON cancellation_fees(booking_id);

-- ================================
-- Profile Boost
-- ================================
CREATE TABLE IF NOT EXISTS boost_packages (
    package_id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    boost_type VARCHAR(50) NOT NULL,
    duration INT NOT NULL, -- hours
    price DECIMAL(10, 2) NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Insert default boost packages
INSERT INTO boost_packages (name, boost_type, duration, price, description) VALUES
('Featured 1 Hour', 'featured', 1, 50.00, 'Show at top of search results for 1 hour'),
('Featured 6 Hours', 'featured', 6, 250.00, 'Show at top of search results for 6 hours'),
('Featured 24 Hours', 'featured', 24, 800.00, 'Show at top of search results for 24 hours'),
('Spotlight 1 Hour', 'spotlight', 1, 100.00, 'Featured with special badge for 1 hour'),
('Spotlight 6 Hours', 'spotlight', 6, 500.00, 'Featured with special badge for 6 hours'),
('Top Search 24 Hours', 'top_search', 24, 1500.00, 'Guaranteed top 3 in search for 24 hours')
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS profile_boosts (
    boost_id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    boost_type VARCHAR(50) NOT NULL,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    amount DECIMAL(10, 2) NOT NULL,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'expired', 'cancelled')),
    payment_id INT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_boosts_user ON profile_boosts(user_id);
CREATE INDEX IF NOT EXISTS idx_boosts_active ON profile_boosts(status, end_time);

-- ================================
-- Coupons
-- ================================
CREATE TABLE IF NOT EXISTS coupons (
    coupon_id SERIAL PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    discount_type VARCHAR(20) NOT NULL CHECK (discount_type IN ('percentage', 'fixed')),
    discount_value DECIMAL(10, 2) NOT NULL,
    min_booking_amount DECIMAL(10, 2),
    max_discount DECIMAL(10, 2),
    valid_from TIMESTAMPTZ NOT NULL,
    valid_until TIMESTAMPTZ NOT NULL,
    usage_limit INT,
    used_count INT DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_by INT NOT NULL REFERENCES users(user_id),
    provider_id INT REFERENCES users(user_id), -- NULL = platform-wide
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_coupons_code ON coupons(code);
CREATE INDEX IF NOT EXISTS idx_coupons_provider ON coupons(provider_id);

CREATE TABLE IF NOT EXISTS coupon_usages (
    usage_id SERIAL PRIMARY KEY,
    coupon_id INT NOT NULL REFERENCES coupons(coupon_id) ON DELETE CASCADE,
    user_id INT NOT NULL REFERENCES users(user_id),
    booking_id INT NOT NULL REFERENCES bookings(booking_id),
    discount_amount DECIMAL(10, 2) NOT NULL,
    used_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_coupon_usages_coupon ON coupon_usages(coupon_id);
CREATE INDEX IF NOT EXISTS idx_coupon_usages_user ON coupon_usages(user_id);

-- ================================
-- Photo Verification Badge
-- ================================
CREATE TABLE IF NOT EXISTS photo_verifications (
    verification_id SERIAL PRIMARY KEY,
    photo_id INT NOT NULL REFERENCES user_photos(photo_id) ON DELETE CASCADE,
    user_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'verified', 'rejected')),
    verified_at TIMESTAMPTZ,
    verified_by INT REFERENCES users(user_id),
    rejection_reason TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_photo_verifications_photo ON photo_verifications(photo_id);
CREATE INDEX IF NOT EXISTS idx_photo_verifications_status ON photo_verifications(status);

-- Add is_verified column to user_photos if not exists
ALTER TABLE user_photos ADD COLUMN IF NOT EXISTS is_verified BOOLEAN DEFAULT false;

-- ================================
-- Add deposit_paid status to bookings
-- ================================
-- No need to alter - status is VARCHAR and accepts any value

-- Create trigger for check-in overdue detection
CREATE OR REPLACE FUNCTION check_overdue_checkins() RETURNS void AS $$
BEGIN
    UPDATE booking_check_ins 
    SET status = 'overdue', updated_at = NOW()
    WHERE status = 'active' 
    AND expected_end_time < NOW() - INTERVAL '15 minutes';
END;
$$ LANGUAGE plpgsql;
