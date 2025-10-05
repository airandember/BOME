-- Unified Plans Query: Treating subscription_plans and stripe_products as one dataset
-- This approach creates a unified "plans" view that combines both legacy and Stripe plans

-- Step 1: Create a unified plans CTE
WITH unified_plans AS (
    -- Legacy subscription plans
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
),

-- Step 2: Create user-plan associations
user_plans AS (
    SELECT 
        u.id as user_id,
        u.email,
        u.first_name,
        u.last_name,
        u.role,
        u.email_verified,
        u.stripe_customer_id,
        u.last_login,
        u.created_at as user_created_at,
        u.updated_at as user_updated_at,
        
        -- Plan information from unified plans
        up.plan_source,
        up.plan_id,
        up.plan_name,
        up.plan_price,
        up.plan_currency,
        up.plan_interval,
        up.plan_interval_count,
        
        -- Subscription status information
        CASE 
            WHEN up.plan_source = 'legacy' THEN 'active'
            ELSE COALESCE(ss.status, 'active')
        END as subscription_status,
        
        -- Subscription periods (only for Stripe)
        ss.current_period_start,
        ss.current_period_end,
        ss.stripe_id as stripe_subscription_id,
        
        -- Priority for DISTINCT ON ordering
        CASE 
            WHEN up.plan_source = 'legacy' THEN 1  -- Legacy plans highest priority
            WHEN up.plan_source = 'stripe' AND ss.stripe_price_id IS NOT NULL THEN 2
            WHEN up.plan_source = 'stripe' AND up.plan_name IS NOT NULL THEN 3
            ELSE 4
        END as plan_priority
        
    FROM users u
    
    -- Join with unified plans via legacy subscription
    LEFT JOIN unified_plans up ON (
        (up.plan_source = 'legacy' AND u.sub_id::text = up.plan_id) OR
        (up.plan_source = 'stripe' AND EXISTS (
            SELECT 1 FROM stripe_customers sc 
            JOIN stripe_subscriptions ss2 ON sc.id = ss2.customer_id
            WHERE (u.stripe_customer_id = sc.stripe_id OR sc.stripe_id = ANY(COALESCE(u.stripe_customer_ids, '{}')))
              AND ss2.stripe_product_id = up.stripe_product_id
              AND ss2.status IN ('active', 'trialing')
              AND (ss2.current_period_end IS NULL OR ss2.current_period_end > NOW())
        ))
    )
    
    -- Join Stripe subscription details for Stripe plans
    LEFT JOIN stripe_customers sc ON (
        up.plan_source = 'stripe' AND 
        (u.stripe_customer_id = sc.stripe_id OR sc.stripe_id = ANY(COALESCE(u.stripe_customer_ids, '{}')))
    )
    LEFT JOIN stripe_subscriptions ss ON (
        up.plan_source = 'stripe' AND 
        sc.id = ss.customer_id AND 
        ss.stripe_product_id = up.stripe_product_id AND
        ss.status IN ('active', 'trialing') AND
        (ss.current_period_end IS NULL OR ss.current_period_end > NOW())
    )
    
    WHERE up.plan_id IS NOT NULL  -- Only users with plans
      AND u.is_active = true
)

-- Step 3: Final query with unified plan data
SELECT DISTINCT ON (user_id)
    user_id as id,
    email,
    first_name,
    last_name,
    role,
    email_verified,
    stripe_customer_id,
    last_login,
    user_created_at as created_at,
    user_updated_at as updated_at,
    
    -- Unified plan information
    plan_id as subscription_id,
    plan_id,
    plan_name,
    plan_price,
    plan_currency,
    plan_interval,
    plan_interval_count,
    subscription_status,
    current_period_start,
    current_period_end,
    stripe_subscription_id,
    
    -- Additional metadata
    plan_source  -- 'legacy' or 'stripe'
    
FROM user_plans
ORDER BY user_id, plan_priority ASC;  -- Priority ensures legacy plans come first
