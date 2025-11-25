# 💰 Financial System Documentation

## 📋 ภาพรวมระบบการเงิน

ระบบการเงินของ SkillMatch ประกอบด้วย:
- 💳 **ระบบการจ่ายเงิน** - Stripe Integration
- 💰 **กระเป๋าเงิน (Wallet)** - เก็บยอดเงินผู้ให้บริการ
- 🏦 **บัญชีธนาคาร** - สำหรับถอนเงิน
- 📊 **ธุรกรรม (Transactions)** - บันทึกทุกการเคลื่อนไหวเงิน
- 💸 **การถอนเงิน (Withdrawals)** - ขอถอนเงินไปบัญชีธนาคาร
- 📈 **รายงานทางการเงิน** - สรุปรายได้/รายจ่าย

---

## 💵 การคำนวณค่าธรรมเนียมและคอมมิชชั่น

### โครงสร้างค่าธรรมเนียม (Total 12.75%):
```
ราคาบริการ (ที่ลูกค้าจ่าย): 1,000 บาท
├── Payment Gateway (Stripe 2.75%): 27.5 บาท
├── Platform Commission (10%): 100 บาท
└── รายได้ Provider (87.25%): 872.5 บาท
```

**หลักการ:**
- ลูกค้าจ่ายราคาเต็ม (ไม่เห็นค่าธรรมเนียม)
- Provider เห็นว่าได้รับเงิน 87.25% (หักค่าธรรมเนียมรวม 12.75%)
- ระบบแสดงรายละเอียดค่าธรรมเนียมเฉพาะฝั่ง Provider

### การแจ้งเตือนค่าธรรมเนียม:

**1. เมื่อ Provider สร้าง Account:**
```
⚠️ ข้อมูลสำคัญเกี่ยวกับค่าธรรมเนียม

ระบบจะหักค่าธรรมเนียมรวม 12.75% จากทุกการจอง:

📊 รายละเอียดค่าธรรมเนียม:
• Payment Gateway (Stripe): 2.75%
• Platform Commission: 10.00%
• รวมทั้งหมด: 12.75%

คุณจะได้รับ: 87.25% ของราคาบริการ

ตัวอย่าง:
- ลูกค้าจ่าย: 1,000 บาท
- คุณได้รับ: 872.50 บาท
- ค่าธรรมเนียม: 127.50 บาท

✅ ยอมรับและดำเนินการต่อ
```

**2. ในหน้า Dashboard Provider:**
- แสดงยอดรายได้สุทธิ (หลังหักค่าธรรมเนียม 12.75%)
- แสดง breakdown ค่าธรรมเนียมแต่ละรายการ
- ไม่แสดงข้อมูลนี้ในหน้าของลูกค้า

### ตัวอย่างการคำนวณ:

| ราคาบริการ | Stripe (2.75%) | Platform (10%) | รวมค่าธรรมเนียม (12.75%) | รายได้ Provider (87.25%) |
|-----------|----------------|----------------|--------------------------|------------------------|
| 500 บาท   | 13.75 บาท      | 50 บาท         | 63.75 บาท                | 436.25 บาท              |
| 1,000 บาท | 27.50 บาท      | 100 บาท        | 127.50 บาท               | 872.50 บาท              |
| 2,000 บาท | 55.00 บาท      | 200 บาท        | 255.00 บาท               | 1,745.00 บาท            |
| 5,000 บาท | 137.50 บาท     | 500 บาท        | 637.50 บาท               | 4,362.50 บาท            |

**สูตรคำนวณ:**
```
ราคาบริการ = X บาท
ค่าธรรมเนียม Stripe = X × 0.0275
ค่าคอมมิชชั่นแพลตฟอร์ม = X × 0.10
รวมค่าธรรมเนียม = X × 0.1275
รายได้ Provider = X × 0.8725 (หรือ X - (X × 0.1275))
```

### Flow การจ่ายเงิน:

```
ลูกค้าจ่าย 1,000 บาท (ราคาเต็ม)
    ↓
Stripe Payment Gateway
    ├── Stripe หัก 2.75% = 27.5 บาท
    └── แพลตฟอร์มได้รับ 972.5 บาท
    ↓
ระบบคำนวณค่าธรรมเนียมรวม 12.75%:
    ├── Payment Gateway Fee: 2.75% (27.5 บาท)
    ├── Platform Commission: 10% (100 บาท)
    └── รวมค่าธรรมเนียม: 12.75% (127.5 บาท)
    ↓
Provider ได้รับ:
    ├── 1,000 - 127.5 = 872.5 บาท
    └── ไปที่ Pending Balance (รอ 7 วัน)
    ↓
หลัง 7 วัน → Available Balance (872.5 บาท)
    ↓
Provider ขอถอน
    ↓
GOD โอนจากบัญชีแพลตฟอร์ม 872.5 บาท
    ↓
ส่งสลิปการโอน (ซ่อนข้อมูล GOD):
    ├── WebSocket → Chat แบบ real-time
    ├── Email → ส่งพร้อม PDF สลิป
    └── หน้า Withdrawals → ดาวน์โหลด slip
```

**หมายเหตุ:**
- ลูกค้าจ่ายราคาเต็ม 1,000 บาท (ไม่เห็นค่าธรรมเนียม)
- Provider เห็นว่าได้รับ 872.5 บาท (หักค่าธรรมเนียม 12.75% แล้ว)
- GOD เก็บ 100 บาท (10%) + Stripe หัก 27.5 บาท (2.75%)

---

## 🗄️ Database Tables

### 1. bank_accounts - บัญชีธนาคาร

```sql
- bank_account_id (PK)
- user_id (FK → users)
- bank_name (VARCHAR): "ธนาคารกสิกรไทย"
- bank_code: "KBANK"
- account_number: "1234567890"
- account_name: "นาย สมชาย ใจดี"
- account_type: savings | current
- is_verified: BOOLEAN (ต้อง verify ก่อนถอนเงิน)
- is_default: BOOLEAN
- is_active: BOOLEAN
```

**Key Points:**
- ผู้ใช้เพิ่มได้หลายบัญชี
- ต้องผ่านการตรวจสอบจาก Admin (is_verified=true) ก่อนถอนเงิน
- ชื่อบัญชีต้องตรงกับ KYC

