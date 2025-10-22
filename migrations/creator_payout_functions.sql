-- =====================================================
-- MIGRATION: Creator Payout System Database Functions
-- Phase: 7B - Creator Payout Calculation Logic
-- Created: October 21, 2025
-- Description: Automated payout calculation and generation functions
-- =====================================================

-- =====================================================
-- FUNCTION 1: Calculate Individual Video Payout
-- =====================================================

CREATE OR REPLACE FUNCTION calculate_video_payout(
    p_video_id INTEGER,
    p_presenter_id INTEGER,
    p_month DATE,
    p_formula_id INTEGER
) RETURNS TABLE (
    video_id INTEGER,
    video_title TEXT,
    views BIGINT,
    watch_minutes BIGINT,
    attribution_pct DECIMAL(5,2),
    payout_amount DECIMAL(12,2)
) AS $$
DECLARE
    formula_record RECORD;
    video_views BIGINT;
    video_watch_time BIGINT;
    attribution_percentage DECIMAL(5,2);
    calculated_amount DECIMAL(12,2) := 0.00;
BEGIN
    -- Get formula details
    SELECT * INTO formula_record FROM payout_formulas WHERE id = p_formula_id AND is_active = true;
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Formula ID % not found or not active', p_formula_id;
    END IF;
    
    -- Get attribution percentage for this presenter on this video
    SELECT vp.attribution_percentage INTO attribution_percentage
    FROM video_presenters vp
    WHERE vp.video_id = p_video_id AND vp.presenter_id = p_presenter_id;
    
    IF NOT FOUND THEN
        attribution_percentage := 100.00; -- Default to 100% if no attribution record
    END IF;
    
    -- Get video metrics for the month
    SELECT 
        COALESCE(SUM(vm.views), 0),
        COALESCE(SUM(vm.watch_time) / 60, 0) -- Convert seconds to minutes
    INTO video_views, video_watch_time
    FROM video_metrics vm
    WHERE vm.video_id = p_video_id
      AND DATE_TRUNC('month', vm.date) = p_month;
    
    -- Calculate based on formula type
    IF formula_record.formula_type = 'per_view' THEN
        calculated_amount := video_views * formula_record.base_rate;
    
    ELSIF formula_record.formula_type = 'per_watch_minute' THEN
        calculated_amount := video_watch_time * formula_record.base_rate;
    
    ELSIF formula_record.formula_type = 'flat_rate' THEN
        calculated_amount := formula_record.base_rate;
    
    END IF;
    
    -- Apply attribution percentage
    calculated_amount := calculated_amount * (attribution_percentage / 100.0);
    
    -- Return results
    RETURN QUERY
    SELECT 
        p_video_id,
        mvl.title,
        video_views,
        video_watch_time,
        attribution_percentage,
        calculated_amount
    FROM master_video_list mvl
    WHERE mvl.id = p_video_id;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION calculate_video_payout IS 'Calculates payout for a single video for a specific presenter and month';

-- =====================================================
-- FUNCTION 2: Calculate Presenter Monthly Payout
-- =====================================================

CREATE OR REPLACE FUNCTION calculate_presenter_payout(
    p_presenter_id INTEGER,
    p_month DATE,
    p_formula_id INTEGER
) RETURNS TABLE (
    presenter_id INTEGER,
    presenter_name TEXT,
    total_videos INTEGER,
    total_views BIGINT,
    total_watch_minutes BIGINT,
    base_amount DECIMAL(12,2),
    bonus_amount DECIMAL(12,2),
    final_amount DECIMAL(12,2),
    calculation_details JSONB
) AS $$
DECLARE
    formula_record RECORD;
    video_record RECORD;
    base_calc DECIMAL(12,2) := 0.00;
    bonus_calc DECIMAL(12,2) := 0.00;
    total_calc DECIMAL(12,2) := 0.00;
    video_count INTEGER := 0;
    cumulative_views BIGINT := 0;
    cumulative_watch_minutes BIGINT := 0;
    calculation_json JSONB := '{"videos": []}'::JSONB;
    video_details JSONB;
