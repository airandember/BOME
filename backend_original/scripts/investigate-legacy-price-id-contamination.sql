-- CRITICAL INVESTIGATION: How did "YPremium" get into stripe_subscriptions?
-- This should NOT happen - stripe_subscriptions should only have real Stripe price IDs!

-- 1. Check all subscriptions with legacy price IDs (not starting with "price_")
SELECT 
    'LEGACY_PRICE_IDS_IN_STRIPE_SUBS' as issue_type,
    stripe_price_id,
    COUNT(*) as subscription_count,
    array_agg(DISTINCT status) as statuses,
    MIN(created_at) as oldest,
    MAX(created_at) as newest
FROM stripe_subscriptions 
WHERE stripe_price_id IS NOT NULL 
  AND stripe_price_id NOT LIKE 'price_%'  -- Real Stripe price IDs start with "price_"
GROUP BY stripe_price_id
ORDER BY subscription_count DESC;

-- 2. Check if these legacy price IDs exist in subscription_plans
SELECT 
    'MATCHING_LEGACY_PLANS' as check_type,
    sp.stripe_price_id as legacy_price_id,
    sp.name as plan_name,
    sp.price as plan_price,
    sp.stripe_product_id,
    COUNT(ss.id) as stripe_subscriptions_using_this
FROM subscription_plans sp
LEFT JOIN stripe_subscriptions ss ON sp.stripe_price_id = ss.stripe_price_id
WHERE sp.stripe_price_id NOT LIKE 'price_%'  -- Legacy IDs
GROUP BY sp.stripe_price_id, sp.name, sp.price, sp.stripe_product_id
ORDER BY stripe_subscriptions_using_this DESC;

-- 3. Check the specific "YPremium" case
SELECT 
    'YPREMIUM_INVESTIGATION' as test_type,
    ss.stripe_id,
    ss.stripe_price_id,
    ss.unit_amount,
    ss.currency,
    ss.stripe_product_id,
    ss.product_name,
    ss.status,
    ss.created_at,
    sp.name as matching_plan_name,
    sp.price as matching_plan_price,
    sp.stripe_product_id as matching_plan_product_id
FROM stripe_subscriptions ss
LEFT JOIN subscription_plans sp ON ss.stripe_price_id = sp.stripe_price_id
WHERE ss.stripe_price_id = 'YPremium'
LIMIT 5;

-- 4. Check how many subscriptions have this data inconsistency
SELECT 
    'DATA_INCONSISTENCY_SUMMARY' as summary_type,
    COUNT(*) as total_stripe_subscriptions,
    COUNT(*) FILTER (WHERE stripe_price_id LIKE 'price_%') as real_stripe_price_ids,
    COUNT(*) FILTER (WHERE stripe_price_id IS NOT NULL AND stripe_price_id NOT LIKE 'price_%') as legacy_price_ids,
    COUNT(*) FILTER (WHERE stripe_price_id IS NULL) as null_price_ids
FROM stripe_subscriptions;
