-- Migration: Comprehensive Stripe Subscriptions Product Name Fix
-- Version: 042
-- Date: 2024-12-05
-- Description: Ensures product_name column exists and is properly populated

DO $$ 
DECLARE
    updated_count INTEGER;
    missing_count INTEGER;
BEGIN
    RAISE NOTICE 'Starting comprehensive stripe_subscriptions product_name migration';
    
    -- Step 1: Add product_name column if it doesn't exist
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'stripe_subscriptions' AND column_name = 'product_name'
    ) THEN
        ALTER TABLE stripe_subscriptions ADD COLUMN product_name TEXT;
        RAISE NOTICE '✅ Added product_name column to stripe_subscriptions table';
    ELSE
        RAISE NOTICE '✅ product_name column already exists';
    END IF;
    
    -- Step 2: Count how many subscriptions need backfilling
    SELECT COUNT(*) INTO missing_count
    FROM stripe_subscriptions ss
    JOIN stripe_products sp ON ss.stripe_product_id = sp.stripe_id
    WHERE (ss.product_name IS NULL OR ss.product_name = '')
      AND sp.name IS NOT NULL;
    
    RAISE NOTICE '📊 Found % subscriptions needing product_name backfill', missing_count;
    
    -- Step 3: Backfill product_name from stripe_products for existing subscriptions
    UPDATE stripe_subscriptions ss
    SET product_name = sp.name
    FROM stripe_products sp
    WHERE ss.stripe_product_id = sp.stripe_id
      AND (ss.product_name IS NULL OR ss.product_name = '')
      AND sp.name IS NOT NULL;
    
    GET DIAGNOSTICS updated_count = ROW_COUNT;
    RAISE NOTICE '✅ Backfilled product_name for % subscriptions', updated_count;
    
    -- Step 4: Report any subscriptions that still have missing product names
    SELECT COUNT(*) INTO missing_count
    FROM stripe_subscriptions ss
    WHERE ss.status IN ('active', 'trialing')
      AND (ss.product_name IS NULL OR ss.product_name = '')
      AND ss.stripe_product_id IS NOT NULL;
    
    IF missing_count > 0 THEN
        RAISE NOTICE '⚠️ Warning: % active subscriptions still have missing product names', missing_count;
        RAISE NOTICE 'These may have orphaned stripe_product_id references';
    ELSE
        RAISE NOTICE '✅ All active subscriptions now have product names';
    END IF;
    
    -- Step 5: Add performance index
    CREATE INDEX IF NOT EXISTS idx_stripe_subscriptions_product_name 
    ON stripe_subscriptions(product_name) WHERE product_name IS NOT NULL;
    
    -- Step 6: Add composite index for common query patterns
    CREATE INDEX IF NOT EXISTS idx_stripe_subscriptions_status_product 
    ON stripe_subscriptions(status, stripe_product_id, product_name) 
    WHERE status IN ('active', 'trialing');
    
    RAISE NOTICE '✅ Added performance indexes';
    
    -- Step 7: Verify the fix with a sample query
    RAISE NOTICE '📋 Sample of fixed subscriptions:';
    FOR rec IN 
        SELECT ss.stripe_id, ss.status, ss.product_name, sp.name as source_name
        FROM stripe_subscriptions ss
        JOIN stripe_products sp ON ss.stripe_product_id = sp.stripe_id
        WHERE ss.status IN ('active', 'trialing')
        AND ss.product_name IS NOT NULL
        LIMIT 5
    LOOP
        RAISE NOTICE '  - %: % (from %)', rec.stripe_id, rec.product_name, rec.source_name;
    END LOOP;
    
    RAISE NOTICE '🎉 Migration 042 completed successfully!';
    
EXCEPTION
    WHEN OTHERS THEN
        RAISE EXCEPTION '❌ Migration 042 failed: %', SQLERRM;
END $$;
