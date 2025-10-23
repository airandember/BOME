-- Test the enhanced unified query for the problem cases
-- This query should now handle incomplete Stripe data properly

WITH unified_plans AS (
    -- Legacy subscription plans (highest priority)
    SELECT 
        'legacy' as plan_source,
        sp.id::text as plan_id,
        sp.name as plan_name,
        sp.price as plan_price,
        sp.currency as plan_currency,
        sp.interval as plan_interval,
        sp.interval_count as plan_interval_count,
        sp.is_active as is_active,
        sp.created_at as created_at,
        sp.updated_at as updated_at,
        NULL as stripe_product_id,
        NULL as stripe_price_id
    FROM subscription_plans sp
    WHERE sp.is_active = true 
      AND sp.deleted_at IS NULL
    
    UNION ALL
    
    -- Stripe products as plans (with pricing from stripe_prices)
    SELECT 
        'stripe' as plan_source,
        stripe_prod.stripe_id as plan_id,
        stripe_prod.name as plan_name,
        CASE WHEN stripe_price.unit_amount IS NOT NULL 
            THEN stripe_price.unit_amount::float / 100.0 
            ELSE 0.0 
        END as plan_price,
        COALESCE(stripe_price.currency, 'USD') as plan_currency,
        COALESCE(stripe_price.recurring_interval, 'month') as plan_interval,
        COALESCE(stripe_price.recurring_interval_count, 1) as plan_interval_count,
        stripe_prod.active as is_active,
        stripe_prod.created_at as created_at,
        stripe_prod.updated_at as updated_at,
        stripe_prod.stripe_id as stripe_product_id,
        stripe_price.stripe_id as stripe_price_id
    FROM stripe_products stripe_prod
    LEFT JOIN stripe_prices stripe_price ON stripe_prod.stripe_id = stripe_price.product_id
    WHERE stripe_prod.active = true
    
    UNION ALL
    
    -- Fallback plans for Stripe subscriptions without product_id (incomplete data)
    SELECT 
        'stripe_fallback' as plan_source,
        ss.stripe_id as plan_id,
        COALESCE(
            ss.product_name,
            CASE 
                WHEN ss.unit_amount = 4500 THEN 'Basic Monthly'
                WHEN ss.unit_amount = 9500 THEN 'Premium Monthly'
                WHEN ss.unit_amount = 8982 THEN 'Premium Semi-Annual'
                WHEN ss.unit_amount = 15564 THEN 'Premium Yearly'
                ELSE 'Subscription Plan'
            END
        ) as plan_name,
        CASE WHEN ss.unit_amount IS NOT NULL 
            THEN ss.unit_amount::float / 100.0 
            ELSE 0.0 
        END as plan_price,
        COALESCE(ss.currency, 'USD') as plan_currency,
        'month' as plan_interval,
        1 as plan_interval_count,
        true as is_active,
        ss.created_at as created_at,
        NOW() as updated_at,
        ss.stripe_product_id as stripe_product_id,
        ss.stripe_price_id as stripe_price_id
    FROM stripe_subscriptions ss
    WHERE ss.status IN ('active', 'trialing')
      AND (ss.current_period_end IS NULL OR ss.current_period_end > NOW())
      AND (ss.stripe_product_id IS NULL OR ss.stripe_product_id = '')
      AND ss.unit_amount IS NOT NULL
)

-- Test query for the problem users
SELECT 
    u.email,
    up.plan_source,
    up.plan_name,
    up.plan_price,
    up.plan_currency,
    'SHOULD_FIX' as status
FROM users u
JOIN unified_plans up ON (
    -- Legacy subscription match
    (up.plan_source = 'legacy' AND u.sub_id::text = up.plan_id) 
    OR
    -- Stripe product match
    (up.plan_source = 'stripe' AND EXISTS (
        SELECT 1 FROM stripe_customers sc2 
        JOIN stripe_subscriptions ss2 ON sc2.id = ss2.customer_id
        WHERE (u.stripe_customer_id = sc2.stripe_id OR sc2.stripe_id = ANY(COALESCE(u.stripe_customer_ids, '{}')))
          AND ss2.stripe_product_id = up.stripe_product_id
          AND ss2.status IN ('active', 'trialing')
          AND (ss2.current_period_end IS NULL OR ss2.current_period_end > NOW())
    ))
    OR
    -- Stripe fallback match (for subscriptions without product_id)
    (up.plan_source = 'stripe_fallback' AND EXISTS (
        SELECT 1 FROM stripe_customers sc3 
        JOIN stripe_subscriptions ss3 ON sc3.id = ss3.customer_id
        WHERE (u.stripe_customer_id = sc3.stripe_id OR sc3.stripe_id = ANY(COALESCE(u.stripe_customer_ids, '{}')))
          AND ss3.stripe_id = up.plan_id
          AND ss3.status IN ('active', 'trialing')
          AND (ss3.current_period_end IS NULL OR ss3.current_period_end > NOW())
    ))
)
WHERE u.email IN ('kainthevamp@hotmail.com', 'tooldaddy@comcast.net', 'hoogyfrom@gmail.com')
ORDER BY u.email, 
    CASE 
        WHEN up.plan_source = 'legacy' THEN 1
        WHEN up.plan_source = 'stripe' THEN 2
        WHEN up.plan_source = 'stripe_fallback' THEN 3
        ELSE 4
    END;