---

### 2. wallets - กระเป๋าเงิน

```sql
- wallet_id (PK)
- user_id (FK → users, UNIQUE)
- available_balance: ยอดพร้อมถอน
- pending_balance: ยอดรอยืนยัน (booking เสร็จแล้ว รอ 7 วัน)
- total_earned: รายได้สะสมทั้งหมด
- total_withdrawn: ถอนไปแล้วทั้งหมด
- total_commission_paid: ค่าคอมฯ ที่จ่ายไป
```

**Business Logic:**
- **Pending Balance**: เมื่อ booking สำเร็จ เงินจะเข้า pending_balance รอ 7 วัน
- **Available Balance**: หลัง 7 วัน เงินจะย้ายไป available_balance (พร้อมถอน)
- เหตุผล: ป้องกันการฉ้อโกง/ยกเลิกหลังจากการให้บริการ

---

### 3. transactions - ธุรกรรม

```sql
- transaction_id (PK)
- transaction_uuid (UUID, UNIQUE)
- user_id (FK → users)
- related_user_id: คู่กรณี (เช่น provider ในการจอง)
- type: ENUM (booking_payment, commission, provider_earning, withdrawal, etc.)
- status: ENUM (pending, processing, completed, failed, cancelled)
- amount: จำนวนเงินต้น
- commission_amount: ค่าคอมฯ
- net_amount: จำนวนสุทธิ
- booking_id (FK → bookings)
- payment_method: stripe, promptpay, bank_transfer
- payment_intent_id: Stripe Payment Intent ID
```

**Transaction Types:**
- `booking_payment`: ลูกค้าจ่ายเงิน
- `commission`: หักค่าคอมฯ แพลตฟอร์ม
- `provider_earning`: รายได้เข้า wallet provider
- `withdrawal`: ถอนเงิน
- `subscription_fee`: ค่าสมาชิก
- `admin_adjustment`: admin ปรับยอด
- `bonus`: โบนัส
- `penalty`: ค่าปรับ

---

### 4. withdrawals - การถอนเงิน

```sql
- withdrawal_id (PK)
- withdrawal_uuid (UUID)
- user_id (FK)
- bank_account_id (FK)
- requested_amount: จำนวนที่ขอถอน
- fee: ค่าธรรมเนียม (10 บาท)
- net_amount: ที่ได้รับจริง
- status: pending | approved | processing | completed | rejected | failed
- requested_at, approved_at, completed_at
- approved_by: admin ที่อนุมัติ
- transfer_reference: เลขที่อ้างอิงการโอน
- rejection_reason: เหตุผลปฏิเสธ
```

**Withdrawal Flow:**
```
1. Provider ขอถอน → status: pending
2. Admin ตรวจสอบ → approve/reject
3. ถ้า approve → status: approved
4. GOD/Admin โอนเงินจากบัญชีธนาคารแพลตฟอร์ม → status: processing
   - โอนจาก: บัญชี GOD (Platform Bank Account)
   - โอนไป: บัญชีของ Provider
   - บันทึก: เลขอ้างอิง + สลิป + บัญชีที่โอนออก
5. โอนสำเร็จ → status: completed (+ slip URL + transfer reference)
```

**Platform Bank Account (GOD Account):**
- บัญชีธนาคารหลักของแพลตฟอร์ม (ควบคุมโดย GOD)
- ทุกการถอนเงินต้องโอนผ่านบัญชีนี้
- เก็บประวัติการโอนทั้งหมดสำหรับ audit และภาษี
- ป้องกันการฉ้อโกงและให้ transparency

---

### 5. commission_rules - กฎค่าคอมมิชชั่น

```sql
- rule_id (PK)
- name: "Default Platform Commission"
- platform_rate: 0.1000 (10%)
- payment_gateway_rate: 0.0275 (2.75% Stripe)
- tier_id: สำหรับ tier พิเศษ (optional)
- effective_from, effective_until
- is_active: BOOLEAN
```

**Use Case:**
- แยก rate ตาม tier (Premium tier อาจได้ค่าคอมฯ น้อยกว่า)
- เปลี่ยน rate ในอนาคตได้ (effective_from)

---

### 6. financial_reports - รายงานทางการเงิน

```sql
- report_id (PK)
- report_type: daily | weekly | monthly | yearly
- period_start, period_end
- total_bookings: จำนวน booking
- total_revenue: รายได้รวม
- total_commission: ค่าคอมฯ รวม
- total_provider_earnings: รายได้ provider รวม
- total_withdrawals: ถอนเงินรวม
- total_subscriptions: ค่าสมาชิกรวม
- breakdown: JSONB (แยกตาม category, tier)
- generated_by: admin ที่สร้าง
```

---

## 🔄 Business Flow Examples

### 1. Customer จอง Provider

```sql
-- ลูกค้าจอง booking_id=123, ราคา 1,000 บาท
BEGIN TRANSACTION;

-- 1. สร้าง booking (ใน booking_handlers.go)
INSERT INTO bookings (...) VALUES (...);

-- 2. บันทึก transaction: ลูกค้าจ่าย
INSERT INTO transactions (
    user_id,                    -- customer_id
    related_user_id,            -- provider_id
    type,                       -- 'booking_payment'
    amount,                     -- 1000
    commission_amount,          -- 100 (10%)
    net_amount,                 -- 900
    booking_id,                 -- 123
    payment_intent_id           -- Stripe ID
) VALUES (...);

-- 3. บันทึก commission แพลตฟอร์ม
INSERT INTO transactions (
    type,                       -- 'commission'
    amount,                     -- 100
    description                 -- 'Platform commission 10%'
) VALUES (...);

-- 4. เพิ่มเงินเข้า pending_balance ของ provider
UPDATE wallets
SET pending_balance = pending_balance + 900,
    total_earned = total_earned + 900,
    total_commission_paid = total_commission_paid + 100
WHERE user_id = provider_id;

-- 5. บันทึก provider earning
INSERT INTO transactions (
    user_id,                    -- provider_id
    type,                       -- 'provider_earning'
    amount,                     -- 900
    status,                     -- 'pending'
    booking_id                  -- 123
) VALUES (...);

COMMIT;
```

