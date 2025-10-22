-- Migration: Fix Production Database Schema Issues
-- Description: Adds missing columns and fixes foreign key constraints for Stripe sync
-- Date: 2025-10-07
-- Issue: Production simple-sync failures due to schema mismatch

-- =====================================================
-- 1. FIX STRIPE_PRICES TABLE - ADD MISSING COLUMN
-- =====================================================

DO $$ 
BEGIN
    -- Check if stripe_product_id column exists in stripe_prices
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'stripe_prices' AND column_name = 'stripe_product_id'
    ) THEN
        -- Add the missing stripe_product_id column
        ALTER TABLE stripe_prices ADD COLUMN stripe_product_id VARCHAR(255);
        
        -- Add index for performance
        CREATE INDEX IF NOT EXISTS idx_stripe_prices_stripe_product_id ON stripe_prices(stripe_product_id);
        
        RAISE NOTICE 'Added stripe_product_id column to stripe_prices table';
    ELSE
        RAISE NOTICE 'stripe_product_id column already exists in stripe_prices table';
    END IF;
END $$;

-- =====================================================
-- 2. FIX FOREIGN KEY CONSTRAINTS
-- =====================================================

-- First, let's check what foreign key constraints exist
DO $$ 
DECLARE
    constraint_exists BOOLEAN;
BEGIN
    -- Check if fk_stripe_product_id constraint exists
    SELECT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_stripe_product_id' 
        AND table_name = 'stripe_subscriptions'
    ) INTO constraint_exists;
    
    IF constraint_exists THEN
        -- Drop the problematic foreign key constraint
        ALTER TABLE stripe_subscriptions DROP CONSTRAINT IF EXISTS fk_stripe_product_id;
        RAISE NOTICE 'Dropped existing fk_stripe_product_id constraint from stripe_subscriptions';
    END IF;
END $$;

-- =====================================================
-- 3. ENSURE PROPER TABLE STRUCTURE
-- =====================================================

-- Make sure stripe_subscriptions has the right columns
DO $$ 
BEGIN
    -- Add stripe_product_id column to stripe_subscriptions if missing
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'stripe_subscriptions' AND column_name = 'stripe_product_id'
    ) THEN
        ALTER TABLE stripe_subscriptions ADD COLUMN stripe_product_id VARCHAR(255);
        RAISE NOTICE 'Added stripe_product_id column to stripe_subscriptions table';
    END IF;
    
    -- Add product_name column to stripe_subscriptions if missing
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'stripe_subscriptions' AND column_name = 'product_name'
    ) THEN
        ALTER TABLE stripe_subscriptions ADD COLUMN product_name TEXT;
        RAISE NOTICE 'Added product_name column to stripe_subscriptions table';
    END IF;
END $$;

-- =====================================================
-- 4. CREATE PROPER FOREIGN KEY RELATIONSHIPS
-- =====================================================

-- Add foreign key constraint from stripe_prices to stripe_products (if both tables exist)
DO $$ 
BEGIN
    -- Check if both tables exist
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'stripe_prices') 
       AND EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'stripe_products') THEN
        
        -- Add foreign key constraint from stripe_prices.product_id to stripe_products.id
        -- (This uses internal database IDs, not Stripe IDs)
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.table_constraints 
            WHERE constraint_name = 'fk_stripe_prices_product_id' 
            AND table_name = 'stripe_prices'
        ) THEN
            ALTER TABLE stripe_prices 
            ADD CONSTRAINT fk_stripe_prices_product_id 
            FOREIGN KEY (product_id) REFERENCES stripe_products(id) ON DELETE CASCADE;
            
            RAISE NOTICE 'Added foreign key constraint fk_stripe_prices_product_id';
        END IF;
        
    END IF;
END $$;

-- =====================================================
-- 5. ADD INDEXES FOR PERFORMANCE
-- =====================================================

-- Add missing indexes
CREATE INDEX IF NOT EXISTS idx_stripe_subscriptions_stripe_product_id ON stripe_subscriptions(stripe_product_id);
CREATE INDEX IF NOT EXISTS idx_stripe_subscriptions_customer_id ON stripe_subscriptions(customer_id);
CREATE INDEX IF NOT EXISTS idx_stripe_subscriptions_price_id ON stripe_subscriptions(price_id);
CREATE INDEX IF NOT EXISTS idx_stripe_subscriptions_status ON stripe_subscriptions(status);

-- =====================================================
-- 6. CLEAN UP ORPHANED DATA (OPTIONAL - BE CAREFUL!)
-- =====================================================

-- Remove any subscriptions that reference non-existent products
-- UNCOMMENT ONLY IF YOU WANT TO CLEAN UP ORPHANED DATA
/*
DELETE FROM stripe_subscriptions 
WHERE stripe_product_id IS NOT NULL 
AND stripe_product_id NOT IN (
    SELECT stripe_id FROM stripe_products WHERE stripe_id IS NOT NULL
);
*/

-- =====================================================
-- 7. VERIFICATION QUERIES
-- =====================================================

-- Show table structures for verification
DO $$ 
BEGIN
    RAISE NOTICE '=== VERIFICATION ===';
    RAISE NOTICE 'Migration completed. Please verify the following:';
    RAISE NOTICE '1. stripe_prices table should have stripe_product_id column';
    RAISE NOTICE '2. stripe_subscriptions table should have stripe_product_id and product_name columns';
    RAISE NOTICE '3. Foreign key constraints should be properly set up';
    RAISE NOTICE '4. Indexes should be created for performance';
END $$;

-- Show column information for key tables
SELECT 
    table_name,
    column_name,
    data_type,
    is_nullable
FROM information_schema.columns 
WHERE table_name IN ('stripe_prices', 'stripe_subscriptions', 'stripe_products')
ORDER BY table_name, ordinal_position;