BEGIN
    -- Get formula details
    SELECT * INTO formula_record FROM payout_formulas WHERE id = p_formula_id AND is_active = true;
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Formula ID % not found or not active', p_formula_id;
    END IF;
    
    -- Loop through all videos for this presenter
    FOR video_record IN
        SELECT 
            vp.video_id,
            vp.attribution_percentage,
            mvl.title,
            COALESCE(SUM(vm.views), 0) as views,
            COALESCE(SUM(vm.watch_time) / 60, 0) as watch_minutes
        FROM video_presenters vp
        INNER JOIN master_video_list mvl ON vp.video_id = mvl.id
        LEFT JOIN video_metrics vm ON vp.video_id = vm.video_id 
            AND DATE_TRUNC('month', vm.date) = p_month
        WHERE vp.presenter_id = p_presenter_id
        GROUP BY vp.video_id, vp.attribution_percentage, mvl.title
    LOOP
        DECLARE
            video_amount DECIMAL(12,2) := 0.00;
        BEGIN
            -- Calculate for this video
            IF formula_record.formula_type = 'per_view' THEN
                video_amount := video_record.views * formula_record.base_rate;
            
            ELSIF formula_record.formula_type = 'per_watch_minute' THEN
                video_amount := video_record.watch_minutes * formula_record.base_rate;
            
            ELSIF formula_record.formula_type = 'flat_rate' THEN
                video_amount := formula_record.base_rate;
            
            END IF;
            
            -- Apply attribution percentage
            video_amount := video_amount * (video_record.attribution_percentage / 100.0);
            
            -- Add to base calculation
            base_calc := base_calc + video_amount;
            
            -- Track cumulative stats
            video_count := video_count + 1;
            cumulative_views := cumulative_views + video_record.views;
            cumulative_watch_minutes := cumulative_watch_minutes + video_record.watch_minutes;
            
            -- Store video details
            video_details := jsonb_build_object(
                'video_id', video_record.video_id,
                'title', video_record.title,
                'views', video_record.views,
                'watch_minutes', video_record.watch_minutes,
                'attribution_pct', video_record.attribution_percentage,
                'amount', video_amount
            );
            
            calculation_json := jsonb_set(
                calculation_json,
                array['videos', video_count::TEXT],
                video_details
            );
        END;
    END LOOP;
    
    -- Calculate bonuses (simplified for now)
    -- TODO: Add subscriber multiplier, completion multiplier logic
    bonus_calc := 0.00;
    
    -- Calculate final amount
    total_calc := base_calc + bonus_calc;
    
    -- Apply min/max limits
    IF total_calc < formula_record.min_payout THEN
        total_calc := formula_record.min_payout;
    END IF;
    
    IF formula_record.max_payout IS NOT NULL AND total_calc > formula_record.max_payout THEN
        total_calc := formula_record.max_payout;
    END IF;
    
    -- Add summary to calculation JSON
    calculation_json := calculation_json || jsonb_build_object(
        'formula_type', formula_record.formula_type,
        'base_rate', formula_record.base_rate,
        'total_videos', video_count,
        'total_views', cumulative_views,
        'total_watch_minutes', cumulative_watch_minutes,
        'base_amount', base_calc,
        'bonus_amount', bonus_calc,
        'final_amount', total_calc
    );
    
    -- Return results
    RETURN QUERY
    SELECT 
        p_presenter_id,
        p.name,
        video_count,
        cumulative_views,
        cumulative_watch_minutes,
        base_calc,
        bonus_calc,
        total_calc,
        calculation_json
    FROM presenters p
    WHERE p.id = p_presenter_id;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION calculate_presenter_payout IS 'Calculates total monthly payout for a presenter across all their videos';

-- =====================================================
-- FUNCTION 3: Generate Monthly Payouts for All Presenters
-- =====================================================

CREATE OR REPLACE FUNCTION generate_monthly_payouts(
    p_month DATE,
    p_calculated_by INTEGER
) RETURNS TABLE (
    generated_count INTEGER,
    total_amount DECIMAL(12,2),
    presenter_details JSONB
) AS $$
DECLARE
    presenter_record RECORD;
    default_formula_id INTEGER;
    calculation_result RECORD;
    generated INTEGER := 0;
    total_payout DECIMAL(12,2) := 0.00;
    results_json JSONB := '{"presenters": []}'::JSONB;
    presenter_detail JSONB;
