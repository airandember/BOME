-- Migration 050.1: Alter user_stripe_customers_v2 to add missing columns
-- Purpose: Add first_linked_at and last_synced_at columns to existing table
-- Date: 2025-10-30

-- Add first_linked_at column (replaces linked_at)
DO $$ 
BEGIN
    -- Check if first_linked_at exists
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'user_stripe_customers_v2' 
        AND column_name = 'first_linked_at'
    ) THEN
        -- If linked_at exists, rename it to first_linked_at
        IF EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_name = 'user_stripe_customers_v2' 
            AND column_name = 'linked_at'
        ) THEN
            ALTER TABLE user_stripe_customers_v2 RENAME COLUMN linked_at TO first_linked_at;
            RAISE NOTICE '✅ Renamed linked_at to first_linked_at';
        ELSE
            -- Otherwise, add the column
            ALTER TABLE user_stripe_customers_v2 ADD COLUMN first_linked_at TIMESTAMPTZ DEFAULT NOW();
            RAISE NOTICE '✅ Added first_linked_at column';
        END IF;
    ELSE
        RAISE NOTICE 'ℹ️  first_linked_at column already exists';
    END IF;
END $$;

-- Add last_synced_at column
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'user_stripe_customers_v2' 
        AND column_name = 'last_synced_at'
    ) THEN
        ALTER TABLE user_stripe_customers_v2 ADD COLUMN last_synced_at TIMESTAMPTZ DEFAULT NOW();
        RAISE NOTICE '✅ Added last_synced_at column';
    ELSE
        RAISE NOTICE 'ℹ️  last_synced_at column already exists';
    END IF;
END $$;

-- Update linked_by to have a default if it doesn't
DO $$ 
BEGIN
    ALTER TABLE user_stripe_customers_v2 ALTER COLUMN linked_by SET DEFAULT 'manual_sync';
    RAISE NOTICE '✅ Set default for linked_by';
EXCEPTION
    WHEN OTHERS THEN
        RAISE NOTICE 'ℹ️  Could not set default for linked_by (may not exist or already has default)';
END $$;

-- Verify the changes
DO $$
DECLARE
    has_first_linked BOOLEAN;
    has_last_synced BOOLEAN;
BEGIN
    SELECT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'user_stripe_customers_v2' 
        AND column_name = 'first_linked_at'
    ) INTO has_first_linked;
    
    SELECT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'user_stripe_customers_v2' 
        AND column_name = 'last_synced_at'
    ) INTO has_last_synced;
    
    IF has_first_linked AND has_last_synced THEN
        RAISE NOTICE '✅ Migration successful! user_stripe_customers_v2 table updated.';
        RAISE NOTICE '   - first_linked_at: ✓';
        RAISE NOTICE '   - last_synced_at: ✓';
    ELSE
        RAISE WARNING '⚠️  Migration incomplete!';
        IF NOT has_first_linked THEN
            RAISE WARNING '   - first_linked_at: ✗';
        END IF;
        IF NOT has_last_synced THEN
            RAISE WARNING '   - last_synced_at: ✗';
        END IF;
    END IF;
END $$;

