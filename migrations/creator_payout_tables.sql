-- =====================================================
-- MIGRATION: Creator Payout System Tables
-- Phase: 7B - Creator Payout System
-- Created: October 21, 2025
-- Description: Complete creator/presenter payout system with configurable formulas
-- =====================================================

-- =====================================================
-- 1. PRESENTERS TABLE
-- =====================================================

CREATE TABLE IF NOT EXISTS presenters (
    id SERIAL PRIMARY KEY,
    
    -- Basic Information
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL, -- Optional link to user account
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE,
    bio TEXT,
    avatar_url TEXT,
    
    -- Payment Details
    payment_method VARCHAR(50), -- stripe, paypal, wire, check
    stripe_connect_id VARCHAR(255),
    paypal_email VARCHAR(255),
    tax_id VARCHAR(100),
    bank_account_last4 VARCHAR(4),
    
    -- Address (for tax/payment purposes)
    address_line1 VARCHAR(255),
    address_line2 VARCHAR(255),
    city VARCHAR(100),
    state VARCHAR(100),
    postal_code VARCHAR(20),
    country VARCHAR(2) DEFAULT 'US',
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    verified BOOLEAN DEFAULT false,
    verified_at TIMESTAMP,
    verified_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    
    -- Cached Statistics
    total_videos INTEGER DEFAULT 0,
    total_views BIGINT DEFAULT 0,
    total_watch_minutes BIGINT DEFAULT 0,
    total_earnings DECIMAL(12,2) DEFAULT 0.00,
    lifetime_paid DECIMAL(12,2) DEFAULT 0.00,
    
    -- Metadata
    notes TEXT,
    internal_notes TEXT, -- Admin-only notes
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for presenters
CREATE INDEX IF NOT EXISTS idx_presenters_user_id ON presenters(user_id);
CREATE INDEX IF NOT EXISTS idx_presenters_email ON presenters(email);
CREATE INDEX IF NOT EXISTS idx_presenters_is_active ON presenters(is_active);
CREATE INDEX IF NOT EXISTS idx_presenters_verified ON presenters(verified);
CREATE INDEX IF NOT EXISTS idx_presenters_name ON presenters(name);

-- Comments
COMMENT ON TABLE presenters IS 'Registry of content creators/presenters who appear in videos and receive payouts';
COMMENT ON COLUMN presenters.user_id IS 'Optional link to platform user account if presenter has login access';
COMMENT ON COLUMN presenters.stripe_connect_id IS 'Stripe Connect account ID for automated payouts';
COMMENT ON COLUMN presenters.total_earnings IS 'Cumulative earnings calculated across all payouts';
COMMENT ON COLUMN presenters.lifetime_paid IS 'Actual amount paid out (may differ from earnings due to deductions)';

-- =====================================================
-- 2. VIDEO PRESENTERS TABLE (Many-to-Many Attribution)
-- =====================================================

CREATE TABLE IF NOT EXISTS video_presenters (
    id SERIAL PRIMARY KEY,
    
    -- Relationships
    video_id INTEGER NOT NULL REFERENCES master_video_list(id) ON DELETE CASCADE,
    presenter_id INTEGER NOT NULL REFERENCES presenters(id) ON DELETE CASCADE,
    
    -- Attribution Details
    role VARCHAR(50) DEFAULT 'presenter', -- host, guest, interviewer, presenter, expert
    attribution_percentage DECIMAL(5,2) DEFAULT 100.00, -- Percentage of video earnings (0.00-100.00)
    is_primary BOOLEAN DEFAULT false, -- Primary presenter for the video
    
    -- Display Order
    display_order INTEGER DEFAULT 0, -- Order to display presenters (lower = first)
    
    -- Metadata
    added_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    notes TEXT, -- Attribution notes or special agreements
    
    -- Constraints
    UNIQUE(video_id, presenter_id),
    CONSTRAINT chk_attribution_percentage CHECK (attribution_percentage >= 0 AND attribution_percentage <= 100)
);

-- Indexes for video_presenters
CREATE INDEX IF NOT EXISTS idx_video_presenters_video_id ON video_presenters(video_id);
CREATE INDEX IF NOT EXISTS idx_video_presenters_presenter_id ON video_presenters(presenter_id);
CREATE INDEX IF NOT EXISTS idx_video_presenters_primary ON video_presenters(is_primary) WHERE is_primary = true;
CREATE INDEX IF NOT EXISTS idx_video_presenters_video_order ON video_presenters(video_id, display_order);

-- Comments
COMMENT ON TABLE video_presenters IS 'Many-to-many relationship linking videos to presenters with attribution details';
COMMENT ON COLUMN video_presenters.attribution_percentage IS 'Percentage of video earnings allocated to this presenter (multiple presenters can split earnings)';
COMMENT ON COLUMN video_presenters.is_primary IS 'Marks the main/primary presenter for the video';

-- =====================================================
-- 3. PAYOUT FORMULAS TABLE (Configurable by Admin)
-- =====================================================

CREATE TABLE IF NOT EXISTS payout_formulas (
    id SERIAL PRIMARY KEY,
    
    -- Formula Identity
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    
    -- Formula Type & Configuration
    formula_type VARCHAR(50) NOT NULL, -- per_view, per_watch_minute, tier_based, flat_rate, hybrid
    base_rate DECIMAL(10,6) DEFAULT 0.000000, -- Base rate per unit (e.g., $0.01 per view)
    
    -- Tier-Based Configuration (JSONB for flexibility)
    tier_config JSONB, 
    /* Example tier_config:
    {
      "tiers": [
        {"min": 0, "max": 1000, "rate": 0.005},
        {"min": 1001, "max": 10000, "rate": 0.010},
        {"min": 10001, "max": null, "rate": 0.015}
      ]
    }
    */
    
    -- Multipliers (for bonus calculations)
    subscriber_multiplier DECIMAL(5,2) DEFAULT 1.00, -- Bonus multiplier for subscriber views (1.5 = 50% bonus)
    completion_multiplier DECIMAL(5,2) DEFAULT 1.00, -- Bonus for high completion rate
    engagement_multiplier DECIMAL(5,2) DEFAULT 1.00, -- Bonus for likes/comments
    
    -- Thresholds for Multipliers
    completion_threshold DECIMAL(5,2) DEFAULT 80.00, -- % completion needed for bonus
    engagement_threshold INTEGER DEFAULT 10, -- Min likes+comments for bonus
    
    -- Min/Max Payout Limits
    min_payout DECIMAL(10,2) DEFAULT 0.00, -- Minimum payout per period
    max_payout DECIMAL(10,2), -- Maximum payout per period (NULL = unlimited)
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    is_default BOOLEAN DEFAULT false, -- Only one formula can be default
    
    -- Metadata
    effective_date DATE, -- When this formula becomes active
    expiration_date DATE, -- When this formula expires
    created_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for payout_formulas
CREATE INDEX IF NOT EXISTS idx_payout_formulas_active ON payout_formulas(is_active);
CREATE INDEX IF NOT EXISTS idx_payout_formulas_default ON payout_formulas(is_default) WHERE is_default = true;
CREATE INDEX IF NOT EXISTS idx_payout_formulas_type ON payout_formulas(formula_type);
CREATE INDEX IF NOT EXISTS idx_payout_formulas_effective_date ON payout_formulas(effective_date);

-- Comments
COMMENT ON TABLE payout_formulas IS 'Configurable payout calculation formulas for presenter earnings';
COMMENT ON COLUMN payout_formulas.formula_type IS 'Type of calculation: per_view, per_watch_minute, tier_based, flat_rate, hybrid';
COMMENT ON COLUMN payout_formulas.tier_config IS 'JSONB configuration for tier-based payouts with different rates per view range';
COMMENT ON COLUMN payout_formulas.subscriber_multiplier IS 'Multiplier applied to views from subscribers (e.g., 1.5 = 50% bonus)';

-- =====================================================
-- 4. PRESENTER PAYOUTS TABLE (Monthly Payout Records)
-- =====================================================

CREATE TABLE IF NOT EXISTS presenter_payouts (
    id SERIAL PRIMARY KEY,
    
    -- Relationships
    presenter_id INTEGER NOT NULL REFERENCES presenters(id) ON DELETE CASCADE,
    formula_id INTEGER REFERENCES payout_formulas(id) ON DELETE SET NULL,
    
    -- Payout Period
    payout_month DATE NOT NULL, -- First day of month (e.g., 2025-10-01)
    
    -- Performance Metrics
    total_videos INTEGER DEFAULT 0,
    total_views BIGINT DEFAULT 0,
    total_watch_minutes BIGINT DEFAULT 0,
    unique_viewers INTEGER DEFAULT 0,
    subscriber_views BIGINT DEFAULT 0,
    avg_completion_rate DECIMAL(5,2) DEFAULT 0.00,
    total_engagement INTEGER DEFAULT 0, -- likes + comments + shares
    
    -- Financial Calculations
    base_amount DECIMAL(12,2) DEFAULT 0.00, -- Amount calculated from formula
    bonus_amount DECIMAL(12,2) DEFAULT 0.00, -- Bonuses (subscriber, completion, etc.)
    adjustment_amount DECIMAL(12,2) DEFAULT 0.00, -- Manual adjustments (+ or -)
    deductions DECIMAL(12,2) DEFAULT 0.00, -- Deductions (chargebacks, corrections)
    final_amount DECIMAL(12,2) NOT NULL, -- Final payout amount
    
    -- Currency
    currency CHAR(3) DEFAULT 'USD',
    
    -- Status Tracking
    status VARCHAR(50) DEFAULT 'pending', 
    /* Status values:
       - pending: Calculated, awaiting review
       - approved: Approved for payment
       - processing: Payment in progress
       - paid: Successfully paid
       - failed: Payment failed
       - cancelled: Payout cancelled
       - on_hold: Temporarily held
    */
    
    -- Payment Details
    payment_method VARCHAR(50), -- stripe, paypal, wire, check
    payment_reference VARCHAR(255), -- Transaction ID or check number
    payment_fee DECIMAL(10,2) DEFAULT 0.00, -- Processing fees
    paid_at TIMESTAMP,
    
    -- Calculation Details
    calculation_data JSONB, -- Store detailed breakdown for transparency
    /* Example calculation_data:
    {
      "formula_used": "per_view_tier",
      "breakdown": {
        "tier_1_views": 1000,
        "tier_1_rate": 0.005,
        "tier_1_amount": 5.00,
        "tier_2_views": 5000,
        "tier_2_rate": 0.010,
        "tier_2_amount": 50.00
      },
      "bonuses": {
        "subscriber_bonus": 10.00,
        "completion_bonus": 5.00
      }
    }
    */
    
    -- Metadata
    notes TEXT, -- Public notes (visible to presenter)
    admin_notes TEXT, -- Admin-only notes
    
    -- Audit Trail
    calculated_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    calculated_at TIMESTAMP,
    approved_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    approved_at TIMESTAMP,
    paid_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    UNIQUE(presenter_id, payout_month),
    CONSTRAINT chk_payout_status CHECK (status IN ('pending', 'approved', 'processing', 'paid', 'failed', 'cancelled', 'on_hold'))
);

-- Indexes for presenter_payouts
CREATE INDEX IF NOT EXISTS idx_presenter_payouts_presenter_id ON presenter_payouts(presenter_id);
CREATE INDEX IF NOT EXISTS idx_presenter_payouts_month ON presenter_payouts(payout_month DESC);
CREATE INDEX IF NOT EXISTS idx_presenter_payouts_status ON presenter_payouts(status);
CREATE INDEX IF NOT EXISTS idx_presenter_payouts_presenter_month ON presenter_payouts(presenter_id, payout_month DESC);
CREATE INDEX IF NOT EXISTS idx_presenter_payouts_pending ON presenter_payouts(status, payout_month) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_presenter_payouts_approved ON presenter_payouts(status, payout_month) WHERE status = 'approved';

-- Comments
COMMENT ON TABLE presenter_payouts IS 'Monthly payout records for presenters with detailed calculations and status tracking';
COMMENT ON COLUMN presenter_payouts.payout_month IS 'First day of the payout month (e.g., 2025-10-01 for October 2025)';
COMMENT ON COLUMN presenter_payouts.calculation_data IS 'JSONB storing detailed breakdown of payout calculation for transparency';
COMMENT ON COLUMN presenter_payouts.final_amount IS 'Final payout amount after all bonuses, adjustments, and deductions';

-- =====================================================
-- 5. PAYOUT TRANSACTIONS TABLE (Payment History)
-- =====================================================

CREATE TABLE IF NOT EXISTS payout_transactions (
    id SERIAL PRIMARY KEY,
    
    -- Relationships
    payout_id INTEGER REFERENCES presenter_payouts(id) ON DELETE CASCADE,
    presenter_id INTEGER NOT NULL REFERENCES presenters(id) ON DELETE CASCADE,
    
    -- Transaction Details
    transaction_type VARCHAR(50) NOT NULL, -- payment, adjustment, refund, chargeback, bonus
    amount DECIMAL(12,2) NOT NULL,
    currency CHAR(3) DEFAULT 'USD',
    
    -- Payment Details
    payment_method VARCHAR(50), -- stripe, paypal, wire, check, manual
    payment_provider VARCHAR(50), -- stripe, paypal, bank, internal
    provider_transaction_id VARCHAR(255), -- External transaction ID
    
    -- Status
    status VARCHAR(50) DEFAULT 'pending', 
    /* Status values:
       - pending: Transaction initiated
       - processing: In progress
       - completed: Successfully completed
       - failed: Transaction failed
       - reversed: Transaction reversed/refunded
    */
    
    -- Metadata
    description TEXT,
    notes TEXT,
    error_message TEXT,
    
    -- Payment Details
    fee_amount DECIMAL(10,2) DEFAULT 0.00, -- Transaction fees
    net_amount DECIMAL(12,2), -- Amount after fees (calculated)
    
    -- Audit Trail
    processed_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    processed_at TIMESTAMP,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT chk_transaction_type CHECK (transaction_type IN ('payment', 'adjustment', 'refund', 'chargeback', 'bonus', 'correction')),
    CONSTRAINT chk_transaction_status CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'reversed'))
);