### 2. หลัง 7 วัน: ย้ายเงินจาก Pending → Available

```sql
-- Cron Job รันทุกวัน
BEGIN TRANSACTION;

-- หา transactions ที่เกิน 7 วันแล้ว
UPDATE transactions
SET status = 'completed'
WHERE type = 'provider_earning'
  AND status = 'pending'
  AND created_at < NOW() - INTERVAL '7 days';

-- ย้ายเงิน pending → available
UPDATE wallets w
SET available_balance = available_balance + (
    SELECT COALESCE(SUM(net_amount), 0)
    FROM transactions t
    WHERE t.user_id = w.user_id
      AND t.type = 'provider_earning'
      AND t.status = 'completed'
      AND t.processed_at IS NULL
),
pending_balance = pending_balance - (
    SELECT COALESCE(SUM(net_amount), 0)
    FROM transactions t
    WHERE t.user_id = w.user_id
      AND t.type = 'provider_earning'
      AND t.status = 'completed'
      AND t.processed_at IS NULL
);

-- Mark transactions as processed
UPDATE transactions
SET processed_at = NOW()
WHERE type = 'provider_earning'
  AND status = 'completed'
  AND processed_at IS NULL;

COMMIT;
```

### 3. Provider ถอนเงิน

```sql
-- Provider ขอถอน 1,000 บาท
BEGIN TRANSACTION;

-- 1. สร้าง withdrawal request
INSERT INTO withdrawals (
    user_id, bank_account_id, 
    requested_amount, fee, net_amount, status
) VALUES (
    provider_id, bank_acc_id,
    1000, 10, 990, 'pending'
);

-- 2. หักเงินจาก available_balance
UPDATE wallets
SET available_balance = available_balance - 1000
WHERE user_id = provider_id;

-- 3. สร้าง transaction
INSERT INTO transactions (
    user_id, type, amount, commission_amount, net_amount,
    withdrawal_id, description
) VALUES (
    provider_id, 'withdrawal', 1000, 10, 990,
    withdrawal_id, 'Withdrawal request #...'
);

COMMIT;
```

### 4. Admin Approve & Complete Withdrawal (via Platform Bank)

```sql
BEGIN TRANSACTION;

-- 1. Admin อนุมัติ
UPDATE withdrawals
SET status = 'approved',
    approved_at = NOW(),
    approved_by = admin_id
WHERE withdrawal_id = xxx;

-- 2. GOD โอนเงิน 90% จากบัญชีแพลตฟอร์ม (10% เก็บไว้แล้ว)
-- บันทึกข้อมูลการโอน
UPDATE withdrawals
SET status = 'completed',
    completed_at = NOW(),
    transfer_reference = 'TXN123456',        -- เลขที่อ้างอิงจากธนาคาร
    transfer_slip_url = 'https://gcs.../slip_masked.jpg', -- สลิปที่ซ่อนข้อมูล GOD
    platform_bank_account_id = god_bank_id,  -- บัญชีแพลตฟอร์มที่โอนออก
    platform_transfer_by = god_user_id,      -- GOD ที่ทำการโอน
    platform_transfer_timestamp = NOW(),
    notes = 'Transferred 90% to provider (10% platform commission retained)'
WHERE withdrawal_id = xxx;

-- 3. อัพเดท wallet
UPDATE wallets
SET total_withdrawn = total_withdrawn + 900,  -- ยอดที่โอนจริง (90%)
    last_updated = NOW()
WHERE user_id = provider_id;

-- 4. อัพเดท transaction
UPDATE transactions
SET status = 'completed', 
    processed_at = NOW(),
    notes = 'Paid via platform bank account - 10% commission withheld'
WHERE withdrawal_id = xxx;

-- 5. สร้าง transfer log (audit trail)
INSERT INTO withdrawal_transfer_logs (
    withdrawal_id, platform_bank_account_id,
    platform_account_number, platform_account_name,
    provider_account_number, provider_account_name, provider_bank_name,
    transfer_amount, transfer_timestamp, transfer_reference,
    transfer_slip_url, transferred_by, transfer_method
) VALUES (
    xxx, god_bank_id,
    'XXX-X-XXXXX-X', 'บริษัท SkillMatch จำกัด',  -- ข้อมูล GOD (เก็บใน DB)
    'xxx-x-xxxxx-x', 'นาย Provider', 'ธนาคารกสิกรไทย',
    900.00, NOW(), 'TXN123456',
    'https://gcs.../slip_masked.jpg',  -- สลิปที่แก้ไขแล้ว
    god_user_id, 'mobile_banking'
);

-- 6. ส่ง Notification
INSERT INTO notifications (
    user_id, notification_type, title, message, 
    related_entity_type, related_entity_id, metadata
) VALUES (
    provider_id, 'withdrawal_completed',
    'การถอนเงินสำเร็จ',
    'โอนเงิน 900 บาท เข้าบัญชี xxx-x-xxxxx-x แล้ว',
    'withdrawal', xxx,
    '{"amount": 900, "transfer_ref": "TXN123456", "slip_url": "https://..."}' ::jsonb
);

COMMIT;

-- 7. ส่ง WebSocket notification (real-time)
-- (ทำใน Go code)
wsManager.BroadcastToUser(provider_id, {
    "type": "withdrawal_completed",
    "payload": {
        "withdrawal_id": xxx,
        "amount": 900,
        "transfer_reference": "TXN123456",
        "slip_url": "https://gcs.../slip_masked.jpg",  -- ซ่อนข้อมูล GOD
        "completed_at": "2025-11-14T10:30:00Z"
    }
})

-- 8. ส่ง Email พร้อมสลิป (ซ่อนข้อมูล GOD)
-- (ทำใน Go code - ใช้ SMTP หรือ SendGrid)
sendEmail({
    to: "provider@example.com",
    subject: "การถอนเงินสำเร็จ - SkillMatch",
    body: "โอนเงิน 900 บาท เข้าบัญชีของคุณแล้ว",
    attachments: ["slip_masked.jpg"]  -- สลิปที่แก้ไขแล้ว
})
```

