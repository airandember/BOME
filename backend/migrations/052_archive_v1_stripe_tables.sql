-- Migration 052: Archive V1 Stripe Tables
-- Purpose: Rename old v1 Stripe tables with _deprecated suffix
-- Date: 2025-10-31
-- Phase: 11.1 - V1 Archival

-- ⚠️ IMPORTANT: This migration archives v1 tables. Only run after:
--    1. Phase 9 confirms 100% data integrity (✅ Complete)
--    2. All services using v2 tables (✅ Complete)
--    3. Frontend using v2 data (✅ Complete)
--    4. Webhooks writing to v2 (✅ Complete)

-- Archive v1 subscription table (if it exists)
DO $$ 
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.tables 
        WHERE table_name = 'user_subscriptions'
        AND table_schema = 'public'
    ) THEN
        ALTER TABLE user_subscriptions RENAME TO user_subscriptions_deprecated_v1;
        RAISE NOTICE '✅ Renamed user_subscriptions → user_subscriptions_deprecated_v1';
    ELSE
        RAISE NOTICE 'ℹ️  user_subscriptions table does not exist (may already be archived)';
    END IF;
END $$;

-- Archive old subscription_plans if they reference v1 Stripe IDs
-- (Keep this table active but mark old entries)
DO $$ 
BEGIN
    -- Add a column to track v1 vs v2 if it doesn't exist
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'subscription_plans' 
        AND column_name = 'is_v1_legacy'
    ) THEN
        ALTER TABLE subscription_plans ADD COLUMN is_v1_legacy BOOLEAN DEFAULT false;
        RAISE NOTICE '✅ Added is_v1_legacy column to subscription_plans';
    ELSE
        RAISE NOTICE 'ℹ️  is_v1_legacy column already exists in subscription_plans';
    END IF;
END $$;

-- Create a snapshot of the v1 table before archival (for audit purposes)
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.tables 
        WHERE table_name = 'user_subscriptions_deprecated_v1'
        AND table_schema = 'public'
    ) THEN
        -- Create audit record
        CREATE TABLE IF NOT EXISTS v1_archive_audit (
            id SERIAL PRIMARY KEY,
            table_name VARCHAR(255),
            archived_at TIMESTAMPTZ DEFAULT NOW(),
            row_count INTEGER,
            notes TEXT
        );
        
        -- Record the archival
        INSERT INTO v1_archive_audit (table_name, row_count, notes)
        SELECT 
            'user_subscriptions',
            COUNT(*),
            'Archived during Phase 11 - V2 cutover. All data migrated to v2 tables.'
        FROM user_subscriptions_deprecated_v1;
        
        RAISE NOTICE '✅ Created audit record for user_subscriptions archival';
    END IF;
END $$;

-- Add comments to archived tables
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.tables 
        WHERE table_name = 'user_subscriptions_deprecated_v1'
        AND table_schema = 'public'
    ) THEN
        COMMENT ON TABLE user_subscriptions_deprecated_v1 IS 
            'DEPRECATED: Legacy v1 subscription table. Archived on 2025-10-31 during Phase 11. ' ||
            'All data migrated to stripe_subscriptions_v2 and user_stripe_customers_v2. ' ||
            'DO NOT USE. Kept for audit purposes only. Can be dropped after 90 days.';
        RAISE NOTICE '✅ Added deprecation comment to user_subscriptions_deprecated_v1';
    END IF;
END $$;

-- Final verification
DO $$
DECLARE
    v1_exists BOOLEAN;
    v2_exists BOOLEAN;
    v1_deprecated_exists BOOLEAN;
BEGIN
    SELECT EXISTS (
        SELECT 1 FROM information_schema.tables 
        WHERE table_name = 'user_subscriptions'
    ) INTO v1_exists;
    
    SELECT EXISTS (
        SELECT 1 FROM information_schema.tables 
        WHERE table_name = 'stripe_subscriptions_v2'
    ) INTO v2_exists;
    
    SELECT EXISTS (
        SELECT 1 FROM information_schema.tables 
        WHERE table_name = 'user_subscriptions_deprecated_v1'
    ) INTO v1_deprecated_exists;
    
    RAISE NOTICE '';
    RAISE NOTICE '═══════════════════════════════════════════════════';
    RAISE NOTICE '✅ Migration 052: V1 Table Archival Complete!';
    RAISE NOTICE '═══════════════════════════════════════════════════';
    RAISE NOTICE '';
    RAISE NOTICE '📊 Table Status:';
    
    IF v1_deprecated_exists THEN
        RAISE NOTICE '   ✅ user_subscriptions_deprecated_v1: EXISTS (archived)';
    ELSE
        RAISE NOTICE '   ℹ️  user_subscriptions_deprecated_v1: NOT FOUND';
    END IF;
    
    IF v2_exists THEN
        RAISE NOTICE '   ✅ stripe_subscriptions_v2: EXISTS (active)';
    ELSE
        RAISE WARNING '   ⚠️  stripe_subscriptions_v2: NOT FOUND!';
    END IF;
    
    IF NOT v1_exists THEN
        RAISE NOTICE '   ✅ user_subscriptions: ARCHIVED (no longer exists)';
    ELSE
        RAISE WARNING '   ⚠️  user_subscriptions: STILL EXISTS!';
    END IF;
    
    RAISE NOTICE '';
    RAISE NOTICE '🎯 Next Steps:';
    RAISE NOTICE '   1. Update backend code to remove v1 references';
    RAISE NOTICE '   2. Remove deprecated v1 services and routes';
    RAISE NOTICE '   3. Monitor for 48 hours';
    RAISE NOTICE '   4. Drop archived tables after 90 days';
    RAISE NOTICE '';
END $$;

