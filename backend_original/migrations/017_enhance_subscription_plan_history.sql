-- Migration: Enhance Subscription Plan History Tracking
-- Description: Adds enhanced history tracking columns and updates existing structure
-- Version: 017
-- Date: 2025-07-28

-- First, rename the existing promotion_history to plan_change_history for general tracking
ALTER TABLE subscription_plans 
RENAME COLUMN promotion_history TO plan_change_history;

-- Add new promotion_metadata column for promotion-specific analytics
ALTER TABLE subscription_plans 
ADD COLUMN IF NOT EXISTS promotion_metadata JSONB DEFAULT '{}'::jsonb;

-- Update sub_type to use string values instead of integers
ALTER TABLE subscription_plans 
ALTER COLUMN sub_type TYPE VARCHAR(8);

-- Update existing sub_type values to use string format
UPDATE subscription_plans 
SET sub_type = CASE 
    WHEN sub_type = '100' OR sub_type = '100'::integer THEN 'stnd'
    WHEN sub_type = '300' OR sub_type = '300'::integer THEN 'prmo'
    ELSE 'stnd'
END;

-- Add constraint for sub_type string values
ALTER TABLE subscription_plans 
ADD CONSTRAINT check_sub_type CHECK (sub_type IN ('stnd', 'prmo'));

-- Add comments for the enhanced columns
COMMENT ON COLUMN subscription_plans.plan_change_history IS 'Complete audit trail of all plan changes (both stnd and prmo)';
COMMENT ON COLUMN subscription_plans.promotion_metadata IS 'Promotion-specific analytics and metadata (only for prmo plans)';
COMMENT ON COLUMN subscription_plans.sub_type IS 'Plan type: stnd = standard plan, prmo = promotional plan';

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_plan_change_history ON subscription_plans USING GIN (plan_change_history);
CREATE INDEX IF NOT EXISTS idx_promotion_metadata ON subscription_plans USING GIN (promotion_metadata);

-- Update existing index to work with string sub_type
DROP INDEX IF EXISTS idx_subscription_plans_sub_type;
CREATE INDEX IF NOT EXISTS idx_subscription_plans_sub_type ON subscription_plans(sub_type) WHERE deleted_at IS NULL;

-- Initialize plan_change_history for existing plans with creation events
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

-- Initialize promotion_metadata for existing promotional plans
UPDATE subscription_plans 
SET promotion_metadata = jsonb_build_object(
    'promotion_stats', jsonb_build_object(
        'total_promotions', 1,
        'current_promotion', CASE 
            WHEN promotion_start_date IS NOT NULL AND promotion_end_date IS NOT NULL THEN
                jsonb_build_object(
                    'start_date', promotion_start_date,
                    'end_date', promotion_end_date,
                    'duration_days', extract(day from (promotion_end_date - promotion_start_date)),
                    'status', CASE 
                        WHEN promotion_end_date < NOW() THEN 'expired'
                        WHEN promotion_start_date > NOW() THEN 'upcoming'
                        ELSE 'active'
                    END
                )
            ELSE null
        END,
        'performance_metrics', jsonb_build_object(
            'total_revenue_generated', 0,
            'average_conversion_rate', 0,
            'best_performing_duration', 0
        )
    ),
    'promotion_settings', jsonb_build_object(
        'auto_extend', false,
        'notify_before_expiry', true,
        'expiry_notification_days', 3
    )
)
WHERE sub_type = 'prmo' AND (promotion_metadata IS NULL OR promotion_metadata = '{}'::jsonb); 