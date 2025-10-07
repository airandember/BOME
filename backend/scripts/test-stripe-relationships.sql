-- Test query to verify the stripe relationships work correctly
-- This will help us understand why plan names aren't showing

-- 1. Test the stripe_subscriptions -> stripe_prices -> stripe_products relationship
SELECT 
    ss.stripe_id as subscription_id,
    ss.stripe_price_id,
    ss.unit_amount,
    ss.currency,
    ss.stripe_product_id as sub_product_id,
    ss.product_name as sub_product_name,
    sp.stripe_id as price_stripe_id,
    sp.unit_amount as price_unit_amount,
    sp.stripe_product_id as price_product_id,
    prod.stripe_id as product_stripe_id,
    prod.name as product_name
FROM stripe_subscriptions ss
LEFT JOIN stripe_prices sp ON ss.stripe_price_id = sp.stripe_id
LEFT JOIN stripe_products prod ON sp.stripe_product_id = prod.stripe_id
WHERE ss.status IN ('active', 'trialing')
  AND (ss.current_period_end IS NULL OR ss.current_period_end > NOW())
ORDER BY ss.created_at DESC
LIMIT 10;

-- 2. Check if stripe_prices has the stripe_product_id column we need
SELECT column_name, data_type, is_nullable
FROM information_schema.columns 
WHERE table_name = 'stripe_prices' 
AND table_schema = 'public'
ORDER BY ordinal_position;

-- 3. Test a specific example - find Aaron Andrew's subscription
SELECT 
    u.first_name,
    u.last_name,
    u.email,
    u.stripe_customer_id,
    sc.stripe_id as customer_stripe_id,
    ss.stripe_id as subscription_id,
    ss.stripe_price_id,
    ss.unit_amount,
    ss.product_name as subscription_product_name,
    sp.stripe_id as price_id,
    sp.stripe_product_id as price_product_id,
    prod.name as product_name
FROM users u
JOIN stripe_customers sc ON u.stripe_customer_id = sc.stripe_id
JOIN stripe_subscriptions ss ON sc.id = ss.customer_id
LEFT JOIN stripe_prices sp ON ss.stripe_price_id = sp.stripe_id  
LEFT JOIN stripe_products prod ON sp.stripe_product_id = prod.stripe_id
WHERE u.first_name = 'Aaron' AND u.last_name = 'Andrew'
  AND ss.status IN ('active', 'trialing')
ORDER BY ss.created_at DESC;
