-- Migration: Add password_changed field to users table
-- This migration adds a field to track if users have changed their password from the default

-- Add password_changed field to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_changed BOOLEAN DEFAULT FALSE;

-- Update existing users to have password_changed = true (they've been using the system)
UPDATE users SET password_changed = TRUE WHERE password_changed IS NULL;

-- Add index for password_changed field for efficient queries
CREATE INDEX IF NOT EXISTS idx_users_password_changed ON users(password_changed);

-- Add comment to document the field
COMMENT ON COLUMN users.password_changed IS 'Indicates if user has changed their password from the default/temporary password';
