-- Migration: Create Subscriber History Table
-- Description: Creates a table for tracking subscriber history including subscriptions, offers, updates, and notes
-- Version: 026
-- Date: 2025-01-15

-- Create the subscriber_history table
CREATE TABLE IF NOT EXISTS subscriber_history (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    usr_sub_hstry JSONB DEFAULT '{}'::jsonb,
    usr_off_hstry JSONB DEFAULT '{}'::jsonb,
    updated_at JSONB DEFAULT '{}'::jsonb,
    notes JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_subscriber_history_user_id ON subscriber_history(user_id);
CREATE INDEX IF NOT EXISTS idx_subscriber_history_created_at ON subscriber_history(created_at);
CREATE INDEX IF NOT EXISTS idx_subscriber_history_usr_sub_hstry ON subscriber_history USING GIN (usr_sub_hstry);
CREATE INDEX IF NOT EXISTS idx_subscriber_history_usr_off_hstry ON subscriber_history USING GIN (usr_off_hstry);
CREATE INDEX IF NOT EXISTS idx_subscriber_history_updated_at ON subscriber_history USING GIN (updated_at);
CREATE INDEX IF NOT EXISTS idx_subscriber_history_notes ON subscriber_history USING GIN (notes);

-- Add table and column comments for documentation
COMMENT ON TABLE subscriber_history IS 'Comprehensive history tracking for subscribers including subscriptions, offers, updates, and notes';
COMMENT ON COLUMN subscriber_history.id IS 'Unique identifier for history record';
COMMENT ON COLUMN subscriber_history.user_id IS 'Foreign key reference to users table';
COMMENT ON COLUMN subscriber_history.usr_sub_hstry IS 'JSONB field storing subscription history entries';
COMMENT ON COLUMN subscriber_history.usr_off_hstry IS 'JSONB field storing offer history entries';
COMMENT ON COLUMN subscriber_history.updated_at IS 'JSONB field storing account update entries';
COMMENT ON COLUMN subscriber_history.notes IS 'JSONB field storing admin, system, and user notes';
COMMENT ON COLUMN subscriber_history.created_at IS 'When this history record was created';

-- Add constraint to ensure user_id is not null
ALTER TABLE subscriber_history 
ADD CONSTRAINT chk_user_id_not_null 
CHECK (user_id IS NOT NULL);

-- Add constraint to ensure JSONB fields are valid JSON when not empty
ALTER TABLE subscriber_history 
ADD CONSTRAINT chk_usr_sub_hstry_jsonb 
CHECK (usr_sub_hstry IS NULL OR jsonb_typeof(usr_sub_hstry) = 'object');

ALTER TABLE subscriber_history 
ADD CONSTRAINT chk_usr_off_hstry_jsonb 
CHECK (usr_off_hstry IS NULL OR jsonb_typeof(usr_off_hstry) = 'object');

ALTER TABLE subscriber_history 
ADD CONSTRAINT chk_updated_at_jsonb 
CHECK (updated_at IS NULL OR jsonb_typeof(updated_at) = 'object');

ALTER TABLE subscriber_history 
ADD CONSTRAINT chk_notes_jsonb 
CHECK (notes IS NULL OR jsonb_typeof(notes) = 'object'); 