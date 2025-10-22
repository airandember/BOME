-- Migration: Create stripe_ghosts table for tracking data quality issues
-- Purpose: Identify and manage invalid/ghost customer records
-- Created: October 21, 2025

-- Create stripe_ghosts table
CREATE TABLE IF NOT EXISTS stripe_ghosts (
    id SERIAL PRIMARY KEY,
    local_customer_id INTEGER,
    stripe_customer_id VARCHAR(255) NOT NULL,
    customer_email VARCHAR(255),
    customer_name VARCHAR(255),
    ghost_type VARCHAR(50) NOT NULL DEFAULT 'customer',
    ghost_reason TEXT NOT NULL,
    purge_status VARCHAR(50) DEFAULT 'detected',
    detection_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    purged_at TIMESTAMP,
    purged_by VARCHAR(255),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_stripe_ghosts_user 
        FOREIGN KEY (local_customer_id) 
        REFERENCES users(id) 
        ON DELETE SET NULL
);

-- Create unique constraint on stripe_customer_id
CREATE UNIQUE INDEX IF NOT EXISTS stripe_ghosts_customer_id_key 
    ON stripe_ghosts(stripe_customer_id);

-- Create indexes for common queries
CREATE INDEX IF NOT EXISTS idx_stripe_ghosts_customer_id 
    ON stripe_ghosts(stripe_customer_id);

CREATE INDEX IF NOT EXISTS idx_stripe_ghosts_purge_status 
    ON stripe_ghosts(purge_status);

CREATE INDEX IF NOT EXISTS idx_stripe_ghosts_detection_date 
    ON stripe_ghosts(detection_date DESC);

CREATE INDEX IF NOT EXISTS idx_stripe_ghosts_local_customer 
    ON stripe_ghosts(local_customer_id) 
    WHERE local_customer_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_stripe_ghosts_ghost_type 
    ON stripe_ghosts(ghost_type);

-- Add comments for documentation
COMMENT ON TABLE stripe_ghosts IS 'Tracks invalid, duplicate, or problematic Stripe customer records for data quality monitoring';
COMMENT ON COLUMN stripe_ghosts.local_customer_id IS 'Reference to users table if ghost is linked to a local user';
COMMENT ON COLUMN stripe_ghosts.stripe_customer_id IS 'The problematic Stripe customer ID (e.g., #hash_id, invalid format, etc.)';
COMMENT ON COLUMN stripe_ghosts.ghost_type IS 'Type of ghost: customer, subscription, product, etc.';
COMMENT ON COLUMN stripe_ghosts.ghost_reason IS 'Explanation of why this is flagged as a ghost (e.g., hash_id format, not found in Stripe, duplicate)';
COMMENT ON COLUMN stripe_ghosts.purge_status IS 'Status: detected, marked_for_purge, purged, kept';
COMMENT ON COLUMN stripe_ghosts.detection_date IS 'When this ghost was first detected';
COMMENT ON COLUMN stripe_ghosts.purged_at IS 'When this ghost was purged (if applicable)';
COMMENT ON COLUMN stripe_ghosts.purged_by IS 'Admin user who performed the purge';

-- Success message
DO $$
BEGIN
    RAISE NOTICE '✅ stripe_ghosts table created successfully';
    RAISE NOTICE '📊 Indexes created for optimal query performance';
END $$;

