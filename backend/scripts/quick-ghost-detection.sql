-- 🕵️ QUICK GHOST DETECTION - Find customers like #J6lcu7t8
-- Run this query in your database admin tool (pgAdmin, etc.)

-- 1. Quick check for obvious ghost customers
SELECT 
    'OBVIOUS_GHOSTS' as detection_type,
    sc.id,
    sc.stripe_id,
    sc.email,
    sc.name,
    sc.created_at,
    CASE 
        WHEN sc.stripe_id LIKE '#%' THEN 'HASH_ID_GHOST (like #J6lcu7t8)'
        WHEN sc.stripe_id NOT LIKE 'cus_%' THEN 'INVALID_STRIPE_FORMAT'
        ELSE 'OTHER'
    END as ghost_reason
FROM stripe_customers sc
WHERE 
    sc.stripe_id LIKE '#%' 
    OR sc.stripe_id NOT LIKE 'cus_%'
ORDER BY sc.created_at DESC;

-- 2. Count the ghosts by type
SELECT 
    'GHOST_SUMMARY' as summary_type,
    COUNT(*) as total_potential_ghosts,
    COUNT(*) FILTER (WHERE stripe_id LIKE '#%') as hash_id_ghosts,
    COUNT(*) FILTER (WHERE stripe_id NOT LIKE 'cus_%' AND stripe_id NOT LIKE '#%') as invalid_format_ghosts
FROM stripe_customers
WHERE 
    stripe_id LIKE '#%' 
    OR stripe_id NOT LIKE 'cus_%';

-- 3. Find the specific Ebenezer customer you mentioned
SELECT 
    'EBENEZER_SEARCH' as search_type,
    sc.id,
    sc.stripe_id,
    sc.email,
    sc.name,
    sc.created_at,
    'FOUND_EBENEZER_GHOST' as status
FROM stripe_customers sc
WHERE 
    sc.email LIKE '%super.admin@bome.test%'
    OR sc.stripe_id = '#J6lcu7t8'
    OR sc.name ILIKE '%ebenezer%';

-- 4. Check related data for ghost customers
SELECT 
    'GHOST_IMPACT_ANALYSIS' as analysis_type,
    sc.stripe_id,
    sc.email,
    COUNT(DISTINCT ss.id) as subscription_count,
    COUNT(DISTINCT si.id) as invoice_count,
    'RELATED_DATA_FOUND' as impact_level
FROM stripe_customers sc
LEFT JOIN stripe_subscriptions ss ON sc.id = ss.customer_id
LEFT JOIN stripe_invoices si ON sc.id = si.customer_id
WHERE 
    sc.stripe_id LIKE '#%' 
    OR sc.stripe_id NOT LIKE 'cus_%'
GROUP BY sc.stripe_id, sc.email
ORDER BY subscription_count DESC, invoice_count DESC;
