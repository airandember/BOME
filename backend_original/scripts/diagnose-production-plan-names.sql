-- Diagnostic script to identify missing plan names in production
-- Run this in production database to diagnose the Adam Arp issue

-- 1. Check Adam Arp's user record
SELECT 'Adam Arp User Record' as check_type;
SELECT id, email, first_name, last_name, sub_id, stripe_customer_id, created_at
FROM users 
WHERE email = 'mycommonsensefinancial@yahoo.com';

-- 2. Check if his sub_id exists in subscription_plans
SELECT 'Subscription Plan Lookup' as check_type;
SELECT u.id as user_id, u.email, u.sub_id, sp.id as plan_id, sp.name as plan_name, sp.is_active, sp.deleted_at
FROM users u
LEFT JOIN subscription_plans sp ON u.sub_id = sp.id
WHERE u.email = 'mycommonsensefinancial@yahoo.com';

-- 3. Check his Stripe customer record
SELECT 'Stripe Customer Record' as check_type;
SELECT sc.id, sc.stripe_id, sc.email, sc.name, sc.created_at
FROM stripe_customers sc
JOIN users u ON (u.stripe_customer_id = sc.stripe_id)
WHERE u.email = 'mycommonsensefinancial@yahoo.com';

-- 4. Check his Stripe subscriptions
SELECT 'Stripe Subscriptions' as check_type;
SELECT ss.id, ss.stripe_id, ss.status, ss.stripe_product_id, ss.stripe_price_id, 
       ss.product_name, ss.current_period_start, ss.current_period_end, ss.created_at
FROM stripe_subscriptions ss
JOIN stripe_customers sc ON ss.customer_id = sc.id
JOIN users u ON u.stripe_customer_id = sc.stripe_id
WHERE u.email = 'mycommonsensefinancial@yahoo.com';

-- 5. Check Stripe products for his subscriptions
SELECT 'Stripe Products' as check_type;
SELECT sp.stripe_id, sp.name, sp.description, sp.active, sp.created_at
FROM stripe_products sp
JOIN stripe_subscriptions ss ON ss.stripe_product_id = sp.stripe_id
JOIN stripe_customers sc ON ss.customer_id = sc.id
JOIN users u ON u.stripe_customer_id = sc.stripe_id
WHERE u.email = 'mycommonsensefinancial@yahoo.com';

-- 6. Check Stripe prices for his subscriptions
SELECT 'Stripe Prices' as check_type;
SELECT sp.stripe_id, sp.unit_amount, sp.currency, sp.recurring_interval, sp.active, sp.created_at
FROM stripe_prices sp
JOIN stripe_subscriptions ss ON ss.stripe_price_id = sp.stripe_id
JOIN stripe_customers sc ON ss.customer_id = sc.id
JOIN users u ON u.stripe_customer_id = sc.stripe_id
WHERE u.email = 'mycommonsensefinancial@yahoo.com';

-- 7. Test the exact COALESCE logic from the query
SELECT 'COALESCE Logic Test' as check_type;
SELECT 
    u.id, u.email,
    u.sub_id,
    sp.id as legacy_plan_id,
    sp.name as legacy_plan_name,
    ss.stripe_product_id,
    stripe_prod.name as stripe_product_name,
    ss.product_name as fallback_product_name,
    COALESCE(
        sp.name,                    -- Legacy subscription plan name (highest priority)
        stripe_prod.name,           -- Current Stripe product name  
        ss.product_name,            -- Fallback from stripe_subscriptions
        CASE 
            WHEN ss.status = 'active' THEN 'Active Subscription'
            WHEN ss.status = 'trialing' THEN 'Trial Subscription'
            ELSE 'Subscription'
        END
    ) as final_plan_name
FROM users u
-- Join legacy subscription plans (for users with sub_id)
LEFT JOIN subscription_plans sp ON u.sub_id = sp.id 
    AND sp.is_active = true 
    AND sp.deleted_at IS NULL
-- Join Stripe customers (for users with stripe_customer_id)
LEFT JOIN stripe_customers sc ON (
    u.stripe_customer_id = sc.stripe_id OR 
    sc.stripe_id = ANY(COALESCE(u.stripe_customer_ids, '{}'))
)
-- Join current Stripe subscriptions (for ACTIVE Stripe subs only)
LEFT JOIN stripe_subscriptions ss ON sc.id = ss.customer_id 
    AND ss.status IN ('active', 'trialing')
    AND (ss.current_period_end IS NULL OR ss.current_period_end > NOW())
-- Join Stripe products (to get current product names)
LEFT JOIN stripe_products stripe_prod ON ss.stripe_product_id = stripe_prod.stripe_id
WHERE u.email = 'mycommonsensefinancial@yahoo.com';

-- 8. Check for missing foreign keys that could cause issues
SELECT 'Missing Foreign Key Check' as check_type;
-- Users with sub_id that don't exist in subscription_plans
SELECT 'Users with invalid sub_id' as issue, COUNT(*) as count
FROM users u
LEFT JOIN subscription_plans sp ON u.sub_id = sp.id
WHERE u.sub_id IS NOT NULL AND sp.id IS NULL;

-- Stripe subscriptions with missing product references
SELECT 'Stripe subs with missing products' as issue, COUNT(*) as count
FROM stripe_subscriptions ss
LEFT JOIN stripe_products sp ON ss.stripe_product_id = sp.stripe_id
WHERE ss.stripe_product_id IS NOT NULL AND sp.stripe_id IS NULL;

-- Stripe subscriptions with missing price references
SELECT 'Stripe subs with missing prices' as issue, COUNT(*) as count
FROM stripe_subscriptions ss
LEFT JOIN stripe_prices sp ON ss.stripe_price_id = sp.stripe_id
WHERE ss.stripe_price_id IS NOT NULL AND sp.stripe_id IS NULL;
