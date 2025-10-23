-- Migration: Fix Subscription Plans Schema
-- Description: Fixes interval values and adds missing short_desc field
-- Version: 015
-- Date: 2024-12-19

-- Add short_desc column if it doesn't exist
ALTER TABLE subscription_plans 
ADD COLUMN IF NOT EXISTS short_desc VARCHAR(500);

-- Update interval constraint to match actual database values
ALTER TABLE subscription_plans 
DROP CONSTRAINT IF EXISTS subscription_plans_interval_check;

ALTER TABLE subscription_plans 
ADD CONSTRAINT subscription_plans_interval_check 
CHECK (interval IN ('month', 'year', 'week', 'day'));

-- Update existing data to use correct interval values
UPDATE subscription_plans 
SET interval = 'month' 
WHERE interval = 'monthly';

UPDATE subscription_plans 
SET interval = 'year' 
WHERE interval = 'annual';

UPDATE subscription_plans 
SET interval = 'week' 
WHERE interval = 'weekly';

UPDATE subscription_plans 
SET interval = 'day' 
WHERE interval = 'daily';

-- Add default short_desc values for existing plans
UPDATE subscription_plans 
SET short_desc = 'Essential Monthly Access'
WHERE name = 'Essential Monthly' AND short_desc IS NULL;

UPDATE subscription_plans 
SET short_desc = 'Best Value - Save 33%'
WHERE name = 'Premium Annual' AND short_desc IS NULL;

UPDATE subscription_plans 
SET short_desc = 'Complete Library Access'
WHERE name = 'Premium Monthly' AND short_desc IS NULL;

UPDATE subscription_plans 
SET short_desc = 'Maximum Savings - Pro Benefits'
WHERE name = 'Annual Pro' AND short_desc IS NULL;

UPDATE subscription_plans 
SET short_desc = 'Complete Conference Access'
WHERE name = 'Conference + Library Bundle' AND short_desc IS NULL;

UPDATE subscription_plans 
SET short_desc = 'Premium + Conference Benefits'
WHERE name = 'Semi-Annual Premium' AND short_desc IS NULL;

UPDATE subscription_plans 
SET short_desc = 'Ultimate Annual Package'
WHERE name = 'Annual Premium' AND short_desc IS NULL;

UPDATE subscription_plans 
SET short_desc = 'Get Started Today'
WHERE name = 'Starter Monthly' AND short_desc IS NULL;

UPDATE subscription_plans 
SET short_desc = 'Professional Choice'
WHERE name = 'Professional Annual' AND short_desc IS NULL;

-- Set default short_desc for any remaining plans
UPDATE subscription_plans 
SET short_desc = name
WHERE short_desc IS NULL;

-- Add comment for the new column
COMMENT ON COLUMN subscription_plans.short_desc IS 'Short description or tagline for the subscription plan';

-- Update the interval comment
COMMENT ON COLUMN subscription_plans.interval IS 'Billing interval (month, year, week, day)'; 