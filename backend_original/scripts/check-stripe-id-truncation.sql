-- 🚨 URGENT: Check if database contains truncated Stripe IDs or full ones
-- This will determine if the issue is display-only or data corruption

-- 1. Check what's actually in the users table
SELECT 
    'USERS_TABLE_STRIPE_IDS' as check_type,
    u.id,
    u.email,
    u.stripe_customer_id,
    LENGTH(u.stripe_customer_id) as id_length,
    CASE 
        WHEN u.stripe_customer_id LIKE 'cus_%' THEN 'FULL_STRIPE_ID'
        WHEN u.stripe_customer_id LIKE '#%' AND LENGTH(u.stripe_customer_id) = 9 THEN 'TRUNCATED_WITH_HASH'
        WHEN LENGTH(u.stripe_customer_id) = 8 AND u.stripe_customer_id NOT LIKE 'cus_%' THEN 'TRUNCATED_NO_HASH'
        ELSE 'OTHER_FORMAT'
    END as id_format
FROM users u
WHERE u.stripe_customer_id IS NOT NULL
ORDER BY id_format, u.created_at DESC
LIMIT 50;

-- 2. Count by format type
SELECT 
    'ID_FORMAT_SUMMARY' as summary_type,
    CASE 
        WHEN stripe_customer_id LIKE 'cus_%' THEN 'FULL_STRIPE_ID'
        WHEN stripe_customer_id LIKE '#%' AND LENGTH(stripe_customer_id) = 9 THEN 'TRUNCATED_WITH_HASH'
        WHEN LENGTH(stripe_customer_id) = 8 AND stripe_customer_id NOT LIKE 'cus_%' THEN 'TRUNCATED_NO_HASH'
        ELSE 'OTHER_FORMAT'
    END as id_format,
    COUNT(*) as count,
    MIN(LENGTH(stripe_customer_id)) as min_length,
    MAX(LENGTH(stripe_customer_id)) as max_length,
    array_agg(DISTINCT stripe_customer_id) FILTER (WHERE stripe_customer_id IS NOT NULL) as sample_ids
FROM users
WHERE stripe_customer_id IS NOT NULL
GROUP BY id_format
ORDER BY count DESC;

-- 3. Check if we can find full Stripe IDs that match the truncated ones
SELECT 
    'TRUNCATION_MATCH_TEST' as test_type,
    u.stripe_customer_id as truncated_id,
    sc.stripe_id as full_stripe_id,
    CASE 
        WHEN sc.stripe_id LIKE '%' || RIGHT(u.stripe_customer_id, 8) THEN 'MATCH_FOUND'
        ELSE 'NO_MATCH'
    END as match_status
FROM users u
LEFT JOIN stripe_customers sc ON sc.stripe_id LIKE '%' || RIGHT(u.stripe_customer_id, 8)
WHERE u.stripe_customer_id LIKE '#%'
  AND LENGTH(u.stripe_customer_id) = 9
LIMIT 20;
