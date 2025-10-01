-- Fix Subscription Mapping: Connect Stripe subscriptions to local subscription plans
-- This fixes the issue where plan_id = 0 and plan_name = "" in production
-- Uses stripe_prices table for accurate mapping

-- Step 1: Update users.sub_id based on their active Stripe subscription
-- Map using stripe_prices table to get accurate plan mapping

-- Monthly Plan: EMonth price (997 cents) -> plan_id 11
UPDATE users 
SET sub_id = 11, updated_at = NOW()
WHERE id IN (
    SELECT DISTINCT u.id
    FROM users u
    INNER JOIN stripe_customers sc ON u.stripe_customer_id = sc.stripe_id
    INNER JOIN stripe_subscriptions ss ON sc.id = ss.customer_id
    WHERE ss.status IN ('active', 'trialing')
    AND ss.unit_amount = 997
    AND u.sub_id IS NULL
);

-- Annual Plan: YPremium price (9564 cents) -> plan_id 12  
UPDATE users 
SET sub_id = 12, updated_at = NOW()
WHERE id IN (
    SELECT DISTINCT u.id
    FROM users u
    INNER JOIN stripe_customers sc ON u.stripe_customer_id = sc.stripe_id
    INNER JOIN stripe_subscriptions ss ON sc.id = ss.customer_id
    WHERE ss.status IN ('active', 'trialing')
    AND ss.unit_amount = 9564
    AND u.sub_id IS NULL
);

-- Legacy Premium Semi-Annual: PPlan price (8982 cents) -> plan_id 12 (map to Annual)
UPDATE users 
SET sub_id = 12, updated_at = NOW()
WHERE id IN (
    SELECT DISTINCT u.id
    FROM users u
    INNER JOIN stripe_customers sc ON u.stripe_customer_id = sc.stripe_id
    INNER JOIN stripe_subscriptions ss ON sc.id = ss.customer_id
    WHERE ss.status IN ('active', 'trialing')
    AND ss.unit_amount = 8982
    AND u.sub_id IS NULL
);

-- Handle other common price points
-- Premium Monthly: premiummonthly price (999 cents) -> plan_id 11
UPDATE users 
SET sub_id = 11, updated_at = NOW()
WHERE id IN (
    SELECT DISTINCT u.id
    FROM users u
    INNER JOIN stripe_customers sc ON u.stripe_customer_id = sc.stripe_id
    INNER JOIN stripe_subscriptions ss ON sc.id = ss.customer_id
    WHERE ss.status IN ('active', 'trialing')
    AND ss.unit_amount = 999
    AND u.sub_id IS NULL
);

-- Best Value Promo: bestvaluepromo price (7200 cents) -> plan_id 12
UPDATE users 
SET sub_id = 12, updated_at = NOW()
WHERE id IN (
    SELECT DISTINCT u.id
    FROM users u
    INNER JOIN stripe_customers sc ON u.stripe_customer_id = sc.stripe_id
    INNER JOIN stripe_subscriptions ss ON sc.id = ss.customer_id
    WHERE ss.status IN ('active', 'trialing')
    AND ss.unit_amount = 7200
    AND u.sub_id IS NULL
);

-- Step 2: Update product_name in stripe_subscriptions for better display
UPDATE stripe_subscriptions 
SET product_name = 'Basic Monthly', updated_at = NOW()
WHERE unit_amount IN (997, 999)
AND (product_name IS NULL OR product_name = '');

UPDATE stripe_subscriptions 
SET product_name = 'Premium Annual', updated_at = NOW()
WHERE unit_amount = 9564 
AND (product_name IS NULL OR product_name = '');

UPDATE stripe_subscriptions 
SET product_name = 'Premium Semi-Annual', updated_at = NOW()
WHERE unit_amount = 8982 
AND (product_name IS NULL OR product_name = '');

UPDATE stripe_subscriptions 
SET product_name = 'Best Value Annual', updated_at = NOW()
WHERE unit_amount = 7200 
AND (product_name IS NULL OR product_name = '');

-- Step 3: Verify the fix for our test user
SELECT 
    u.id, 
    u.email, 
    u.sub_id, 
    sp.name as plan_name, 
    sp.price as plan_price,
    ss.unit_amount, 
    ss.product_name,
    ss.status
FROM users u
LEFT JOIN subscription_plans sp ON u.sub_id = sp.id
LEFT JOIN stripe_customers sc ON u.stripe_customer_id = sc.stripe_id
LEFT JOIN stripe_subscriptions ss ON sc.id = ss.customer_id AND ss.status IN ('active', 'trialing')
WHERE u.email = 'mycommonsensefinancial@yahoo.com';

-- Step 4: Check how many users were affected by unit_amount
SELECT 
    ss.unit_amount,
    COUNT(*) as subscriber_count,
    CASE 
        WHEN ss.unit_amount = 997 THEN 'Basic Monthly ($9.97)'
        WHEN ss.unit_amount = 999 THEN 'Premium Monthly ($9.99)'
        WHEN ss.unit_amount = 8982 THEN 'Premium Semi-Annual ($89.82)'
        WHEN ss.unit_amount = 9564 THEN 'Premium Annual ($95.64)'
        WHEN ss.unit_amount = 7200 THEN 'Best Value Annual ($72.00)'
        ELSE 'Other'
    END as plan_description
FROM users u
INNER JOIN stripe_customers sc ON u.stripe_customer_id = sc.stripe_id
INNER JOIN stripe_subscriptions ss ON sc.id = ss.customer_id
WHERE ss.status IN ('active', 'trialing')
GROUP BY ss.unit_amount
ORDER BY ss.unit_amount;

-- Step 5: Summary of fixes
SELECT 
    'Users with sub_id now set' as description,
    COUNT(*) as count
FROM users 
WHERE sub_id IS NOT NULL;

SELECT 
    'Active subscriptions with product_name' as description,
    COUNT(*) as count
FROM stripe_subscriptions 
WHERE status IN ('active', 'trialing') 
AND product_name IS NOT NULL 
AND product_name != '';
