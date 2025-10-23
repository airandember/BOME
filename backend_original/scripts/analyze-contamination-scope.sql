-- URGENT: Show me the legacy price IDs that are contaminating stripe_subscriptions
-- This will reveal the scope of the data migration disaster

SELECT 
    'TOP_LEGACY_CONTAMINANTS' as analysis_type,
    stripe_price_id as legacy_id,
    COUNT(*) as subscription_count,
    ROUND(COUNT(*) * 100.0 / 1484, 2) as percentage_of_legacy,
    array_agg(DISTINCT status) as statuses,
    MIN(created_at) as oldest_contamination,
    MAX(created_at) as newest_contamination
FROM stripe_subscriptions 
WHERE stripe_price_id IS NOT NULL 
  AND stripe_price_id NOT LIKE 'price_%'
GROUP BY stripe_price_id
ORDER BY subscription_count DESC
LIMIT 20;

-- Show me what the 7 REAL Stripe subscriptions look like
SELECT 
    'REAL_STRIPE_SUBSCRIPTIONS' as analysis_type,
    stripe_id,
    stripe_price_id,
    unit_amount,
    currency,
    stripe_product_id,
    product_name,
    status,
    created_at
FROM stripe_subscriptions 
WHERE stripe_price_id LIKE 'price_%'
ORDER BY created_at DESC;
