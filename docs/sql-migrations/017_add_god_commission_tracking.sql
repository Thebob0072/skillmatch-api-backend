-- Migration 017: GOD Commission Balance Tracking
-- ติดตามยอดค่าคอมมิชชั่นที่เก็บไว้ในบัญชี GOD

-- ================================
-- 1. GOD Commission Balance Table
-- ================================
CREATE TABLE god_commission_balance (
    balance_id SERIAL PRIMARY KEY,
    
    -- GOD Account
    god_user_id INTEGER NOT NULL REFERENCES users(user_id),
    platform_bank_account_id INTEGER NOT NULL REFERENCES platform_bank_accounts(platform_bank_id),
    
    -- Balance Tracking
    total_commission_collected DECIMAL(12, 2) DEFAULT 0.00 NOT NULL, -- รวมค่าคอมฯ ที่เก็บได้ทั้งหมด
    total_transferred DECIMAL(12, 2) DEFAULT 0.00 NOT NULL,          -- รวมที่โอนไป provider แล้ว
    current_balance DECIMAL(12, 2) DEFAULT 0.00 NOT NULL,            -- ยอดคงเหลือในบัญชี GOD
    
    -- Statistics
    total_withdrawals_processed INTEGER DEFAULT 0,                   -- จำนวน withdrawal ที่ทำแล้ว
    average_withdrawal_amount DECIMAL(12, 2) DEFAULT 0.00,           -- ยอดเฉลี่ยต่อครั้ง
    
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(god_user_id, platform_bank_account_id),
    CONSTRAINT positive_balances CHECK (
        total_commission_collected >= 0 AND 
        total_transferred >= 0 AND 
        current_balance >= 0
    )
);

