-- Migration: Create subscription offers tables
-- Description: Creates tables for subscription offers and offer history

-- Create subscription_offers table
CREATE TABLE IF NOT EXISTS subscription_offers (
    id SERIAL PRIMARY KEY,
    plan_id INTEGER NOT NULL REFERENCES subscription_plans(id) ON DELETE CASCADE,
    item_id VARCHAR(255),
    off_discount_type VARCHAR(50) NOT NULL DEFAULT 'percentage',
    off_discount_value DECIMAL(10,2) NOT NULL,
    offer_start_date TIMESTAMP,
    off_end_date TIMESTAMP,
    is_active BOOLEAN NOT NULL DEFAULT true,
    off_description TEXT,
    off_name VARCHAR(255) NOT NULL,
    off_code VARCHAR(100),
    off_max_uses INTEGER DEFAULT 100,
    off_current_uses INTEGER DEFAULT 0,
    off_terms_conditions TEXT,
    off_target VARCHAR(100),
    off_priority INTEGER DEFAULT 1,
    off_auto_apply BOOLEAN DEFAULT false,
    off_created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    off_updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create subscription_offers_history table
CREATE TABLE IF NOT EXISTS subscription_offers_history (
    id SERIAL PRIMARY KEY,
    offer_id INTEGER NOT NULL REFERENCES subscription_offers(id) ON DELETE CASCADE,
    user_id JSONB,
    sub_plan_id INTEGER REFERENCES subscription_plans(id) ON DELETE SET NULL,
    accepted BOOLEAN DEFAULT false,
    off_user_ip VARCHAR(45),
    device_info TEXT,
    event_type VARCHAR(100) NOT NULL,
    discount_amount DECIMAL(10,2),
    original_price DECIMAL(10,2),
    final_price DECIMAL(10,2),
    session_id VARCHAR(255),
    referrer_url TEXT,
    user_agent TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_subscription_offers_plan_id ON subscription_offers(plan_id);
CREATE INDEX IF NOT EXISTS idx_subscription_offers_is_active ON subscription_offers(is_active);
CREATE INDEX IF NOT EXISTS idx_subscription_offers_priority ON subscription_offers(off_priority);
CREATE INDEX IF NOT EXISTS idx_subscription_offers_history_offer_id ON subscription_offers_history(offer_id);
CREATE INDEX IF NOT EXISTS idx_subscription_offers_history_event_type ON subscription_offers_history(event_type);
CREATE INDEX IF NOT EXISTS idx_subscription_offers_history_created_at ON subscription_offers_history(created_at);

-- Add comments for documentation
COMMENT ON TABLE subscription_offers IS 'Stores subscription offers that can be applied to subscription plans';
COMMENT ON TABLE subscription_offers_history IS 'Stores history of offer interactions and events';
COMMENT ON COLUMN subscription_offers.off_discount_type IS 'Type of discount: percentage, fixed_amount, etc.';
COMMENT ON COLUMN subscription_offers.off_discount_value IS 'Value of the discount (percentage or fixed amount)';
COMMENT ON COLUMN subscription_offers.off_priority IS 'Priority level for offer application (higher = more priority)';
COMMENT ON COLUMN subscription_offers.off_auto_apply IS 'Whether the offer should be automatically applied';
COMMENT ON COLUMN subscription_offers_history.event_type IS 'Type of event: offer_created, offer_viewed, offer_accepted, etc.'; 