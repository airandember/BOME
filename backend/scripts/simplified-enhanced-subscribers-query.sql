-- SIMPLIFIED Enhanced Subscribers Query
-- Uses the EXISTING relationships: stripe_prices.product_id -> stripe_products.id

WITH active_stripe_subscriptions AS (
    -- Get all active Stripe subscriptions with plan details
    SELECT DISTINCT
        u.id as user_id,
        u.email,
        u.first_name,
        u.last_name,
        u.role,
        u.email_verified,
        u.stripe_customer_id,
        u.last_login,
        u.created_at,
        u.updated_at,
        
        -- Subscription details
        ss.stripe_id as stripe_subscription_id,
        ss.status as subscription_status,
        ss.current_period_start,
        ss.current_period_end,
        
        -- Plan details from Stripe products (via the existing FK relationship)
        prod.name as plan_name,
        CASE WHEN sp.unit_amount IS NOT NULL 
            THEN sp.unit_amount::float / 100.0 
            ELSE 0.0 
        END as plan_price,
        COALESCE(sp.currency, 'USD') as plan_currency,
        COALESCE(sp.recurring_interval, 'month') as plan_interval,
        1::INTEGER as plan_interval_count,
        
        -- Priority for ordering
        1::INTEGER as plan_priority
        
    FROM users u
    JOIN stripe_customers sc ON u.stripe_customer_id = sc.stripe_id
    JOIN stripe_subscriptions ss ON sc.id = ss.customer_id
    LEFT JOIN stripe_prices sp ON ss.stripe_price_id = sp.stripe_id
    LEFT JOIN stripe_products prod ON sp.product_id = prod.id  -- EXISTING FK RELATIONSHIP
    
    WHERE ss.status IN ('active', 'trialing')
      AND (ss.current_period_end IS NULL OR ss.current_period_end > NOW())
      AND u.is_active = true
),

legacy_subscriptions AS (
    -- Get legacy subscription plan users
    SELECT DISTINCT
        u.id as user_id,
        u.email,
        u.first_name,
        u.last_name,
        u.role,
        u.email_verified,
        u.stripe_customer_id,
        u.last_login,
        u.created_at,
        u.updated_at,
        
        -- Legacy subscription details (with proper types)
        NULL::VARCHAR as stripe_subscription_id,
        'active'::VARCHAR as subscription_status,
        NULL::TIMESTAMP WITH TIME ZONE as current_period_start,
        NULL::TIMESTAMP WITH TIME ZONE as current_period_end,
        
        -- Plan details from legacy subscription_plans (with proper types)
        sp.name as plan_name,
        sp.price::FLOAT as plan_price,
        sp.currency as plan_currency,
        sp.interval as plan_interval,
        sp.interval_count::INTEGER as plan_interval_count,
        
        -- Priority for ordering (legacy gets priority)
        0::INTEGER as plan_priority
        
    FROM users u
    JOIN subscription_plans sp ON u.sub_id = sp.id
    
    WHERE sp.is_active = true 
      AND sp.deleted_at IS NULL
      AND u.is_active = true
      -- Only include if user doesn't have active Stripe subscription
      AND NOT EXISTS (
          SELECT 1 FROM stripe_customers sc2
          JOIN stripe_subscriptions ss2 ON sc2.id = ss2.customer_id
          WHERE u.stripe_customer_id = sc2.stripe_id
            AND ss2.status IN ('active', 'trialing')
            AND (ss2.current_period_end IS NULL OR ss2.current_period_end > NOW())
      )
)

-- Final unified result
SELECT 
    user_id as id,
    email,
    first_name,
    last_name,
    role,
    email_verified,
    stripe_customer_id,
    last_login,
    created_at,
    updated_at,
    
    -- Plan information
    plan_name,
    plan_price,
    plan_currency,
    plan_interval,
    plan_interval_count,
    
    -- Subscription information
    subscription_status,
    current_period_start,
    current_period_end,
    stripe_subscription_id,
    
    -- For backward compatibility
    user_id as subscription_id,
    user_id as plan_id

FROM (
    SELECT * FROM active_stripe_subscriptions
    UNION ALL
    SELECT * FROM legacy_subscriptions
) combined

ORDER BY plan_priority ASC, user_id ASC
LIMIT 50;
