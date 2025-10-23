-- Migration: Add monthly email tracking
-- This migration adds monthly email usage tracking for better analytics

-- Add monthly email usage table
CREATE TABLE IF NOT EXISTS monthly_email_usage (
    id SERIAL PRIMARY KEY,
    year INTEGER NOT NULL,
    month INTEGER NOT NULL,
    provider VARCHAR(50) NOT NULL,
    emails_sent INTEGER DEFAULT 0,
    emails_failed INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(year, month, provider)
);

-- Add indexes for better performance
CREATE INDEX IF NOT EXISTS idx_monthly_email_usage_year_month ON monthly_email_usage(year, month);
CREATE INDEX IF NOT EXISTS idx_monthly_email_usage_provider ON monthly_email_usage(provider);
CREATE INDEX IF NOT EXISTS idx_monthly_email_usage_year_month_provider ON monthly_email_usage(year, month, provider);

-- Add monthly limits to email_settings
INSERT INTO email_settings (setting_key, setting_value, created_at) 
VALUES 
    ('monthly_email_limit_resend', '3000', CURRENT_TIMESTAMP),
    ('monthly_email_limit_mailgun', '5000', CURRENT_TIMESTAMP)
ON CONFLICT (setting_key) DO NOTHING;

-- Add trigger to update last_updated timestamp
CREATE OR REPLACE FUNCTION update_monthly_email_usage_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.last_updated = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_monthly_email_usage_timestamp
    BEFORE UPDATE ON monthly_email_usage
    FOR EACH ROW
    EXECUTE FUNCTION update_monthly_email_usage_timestamp();
