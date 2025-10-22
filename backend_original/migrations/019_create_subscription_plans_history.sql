-- Migration: Create Subscription Plans History Table
-- Description: Creates a separate table for tracking subscription plan changes
-- This provides better querying capabilities and analytics support
-- Version: 019
-- Date: 2025-07-29

-- Create the history table
CREATE TABLE subscription_plans_history (
    id SERIAL PRIMARY KEY,
    plan_id INTEGER NOT NULL REFERENCES subscription_plans(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    user_id VARCHAR(100),
    description TEXT,
    old_values JSONB,
    new_values JSONB,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX idx_subscription_plans_history_plan_id ON subscription_plans_history(plan_id);
CREATE INDEX idx_subscription_plans_history_timestamp ON subscription_plans_history(timestamp);
CREATE INDEX idx_subscription_plans_history_event_type ON subscription_plans_history(event_type);
CREATE INDEX idx_subscription_plans_history_user_id ON subscription_plans_history(user_id);
CREATE INDEX idx_subscription_plans_history_plan_timestamp ON subscription_plans_history(plan_id, timestamp DESC);

-- Add table and column comments for documentation
COMMENT ON TABLE subscription_plans_history IS 'Audit trail for subscription plan changes and events';
COMMENT ON COLUMN subscription_plans_history.id IS 'Unique identifier for history record';
COMMENT ON COLUMN subscription_plans_history.plan_id IS 'Foreign key reference to subscription_plans table';
COMMENT ON COLUMN subscription_plans_history.event_type IS 'Type of change event (plan_created, status_toggled, etc.)';
COMMENT ON COLUMN subscription_plans_history.timestamp IS 'When the event occurred';
COMMENT ON COLUMN subscription_plans_history.user_id IS 'User who performed the action';
COMMENT ON COLUMN subscription_plans_history.description IS 'Human-readable description of the change';
COMMENT ON COLUMN subscription_plans_history.old_values IS 'Previous field values before the change';
COMMENT ON COLUMN subscription_plans_history.new_values IS 'New field values after the change';
COMMENT ON COLUMN subscription_plans_history.metadata IS 'Additional context and analytics data';
COMMENT ON COLUMN subscription_plans_history.created_at IS 'When this history record was created';

-- Add constraint to ensure valid event types
ALTER TABLE subscription_plans_history 
ADD CONSTRAINT chk_event_type 
CHECK (event_type IN (
    'plan_created',
    'plan_updated',
    'status_activated',
    'status_deactivated',
    'promotion_started',
    'promotion_ended',
    'promotion_expired',
    'price_changed',
    'type_changed',
    'plan_deleted'
));

-- Optional: Migrate existing history data from JSONB column
-- Uncomment the following section if you want to migrate existing data
/*
INSERT INTO subscription_plans_history (plan_id, event_type, timestamp, user_id, description, old_values, new_values, metadata)
SELECT 
    id as plan_id,
    (event->>'event_type')::VARCHAR(50) as event_type,
    (event->>'timestamp')::TIMESTAMP WITH TIME ZONE as timestamp,
    event->>'user_id' as user_id,
    event->>'description' as description,
    event->'old_values' as old_values,
    event->'new_values' as new_values,
    event->'metadata' as metadata
FROM subscription_plans,
     jsonb_array_elements(plan_change_history) as event
WHERE plan_change_history IS NOT NULL 
  AND plan_change_history != '[]'::jsonb;
*/