**Security Benefits:**
- ✅ **Audit Trail**: ทุกการโอนผ่านบัญชีเดียวกัน (GOD account) ตรวจสอบได้
- ✅ **Fraud Prevention**: ไม่มีการโอนเงินตรงจากแพลตฟอร์มไป provider
- ✅ **Tax Compliance**: รวมรายการโอนเงินสำหรับภาษี
- ✅ **Fund Control**: GOD ควบคุมกระแสเงินสด
- ✅ **Dispute Resolution**: มีหลักฐานการโอนชัดเจนทุกรายการ

---

## 📊 Algorithm & Business Rules

### 1. การคำนวณค่าคอมมิชชั่นแบบ Dynamic

```go
// Get commission rule
func getCommissionRate(userID int, tierID *int) (float64, error) {
    var rate float64
    
    if tierID != nil {
        // Check for tier-specific rate
        err := db.QueryRow(`
            SELECT platform_rate FROM commission_rules
            WHERE tier_id = $1 AND is_active = true
              AND effective_from <= CURRENT_DATE
              AND (effective_until IS NULL OR effective_until >= CURRENT_DATE)
            ORDER BY effective_from DESC LIMIT 1
        `, tierID).Scan(&rate)
        
        if err == nil {
            return rate, nil
        }
    }
    
    // Fallback to default rate
    err := db.QueryRow(`
        SELECT platform_rate FROM commission_rules
        WHERE tier_id IS NULL AND is_active = true
        ORDER BY effective_from DESC LIMIT 1
    `).Scan(&rate)
    
    return rate, err
}

// Calculate commission
commissionRate := 0.10 // Default 10%
if rate, err := getCommissionRate(providerID, tierID); err == nil {
    commissionRate = rate
}

totalAmount := 1000.0
commissionAmount := totalAmount * commissionRate
providerEarning := totalAmount - commissionAmount
```

### 2. Algorithm สำหรับ Recommendation Provider

```go
// Recommend providers based on earning potential
func recommendProviders(location string, category int, priceRange string) []Provider {
    /*
    Scoring Algorithm:
    - Rating (40%): average_rating * 0.4
    - Availability (30%): available_slots / total_slots * 0.3
    - Price competitiveness (20%): (max_price - provider_price) / max_price * 0.2
    - Response time (10%): 1 - (avg_response_hours / 24) * 0.1
    */
    
    query := `
        SELECT p.user_id, p.username,
               -- Calculate score
               (COALESCE(AVG(r.rating), 0) * 0.4 +
                (COUNT(DISTINCT pa.slot_id)::FLOAT / NULLIF(COUNT(DISTINCT pa.slot_id) + COUNT(DISTINCT b.booking_id), 0)) * 0.3 +
                ((SELECT MAX(price) FROM service_packages) - MIN(sp.price)) / (SELECT MAX(price) FROM service_packages) * 0.2 +
                (1 - (AVG(EXTRACT(EPOCH FROM (m.created_at - m.sent_at)) / 3600) / 24)) * 0.1
               ) as score
        FROM users p
        LEFT JOIN reviews r ON p.user_id = r.provider_id
        LEFT JOIN service_packages sp ON p.user_id = sp.provider_id
        LEFT JOIN provider_availability pa ON p.user_id = pa.provider_id
        LEFT JOIN bookings b ON p.user_id = b.provider_id AND b.status IN ('pending', 'confirmed')
        LEFT JOIN messages m ON p.user_id = m.recipient_id
        WHERE p.verification_status = 'approved'
          AND p.province = ?
          AND EXISTS (
              SELECT 1 FROM provider_categories pc
              WHERE pc.provider_id = p.user_id AND pc.category_id = ?
          )
        GROUP BY p.user_id
        ORDER BY score DESC
        LIMIT 20
    `
    
    // Execute query...
}
```

### 3. Fraud Detection Algorithm

```go
// Detect suspicious withdrawal patterns
func detectSuspiciousWithdrawal(userID int, amount float64) (bool, string) {
    // Rule 1: Withdrawal > 80% of total earned in last 7 days
    var last7DaysEarning float64
    db.QueryRow(`
        SELECT COALESCE(SUM(net_amount), 0)
        FROM transactions
        WHERE user_id = $1 AND type = 'provider_earning'
          AND created_at >= NOW() - INTERVAL '7 days'
    `, userID).Scan(&last7DaysEarning)
    
    if amount > last7DaysEarning * 0.8 {
        return true, "Withdrawal amount exceeds 80% of recent earnings"
    }
    
    // Rule 2: More than 3 withdrawals in 24 hours
    var count24h int
    db.QueryRow(`
        SELECT COUNT(*) FROM withdrawals
        WHERE user_id = $1
          AND requested_at >= NOW() - INTERVAL '24 hours'
    `, userID).Scan(&count24h)
    
    if count24h >= 3 {
        return true, "Too many withdrawal requests in 24 hours"
    }
    
    // Rule 3: First withdrawal > 10,000 THB
    var totalWithdrawals int
    db.QueryRow(`
        SELECT COUNT(*) FROM withdrawals
        WHERE user_id = $1 AND status = 'completed'
    `, userID).Scan(&totalWithdrawals)
    
    if totalWithdrawals == 0 && amount > 10000 {
        return true, "First withdrawal exceeds 10,000 THB"
    }
    
    return false, ""
}
```

---

## 🔐 Security & Validation

### Bank Account Verification Process

```
1. User เพิ่มบัญชีธนาคาร → is_verified = false
2. Admin ตรวจสอบ:
   - ชื่อบัญชีตรงกับ KYC?
   - เลขบัญชีถูกต้อง?
   - Test โอนเงิน 1 บาท (optional)
3. Admin อนุมัติ → is_verified = true
4. User สามารถถอนเงินได้
```

### Withdrawal Rules

- ✅ Minimum: 100 บาท
- ✅ Fee: 10 บาท ต่อครั้ง
- ✅ Bank account ต้อง verified
- ✅ available_balance เพียงพอ
- ✅ Daily limit: 3 ครั้ง/วัน
- ✅ Max per withdrawal: 100,000 บาท

