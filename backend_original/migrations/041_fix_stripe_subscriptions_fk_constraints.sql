-- Fix missing and incorrect FK constraints in stripe_subscriptions table
-- Production is missing the crucial price_id -> stripe_prices(id) relationship

-- Step 1: Drop the incorrectly named constraint that maps customer_id to stripe_customers
-- (This should be named stripe_subscriptions_customer_id_fkey, not fk_stripe_subscriptions_price_id)
DO $$
BEGIN
    -- Check if the misnamed constraint exists and drop it
    IF EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_stripe_subscriptions_price_id'
        AND table_name = 'stripe_subscriptions'
    ) THEN
        -- Check what this constraint actually references
        PERFORM 1 FROM information_schema.key_column_usage kcu
        JOIN information_schema.referential_constraints rc ON kcu.constraint_name = rc.constraint_name
        WHERE kcu.constraint_name = 'fk_stripe_subscriptions_price_id'
        AND kcu.table_name = 'stripe_subscriptions'
        AND kcu.column_name = 'customer_id';
        
        IF FOUND THEN
            -- This is the misnamed constraint - drop it
            ALTER TABLE stripe_subscriptions 
            DROP CONSTRAINT fk_stripe_subscriptions_price_id;
            
            RAISE NOTICE 'Dropped misnamed FK constraint that was mapping customer_id instead of price_id';
        END IF;
    END IF;
END $$;

-- Step 2: Add the correct FK constraint for price_id -> stripe_prices(id)
DO $$
BEGIN
    -- Add the missing FK constraint for price_id
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.key_column_usage kcu
        JOIN information_schema.referential_constraints rc ON kcu.constraint_name = rc.constraint_name
        WHERE kcu.table_name = 'stripe_subscriptions'
        AND kcu.column_name = 'price_id'
        AND kcu.referenced_table_name = 'stripe_prices'
        AND kcu.referenced_column_name = 'id'
    ) THEN
        -- Clean up any orphaned price_id references first
        UPDATE stripe_subscriptions 
        SET price_id = NULL
        WHERE price_id IS NOT NULL 
        AND price_id NOT IN (SELECT id FROM stripe_prices WHERE id IS NOT NULL);
        
        -- Add the correct FK constraint
        ALTER TABLE stripe_subscriptions 
        ADD CONSTRAINT fk_stripe_subscriptions_price_id_correct 
        FOREIGN KEY (price_id) REFERENCES stripe_prices(id);
        
        RAISE NOTICE 'Added correct FK constraint: stripe_subscriptions.price_id -> stripe_prices.id';
    ELSE
        RAISE NOTICE 'Correct FK constraint already exists: stripe_subscriptions.price_id -> stripe_prices.id';
    END IF;
END $$;

-- Step 3: Ensure the customer_id FK constraint exists with the correct name
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'stripe_subscriptions_customer_id_fkey'
        AND table_name = 'stripe_subscriptions'
    ) THEN
        -- Clean up any orphaned customer_id references first
        UPDATE stripe_subscriptions 
        SET customer_id = NULL
        WHERE customer_id IS NOT NULL 
        AND customer_id NOT IN (SELECT id FROM stripe_customers WHERE id IS NOT NULL);
        
        -- Add the customer_id FK constraint with correct name
        ALTER TABLE stripe_subscriptions 
        ADD CONSTRAINT stripe_subscriptions_customer_id_fkey 
        FOREIGN KEY (customer_id) REFERENCES stripe_customers(id);
        
        RAISE NOTICE 'Added customer FK constraint: stripe_subscriptions.customer_id -> stripe_customers.id';
    ELSE
        RAISE NOTICE 'Customer FK constraint already exists with correct name';
    END IF;
END $$;

-- Step 4: Add missing indexes for performance
CREATE INDEX IF NOT EXISTS idx_stripe_subscriptions_price_id 
ON stripe_subscriptions(price_id);

CREATE INDEX IF NOT EXISTS idx_stripe_subscriptions_stripe_price_id 
ON stripe_subscriptions(stripe_price_id);

CREATE INDEX IF NOT EXISTS idx_stripe_subscriptions_stripe_product_id 
ON stripe_subscriptions(stripe_product_id);

-- Step 5: Verify all FK constraints are now correct
SELECT 
    tc.constraint_name,
    kcu.column_name,
    ccu.table_name AS referenced_table,
    ccu.column_name AS referenced_column,
    tc.is_deferrable,
    rc.match_option,
    rc.update_rule,
    rc.delete_rule
FROM information_schema.table_constraints tc
JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name
JOIN information_schema.constraint_column_usage ccu ON ccu.constraint_name = tc.constraint_name
JOIN information_schema.referential_constraints rc ON tc.constraint_name = rc.constraint_name
WHERE tc.table_name = 'stripe_subscriptions' 
AND tc.constraint_type = 'FOREIGN KEY'
ORDER BY tc.constraint_name;

-- Step 6: Test the relationships work properly
SELECT 
    ss.stripe_id as subscription_id,
    ss.status,
    ss.price_id,
    spr.stripe_id as price_stripe_id,
    spr.unit_amount,
    ss.customer_id,
    sc.stripe_id as customer_stripe_id,
    sc.email
FROM stripe_subscriptions ss
LEFT JOIN stripe_prices spr ON ss.price_id = spr.id
LEFT JOIN stripe_customers sc ON ss.customer_id = sc.id
WHERE ss.status IN ('active', 'trialing')
LIMIT 5;
