-- 🛡️ GHOST-PROOFING: Add constraints to prevent ghost re-entry
-- This will block ghost subscriptions from being re-inserted

-- OPTION 1: Add a check constraint to reject ghost product IDs
ALTER TABLE stripe_subscriptions 
ADD CONSTRAINT check_no_ghost_products 
CHECK (
    stripe_product_id IS NULL 
    OR stripe_product_id NOT IN (
        'prod_HjYKGcWGP9r4EC', 'prod_HEmcX1PE8TO2CO', 'prod_FvNAeI348dup9w',
        'prod_FvNAlEGGL452nN', 'prod_HF5YzcBH5Rwr0d', 'prod_GVV5efccnh13h9', 'prod_FvNAJgnw48hwpZ'
    )
);

-- OPTION 2: Create a whitelist of allowed product IDs (safer)
CREATE TABLE IF NOT EXISTS allowed_stripe_products (
    stripe_product_id VARCHAR(255) PRIMARY KEY,
    product_name VARCHAR(255),
    added_date TIMESTAMP DEFAULT NOW(),
    notes TEXT
);

-- Add your known good products to the whitelist
INSERT INTO allowed_stripe_products (stripe_product_id, product_name, notes) VALUES
('prod_RVBYX35dZw99f8', 'premium yearly', 'Current production product'),
('prod_KExafNlUlt1H8M', 'Combo Conference + Streaming Library​', 'Conference combo product'),
('prod_SrQV19Fb1GCWVz', 'STRIPE TEST REVENUE', 'Development test product')
ON CONFLICT (stripe_product_id) DO NOTHING;

-- Add constraint to only allow whitelisted products
ALTER TABLE stripe_subscriptions 
ADD CONSTRAINT check_only_allowed_products 
CHECK (
    stripe_product_id IS NULL 
    OR stripe_product_id IN (SELECT stripe_product_id FROM allowed_stripe_products)
);

-- OPTION 3: Add a ghost detection function
CREATE OR REPLACE FUNCTION is_ghost_product(product_id TEXT) 
RETURNS BOOLEAN AS $$
BEGIN
    -- Return TRUE if this is a known ghost product
    RETURN product_id IN (
        'prod_HjYKGcWGP9r4EC', 'prod_HEmcX1PE8TO2CO', 'prod_FvNAeI348dup9w',
        'prod_FvNAlEGGL452nN', 'prod_HF5YzcBH5Rwr0d', 'prod_GVV5efccnh13h9', 'prod_FvNAJgnw48hwpZ'
    );
END;
$$ LANGUAGE plpgsql;
