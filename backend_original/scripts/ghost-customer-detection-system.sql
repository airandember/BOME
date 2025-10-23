-- 🕵️ GHOST CUSTOMER DETECTION SYSTEM
-- This creates infrastructure to detect and manage customers that exist locally but not in Stripe

-- 1. Create the "Not Found with Stripe" table
CREATE TABLE IF NOT EXISTS stripe_ghosts (
    id SERIAL PRIMARY KEY,
    local_customer_id INTEGER,
    stripe_customer_id VARCHAR(255),
    customer_email VARCHAR(255),
    customer_name VARCHAR(255),
    ghost_type VARCHAR(50), -- 'customer', 'subscription', 'product', etc.
    detection_date TIMESTAMP DEFAULT NOW(),
    last_seen_in_stripe TIMESTAMP,
    ghost_reason TEXT, -- 'not_found_in_api', 'test_data', 'deleted_account', etc.
    purge_status VARCHAR(50) DEFAULT 'detected', -- 'detected', 'marked_for_purge', 'purged'
    purged_at TIMESTAMP,
    purged_by VARCHAR(255),
    notes TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 2. Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_stripe_ghosts_customer_id ON stripe_ghosts(stripe_customer_id);
CREATE INDEX IF NOT EXISTS idx_stripe_ghosts_email ON stripe_ghosts(customer_email);
CREATE INDEX IF NOT EXISTS idx_stripe_ghosts_type ON stripe_ghosts(ghost_type);
CREATE INDEX IF NOT EXISTS idx_stripe_ghosts_status ON stripe_ghosts(purge_status);

-- 3. Create a function to detect ghost customers
CREATE OR REPLACE FUNCTION detect_ghost_customers() 
RETURNS TABLE(
    customer_id INTEGER,
    stripe_id VARCHAR(255),
    email VARCHAR(255),
    name VARCHAR(255),
    ghost_reason TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        sc.id as customer_id,
        sc.stripe_id,
        sc.email,
        sc.name,
        CASE 
            WHEN sc.stripe_id LIKE '#%' THEN 'legacy_test_customer'
            WHEN sc.stripe_id NOT LIKE 'cus_%' THEN 'invalid_stripe_id_format'
            WHEN sc.created_at < NOW() - INTERVAL '2 years' THEN 'very_old_customer'
            ELSE 'unknown_ghost_reason'
        END as ghost_reason
    FROM stripe_customers sc
    WHERE 
        -- Detect obvious ghost patterns
        (sc.stripe_id LIKE '#%' OR sc.stripe_id NOT LIKE 'cus_%')
        -- Add more detection logic as needed
    ORDER BY sc.created_at DESC;
END;
$$ LANGUAGE plpgsql;

-- 4. Create a function to mark customers for purging
CREATE OR REPLACE FUNCTION mark_customer_for_purge(customer_stripe_id VARCHAR(255), reason TEXT DEFAULT 'admin_decision')
RETURNS BOOLEAN AS $$
DECLARE
    customer_record RECORD;
BEGIN
    -- Get customer details
    SELECT * INTO customer_record 
    FROM stripe_customers 
    WHERE stripe_id = customer_stripe_id;
    
    IF NOT FOUND THEN
        RAISE NOTICE 'Customer % not found', customer_stripe_id;
        RETURN FALSE;
    END IF;
    
    -- Insert into ghosts table
    INSERT INTO stripe_ghosts (
        local_customer_id, stripe_customer_id, customer_email, 
        customer_name, ghost_type, ghost_reason, purge_status
    ) VALUES (
        customer_record.id, customer_record.stripe_id, customer_record.email,
        customer_record.name, 'customer', reason, 'marked_for_purge'
    ) ON CONFLICT (stripe_customer_id) 
    DO UPDATE SET 
        purge_status = 'marked_for_purge',
        ghost_reason = EXCLUDED.ghost_reason,
        updated_at = NOW();
    
    RAISE NOTICE 'Customer % marked for purge: %', customer_stripe_id, reason;
    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;

-- 5. Create cascading purge function
CREATE OR REPLACE FUNCTION purge_ghost_customer(customer_stripe_id VARCHAR(255), admin_user VARCHAR(255) DEFAULT 'system')
RETURNS BOOLEAN AS $$
DECLARE
    customer_id INTEGER;
    deleted_counts RECORD;
BEGIN
    -- Get the customer ID
    SELECT id INTO customer_id 
    FROM stripe_customers 
    WHERE stripe_id = customer_stripe_id;
    
    IF NOT FOUND THEN
        RAISE NOTICE 'Customer % not found for purging', customer_stripe_id;
        RETURN FALSE;
    END IF;
    
    -- Start transaction for cascading deletion
    BEGIN
        -- Delete from stripe_subscriptions
        DELETE FROM stripe_subscriptions WHERE customer_id = customer_id;
        GET DIAGNOSTICS deleted_counts.subscriptions = ROW_COUNT;
        
        -- Delete from stripe_invoices  
        DELETE FROM stripe_invoices WHERE customer_id = customer_id;
        GET DIAGNOSTICS deleted_counts.invoices = ROW_COUNT;
        
        -- Delete from any other related tables
        -- Add more as needed based on your schema
        
        -- Delete the customer record
        DELETE FROM stripe_customers WHERE id = customer_id;
        GET DIAGNOSTICS deleted_counts.customers = ROW_COUNT;
        
        -- Update ghost record
        UPDATE stripe_ghosts 
        SET 
            purge_status = 'purged',
            purged_at = NOW(),
            purged_by = admin_user,
            notes = format('Cascading deletion: %s subscriptions, %s invoices, %s customers', 
                          deleted_counts.subscriptions, deleted_counts.invoices, deleted_counts.customers)
        WHERE stripe_customer_id = customer_stripe_id;
        
        RAISE NOTICE 'Successfully purged customer %: % subscriptions, % invoices', 
                     customer_stripe_id, deleted_counts.subscriptions, deleted_counts.invoices;
        
        RETURN TRUE;
        
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Error purging customer %: %', customer_stripe_id, SQLERRM;
        RETURN FALSE;
    END;
END;
$$ LANGUAGE plpgsql;

-- 6. Create a view for easy ghost management
CREATE OR REPLACE VIEW v_ghost_customers AS
SELECT 
    sg.id as ghost_id,
    sg.stripe_customer_id,
    sg.customer_email,
    sg.customer_name,
    sg.ghost_type,
    sg.ghost_reason,
    sg.purge_status,
    sg.detection_date,
    sg.notes,
    -- Check if customer still exists in our database
    CASE WHEN sc.id IS NOT NULL THEN 'exists' ELSE 'already_deleted' END as current_status,
    -- Count related records
    COALESCE(sub_count.count, 0) as subscription_count,
    COALESCE(inv_count.count, 0) as invoice_count
FROM stripe_ghosts sg
LEFT JOIN stripe_customers sc ON sg.stripe_customer_id = sc.stripe_id
LEFT JOIN (
    SELECT customer_id, COUNT(*) as count 
    FROM stripe_subscriptions 
    GROUP BY customer_id
) sub_count ON sc.id = sub_count.customer_id
LEFT JOIN (
    SELECT customer_id, COUNT(*) as count 
    FROM stripe_invoices 
    GROUP BY customer_id
) inv_count ON sc.id = inv_count.customer_id
ORDER BY sg.detection_date DESC;

-- 7. Initial ghost detection - find the obvious ones
INSERT INTO stripe_ghosts (
    local_customer_id, stripe_customer_id, customer_email, 
    customer_name, ghost_type, ghost_reason, purge_status
)
SELECT 
    sc.id,
    sc.stripe_id,
    sc.email,
    sc.name,
    'customer',
    CASE 
        WHEN sc.stripe_id LIKE '#%' THEN 'legacy_test_customer_with_hash_id'
        WHEN sc.stripe_id NOT LIKE 'cus_%' THEN 'invalid_stripe_id_format'
        ELSE 'detected_ghost_customer'
    END,
    'detected'
FROM stripe_customers sc
WHERE 
    -- Detect obvious ghost patterns
    (sc.stripe_id LIKE '#%' OR sc.stripe_id NOT LIKE 'cus_%')
    -- Don't duplicate existing ghost records
    AND NOT EXISTS (
        SELECT 1 FROM stripe_ghosts sg 
        WHERE sg.stripe_customer_id = sc.stripe_id
    );

-- 8. Show immediate results
SELECT 
    'GHOST_DETECTION_SUMMARY' as summary_type,
    COUNT(*) as total_ghosts_detected,
    COUNT(*) FILTER (WHERE ghost_reason LIKE '%hash_id%') as hash_id_ghosts,
    COUNT(*) FILTER (WHERE ghost_reason LIKE '%invalid%') as invalid_format_ghosts
FROM stripe_ghosts
WHERE purge_status = 'detected';
