-- Migration: Add Video Access Tracking Columns
-- Description: Adds timestamp and source tracking for video access grants

-- Add video_access_granted_at column (if it doesn't exist)
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS video_access_granted_at TIMESTAMPTZ;

-- Add video_access_source column (if it doesn't exist)
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS video_access_source VARCHAR(255);

-- Add comments
COMMENT ON COLUMN users.video_access_granted_at IS 'Timestamp when video access was granted';
COMMENT ON COLUMN users.video_access_source IS 'Source of video access grant (e.g., session_verification:cs_xxx, webhook:sub_xxx, retroactive_linking)';
