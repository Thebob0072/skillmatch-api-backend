-- Migration 014: Add Email Verification System
-- Created: 2025-11-14
-- Purpose: Store OTP codes for email verification during registration

-- Create email_verifications table
CREATE TABLE IF NOT EXISTS email_verifications (
    email VARCHAR(255) PRIMARY KEY,
    otp CHAR(6) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index for faster lookups
CREATE INDEX IF NOT EXISTS idx_email_verifications_expires 
ON email_verifications(expires_at);

-- Add phone_number to users table if not exists
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS phone_number VARCHAR(20);

COMMENT ON TABLE email_verifications IS 'Temporary storage for email verification OTP codes (10 minute expiry)';
COMMENT ON COLUMN email_verifications.otp IS '6-digit verification code';
COMMENT ON COLUMN email_verifications.expires_at IS 'OTP expiration time (10 minutes from creation)';
