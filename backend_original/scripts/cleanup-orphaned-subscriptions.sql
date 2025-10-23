-- =====================================================
-- STRIPE SUBSCRIPTIONS CLEANUP SCRIPT
-- =====================================================
-- This script identifies and cleans up orphaned Stripe subscriptions
-- that have invalid product references or missing product names.

-- =====================================================
-- 1. DIAGNOSTIC QUERIES - Run these first to assess the scope
-- =====================================================

-- Check total active subscriptions
SELECT 
    'Total Active Subscriptions' as metric,
    COUNT(*) as count
FROM stripe_subscriptions 
WHERE status IN ('active', 'trialing') 
    AND (current_period_end IS NULL OR current_period_end > NOW());

-- Check subscriptions with missing product names
SELECT 
    'Missing Product Names' as metric,
    COUNT(*) as count
FROM stripe_subscriptions 
WHERE status IN ('active', 'trialing') 
    AND (current_period_end IS NULL OR current_period_end > NOW())
    AND (product_name IS NULL OR product_name = '');

-- Check subscriptions with orphaned product IDs
SELECT 
    'Orphaned Product IDs' as metric,
    COUNT(*) as count
FROM stripe_subscriptions ss
LEFT JOIN stripe_products sp ON ss.stripe_product_id = sp.stripe_id
WHERE ss.status IN ('active', 'trialing') 
    AND (ss.current_period_end IS NULL OR ss.current_period_end > NOW())
    AND ss.stripe_product_id IS NOT NULL
    AND sp.stripe_id IS NULL;

-- Check subscriptions with valid product references
SELECT 
    'Valid Product References' as metric,
    COUNT(*) as count
FROM stripe_subscriptions ss
INNER JOIN stripe_products sp ON ss.stripe_product_id = sp.stripe_id
WHERE ss.status IN ('active', 'trialing') 
    AND (ss.current_period_end IS NULL OR ss.current_period_end > NOW())
    AND ss.product_name IS NOT NULL 
    AND ss.product_name != '';

-- =====================================================
-- 2. DETAILED ANALYSIS - See specific problematic records
-- =====================================================

-- Show subscriptions with missing/empty product names
SELECT 
    ss.id,
    ss.stripe_id,
    ss.status,
    ss.stripe_product_id,
    ss.product_name,
    ss.current_period_start,
    ss.current_period_end,
    sc.stripe_id as customer_stripe_id,
    u.email as user_email
FROM stripe_subscriptions ss
LEFT JOIN stripe_customers sc ON ss.customer_id = sc.id
LEFT JOIN users u ON (u.stripe_customer_id = sc.stripe_id OR sc.stripe_id = ANY(COALESCE(u.stripe_customer_ids, '{}')))
WHERE ss.status IN ('active', 'trialing') 
    AND (ss.current_period_end IS NULL OR ss.current_period_end > NOW())
    AND (ss.product_name IS NULL OR ss.product_name = '')
ORDER BY ss.current_period_end DESC
LIMIT 20;

-- Show subscriptions with orphaned product IDs
SELECT 
    ss.id,
    ss.stripe_id,
    ss.status,
    ss.stripe_product_id,
    ss.product_name,
    ss.current_period_start,
    ss.current_period_end,
    sc.stripe_id as customer_stripe_id,
    u.email as user_email
FROM stripe_subscriptions ss
LEFT JOIN stripe_products sp ON ss.stripe_product_id = sp.stripe_id
LEFT JOIN stripe_customers sc ON ss.customer_id = sc.id
LEFT JOIN users u ON (u.stripe_customer_id = sc.stripe_id OR sc.stripe_id = ANY(COALESCE(u.stripe_customer_ids, '{}')))
WHERE ss.status IN ('active', 'trialing') 
    AND (ss.current_period_end IS NULL OR ss.current_period_end > NOW())
    AND ss.stripe_product_id IS NOT NULL
    AND sp.stripe_id IS NULL
ORDER BY ss.current_period_end DESC
LIMIT 20;

-- =====================================================
-- 3. CLEANUP OPTIONS - Choose the appropriate action
-- =====================================================

