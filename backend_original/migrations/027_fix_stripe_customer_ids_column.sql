-- Migration: Fix stripe_customer_ids column to handle NULL values properly
-- This migration ensures the stripe_customer_ids column exists and can handle NULL values

-- Add the stripe_customer_ids column if it doesn't exist
ALTER TABLE users ADD COLUMN IF NOT EXISTS stripe_customer_ids TEXT[];

-- Create an index for better performance on Stripe ID lookups
CREATE INDEX IF NOT EXISTS idx_users_stripe_customer_ids ON users USING GIN(stripe_customer_ids);

-- Update any existing users that have a stripe_customer_id but no array entry
UPDATE users 
SET stripe_customer_ids = ARRAY[stripe_customer_id] 
WHERE stripe_customer_id IS NOT NULL 
  AND stripe_customer_id != '' 
  AND (stripe_customer_ids IS NULL OR array_length(stripe_customer_ids, 1) IS NULL);
