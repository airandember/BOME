-- 🔥 GHOST EXORCISM EXECUTION PLAN 🔥
-- DANGER: These are destructive operations - run with extreme caution!

-- ========================================
-- PHASE 1: BACKUP BEFORE EXORCISM
-- ========================================

-- Create backup table before any ghost removal
CREATE TABLE stripe_subscriptions_backup_pre_exorcism AS 
SELECT * FROM stripe_subscriptions;

-- ========================================
-- PHASE 2: GENTLE EXORCISM (RECOMMENDED)
-- ========================================

-- Option A: Mark ghosts as "archived" instead of deleting
ALTER TABLE stripe_subscriptions ADD COLUMN IF NOT EXISTS is_ghost BOOLEAN DEFAULT FALSE;
ALTER TABLE stripe_subscriptions ADD COLUMN IF NOT EXISTS ghost_reason TEXT;

-- Mark legacy contamination ghosts
UPDATE stripe_subscriptions 
SET is_ghost = TRUE, 
    ghost_reason = 'LEGACY_CONTAMINATION: Non-Stripe price ID'
WHERE stripe_price_id IS NOT NULL 
  AND stripe_price_id NOT LIKE 'price_%';

-- Mark orphaned product ghosts  
UPDATE stripe_subscriptions 
SET is_ghost = TRUE,
    ghost_reason = 'ORPHANED_PRODUCT: Product does not exist in current Stripe account'
WHERE stripe_product_id IN (
    'prod_HjYKGcWGP9r4EC', 'prod_HEmcX1PE8TO2CO', 'prod_FvNAeI348dup9w',
    'prod_FvNAlEGGL452nN', 'prod_HF5YzcBH5Rwr0d', 'prod_GVV5efccnh13h9', 'prod_FvNAJgnw48hwpZ'
);

-- Mark completely broken records
UPDATE stripe_subscriptions 
SET is_ghost = TRUE,
    ghost_reason = 'BROKEN_RECORD: Missing critical Stripe identifiers'
WHERE stripe_price_id IS NULL 
  AND (stripe_product_id IS NULL OR stripe_product_id = '');

-- ========================================
-- PHASE 3: NUCLEAR EXORCISM (DANGEROUS!)
-- ========================================

-- DANGER ZONE: Only run these if you're absolutely sure!

-- Option B: Delete legacy contamination (moves them out of Stripe system)
-- UNCOMMENT ONLY IF YOU'RE SURE:
/*
DELETE FROM stripe_subscriptions 
WHERE stripe_price_id IS NOT NULL 
  AND stripe_price_id NOT LIKE 'price_%'
  AND status = 'canceled';  -- Start with canceled ones only
*/

-- Option C: Delete orphaned product ghosts  
-- UNCOMMENT ONLY IF YOU'RE SURE:
/*
DELETE FROM stripe_subscriptions 
WHERE stripe_product_id IN (
    'prod_HjYKGcWGP9r4EC', 'prod_HEmcX1PE8TO2CO', 'prod_FvNAeI348dup9w',
    'prod_FvNAlEGGL452nN', 'prod_HF5YzcBH5Rwr0d', 'prod_GVV5efccnh13h9', 'prod_FvNAJgnw48hwpZ'
) AND status = 'canceled';  -- Start with canceled ones only
*/

-- ========================================
-- PHASE 4: VERIFICATION
-- ========================================

-- Count remaining records after exorcism
SELECT 
    'POST_EXORCISM_COUNT' as check_type,
    COUNT(*) as total_subscriptions,
    COUNT(*) FILTER (WHERE stripe_price_id LIKE 'price_%') as real_stripe_subscriptions,
    COUNT(*) FILTER (WHERE is_ghost = TRUE) as marked_ghosts,
    COUNT(*) FILTER (WHERE is_ghost IS NULL OR is_ghost = FALSE) as clean_records
FROM stripe_subscriptions;
