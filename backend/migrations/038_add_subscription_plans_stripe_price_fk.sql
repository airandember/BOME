-- Add missing foreign key constraint for subscription_plans.stripe_price_id
-- This ensures proper referential integrity between subscription plans and Stripe prices

-- First, let's check if the constraint already exists
DO $$
BEGIN
    -- Add the foreign key constraint if it doesn't exist
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_subscription_plans_stripe_price_id'
        AND table_name = 'subscription_plans'
    ) THEN
        -- Add the foreign key constraint
        ALTER TABLE subscription_plans 
        ADD CONSTRAINT fk_subscription_plans_stripe_price_id 
        FOREIGN KEY (stripe_price_id) REFERENCES stripe_prices(stripe_id);
        
        RAISE NOTICE 'Added FK constraint: subscription_plans.stripe_price_id -> stripe_prices.stripe_id';
    ELSE
        RAISE NOTICE 'FK constraint already exists: subscription_plans.stripe_price_id -> stripe_prices.stripe_id';
    END IF;
END $$;

-- Add index for better performance on the FK column
CREATE INDEX IF NOT EXISTS idx_subscription_plans_stripe_price_id 
ON subscription_plans(stripe_price_id);

-- Verify the relationships work correctly
SELECT 
    sp.id as plan_id,
    sp.name as plan_name,
    sp.stripe_price_id,
    spr.stripe_id as price_stripe_id,
    spr.unit_amount,
    spr.currency,
    spr.recurring_interval
FROM subscription_plans sp
LEFT JOIN stripe_prices spr ON sp.stripe_price_id = spr.stripe_id
WHERE sp.is_active = true
ORDER BY sp.id;
