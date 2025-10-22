-- Migration: Add Missing Subscription Plan Columns
-- Description: Adds missing columns that the Go code expects but are not in the database schema
-- Version: 016
-- Date: 2024-12-19

-- Add sub_type column for plan classification (100 = standard, 300 = promotional)
ALTER TABLE subscription_plans 
ADD COLUMN IF NOT EXISTS sub_type INTEGER DEFAULT 100 CHECK (sub_type IN (100, 300));

-- Add promotion_start_date column
ALTER TABLE subscription_plans 
ADD COLUMN IF NOT EXISTS promotion_start_date TIMESTAMP WITH TIME ZONE;

-- Add promotion_history column for tracking promotion changes
ALTER TABLE subscription_plans 
ADD COLUMN IF NOT EXISTS promotion_history JSONB DEFAULT '[]'::jsonb;

-- Add is_deleted column for soft delete status
ALTER TABLE subscription_plans 
ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN DEFAULT false;

-- Update existing plans to have proper sub_type values
UPDATE subscription_plans 
SET sub_type = CASE 
    WHEN is_promoted = true THEN 300 
    ELSE 100 
END
WHERE sub_type IS NULL;

-- Add comments for the new columns
COMMENT ON COLUMN subscription_plans.sub_type IS 'Plan type: 100 = standard plan, 300 = promotional plan';
COMMENT ON COLUMN subscription_plans.promotion_start_date IS 'When the promotion started (NULL if not promoted)';
COMMENT ON COLUMN subscription_plans.promotion_history IS 'JSON array of promotion history events';
COMMENT ON COLUMN subscription_plans.is_deleted IS 'Soft delete flag (true if plan is deleted)';

-- Create index on sub_type for better performance
CREATE INDEX IF NOT EXISTS idx_subscription_plans_sub_type ON subscription_plans(sub_type) WHERE deleted_at IS NULL; 