-- ================================
-- 2. Commission Transaction History
-- ================================
-- ตารางเก็บประวัติการรับค่าคอมมิชชั่นแต่ละครั้ง
CREATE TABLE commission_transactions (
    commission_txn_id SERIAL PRIMARY KEY,
    
    -- Related Records
    booking_id INTEGER REFERENCES bookings(booking_id),
    transaction_id INTEGER REFERENCES transactions(transaction_id),
    
    -- Commission Details
    booking_amount DECIMAL(12, 2) NOT NULL,          -- ยอด booking เต็ม
    commission_rate DECIMAL(5, 4) DEFAULT 0.1000,    -- อัตรา (10%)
    commission_amount DECIMAL(12, 2) NOT NULL,       -- ค่าคอมฯ (เช่น 100 บาท)
    provider_amount DECIMAL(12, 2) NOT NULL,         -- ยอดที่ provider ได้ (เช่น 900 บาท)
    
    -- Provider Info
    provider_id INTEGER NOT NULL REFERENCES users(user_id),
    
    -- Platform Bank Account
    platform_bank_account_id INTEGER REFERENCES platform_bank_accounts(platform_bank_id),
    
    -- Status
    status VARCHAR(20) DEFAULT 'collected',          -- collected, refunded
    
    collected_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    refunded_at TIMESTAMP,
    refund_reason TEXT,
    
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ================================
-- 3. Update Withdrawals Table
-- ================================
-- เพิ่ม columns สำหรับเก็บ original slip และ commission amount
ALTER TABLE withdrawals 
    ADD COLUMN IF NOT EXISTS original_slip_url TEXT,           -- สลิปต้นฉบับ (ไม่แสดง provider)
    ADD COLUMN IF NOT EXISTS commission_withheld DECIMAL(12, 2) DEFAULT 0.00, -- ค่าคอมฯ ที่หักไว้
    ADD COLUMN IF NOT EXISTS notification_sent BOOLEAN DEFAULT false,         -- ส่ง notification แล้ว
    ADD COLUMN IF NOT EXISTS email_sent BOOLEAN DEFAULT false;                -- ส่ง email แล้ว

COMMENT ON COLUMN withdrawals.original_slip_url IS 'สลิปต้นฉบับจากธนาคาร (มีข้อมูล GOD - เก็บไว้ใน DB อย่างเดียว)';
COMMENT ON COLUMN withdrawals.transfer_slip_url IS 'สลิปที่แก้ไขแล้ว (ซ่อนข้อมูล GOD - แสดงให้ provider)';
COMMENT ON COLUMN withdrawals.commission_withheld IS 'ค่าคอมมิชชั่น 10% ที่เก็บไว้ในบัญชี GOD';

-- ================================
-- 4. Triggers for Auto-Update GOD Balance
-- ================================

-- Trigger 1: เมื่อมี booking สำเร็จ → เพิ่มค่าคอมฯ
CREATE OR REPLACE FUNCTION increment_god_commission()
RETURNS TRIGGER AS $$
DECLARE
    commission_amt DECIMAL(12, 2);
    provider_amt DECIMAL(12, 2);
    default_platform_bank_id INTEGER;
BEGIN
    -- คำนวณค่าคอมฯ 10%
    commission_amt := NEW.amount * 0.10;
    provider_amt := NEW.amount - commission_amt;
    
    -- หา default platform bank account
    SELECT platform_bank_id INTO default_platform_bank_id
    FROM platform_bank_accounts
    WHERE is_default = true AND is_active = true
    LIMIT 1;
    
    IF default_platform_bank_id IS NOT NULL THEN
        -- อัพเดท GOD commission balance
        INSERT INTO god_commission_balance (
            god_user_id, platform_bank_account_id,
            total_commission_collected, current_balance
        ) VALUES (
            1, default_platform_bank_id,  -- god_user_id = 1
            commission_amt, commission_amt
        )
        ON CONFLICT (god_user_id, platform_bank_account_id) 
        DO UPDATE SET
            total_commission_collected = god_commission_balance.total_commission_collected + commission_amt,
            current_balance = god_commission_balance.current_balance + commission_amt,
            last_updated = CURRENT_TIMESTAMP;
        
        -- บันทึกประวัติ commission
        INSERT INTO commission_transactions (
            booking_id, transaction_id, booking_amount,
            commission_rate, commission_amount, provider_amount,
            provider_id, platform_bank_account_id, status
        ) VALUES (
            NEW.booking_id, NEW.transaction_id, NEW.amount,
            0.1000, commission_amt, provider_amt,
            NEW.related_user_id, default_platform_bank_id, 'collected'
        );
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_increment_god_commission
AFTER INSERT ON transactions
FOR EACH ROW
WHEN (NEW.type = 'booking_payment' AND NEW.status = 'completed')
EXECUTE FUNCTION increment_god_commission();

-- Trigger 2: เมื่อ withdrawal completed → หักยอดออกจาก GOD balance
CREATE OR REPLACE FUNCTION decrement_god_commission()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status = 'completed' AND OLD.status != 'completed' THEN
        -- หักยอดที่โอนออก + ค่าคอมฯ ที่หักไว้
        UPDATE god_commission_balance
        SET total_transferred = total_transferred + NEW.net_amount,
            current_balance = current_balance - (NEW.requested_amount - NEW.fee), -- หักยอดรวม (ไม่รวมค่าธรรมเนียม)
            total_withdrawals_processed = total_withdrawals_processed + 1,
            average_withdrawal_amount = (
                (average_withdrawal_amount * total_withdrawals_processed + NEW.net_amount) / 
                (total_withdrawals_processed + 1)
            ),
            last_updated = CURRENT_TIMESTAMP
        WHERE platform_bank_account_id = NEW.platform_bank_account_id;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_decrement_god_commission
AFTER UPDATE ON withdrawals
FOR EACH ROW
EXECUTE FUNCTION decrement_god_commission();

-- ================================
-- 5. Indexes
-- ================================
CREATE INDEX idx_god_commission_balance_user ON god_commission_balance(god_user_id);
CREATE INDEX idx_god_commission_balance_platform ON god_commission_balance(platform_bank_account_id);

CREATE INDEX idx_commission_transactions_booking ON commission_transactions(booking_id);
CREATE INDEX idx_commission_transactions_provider ON commission_transactions(provider_id);
CREATE INDEX idx_commission_transactions_platform ON commission_transactions(platform_bank_account_id);
CREATE INDEX idx_commission_transactions_status ON commission_transactions(status);
CREATE INDEX idx_commission_transactions_collected ON commission_transactions(collected_at DESC);

CREATE INDEX idx_withdrawals_notification ON withdrawals(notification_sent) WHERE notification_sent = false;
CREATE INDEX idx_withdrawals_email ON withdrawals(email_sent) WHERE email_sent = false;

-- ================================
-- 6. Initialize GOD Commission Balance
-- ================================
-- สร้าง record สำหรับ GOD user (user_id = 1)
INSERT INTO god_commission_balance (
    god_user_id, 
    platform_bank_account_id,
    total_commission_collected,
    total_transferred,
    current_balance
)
SELECT 
    1,  -- GOD user_id
    platform_bank_id,
    0.00,
    0.00,
    0.00
FROM platform_bank_accounts
WHERE is_default = true AND is_active = true
LIMIT 1
ON CONFLICT (god_user_id, platform_bank_account_id) DO NOTHING;

-- ================================
-- 7. Views for Easy Querying
-- ================================

-- View: GOD Dashboard Summary
CREATE OR REPLACE VIEW v_god_financial_summary AS
SELECT 
    gcb.god_user_id,
    gcb.platform_bank_account_id,
    pba.account_name as platform_account_name,
    pba.bank_name,
    pba.account_number,
    gcb.total_commission_collected,
    gcb.total_transferred,
    gcb.current_balance,
    gcb.total_withdrawals_processed,
    gcb.average_withdrawal_amount,
    -- Pending withdrawals
    (SELECT COUNT(*) FROM withdrawals WHERE status = 'pending') as pending_withdrawals_count,
    (SELECT COALESCE(SUM(net_amount), 0) FROM withdrawals WHERE status = 'pending') as pending_withdrawals_amount,
    gcb.last_updated
FROM god_commission_balance gcb
JOIN platform_bank_accounts pba ON gcb.platform_bank_account_id = pba.platform_bank_id;

-- View: Commission Transaction History
CREATE OR REPLACE VIEW v_commission_history AS
SELECT 
    ct.commission_txn_id,
    ct.booking_id,
    b.booking_date,
    ct.provider_id,
    u.username as provider_username,
    ct.booking_amount,
    ct.commission_rate,
    ct.commission_amount,
    ct.provider_amount,
    ct.status,
    ct.collected_at,
    ct.refunded_at,
    ct.refund_reason
FROM commission_transactions ct
LEFT JOIN bookings b ON ct.booking_id = b.booking_id
LEFT JOIN users u ON ct.provider_id = u.user_id
ORDER BY ct.collected_at DESC;

-- ================================
-- 8. Comments
-- ================================
COMMENT ON TABLE god_commission_balance IS 'ติดตามยอดค่าคอมมิชชั่นที่เก็บไว้ในบัญชี GOD/แพลตฟอร์ม';
COMMENT ON TABLE commission_transactions IS 'ประวัติการเก็บค่าคอมมิชชั่นจาก bookings';

COMMENT ON COLUMN god_commission_balance.total_commission_collected IS 'รวมค่าคอมฯ 10% ที่เก็บได้ทั้งหมด';
COMMENT ON COLUMN god_commission_balance.total_transferred IS 'รวมยอด 90% ที่โอนไป provider แล้ว';
COMMENT ON COLUMN god_commission_balance.current_balance IS 'ยอดคงเหลือในบัญชี GOD = collected - transferred';

COMMENT ON COLUMN commission_transactions.booking_amount IS 'ยอด booking เต็ม (เช่น 1000 บาท)';
COMMENT ON COLUMN commission_transactions.commission_amount IS 'ค่าคอมฯ 10% (เช่น 100 บาท)';
COMMENT ON COLUMN commission_transactions.provider_amount IS 'ยอดที่ provider ได้ 90% (เช่น 900 บาท)';

```