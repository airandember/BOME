-- Fix stripe_prices table schema by adding missing stripe_product_id column
-- This column should reference the Stripe product ID that this price belongs to

-- 1. Add the stripe_product_id column (allowing NULL initially for existing data)
ALTER TABLE stripe_prices 
ADD COLUMN IF NOT EXISTS stripe_product_id VARCHAR(255);

-- 2. Create an index on the new column for better performance
CREATE INDEX IF NOT EXISTS idx_stripe_prices_stripe_product_id 
ON stripe_prices(stripe_product_id);

-- 3. Add foreign key constraint to ensure data integrity
-- (We'll add this after we populate the column with correct data)
-- ALTER TABLE stripe_prices 
-- ADD CONSTRAINT fk_stripe_prices_stripe_product_id 
-- FOREIGN KEY (stripe_product_id) REFERENCES stripe_products(stripe_id);

-- 4. Verify the schema change
SELECT column_name, data_type, is_nullable
FROM information_schema.columns 
WHERE table_name = 'stripe_prices' 
AND table_schema = 'public'
ORDER BY ordinal_position;
