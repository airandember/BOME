-- Migration: Add metadata columns to subscription_offers_history
-- Description: Add metadata, old_values, and new_values columns for better history tracking

-- Add the missing columns to subscription_offers_history table
ALTER TABLE subscription_offers_history 
ADD COLUMN IF NOT EXISTS metadata JSONB,
ADD COLUMN IF NOT EXISTS old_values JSONB,
ADD COLUMN IF NOT EXISTS new_values JSONB,
ADD COLUMN IF NOT EXISTS description TEXT;

-- Add comments for documentation
COMMENT ON COLUMN subscription_offers_history.metadata IS 'Additional metadata for the history event';
COMMENT ON COLUMN subscription_offers_history.old_values IS 'Previous values before the change';
COMMENT ON COLUMN subscription_offers_history.new_values IS 'New values after the change';
COMMENT ON COLUMN subscription_offers_history.description IS 'Human-readable description of the event'; 