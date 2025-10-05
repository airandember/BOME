-- URGENT: Fix missing product names in production stripe_subscriptions
-- This will resolve Adam Arp and all other missing plan names immediately

-- Step 1: Update all subscriptions with missing product names
UPDATE stripe_subscriptions ss
SET product_name = sp.name
FROM stripe_products sp
WHERE ss.stripe_product_id = sp.stripe_id
  AND (ss.product_name IS NULL OR ss.product_name = '')
  AND ss.status IN ('active', 'trialing');

-- Step 2: Verify the fix worked
SELECT 
    COUNT(*) as total_fixed,
    COUNT(*) FILTER (WHERE product_name IS NOT NULL AND product_name != '') as now_have_names
FROM stripe_subscriptions ss
JOIN stripe_products sp ON ss.stripe_product_id = sp.stripe_id
WHERE ss.status IN ('active', 'trialing');

-- Step 3: Show the specific Adam Arp fix
SELECT 
    u.email,
    ss.stripe_id,
    ss.stripe_product_id,
    ss.product_name as updated_product_name,
    sp.name as source_product_name,
    'FIXED' as status
FROM users u
JOIN stripe_customers sc ON u.stripe_customer_id = sc.stripe_id
JOIN stripe_subscriptions ss ON sc.id = ss.customer_id
JOIN stripe_products sp ON ss.stripe_product_id = sp.stripe_id
WHERE u.email = 'mycommonsensefinancial@yahoo.com'
  AND ss.status IN ('active', 'trialing');
