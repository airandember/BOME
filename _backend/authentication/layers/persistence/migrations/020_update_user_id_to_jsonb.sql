-- Migration: Update user_id to JSONB in subscription_plans_history
-- Description: Changes user_id column from VARCHAR(100) to JSONB to store complete user information
-- This allows storing full user details like {"id":4,"email":"super.admin@bome.test","role":"super_admin","first_name":"Super","last_name":"Administrator"}
-- Version: 020
-- Date: 2025-07-30

-- First, create a temporary column with JSONB type
ALTER TABLE subscription_plans_history 
ADD COLUMN user_data JSONB;

-- Update the temporary column with existing user_id data as JSON
UPDATE subscription_plans_history 
SET user_data = CASE 
    WHEN user_id IS NULL OR user_id = '' THEN NULL
    WHEN user_id = 'system' THEN '{"id": "system", "email": "system", "role": "system", "first_name": "System", "last_name": ""}'::jsonb
    WHEN user_id = 'System (Auto-Expiration)' THEN '{"id": "system", "email": "system", "role": "system", "first_name": "System", "last_name": "(Auto-Expiration)"}'::jsonb
    ELSE jsonb_build_object('id', user_id, 'email', user_id, 'role', 'user', 'first_name', user_id, 'last_name', '')
END;

-- Drop the old user_id column
ALTER TABLE subscription_plans_history 
DROP COLUMN user_id;

-- Rename the new column to user_id
ALTER TABLE subscription_plans_history 
RENAME COLUMN user_data TO user_id;

-- Update the index to work with JSONB
DROP INDEX IF EXISTS idx_subscription_plans_history_user_id;
CREATE INDEX idx_subscription_plans_history_user_id ON subscription_plans_history USING GIN (user_id);

-- Update the comment
COMMENT ON COLUMN subscription_plans_history.user_id IS 'Complete user information as JSONB (id, email, role, first_name, last_name, etc.)';

-- Add a constraint to ensure user_id is valid JSONB when not null
ALTER TABLE subscription_plans_history 
ADD CONSTRAINT chk_user_id_jsonb 
CHECK (user_id IS NULL OR jsonb_typeof(user_id) = 'object'); 