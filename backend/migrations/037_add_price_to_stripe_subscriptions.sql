-- Migration: Add price information to stripe_subscriptions table
-- This allows us to link subscriptions directly to their pricing information

-- Add price_id column to stripe_subscriptions
ALTER TABLE stripe_subscriptions 
ADD COLUMN IF NOT EXISTS price_id INTEGER REFERENCES stripe_prices(id);

-- Add index for better performance
CREATE INDEX IF NOT EXISTS idx_stripe_subscriptions_price_id ON stripe_subscriptions(price_id);

-- Add comment for documentation
COMMENT ON COLUMN stripe_subscriptions.price_id IS 'Foreign key reference to stripe_prices table';
