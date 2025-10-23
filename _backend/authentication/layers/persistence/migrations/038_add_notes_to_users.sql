-- Migration: Add notes column to users table
-- This allows admins to add internal notes about subscribers

ALTER TABLE users ADD COLUMN IF NOT EXISTS notes TEXT DEFAULT '';

-- Add index for better performance when searching notes
CREATE INDEX IF NOT EXISTS idx_users_notes ON users USING gin(to_tsvector('english', notes));

-- Add comment for documentation
COMMENT ON COLUMN users.notes IS 'Internal admin notes about the user/subscriber';
