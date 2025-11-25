-- Migration 010: Profile Views Tracking
-- Created: November 13, 2025
-- Purpose: Track profile views for analytics

-- Profile Views Table
CREATE TABLE IF NOT EXISTS profile_views (
    id SERIAL PRIMARY KEY,
    provider_id INTEGER NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    viewer_id INTEGER REFERENCES users(user_id) ON DELETE SET NULL, -- NULL for anonymous views
    view_count INTEGER DEFAULT 1,
    last_viewed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(provider_id, COALESCE(viewer_id, -1)) -- Allow one record per provider-viewer pair
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_profile_views_provider ON profile_views(provider_id);
CREATE INDEX IF NOT EXISTS idx_profile_views_viewer ON profile_views(viewer_id) WHERE viewer_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_profile_views_date ON profile_views(last_viewed_at);

-- Comments
COMMENT ON TABLE profile_views IS 'Tracks profile views for analytics';
COMMENT ON COLUMN profile_views.provider_id IS 'The provider being viewed';
COMMENT ON COLUMN profile_views.viewer_id IS 'The user viewing the profile (NULL for anonymous)';
COMMENT ON COLUMN profile_views.view_count IS 'Total views from this viewer';
COMMENT ON COLUMN profile_views.last_viewed_at IS 'Last time this viewer viewed the profile';