BEGIN
    -- Get default formula
    SELECT id INTO default_formula_id 
    FROM payout_formulas 
    WHERE is_default = true AND is_active = true
    LIMIT 1;
    
    IF default_formula_id IS NULL THEN
        RAISE EXCEPTION 'No default payout formula configured. Please set a default formula.';
    END IF;
    
    -- Loop through active presenters
    FOR presenter_record IN 
        SELECT id, name, email FROM presenters WHERE is_active = true
    LOOP
        -- Calculate payout for this presenter
        SELECT * INTO calculation_result
        FROM calculate_presenter_payout(
            presenter_record.id,
            p_month,
            default_formula_id
        );
        
        -- Only create payout if there's activity (views > 0)
        IF calculation_result.total_views > 0 THEN
            -- Insert or update payout record
            INSERT INTO presenter_payouts (
                presenter_id,
                payout_month,
                formula_id,
                total_videos,
                total_views,
                total_watch_minutes,
                base_amount,
                bonus_amount,
                final_amount,
                status,
                calculation_data,
                calculated_by,
                calculated_at
            ) VALUES (
                presenter_record.id,
                p_month,
                default_formula_id,
                calculation_result.total_videos,
                calculation_result.total_views,
                calculation_result.total_watch_minutes,
                calculation_result.base_amount,
                calculation_result.bonus_amount,
                calculation_result.final_amount,
                'pending',
                calculation_result.calculation_details,
                p_calculated_by,
                NOW()
            )
            ON CONFLICT (presenter_id, payout_month) 
            DO UPDATE SET 
                formula_id = EXCLUDED.formula_id,
                total_videos = EXCLUDED.total_videos,
                total_views = EXCLUDED.total_views,
                total_watch_minutes = EXCLUDED.total_watch_minutes,
                base_amount = EXCLUDED.base_amount,
                bonus_amount = EXCLUDED.bonus_amount,
                final_amount = EXCLUDED.final_amount,
                calculation_data = EXCLUDED.calculation_data,
                calculated_by = EXCLUDED.calculated_by,
                calculated_at = EXCLUDED.calculated_at,
                updated_at = NOW();
            
            generated := generated + 1;
            total_payout := total_payout + calculation_result.final_amount;
            
            -- Build presenter detail JSON
            presenter_detail := jsonb_build_object(
                'presenter_id', presenter_record.id,
                'name', presenter_record.name,
                'email', presenter_record.email,
                'videos', calculation_result.total_videos,
                'views', calculation_result.total_views,
                'amount', calculation_result.final_amount
            );
            
            results_json := jsonb_set(
                results_json,
                array['presenters', generated::TEXT],
                presenter_detail
            );
        END IF;
    END LOOP;
    
    -- Add summary
    results_json := results_json || jsonb_build_object(
        'month', p_month,
        'generated_count', generated,
        'total_amount', total_payout,
        'formula_id', default_formula_id,
        'generated_at', NOW()
    );
    
    -- Return results
    RETURN QUERY
    SELECT generated, total_payout, results_json;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION generate_monthly_payouts IS 'Generates payout records for all active presenters for a specific month';

-- =====================================================
-- FUNCTION 4: Update Presenter Statistics
-- =====================================================

