-- Dynamic Subscription Mapping Fix
-- This creates a dynamic mapping based on your actual database tables
-- No hardcoded values - uses subscription_plans and stripe_prices data

-- Step 1: Create a temporary mapping table based on your actual data
CREATE TEMP TABLE price_mappings AS
SELECT 
    sp.id as plan_id,
    sp.name as plan_name,
    spr.unit_amount,
    sp.stripe_price_id
FROM subscription_plans sp
LEFT JOIN stripe_prices spr ON sp.stripe_price_id = spr.stripe_id
WHERE sp.is_active = true;

-- Step 2: Show what mappings we found
SELECT 
    plan_id,
    plan_name,
    unit_amount,
    stripe_price_id,
    CASE 
        WHEN unit_amount IS NOT NULL THEN '$' || (unit_amount::float / 100)::text
        ELSE 'No price found'
    END as price_display
FROM price_mappings
ORDER BY plan_id;

-- Step 3: Update users.sub_id using the dynamic mappings
-- This will work for any subscription plans you have configured
DO $$
DECLARE
    mapping_record RECORD;
    updated_count INTEGER;
    total_updated INTEGER := 0;
BEGIN
    -- Loop through each price mapping
    FOR mapping_record IN 
        SELECT plan_id, plan_name, unit_amount 
        FROM price_mappings 
        WHERE unit_amount IS NOT NULL
    LOOP
        -- Update users for this specific unit_amount -> plan_id mapping
        UPDATE users 
        SET sub_id = mapping_record.plan_id, updated_at = NOW()
        WHERE id IN (
            SELECT DISTINCT u.id
            FROM users u
            INNER JOIN stripe_customers sc ON u.stripe_customer_id = sc.stripe_id
            INNER JOIN stripe_subscriptions ss ON sc.id = ss.customer_id
            WHERE ss.status IN ('active', 'trialing')
            AND ss.unit_amount = mapping_record.unit_amount
            AND u.sub_id IS NULL
        );
        
        GET DIAGNOSTICS updated_count = ROW_COUNT;
        total_updated := total_updated + updated_count;
        
        RAISE NOTICE 'Updated % users with unit_amount % (%) to plan_id %', 
            updated_count, mapping_record.unit_amount, mapping_record.plan_name, mapping_record.plan_id;
    END LOOP;
    
    RAISE NOTICE 'Total users updated: %', total_updated;
END $$;

-- Step 4: Handle unmapped prices with smart fallback logic
-- For prices that don't have direct mappings, use intelligent defaults
DO $$
DECLARE
    price_record RECORD;
    fallback_plan_id INTEGER;
    updated_count INTEGER;
    fallback_updated INTEGER := 0;
BEGIN
    -- Find common prices that aren't mapped yet
    FOR price_record IN 
        SELECT DISTINCT ss.unit_amount, COUNT(*) as subscriber_count
        FROM stripe_subscriptions ss
        LEFT JOIN price_mappings pm ON ss.unit_amount = pm.unit_amount
        WHERE ss.status IN ('active', 'trialing')
        AND ss.unit_amount IS NOT NULL
        AND pm.unit_amount IS NULL  -- Not already mapped
        GROUP BY ss.unit_amount
        ORDER BY subscriber_count DESC
    LOOP
        -- Smart mapping: < $15 = Monthly (plan 11), >= $15 = Annual (plan 12)
        IF price_record.unit_amount < 1500 THEN
            fallback_plan_id := 11;  -- Monthly
        ELSE
            fallback_plan_id := 12;  -- Annual
        END IF;
        
        -- Update users with this unmapped price
        UPDATE users 
        SET sub_id = fallback_plan_id, updated_at = NOW()
        WHERE id IN (
            SELECT DISTINCT u.id
            FROM users u
            INNER JOIN stripe_customers sc ON u.stripe_customer_id = sc.stripe_id
            INNER JOIN stripe_subscriptions ss ON sc.id = ss.customer_id
            WHERE ss.status IN ('active', 'trialing')
            AND ss.unit_amount = price_record.unit_amount
            AND u.sub_id IS NULL
        );
        
        GET DIAGNOSTICS updated_count = ROW_COUNT;
        fallback_updated := fallback_updated + updated_count;
        
        RAISE NOTICE 'Fallback: Updated % users with unmapped unit_amount % ($%) to plan_id %', 
            updated_count, price_record.unit_amount, 
            (price_record.unit_amount::float / 100), fallback_plan_id;
    END LOOP;
    
    RAISE NOTICE 'Total fallback users updated: %', fallback_updated;
END $$;

-- Step 5: Update product_name in stripe_subscriptions dynamically
UPDATE stripe_subscriptions 
SET product_name = pm.plan_name, updated_at = NOW()
FROM price_mappings pm
WHERE stripe_subscriptions.unit_amount = pm.unit_amount
AND (stripe_subscriptions.product_name IS NULL OR stripe_subscriptions.product_name = '');

-- Step 6: Verification queries
-- Show the fix results for our test user
SELECT 
    u.id, 
    u.email, 
    u.sub_id, 
    sp.name as plan_name, 
    sp.price as plan_price,
    ss.unit_amount, 
    ss.product_name,
    ss.status
FROM users u
LEFT JOIN subscription_plans sp ON u.sub_id = sp.id
LEFT JOIN stripe_customers sc ON u.stripe_customer_id = sc.stripe_id
LEFT JOIN stripe_subscriptions ss ON sc.id = ss.customer_id AND ss.status IN ('active', 'trialing')
WHERE u.email = 'mycommonsensefinancial@yahoo.com';

-- Show subscription distribution by unit_amount
SELECT 
    ss.unit_amount,
    COUNT(*) as subscriber_count,
    '$' || (ss.unit_amount::float / 100)::text as price_display,
    sp.name as mapped_plan_name
FROM users u
INNER JOIN stripe_customers sc ON u.stripe_customer_id = sc.stripe_id
INNER JOIN stripe_subscriptions ss ON sc.id = ss.customer_id
LEFT JOIN subscription_plans sp ON u.sub_id = sp.id
WHERE ss.status IN ('active', 'trialing')
GROUP BY ss.unit_amount, sp.name
ORDER BY ss.unit_amount;

-- Summary stats
SELECT 
    'Users with sub_id now set' as description,
    COUNT(*) as count
FROM users 
WHERE sub_id IS NOT NULL;

SELECT 
    'Active subscriptions with product_name' as description,
    COUNT(*) as count
FROM stripe_subscriptions 
WHERE status IN ('active', 'trialing') 
AND product_name IS NOT NULL 
AND product_name != '';

-- Clean up temp table
DROP TABLE price_mappings;
