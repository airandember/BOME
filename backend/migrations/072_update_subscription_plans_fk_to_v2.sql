-- Migration 072: Remove subscription_plans FK constraints for Stripe columns
-- Purpose: Remove legacy FK constraints to allow importing from V2 tables
-- Date: 2025-12-25
-- Related: Stripe Products V2 Migration

-- ============================================================
-- PROBLEM: The subscription_plans table has FK constraints on:
--   - stripe_product_id -> stripe_products.stripe_id (legacy V1)
--   - stripe_price_id -> stripe_prices.stripe_id (legacy V1)
-- When importing from V2 tables, these fail if the product/price
-- doesn't exist in the legacy tables.
--
-- SOLUTION: Remove ALL Stripe-related FK constraints. We validate
-- data integrity at the application level.
-- ============================================================

-- Drop ALL FK constraints on Stripe-related columns
DO $$
DECLARE
    constraint_record RECORD;
    dropped_count INT := 0;
BEGIN
    -- Find and drop ALL FK constraints on stripe_product_id AND stripe_price_id
    FOR constraint_record IN
        SELECT DISTINCT tc.constraint_name
        FROM information_schema.table_constraints tc
        JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name
        WHERE tc.table_name = 'subscription_plans'
        AND tc.constraint_type = 'FOREIGN KEY'
        AND kcu.column_name IN ('stripe_product_id', 'stripe_price_id')
    LOOP
        EXECUTE 'ALTER TABLE subscription_plans DROP CONSTRAINT IF EXISTS ' || quote_ident(constraint_record.constraint_name);
        RAISE NOTICE '✅ Dropped FK constraint: %', constraint_record.constraint_name;
        dropped_count := dropped_count + 1;
    END LOOP;
    
    IF dropped_count = 0 THEN
        RAISE NOTICE 'ℹ️  No Stripe-related FK constraints found on subscription_plans';
    ELSE
        RAISE NOTICE '✅ Dropped % FK constraint(s) total', dropped_count;
    END IF;
END $$;

-- Update column comments to reflect the change
COMMENT ON COLUMN subscription_plans.stripe_product_id IS 
    'Stores Stripe product ID for reference (no FK constraint - validated at application level)';
COMMENT ON COLUMN subscription_plans.stripe_price_id IS 
    'Stores Stripe price ID for reference (no FK constraint - validated at application level)';

-- Verification
DO $$
DECLARE
    fk_count INT;
BEGIN
    SELECT COUNT(*) INTO fk_count
    FROM information_schema.table_constraints tc
    JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name
    WHERE tc.table_name = 'subscription_plans'
    AND tc.constraint_type = 'FOREIGN KEY'
    AND kcu.column_name IN ('stripe_product_id', 'stripe_price_id');
    
    RAISE NOTICE '';
    RAISE NOTICE '═══════════════════════════════════════════════════';
    RAISE NOTICE '✅ Migration 072: FK Constraints Removal Complete!';
    RAISE NOTICE '═══════════════════════════════════════════════════';
    RAISE NOTICE '';
    RAISE NOTICE '📊 Remaining Stripe-related FK constraints: %', fk_count;
    IF fk_count = 0 THEN
        RAISE NOTICE '✅ Both stripe_product_id and stripe_price_id are now simple VARCHAR columns';
        RAISE NOTICE '   Products can be imported from stripe_products_v2 without issues';
    END IF;
    RAISE NOTICE '';
END $$;