---

## 📱 API Endpoints Summary

### User Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/bank-accounts` | เพิ่มบัญชีธนาคาร |
| GET | `/bank-accounts` | ดูบัญชีของตัวเอง |
| DELETE | `/bank-accounts/:id` | ลบบัญชี |
| GET | `/wallet` | ดู wallet ของตัวเอง |
| POST | `/withdrawals` | ขอถอนเงิน |
| GET | `/withdrawals` | ดูประวัติถอนเงิน |
| GET | `/transactions` | ดูประวัติธุรกรรม |

### Admin Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/admin/withdrawals` | ดูคำขอถอนเงินทั้งหมด |
| POST | `/admin/withdrawals/:id/process` | อนุมัติ/ปฏิเสธ/complete |
| POST | `/admin/bank-accounts/:id/verify` | ยืนยันบัญชีธนาคาร |
| GET | `/admin/financial/summary` | สรุปรายได้วันนี้/เดือนนี้ |
| POST | `/admin/financial/report` | สร้างรายงาน |
| GET | `/admin/commission-rules` | ดูกฎค่าคอมมิชชั่น |
| PUT | `/admin/commission-rules/:id` | แก้ไขกฎ |
| GET | `/admin/wallets/:user_id` | ดู wallet user |
| POST | `/admin/wallets/:user_id/adjust` | ปรับยอด (bonus/penalty) |

### GOD Endpoints (Platform Bank Management)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/admin/god/platform-banks` | ดูบัญชีธนาคารแพลตฟอร์มทั้งหมด |
| POST | `/admin/god/platform-banks` | เพิ่มบัญชีแพลตฟอร์มใหม่ |
| PUT | `/admin/god/platform-banks/:id` | แก้ไขบัญชีแพลตฟอร์ม |
| POST | `/admin/god/platform-banks/:id/set-default` | ตั้งเป็นบัญชีหลัก |
| GET | `/admin/god/transfer-logs` | ดูประวัติการโอนเงินทั้งหมด |
| GET | `/admin/god/transfer-logs/:withdrawal_id` | ดูประวัติการโอนของ withdrawal นี้ |

---

## 🚀 Implementation Checklist

**Database:**
- [x] Create migration 013
- [x] 6 tables created (bank_accounts, wallets, transactions, withdrawals, commission_rules, financial_reports)
- [x] Indexes for performance
- [x] Default commission rule inserted (10%)

**Backend:**
- [x] financial_models.go - All data models
- [x] financial_handlers.go - User endpoints
- [x] financial_admin_handlers.go - Admin endpoints
- [ ] Update booking_handlers.go - Auto create transactions
- [ ] Update payment_handlers.go - Record subscription transactions
- [ ] Add cron job - Move pending → available (7 days)

**Frontend:**
- [ ] Provider Dashboard: Wallet component
- [ ] Provider Settings: Bank account management
- [ ] Provider Finance: Withdrawal request form
- [ ] Admin Panel: Withdrawal approval queue
- [ ] Admin Panel: Financial reports

---

## 📈 Next Steps

1. **Integrate with Booking System**
   - เมื่อ booking confirmed → สร้าง transactions
   - เมื่อ booking completed → เงินเข้า pending_balance

2. **Cron Jobs**
   - Daily: ย้าย pending → available (7 days)
   - Daily: Generate daily financial reports
   - Monthly: Send earning statements to providers

3. **Notifications**
   - แจ้งเตือนเมื่อเงินเข้า wallet
   - แจ้งเตือนเมื่อ withdrawal approved/rejected
   - แจ้งเตือนเมื่อเงินพร้อมถอน (pending → available)

4. **Dashboard & Analytics**
   - Provider: กราฟรายได้ย้อนหลัง 30 วัน
   - Admin: Dashboard รายได้แพลตฟอร์ม
   - Charts: Booking trends, Popular categories

5. **Advanced Features**
   - Auto-payout: โอนเงินอัตโนมัติทุกวันจันทร์
   - Tax documents: Export ภาษี (หนังสือรับรองฯ)
   - Referral system: แนะนำเพื่อนรับเงิน
   - Loyalty rewards: คะแนนสะสมสำหรับลูกค้า

---

## 🔐 Platform Bank Account Security

### GOD Account Requirements
- ต้องเป็น user ที่มี `tier_id = 5` และ `is_admin = true`
- เก็บข้อมูลบัญชีธนาคารที่ verified แล้วใน `platform_bank_accounts`
- ควรมี backup account อย่างน้อย 1 บัญชี

### Withdrawal Approval Process
```
1. Provider ขอถอน 1,000 บาท → withdrawal status: pending
   - ระบบตรวจสอบ available_balance
   - หักเงินจาก wallet ทันที (lock amount)

2. Admin/GOD ตรวจสอบ:
   ✓ Bank account verified?
   ✓ Available balance sufficient?
   ✓ No suspicious patterns?
   ✓ Provider มีประวัติดีไหม?

3. Admin/GOD อนุมัติ → status: approved
   - บัญชี GOD มียอด 100 บาท (10% commission)
   - ต้องโอน 900 บาท (90%) ไป provider

4. GOD เลือก platform bank account → โอนเงิน 900 บาท
   - โอนผ่าน mobile banking / internet banking
   - ได้สลิปจากธนาคาร (มีข้อมูลบัญชี GOD)

5. ระบบ Process สลิป:
   - Upload สลิปต้นฉบับไปยัง GCS (private)
   - ✂️ **แก้ไขสลิป**: Mask/Remove ข้อมูล GOD
     * ชื่อบัญชีผู้โอน → "SkillMatch Platform"
     * เลขบัญชีผู้โอน → "XXX-X-XXXXX-X"
     * ลบรายละเอียดที่ไม่จำเป็น
   - Upload สลิปที่แก้แล้วไปยัง GCS (public)

6. บันทึกข้อมูล:
   - transfer_reference (เลขอ้างอิงจากธนาคาร)
   - transfer_slip_url (URL สลิปที่แก้แล้ว - ซ่อนข้อมูล GOD)
   - platform_bank_account_id (บัญชีที่โอนออก - เก็บใน DB อย่างเดียว)
   - platform_transfer_by (GOD user_id)
   - platform_transfer_timestamp

7. สร้าง audit log ใน withdrawal_transfer_logs
   - เก็บข้อมูล GOD ครบถ้วน (สำหรับ internal audit)
   - แยกจาก public data

8. ส่งการแจ้งเตือน:
   ✉️ **Email**: ส่งสลิปที่แก้แล้ว พร้อมรายละเอียด
   💬 **Chat/WebSocket**: แจ้งแบบ real-time พร้อม link ดาวน์โหลดสลิป
   🔔 **Notification**: แจ้งเตือนในแอป

9. Status → completed
   - Provider ได้รับเงิน 900 บาท
   - GOD เก็บค่าคอมมิชชั่น 100 บาท ไว้ในบัญชีแพลตฟอร์ม
```

