-- Migration: Add video_approved column to stripe_products table
-- This allows granular control over which Stripe products grant video access

-- Add the video_approved column with default false
ALTER TABLE stripe_products 
ADD COLUMN video_approved BOOLEAN DEFAULT FALSE;

-- Add a comment to document the purpose
COMMENT ON COLUMN stripe_products.video_approved IS 'Whether this product grants video access to users';

-- Create an index for performance when filtering by video_approved
CREATE INDEX idx_stripe_products_video_approved ON stripe_products(video_approved);

-- Set some reasonable defaults for existing products (you can adjust these)
-- Example: Enable video access for products with "premium", "pro", or "plus" in the name
UPDATE stripe_products 
SET video_approved = TRUE 
WHERE LOWER(name) LIKE '%premium%' 
   OR LOWER(name) LIKE '%pro%' 
   OR LOWER(name) LIKE '%plus%'
   OR LOWER(name) LIKE '%unlimited%'
   OR LOWER(name) LIKE '%full%';

-- Log the changes
DO $$
DECLARE
    approved_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO approved_count FROM stripe_products WHERE video_approved = TRUE;
    RAISE NOTICE 'Migration completed: % products now have video_approved = TRUE', approved_count;
END $$;
