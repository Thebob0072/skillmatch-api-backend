-- Migration 018: Update Fee Structure to 12.75% Total
-- อัปเดตโครงสร้างค่าธรรมเนียมให้รวม Stripe (2.75%) + Platform (10%) = 12.75%

-- ================================
-- 1. Update Commission Rules Table
-- ================================

-- เพิ่ม column สำหรับ total_rate (auto-calculated)
ALTER TABLE commission_rules
    ADD COLUMN IF NOT EXISTS total_rate DECIMAL(5, 4) 
        GENERATED ALWAYS AS (platform_rate + payment_gateway_rate) STORED;

COMMENT ON COLUMN commission_rules.total_rate IS 'รวมค่าธรรมเนียมทั้งหมด (Platform + Payment Gateway) - คำนวณอัตโนมัติ';

-- อัปเดต default commission rule
UPDATE commission_rules
SET 
    platform_rate = 0.1000,           -- Platform Commission 10%
    payment_gateway_rate = 0.0275,    -- Stripe Payment Gateway 2.75%
    description = 'Total fee: 12.75% (Platform 10% + Payment Gateway 2.75%) - Provider receives 87.25%',
    name = 'Default Fee Structure',
    updated_at = CURRENT_TIMESTAMP
WHERE rule_id = 1;

-- ================================
-- 2. Provider Fee Notification Table
-- ================================
-- เก็บบันทึกว่า provider ได้รับการแจ้งเตือนเรื่องค่าธรรมเนียมแล้ว

