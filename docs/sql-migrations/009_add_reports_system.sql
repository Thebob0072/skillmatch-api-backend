-- Migration 009: Add Reports System
-- Create reports table for user reporting and content moderation

-- Table: reports
-- Stores user reports for abuse, inappropriate content, etc.
CREATE TABLE IF NOT EXISTS reports (
    id SERIAL PRIMARY KEY,
    reporter_id INTEGER NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    reported_user_id INTEGER REFERENCES users(user_id) ON DELETE CASCADE,
    report_type VARCHAR(50) NOT NULL,
    reason TEXT NOT NULL,
    additional_details TEXT,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'reviewing', 'resolved', 'dismissed')),
    admin_notes TEXT,
    resolved_by INTEGER REFERENCES users(user_id) ON DELETE SET NULL,
    resolved_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT check_report_type CHECK (report_type IN (
        'harassment',
        'inappropriate_content',
        'fake_profile',
        'scam',
        'violence_threat',
        'underage',
        'spam',
        'other'
    )),
    CONSTRAINT check_not_self_report CHECK (reporter_id != reported_user_id)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_reports_reporter ON reports(reporter_id);
CREATE INDEX IF NOT EXISTS idx_reports_reported_user ON reports(reported_user_id);
CREATE INDEX IF NOT EXISTS idx_reports_status ON reports(status);
CREATE INDEX IF NOT EXISTS idx_reports_type ON reports(report_type);
CREATE INDEX IF NOT EXISTS idx_reports_created_at ON reports(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_reports_pending ON reports(status) WHERE status = 'pending';

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_report_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to automatically update updated_at
DROP TRIGGER IF EXISTS trigger_update_report_timestamp ON reports;
CREATE TRIGGER trigger_update_report_timestamp
    BEFORE UPDATE ON reports
    FOR EACH ROW
    EXECUTE FUNCTION update_report_timestamp();

-- Comments for documentation
COMMENT ON TABLE reports IS 'Stores user reports for content moderation';
COMMENT ON COLUMN reports.report_type IS 'Type of report: harassment, inappropriate_content, fake_profile, scam, violence_threat, underage, spam, other';
COMMENT ON COLUMN reports.status IS 'Report status: pending, reviewing, resolved, dismissed';
COMMENT ON COLUMN reports.admin_notes IS 'Internal notes for admin review';
COMMENT ON COLUMN reports.resolved_by IS 'Admin user ID who resolved the report';
