-- Migration: Add stripe_product_id to subscription_plans table
-- This links subscription plans to their corresponding Stripe products

-- Add stripe_product_id column
ALTER TABLE subscription_plans 
ADD COLUMN IF NOT EXISTS stripe_product_id VARCHAR(255);

-- Add foreign key constraint to stripe_products table
ALTER TABLE subscription_plans 
ADD CONSTRAINT fk_subscription_plans_stripe_product 
FOREIGN KEY (stripe_product_id) 
REFERENCES stripe_products(stripe_id) 
ON DELETE SET NULL;

-- Add index for better performance
CREATE INDEX IF NOT EXISTS idx_subscription_plans_stripe_product_id 
ON subscription_plans (stripe_product_id);

-- Add comment for documentation
COMMENT ON COLUMN subscription_plans.stripe_product_id IS 'Links to stripe_products.stripe_id for Stripe integration';
