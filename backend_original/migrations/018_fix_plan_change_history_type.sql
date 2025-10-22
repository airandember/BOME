-- Migration: Fix Plan Change History Column Type
-- Description: Changes plan_change_history from jsonb[] to jsonb for better flexibility
-- Version: 018
-- Date: 2025-07-29

-- First, let's backup the existing data by converting jsonb[] to jsonb
-- We'll create a temporary column to store the converted data
ALTER TABLE subscription_plans 
ADD COLUMN IF NOT EXISTS plan_change_history_temp JSONB;

-- Convert existing jsonb[] data to jsonb format
UPDATE subscription_plans 
SET plan_change_history_temp = CASE 
    WHEN plan_change_history IS NULL OR plan_change_history = '{}'::jsonb[] THEN '[]'::jsonb
    ELSE plan_change_history::jsonb
END;

-- Drop the old column and rename the temp column
ALTER TABLE subscription_plans DROP COLUMN IF EXISTS plan_change_history;
ALTER TABLE subscription_plans RENAME COLUMN plan_change_history_temp TO plan_change_history;

-- Add comment for the corrected column
COMMENT ON COLUMN subscription_plans.plan_change_history IS 'Complete audit trail of all plan changes as JSONB array';

-- Recreate the index for the corrected column type
DROP INDEX IF EXISTS idx_plan_change_history;
CREATE INDEX IF NOT EXISTS idx_plan_change_history ON subscription_plans USING GIN (plan_change_history);

-- Initialize plan_change_history for existing plans with creation events if empty
UPDATE subscription_plans 
SET plan_change_history = jsonb_build_array(
    jsonb_build_object(
        'id', 'evt_' || extract(epoch from created_at)::bigint,
        'event_type', 'plan_created',
        'timestamp', created_at,
        'description', 'Plan ''' || name || ''' was created',
        'old_values', null,
        'new_values', jsonb_build_object(
            'name', name,
            'price', price,
            'sub_type', COALESCE(sub_type, 'stnd'),
            'is_active', is_active
        ),
        'metadata', jsonb_build_object(
            'plan_name', name,
            'action', 'creation'
        )
    )
)
WHERE plan_change_history IS NULL OR plan_change_history = '[]'::jsonb; 