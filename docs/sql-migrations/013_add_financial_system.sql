-- Migration 013: Financial System
-- เพิ่มระบบการเงิน: บัญชีธนาคาร, กระเป๋าเงิน, ธุรกรรม, การถอนเงิน

-- ================================
-- 1. Bank Accounts (บัญชีธนาคาร)
-- ================================
CREATE TABLE bank_accounts (
    bank_account_id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    
    -- Bank Information
    bank_name VARCHAR(100) NOT NULL,           -- ชื่อธนาคาร (e.g., "ธนาคารกสิกรไทย")
    bank_code VARCHAR(10),                     -- รหัสธนาคาร (e.g., "KBANK")
    account_number VARCHAR(50) NOT NULL,       -- เลขที่บัญชี
    account_name VARCHAR(200) NOT NULL,        -- ชื่อบัญชี (ต้องตรงกับ KYC)
    account_type VARCHAR(20) DEFAULT 'savings', -- savings, current
    branch_name VARCHAR(100),                  -- สาขา
    
    -- Verification
    is_verified BOOLEAN DEFAULT false,         -- ยืนยันบัญชีแล้ว
    verified_at TIMESTAMP,                     -- วันที่ยืนยัน
    verified_by INTEGER REFERENCES users(user_id), -- ยืนยันโดย admin
    
    -- Status
    is_default BOOLEAN DEFAULT false,          -- บัญชีหลัก
    is_active BOOLEAN DEFAULT true,            -- ใช้งานได้
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(user_id, account_number)
);

-- ================================
-- 2. Wallet (กระเป๋าเงิน)
-- ================================
CREATE TABLE wallets (
    wallet_id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL UNIQUE REFERENCES users(user_id) ON DELETE CASCADE,
    
    -- Balances
    available_balance DECIMAL(12, 2) DEFAULT 0.00 NOT NULL,  -- ยอดพร้อมถอน
    pending_balance DECIMAL(12, 2) DEFAULT 0.00 NOT NULL,    -- ยอดรอยืนยัน (7 วัน)
    total_earned DECIMAL(12, 2) DEFAULT 0.00 NOT NULL,       -- รายได้สะสม
    total_withdrawn DECIMAL(12, 2) DEFAULT 0.00 NOT NULL,    -- ถอนไปแล้ว
    
    -- Commission tracking
    total_commission_paid DECIMAL(12, 2) DEFAULT 0.00,       -- ค่าคอมฯ ที่จ่ายไป
    
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT positive_balance CHECK (available_balance >= 0 AND pending_balance >= 0)
);

-- ================================
-- 3. Transactions (ธุรกรรม)
-- ================================
CREATE TYPE transaction_type AS ENUM (
    'booking_payment',      -- ลูกค้าจ่ายค่าบริการ
    'booking_refund',       -- คืนเงินลูกค้า
    'commission',           -- หักค่าคอมมิชชั่น 10%
    'provider_earning',     -- รายได้ผู้ให้บริการ (90%)
    'withdrawal',           -- ถอนเงิน
    'subscription_fee',     -- ค่าสมาชิกรายเดือน
    'admin_adjustment',     -- ปรับยอดโดย admin
    'bonus',                -- โบนัส/สิทธิพิเศษ
    'penalty'               -- ค่าปรับ
);

CREATE TYPE transaction_status AS ENUM (
    'pending',              -- รอดำเนินการ
    'processing',           -- กำลังดำเนินการ
    'completed',            -- สำเร็จ
    'failed',               -- ล้มเหลว
    'cancelled',            -- ยกเลิก
    'refunded'              -- คืนเงินแล้ว
);

