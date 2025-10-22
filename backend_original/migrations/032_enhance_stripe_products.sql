-- Migration: Enhance stripe_products table with additional Stripe fields
-- This adds fields that Stripe provides but we're not currently storing

-- Add new columns to stripe_products
ALTER TABLE stripe_products 
ADD COLUMN IF NOT EXISTS metadata JSONB DEFAULT '{}',
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
ADD COLUMN IF NOT EXISTS url TEXT,
ADD COLUMN IF NOT EXISTS images TEXT[],
ADD COLUMN IF NOT EXISTS package_dimensions JSONB,
ADD COLUMN IF NOT EXISTS shippable BOOLEAN DEFAULT false,
ADD COLUMN IF NOT EXISTS statement_descriptor TEXT,
ADD COLUMN IF NOT EXISTS tax_code TEXT,
ADD COLUMN IF NOT EXISTS unit_label TEXT,
ADD COLUMN IF NOT EXISTS livemode BOOLEAN DEFAULT false;

-- Add indexes for better performance
CREATE INDEX IF NOT EXISTS idx_stripe_products_metadata ON stripe_products USING GIN (metadata);
CREATE INDEX IF NOT EXISTS idx_stripe_products_updated_at ON stripe_products (updated_at);
CREATE INDEX IF NOT EXISTS idx_stripe_products_livemode ON stripe_products (livemode);

-- Add trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_stripe_products_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER trigger_stripe_products_updated_at
    BEFORE UPDATE ON stripe_products
    FOR EACH ROW
    EXECUTE FUNCTION update_stripe_products_updated_at();

-- Add comments for documentation
COMMENT ON COLUMN stripe_products.metadata IS 'Stripe metadata key-value pairs';
COMMENT ON COLUMN stripe_products.url IS 'Product URL from Stripe';
COMMENT ON COLUMN stripe_products.images IS 'Array of product image URLs';
COMMENT ON COLUMN stripe_products.package_dimensions IS 'Physical product dimensions (height, length, weight, width)';
COMMENT ON COLUMN stripe_products.shippable IS 'Whether product can be shipped';
COMMENT ON COLUMN stripe_products.statement_descriptor IS 'Statement descriptor for billing';
COMMENT ON COLUMN stripe_products.tax_code IS 'Tax code for the product';
COMMENT ON COLUMN stripe_products.unit_label IS 'Unit label (e.g., "per seat", "per GB")';
COMMENT ON COLUMN stripe_products.livemode IS 'Whether this is from live or test mode';
