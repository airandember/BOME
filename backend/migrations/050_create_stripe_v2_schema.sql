-- ================================================================
-- STRIPE V2 SCHEMA - CLEAN BRAID ARCHITECTURE
-- ================================================================
-- Migration: 050
-- Description: Creates new Stripe tables with proper foreign keys,
--              audit trails, and performance optimizations
-- Strategy: Build parallel system, then swap
-- Date: 2025-10-26
-- ================================================================

-- ================================================================
-- TABLE 1: stripe_customers_v2
-- ================================================================
-- Stores Stripe customer data with audit trail
-- ================================================================

CREATE TABLE IF NOT EXISTS stripe_customers_v2 (
    id SERIAL PRIMARY KEY,
    stripe_id VARCHAR(255) UNIQUE NOT NULL,  -- cus_xxxxx
    email VARCHAR(512),
    name VARCHAR(255),
    phone VARCHAR(50),
    address JSONB,
    metadata JSONB,
    balance INTEGER DEFAULT 0,               -- cents
    currency VARCHAR(3),                     -- 'usd', 'eur', etc.
    delinquent BOOLEAN DEFAULT false,
    
    -- Stripe timestamps
    stripe_created_at TIMESTAMPTZ NOT NULL,
    
    -- Our timestamps
    first_synced_at TIMESTAMPTZ DEFAULT NOW(),
    last_synced_at TIMESTAMPTZ DEFAULT NOW(),
    sync_source VARCHAR(50)                  -- 'webhook', 'manual_sync', 'checkout'
);

-- Indexes for fast lookups
CREATE INDEX IF NOT EXISTS idx_stripe_customers_v2_email 
    ON stripe_customers_v2(email);
CREATE INDEX IF NOT EXISTS idx_stripe_customers_v2_stripe_created 
    ON stripe_customers_v2(stripe_created_at DESC);

COMMENT ON TABLE stripe_customers_v2 IS 'Stripe customers (v2) - synced from Stripe API with audit trail';
COMMENT ON COLUMN stripe_customers_v2.stripe_id IS 'Stripe customer ID (cus_xxxxx)';
COMMENT ON COLUMN stripe_customers_v2.sync_source IS 'Where this data came from: webhook, manual_sync, checkout';

-- ================================================================
-- TABLE 2: stripe_products_v2
-- ================================================================
-- Stores Stripe products with custom fields for video access
-- ================================================================

