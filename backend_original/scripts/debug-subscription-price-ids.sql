-- Debug script to understand the subscription price ID issue
-- This will help us see what's happening with the price IDs

-- 1. Check current subscription data to see the price ID pattern
SELECT 
    stripe_id,
    stripe_price_id,
    unit_amount,
    currency,
    stripe_product_id,
    product_name,
    status,
    created_at
FROM stripe_subscriptions 
WHERE status IN ('active', 'trialing')
ORDER BY created_at DESC
LIMIT 10;

-- 2. Check if we have real Stripe price IDs vs legacy ones
SELECT 
    'REAL_STRIPE_PRICE_IDS' as analysis_type,
    COUNT(*) as count
FROM stripe_subscriptions 
WHERE stripe_price_id LIKE 'price_%'
UNION ALL
SELECT 
    'LEGACY_PRICE_IDS' as analysis_type,
    COUNT(*) as count  
FROM stripe_subscriptions 
WHERE stripe_price_id NOT LIKE 'price_%' AND stripe_price_id IS NOT NULL
UNION ALL
SELECT 
    'NULL_PRICE_IDS' as analysis_type,
    COUNT(*) as count
FROM stripe_subscriptions 
WHERE stripe_price_id IS NULL;

-- 3. Check which legacy price IDs we still have
SELECT 
    stripe_price_id,
    COUNT(*) as subscription_count,
    MIN(created_at) as earliest_subscription,
    MAX(created_at) as latest_subscription
FROM stripe_subscriptions 
WHERE stripe_price_id NOT LIKE 'price_%' 
    AND stripe_price_id IS NOT NULL
    AND status IN ('active', 'trialing')
GROUP BY stripe_price_id
ORDER BY subscription_count DESC;

-- 4. Check the stripe_prices table for real vs legacy price IDs
SELECT 
    'REAL_STRIPE_PRICES' as analysis_type,
    COUNT(*) as count
FROM stripe_prices 
WHERE stripe_id LIKE 'price_%'
UNION ALL
SELECT 
    'LEGACY_PRICES' as analysis_type,
    COUNT(*) as count
FROM stripe_prices 
WHERE stripe_id NOT LIKE 'price_%';
