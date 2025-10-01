-- Fix the existing NOT VALID foreign key constraint
-- The constraint exists but is not being enforced due to data violations

-- Step 1: Find and fix any orphaned data that violates the constraint
-- Show subscriptions with invalid stripe_product_id references
SELECT 
    ss.id,
    ss.stripe_id,
    ss.stripe_product_id,
    'Missing product' as issue
FROM stripe_subscriptions ss
LEFT JOIN stripe_products sp ON ss.stripe_product_id = sp.stripe_id
WHERE ss.stripe_product_id IS NOT NULL 
AND sp.stripe_id IS NULL
LIMIT 10;

-- Step 2: Clean up orphaned data (set to NULL if no matching product exists)
UPDATE stripe_subscriptions 
SET stripe_product_id = NULL
WHERE stripe_product_id IS NOT NULL 
AND stripe_product_id NOT IN (
    SELECT stripe_id FROM stripe_products WHERE stripe_id IS NOT NULL
);

-- Step 3: Now validate the existing constraint to make it enforceable
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

-- Step 4: Verify the constraint is now valid and working
SELECT 
    tc.constraint_name,
    tc.table_name,
    tc.constraint_type,
    cc.is_deferrable,
    cc.initially_deferred
FROM information_schema.table_constraints tc
JOIN information_schema.check_constraints cc ON tc.constraint_name = cc.constraint_name
WHERE tc.table_name = 'stripe_subscriptions' 
AND tc.constraint_name = 'fk_stripe_product_id';

-- Step 5: Test the relationship now works properly
SELECT 
    ss.stripe_id as subscription_id,
    ss.status,
    ss.stripe_product_id,
    sp.name as product_name,
    sp.active as product_active
FROM stripe_subscriptions ss
LEFT JOIN stripe_products sp ON ss.stripe_product_id = sp.stripe_id
WHERE ss.status IN ('active', 'trialing')
AND ss.stripe_product_id IS NOT NULL
LIMIT 5;
