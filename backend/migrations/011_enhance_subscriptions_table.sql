-- Migration: Enhance Subscriptions Table
-- Description: Adds new fields to existing subscriptions table for enhanced functionality
-- Version: 011
-- Date: 2024-12-19

-- Add new columns to existing subscriptions table
ALTER TABLE subscriptions 
ADD COLUMN IF NOT EXISTS plan_id INTEGER REFERENCES subscription_plans(id) ON DELETE SET NULL,
ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE NULL,
ADD COLUMN IF NOT EXISTS cancellation_reason TEXT,
ADD COLUMN IF NOT EXISTS refund_amount DECIMAL(10,2) NULL CHECK (refund_amount >= 0),
ADD COLUMN IF NOT EXISTS refund_reason TEXT;

-- Create indexes for the new columns
CREATE INDEX IF NOT EXISTS idx_subscriptions_plan_id ON subscriptions(plan_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_subscriptions_deleted_at ON subscriptions(deleted_at) WHERE deleted_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_subscriptions_cancellation_reason ON subscriptions(cancellation_reason) WHERE cancellation_reason IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_subscriptions_refund_amount ON subscriptions(refund_amount) WHERE refund_amount IS NOT NULL;

-- Update existing indexes to include deleted_at filter where appropriate
DROP INDEX IF EXISTS idx_subscriptions_user_id;
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id) WHERE deleted_at IS NULL;

DROP INDEX IF EXISTS idx_subscriptions_status;
CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions(status) WHERE deleted_at IS NULL;

DROP INDEX IF EXISTS idx_subscriptions_stripe_subscription_id;
CREATE INDEX IF NOT EXISTS idx_subscriptions_stripe_subscription_id ON subscriptions(stripe_subscription_id) WHERE deleted_at IS NULL;

-- Add comments for the new columns
COMMENT ON COLUMN subscriptions.plan_id IS 'Reference to the subscription plan this subscription is for';
COMMENT ON COLUMN subscriptions.deleted_at IS 'Soft delete timestamp (NULL if not deleted)';
COMMENT ON COLUMN subscriptions.cancellation_reason IS 'Reason provided for subscription cancellation';
COMMENT ON COLUMN subscriptions.refund_amount IS 'Amount refunded for this subscription';
COMMENT ON COLUMN subscriptions.refund_reason IS 'Reason for the refund';

-- Update existing subscriptions to have a default plan_id if they don't have one
-- This assumes there's a default plan with ID 1 (Basic Monthly)
UPDATE subscriptions 
SET plan_id = 1 
WHERE plan_id IS NULL 
AND EXISTS (SELECT 1 FROM subscription_plans WHERE id = 1);

-- Add constraint to ensure refund_amount is only set when there's a refund_reason
ALTER TABLE subscriptions 
ADD CONSTRAINT IF NOT EXISTS check_refund_reason 
CHECK ((refund_amount IS NULL AND refund_reason IS NULL) OR (refund_amount IS NOT NULL AND refund_reason IS NOT NULL)); 