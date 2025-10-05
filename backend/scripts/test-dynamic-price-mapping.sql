-- Test the dynamic price mapping query
-- This verifies that we're properly mapping unit_amount to plan names from existing data

WITH price_mapping AS (
    -- Create a mapping of unit_amount to plan names from existing data
    SELECT DISTINCT 
        sp.unit_amount,
        sp.currency,
        FIRST_VALUE(prod.name) OVER (
            PARTITION BY sp.unit_amount, sp.currency 
            ORDER BY sp.created_at DESC
        ) as plan_name
    FROM stripe_prices sp
    JOIN stripe_products prod ON sp.product_id = prod.stripe_id
    WHERE sp.unit_amount IS NOT NULL 
      AND prod.name IS NOT NULL
      AND prod.active = true
    
    UNION ALL
    
    -- Add legacy plan mappings (convert to cents for consistency)
    SELECT DISTINCT
        (legacy.price * 100)::integer as unit_amount,
        legacy.currency,
        legacy.name as plan_name
    FROM subscription_plans legacy
    WHERE legacy.is_active = true 
      AND legacy.deleted_at IS NULL
      AND legacy.price IS NOT NULL
)

-- Test the mapping for our known problem cases
SELECT 
    pm.unit_amount,
    pm.currency,
    pm.plan_name,
    'DYNAMIC_MAPPING' as source
FROM price_mapping pm
WHERE pm.unit_amount IN (4500, 9500, 8982, 15564)  -- Our known problem amounts
ORDER BY pm.unit_amount;

-- Also test what subscriptions would get mapped
SELECT 
    ss.stripe_id,
    ss.unit_amount,
    ss.currency,
    ss.product_name as current_product_name,
    COALESCE(
        ss.product_name,
        pm.plan_name,  -- Dynamic lookup from price_mapping
        'Subscription Plan (' || COALESCE(ss.currency, 'USD') || ' ' || 
        CASE WHEN ss.unit_amount IS NOT NULL 
            THEN (ss.unit_amount::float / 100.0)::text 
            ELSE '0' 
        END || ')'
    ) as new_plan_name,
    CASE 
        WHEN ss.product_name IS NOT NULL THEN 'EXISTING_NAME'
        WHEN pm.plan_name IS NOT NULL THEN 'DYNAMIC_MAPPING'
        ELSE 'FALLBACK_GENERATED'
    END as name_source
FROM stripe_subscriptions ss
LEFT JOIN (
    -- Inline the price_mapping logic for this test
    SELECT DISTINCT 
        sp.unit_amount,
        sp.currency,
        FIRST_VALUE(prod.name) OVER (
            PARTITION BY sp.unit_amount, sp.currency 
            ORDER BY sp.created_at DESC
        ) as plan_name
    FROM stripe_prices sp
    JOIN stripe_products prod ON sp.product_id = prod.stripe_id
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
) pm ON ss.unit_amount = pm.unit_amount 
    AND COALESCE(ss.currency, 'USD') = pm.currency
WHERE ss.status IN ('active', 'trialing')
  AND (ss.stripe_product_id IS NULL OR ss.stripe_product_id = '')
  AND ss.unit_amount IS NOT NULL
ORDER BY ss.unit_amount, ss.created_at;
