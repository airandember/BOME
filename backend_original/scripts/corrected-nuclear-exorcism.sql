-- ☢️ CORRECTED NUCLEAR EXORCISM CODES ☢️
-- Fixed SQL syntax - no bash commands in SQL!

-- STEP 1: BACKUP (just in case)
CREATE TABLE stripe_subscriptions_backup_20241006 AS 
SELECT * FROM stripe_subscriptions;

-- STEP 2: IDENTIFY WHAT TO KEEP (real Stripe data only)
SELECT 
    'PRE_NUCLEAR_INVENTORY' as inventory_type,
    COUNT(*) as total_before_nuclear,
    COUNT(*) FILTER (WHERE stripe_price_id LIKE 'price_%') as real_stripe_to_keep,
    COUNT(*) FILTER (WHERE stripe_price_id NOT LIKE 'price_%' OR stripe_price_id IS NULL) as contamination_to_delete
FROM stripe_subscriptions;

-- STEP 3: NUCLEAR STRIKE! 💥
-- Delete all non-Stripe data
DELETE FROM stripe_subscriptions 
WHERE stripe_price_id IS NULL 
   OR stripe_price_id NOT LIKE 'price_%';

-- STEP 4: VERIFY NUCLEAR RESULTS
SELECT 
    'POST_NUCLEAR_INVENTORY' as inventory_type,
    COUNT(*) as total_after_nuclear,
    COUNT(*) FILTER (WHERE stripe_price_id LIKE 'price_%') as confirmed_real_stripe,
    COUNT(*) FILTER (WHERE stripe_product_id LIKE 'prod_%') as confirmed_real_products
FROM stripe_subscriptions;

-- STEP 5: CLEAN UP RELATED GHOST DATA
-- Also clean up any orphaned stripe_customers that no longer have subscriptions
WITH customers_with_subscriptions AS (
    SELECT DISTINCT customer_id 
    FROM stripe_subscriptions 
    WHERE customer_id IS NOT NULL
)
SELECT 
    'ORPHANED_CUSTOMERS_CHECK' as check_type,
    COUNT(*) as total_stripe_customers,
    COUNT(*) FILTER (WHERE sc.id IN (SELECT customer_id FROM customers_with_subscriptions)) as customers_with_subscriptions,
    COUNT(*) FILTER (WHERE sc.id NOT IN (SELECT customer_id FROM customers_with_subscriptions)) as orphaned_customers
FROM stripe_customers sc;

-- STEP 6: READY FOR FRESH SYNC
SELECT 
    'EXORCISM_COMPLETE' as status,
    '🎉 All ghost data has been banished! Ready for fresh Stripe sync! 🎉' as message;
