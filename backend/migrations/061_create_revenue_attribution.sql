-- Migration: Create revenue attribution tables with custom formula support
-- Description: Track which videos lead to subscriptions with configurable attribution logic

-- Attribution formulas table - stores custom attribution calculation logic
CREATE TABLE IF NOT EXISTS revenue_attribution_formulas (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    formula_type VARCHAR(50) NOT NULL DEFAULT 'last_touch', -- last_touch, first_touch, linear, time_decay, position_based, custom
    
    -- Custom formula configuration (JSON)
    -- Example: {"weights": {"first": 0.4, "last": 0.4, "middle": 0.2}, "decay_rate": 0.5}
    formula_config JSONB DEFAULT '{}',
    
    -- Attribution window settings
    attribution_window_days INTEGER DEFAULT 7, -- How far back to look for video views
    min_watch_percentage DECIMAL(5,2) DEFAULT 25.00, -- Minimum % watched to count (25%)
    
    -- Status and metadata
    is_active BOOLEAN DEFAULT true,
    is_default BOOLEAN DEFAULT false,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(id) ON DELETE SET NULL
);

-- Revenue attribution tracking - actual attribution records
CREATE TABLE IF NOT EXISTS video_revenue_attribution (
    id SERIAL PRIMARY KEY,
    
    -- Entities
    video_id INTEGER NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subscription_id VARCHAR(255), -- Stripe subscription ID
    
    -- Attribution details
    formula_id INTEGER REFERENCES revenue_attribution_formulas(id) ON DELETE SET NULL,
    attribution_type VARCHAR(50) NOT NULL, -- first_touch, last_touch, assisted, etc.
    attribution_weight DECIMAL(5,4) DEFAULT 1.0000, -- 0.0 to 1.0 (for multi-touch)
    
    -- Financial data
    attributed_revenue DECIMAL(10,2) NOT NULL, -- Revenue attributed to this video
    subscription_value DECIMAL(10,2), -- Full subscription value
    
    -- Context
    views_before_conversion INTEGER DEFAULT 0,
    total_watch_time_seconds INTEGER DEFAULT 0,
    last_view_before_conversion TIMESTAMP,
    conversion_timestamp TIMESTAMP NOT NULL,
    
    -- Metadata
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(video_id, user_id, subscription_id, formula_id)
);

-- Video conversion metrics - aggregated conversion data per video
CREATE TABLE IF NOT EXISTS video_conversion_metrics (
    id SERIAL PRIMARY KEY,
    video_id INTEGER NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    formula_id INTEGER REFERENCES revenue_attribution_formulas(id) ON DELETE SET NULL,
    
    -- Conversion metrics
    total_conversions INTEGER DEFAULT 0,
    assisted_conversions INTEGER DEFAULT 0, -- Video was in the journey but not final touch
    
    -- Revenue metrics
    total_attributed_revenue DECIMAL(12,2) DEFAULT 0.00,
    avg_revenue_per_conversion DECIMAL(10,2) DEFAULT 0.00,
    
    -- Conversion funnel
    total_qualified_views INTEGER DEFAULT 0, -- Views that met min watch %
    conversion_rate DECIMAL(5,4) DEFAULT 0.0000, -- conversions / qualified views
    
    -- Time to conversion
    avg_time_to_conversion_hours DECIMAL(10,2), -- Average hours from view to conversion
    
    -- Last updated
    last_calculated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(video_id, formula_id)
);

-- Create indexes for performance
CREATE INDEX idx_revenue_attribution_video ON video_revenue_attribution(video_id);
CREATE INDEX idx_revenue_attribution_user ON video_revenue_attribution(user_id);
CREATE INDEX idx_revenue_attribution_subscription ON video_revenue_attribution(subscription_id);
CREATE INDEX idx_revenue_attribution_formula ON video_revenue_attribution(formula_id);
CREATE INDEX idx_revenue_attribution_created ON video_revenue_attribution(created_at);
CREATE INDEX idx_revenue_attribution_conversion ON video_revenue_attribution(conversion_timestamp);

CREATE INDEX idx_conversion_metrics_video ON video_conversion_metrics(video_id);
CREATE INDEX idx_conversion_metrics_formula ON video_conversion_metrics(formula_id);

CREATE INDEX idx_attribution_formulas_active ON revenue_attribution_formulas(is_active) WHERE is_active = true;
CREATE INDEX idx_attribution_formulas_default ON revenue_attribution_formulas(is_default) WHERE is_default = true;

-- Create default attribution formulas
INSERT INTO revenue_attribution_formulas (name, description, formula_type, attribution_window_days, min_watch_percentage, is_default) VALUES
('Last Touch', 'Full credit to the last video watched before subscription', 'last_touch', 7, 25.00, true),
('First Touch', 'Full credit to the first video that introduced the user', 'first_touch', 30, 25.00, false),
('Linear', 'Equal credit distributed across all videos in the journey', 'linear', 14, 25.00, false),
('Time Decay', 'More credit to recent videos, exponentially decaying for older ones', 'time_decay', 14, 25.00, false),
('Position Based', '40% first, 40% last, 20% distributed to middle videos', 'position_based', 14, 25.00, false);

-- Update formula config for position-based (example of JSON config)
UPDATE revenue_attribution_formulas 
SET formula_config = '{"first_weight": 0.4, "last_weight": 0.4, "middle_weight": 0.2}'
WHERE formula_type = 'position_based';

-- Update formula config for time decay (example of decay rate)
UPDATE revenue_attribution_formulas 
SET formula_config = '{"decay_rate": 0.5, "half_life_days": 3.5}'
WHERE formula_type = 'time_decay';

COMMENT ON TABLE revenue_attribution_formulas IS 'Custom attribution formulas that admins can create and configure';
COMMENT ON TABLE video_revenue_attribution IS 'Individual attribution records linking videos to subscription revenue';
COMMENT ON TABLE video_conversion_metrics IS 'Aggregated conversion and revenue metrics per video';
COMMENT ON COLUMN revenue_attribution_formulas.formula_config IS 'JSON configuration for custom formula parameters';
COMMENT ON COLUMN video_revenue_attribution.attribution_weight IS 'Weight for multi-touch attribution (0.0 to 1.0)';

