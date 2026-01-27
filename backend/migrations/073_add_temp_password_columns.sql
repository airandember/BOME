-- Migration: Add temp password columns for simplified signup flow
-- This allows users with existing Stripe subscriptions to receive a temporary password
-- instead of going through email verification

-- Add temp password tracking columns
ALTER TABLE users ADD COLUMN IF NOT EXISTS temp_password_active BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS temp_password TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS temp_password_created_at TIMESTAMP;

-- Add comments for documentation
COMMENT ON COLUMN users.temp_password_active IS 'True if user is using a temporary password (BOME_[user_id])';
COMMENT ON COLUMN users.temp_password IS 'Plaintext temporary password for simplified login';
COMMENT ON COLUMN users.temp_password_created_at IS 'When the temp password was issued';

-- Create index for temp password lookups
CREATE INDEX IF NOT EXISTS idx_users_temp_password_active ON users(temp_password_active) WHERE temp_password_active = TRUE;
