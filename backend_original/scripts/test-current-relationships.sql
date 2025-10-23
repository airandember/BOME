-- Test the CURRENT stripe relationships using the existing product_id column
-- This should work since product_id points to stripe_products.id

SELECT 
    ss.stripe_id as subscription_id,
    ss.stripe_price_id,
    ss.unit_amount,
    ss.currency,
    ss.product_name as sub_product_name,
    sp.stripe_id as price_stripe_id,
    sp.product_id as price_internal_product_id,
    prod.id as product_internal_id,
    prod.stripe_id as product_stripe_id,
    prod.name as product_name
FROM stripe_subscriptions ss
LEFT JOIN stripe_prices sp ON ss.stripe_price_id = sp.stripe_id
LEFT JOIN stripe_products prod ON sp.product_id = prod.id  -- Using INTERNAL ID relationship
WHERE ss.status IN ('active', 'trialing')
  AND (ss.current_period_end IS NULL OR ss.current_period_end > NOW())
ORDER BY ss.created_at DESC
LIMIT 10;

-- Also test a specific example - Aaron Andrew
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
    sp.product_id as price_internal_product_id,
    prod.id as product_internal_id,
    prod.stripe_id as product_stripe_id,
    prod.name as product_name
FROM users u
JOIN stripe_customers sc ON u.stripe_customer_id = sc.stripe_id
JOIN stripe_subscriptions ss ON sc.id = ss.customer_id
LEFT JOIN stripe_prices sp ON ss.stripe_price_id = sp.stripe_id  
LEFT JOIN stripe_products prod ON sp.product_id = prod.id  -- Using INTERNAL ID
WHERE u.first_name = 'Aaron' AND u.last_name = 'Andrew'
  AND ss.status IN ('active', 'trialing')
ORDER BY ss.created_at DESC;