CREATE TABLE IF NOT EXISTS stripe_products_v2 (
    id SERIAL PRIMARY KEY,
    stripe_id VARCHAR(255) UNIQUE NOT NULL,  -- prod_xxxxx
    name VARCHAR(500) NOT NULL,
    description TEXT,
    active BOOLEAN DEFAULT true,
    metadata JSONB,
    
    -- Stripe timestamps
    stripe_created_at TIMESTAMPTZ NOT NULL,
    stripe_updated_at TIMESTAMPTZ,
    
    -- Our custom fields (NOT from Stripe)
    video_approved BOOLEAN DEFAULT false,    -- Custom: does this grant video access?
    is_legacy BOOLEAN DEFAULT false,         -- Custom: old/deprecated product
    
    -- Our timestamps
    first_synced_at TIMESTAMPTZ DEFAULT NOW(),
    last_synced_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_stripe_products_v2_active 
    ON stripe_products_v2(active) 
    WHERE active = true;
CREATE INDEX IF NOT EXISTS idx_stripe_products_v2_video_approved 
    ON stripe_products_v2(video_approved) 
    WHERE video_approved = true;

COMMENT ON TABLE stripe_products_v2 IS 'Stripe products (v2) - subscription plans and offerings';
COMMENT ON COLUMN stripe_products_v2.video_approved IS 'Custom field: does this product grant video access?';
COMMENT ON COLUMN stripe_products_v2.is_legacy IS 'Custom field: is this a deprecated/old product?';

-- ================================================================
-- TABLE 3: stripe_prices_v2
-- ================================================================
-- Stores Stripe prices with proper FK to products
-- ================================================================

CREATE TABLE IF NOT EXISTS stripe_prices_v2 (
    id SERIAL PRIMARY KEY,
    stripe_id VARCHAR(255) UNIQUE NOT NULL,  -- price_xxxxx
    
    -- ✅ PROPER FOREIGN KEY TO PRODUCTS
    product_id INTEGER NOT NULL REFERENCES stripe_products_v2(id) ON DELETE CASCADE,
    
    unit_amount BIGINT NOT NULL,             -- cents (can be 0 for free)
    currency VARCHAR(3) NOT NULL DEFAULT 'usd',
    active BOOLEAN DEFAULT true,
    
    -- Recurring details
    recurring_interval VARCHAR(20),          -- 'month', 'year', 'week', 'day'
    recurring_interval_count INTEGER DEFAULT 1,
    
    metadata JSONB,
    
    -- Stripe timestamps
    stripe_created_at TIMESTAMPTZ NOT NULL,
    
    -- Our timestamps
    first_synced_at TIMESTAMPTZ DEFAULT NOW(),
    last_synced_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for fast lookups
CREATE INDEX IF NOT EXISTS idx_stripe_prices_v2_product 
    ON stripe_prices_v2(product_id);
CREATE INDEX IF NOT EXISTS idx_stripe_prices_v2_active 
    ON stripe_prices_v2(active) 
    WHERE active = true;
CREATE INDEX IF NOT EXISTS idx_stripe_prices_v2_recurring 
    ON stripe_prices_v2(recurring_interval);

COMMENT ON TABLE stripe_prices_v2 IS 'Stripe prices (v2) - pricing for products';
COMMENT ON COLUMN stripe_prices_v2.product_id IS 'FK to stripe_products_v2.id (integer for fast joins)';
COMMENT ON COLUMN stripe_prices_v2.unit_amount IS 'Price in cents (divide by 100 for dollars)';

-- ================================================================
-- TABLE 4: stripe_subscriptions_v2
-- ================================================================
-- Stores Stripe subscriptions with proper FKs
-- ================================================================

CREATE TABLE IF NOT EXISTS stripe_subscriptions_v2 (
    id SERIAL PRIMARY KEY,
    stripe_id VARCHAR(255) UNIQUE NOT NULL,  -- sub_xxxxx
    
    -- ✅ PROPER FOREIGN KEYS
    customer_id INTEGER NOT NULL REFERENCES stripe_customers_v2(id) ON DELETE CASCADE,
    price_id INTEGER NOT NULL REFERENCES stripe_prices_v2(id) ON DELETE RESTRICT,
    
    status VARCHAR(50) NOT NULL,             -- 'active', 'canceled', 'past_due', etc.
    
    -- Billing periods
    current_period_start TIMESTAMPTZ NOT NULL,
    current_period_end TIMESTAMPTZ NOT NULL,
    cancel_at_period_end BOOLEAN DEFAULT false,
    canceled_at TIMESTAMPTZ,
    
    -- Stripe timestamps
    stripe_created_at TIMESTAMPTZ NOT NULL,
    
    -- Our timestamps
    first_synced_at TIMESTAMPTZ DEFAULT NOW(),
    last_synced_at TIMESTAMPTZ DEFAULT NOW(),
    
    metadata JSONB
);

-- Indexes for fast lookups
CREATE INDEX IF NOT EXISTS idx_stripe_subscriptions_v2_customer 
    ON stripe_subscriptions_v2(customer_id);
CREATE INDEX IF NOT EXISTS idx_stripe_subscriptions_v2_status 
    ON stripe_subscriptions_v2(status);
CREATE INDEX IF NOT EXISTS idx_stripe_subscriptions_v2_active 
    ON stripe_subscriptions_v2(customer_id, status) 
    WHERE status IN ('active', 'trialing');
CREATE INDEX IF NOT EXISTS idx_stripe_subscriptions_v2_period_end 
    ON stripe_subscriptions_v2(current_period_end);

COMMENT ON TABLE stripe_subscriptions_v2 IS 'Stripe subscriptions (v2) - user subscription records';
COMMENT ON COLUMN stripe_subscriptions_v2.customer_id IS 'FK to stripe_customers_v2.id';
COMMENT ON COLUMN stripe_subscriptions_v2.price_id IS 'FK to stripe_prices_v2.id';
COMMENT ON COLUMN stripe_subscriptions_v2.status IS 'active, canceled, past_due, trialing, etc.';

-- ================================================================
-- TABLE 5: user_stripe_customers_v2 (THE KEY INNOVATION)
-- ================================================================
-- Links users to Stripe customers with audit trail
-- Replaces the array field in users table
-- ================================================================

CREATE TABLE IF NOT EXISTS user_stripe_customers_v2 (
    id SERIAL PRIMARY KEY,
    
    -- ✅ EXPLICIT MANY-TO-MANY LINK
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    stripe_customer_id INTEGER NOT NULL REFERENCES stripe_customers_v2(id) ON DELETE CASCADE,
    
    -- ✅ PRIMARY CUSTOMER TRACKING
    is_primary BOOLEAN DEFAULT false,
    
    -- ✅ AUDIT TRAIL
    first_linked_at TIMESTAMPTZ DEFAULT NOW(),
    last_synced_at TIMESTAMPTZ DEFAULT NOW(),
    linked_by VARCHAR(50) DEFAULT 'manual_sync',  -- 'webhook', 'manual_sync', 'admin', 'checkout'
    linked_reason TEXT,                            -- 'new_subscription', 'email_match', 'manual_link', etc.
    
    -- Optional: why is this customer kept?
    notes TEXT,
    
    CONSTRAINT user_stripe_customers_v2_unique UNIQUE (user_id, stripe_customer_id)
);

-- ⚠️ CRITICAL: Ensure only one primary customer per user
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_stripe_customers_v2_primary 
    ON user_stripe_customers_v2(user_id) 
    WHERE is_primary = true;

-- Indexes for fast lookups
CREATE INDEX IF NOT EXISTS idx_user_stripe_customers_v2_user 
    ON user_stripe_customers_v2(user_id);
CREATE INDEX IF NOT EXISTS idx_user_stripe_customers_v2_customer 
    ON user_stripe_customers_v2(stripe_customer_id);

COMMENT ON TABLE user_stripe_customers_v2 IS 'Links users to Stripe customers (many-to-many with primary tracking)';
COMMENT ON COLUMN user_stripe_customers_v2.is_primary IS 'Only ONE primary customer per user (enforced by unique index)';
COMMENT ON COLUMN user_stripe_customers_v2.linked_by IS 'How was this link created: webhook, manual_sync, admin, checkout';
COMMENT ON COLUMN user_stripe_customers_v2.linked_reason IS 'Why was this link created: new_subscription, email_match, etc.';

-- ================================================================
-- SUCCESS MESSAGE
-- ================================================================

DO $$
BEGIN
    RAISE NOTICE '✅ Stripe v2 schema created successfully!';
    RAISE NOTICE '📊 Tables created:';
    RAISE NOTICE '   - stripe_customers_v2';
    RAISE NOTICE '   - stripe_products_v2';
    RAISE NOTICE '   - stripe_prices_v2';
    RAISE NOTICE '   - stripe_subscriptions_v2';
    RAISE NOTICE '   - user_stripe_customers_v2';
    RAISE NOTICE '🔗 All foreign keys and indexes created';
    RAISE NOTICE '🚀 Ready for Stripe sync!';
END $$;

