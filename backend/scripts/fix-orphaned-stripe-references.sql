-- Fix orphaned stripe_price_id references in production
-- This addresses the "Combo" price ID issue and similar orphaned references

-- 1. First, let's see what orphaned price IDs we have
SELECT 'Orphaned Price IDs' as check_type;
SELECT ss.stripe_price_id, COUNT(*) as count
FROM stripe_subscriptions ss
LEFT JOIN stripe_prices sp ON ss.stripe_price_id = sp.stripe_id
WHERE ss.stripe_price_id IS NOT NULL 
  AND sp.stripe_id IS NULL
GROUP BY ss.stripe_price_id
ORDER BY count DESC;

-- 2. Check what orphaned product IDs we have
SELECT 'Orphaned Product IDs' as check_type;
SELECT ss.stripe_product_id, COUNT(*) as count
FROM stripe_subscriptions ss
LEFT JOIN stripe_products sp ON ss.stripe_product_id = sp.stripe_id
WHERE ss.stripe_product_id IS NOT NULL 
  AND sp.stripe_id IS NULL
GROUP BY ss.stripe_product_id
ORDER BY count DESC;

-- 3. Fix orphaned price references by setting them to NULL
-- This allows the COALESCE to fall back to other price sources
UPDATE stripe_subscriptions 
SET stripe_price_id = NULL
WHERE stripe_price_id IS NOT NULL 
  AND stripe_price_id NOT IN (SELECT stripe_id FROM stripe_prices WHERE stripe_id IS NOT NULL);

-- 4. Fix orphaned product references by setting them to NULL
-- This allows the COALESCE to fall back to other product name sources
UPDATE stripe_subscriptions 
SET stripe_product_id = NULL
WHERE stripe_product_id IS NOT NULL 
  AND stripe_product_id NOT IN (SELECT stripe_id FROM stripe_products WHERE stripe_id IS NOT NULL);

-- 5. Show the results
SELECT 'Fixed Records Count' as check_type;
SELECT 
  COUNT(*) as total_subscriptions,
  COUNT(stripe_price_id) as with_price_id,
  COUNT(stripe_product_id) as with_product_id
FROM stripe_subscriptions;

-- 6. Test Adam Arp's record again after the fix
SELECT 'Adam Arp After Fix' as check_type;
SELECT 
    u.id, u.email,
    u.sub_id,
    sp.id as legacy_plan_id,
    sp.name as legacy_plan_name,
    ss.stripe_product_id,
    stripe_prod.name as stripe_product_name,
    ss.product_name as fallback_product_name,
    COALESCE(
        sp.name,                    -- Legacy subscription plan name (highest priority)
        stripe_prod.name,           -- Current Stripe product name  
        ss.product_name,            -- Fallback from stripe_subscriptions
        CASE 
            WHEN ss.status = 'active' THEN 'Active Subscription'
            WHEN ss.status = 'trialing' THEN 'Trial Subscription'
            ELSE 'Subscription'
        END
    ) as final_plan_name
FROM users u
-- Join legacy subscription plans (for users with sub_id)
LEFT JOIN subscription_plans sp ON u.sub_id = sp.id 
    AND sp.is_active = true 
    AND sp.deleted_at IS NULL
-- Join Stripe customers (for users with stripe_customer_id)
LEFT JOIN stripe_customers sc ON (
    u.stripe_customer_id = sc.stripe_id OR 
    sc.stripe_id = ANY(COALESCE(u.stripe_customer_ids, '{}'))
)
-- Join current Stripe subscriptions (for ACTIVE Stripe subs only)
LEFT JOIN stripe_subscriptions ss ON sc.id = ss.customer_id 
    AND ss.status IN ('active', 'trialing')
    AND (ss.current_period_end IS NULL OR ss.current_period_end > NOW())
-- Join Stripe products (to get current product names)
LEFT JOIN stripe_products stripe_prod ON ss.stripe_product_id = stripe_prod.stripe_id
WHERE u.email = 'mycommonsensefinancial@yahoo.com';
