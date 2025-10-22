-- TARGET TEST: Test the specific subscription from the logs
-- sub_I5zFYVVuO032AW with product prod_HjYKGcWGP9r4EC and unit_amount 4500

WITH price_mapping AS (
    SELECT DISTINCT 
        sp.unit_amount,
        sp.currency,
        FIRST_VALUE(prod.name) OVER (
            PARTITION BY sp.unit_amount, sp.currency 
            ORDER BY sp.created_at DESC
        ) as plan_name
    FROM stripe_prices sp
    JOIN stripe_products prod ON sp.product_id = prod.id
    WHERE sp.unit_amount IS NOT NULL 
      AND prod.name IS NOT NULL
      AND prod.active = true
    
    UNION ALL
    
    SELECT DISTINCT
        (legacy.price * 100)::integer as unit_amount,
        legacy.currency,
        legacy.name as plan_name
    FROM subscription_plans legacy
    WHERE legacy.is_active = true 
      AND legacy.deleted_at IS NULL
      AND legacy.price IS NOT NULL
)

-- Test the SPECIFIC subscription from the logs
SELECT 
    'TARGET_SUBSCRIPTION_TEST' as test_type,
    ss.stripe_id,
    ss.unit_amount,
    ss.currency,
    ss.stripe_product_id,
    ss.product_name as current_product_name,
    pm.plan_name as mapped_plan_name,
    -- Our corrected COALESCE logic
    COALESCE(
        CASE WHEN ss.product_name IS NOT NULL AND ss.product_name != '' THEN ss.product_name ELSE NULL END,
        pm.plan_name,
        'Subscription Plan (' || COALESCE(ss.currency, 'USD') || ' ' || 
        CASE WHEN ss.unit_amount IS NOT NULL 
            THEN (ss.unit_amount::float / 100.0)::text 
            ELSE '0' 
        END || ')'
    ) as final_plan_name,
    CASE 
        WHEN ss.product_name IS NOT NULL AND ss.product_name != '' THEN 'EXISTING_NAME'
        WHEN pm.plan_name IS NOT NULL THEN 'DYNAMIC_MAPPING'
        ELSE 'FALLBACK_GENERATED'
    END as name_source
FROM stripe_subscriptions ss
LEFT JOIN price_mapping pm ON ss.unit_amount = pm.unit_amount 
    AND COALESCE(ss.currency, 'USD') = pm.currency
WHERE ss.stripe_id = 'sub_I5zFYVVuO032AW'  -- The exact subscription from logs
   OR ss.unit_amount = 4500  -- All $45 subscriptions
ORDER BY ss.stripe_id;

-- Also check what's in stripe_products for that product
SELECT 
    'PRODUCT_CHECK' as test_type,
    stripe_id,
    name,
    active,
    CASE 
        WHEN name IS NULL THEN 'NULL_NAME'
        WHEN name = '' THEN 'EMPTY_NAME'
        ELSE 'HAS_NAME'
    END as name_status
FROM stripe_products 
WHERE stripe_id = 'prod_HjYKGcWGP9r4EC';