### Audit Trail Benefits
✅ **Transparency**: ทุกการโอนมี log ครบถ้วน  
✅ **Accountability**: รู้ว่าใครโอน เมื่อไหร่ จากบัญชีไหน  
✅ **Tax Compliance**: Export รายงานสำหรับภาษีได้ทันที  
✅ **Fraud Detection**: ตรวจจับรูปแบบผิดปกติได้ง่าย  
✅ **Dispute Resolution**: มีหลักฐานชัดเจนทุกรายการ  

---

---

## 🎭 Transfer Slip Masking System

### Why Mask Transfer Slips?
- **Privacy**: ป้องกันการเปิดเผยข้อมูลบัญชี GOD
- **Security**: ลดความเสี่ยงจากการโจมตี
- **Professionalism**: Provider เห็นแค่ "SkillMatch Platform" ไม่เห็นข้อมูลส่วนตัว GOD

### Slip Masking Process

```go
// 1. Upload Original Slip (Private)
originalSlipURL := uploadToGCS(slipFile, "transfer-slips/original/")
// URL: gs://bucket/transfer-slips/original/withdrawal_123_original.jpg

// 2. Mask Sensitive Information
maskedSlipFile := maskTransferSlip(slipFile, {
    "senderName": "SkillMatch Platform",         // แทนชื่อ GOD
    "senderAccount": "XXX-X-XXXXX-X",           // แทนเลขบัญชี GOD
    "hideFields": ["branch", "idCard"]          // ซ่อนข้อมูลเพิ่มเติม
})

// 3. Upload Masked Slip (Public)
maskedSlipURL := uploadToGCS(maskedSlipFile, "transfer-slips/public/")
// URL: https://storage.googleapis.com/.../withdrawal_123_masked.jpg

// 4. Store Both URLs
updateWithdrawal({
    original_slip_url: originalSlipURL,  // เก็บไว้ใน DB (ไม่แสดง provider)
    transfer_slip_url: maskedSlipURL     // แสดงให้ provider เห็น
})
```

### Masking Rules

| Field | Original | Masked |
|-------|----------|--------|
| ชื่อผู้โอน | นาย GOD Master | **SkillMatch Platform** |
| เลขบัญชีผู้โอน | 123-4-56789-0 | **XXX-X-XXXXX-X** |
| สาขา | สีลม | *(ลบออก)* |
| เลขบัตรประชาชน | 1-XXXX-XXXXX-XX-X | *(ลบออก)* |
| ชื่อผู้รับ | นาย Provider | **นาย Provider** ✅ |
| เลขบัญชีผู้รับ | 987-6-54321-0 | **987-6-54321-0** ✅ |
| จำนวนเงิน | 900.00 | **900.00** ✅ |
| เลขที่อ้างอิง | TXN123456 | **TXN123456** ✅ |

---

## 📧 Notification System for Withdrawals

### 1. WebSocket Notification (Real-time Chat)

```javascript
// Provider รับ notification ทาง WebSocket
{
    "type": "withdrawal_completed",
    "payload": {
        "withdrawal_id": 123,
        "withdrawal_uuid": "abc-def-ghi",
        "requested_amount": 1000.00,
        "commission": 100.00,        // 10%
        "net_amount": 900.00,        // ที่ได้รับจริง
        "transfer_reference": "TXN123456",
        "transfer_slip_url": "https://storage.../withdrawal_123_masked.jpg",
        "completed_at": "2025-11-14T10:30:00Z",
        "message": "โอนเงิน 900 บาท เข้าบัญชี xxx-x-xxxxx-x สำเร็จ"
    }
}
```

**Go Implementation:**
```go
// In financial_admin_handlers.go
func completeWithdrawal(withdrawalID int, slipURL string) {
    // ... update database ...
    
    // Send WebSocket notification
    wsManager.BroadcastToUser(providerID, WebSocketMessage{
        Type: "withdrawal_completed",
        Payload: map[string]interface{}{
            "withdrawal_id":      withdrawalID,
            "net_amount":         netAmount,
            "transfer_slip_url":  maskedSlipURL,  // สลิปที่แก้แล้ว
            "message":            "โอนเงินสำเร็จ",
        },
    })
}
```

### 2. Email Notification

**Email Template:**

```
เรียน [Provider Name]

การถอนเงินของคุณสำเร็จแล้ว

รายละเอียด:
- จำนวนที่ขอถอน: 1,000.00 บาท
- ค่าธรรมเนียม: 10.00 บาท
- ค่าคอมมิชชั่นแพลตฟอร์ม (10%): 100.00 บาท
- โอนเข้าบัญชีของคุณ: 900.00 บาท

เลขที่อ้างอิง: TXN123456
เวลาโอน: 14 พ.ย. 2568 เวลา 10:30 น.

กรุณาตรวจสอบยอดเงินในบัญชีธนาคารของคุณภายใน 24 ชั่วโมง

ดูสลิปการโอน: [ดาวน์โหลดที่นี่]

ขอบคุณที่ใช้บริการ SkillMatch
```

