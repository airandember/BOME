-- 🚨 NUCLEAR IMPACT ASSESSMENT 🚨
-- Before going nuclear, let's see what we'd lose

-- CRITICAL: Are legacy customers still active in your business?
SELECT 
    'LEGACY_CUSTOMER_STATUS' as analysis_type,
    ss.stripe_price_id as legacy_price_id,
    COUNT(*) as total_subscriptions,
    COUNT(*) FILTER (WHERE ss.status IN ('active', 'trialing')) as currently_active,
    COUNT(*) FILTER (WHERE ss.status = 'canceled') as canceled,
    COUNT(*) FILTER (WHERE ss.current_period_end > NOW()) as still_valid_period,
    -- Check if these customers exist in users table
    COUNT(DISTINCT u.id) as linked_users,
    COUNT(DISTINCT u.id) FILTER (WHERE u.is_active = true) as active_users
FROM stripe_subscriptions ss
LEFT JOIN stripe_customers sc ON ss.customer_id = sc.id  
LEFT JOIN users u ON sc.stripe_id = u.stripe_customer_id
WHERE ss.stripe_price_id IS NOT NULL 
  AND ss.stripe_price_id NOT LIKE 'price_%'
GROUP BY ss.stripe_price_id
ORDER BY currently_active DESC, total_subscriptions DESC;

-- NUCLEAR SIMULATION: What would remain after deletion?
SELECT 
    'NUCLEAR_SIMULATION' as simulation_type,
    'BEFORE_NUCLEAR' as phase,
    COUNT(*) as subscription_count,
    COUNT(DISTINCT sc.stripe_id) as unique_customers,
    COUNT(DISTINCT u.id) as linked_users
FROM stripe_subscriptions ss
LEFT JOIN stripe_customers sc ON ss.customer_id = sc.id
LEFT JOIN users u ON sc.stripe_id = u.stripe_customer_id

UNION ALL

SELECT 
    'NUCLEAR_SIMULATION' as simulation_type,
    'AFTER_NUCLEAR' as phase,
    COUNT(*) as subscription_count,
    COUNT(DISTINCT sc.stripe_id) as unique_customers,
    COUNT(DISTINCT u.id) as linked_users
FROM stripe_subscriptions ss
LEFT JOIN stripe_customers sc ON ss.customer_id = sc.id
LEFT JOIN users u ON sc.stripe_id = u.stripe_customer_id
WHERE ss.stripe_price_id LIKE 'price_%';  -- Only real Stripe data would remain

-- CUSTOMER IMPACT: Who would lose their subscription history?
SELECT 
    'CUSTOMER_IMPACT' as impact_type,
    COUNT(DISTINCT u.email) as customers_losing_history,
    COUNT(DISTINCT u.id) FILTER (WHERE u.is_active = true) as active_customers_affected,
    COUNT(DISTINCT u.id) FILTER (WHERE u.last_login > NOW() - INTERVAL '30 days') as recently_active_affected
FROM stripe_subscriptions ss
JOIN stripe_customers sc ON ss.customer_id = sc.id
JOIN users u ON sc.stripe_id = u.stripe_customer_id
WHERE ss.stripe_price_id IS NOT NULL 
  AND ss.stripe_price_id NOT LIKE 'price_%';  -- Legacy data that would be deleted
