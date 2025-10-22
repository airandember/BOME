-- Migration: Create database functions for ghost customer management
-- Purpose: Automated detection and management of invalid Stripe customer records
-- Created: October 21, 2025

-- Function 1: Detect Ghost Customers
-- Finds customers with invalid Stripe customer IDs (# prefix, invalid format, etc.)
CREATE OR REPLACE FUNCTION detect_ghost_customers()
RETURNS TABLE (
    customer_id INT,
    stripe_id VARCHAR,
    email VARCHAR,
    name VARCHAR,
    reason TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        u.id as customer_id,
        u.stripe_customer_id as stripe_id,
        u.email,
        CONCAT(COALESCE(u.first_name, ''), ' ', COALESCE(u.last_name, '')) as name,
        CASE 
            WHEN u.stripe_customer_id LIKE '#%' THEN 'hash_id: Invalid Stripe customer ID format (starts with #)'
            WHEN u.stripe_customer_id IS NOT NULL 
                 AND u.stripe_customer_id != '' 
                 AND NOT u.stripe_customer_id LIKE 'cus_%' THEN 'invalid_format: Stripe customer ID does not start with cus_'
            WHEN u.stripe_customer_id IS NULL AND u.sub_id IS NOT NULL THEN 'missing_stripe_id: User has subscription but no Stripe customer ID'
            ELSE 'unknown: Unable to determine ghost type'
        END as reason
    FROM users u
    WHERE 
        -- Case 1: Hash IDs (# prefix)
        (u.stripe_customer_id LIKE '#%')
        OR
        -- Case 2: Invalid format (not starting with cus_)
        (u.stripe_customer_id IS NOT NULL 
         AND u.stripe_customer_id != '' 
         AND NOT u.stripe_customer_id LIKE 'cus_%')
        OR
        -- Case 3: Missing Stripe ID but has subscription
        (u.stripe_customer_id IS NULL AND u.sub_id IS NOT NULL)
    ORDER BY u.created_at DESC;
END;
$$ LANGUAGE plpgsql;

-- Function 2: Mark Customer for Purge
-- Marks a ghost customer for purging with admin approval
CREATE OR REPLACE FUNCTION mark_customer_for_purge(
    p_stripe_customer_id VARCHAR,
    p_reason TEXT DEFAULT 'admin_decision'
)
RETURNS BOOLEAN AS $$
DECLARE
    v_rows_updated INT;
BEGIN
    -- Update the ghost record to mark for purge
    UPDATE stripe_ghosts
    SET 
        purge_status = 'marked_for_purge',
        notes = COALESCE(notes || E'\n', '') || 
                'Marked for purge: ' || p_reason || ' (at ' || NOW()::TEXT || ')',
        updated_at = NOW()
    WHERE stripe_customer_id = p_stripe_customer_id
      AND purge_status != 'purged';
    
    GET DIAGNOSTICS v_rows_updated = ROW_COUNT;
    
    -- If no rows updated, try to insert a new record
    IF v_rows_updated = 0 THEN
        INSERT INTO stripe_ghosts (
            stripe_customer_id,
            customer_email,
            customer_name,
            ghost_type,
            ghost_reason,
            purge_status,
            notes
        )
        SELECT 
            u.stripe_customer_id,
            u.email,
            CONCAT(COALESCE(u.first_name, ''), ' ', COALESCE(u.last_name, '')),
            'customer',
            p_reason,
            'marked_for_purge',
            'Marked for purge: ' || p_reason
        FROM users u
        WHERE u.stripe_customer_id = p_stripe_customer_id
        ON CONFLICT (stripe_customer_id) DO UPDATE
        SET 
            purge_status = 'marked_for_purge',
            notes = EXCLUDED.notes,
            updated_at = NOW();
        
        GET DIAGNOSTICS v_rows_updated = ROW_COUNT;
    END IF;
    
    RETURN v_rows_updated > 0;
END;
$$ LANGUAGE plpgsql;

-- Function 3: Purge Ghost Customer
-- Permanently removes a ghost customer and updates ghost tracking
CREATE OR REPLACE FUNCTION purge_ghost_customer(
    p_stripe_customer_id VARCHAR,
    p_admin_user VARCHAR DEFAULT 'system'
)
RETURNS BOOLEAN AS $$
DECLARE
    v_local_customer_id INT;
    v_rows_deleted INT := 0;
BEGIN
    -- Get local customer ID if exists
    SELECT id INTO v_local_customer_id
    FROM users
    WHERE stripe_customer_id = p_stripe_customer_id;
    
    -- Delete from stripe_customers (cascade will handle subscriptions, invoices)
    DELETE FROM stripe_customers
    WHERE stripe_id = p_stripe_customer_id;
    
    GET DIAGNOSTICS v_rows_deleted = ROW_COUNT;
    
    -- Clear the stripe_customer_id from users table if exists
    IF v_local_customer_id IS NOT NULL THEN
        UPDATE users
        SET 
            stripe_customer_id = NULL,
            updated_at = NOW()
        WHERE id = v_local_customer_id;
    END IF;
    
    -- Update ghost record to mark as purged
    UPDATE stripe_ghosts
    SET 
        purge_status = 'purged',
        purged_at = NOW(),
        purged_by = p_admin_user,
        notes = COALESCE(notes || E'\n', '') || 
                'Purged by ' || p_admin_user || ' at ' || NOW()::TEXT,
        updated_at = NOW()
    WHERE stripe_customer_id = p_stripe_customer_id;
    
    -- Return true if we deleted something or updated the ghost record
    RETURN v_rows_deleted > 0 OR FOUND;
END;
$$ LANGUAGE plpgsql;

-- Function 4: Get Ghost Customer Statistics
-- Returns summary statistics for ghost detection
CREATE OR REPLACE FUNCTION get_ghost_statistics()
RETURNS TABLE (
    total_ghosts BIGINT,
    hash_id_ghosts BIGINT,
    invalid_format_ghosts BIGINT,
    missing_stripe_id_ghosts BIGINT,
    marked_for_purge BIGINT,
    already_purged BIGINT,
    detected_today BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        COUNT(*)::BIGINT as total_ghosts,
        COUNT(*) FILTER (WHERE ghost_reason LIKE '%hash_id%')::BIGINT as hash_id_ghosts,
        COUNT(*) FILTER (WHERE ghost_reason LIKE '%invalid_format%')::BIGINT as invalid_format_ghosts,
        COUNT(*) FILTER (WHERE ghost_reason LIKE '%missing_stripe_id%')::BIGINT as missing_stripe_id_ghosts,
        COUNT(*) FILTER (WHERE purge_status = 'marked_for_purge')::BIGINT as marked_for_purge,
        COUNT(*) FILTER (WHERE purge_status = 'purged')::BIGINT as already_purged,
        COUNT(*) FILTER (WHERE detection_date::DATE = CURRENT_DATE)::BIGINT as detected_today
    FROM stripe_ghosts;
END;
$$ LANGUAGE plpgsql;

-- Grant execute permissions
GRANT EXECUTE ON FUNCTION detect_ghost_customers() TO PUBLIC;
GRANT EXECUTE ON FUNCTION mark_customer_for_purge(VARCHAR, TEXT) TO PUBLIC;
GRANT EXECUTE ON FUNCTION purge_ghost_customer(VARCHAR, VARCHAR) TO PUBLIC;
GRANT EXECUTE ON FUNCTION get_ghost_statistics() TO PUBLIC;

-- Add comments for documentation
COMMENT ON FUNCTION detect_ghost_customers() IS 'Detects users with invalid or problematic Stripe customer IDs';
COMMENT ON FUNCTION mark_customer_for_purge(VARCHAR, TEXT) IS 'Marks a ghost customer for administrative review and purging';
COMMENT ON FUNCTION purge_ghost_customer(VARCHAR, VARCHAR) IS 'Permanently deletes a ghost customer and all related Stripe data';
COMMENT ON FUNCTION get_ghost_statistics() IS 'Returns summary statistics for ghost customer detection';

-- Success message
DO $$
BEGIN
    RAISE NOTICE '✅ Ghost customer management functions created successfully';
    RAISE NOTICE '🔧 Functions: detect_ghost_customers, mark_customer_for_purge, purge_ghost_customer, get_ghost_statistics';
END $$;

