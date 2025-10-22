-- Migration: Create user subscriptions and email verification tables
-- Created: 2025-01-03

-- Create user_subscriptions table to track active subscriptions
CREATE TABLE IF NOT EXISTS user_subscriptions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subscription_plan_id INTEGER NOT NULL REFERENCES subscription_plans(id),
    stripe_subscription_id VARCHAR(255) UNIQUE,
    stripe_customer_id VARCHAR(255),
    stripe_session_id VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, cancelled, past_due, unpaid, incomplete
    current_period_start TIMESTAMP,
    current_period_end TIMESTAMP,
    cancel_at_period_end BOOLEAN DEFAULT FALSE,
    cancelled_at TIMESTAMP,
    trial_start TIMESTAMP,
    trial_end TIMESTAMP,
    amount_paid DECIMAL(10,2),
    currency VARCHAR(3) DEFAULT 'USD',
    payment_method VARCHAR(100),
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create email_verification_tokens table
CREATE TABLE IF NOT EXISTS email_verification_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    verified_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create email_notifications table to track sent emails
CREATE TABLE IF NOT EXISTS email_notifications (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    email_to VARCHAR(255) NOT NULL,
    email_type VARCHAR(100) NOT NULL, -- verification, subscription_confirmation, password_reset, etc.
    subject VARCHAR(500) NOT NULL,
    template_name VARCHAR(100),
    template_data JSONB,
    status VARCHAR(50) DEFAULT 'pending', -- pending, sent, failed, delivered, bounced
    sent_at TIMESTAMP,
    delivered_at TIMESTAMP,
    error_message TEXT,
    provider VARCHAR(50), -- smtp, sendgrid, ses, etc.
    provider_message_id VARCHAR(255),
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add email_verified column to users table if it doesn't exist
DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'users' AND column_name = 'email_verified') THEN
        ALTER TABLE users ADD COLUMN email_verified BOOLEAN DEFAULT FALSE;
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'users' AND column_name = 'email_verified_at') THEN
        ALTER TABLE users ADD COLUMN email_verified_at TIMESTAMP;
    END IF;
END $$;

-- Create email_settings table for SMTP configuration
CREATE TABLE IF NOT EXISTS email_settings (
    id SERIAL PRIMARY KEY,
    setting_key VARCHAR(100) NOT NULL UNIQUE,
    setting_value TEXT,
    is_encrypted BOOLEAN DEFAULT FALSE,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert default email settings
INSERT INTO email_settings (setting_key, setting_value, is_encrypted, description) VALUES
('smtp_host', '', FALSE, 'SMTP server hostname'),
('smtp_port', '587', FALSE, 'SMTP server port'),
('smtp_username', '', FALSE, 'SMTP username'),
('smtp_password', '', TRUE, 'SMTP password (encrypted)'),
('smtp_from_email', '', FALSE, 'Default from email address'),
('smtp_from_name', 'BOME', FALSE, 'Default from name'),
('email_enabled', 'false', FALSE, 'Whether email sending is enabled'),
('verification_required', 'true', FALSE, 'Whether email verification is required for new accounts')
ON CONFLICT (setting_key) DO NOTHING;

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
DROP TRIGGER IF EXISTS update_user_subscriptions_updated_at ON user_subscriptions;
CREATE TRIGGER update_user_subscriptions_updated_at 
    BEFORE UPDATE ON user_subscriptions 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_email_notifications_updated_at ON email_notifications;
CREATE TRIGGER update_email_notifications_updated_at 
    BEFORE UPDATE ON email_notifications 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_email_settings_updated_at ON email_settings;
CREATE TRIGGER update_email_settings_updated_at 
    BEFORE UPDATE ON email_settings 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_user_subscriptions_user_id ON user_subscriptions (user_id);
CREATE INDEX IF NOT EXISTS idx_user_subscriptions_stripe_subscription_id ON user_subscriptions (stripe_subscription_id);
CREATE INDEX IF NOT EXISTS idx_user_subscriptions_status ON user_subscriptions (status);
CREATE INDEX IF NOT EXISTS idx_user_subscriptions_current_period_end ON user_subscriptions (current_period_end);

CREATE INDEX IF NOT EXISTS idx_email_verification_tokens_token ON email_verification_tokens (token);
CREATE INDEX IF NOT EXISTS idx_email_verification_tokens_user_id ON email_verification_tokens (user_id);
CREATE INDEX IF NOT EXISTS idx_email_verification_tokens_expires_at ON email_verification_tokens (expires_at);

CREATE INDEX IF NOT EXISTS idx_email_notifications_user_id ON email_notifications (user_id);
CREATE INDEX IF NOT EXISTS idx_email_notifications_email_type ON email_notifications (email_type);
CREATE INDEX IF NOT EXISTS idx_email_notifications_status ON email_notifications (status);
CREATE INDEX IF NOT EXISTS idx_email_notifications_created_at ON email_notifications (created_at);

-- Create daily_email_usage table for tracking email provider usage
CREATE TABLE IF NOT EXISTS daily_email_usage (
    id SERIAL PRIMARY KEY,
    date DATE NOT NULL DEFAULT CURRENT_DATE,
    provider VARCHAR(50) NOT NULL, -- 'resend' or 'mailgun'
    emails_sent INTEGER DEFAULT 0,
    emails_failed INTEGER DEFAULT 0,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(date, provider)
);

-- Create index for fast lookups
CREATE INDEX IF NOT EXISTS idx_daily_email_usage_date_provider ON daily_email_usage(date, provider);
CREATE INDEX IF NOT EXISTS idx_daily_email_usage_date ON daily_email_usage(date);
CREATE INDEX IF NOT EXISTS idx_daily_email_usage_provider ON daily_email_usage(provider);

-- Insert hybrid email system configuration
INSERT INTO email_settings (setting_key, setting_value, is_encrypted) VALUES
('email_provider_primary', 'resend', false),
('email_provider_secondary', 'mailgun', false),
('daily_email_limit_resend', '100', false),
('daily_email_limit_mailgun', '100', false),
('auto_failover_enabled', 'true', false),
('email_enabled', 'true', false),
('smtp_from_name', 'BOME Support', false),
('smtp_from_email', 'support@yourdomain.com', false)
ON CONFLICT (setting_key) DO UPDATE SET 
    setting_value = EXCLUDED.setting_value,
    updated_at = CURRENT_TIMESTAMP;

-- Insert encrypted API keys from environment variables (these will be set via backend code)
-- The backend will read EMAIL_RSND_100 and EMAIL_MG_100 from .env and store them encrypted
