-- Fix missing product_name in stripe_subscriptions table
-- This addresses the root cause of Adam Arp's missing plan name

-- Step 1: Update existing subscriptions with missing product names
UPDATE stripe_subscriptions ss
SET product_name = sp.name
FROM stripe_products sp
WHERE ss.stripe_product_id = sp.stripe_id
  AND (ss.product_name IS NULL OR ss.product_name = '');

-- Step 2: Verify the fix
SELECT 
    ss.stripe_id,
    ss.status,
    ss.stripe_product_id,
    ss.product_name as current_product_name,
    sp.name as actual_product_name,
    CASE 
        WHEN ss.product_name = sp.name THEN '✅ FIXED'
        WHEN ss.product_name IS NULL OR ss.product_name = '' THEN '❌ STILL MISSING'
        ELSE '⚠️ MISMATCH'
    END as sync_status
FROM stripe_subscriptions ss
LEFT JOIN stripe_products sp ON ss.stripe_product_id = sp.stripe_id
WHERE ss.status IN ('active', 'trialing')
  AND ss.stripe_product_id IS NOT NULL
ORDER BY sync_status, ss.stripe_id;

-- Step 3: Show Adam Arp's corrected data
SELECT 
    u.email,
    ss.stripe_id,
    ss.product_name,
    sp.name as stripe_product_name
FROM users u
JOIN stripe_customers sc ON u.stripe_customer_id = sc.stripe_id
JOIN stripe_subscriptions ss ON sc.id = ss.customer_id
LEFT JOIN stripe_products sp ON ss.stripe_product_id = sp.stripe_id
WHERE u.email = 'mycommonsensefinancial@yahoo.com'
  AND ss.status IN ('active', 'trialing');