-- OPTION A: Mark problematic subscriptions as 'canceled' (SAFEST)
-- This preserves the data but removes them from active consideration
/*
UPDATE stripe_subscriptions 
SET 
    status = 'canceled',
    updated_at = NOW()
WHERE id IN (
    SELECT ss.id
    FROM stripe_subscriptions ss
    LEFT JOIN stripe_products sp ON ss.stripe_product_id = sp.stripe_id
    WHERE ss.status IN ('active', 'trialing') 
        AND (ss.current_period_end IS NULL OR ss.current_period_end > NOW())
        AND (
            -- Missing product name
            (ss.product_name IS NULL OR ss.product_name = '')
            OR 
            -- Orphaned product ID
            (ss.stripe_product_id IS NOT NULL AND sp.stripe_id IS NULL)
        )
);
*/

-- OPTION B: Set current_period_end to past date (MODERATE)
-- This makes subscriptions appear expired without changing status
/*
UPDATE stripe_subscriptions 
SET 
    current_period_end = NOW() - INTERVAL '1 day',
    updated_at = NOW()
WHERE id IN (
    SELECT ss.id
    FROM stripe_subscriptions ss
    LEFT JOIN stripe_products sp ON ss.stripe_product_id = sp.stripe_id
    WHERE ss.status IN ('active', 'trialing') 
        AND (ss.current_period_end IS NULL OR ss.current_period_end > NOW())
        AND (
            -- Missing product name
            (ss.product_name IS NULL OR ss.product_name = '')
            OR 
            -- Orphaned product ID
            (ss.stripe_product_id IS NOT NULL AND sp.stripe_id IS NULL)
        )
);
*/

-- OPTION C: Delete orphaned subscriptions (MOST AGGRESSIVE - USE WITH CAUTION)
-- Only use this if you're certain these are invalid/test subscriptions
/*
DELETE FROM stripe_subscriptions 
WHERE id IN (
    SELECT ss.id
    FROM stripe_subscriptions ss
    LEFT JOIN stripe_products sp ON ss.stripe_product_id = sp.stripe_id
    WHERE ss.status IN ('active', 'trialing') 
        AND (ss.current_period_end IS NULL OR ss.current_period_end > NOW())
        AND (
            -- Missing product name
            (ss.product_name IS NULL OR ss.product_name = '')
            OR 
            -- Orphaned product ID
            (ss.stripe_product_id IS NOT NULL AND sp.stripe_id IS NULL)
        )
);
*/

-- =====================================================
-- 4. VERIFICATION QUERIES - Run after cleanup
-- =====================================================

-- Verify cleanup results
SELECT 
    'After Cleanup - Active Subscriptions' as metric,
    COUNT(*) as count
FROM stripe_subscriptions 
WHERE status IN ('active', 'trialing') 
    AND (current_period_end IS NULL OR current_period_end > NOW());

-- Verify no more orphaned subscriptions
SELECT 
    'After Cleanup - Orphaned Subscriptions' as metric,
    COUNT(*) as count
FROM stripe_subscriptions ss
LEFT JOIN stripe_products sp ON ss.stripe_product_id = sp.stripe_id
WHERE ss.status IN ('active', 'trialing') 
    AND (ss.current_period_end IS NULL OR ss.current_period_end > NOW())
    AND (
        (ss.product_name IS NULL OR ss.product_name = '')
        OR 
        (ss.stripe_product_id IS NOT NULL AND sp.stripe_id IS NULL)
    );

-- =====================================================
-- 5. RECOMMENDED APPROACH
-- =====================================================
-- 1. Run diagnostic queries first to understand scope
-- 2. Use OPTION A (mark as canceled) for safety
-- 3. Monitor the application to ensure no issues
-- 4. If needed, can always revert by changing status back
-- 5. Run verification queries to confirm cleanup

-- =====================================================
-- NOTES:
-- - Always backup your database before running cleanup
-- - Test on a staging environment first
-- - Consider the business impact of marking subscriptions as canceled
-- - You may want to notify affected users or investigate why these exist
-- =====================================================
