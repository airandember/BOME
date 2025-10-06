-- 👻 GHOST EXORCISM RITUAL 👻
-- Step-by-step ghost removal with safety checks

-- STEP 1: IDENTIFY ALL GHOSTS
-- Show exactly what we're dealing with
SELECT 
    'GHOST_CENSUS' as ghost_type,
    CASE 
        WHEN stripe_price_id IS NOT NULL AND stripe_price_id NOT LIKE 'price_%' THEN 'LEGACY_CONTAMINATION'
        WHEN stripe_product_id IS NOT NULL AND stripe_product_id IN (
            'prod_HjYKGcWGP9r4EC', 'prod_HEmcX1PE8TO2CO', 'prod_FvNAeI348dup9w',
            'prod_FvNAlEGGL452nN', 'prod_HF5YzcBH5Rwr0d', 'prod_GVV5efccnh13h9', 'prod_FvNAJgnw48hwpZ'
        ) THEN 'ORPHANED_PRODUCT_GHOST'
        WHEN stripe_price_id IS NULL AND stripe_product_id IS NULL THEN 'COMPLETELY_BROKEN'
        WHEN stripe_price_id IS NULL THEN 'MISSING_PRICE_ID'
        WHEN stripe_product_id IS NULL OR stripe_product_id = '' THEN 'MISSING_PRODUCT_ID'
        ELSE 'UNKNOWN_GHOST'
    END as ghost_category,
    COUNT(*) as ghost_count,
    COUNT(*) FILTER (WHERE status IN ('active', 'trialing')) as active_ghosts,
    COUNT(*) FILTER (WHERE status = 'canceled') as canceled_ghosts,
    array_agg(DISTINCT status) as ghost_statuses
FROM stripe_subscriptions 
WHERE stripe_price_id NOT LIKE 'price_%' OR stripe_price_id IS NULL
GROUP BY ghost_category
ORDER BY ghost_count DESC;

-- STEP 2: SAFETY CHECK - Find customers who would be affected
SELECT 
    'AFFECTED_CUSTOMERS' as check_type,
    COUNT(DISTINCT sc.stripe_id) as customers_with_ghosts,
    COUNT(DISTINCT u.id) as users_with_ghosts
FROM stripe_subscriptions ss
LEFT JOIN stripe_customers sc ON ss.customer_id = sc.id
LEFT JOIN users u ON sc.stripe_id = u.stripe_customer_id
WHERE (ss.stripe_price_id NOT LIKE 'price_%' OR ss.stripe_price_id IS NULL)
  AND ss.status IN ('active', 'trialing');

-- STEP 3: IDENTIFY LEGACY DATA THAT SHOULD MOVE BACK TO subscription_plans
SELECT 
    'LEGACY_RELOCATION_CANDIDATES' as analysis_type,
    ss.stripe_price_id as legacy_price_id,
    COUNT(*) as subscription_count,
    sp.name as matching_plan_name,
    sp.price as matching_plan_price,
    CASE 
        WHEN sp.stripe_price_id IS NOT NULL THEN 'HAS_MATCHING_PLAN'
        ELSE 'NO_MATCHING_PLAN'
    END as plan_match_status
FROM stripe_subscriptions ss
LEFT JOIN subscription_plans sp ON ss.stripe_price_id = sp.stripe_price_id
WHERE ss.stripe_price_id IS NOT NULL 
  AND ss.stripe_price_id NOT LIKE 'price_%'
GROUP BY ss.stripe_price_id, sp.name, sp.price, sp.stripe_price_id
ORDER BY subscription_count DESC
LIMIT 10;
