-- 🕵️ REAL GHOST DETECTION - Check the users table!
-- These hash IDs are probably in the main users table, not stripe_customers

-- 1. Find ALL hash ID customers in users table
SELECT 
    'USERS_TABLE_GHOSTS' as detection_type,
    u.id,
    u.stripe_customer_id,
    u.email,
    u.first_name,
    u.last_name,
    u.created_at,
    u.role,
    CASE 
        WHEN u.stripe_customer_id LIKE '#%' THEN 'HASH_ID_GHOST'
        WHEN u.stripe_customer_id NOT LIKE 'cus_%' AND u.stripe_customer_id IS NOT NULL THEN 'INVALID_STRIPE_FORMAT'
        ELSE 'POTENTIALLY_VALID'
    END as ghost_reason
FROM users u
WHERE 
    u.stripe_customer_id LIKE '#%' 
    OR (u.stripe_customer_id NOT LIKE 'cus_%' AND u.stripe_customer_id IS NOT NULL)
ORDER BY u.created_at DESC
LIMIT 50;

-- 2. Count the scope of the ghost problem
SELECT 
    'GHOST_CONTAMINATION_SCOPE' as analysis_type,
    COUNT(*) as total_users_with_stripe_ids,
    COUNT(*) FILTER (WHERE stripe_customer_id LIKE 'cus_%') as real_stripe_customers,
    COUNT(*) FILTER (WHERE stripe_customer_id LIKE '#%') as hash_id_ghosts,
    COUNT(*) FILTER (WHERE stripe_customer_id NOT LIKE 'cus_%' AND stripe_customer_id NOT LIKE '#%' AND stripe_customer_id IS NOT NULL) as other_invalid_format,
    COUNT(*) FILTER (WHERE stripe_customer_id IS NULL) as no_stripe_id
FROM users
WHERE stripe_customer_id IS NOT NULL;

-- 3. Find the specific customers you mentioned
SELECT 
    'SPECIFIC_GHOSTS' as search_type,
    u.id,
    u.stripe_customer_id,
    u.email,
    u.first_name,
    u.last_name,
    u.created_at,
    'FOUND_IN_USERS_TABLE' as location
FROM users u
WHERE 
    u.stripe_customer_id IN ('#4voKsKQ1', '#7oozIiWW', '#Nb8WZRCz', '#rLOwMt7I', '#GTt7BZQg')
    OR u.email = 'super.admin@bome.test'
    OR u.stripe_customer_id = '#J6lcu7t8';

-- 4. Check if these ghost customers have any stripe_customers entries
SELECT 
    'CROSS_TABLE_ANALYSIS' as analysis_type,
    u.stripe_customer_id as user_stripe_id,
    u.email,
    sc.stripe_id as stripe_customers_id,
    sc.email as stripe_customers_email,
    CASE 
        WHEN sc.stripe_id IS NOT NULL THEN 'EXISTS_IN_BOTH_TABLES'
        ELSE 'ONLY_IN_USERS_TABLE'
    END as table_status
FROM users u
LEFT JOIN stripe_customers sc ON u.stripe_customer_id = sc.stripe_id
WHERE u.stripe_customer_id LIKE '#%'
ORDER BY u.created_at DESC
LIMIT 20;

-- 5. Show the plan assignment problem
SELECT 
    'PLAN_ASSIGNMENT_ISSUE' as issue_type,
    u.id,
    u.stripe_customer_id,
    u.email,
    u.sub_id as legacy_subscription_id,
    sp.name as legacy_plan_name,
    'NO_STRIPE_PLAN_BECAUSE_GHOST_ID' as stripe_plan_status
FROM users u
LEFT JOIN subscription_plans sp ON u.sub_id = sp.id
WHERE u.stripe_customer_id LIKE '#%'
ORDER BY u.created_at DESC
LIMIT 10;
