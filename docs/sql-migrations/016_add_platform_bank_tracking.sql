-- Migration 016: Add Platform Bank Account Tracking for Withdrawals
-- เพิ่มการติดตามบัญชีธนาคารแพลตฟอร์มที่ใช้โอนเงิน

-- ================================
-- 1. Platform Bank Accounts (บัญชีธนาคารแพลตฟอร์ม - GOD)
-- ================================
CREATE TABLE platform_bank_accounts (
    platform_bank_id SERIAL PRIMARY KEY,
    
    -- Bank Information
    bank_name VARCHAR(100) NOT NULL,           -- ชื่อธนาคาร (e.g., "ธนาคารกสิกรไทย")
    bank_code VARCHAR(10),                     -- รหัสธนาคาร (e.g., "KBANK")
    account_number VARCHAR(50) NOT NULL UNIQUE, -- เลขที่บัญชีแพลตฟอร์ม
    account_name VARCHAR(200) NOT NULL,        -- ชื่อบัญชี (ชื่อบริษัท/GOD)
    account_type VARCHAR(20) DEFAULT 'current', -- savings, current
    branch_name VARCHAR(100),                  -- สาขา
    
    -- Account Details
    account_holder VARCHAR(200),               -- ผู้ถือบัญชี (ชื่อจริง GOD/บริษัท)
    account_holder_id_card VARCHAR(50),        -- เลขบัตรประชาชน/เลขทะเบียนนิติบุคคล
    
    -- Balance Tracking (Optional - for reconciliation)
    current_balance DECIMAL(12, 2) DEFAULT 0.00,
    total_inflow DECIMAL(12, 2) DEFAULT 0.00,  -- เงินเข้ารวม
    total_outflow DECIMAL(12, 2) DEFAULT 0.00, -- เงินออกรวม (withdrawals)
    
    -- Status
    is_active BOOLEAN DEFAULT true,            -- ใช้งานอยู่
    is_default BOOLEAN DEFAULT false,          -- บัญชีหลัก (ใช้โอนเงินอัตโนมัติ)
    
    -- Ownership
    owned_by INTEGER REFERENCES users(user_id), -- GOD user_id (tier_id = 5)
    
    notes TEXT,                                -- หมายเหตุ
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT one_default_platform_account CHECK (
        is_default = false OR 
        (SELECT COUNT(*) FROM platform_bank_accounts WHERE is_default = true) <= 1
    )
);

-- ================================
-- 2. Add Platform Bank Reference to Withdrawals
-- ================================
-- เพิ่ม column สำหรับติดตามว่าโอนจากบัญชีแพลตฟอร์มไหน
ALTER TABLE withdrawals 
    ADD COLUMN platform_bank_account_id INTEGER REFERENCES platform_bank_accounts(platform_bank_id),
    ADD COLUMN platform_transfer_timestamp TIMESTAMP, -- เวลาที่โอนจากธนาคารแพลตฟอร์ม
    ADD COLUMN platform_transfer_by INTEGER REFERENCES users(user_id); -- GOD/Admin ที่ทำการโอน

-- ================================
-- 3. Insert Default Platform Bank Account (GOD Account)
-- ================================
-- สมมติว่า GOD user_id = 1
-- แก้ไขข้อมูลตามจริง
INSERT INTO platform_bank_accounts (
    bank_name,
    bank_code,
    account_number,
    account_name,
    account_type,
    branch_name,
    account_holder,
    account_holder_id_card,
    is_active,
    is_default,
    owned_by,
    notes
) VALUES (
    'ธนาคารกสิกรไทย',                        -- ชื่อธนาคาร
    'KBANK',                                -- รหัสธนาคาร
    'XXX-X-XXXXX-X',                        -- เลขบัญชี (แก้ตามจริง)
    'บริษัท SkillMatch จำกัด',              -- ชื่อบัญชี
    'current',                              -- ประเภทบัญชี
    'สาขาสีลม',                             -- สาขา
    'นาย GOD Master',                       -- ชื่อผู้ถือบัญชี
    '1-XXXX-XXXXX-XX-X',                   -- เลขบัตรประชาชน
    true,                                   -- is_active
    true,                                   -- is_default (บัญชีหลัก)
    1,                                      -- owned_by (user_id = 1 = GOD)
    'บัญชีธนาคารหลักของแพลตฟอร์ม ใช้สำหรับโอนเงินให้ providers ทั้งหมด'
);

