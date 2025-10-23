-- Migration: Create Subscription Plans Table
-- Description: Creates table for subscription plans that users can subscribe to
-- Version: 010
-- Date: 2024-12-19

-- Create subscription_plans table
CREATE TABLE IF NOT EXISTS subscription_plans (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL CHECK (price >= 0),
    currency VARCHAR(3) DEFAULT 'USD' CHECK (currency IN ('USD', 'EUR', 'GBP', 'CAD')),
    interval VARCHAR(20) NOT NULL CHECK (interval IN ('monthly', 'annual', 'weekly', 'daily')),
    interval_count INTEGER DEFAULT 1 CHECK (interval_count > 0),
    stripe_price_id VARCHAR(255) UNIQUE,
    features JSONB, -- Array of features included in this plan
    is_active BOOLEAN DEFAULT true,
    is_promoted BOOLEAN DEFAULT false,
    promotion_end_date TIMESTAMP WITH TIME ZONE,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE NULL
);

-- Create indexes for performance optimization
CREATE INDEX IF NOT EXISTS idx_subscription_plans_active ON subscription_plans(is_active) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_subscription_plans_promoted ON subscription_plans(is_promoted) WHERE is_active = true AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_subscription_plans_sort_order ON subscription_plans(sort_order) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_subscription_plans_stripe_price_id ON subscription_plans(stripe_price_id) WHERE stripe_price_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_subscription_plans_created_at ON subscription_plans(created_at);
CREATE INDEX IF NOT EXISTS idx_subscription_plans_deleted_at ON subscription_plans(deleted_at) WHERE deleted_at IS NOT NULL;

-- Create unique constraint for active plans with same name
CREATE UNIQUE INDEX IF NOT EXISTS idx_subscription_plans_name_unique ON subscription_plans(name) WHERE deleted_at IS NULL;

-- Add comments for documentation
COMMENT ON TABLE subscription_plans IS 'Subscription plans that users can subscribe to';
COMMENT ON COLUMN subscription_plans.id IS 'Primary key for subscription plan';
COMMENT ON COLUMN subscription_plans.name IS 'Display name of the subscription plan';
COMMENT ON COLUMN subscription_plans.description IS 'Detailed description of what the plan includes';
COMMENT ON COLUMN subscription_plans.price IS 'Price of the subscription plan';
COMMENT ON COLUMN subscription_plans.currency IS 'Currency code for the price (USD, EUR, GBP, CAD)';
COMMENT ON COLUMN subscription_plans.interval IS 'Billing interval (monthly, annual, weekly, daily)';
COMMENT ON COLUMN subscription_plans.interval_count IS 'Number of intervals between billings';
COMMENT ON COLUMN subscription_plans.stripe_price_id IS 'Stripe price ID for this plan';
COMMENT ON COLUMN subscription_plans.features IS 'JSON array of features included in this plan';
COMMENT ON COLUMN subscription_plans.is_active IS 'Whether this plan is available for subscription';
COMMENT ON COLUMN subscription_plans.is_promoted IS 'Whether this plan is currently being promoted';
COMMENT ON COLUMN subscription_plans.promotion_end_date IS 'When the promotion ends (NULL for no end date)';
COMMENT ON COLUMN subscription_plans.sort_order IS 'Order for displaying plans (lower numbers first)';
COMMENT ON COLUMN subscription_plans.created_at IS 'When the plan was created';
COMMENT ON COLUMN subscription_plans.updated_at IS 'When the plan was last updated';
COMMENT ON COLUMN subscription_plans.deleted_at IS 'Soft delete timestamp (NULL if not deleted)';

-- Insert some default subscription plans
INSERT INTO subscription_plans (name, description, price, currency, interval, interval_count, features, is_active, sort_order) VALUES
('Basic Monthly', 'Access to basic streaming content', 9.99, 'USD', 'monthly', 1, '["HD Streaming", "Basic Support", "Ad-free Experience"]', true, 1),
('Premium Monthly', 'Full access to all streaming content', 19.99, 'USD', 'monthly', 1, '["4K Streaming", "Premium Support", "Ad-free Experience", "Offline Downloads", "Multiple Devices"]', true, 2),
('Basic Annual', 'Access to basic streaming content (annual)', 99.99, 'USD', 'annual', 1, '["HD Streaming", "Basic Support", "Ad-free Experience", "2 Months Free"]', true, 3),
('Premium Annual', 'Full access to all streaming content (annual)', 199.99, 'USD', 'annual', 1, '["4K Streaming", "Premium Support", "Ad-free Experience", "Offline Downloads", "Multiple Devices", "3 Months Free"]', true, 4)
ON CONFLICT (name) DO NOTHING; 