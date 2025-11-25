-- Migration 020: Face Verification System for Provider KYC
-- เพิ่มระบบแสกนใบหน้าเพื่อยืนยันตัวตนของ Provider

-- 1. Face Verification Table
CREATE TABLE IF NOT EXISTS face_verifications (
    verification_id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    
    -- Selfie Photos
    selfie_url TEXT NOT NULL,                          -- รูป selfie ที่อัปโหลด
    liveness_video_url TEXT,                           -- วิดีโอ liveness check (optional)
    
    -- Face Matching Results
    match_confidence DECIMAL(5, 2),                    -- % ความแม่นยำ (0-100)
    is_match BOOLEAN DEFAULT false,                    -- ตรงกับบัตรประชาชนหรือไม่
    national_id_photo_url TEXT,                        -- รูปจากบัตรประชาชนที่ใช้เปรียบเทียบ
    
    -- Liveness Detection
    liveness_passed BOOLEAN DEFAULT false,             -- ผ่าน liveness check หรือไม่
    liveness_confidence DECIMAL(5, 2),                 -- % ความมั่นใจว่าเป็นคนจริง
    
    -- Verification Status
    verification_status VARCHAR(20) NOT NULL DEFAULT 'pending',
    -- 'pending': รอตรวจสอบ
    -- 'approved': อนุมัติแล้ว
    -- 'rejected': ปฏิเสธ
    -- 'needs_retry': ต้องลองใหม่
    
    -- API Provider Info (ถ้าใช้ third-party service)
    api_provider VARCHAR(50),                          -- 'aws_rekognition', 'azure_face', 'onfido', etc.
    api_response_data JSONB,                           -- เก็บ response จาก API
    
    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT NOW(),
    verified_at TIMESTAMPTZ,
    verified_by INTEGER REFERENCES users(user_id),     -- Admin ที่ตรวจสอบ
    
    -- Rejection
    rejection_reason TEXT,
    retry_count INTEGER DEFAULT 0,
    
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(user_id),
    CONSTRAINT fk_verified_by FOREIGN KEY (verified_by) REFERENCES users(user_id)
);

-- Index for performance
CREATE INDEX IF NOT EXISTS idx_face_verifications_user_id ON face_verifications(user_id);
CREATE INDEX IF NOT EXISTS idx_face_verifications_status ON face_verifications(verification_status);
CREATE INDEX IF NOT EXISTS idx_face_verifications_created_at ON face_verifications(created_at DESC);

-- 2. Add face verification requirement to provider documents
ALTER TABLE provider_documents 
ADD COLUMN IF NOT EXISTS requires_face_match BOOLEAN DEFAULT false;

-- 3. Update users table to track face verification
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS face_verified BOOLEAN DEFAULT false,
ADD COLUMN IF NOT EXISTS face_verification_id INTEGER REFERENCES face_verifications(verification_id);

-- 4. Comments
COMMENT ON TABLE face_verifications IS 'ระบบยืนยันใบหน้า Provider (Face Recognition + Liveness Detection)';
COMMENT ON COLUMN face_verifications.liveness_passed IS 'ผ่าน liveness detection (ป้องกันการใช้รูปถ่าย)';
COMMENT ON COLUMN face_verifications.match_confidence IS 'ความแม่นยำในการจับคู่ใบหน้ากับบัตรประชาชน (0-100%)';
COMMENT ON COLUMN face_verifications.api_response_data IS 'เก็บ raw response จาก face recognition API';

-- 5. Function to update user face verification status
CREATE OR REPLACE FUNCTION update_user_face_verification()
RETURNS TRIGGER AS $$
BEGIN
    -- เมื่อ face verification ถูก approve ให้อัพเดท user
    IF NEW.verification_status = 'approved' AND OLD.verification_status != 'approved' THEN
        UPDATE users 
        SET face_verified = true,
            face_verification_id = NEW.verification_id
        WHERE user_id = NEW.user_id;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 6. Trigger
DROP TRIGGER IF EXISTS trigger_update_user_face_verification ON face_verifications;
CREATE TRIGGER trigger_update_user_face_verification
    AFTER UPDATE ON face_verifications
    FOR EACH ROW
    EXECUTE FUNCTION update_user_face_verification();
