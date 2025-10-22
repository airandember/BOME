-- Migration: Add Subscription Analytics
-- Description: Adds subscription-specific analytics events and webhook tracking
-- Version: 013
-- Date: 2024-12-19

-- Add subscription-specific analytics events to existing analytics_events table
-- These events will be tracked for subscription lifecycle

-- Add subscription-specific webhook events to existing webhook_events table
-- These events will track Stripe webhook processing for subscriptions

-- Create subscription_metrics table for aggregated subscription data
CREATE TABLE IF NOT EXISTS subscription_metrics (
    id SERIAL PRIMARY KEY,
    date DATE NOT NULL,
    plan_id INTEGER REFERENCES subscription_plans(id) ON DELETE SET NULL,
    metric_type VARCHAR(50) NOT NULL CHECK (metric_type IN ('new_subscriptions', 'cancellations', 'renewals', 'revenue', 'churn_rate', 'mrr')),
    metric_value DECIMAL(15,2) NOT NULL,
    metric_count INTEGER DEFAULT 0,
    currency VARCHAR(3) DEFAULT 'USD',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(date, plan_id, metric_type)
);

-- Create indexes for subscription metrics
CREATE INDEX IF NOT EXISTS idx_subscription_metrics_date ON subscription_metrics(date);
CREATE INDEX IF NOT EXISTS idx_subscription_metrics_plan_id ON subscription_metrics(plan_id);
CREATE INDEX IF NOT EXISTS idx_subscription_metrics_type ON subscription_metrics(metric_type);
CREATE INDEX IF NOT EXISTS idx_subscription_metrics_date_type ON subscription_metrics(date, metric_type);

-- Create subscription_events table for detailed subscription lifecycle tracking
CREATE TABLE IF NOT EXISTS subscription_events (
    id SERIAL PRIMARY KEY,
    subscription_id INTEGER REFERENCES subscriptions(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    plan_id INTEGER REFERENCES subscription_plans(id) ON DELETE SET NULL,
    event_type VARCHAR(50) NOT NULL CHECK (event_type IN (
        'subscription_created', 'subscription_activated', 'subscription_cancelled', 
        'subscription_renewed', 'subscription_expired', 'payment_succeeded', 
        'payment_failed', 'refund_processed', 'plan_changed', 'promotion_applied'
    )),
    event_data JSONB, -- Additional event-specific data
    stripe_event_id VARCHAR(255), -- Reference to Stripe event if applicable
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for subscription events
CREATE INDEX IF NOT EXISTS idx_subscription_events_subscription_id ON subscription_events(subscription_id);
CREATE INDEX IF NOT EXISTS idx_subscription_events_user_id ON subscription_events(user_id);
CREATE INDEX IF NOT EXISTS idx_subscription_events_plan_id ON subscription_events(plan_id);
CREATE INDEX IF NOT EXISTS idx_subscription_events_type ON subscription_events(event_type);
CREATE INDEX IF NOT EXISTS idx_subscription_events_created_at ON subscription_events(created_at);
CREATE INDEX IF NOT EXISTS idx_subscription_events_stripe_event_id ON subscription_events(stripe_event_id) WHERE stripe_event_id IS NOT NULL;

-- Create subscription_webhooks table for tracking Stripe webhook processing
CREATE TABLE IF NOT EXISTS subscription_webhooks (
    id SERIAL PRIMARY KEY,
    stripe_event_id VARCHAR(255) UNIQUE NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    subscription_id INTEGER REFERENCES subscriptions(id) ON DELETE SET NULL,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    processing_time INTEGER, -- milliseconds
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    payload_hash VARCHAR(64), -- SHA256 hash of the webhook payload for deduplication
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    processed_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for subscription webhooks
CREATE INDEX IF NOT EXISTS idx_subscription_webhooks_stripe_event_id ON subscription_webhooks(stripe_event_id);
CREATE INDEX IF NOT EXISTS idx_subscription_webhooks_status ON subscription_webhooks(status);
CREATE INDEX IF NOT EXISTS idx_subscription_webhooks_type ON subscription_webhooks(event_type);
CREATE INDEX IF NOT EXISTS idx_subscription_webhooks_subscription_id ON subscription_webhooks(subscription_id);
CREATE INDEX IF NOT EXISTS idx_subscription_webhooks_created_at ON subscription_webhooks(created_at);
CREATE INDEX IF NOT EXISTS idx_subscription_webhooks_payload_hash ON subscription_webhooks(payload_hash);

-- Add comments for documentation
COMMENT ON TABLE subscription_metrics IS 'Aggregated subscription metrics for reporting and analytics';
COMMENT ON TABLE subscription_events IS 'Detailed subscription lifecycle events for tracking and debugging';
COMMENT ON TABLE subscription_webhooks IS 'Stripe webhook processing tracking for subscriptions';

COMMENT ON COLUMN subscription_metrics.metric_type IS 'Type of metric being tracked';
COMMENT ON COLUMN subscription_metrics.metric_value IS 'Numeric value of the metric';
COMMENT ON COLUMN subscription_metrics.metric_count IS 'Count of events for this metric';

COMMENT ON COLUMN subscription_events.event_type IS 'Type of subscription event';
COMMENT ON COLUMN subscription_events.event_data IS 'JSON data specific to this event type';
COMMENT ON COLUMN subscription_events.stripe_event_id IS 'Reference to Stripe event if this was triggered by Stripe';

COMMENT ON COLUMN subscription_webhooks.stripe_event_id IS 'Stripe event ID for deduplication';
COMMENT ON COLUMN subscription_webhooks.processing_time IS 'Time taken to process the webhook in milliseconds';
COMMENT ON COLUMN subscription_webhooks.payload_hash IS 'SHA256 hash of webhook payload for deduplication';

-- Create a function to automatically update subscription metrics
CREATE OR REPLACE FUNCTION update_subscription_metrics()
RETURNS TRIGGER AS $$
BEGIN
    -- This function will be called by triggers to update aggregated metrics
    -- Implementation will be added when we create the analytics service
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Insert some subscription-specific webhook event types into existing webhook_events table
-- These are examples of events that will be tracked
INSERT INTO webhook_events (event_type, subsite, endpoint, status, response_time, payload_size, status_code) VALUES
('subscription.created', 'streaming', 'https://api.stripe.com/v1/webhook-endpoints/subscription', 'success', 150, 1024, 200),
('subscription.updated', 'streaming', 'https://api.stripe.com/v1/webhook-endpoints/subscription', 'success', 120, 896, 200),
('subscription.deleted', 'streaming', 'https://api.stripe.com/v1/webhook-endpoints/subscription', 'success', 100, 512, 200),
('invoice.payment_succeeded', 'streaming', 'https://api.stripe.com/v1/webhook-endpoints/subscription', 'success', 180, 1536, 200),
('invoice.payment_failed', 'streaming', 'https://api.stripe.com/v1/webhook-endpoints/subscription', 'success', 160, 1280, 200)
ON CONFLICT DO NOTHING; 