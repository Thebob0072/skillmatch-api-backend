-- Migration 011: Blocks System
-- Created: November 13, 2025
-- Purpose: Allow users to block other users

-- Blocks Table
CREATE TABLE IF NOT EXISTS blocks (
    id SERIAL PRIMARY KEY,
    blocker_id INTEGER NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    blocked_id INTEGER NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_not_self_block CHECK (blocker_id != blocked_id),
    UNIQUE(blocker_id, blocked_id) -- One block per pair
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_blocks_blocker ON blocks(blocker_id);
CREATE INDEX IF NOT EXISTS idx_blocks_blocked ON blocks(blocked_id);
CREATE INDEX IF NOT EXISTS idx_blocks_pair ON blocks(blocker_id, blocked_id);

-- Comments
COMMENT ON TABLE blocks IS 'User blocking system for privacy and safety';
COMMENT ON COLUMN blocks.blocker_id IS 'User who initiated the block';
COMMENT ON COLUMN blocks.blocked_id IS 'User who is being blocked';
COMMENT ON COLUMN blocks.reason IS 'Optional reason for blocking';