-- ================================
-- 4. Withdrawal Transfer History (ประวัติการโอนเงิน)
-- ================================
-- ตารางเก็บประวัติการโอนเงินแต่ละครั้ง (for audit trail)
CREATE TABLE withdrawal_transfer_logs (
    log_id SERIAL PRIMARY KEY,
    withdrawal_id INTEGER NOT NULL REFERENCES withdrawals(withdrawal_id),
    
    -- Platform Bank Details
    platform_bank_account_id INTEGER NOT NULL REFERENCES platform_bank_accounts(platform_bank_id),
    platform_account_number VARCHAR(50) NOT NULL, -- เลขบัญชีที่โอนออก
    platform_account_name VARCHAR(200) NOT NULL,  -- ชื่อบัญชีที่โอนออก
    
    -- Provider Bank Details (snapshot)
    provider_account_number VARCHAR(50) NOT NULL, -- เลขบัญชีที่โอนเข้า
    provider_account_name VARCHAR(200) NOT NULL,  -- ชื่อบัญชีที่โอนเข้า
    provider_bank_name VARCHAR(100) NOT NULL,     -- ธนาคารของ provider
    
    -- Transfer Details
    transfer_amount DECIMAL(12, 2) NOT NULL,      -- จำนวนเงินที่โอน (net_amount)
    transfer_timestamp TIMESTAMP NOT NULL,        -- เวลาที่โอน
    transfer_reference VARCHAR(100),              -- เลขที่อ้างอิงจากธนาคาร
    transfer_slip_url TEXT,                       -- URL สลิปการโอน
    
    -- Who did this transfer
    transferred_by INTEGER NOT NULL REFERENCES users(user_id), -- GOD/Admin
    transfer_method VARCHAR(50),                  -- mobile_banking, internet_banking, atm, counter
    
    -- Verification
    verified BOOLEAN DEFAULT false,               -- ยืนยันการโอนแล้ว
    verified_at TIMESTAMP,
    verified_by INTEGER REFERENCES users(user_id),
    
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ================================
-- 5. Indexes
-- ================================
CREATE INDEX idx_platform_bank_active ON platform_bank_accounts(is_active) WHERE is_active = true;
CREATE INDEX idx_platform_bank_default ON platform_bank_accounts(is_default) WHERE is_default = true;
CREATE INDEX idx_platform_bank_owner ON platform_bank_accounts(owned_by);

CREATE INDEX idx_withdrawals_platform_bank ON withdrawals(platform_bank_account_id);
CREATE INDEX idx_withdrawal_transfer_logs_withdrawal ON withdrawal_transfer_logs(withdrawal_id);
CREATE INDEX idx_withdrawal_transfer_logs_platform_bank ON withdrawal_transfer_logs(platform_bank_account_id);
CREATE INDEX idx_withdrawal_transfer_logs_transferrer ON withdrawal_transfer_logs(transferred_by);

-- ================================
-- 6. Trigger for Updated_at
-- ================================
CREATE TRIGGER update_platform_bank_accounts_updated_at BEFORE UPDATE ON platform_bank_accounts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ================================
-- 7. Comments
-- ================================
COMMENT ON TABLE platform_bank_accounts IS 'บัญชีธนาคารแพลตฟอร์ม (GOD) สำหรับโอนเงินให้ providers';
COMMENT ON TABLE withdrawal_transfer_logs IS 'ประวัติการโอนเงินแต่ละครั้ง (audit trail สำหรับตรวจสอบ)';

COMMENT ON COLUMN platform_bank_accounts.is_default IS 'บัญชีหลักที่ใช้โอนเงินอัตโนมัติ (ควรมีแค่ 1 บัญชี)';
COMMENT ON COLUMN platform_bank_accounts.owned_by IS 'GOD user_id (tier_id = 5) ที่เป็นเจ้าของบัญชีนี้';
COMMENT ON COLUMN withdrawals.platform_bank_account_id IS 'บัญชีแพลตฟอร์มที่ใช้โอนเงินครั้งนี้';
COMMENT ON COLUMN withdrawals.platform_transfer_by IS 'GOD/Admin ที่ทำการโอนเงิน';
COMMENT ON COLUMN withdrawal_transfer_logs.transfer_method IS 'วิธีการโอน: mobile_banking, internet_banking, atm, counter';

-- ================================
-- 8. Security: Prevent Deletion
-- ================================
-- ป้องกันการลบบัญชีแพลตฟอร์มที่มี withdrawals อ้างอิง
ALTER TABLE platform_bank_accounts 
    ADD CONSTRAINT no_delete_if_used CHECK (
        is_active = true OR 
        NOT EXISTS (
            SELECT 1 FROM withdrawals 
            WHERE platform_bank_account_id = platform_bank_id
        )
    );

```