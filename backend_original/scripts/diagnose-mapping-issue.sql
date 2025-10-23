-- DIAGNOSTIC: Why is the dynamic mapping not working?
-- Let's investigate step by step

-- Step 1: Check if we have any price mappings at all
WITH price_mapping AS (
    SELECT DISTINCT 
        sp.unit_amount,
        sp.currency,
        FIRST_VALUE(prod.name) OVER (
            PARTITION BY sp.unit_amount, sp.currency 
            ORDER BY sp.created_at DESC
        ) as plan_name,
        'stripe_prices' as source
    FROM stripe_prices sp
    JOIN stripe_products prod ON sp.product_id = prod.id
    WHERE sp.unit_amount IS NOT NULL 
      AND prod.name IS NOT NULL
      AND prod.active = true
    
    UNION ALL
    
    SELECT DISTINCT
        (legacy.price * 100)::integer as unit_amount,
        legacy.currency,
        legacy.name as plan_name,
        'legacy_plans' as source
    FROM subscription_plans legacy
    WHERE legacy.is_active = true 
      AND legacy.deleted_at IS NULL
      AND legacy.price IS NOT NULL
)

SELECT 'STEP 1: Available Price Mappings' as step;
SELECT * FROM price_mapping ORDER BY unit_amount;

-- Step 2: Check the COALESCE logic issue
SELECT 'STEP 2: COALESCE Logic Test' as step;
SELECT 
    ss.stripe_id,
    ss.unit_amount,
    ss.currency,
    ss.product_name,
    pm.plan_name as mapped_name,
    -- Test each step of COALESCE
    CASE WHEN ss.product_name IS NOT NULL AND ss.product_name != '' THEN ss.product_name ELSE NULL END as step1,
    CASE WHEN pm.plan_name IS NOT NULL THEN pm.plan_name ELSE NULL END as step2,
    'Subscription Plan (' || COALESCE(ss.currency, 'USD') || ' ' || 
    CASE WHEN ss.unit_amount IS NOT NULL 
        THEN (ss.unit_amount::float / 100.0)::text 
        ELSE '0' 
    END || ')' as step3,
    -- Final COALESCE result
    COALESCE(
        CASE WHEN ss.product_name IS NOT NULL AND ss.product_name != '' THEN ss.product_name ELSE NULL END,
        pm.plan_name,
        'Subscription Plan (' || COALESCE(ss.currency, 'USD') || ' ' || 
        CASE WHEN ss.unit_amount IS NOT NULL 
            THEN (ss.unit_amount::float / 100.0)::text 
            ELSE '0' 
        END || ')'
    ) as final_name
FROM stripe_subscriptions ss
LEFT JOIN price_mapping pm ON ss.unit_amount = pm.unit_amount 
    AND COALESCE(ss.currency, 'USD') = pm.currency
WHERE ss.status IN ('active', 'trialing')
  AND (ss.stripe_product_id IS NULL OR ss.stripe_product_id = '')
  AND ss.unit_amount IS NOT NULL
  AND ss.stripe_id IN (
    'sub_HjiahXik56JbZQ',  -- 4500 USD
    'sub_Fx1YapCgFisoYr',  -- 9564 USD  
    'sub_JGLu9d2drFcxw5'   -- 4500 USD (Alan Howard)
  )
ORDER BY ss.unit_amount;
