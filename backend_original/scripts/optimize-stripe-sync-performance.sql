-- Stripe Sync Performance Optimization
-- Add missing indexes and optimize query patterns

-- 1. Composite indexes for common lookup patterns
CREATE INDEX IF NOT EXISTS idx_stripe_subscriptions_customer_status_active 
ON stripe_subscriptions(customer_id, status) 
WHERE status IN ('active', 'trialing');

CREATE INDEX IF NOT EXISTS idx_stripe_subscriptions_stripe_ids 
ON stripe_subscriptions(stripe_price_id, stripe_product_id) 
WHERE stripe_price_id IS NOT NULL;

-- 2. Partial indexes for webhook processing
CREATE INDEX IF NOT EXISTS idx_stripe_customers_webhook_lookup 
ON stripe_customers(stripe_id) 
INCLUDE (id, email, name);

CREATE INDEX IF NOT EXISTS idx_stripe_products_webhook_lookup 
ON stripe_products(stripe_id) 
INCLUDE (name, active);

CREATE INDEX IF NOT EXISTS idx_stripe_prices_webhook_lookup 
ON stripe_prices(stripe_id) 
INCLUDE (id, unit_amount, currency, product_id);

-- 3. Optimize the enhanced subscribers query with covering index
CREATE INDEX IF NOT EXISTS idx_users_subscriber_lookup 
ON users(stripe_customer_id, sub_id, is_active) 
INCLUDE (id, email, first_name, last_name, created_at);

-- 4. Add GIN index for JSONB metadata searches
CREATE INDEX IF NOT EXISTS idx_stripe_customers_metadata_gin 
ON stripe_customers USING GIN (metadata);

-- 5. Performance monitoring view
CREATE OR REPLACE VIEW v_stripe_sync_performance AS
SELECT 
    'stripe_customers' as table_name,
    COUNT(*) as total_records,
    COUNT(*) FILTER (WHERE updated_at > NOW() - INTERVAL '24 hours') as updated_today,
    MAX(updated_at) as last_updated
FROM stripe_customers
UNION ALL
SELECT 
    'stripe_subscriptions' as table_name,
    COUNT(*) as total_records,
    COUNT(*) FILTER (WHERE created_at > NOW() - INTERVAL '24 hours') as updated_today,
    MAX(created_at) as last_updated
FROM stripe_subscriptions
UNION ALL
SELECT 
    'stripe_products' as table_name,
    COUNT(*) as total_records,
    COUNT(*) FILTER (WHERE updated_at > NOW() - INTERVAL '24 hours') as updated_today,
    MAX(updated_at) as last_updated
FROM stripe_products
UNION ALL
SELECT 
    'stripe_prices' as table_name,
    COUNT(*) as total_records,
    COUNT(*) FILTER (WHERE created_at > NOW() - INTERVAL '24 hours') as updated_today,
    MAX(created_at) as last_updated
FROM stripe_prices;

-- 6. Query to identify sync inefficiencies
SELECT 
    ss.stripe_id,
    ss.status,
    ss.stripe_product_id,
    ss.product_name,
    sp.name as actual_product_name,
    CASE 
        WHEN ss.product_name IS NULL OR ss.product_name = '' THEN 'MISSING_PRODUCT_NAME'
        WHEN ss.stripe_product_id IS NULL THEN 'MISSING_PRODUCT_ID'
        WHEN sp.stripe_id IS NULL THEN 'ORPHANED_PRODUCT_REF'
        WHEN ss.product_name != sp.name THEN 'PRODUCT_NAME_MISMATCH'
        ELSE 'OK'
    END as sync_status
FROM stripe_subscriptions ss
LEFT JOIN stripe_products sp ON ss.stripe_product_id = sp.stripe_id
WHERE ss.status IN ('active', 'trialing')
ORDER BY sync_status, ss.created_at DESC;
