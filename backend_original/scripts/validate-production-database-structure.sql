-- Comprehensive database structure validation for production
-- This checks for missing indexes and foreign keys that could affect query performance

-- ===== PART 1: CHECK EXISTING INDEXES =====
SELECT 'EXISTING INDEXES CHECK' as section;

-- Check critical indexes for the enhanced subscribers query
SELECT 
    schemaname, tablename, indexname, indexdef
FROM pg_indexes 
WHERE tablename IN ('users', 'subscription_plans', 'stripe_customers', 'stripe_subscriptions', 'stripe_products', 'stripe_prices')
ORDER BY tablename, indexname;

-- ===== PART 2: CHECK EXISTING FOREIGN KEYS =====
SELECT 'EXISTING FOREIGN KEYS CHECK' as section;

SELECT 
    tc.table_name, 
    tc.constraint_name, 
    tc.constraint_type,
    kcu.column_name,
    ccu.table_name AS foreign_table_name,
    ccu.column_name AS foreign_column_name 
FROM information_schema.table_constraints AS tc 
JOIN information_schema.key_column_usage AS kcu
    ON tc.constraint_name = kcu.constraint_name
    AND tc.table_schema = kcu.table_schema
JOIN information_schema.constraint_column_usage AS ccu
    ON ccu.constraint_name = tc.constraint_name
    AND ccu.table_schema = tc.table_schema
WHERE tc.constraint_type = 'FOREIGN KEY' 
    AND tc.table_name IN ('users', 'subscription_plans', 'stripe_customers', 'stripe_subscriptions', 'stripe_products', 'stripe_prices')
ORDER BY tc.table_name, tc.constraint_name;

-- ===== PART 3: MISSING INDEXES (CREATE IF NOT EXISTS) =====
SELECT 'CREATING MISSING INDEXES' as section;

-- Critical indexes for enhanced subscribers query performance
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_sub_id ON users(sub_id) WHERE sub_id IS NOT NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_stripe_customer_id ON users(stripe_customer_id) WHERE stripe_customer_id IS NOT NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_active ON users(id, is_active) WHERE is_active = true;

-- Subscription plans indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_subscription_plans_active ON subscription_plans(id, is_active, deleted_at) WHERE is_active = true AND deleted_at IS NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_subscription_plans_lookup ON subscription_plans(id) WHERE is_active = true AND deleted_at IS NULL;

-- Stripe customers indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_stripe_customers_stripe_id ON stripe_customers(stripe_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_stripe_customers_email ON stripe_customers(email) WHERE email IS NOT NULL;

-- Stripe subscriptions indexes (CRITICAL for performance)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_stripe_subscriptions_customer_status ON stripe_subscriptions(customer_id, status) WHERE status IN ('active', 'trialing');
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_stripe_subscriptions_active ON stripe_subscriptions(customer_id, status, current_period_end) WHERE status IN ('active', 'trialing');
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_stripe_subscriptions_price_quality ON stripe_subscriptions(customer_id, stripe_price_id, product_name, unit_amount, created_at) WHERE status IN ('active', 'trialing');

-- Stripe products indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_stripe_products_stripe_id ON stripe_products(stripe_id) WHERE active = true;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_stripe_products_active ON stripe_products(stripe_id, name, active) WHERE active = true;

-- Stripe prices indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_stripe_prices_stripe_id ON stripe_prices(stripe_id) WHERE active = true;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_stripe_prices_active ON stripe_prices(stripe_id, unit_amount, currency, recurring_interval) WHERE active = true;

-- ===== PART 4: MISSING FOREIGN KEYS (ADD IF NOT EXISTS) =====
SELECT 'CREATING MISSING FOREIGN KEYS' as section;

-- Add foreign keys with proper error handling
DO $$
BEGIN
    -- Users to subscription_plans
    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints 
                   WHERE constraint_name = 'fk_users_sub_id' AND table_name = 'users') THEN
        ALTER TABLE users ADD CONSTRAINT fk_users_sub_id 
        FOREIGN KEY (sub_id) REFERENCES subscription_plans(id) DEFERRABLE INITIALLY DEFERRED;
        RAISE NOTICE 'Added FK: users.sub_id -> subscription_plans.id';
    END IF;

    -- Stripe subscriptions to customers
    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints 
                   WHERE constraint_name = 'fk_stripe_subs_customer_id' AND table_name = 'stripe_subscriptions') THEN
        ALTER TABLE stripe_subscriptions ADD CONSTRAINT fk_stripe_subs_customer_id 
        FOREIGN KEY (customer_id) REFERENCES stripe_customers(id) ON DELETE CASCADE;
        RAISE NOTICE 'Added FK: stripe_subscriptions.customer_id -> stripe_customers.id';
    END IF;

    -- Stripe subscriptions to products (nullable, so only if value exists)
    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints 
                   WHERE constraint_name = 'fk_stripe_subs_product_id' AND table_name = 'stripe_subscriptions') THEN
        -- First clean up any orphaned references
        UPDATE stripe_subscriptions 
        SET stripe_product_id = NULL 
        WHERE stripe_product_id IS NOT NULL 
          AND stripe_product_id NOT IN (SELECT stripe_id FROM stripe_products WHERE stripe_id IS NOT NULL);
        
        ALTER TABLE stripe_subscriptions ADD CONSTRAINT fk_stripe_subs_product_id 
        FOREIGN KEY (stripe_product_id) REFERENCES stripe_products(stripe_id) DEFERRABLE INITIALLY DEFERRED;
        RAISE NOTICE 'Added FK: stripe_subscriptions.stripe_product_id -> stripe_products.stripe_id';
    END IF;

    -- Stripe subscriptions to prices (nullable, so only if value exists)
    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints 
                   WHERE constraint_name = 'fk_stripe_subs_price_id' AND table_name = 'stripe_subscriptions') THEN
        -- Clean up any remaining orphaned references (should be done already)
        UPDATE stripe_subscriptions 
        SET stripe_price_id = NULL 
        WHERE stripe_price_id IS NOT NULL 
          AND stripe_price_id NOT IN (SELECT stripe_id FROM stripe_prices WHERE stripe_id IS NOT NULL);
        
        ALTER TABLE stripe_subscriptions ADD CONSTRAINT fk_stripe_subs_price_id 
        FOREIGN KEY (stripe_price_id) REFERENCES stripe_prices(stripe_id) DEFERRABLE INITIALLY DEFERRED;
        RAISE NOTICE 'Added FK: stripe_subscriptions.stripe_price_id -> stripe_prices.stripe_id';
    END IF;

    -- Stripe prices to products
    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints 
                   WHERE constraint_name = 'fk_stripe_prices_product_id' AND table_name = 'stripe_prices') THEN
        ALTER TABLE stripe_prices ADD CONSTRAINT fk_stripe_prices_product_id 
        FOREIGN KEY (product_id) REFERENCES stripe_products(id) ON DELETE CASCADE;
        RAISE NOTICE 'Added FK: stripe_prices.product_id -> stripe_products.id';
    END IF;

EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'Error adding foreign key: %', SQLERRM;
END $$;

-- ===== PART 5: VALIDATION QUERIES =====
SELECT 'VALIDATION RESULTS' as section;

-- Check for orphaned data that might affect query performance
SELECT 'Orphaned Data Check' as check_type;

-- Users with invalid sub_id
SELECT 'users_invalid_sub_id' as issue, COUNT(*) as count
FROM users u
LEFT JOIN subscription_plans sp ON u.sub_id = sp.id
WHERE u.sub_id IS NOT NULL AND sp.id IS NULL;

-- Stripe subscriptions with missing customers (should be 0 after FK)
SELECT 'stripe_subs_missing_customers' as issue, COUNT(*) as count
FROM stripe_subscriptions ss
LEFT JOIN stripe_customers sc ON ss.customer_id = sc.id
WHERE sc.id IS NULL;

-- Stripe subscriptions with missing products (OK to have some)
SELECT 'stripe_subs_missing_products' as issue, COUNT(*) as count
FROM stripe_subscriptions ss
LEFT JOIN stripe_products sp ON ss.stripe_product_id = sp.stripe_id
WHERE ss.stripe_product_id IS NOT NULL AND sp.stripe_id IS NULL;

-- Stripe subscriptions with missing prices (OK to have some)
SELECT 'stripe_subs_missing_prices' as issue, COUNT(*) as count
FROM stripe_subscriptions ss
LEFT JOIN stripe_prices sp ON ss.stripe_price_id = sp.stripe_id
WHERE ss.stripe_price_id IS NOT NULL AND sp.stripe_id IS NULL;

-- ===== PART 6: PERFORMANCE TEST =====
SELECT 'PERFORMANCE TEST' as section;

-- Test the enhanced subscribers query performance for Adam
EXPLAIN (ANALYZE, BUFFERS) 
SELECT DISTINCT ON (u.id)
    u.id, u.email,
    COALESCE(
        sp.name,                    
        stripe_prod.name,           
        ss.product_name,            
        CASE 
            WHEN ss.status = 'active' THEN 'Active Subscription'
            WHEN ss.status = 'trialing' THEN 'Trial Subscription'
            ELSE 'Subscription'
        END
    ) as final_plan_name
FROM users u
LEFT JOIN subscription_plans sp ON u.sub_id = sp.id 
    AND sp.is_active = true 
    AND sp.deleted_at IS NULL
LEFT JOIN stripe_customers sc ON u.stripe_customer_id = sc.stripe_id
LEFT JOIN stripe_subscriptions ss ON sc.id = ss.customer_id 
    AND ss.status IN ('active', 'trialing')
    AND (ss.current_period_end IS NULL OR ss.current_period_end > NOW())
LEFT JOIN stripe_products stripe_prod ON ss.stripe_product_id = stripe_prod.stripe_id
WHERE u.email = 'mycommonsensefinancial@yahoo.com'
ORDER BY u.id, 
    sp.id DESC NULLS LAST,                    
    ss.stripe_price_id IS NOT NULL DESC,      
    ss.product_name IS NOT NULL DESC,         
    ss.unit_amount DESC NULLS LAST,           
    ss.created_at DESC NULLS LAST;
