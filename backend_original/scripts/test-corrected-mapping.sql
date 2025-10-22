-- CORRECTED: Test the dynamic price mapping with proper empty string handling

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
    JOIN stripe_products prod ON sp.product_id = prod.id
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

-- Test what subscriptions would get mapped with CORRECTED logic
SELECT 
    ss.stripe_id,
    ss.unit_amount,
    ss.currency,
    ss.product_name as current_product_name,
    pm.plan_name as mapped_plan_name,
    -- CORRECTED COALESCE: Check for both NULL and empty string
    COALESCE(
        CASE WHEN ss.product_name IS NOT NULL AND ss.product_name != '' THEN ss.product_name ELSE NULL END,
        pm.plan_name,  -- Dynamic lookup from price_mapping
        'Subscription Plan (' || COALESCE(ss.currency, 'USD') || ' ' || 
        CASE WHEN ss.unit_amount IS NOT NULL 
            THEN (ss.unit_amount::float / 100.0)::text 
            ELSE '0' 
        END || ')'
    ) as new_plan_name,
    -- CORRECTED name source logic
    CASE 
        WHEN ss.product_name IS NOT NULL AND ss.product_name != '' THEN 'EXISTING_NAME'
        WHEN pm.plan_name IS NOT NULL THEN 'DYNAMIC_MAPPING'
        ELSE 'FALLBACK_GENERATED'
    END as name_source
FROM stripe_subscriptions ss
LEFT JOIN price_mapping pm ON ss.unit_amount = pm.unit_amount 
    AND COALESCE(ss.currency, 'USD') = pm.currency
WHERE ss.status IN ('active', 'trialing')
  AND (ss.stripe_product_id IS NULL OR ss.stripe_product_id = '')
  AND ss.unit_amount IS NOT NULL
ORDER BY ss.unit_amount, ss.created_at
LIMIT 20;