**Go Implementation:**
```go
import "net/smtp"

func sendWithdrawalEmail(providerEmail string, data WithdrawalData) error {
    // Email content with masked slip URL
    body := fmt.Sprintf(`
        เรียน %s
        
        การถอนเงินของคุณสำเร็จแล้ว
        
        - จำนวนที่โอน: %.2f บาท
        - เลขที่อ้างอิง: %s
        
        ดูสลิปการโอน: %s
    `, data.Name, data.NetAmount, data.Reference, data.MaskedSlipURL)
    
    // Send via SMTP or SendGrid
    return sendEmail(providerEmail, "การถอนเงินสำเร็จ", body)
}
```

### 3. In-App Notification

```sql
INSERT INTO notifications (
    user_id, notification_type, title, message,
    related_entity_type, related_entity_id,
    is_read, metadata
) VALUES (
    provider_id,
    'withdrawal_completed',
    'การถอนเงินสำเร็จ',
    'โอนเงิน 900 บาท เข้าบัญชีของคุณแล้ว (หักค่าคอมมิชชั่น 10%)',
    'withdrawal',
    withdrawal_id,
    false,
    jsonb_build_object(
        'amount', 900.00,
        'commission', 100.00,
        'transfer_ref', 'TXN123456',
        'slip_url', 'https://storage.../masked.jpg'
    )
);
```

---

## 💼 GOD Commission Tracking

### Commission Balance Tracking

```sql
-- เพิ่ม table สำหรับติดตามค่าคอมมิชชั่น GOD
CREATE TABLE god_commission_balance (
    balance_id SERIAL PRIMARY KEY,
    god_user_id INTEGER NOT NULL REFERENCES users(user_id),
    platform_bank_account_id INTEGER REFERENCES platform_bank_accounts(platform_bank_id),
    
    -- Balances
    total_commission_collected DECIMAL(12, 2) DEFAULT 0.00,  -- รวมค่าคอมฯ ที่เก็บได้
    total_transferred DECIMAL(12, 2) DEFAULT 0.00,            -- รวมที่โอนไป provider
    current_balance DECIMAL(12, 2) DEFAULT 0.00,              -- ยอดคงเหลือในบัญชี GOD
    
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- เพิ่ม trigger อัพเดทยอดอัตโนมัติ
CREATE OR REPLACE FUNCTION update_god_commission_balance()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status = 'completed' AND OLD.status != 'completed' THEN
        UPDATE god_commission_balance
        SET total_transferred = total_transferred + NEW.net_amount,
            current_balance = total_commission_collected - (total_transferred + NEW.net_amount),
            last_updated = NOW()
        WHERE platform_bank_account_id = NEW.platform_bank_account_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_god_balance
AFTER UPDATE ON withdrawals
FOR EACH ROW
EXECUTE FUNCTION update_god_commission_balance();
```

**GOD Dashboard:**
- แสดงยอดค่าคอมมิชชั่นรวม
- แสดงยอดที่โอนไปแล้ว
- แสดงยอดคงเหลือในบัญชีแพลตฟอร์ม
- Export รายงานสำหรับบัญชี/ภาษี

---

---

## 🎨 UI/UX Guidelines for Fee Display

### Provider Side (แสดงค่าธรรมเนียม)

**1. Provider Registration:**
```jsx
// Modal แสดงเมื่อสร้าง account
<FeeNotificationModal>
  <Icon>⚠️</Icon>
  <Title>ข้อมูลสำคัญเกี่ยวกับค่าธรรมเนียม</Title>
  
  <FeeBreakdown>
    <Fee>Payment Gateway (Stripe): 2.75%</Fee>
    <Fee>Platform Commission: 10.00%</Fee>
    <Divider />
    <TotalFee>รวมค่าธรรมเนียม: 12.75%</TotalFee>
  </FeeBreakdown>
  
  <Example>
    <Label>ตัวอย่าง:</Label>
    <Item>ลูกค้าจ่าย: 1,000 บาท</Item>
    <Item>คุณได้รับ: 872.50 บาท</Item>
    <Item>ค่าธรรมเนียม: 127.50 บาท</Item>
  </Example>
  
  <Checkbox>
    <Input type="checkbox" required />
    <Label>ฉันรับทราบและยอมรับเงื่อนไขค่าธรรมเนียม 12.75%</Label>
  </Checkbox>
  
  <Button>ยอมรับและดำเนินการต่อ</Button>
</FeeNotificationModal>
```

**2. Provider Dashboard - Wallet:**
```jsx
<WalletCard>
  <Balance>
    <Label>ยอดเงินพร้อมถอน</Label>
    <Amount>872.50 บาท</Amount>
  </Balance>
  
  <FeeInfo>
    <Icon>ℹ️</Icon>
    <Text>
      ยอดเงินนี้หักค่าธรรมเนียม 12.75% แล้ว
      (Payment Gateway 2.75% + Platform 10%)
    </Text>
  </FeeInfo>
</WalletCard>
```

**3. Transaction History:**
```jsx
<TransactionItem>
  <Header>
    <Type>รายได้จากการจอง #12345</Type>
    <Date>14 พ.ย. 2568</Date>
  </Header>
  
  <AmountBreakdown>
    <Row>
      <Label>ราคาบริการ</Label>
      <Value>1,000.00 บาท</Value>
    </Row>
    <Row type="deduction">
      <Label>Payment Gateway (2.75%)</Label>
      <Value>-27.50 บาท</Value>
    </Row>
    <Row type="deduction">
      <Label>Platform Commission (10%)</Label>
      <Value>-100.00 บาท</Value>
    </Row>
    <Divider />
    <Row type="total">
      <Label>ยอดที่คุณได้รับ</Label>
      <Value>872.50 บาท</Value>
    </Row>
  </AmountBreakdown>
</TransactionItem>
```

**4. Service Package Creation:**
```jsx
<PackageForm>
  <PriceInput>
    <Label>ราคาบริการ (ที่ลูกค้าจะจ่าย)</Label>
    <Input type="number" placeholder="1000" />
  </PriceInput>
  
  <FeeCalculator>
    <Row>
      <Label>ราคาที่ลูกค้าจ่าย:</Label>
      <Value>1,000.00 บาท</Value>
    </Row>
    <Row>
      <Label>ค่าธรรมเนียมรวม (12.75%):</Label>
      <Value type="negative">-127.50 บาท</Value>
    </Row>
    <Row type="highlight">
      <Label>คุณจะได้รับ:</Label>
      <Value>872.50 บาท</Value>
    </Row>
  </FeeCalculator>
</PackageForm>
```

