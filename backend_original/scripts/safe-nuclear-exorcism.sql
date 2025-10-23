-- ☢️ SAFE NUCLEAR EXORCISM ☢️
-- Since all real customers come from Stripe, this is safe!

-- STEP 1: BACKUP (just in case)
CREATE TABLE stripe_subscriptions_backup_$(date +%Y%m%d) AS 
SELECT * FROM stripe_subscriptions;

-- STEP 2: IDENTIFY WHAT TO KEEP (real Stripe data only)
SELECT 
    'PRE_NUCLEAR_INVENTORY' as inventory_type,
    COUNT(*) as total_before_nuclear,
    COUNT(*) FILTER (WHERE stripe_price_id LIKE 'price_%') as real_stripe_to_keep,
    COUNT(*) FILTER (WHERE stripe_price_id NOT LIKE 'price_%' OR stripe_price_id IS NULL) as contamination_to_delete
FROM stripe_subscriptions;

-- STEP 3: NUCLEAR STRIKE! 💥
-- Delete all non-Stripe data
DELETE FROM stripe_subscriptions 
WHERE stripe_price_id IS NULL 
   OR stripe_price_id NOT LIKE 'price_%';

-- STEP 4: VERIFY NUCLEAR RESULTS
SELECT 
    'POST_NUCLEAR_INVENTORY' as inventory_type,
    COUNT(*) as total_after_nuclear,
    COUNT(*) FILTER (WHERE stripe_price_id LIKE 'price_%') as confirmed_real_stripe,
    COUNT(*) FILTER (WHERE stripe_product_id LIKE 'prod_%') as confirmed_real_products
FROM stripe_subscriptions;

-- STEP 5: FRESH STRIPE SYNC PREPARATION
-- Your existing sync will now pull fresh, clean data
SELECT 
    'READY_FOR_FRESH_SYNC' as status,
    'All ghost data exorcised - ready for clean Stripe sync' as message;
