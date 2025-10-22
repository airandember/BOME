-- 🎯 BOME Video Access Fix
-- This script updates current subscription products to enable video access

-- First, let's see the current status
SELECT 
    id, 
    name, 
    video_approved,
    active,
    CASE 
        WHEN video_approved THEN '✅ Approved'
        ELSE '❌ Not Approved'
    END as status
FROM stripe_products 
WHERE active = true 
AND (
    name ILIKE '%plan%' 
    OR name ILIKE '%premium%'
    OR name ILIKE '%basic%'
)
ORDER BY id DESC;

-- Update the current subscription products that should have video access
UPDATE stripe_products 
SET video_approved = true 
WHERE id IN (
    993, -- Monthly Plan
    999, -- Annual Plan  
    13,  -- premium yearly
    14,  -- Premium Monthly
    15,  -- Premium Annual
    16,  -- Premium Yearly
    17   -- Premium Semi-Annual
)
AND active = true;

-- Show the results
SELECT 
    id, 
    name, 
    video_approved,
    active,
    CASE 
        WHEN video_approved THEN '✅ Approved'
        ELSE '❌ Not Approved'
    END as status
FROM stripe_products 
WHERE active = true 
AND (
    name ILIKE '%plan%' 
    OR name ILIKE '%premium%'
    OR name ILIKE '%basic%'
)
ORDER BY id DESC;

-- Check which users will be affected by this change
SELECT 
    u.id as user_id,
    u.email,
    u.manual_video_access,
    ss.status as subscription_status,
    sp.name as product_name,
    sp.video_approved
FROM users u
JOIN stripe_subscriptions ss ON u.stripe_customer_id = ss.customer_id
JOIN stripe_prices spr ON ss.price_id = spr.id  
JOIN stripe_products sp ON spr.product_id = sp.id
WHERE ss.status IN ('active', 'trialing')
AND (ss.current_period_end IS NULL OR ss.current_period_end > NOW())
AND sp.id IN (993, 999, 13, 14, 15, 16, 17)
ORDER BY u.email;