### Client Side (ไม่แสดงค่าธรรมเนียม)

**1. Service Package List:**
```jsx
<PackageCard>
  <Title>แพ็กเกจ 1 ชั่วโมง</Title>
  <Price>1,000 บาท</Price>  {/* ราคาเต็ม ไม่มี fee */}
  <Button>จองเลย</Button>
</PackageCard>
```

**2. Booking Checkout:**
```jsx
<CheckoutSummary>
  <Item>
    <Label>แพ็กเกจ 1 ชั่วโมง</Label>
    <Price>1,000 บาท</Price>
  </Item>
  <Divider />
  <Total>
    <Label>ยอดชำระทั้งหมด</Label>
    <Amount>1,000 บาท</Amount>  {/* ไม่มี +fee */}
  </Total>
  <Button>ชำระเงิน</Button>
</CheckoutSummary>
```

---

## 📧 Email Templates

### 1. Provider Registration Confirmation

```html
<EmailTemplate>
  <Header>ยินดีต้อนรับสู่ SkillMatch</Header>
  
  <Body>
    <Greeting>สวัสดี [Provider Name],</Greeting>
    
    <Message>
      ยินดีต้อนรับเข้าสู่ SkillMatch! บัญชีของคุณถูกสร้างเรียบร้อยแล้ว
    </Message>
    
    <FeeNotification>
      <Icon>💰</Icon>
      <Title>ข้อมูลสำคัญเกี่ยวกับค่าธรรมเนียม</Title>
      
      <Text>
        ระบบจะหักค่าธรรมเนียมรวม 12.75% จากรายได้ของคุณ:
      </Text>
      
      <FeeList>
        <Item>• Payment Gateway (Stripe): 2.75%</Item>
        <Item>• Platform Commission: 10.00%</Item>
      </FeeList>
      
      <Example>
        <Strong>ตัวอย่าง:</Strong>
        <br/>
        ลูกค้าจ่าย 1,000 บาท → คุณได้รับ 872.50 บาท
        <br/>
        (หักค่าธรรมเนียม 127.50 บาท)
      </Example>
    </FeeNotification>
    
    <CTA>
      <Button href="/dashboard">ไปยัง Dashboard</Button>
    </CTA>
  </Body>
  
  <Footer>SkillMatch Platform</Footer>
</EmailTemplate>
```

### 2. Booking Completed (Provider)

```html
<EmailTemplate>
  <Header>🎉 คุณมีรายได้ใหม่!</Header>
  
  <Body>
    <Booking>
      <Label>Booking #12345</Label>
      <Date>14 พ.ย. 2568</Date>
    </Booking>
    
    <Breakdown>
      <Row>
        <Label>ราคาบริการ:</Label>
        <Value>1,000.00 บาท</Value>
      </Row>
      <Row type="deduction">
        <Label>Payment Gateway (2.75%):</Label>
        <Value>-27.50 บาท</Value>
      </Row>
      <Row type="deduction">
        <Label>Platform Commission (10%):</Label>
        <Value>-100.00 บาท</Value>
      </Row>
      <Divider />
      <Row type="total">
        <Label>ยอดที่คุณได้รับ:</Label>
        <Value>872.50 บาท</Value>
      </Row>
    </Breakdown>
    
    <Status>
      เงินจะพร้อมถอนใน 7 วัน (21 พ.ย. 2568)
    </Status>
  </Body>
</EmailTemplate>
```

---

## 🔢 Database Updates for Fee Calculation

**Update commission_rules table:**
```sql
-- เปลี่ยนจาก 10% เป็น 12.75%
UPDATE commission_rules
SET platform_rate = 0.1000,           -- Platform Commission 10%
    payment_gateway_rate = 0.0275,    -- Stripe 2.75%
    description = 'Total fee: 12.75% (Platform 10% + Payment Gateway 2.75%)'
WHERE rule_id = 1;

-- เพิ่ม field สำหรับ total_rate
ALTER TABLE commission_rules
ADD COLUMN IF NOT EXISTS total_rate DECIMAL(5, 4) GENERATED ALWAYS AS (platform_rate + payment_gateway_rate) STORED;

COMMENT ON COLUMN commission_rules.total_rate IS 'รวมค่าธรรมเนียมทั้งหมด (auto-calculated)';
```

**Update calculation logic:**
```sql
-- Function คำนวณรายได้ provider
CREATE OR REPLACE FUNCTION calculate_provider_earning(booking_amount DECIMAL)
RETURNS TABLE (
    gross_amount DECIMAL,
    stripe_fee DECIMAL,
    platform_commission DECIMAL,
    total_fee DECIMAL,
    net_amount DECIMAL
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        booking_amount,                           -- ยอดเต็ม
        booking_amount * 0.0275,                  -- Stripe 2.75%
        booking_amount * 0.1000,                  -- Platform 10%
        booking_amount * 0.1275,                  -- รวม 12.75%
        booking_amount * 0.8725;                  -- Provider ได้รับ 87.25%
END;
$$ LANGUAGE plpgsql;

-- ตัวอย่างการใช้งาน
SELECT * FROM calculate_provider_earning(1000.00);
-- Result:
-- gross_amount | stripe_fee | platform_commission | total_fee | net_amount
-- 1000.00      | 27.50      | 100.00              | 127.50    | 872.50
```

---

**Migrations:**  
- 013_add_financial_system.sql (Wallet, Transactions, Withdrawals)  
- 016_add_platform_bank_tracking.sql (Platform Bank Accounts, Transfer Logs)  
- 017_add_god_commission_tracking.sql (GOD Commission Balance)  
- 018_update_fee_structure.sql (Update to 12.75% total fee)

**Last Updated:** November 14, 2025  
**Status:** ✅ Database Ready, 🔄 Implementation Needed:
- Update commission calculation to 12.75%
- Add provider registration fee notification
- Update dashboard to show net amount (after 12.75% fee)
- Hide fee information from client-facing pages
- Slip masking logic
- WebSocket notification on withdrawal completion
- Email notification with masked slip
