-- 🔥 TARGETED GHOST EXORCISM: Legacy Price ID Cleanup
-- Based on the analysis showing NO_MATCHING_PLAN for all legacy IDs

-- STEP 1: Create a comprehensive legacy mapping table
CREATE TABLE IF NOT EXISTS legacy_price_mapping (
    legacy_price_id VARCHAR(50) PRIMARY KEY,
    subscription_count INTEGER,
    suggested_plan_name VARCHAR(100),
    suggested_unit_amount INTEGER,
    suggested_currency VARCHAR(3) DEFAULT 'usd',
    business_notes TEXT
);

-- STEP 2: Insert the discovered legacy mappings with business logic
INSERT INTO legacy_price_mapping (legacy_price_id, subscription_count, suggested_plan_name, suggested_unit_amount, business_notes) VALUES
('Combo', 790, 'Conference + Library Bundle', 4500, 'Your biggest legacy product - conference + streaming library access'),
('EMonth', 588, 'Essential Monthly', 997, 'Essential monthly plan - basic library access'),
('YPremium', 415, 'Premium Yearly', 9564, 'Premium yearly plan - matches your current $95.64 pricing'),
('PYear', 250, 'Premium Yearly', 9564, 'Another premium yearly variant'),
('VirtualExpoMonthly', 195, 'Virtual Expo Monthly', 1397, 'Virtual conference monthly access'),
('PPlan', 183, 'Premium Plan', 1397, 'Generic premium plan'),
('EXPOYEAR', 133, 'Expo Yearly', 15564, 'Annual expo/conference access'),
('bestvaluepromo', 29, 'Best Value Promotion', 8982, 'Promotional pricing plan'),
('premiummonthly', 17, 'Premium Monthly', 1397, 'Premium monthly access'),
('PMonth', 17, 'Premium Monthly', 1397, 'Premium monthly variant')
ON CONFLICT (legacy_price_id) DO NOTHING;

-- STEP 3: GENTLE EXORCISM - Mark legacy contamination with proper mapping
UPDATE stripe_subscriptions 
SET is_ghost = TRUE,
    ghost_reason = 'LEGACY_CONTAMINATION: ' || 
        COALESCE(lpm.suggested_plan_name, 'Unknown legacy plan') || 
        ' ($' || COALESCE(lpm.suggested_unit_amount::float / 100, 0) || ')'
FROM legacy_price_mapping lpm
WHERE stripe_subscriptions.stripe_price_id = lpm.legacy_price_id;

-- STEP 4: Update your dynamic mapping to use this legacy data
-- This enhances your existing price_mapping CTE in subscribers.go
SELECT 
    'ENHANCED_LEGACY_MAPPING' as mapping_type,
    legacy_price_id,
    suggested_plan_name,
    suggested_unit_amount,
    subscription_count,
    'Legacy system data - use for dynamic mapping' as usage_note
FROM legacy_price_mapping
ORDER BY subscription_count DESC;

-- STEP 5: Verification - show the exorcism results
SELECT 
    'EXORCISM_RESULTS' as result_type,
    COUNT(*) as total_subscriptions,
    COUNT(*) FILTER (WHERE stripe_price_id LIKE 'price_%') as real_stripe_subscriptions,
    COUNT(*) FILTER (WHERE is_ghost = TRUE AND ghost_reason LIKE 'LEGACY_CONTAMINATION%') as exorcised_legacy_ghosts,
    COUNT(*) FILTER (WHERE is_ghost = TRUE AND ghost_reason LIKE 'ORPHANED_PRODUCT%') as orphaned_product_ghosts,
    COUNT(*) FILTER (WHERE is_ghost IS NULL OR is_ghost = FALSE) as clean_records
FROM stripe_subscriptions;
