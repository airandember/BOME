-- This FK already exists! stripe_prices.product_id -> stripe_products.id ✅
-- No migration needed - constraint already present as "stripe_prices_product_id_fkey"

-- Just verify the existing relationship works
SELECT 
    spr.stripe_id as price_stripe_id,
    spr.unit_amount,
    spr.currency,
    sprod.stripe_id as product_stripe_id,
    sprod.name as product_name
FROM stripe_prices spr
LEFT JOIN stripe_products sprod ON spr.product_id = sprod.id
WHERE spr.unit_amount IS NOT NULL
LIMIT 5;
