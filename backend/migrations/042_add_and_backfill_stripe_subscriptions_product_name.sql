-- Migration: Add product_name column and backfill existing data
-- Version: 042
-- Date: 2024-12-05
-- Description: Adds product_name column to stripe_subscriptions and backfills from stripe_products

DO $$ 
BEGIN
    -- Step 1: Add product_name column if it doesn't exist
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'stripe_subscriptions' AND column_name = 'product_name'
    ) THEN
        ALTER TABLE stripe_subscriptions ADD COLUMN product_name TEXT;
        RAISE NOTICE 'Added product_name column to stripe_subscriptions table';
    END IF;
    
    -- Step 2: Backfill product_name from stripe_products for existing subscriptions
    -- This is the critical step that was missing from the original migration
    UPDATE stripe_subscriptions ss
    SET product_name = sp.name
    FROM stripe_products sp
    WHERE ss.stripe_product_id = sp.stripe_id
      AND (ss.product_name IS NULL OR ss.product_name = '')
      AND sp.name IS NOT NULL;
    
    -- Get count of updated records
    GET DIAGNOSTICS updated_count = ROW_COUNT;
    RAISE NOTICE 'Backfilled product_name for % existing subscriptions', updated_count;
    
    -- Step 3: Add index for better performance
    CREATE INDEX IF NOT EXISTS idx_stripe_subscriptions_product_name 
    ON stripe_subscriptions(product_name) WHERE product_name IS NOT NULL;
    
    -- Step 4: Add NOT NULL constraint with a default for future records
    -- (We won't make existing NULL values fail, but new ones should be populated)
    -- This is optional but recommended for data integrity
    
    RAISE NOTICE 'Migration 042 completed successfully';
    
EXCEPTION
    WHEN OTHERS THEN
        RAISE EXCEPTION 'Migration 042 failed: %', SQLERRM;
END $$;