CREATE TABLE IF NOT EXISTS provider_fee_notifications (
    notification_id SERIAL PRIMARY KEY,
    provider_id INTEGER NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    
    -- Fee Information (snapshot ณ เวลาที่แจ้ง)
    platform_rate DECIMAL(5, 4) NOT NULL,
    payment_gateway_rate DECIMAL(5, 4) NOT NULL,
    total_rate DECIMAL(5, 4) NOT NULL,
    
    -- Notification Details
    notification_type VARCHAR(50) NOT NULL,  -- registration, update, reminder
    shown_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    acknowledged BOOLEAN DEFAULT false,
    acknowledged_at TIMESTAMP,
    
    -- Channel
    notification_channel VARCHAR(50),        -- modal, email, in_app
    
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_provider_fee_notifications_provider ON provider_fee_notifications(provider_id);
CREATE INDEX idx_provider_fee_notifications_type ON provider_fee_notifications(notification_type);
CREATE INDEX idx_provider_fee_notifications_acknowledged ON provider_fee_notifications(acknowledged) 
    WHERE acknowledged = false;

COMMENT ON TABLE provider_fee_notifications IS 'บันทึกการแจ้งเตือนค่าธรรมเนียมให้ providers';

-- ================================
-- 3. Helper Functions
-- ================================

-- Function: คำนวณรายได้ provider
CREATE OR REPLACE FUNCTION calculate_provider_earning(booking_amount DECIMAL)
RETURNS TABLE (
    gross_amount DECIMAL,
    stripe_fee DECIMAL,
    platform_commission DECIMAL,
    total_fee DECIMAL,
    net_amount DECIMAL,
    provider_percentage DECIMAL
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        booking_amount,                           -- ยอดเต็ม
        ROUND(booking_amount * 0.0275, 2),        -- Stripe 2.75%
        ROUND(booking_amount * 0.1000, 2),        -- Platform 10%
        ROUND(booking_amount * 0.1275, 2),        -- รวม 12.75%
        ROUND(booking_amount * 0.8725, 2),        -- Provider ได้รับ 87.25%
        87.25;                                    -- เปอร์เซ็นต์ที่ provider ได้
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION calculate_provider_earning IS 'คำนวณรายได้สุทธิของ provider หลังหักค่าธรรมเนียม 12.75%';

-- Function: แจ้งเตือนค่าธรรมเนียมเมื่อสร้าง account
CREATE OR REPLACE FUNCTION notify_provider_fee_on_registration()
RETURNS TRIGGER AS $$
DECLARE
    current_platform_rate DECIMAL(5, 4);
    current_gateway_rate DECIMAL(5, 4);
    current_total_rate DECIMAL(5, 4);
BEGIN
    -- ดึงค่า rate ปัจจุบัน
    SELECT platform_rate, payment_gateway_rate, total_rate
    INTO current_platform_rate, current_gateway_rate, current_total_rate
    FROM commission_rules
    WHERE is_active = true
    ORDER BY effective_from DESC
    LIMIT 1;
    
    -- บันทึก notification
    INSERT INTO provider_fee_notifications (
        provider_id, platform_rate, payment_gateway_rate, total_rate,
        notification_type, notification_channel, notes
    ) VALUES (
        NEW.user_id,
        current_platform_rate,
        current_gateway_rate,
        current_total_rate,
        'registration',
        'modal',
        'Fee notification shown during provider account creation'
    );
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger: แจ้งเตือนเมื่อสร้าง provider account
-- (จะ trigger เมื่อ user ที่มี service_packages หรือ provider_level_id > 1)
CREATE OR REPLACE TRIGGER trigger_notify_provider_fee
AFTER INSERT ON user_profiles
FOR EACH ROW
WHEN (NEW.service_type IS NOT NULL)  -- มี service_type = เป็น provider
EXECUTE FUNCTION notify_provider_fee_on_registration();

-- ================================
-- 4. Update Existing Transactions
-- ================================
-- อัปเดต commission_amount ในธุรกรรมที่มีอยู่แล้ว (ถ้าจำเป็น)

-- เพิ่ม columns ใหม่ใน transactions table
ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS stripe_fee DECIMAL(12, 2) DEFAULT 0.00,
    ADD COLUMN IF NOT EXISTS platform_commission DECIMAL(12, 2) DEFAULT 0.00,
    ADD COLUMN IF NOT EXISTS total_fee_percentage DECIMAL(5, 4) DEFAULT 0.1275;

COMMENT ON COLUMN transactions.stripe_fee IS 'ค่าธรรมเนียม Stripe 2.75%';
COMMENT ON COLUMN transactions.platform_commission IS 'ค่าคอมมิชชั่นแพลตฟอร์ม 10%';
COMMENT ON COLUMN transactions.total_fee_percentage IS 'เปอร์เซ็นต์ค่าธรรมเนียมรวม (12.75%)';

-- อัปเดต existing transactions ที่ยังไม่มี breakdown
UPDATE transactions
SET 
    stripe_fee = ROUND(amount * 0.0275, 2),
    platform_commission = ROUND(amount * 0.1000, 2),
    total_fee_percentage = 0.1275
WHERE type IN ('booking_payment', 'provider_earning')
  AND stripe_fee = 0
  AND platform_commission = 0;

-- ================================
-- 5. Views for Provider Dashboard
-- ================================

-- View: Provider Earnings Summary (แสดงรายได้หลังหักค่าธรรมเนียม)
CREATE OR REPLACE VIEW v_provider_earnings_summary AS
SELECT 
    u.user_id,
    u.username,
    u.email,
    COUNT(DISTINCT t.booking_id) as total_bookings,
    COALESCE(SUM(t.amount), 0) as gross_earnings,           -- รายได้รวม (ก่อนหักค่าธรรมเนียม)
    COALESCE(SUM(t.stripe_fee), 0) as total_stripe_fees,   -- Stripe 2.75%
    COALESCE(SUM(t.platform_commission), 0) as total_platform_fees, -- Platform 10%
    COALESCE(SUM(t.stripe_fee + t.platform_commission), 0) as total_fees, -- รวม 12.75%
    COALESCE(SUM(t.net_amount), 0) as net_earnings,         -- รายได้สุทธิ 87.25%
    w.available_balance,
    w.pending_balance,
    w.total_withdrawn
FROM users u
LEFT JOIN transactions t ON u.user_id = t.user_id 
    AND t.type = 'provider_earning' 
    AND t.status = 'completed'
LEFT JOIN wallets w ON u.user_id = w.user_id
WHERE u.provider_level_id > 1 OR EXISTS (
    SELECT 1 FROM service_packages sp WHERE sp.provider_id = u.user_id
)
GROUP BY u.user_id, u.username, u.email, w.available_balance, w.pending_balance, w.total_withdrawn;

COMMENT ON VIEW v_provider_earnings_summary IS 'สรุปรายได้ของ provider แสดงทั้ง gross และ net หลังหักค่าธรรมเนียม 12.75%';

-- View: Fee Breakdown by Transaction
CREATE OR REPLACE VIEW v_transaction_fee_breakdown AS
SELECT 
    t.transaction_id,
    t.transaction_uuid,
    t.user_id,
    u.username,
    t.type,
    t.booking_id,
    t.amount as gross_amount,
    t.stripe_fee,
    t.platform_commission,
    (t.stripe_fee + t.platform_commission) as total_fee,
    t.net_amount,
    t.total_fee_percentage,
    ROUND((t.net_amount / NULLIF(t.amount, 0)) * 100, 2) as net_percentage,
    t.status,
    t.created_at
FROM transactions t
JOIN users u ON t.user_id = u.user_id
WHERE t.type IN ('booking_payment', 'provider_earning')
ORDER BY t.created_at DESC;

COMMENT ON VIEW v_transaction_fee_breakdown IS 'รายละเอียดค่าธรรมเนียมแต่ละธุรกรรม';

-- ================================
-- 6. Insert Sample Fee Notification (For Testing)
-- ================================
-- สำหรับ providers ที่มีอยู่แล้ว ให้สร้าง notification ย้อนหลัง

INSERT INTO provider_fee_notifications (
    provider_id, platform_rate, payment_gateway_rate, total_rate,
    notification_type, notification_channel, acknowledged, notes
)
SELECT 
    u.user_id,
    0.1000,
    0.0275,
    0.1275,
    'update',
    'system',
    false,
    'Fee structure updated to 12.75% - Retroactive notification'
FROM users u
WHERE u.provider_level_id > 1 
   OR EXISTS (SELECT 1 FROM service_packages sp WHERE sp.provider_id = u.user_id)
ON CONFLICT DO NOTHING;

-- ================================
-- 7. Statistics Query Examples
-- ================================

-- Total fees collected by platform
COMMENT ON VIEW v_platform_fee_statistics IS '
-- Query ตัวอย่าง: สถิติค่าธรรมเนียมที่แพลตฟอร์มเก็บได้

SELECT 
    DATE_TRUNC(''month'', created_at) as month,
    COUNT(*) as total_transactions,
    SUM(amount) as total_revenue,
    SUM(stripe_fee) as total_stripe_fees,
    SUM(platform_commission) as total_platform_fees,
    SUM(stripe_fee + platform_commission) as total_fees_collected,
    SUM(net_amount) as total_provider_earnings
FROM transactions
WHERE type = ''booking_payment'' 
  AND status = ''completed''
  AND created_at >= DATE_TRUNC(''year'', CURRENT_DATE)
GROUP BY DATE_TRUNC(''month'', created_at)
ORDER BY month DESC;
';

-- ================================
-- 8. Data Validation
-- ================================

-- ตรวจสอบว่าค่าธรรมเนียมคำนวณถูกต้อง
DO $$
DECLARE
    test_amount DECIMAL := 1000.00;
    result RECORD;
BEGIN
    SELECT * INTO result FROM calculate_provider_earning(test_amount);
    
    RAISE NOTICE 'Fee Calculation Test:';
    RAISE NOTICE 'Gross Amount: %', result.gross_amount;
    RAISE NOTICE 'Stripe Fee (2.75%%): %', result.stripe_fee;
    RAISE NOTICE 'Platform Commission (10%%): %', result.platform_commission;
    RAISE NOTICE 'Total Fee (12.75%%): %', result.total_fee;
    RAISE NOTICE 'Net Amount (87.25%%): %', result.net_amount;
    
    -- Validate
    IF result.total_fee != 127.50 THEN
        RAISE EXCEPTION 'Fee calculation incorrect!';
    END IF;
    
    IF result.net_amount != 872.50 THEN
        RAISE EXCEPTION 'Net amount calculation incorrect!';
    END IF;
    
    RAISE NOTICE '✅ Fee calculation validation passed!';
END $$;

```