CREATE TABLE transactions (
    transaction_id SERIAL PRIMARY KEY,
    transaction_uuid UUID DEFAULT gen_random_uuid() UNIQUE,
    
    -- Parties involved
    user_id INTEGER NOT NULL REFERENCES users(user_id),      -- ผู้ทำธุรกรรม
    related_user_id INTEGER REFERENCES users(user_id),       -- คู่กรณี (ถ้ามี)
    
    -- Transaction details
    type transaction_type NOT NULL,
    status transaction_status DEFAULT 'pending',
    
    -- Amounts
    amount DECIMAL(12, 2) NOT NULL,                          -- จำนวนเงิน
    commission_amount DECIMAL(12, 2) DEFAULT 0.00,           -- ค่าคอมมิชชั่น
    net_amount DECIMAL(12, 2) NOT NULL,                      -- ยอดสุทธิ
    
    -- Related records
    booking_id INTEGER REFERENCES bookings(booking_id),
    withdrawal_id INTEGER,                                    -- FK to withdrawals
    
    -- Payment gateway
    payment_method VARCHAR(50),                               -- stripe, promptpay, bank_transfer
    payment_intent_id VARCHAR(255),                          -- Stripe Payment Intent ID
    
    -- Metadata
    description TEXT,
    notes TEXT,                                              -- หมายเหตุจาก admin
    metadata JSONB,                                          -- ข้อมูลเพิ่มเติม
    
    processed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ================================
-- 4. Withdrawals (การถอนเงิน)
-- ================================
CREATE TYPE withdrawal_status AS ENUM (
    'pending',              -- รอตรวจสอบ
    'approved',             -- อนุมัติแล้ว
    'processing',           -- กำลังโอนเงิน
    'completed',            -- โอนเงินสำเร็จ
    'rejected',             -- ปฏิเสธ
    'failed'                -- โอนเงินล้มเหลว
);

CREATE TABLE withdrawals (
    withdrawal_id SERIAL PRIMARY KEY,
    withdrawal_uuid UUID DEFAULT gen_random_uuid() UNIQUE,
    
    -- User & Bank
    user_id INTEGER NOT NULL REFERENCES users(user_id),
    bank_account_id INTEGER NOT NULL REFERENCES bank_accounts(bank_account_id),
    
    -- Amount
    requested_amount DECIMAL(12, 2) NOT NULL,                -- จำนวนที่ขอถอน
    fee DECIMAL(12, 2) DEFAULT 0.00,                         -- ค่าธรรมเนียมถอน
    net_amount DECIMAL(12, 2) NOT NULL,                      -- ยอดที่ได้รับจริง
    
    -- Status & Processing
    status withdrawal_status DEFAULT 'pending',
    requested_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    approved_at TIMESTAMP,
    approved_by INTEGER REFERENCES users(user_id),           -- admin ที่อนุมัติ
    processed_at TIMESTAMP,
    completed_at TIMESTAMP,
    
    -- Transfer details
    transfer_reference VARCHAR(100),                         -- เลขที่อ้างอิง
    transfer_slip_url TEXT,                                  -- URL slip โอนเงิน
    
    -- Rejection
    rejection_reason TEXT,
    rejected_at TIMESTAMP,
    rejected_by INTEGER REFERENCES users(user_id),
    
    notes TEXT,                                              -- หมายเหตุจาก admin
    metadata JSONB,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT valid_withdrawal_amount CHECK (requested_amount > 0 AND net_amount > 0)
);

-- ================================
-- 5. Commission Rules (กฎค่าคอมมิชชั่น)
-- ================================
CREATE TABLE commission_rules (
    rule_id SERIAL PRIMARY KEY,
    
    -- Rule details
    name VARCHAR(100) NOT NULL,
    description TEXT,
    
    -- Commission rates
    platform_rate DECIMAL(5, 4) DEFAULT 0.1000,              -- 10% default
    payment_gateway_rate DECIMAL(5, 4) DEFAULT 0.0275,       -- Stripe 2.75%
    
    -- Tier-based rates (optional)
    tier_id INTEGER REFERENCES tiers(tier_id),               -- ถ้าระบุ tier แยก rate ได้
    
    -- Validity
    effective_from DATE DEFAULT CURRENT_DATE,
    effective_until DATE,
    is_active BOOLEAN DEFAULT true,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert default commission rule
INSERT INTO commission_rules (name, description, platform_rate, payment_gateway_rate) VALUES
('Default Platform Commission', '10% platform fee + 2.75% payment gateway fee', 0.1000, 0.0275);

-- ================================
-- 6. Financial Reports (รายงานทางการเงิน)
-- ================================
CREATE TABLE financial_reports (
    report_id SERIAL PRIMARY KEY,
    
    -- Report metadata
    report_type VARCHAR(50) NOT NULL,                        -- daily, weekly, monthly, yearly
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    
    -- Summary data
    total_bookings INTEGER DEFAULT 0,
    total_revenue DECIMAL(12, 2) DEFAULT 0.00,               -- รายได้รวม
    total_commission DECIMAL(12, 2) DEFAULT 0.00,            -- ค่าคอมฯ รวม
    total_provider_earnings DECIMAL(12, 2) DEFAULT 0.00,     -- รายได้ provider รวม
    total_withdrawals DECIMAL(12, 2) DEFAULT 0.00,           -- ถอนเงินรวม
    total_subscriptions DECIMAL(12, 2) DEFAULT 0.00,         -- ค่าสมาชิกรวม
    
    -- Detailed breakdown
    breakdown JSONB,                                         -- ข้อมูลแยกตาม category, tier, etc.
    
    generated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    generated_by INTEGER REFERENCES users(user_id),
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ================================
-- Indexes for Performance
-- ================================
CREATE INDEX idx_bank_accounts_user ON bank_accounts(user_id) WHERE is_active = true;
CREATE INDEX idx_bank_accounts_default ON bank_accounts(user_id, is_default) WHERE is_default = true;
CREATE INDEX idx_wallets_user ON wallets(user_id);
CREATE INDEX idx_transactions_user ON transactions(user_id);
CREATE INDEX idx_transactions_type ON transactions(type);
CREATE INDEX idx_transactions_status ON transactions(status);
CREATE INDEX idx_transactions_booking ON transactions(booking_id);
CREATE INDEX idx_transactions_created ON transactions(created_at DESC);
CREATE INDEX idx_withdrawals_user ON withdrawals(user_id);
CREATE INDEX idx_withdrawals_status ON withdrawals(status);
CREATE INDEX idx_withdrawals_requested ON withdrawals(requested_at DESC);
CREATE INDEX idx_financial_reports_period ON financial_reports(period_start, period_end);

-- ================================
-- Triggers for Updated_at
-- ================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_bank_accounts_updated_at BEFORE UPDATE ON bank_accounts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_transactions_updated_at BEFORE UPDATE ON transactions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_withdrawals_updated_at BEFORE UPDATE ON withdrawals
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ================================
-- Comments
-- ================================
COMMENT ON TABLE bank_accounts IS 'เก็บข้อมูลบัญชีธนาคารของผู้ให้บริการสำหรับถอนเงิน';
COMMENT ON TABLE wallets IS 'กระเป๋าเงินของแต่ละ user สำหรับเก็บยอดเงิน';
COMMENT ON TABLE transactions IS 'บันทึกธุรกรรมทางการเงินทั้งหมดในระบบ';
COMMENT ON TABLE withdrawals IS 'คำขอถอนเงินของผู้ให้บริการ';
COMMENT ON TABLE commission_rules IS 'กฎการคำนวณค่าคอมมิชชั่นแพลตฟอร์ม';
COMMENT ON TABLE financial_reports IS 'รายงานสรุปทางการเงินรายวัน/เดือน/ปี';

COMMENT ON COLUMN wallets.available_balance IS 'ยอดเงินที่สามารถถอนได้ทันที';
COMMENT ON COLUMN wallets.pending_balance IS 'ยอดเงินที่รอยืนยัน (จาก booking ที่เสร็จใหม่ๆ)';
COMMENT ON COLUMN transactions.commission_amount IS 'ค่าคอมมิชชั่นแพลตฟอร์ม 10%';
COMMENT ON COLUMN transactions.net_amount IS 'จำนวนเงินสุทธิหลังหักค่าคอมฯ';
