-- Fix Foreign Key Constraint Violations in Production
-- This script cleans up orphaned data in stripe_subscriptions table

-- Step 1: Identify the problematic subscriptions
SELECT 
    'Subscriptions with invalid stripe_product_id:' as status,
    COUNT(*) as count
FROM stripe_subscriptions ss
LEFT JOIN stripe_products sp ON ss.stripe_product_id = sp.stripe_id
WHERE ss.stripe_product_id IS NOT NULL 
AND sp.stripe_id IS NULL;

-- Step 2: Show a sample of the problematic data (first 10 records)
SELECT 
    ss.id,
    ss.stripe_id,
    ss.stripe_product_id,
    'Missing product in stripe_products' as issue
FROM stripe_subscriptions ss
LEFT JOIN stripe_products sp ON ss.stripe_product_id = sp.stripe_id
WHERE ss.stripe_product_id IS NOT NULL 
AND sp.stripe_id IS NULL
LIMIT 10;

-- Step 3: Clean up orphaned data (set to NULL if no matching product exists)
-- This prevents FK violations by removing invalid references
UPDATE stripe_subscriptions 
SET stripe_product_id = NULL
WHERE stripe_product_id IS NOT NULL 
AND stripe_product_id NOT IN (
    SELECT stripe_id FROM stripe_products WHERE stripe_id IS NOT NULL
);

-- Step 4: Validate the existing constraint to make it enforceable
DO $$
BEGIN
    -- Try to validate the existing constraint
    BEGIN
        ALTER TABLE stripe_subscriptions 
        VALIDATE CONSTRAINT fk_stripe_product_id;
        
        RAISE NOTICE 'Successfully validated FK constraint: stripe_subscriptions.stripe_product_id -> stripe_products.stripe_id';
    EXCEPTION 
        WHEN OTHERS THEN
            RAISE NOTICE 'Could not validate constraint, attempting to recreate: %', SQLERRM;
            
            -- Drop and recreate the constraint if validation fails
            ALTER TABLE stripe_subscriptions 
            DROP CONSTRAINT IF EXISTS fk_stripe_product_id;
            
            ALTER TABLE stripe_subscriptions 
            ADD CONSTRAINT fk_stripe_product_id 
            FOREIGN KEY (stripe_product_id) REFERENCES stripe_products(stripe_id);
            
            RAISE NOTICE 'Recreated FK constraint: stripe_subscriptions.stripe_product_id -> stripe_products.stripe_id';
    END;
END $$;

-- Step 5: Verify the fix worked
SELECT 
    'After cleanup - Subscriptions with invalid stripe_product_id:' as status,
    COUNT(*) as count
FROM stripe_subscriptions ss
LEFT JOIN stripe_products sp ON ss.stripe_product_id = sp.stripe_id
WHERE ss.stripe_product_id IS NOT NULL 
AND sp.stripe_id IS NULL;

-- Step 6: Show constraint status
SELECT 
    tc.constraint_name,
    tc.table_name,
    tc.constraint_type,
    kcu.column_name,
    ccu.table_name AS referenced_table,
    ccu.column_name AS referenced_column
FROM information_schema.table_constraints tc
JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name
JOIN information_schema.constraint_column_usage ccu ON ccu.constraint_name = tc.constraint_name
WHERE tc.table_name = 'stripe_subscriptions' 
AND tc.constraint_type = 'FOREIGN KEY'
AND tc.constraint_name = 'fk_stripe_product_id';