CREATE OR REPLACE FUNCTION update_presenter_statistics(
    p_presenter_id INTEGER
) RETURNS VOID AS $$
BEGIN
    UPDATE presenters
    SET 
        total_videos = (
            SELECT COUNT(DISTINCT vp.video_id)
            FROM video_presenters vp
            WHERE vp.presenter_id = p_presenter_id
        ),
        total_views = (
            SELECT COALESCE(SUM(mvl.views), 0)
            FROM video_presenters vp
            INNER JOIN master_video_list mvl ON vp.video_id = mvl.id
            WHERE vp.presenter_id = p_presenter_id
        ),
        total_watch_minutes = (
            SELECT COALESCE(SUM(mvl.average_watch_time * mvl.views / 60), 0)
            FROM video_presenters vp
            INNER JOIN master_video_list mvl ON vp.video_id = mvl.id
            WHERE vp.presenter_id = p_presenter_id
        ),
        total_earnings = (
            SELECT COALESCE(SUM(final_amount), 0)
            FROM presenter_payouts
            WHERE presenter_id = p_presenter_id
        ),
        lifetime_paid = (
            SELECT COALESCE(SUM(final_amount), 0)
            FROM presenter_payouts
            WHERE presenter_id = p_presenter_id
              AND status = 'paid'
        ),
        updated_at = NOW()
    WHERE id = p_presenter_id;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION update_presenter_statistics IS 'Updates cached statistics for a presenter (videos, views, earnings)';

-- =====================================================
-- FUNCTION 5: Get Payout Summary for Month
-- =====================================================

CREATE OR REPLACE FUNCTION get_payout_summary(
    p_month DATE
) RETURNS TABLE (
    payout_month DATE,
    total_presenters INTEGER,
    total_videos INTEGER,
    total_views BIGINT,
    total_amount DECIMAL(12,2),
    pending_amount DECIMAL(12,2),
    approved_amount DECIMAL(12,2),
    paid_amount DECIMAL(12,2),
    status_breakdown JSONB
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        p_month,
        COUNT(*)::INTEGER,
        COALESCE(SUM(pp.total_videos), 0)::INTEGER,
        COALESCE(SUM(pp.total_views), 0),
        COALESCE(SUM(pp.final_amount), 0.00),
        COALESCE(SUM(pp.final_amount) FILTER (WHERE pp.status = 'pending'), 0.00),
        COALESCE(SUM(pp.final_amount) FILTER (WHERE pp.status = 'approved'), 0.00),
        COALESCE(SUM(pp.final_amount) FILTER (WHERE pp.status = 'paid'), 0.00),
        jsonb_build_object(
            'pending', COUNT(*) FILTER (WHERE pp.status = 'pending'),
            'approved', COUNT(*) FILTER (WHERE pp.status = 'approved'),
            'processing', COUNT(*) FILTER (WHERE pp.status = 'processing'),
            'paid', COUNT(*) FILTER (WHERE pp.status = 'paid'),
            'failed', COUNT(*) FILTER (WHERE pp.status = 'failed'),
            'cancelled', COUNT(*) FILTER (WHERE pp.status = 'cancelled')
        )
    FROM presenter_payouts pp
    WHERE pp.payout_month = p_month;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION get_payout_summary IS 'Returns summary statistics for all payouts in a specific month';

-- =====================================================
-- GRANT PERMISSIONS
-- =====================================================

GRANT EXECUTE ON FUNCTION calculate_video_payout(INTEGER, INTEGER, DATE, INTEGER) TO PUBLIC;
GRANT EXECUTE ON FUNCTION calculate_presenter_payout(INTEGER, DATE, INTEGER) TO PUBLIC;
GRANT EXECUTE ON FUNCTION generate_monthly_payouts(DATE, INTEGER) TO PUBLIC;
GRANT EXECUTE ON FUNCTION update_presenter_statistics(INTEGER) TO PUBLIC;
GRANT EXECUTE ON FUNCTION get_payout_summary(DATE) TO PUBLIC;

-- =====================================================
-- SUCCESS MESSAGE
-- =====================================================

DO $$
BEGIN
    RAISE NOTICE '✅ Creator Payout System functions created successfully!';
    RAISE NOTICE '🔧 Functions:';
    RAISE NOTICE '   - calculate_video_payout(): Calculate single video payout';
    RAISE NOTICE '   - calculate_presenter_payout(): Calculate presenter monthly total';
    RAISE NOTICE '   - generate_monthly_payouts(): Generate payouts for all presenters';
    RAISE NOTICE '   - update_presenter_statistics(): Update presenter cached stats';
    RAISE NOTICE '   - get_payout_summary(): Get monthly payout summary';
    RAISE NOTICE '🚀 Ready to calculate presenter earnings!';
END $$;

