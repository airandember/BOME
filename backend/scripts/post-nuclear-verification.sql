-- 🔍 POST-NUCLEAR VERIFICATION
-- Confirm the exorcism was complete

-- This should return ZERO rows (empty result)
SELECT 
    'CONTAMINATION_CHECK' as check_type,
    stripe_price_id as legacy_id,
    COUNT(*) as subscription_count,
    'THIS_SHOULD_BE_EMPTY' as expected_result
FROM stripe_subscriptions 
WHERE stripe_price_id IS NOT NULL 
  AND stripe_price_id NOT LIKE 'price_%'  -- Legacy contamination
GROUP BY stripe_price_id
ORDER BY subscription_count DESC;

-- This should also return ZERO rows (empty result)  
SELECT 
    'GHOST_PRODUCTS_CHECK' as check_type,
    stripe_product_id as ghost_product,
    COUNT(*) as subscription_count,
    'THIS_SHOULD_ALSO_BE_EMPTY' as expected_result
FROM stripe_subscriptions 
WHERE stripe_product_id IN (
    'prod_HjYKGcWGP9r4EC', 'prod_HEmcX1PE8TO2CO', 'prod_FvNAeI348dup9w',
    'prod_FvNAlEGGL452nN', 'prod_HF5YzcBH5Rwr0d', 'prod_GVV5efccnh13h9', 'prod_FvNAJgnw48hwpZ'
)
GROUP BY stripe_product_id;

-- Final clean inventory - this should show only clean data
SELECT 
    'FINAL_CLEAN_INVENTORY' as inventory_type,
    COUNT(*) as total_subscriptions,
    COUNT(*) FILTER (WHERE stripe_price_id LIKE 'price_%') as confirmed_real_stripe,
    COUNT(*) FILTER (WHERE stripe_price_id NOT LIKE 'price_%' OR stripe_price_id IS NULL) as remaining_contamination,
    CASE 
        WHEN COUNT(*) FILTER (WHERE stripe_price_id NOT LIKE 'price_%' OR stripe_price_id IS NULL) = 0 
        THEN '✅ EXORCISM COMPLETE - ALL CLEAN!'
        ELSE '❌ CONTAMINATION REMAINS'
    END as exorcism_status
FROM stripe_subscriptions;
