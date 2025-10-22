-- =====================================================
-- MIGRATION: Creator Payout System Seed Data
-- Phase: 7B - Initial Configuration
-- Created: October 21, 2025
-- Description: Default payout formulas and example data
-- =====================================================

-- =====================================================
-- INSERT DEFAULT PAYOUT FORMULAS
-- =====================================================

-- Formula 1: Per-View with Tiers (DEFAULT)
INSERT INTO payout_formulas (
    name,
    description,
    formula_type,
    base_rate,
    tier_config,
    subscriber_multiplier,
    completion_multiplier,
    completion_threshold,
    min_payout,
    max_payout,
    is_active,
    is_default
) VALUES (
    'Per-View Tiered',
    'Tiered payout based on view count: $0.005 per view (0-1000), $0.010 per view (1001-10000), $0.015 per view (10001+). Includes 50% bonus for subscriber views and 20% bonus for high completion rate.',
    'per_view',
    0.010000, -- Base rate (middle tier)
    '{"tiers": [{"min": 0, "max": 1000, "rate": 0.005}, {"min": 1001, "max": 10000, "rate": 0.010}, {"min": 10001, "max": null, "rate": 0.015}]}'::JSONB,
    1.50, -- 50% bonus for subscriber views
    1.20, -- 20% bonus for high completion
    80.00, -- 80% completion threshold
    5.00, -- Minimum $5.00 per month
    NULL, -- No maximum
    true,
    true -- This is the default formula
) ON CONFLICT (name) DO UPDATE
SET 
    description = EXCLUDED.description,
    formula_type = EXCLUDED.formula_type,
    base_rate = EXCLUDED.base_rate,
    tier_config = EXCLUDED.tier_config,
    subscriber_multiplier = EXCLUDED.subscriber_multiplier,
    completion_multiplier = EXCLUDED.completion_multiplier,
    completion_threshold = EXCLUDED.completion_threshold,
    min_payout = EXCLUDED.min_payout,
    is_active = EXCLUDED.is_active,
    is_default = EXCLUDED.is_default,
    updated_at = CURRENT_TIMESTAMP;

-- Formula 2: Per-Watch-Minute
INSERT INTO payout_formulas (
    name,
    description,
    formula_type,
    base_rate,
    subscriber_multiplier,
    min_payout,
    max_payout,
    is_active,
    is_default
) VALUES (
    'Per-Watch-Minute',
    'Payout based on watch time: $0.001 per minute watched. Rewards engagement and long-form content.',
    'per_watch_minute',
    0.001000, -- $0.001 per minute
    1.00, -- No subscriber bonus for this formula
    5.00, -- Minimum $5.00 per month
    NULL,
    true,
    false
) ON CONFLICT (name) DO UPDATE
SET 
    description = EXCLUDED.description,
    formula_type = EXCLUDED.formula_type,
    base_rate = EXCLUDED.base_rate,
    subscriber_multiplier = EXCLUDED.subscriber_multiplier,
    min_payout = EXCLUDED.min_payout,
    is_active = EXCLUDED.is_active,
    updated_at = CURRENT_TIMESTAMP;

-- Formula 3: Flat Rate per Video
INSERT INTO payout_formulas (
    name,
    description,
    formula_type,
    base_rate,
    min_payout,
    max_payout,
    is_active,
    is_default
) VALUES (
    'Flat Rate per Video',
    'Fixed payment per video published: $50.00 per video. Simple and predictable.',
    'flat_rate',
    50.000000, -- $50 per video
    50.00,
    50.00, -- Max = min for flat rate
    true,
    false
) ON CONFLICT (name) DO UPDATE
SET 
    description = EXCLUDED.description,
    formula_type = EXCLUDED.formula_type,
    base_rate = EXCLUDED.base_rate,
    min_payout = EXCLUDED.min_payout,
    max_payout = EXCLUDED.max_payout,
    is_active = EXCLUDED.is_active,
    updated_at = CURRENT_TIMESTAMP;

-- Formula 4: High-Volume Creators
INSERT INTO payout_formulas (
    name,
    description,
    formula_type,
    base_rate,
    tier_config,
    subscriber_multiplier,
    completion_multiplier,
    completion_threshold,
    min_payout,
    max_payout,
    is_active,
    is_default
) VALUES (
    'High-Volume Creator',
    'Premium tier for high-performing creators: $0.008 per view (0-5000), $0.012 per view (5001-25000), $0.020 per view (25001+). Higher bonuses for quality engagement.',
    'per_view',
    0.012000,
    '{"tiers": [{"min": 0, "max": 5000, "rate": 0.008}, {"min": 5001, "max": 25000, "rate": 0.012}, {"min": 25001, "max": null, "rate": 0.020}]}'::JSONB,
    2.00, -- 100% bonus for subscriber views
    1.50, -- 50% bonus for high completion
    85.00, -- 85% completion threshold
    10.00,
    NULL,
    true,
    false
) ON CONFLICT (name) DO UPDATE
SET 
    description = EXCLUDED.description,
    formula_type = EXCLUDED.formula_type,
    base_rate = EXCLUDED.base_rate,
    tier_config = EXCLUDED.tier_config,
    subscriber_multiplier = EXCLUDED.subscriber_multiplier,
    completion_multiplier = EXCLUDED.completion_multiplier,
    completion_threshold = EXCLUDED.completion_threshold,
    min_payout = EXCLUDED.min_payout,
    is_active = EXCLUDED.is_active,
    updated_at = CURRENT_TIMESTAMP;

-- =====================================================
-- EXAMPLE PRESENTER (Optional - for testing)
-- =====================================================

-- Uncomment to create an example presenter for testing
/*
INSERT INTO presenters (
    name,
    email,
    bio,
    payment_method,
    is_active,
    verified
) VALUES (
    'Dr. Example Presenter',
    'presenter@example.com',
    'Example presenter for testing the payout system.',
    'stripe',
    true,
    true
) ON CONFLICT (email) DO NOTHING;
*/

-- =====================================================
-- SUCCESS MESSAGE
-- =====================================================

DO $$
DECLARE
    formula_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO formula_count FROM payout_formulas WHERE is_active = true;
    
    RAISE NOTICE '✅ Creator Payout System seed data loaded successfully!';
    RAISE NOTICE '📋 Created % active payout formulas:', formula_count;
    RAISE NOTICE '   1. Per-View Tiered (DEFAULT) - $0.005-$0.015 per view with bonuses';
    RAISE NOTICE '   2. Per-Watch-Minute - $0.001 per minute watched';
    RAISE NOTICE '   3. Flat Rate per Video - $50.00 per video';
    RAISE NOTICE '   4. High-Volume Creator - $0.008-$0.020 per view with premium bonuses';
    RAISE NOTICE '';
    RAISE NOTICE '🎯 Default formula configured and ready to use!';
    RAISE NOTICE '💡 Admins can now:';
    RAISE NOTICE '   - Add presenters';
    RAISE NOTICE '   - Link videos to presenters';
    RAISE NOTICE '   - Generate monthly payouts';
    RAISE NOTICE '   - Configure custom formulas';
END $$;

