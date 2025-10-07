-- 🔍 DEBUG: Check what the enhanced subscribers query returns for ghost customers
-- This will help us understand why plan names are showing as N/A

-- 1. Sample of customers with hash IDs and their plan data
SELECT 
    'GHOST_CUSTOMERS_PLAN_DEBUG' as debug_type,
    u.id,
    u.email,
    u.stripe_customer_id,
    u.sub_id as legacy_subscription_id,
    sp.name as legacy_plan_name,
    ss.stripe_id as stripe_subscription_id,
    ss.product_name as stripe_product_name,
    ss.status as stripe_status,
    'PLAN_DATA_CHECK' as result
FROM users u
LEFT JOIN subscription_plans sp ON u.sub_id = sp.id
LEFT JOIN stripe_customers sc ON u.stripe_customer_id = sc.stripe_id
LEFT JOIN stripe_subscriptions ss ON sc.id = ss.customer_id AND ss.status IN ('active', 'trialing')
WHERE u.stripe_customer_id LIKE '#%'
  AND u.is_active = true
ORDER BY u.created_at DESC
LIMIT 20;

-- 2. Check what the enhanced subscribers query would return for these users
SELECT 
    'ENHANCED_QUERY_SIMULATION' as debug_type,
    u.id,
    u.email,
    u.stripe_customer_id,
    COALESCE(sp.name, ss.product_name, 'No Plan Found') as plan_name_result,
    CASE 
        WHEN sp.name IS NOT NULL THEN 'FROM_LEGACY_PLAN'
        WHEN ss.product_name IS NOT NULL THEN 'FROM_STRIPE_PRODUCT'
        ELSE 'NO_PLAN_SOURCE'
    END as plan_source
FROM users u
LEFT JOIN subscription_plans sp ON u.sub_id = sp.id AND sp.is_active = true
LEFT JOIN stripe_customers sc ON u.stripe_customer_id = sc.stripe_id
LEFT JOIN stripe_subscriptions ss ON sc.id = ss.customer_id AND ss.status IN ('active', 'trialing')
WHERE u.stripe_customer_id LIKE '#%'
  AND u.is_active = true
ORDER BY u.created_at DESC
LIMIT 20;

-- 3. Count how many ghost customers have plans vs no plans
SELECT 
    'GHOST_PLAN_SUMMARY' as summary_type,
    COUNT(*) as total_ghost_customers,
    COUNT(*) FILTER (WHERE sp.name IS NOT NULL) as have_legacy_plans,
    COUNT(*) FILTER (WHERE ss.product_name IS NOT NULL) as have_stripe_plans,
    COUNT(*) FILTER (WHERE sp.name IS NULL AND ss.product_name IS NULL) as have_no_plans
FROM users u
LEFT JOIN subscription_plans sp ON u.sub_id = sp.id AND sp.is_active = true
LEFT JOIN stripe_customers sc ON u.stripe_customer_id = sc.stripe_id
LEFT JOIN stripe_subscriptions ss ON sc.id = ss.customer_id AND ss.status IN ('active', 'trialing')
WHERE u.stripe_customer_id LIKE '#%'
  AND u.is_active = true;
