-- Migration: Update subscription plans with professional marketing copy
-- This migration updates existing subscription plans with clean, professional descriptions

-- Update existing plans with professional copy
UPDATE subscription_plans 
SET 
    name = 'Essential Monthly',
    description = 'Unlock unlimited access to our premium streaming library. Perfect for professionals seeking continuous learning and industry insights.',
    short_desc = 'Essential Monthly Access',
    is_active = true,
    is_promoted = false,
    sort_order = 1,
    updated_at = NOW()
WHERE name = 'Monthly';

UPDATE subscription_plans 
SET 
    name = 'Premium Annual',
    description = 'Our most popular plan. Save 33% with annual billing while enjoying complete access to 1,700+ premium videos and live conference streams.',
    short_desc = 'Best Value - Save 33%',
    is_active = true,
    is_promoted = true,
    sort_order = 2,
    updated_at = NOW()
WHERE name = 'Yearly';

UPDATE subscription_plans 
SET 
    name = 'Premium Monthly',
    description = 'Full access to our complete library of 1,700+ premium videos and presentations. New content added monthly. Cancel anytime.',
    short_desc = 'Complete Library Access',
    is_active = true,
    is_promoted = false,
    sort_order = 3,
    updated_at = NOW()
WHERE name = 'Premium Monthly';

UPDATE subscription_plans 
SET 
    name = 'Annual Pro',
    description = 'Exclusive annual plan with maximum savings. Full library access plus priority support and early access to new content.',
    short_desc = 'Maximum Savings - Pro Benefits',
    is_active = true,
    is_promoted = false,
    sort_order = 4,
    updated_at = NOW()
WHERE name = 'Premuim Annual';

-- Update hidden plans to be more professional
UPDATE subscription_plans 
SET 
    name = 'Conference + Library Bundle',
    description = 'Complete access to our streaming library plus exclusive conference content and live event streams.',
    short_desc = 'Complete Conference Access',
    is_active = true,
    is_promoted = false,
    sort_order = 5,
    updated_at = NOW()
WHERE name = 'Expo Conf + Streaming Library';

UPDATE subscription_plans 
SET 
    name = 'Semi-Annual Premium',
    description = 'Premium access with conference benefits. Includes library access, live conference streams, and exclusive event content.',
    short_desc = 'Premium + Conference Benefits',
    is_active = true,
    is_promoted = false,
    sort_order = 6,
    updated_at = NOW()
WHERE name = 'Premium Semi-Annual';

UPDATE subscription_plans 
SET 
    name = 'Annual Premium',
    description = 'Ultimate annual package with maximum benefits. Complete library access, conference streams, and exclusive event access.',
    short_desc = 'Ultimate Annual Package',
    is_active = true,
    is_promoted = false,
    sort_order = 7,
    updated_at = NOW()
WHERE name = 'Premium Yearly';

-- Add new professional plans
INSERT INTO subscription_plans (
    name, 
    description, 
    short_desc, 
    price, 
    currency, 
    interval, 
    interval_count, 
    features, 
    is_active, 
    is_promoted, 
    sort_order, 
    created_at, 
    updated_at
) VALUES 
(
    'Starter Monthly',
    'Perfect for getting started. Access to core library content with essential industry insights.',
    'Get Started Today',
    7.99,
    'USD',
    'month',
    1,
    '["Core Library Access", "Essential Content", "Email Support"]',
    true,
    false,
    0,
    NOW(),
    NOW()
),
(
    'Professional Annual',
    'For serious professionals. Complete library access with advanced features and priority support.',
    'Professional Choice',
    89.99,
    'USD',
    'year',
    1,
    '["Complete Library Access", "Priority Support", "Advanced Features", "Early Access"]',
    true,
    false,
    8,
    NOW(),
    NOW()
);

-- Update features for existing plans
UPDATE subscription_plans 
SET features = '["Complete Library Access", "HD Streaming", "Email Support", "Mobile Access"]'
WHERE name = 'Essential Monthly';

UPDATE subscription_plans 
SET features = '["Complete Library Access", "HD Streaming", "Priority Support", "Mobile Access", "Offline Downloads", "33% Savings"]'
WHERE name = 'Premium Annual';

UPDATE subscription_plans 
SET features = '["Complete Library Access", "HD Streaming", "Email Support", "Mobile Access", "Cancel Anytime"]'
WHERE name = 'Premium Monthly';

UPDATE subscription_plans 
SET features = '["Complete Library Access", "HD Streaming", "Priority Support", "Mobile Access", "Offline Downloads", "Pro Benefits"]'
WHERE name = 'Annual Pro';

-- Add comments for documentation
COMMENT ON TABLE subscription_plans IS 'Professional subscription plans with clean marketing copy and clear value propositions';
COMMENT ON COLUMN subscription_plans.name IS 'Professional plan name for marketing';
COMMENT ON COLUMN subscription_plans.description IS 'Detailed professional description highlighting value proposition';
COMMENT ON COLUMN subscription_plans.short_desc IS 'Concise marketing tagline for plan cards';
COMMENT ON COLUMN subscription_plans.features IS 'JSON array of key features and benefits';
COMMENT ON COLUMN subscription_plans.is_promoted IS 'Whether to highlight this plan as featured';
COMMENT ON COLUMN subscription_plans.sort_order IS 'Display order for plan listing (lower numbers first)'; 