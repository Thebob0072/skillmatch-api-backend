-- Migration: Add Escrow System for Remaining Payments
-- Created: 2024-12-17

-- 1. Create escrow_payments table
CREATE TABLE IF NOT EXISTS escrow_payments (
    escrow_id SERIAL PRIMARY KEY,
    booking_id INTEGER NOT NULL REFERENCES bookings(booking_id),
    
    -- จำนวนเงินที่ล็อคไว้
    amount DECIMAL(10,2) NOT NULL,
    
    -- สถานะ escrow
    status VARCHAR(50) NOT NULL DEFAULT 'locked',
    -- locked: เงินถูกล็อคไว้
    -- released: ปลดล็อคเงินให้ provider แล้ว
    -- refunded: คืนเงินให้ client แล้ว
    -- disputed: มีข้อพิพาท รอ admin ตัดสิน
    -- partially_refunded: แบ่งเงินระหว่าง client และ provider
    
    -- Timestamps สำคัญ
    locked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    client_confirmed_at TIMESTAMP,
    provider_completed_at TIMESTAMP,
    released_at TIMESTAMP,
    
    -- Dispute information
    dispute_reason TEXT,
    disputed_at TIMESTAMP,
    admin_decision VARCHAR(50),
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT valid_status CHECK (status IN (
        'locked', 'released', 'refunded', 'disputed', 
        'partially_refunded'
    ))
);

-- 2. Add columns to bookings table
ALTER TABLE bookings
ADD COLUMN IF NOT EXISTS deposit_required BOOLEAN DEFAULT false,
ADD COLUMN IF NOT EXISTS deposit_amount DECIMAL(10,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS deposit_paid BOOLEAN DEFAULT false,
ADD COLUMN IF NOT EXISTS deposit_paid_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS remaining_amount DECIMAL(10,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS escrow_locked BOOLEAN DEFAULT false,

-- Provider arrival
ADD COLUMN IF NOT EXISTS provider_arrived_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS client_confirmed_arrival_at TIMESTAMP,

-- Service completion
ADD COLUMN IF NOT EXISTS provider_completed_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS provider_completion_notes TEXT,
ADD COLUMN IF NOT EXISTS client_confirmed_at TIMESTAMP,

-- Dispute
ADD COLUMN IF NOT EXISTS dispute_reason TEXT,
ADD COLUMN IF NOT EXISTS dispute_description TEXT,
ADD COLUMN IF NOT EXISTS disputed_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS admin_decision VARCHAR(50),
ADD COLUMN IF NOT EXISTS admin_decision_notes TEXT,
ADD COLUMN IF NOT EXISTS resolved_by_admin_id INTEGER REFERENCES users(user_id),
ADD COLUMN IF NOT EXISTS resolved_at TIMESTAMP;

-- 3. Update bookings status to include new states
-- New statuses: provider_arrived, disputed, dispute_resolved, funds_released

-- 4. Add payment_type to payments table
ALTER TABLE payments
ADD COLUMN IF NOT EXISTS payment_type VARCHAR(20) DEFAULT 'full';
-- 'deposit' = มัดจำ 10%
-- 'full' = เต็มจำนวน
-- 'remaining' = ส่วนที่เหลือ

-- 5. Create indexes
CREATE INDEX IF NOT EXISTS idx_escrow_booking ON escrow_payments(booking_id);
CREATE INDEX IF NOT EXISTS idx_escrow_status ON escrow_payments(status);
CREATE INDEX IF NOT EXISTS idx_bookings_deposit_paid ON bookings(deposit_paid);
CREATE INDEX IF NOT EXISTS idx_bookings_escrow_locked ON bookings(escrow_locked);

-- 6. Add comments
COMMENT ON TABLE escrow_payments IS 'Escrow system for holding remaining payments until service completion';
COMMENT ON COLUMN escrow_payments.status IS 'locked=เงินถูกล็อค, released=จ่ายให้ provider, refunded=คืนให้ client, disputed=มีข้อพิพาท';
COMMENT ON COLUMN bookings.deposit_required IS 'ต้องจ่ายมัดจำ 10% ก่อนหรือไม่';
COMMENT ON COLUMN bookings.escrow_locked IS 'เงินส่วนที่เหลือถูกล็อคใน escrow หรือไม่';

-- 7. Update existing bookings (optional - for testing)
-- คำนวณ deposit_amount และ remaining_amount สำหรับ bookings ที่มีอยู่
UPDATE bookings
SET 
    deposit_amount = total_amount * 0.10,
    remaining_amount = total_amount * 0.90
WHERE deposit_required = true AND deposit_amount = 0;