-- Indexes for payout_transactions
CREATE INDEX IF NOT EXISTS idx_payout_transactions_payout_id ON payout_transactions(payout_id);
CREATE INDEX IF NOT EXISTS idx_payout_transactions_presenter_id ON payout_transactions(presenter_id);
CREATE INDEX IF NOT EXISTS idx_payout_transactions_status ON payout_transactions(status);
CREATE INDEX IF NOT EXISTS idx_payout_transactions_type ON payout_transactions(transaction_type);
CREATE INDEX IF NOT EXISTS idx_payout_transactions_created_at ON payout_transactions(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_payout_transactions_provider_id ON payout_transactions(provider_transaction_id) WHERE provider_transaction_id IS NOT NULL;

-- Comments
COMMENT ON TABLE payout_transactions IS 'Detailed transaction history for all payout-related payments and adjustments';
COMMENT ON COLUMN payout_transactions.transaction_type IS 'Type: payment, adjustment, refund, chargeback, bonus, correction';
COMMENT ON COLUMN payout_transactions.provider_transaction_id IS 'External transaction ID from payment provider (Stripe, PayPal, etc.)';

-- =====================================================
-- TRIGGERS FOR AUTOMATIC TIMESTAMP UPDATES
-- =====================================================

-- Update presenters.updated_at
CREATE OR REPLACE FUNCTION update_presenters_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_presenters_updated_at ON presenters;
CREATE TRIGGER trigger_presenters_updated_at
    BEFORE UPDATE ON presenters
    FOR EACH ROW
    EXECUTE FUNCTION update_presenters_updated_at();

-- Update payout_formulas.updated_at
DROP TRIGGER IF EXISTS trigger_payout_formulas_updated_at ON payout_formulas;
CREATE TRIGGER trigger_payout_formulas_updated_at
    BEFORE UPDATE ON payout_formulas
    FOR EACH ROW
    EXECUTE FUNCTION update_presenters_updated_at();

-- Update presenter_payouts.updated_at
DROP TRIGGER IF EXISTS trigger_presenter_payouts_updated_at ON presenter_payouts;
CREATE TRIGGER trigger_presenter_payouts_updated_at
    BEFORE UPDATE ON presenter_payouts
    FOR EACH ROW
    EXECUTE FUNCTION update_presenters_updated_at();

-- Update payout_transactions.updated_at
DROP TRIGGER IF EXISTS trigger_payout_transactions_updated_at ON payout_transactions;
CREATE TRIGGER trigger_payout_transactions_updated_at
    BEFORE UPDATE ON payout_transactions
    FOR EACH ROW
    EXECUTE FUNCTION update_presenters_updated_at();

-- =====================================================
-- SUCCESS MESSAGE
-- =====================================================

DO $$
BEGIN
    RAISE NOTICE '✅ Creator Payout System tables created successfully!';
    RAISE NOTICE '📊 Tables: presenters, video_presenters, payout_formulas, presenter_payouts, payout_transactions';
    RAISE NOTICE '🔧 Triggers: Auto-update timestamps on all tables';
    RAISE NOTICE '📈 Indexes: 30+ indexes for optimal query performance';
END $$;

