-- GHOST PRODUCTS INVESTIGATION
-- Check if these problematic product IDs exist in our database vs Stripe export

-- 1. Check what's in our stripe_products table
SELECT 
    'OUR_DATABASE' as source,
    stripe_id,
    name,
    active,
    created_at,
    CASE 
        WHEN name IS NULL OR name = '' THEN '❌ MISSING_NAME'
        ELSE '✅ HAS_NAME'
    END as name_status
FROM stripe_products 
WHERE stripe_id IN (
    'prod_HjYKGcWGP9r4EC',
    'prod_HEmcX1PE8TO2CO', 
    'prod_FvNAeI348dup9w',
    'prod_FvNAlEGGL452nN',
    'prod_HF5YzcBH5Rwr0d',
    'prod_GVV5efccnh13h9',
    'prod_FvNAJgnw48hwpZ'
)
ORDER BY stripe_id;

-- 2. Check how many subscriptions reference these ghost products
SELECT 
    stripe_product_id,
    COUNT(*) as subscription_count,
    COUNT(DISTINCT customer_id) as unique_customers,
    MIN(created_at) as oldest_subscription,
    MAX(created_at) as newest_subscription,
    array_agg(DISTINCT status) as statuses
FROM stripe_subscriptions 
WHERE stripe_product_id IN (
    'prod_HjYKGcWGP9r4EC',
    'prod_HEmcX1PE8TO2CO', 
    'prod_FvNAeI348dup9w',
    'prod_FvNAlEGGL452nN',
    'prod_HF5YzcBH5Rwr0d',
    'prod_GVV5efccnh13h9',
    'prod_FvNAJgnw48hwpZ'
)
GROUP BY stripe_product_id
ORDER BY subscription_count DESC;

-- 3. Check if we have any REAL products that match the pricing
SELECT 
    'REAL_PRODUCTS_BY_PRICE' as source,
    sp.stripe_id as product_id,
    sp.name as product_name,
    price.unit_amount,
    price.currency,
    COUNT(subs.id) as subscription_count
FROM stripe_products sp
JOIN stripe_prices price ON sp.stripe_id = price.product_id
LEFT JOIN stripe_subscriptions subs ON sp.stripe_id = subs.stripe_product_id
WHERE price.unit_amount IN (997, 9564, 4500, 1397, 1393)
GROUP BY sp.stripe_id, sp.name, price.unit_amount, price.currency
ORDER BY price.unit_amount;
