-- Migration 008: Add Notifications System
-- Create notifications table for real-time alerts

-- Table: notifications
-- Stores notifications for users (bookings, messages, KYC, etc.)
CREATE TABLE IF NOT EXISTS notifications (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL, -- 'new_message', 'booking_request', 'booking_confirmed', 'booking_cancelled', 'kyc_approved', 'kyc_rejected', 'new_review'
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    metadata JSONB, -- Additional data (e.g., {"conversation_id": 123, "booking_id": 456})
    is_read BOOLEAN DEFAULT FALSE,
    read_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT check_type CHECK (type IN (
        'new_message', 
        'booking_request', 
        'booking_confirmed', 
        'booking_cancelled', 
        'booking_completed',
        'kyc_approved', 
        'kyc_rejected', 
        'new_review',
        'payment_success',
        'payment_failed',
        'tier_upgraded'
    ))
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_type ON notifications(type);
CREATE INDEX IF NOT EXISTS idx_notifications_is_read ON notifications(is_read);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_unread_user ON notifications(user_id, is_read) WHERE is_read = FALSE;

-- Comments for documentation
COMMENT ON TABLE notifications IS 'Stores user notifications for various events';
COMMENT ON COLUMN notifications.type IS 'Type of notification (message, booking, KYC, review, payment, etc.)';
COMMENT ON COLUMN notifications.metadata IS 'JSON field for additional context (IDs, URLs, etc.)';
COMMENT ON COLUMN notifications.is_read IS 'Whether the user has read this notification';
