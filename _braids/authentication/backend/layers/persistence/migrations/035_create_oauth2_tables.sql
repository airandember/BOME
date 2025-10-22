-- OAuth2 Integration Migration
-- Creates tables for OAuth2 authentication and account linking

-- Create oauth2_accounts table for linking OAuth2 accounts to users
CREATE TABLE IF NOT EXISTS oauth2_accounts (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL, -- 'google', 'facebook', 'github', etc.
    provider_user_id VARCHAR(255) NOT NULL, -- OAuth2 provider's user ID
    email VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    picture TEXT, -- Profile picture URL
    access_token TEXT, -- Encrypted OAuth2 access token (optional)
    refresh_token TEXT, -- Encrypted OAuth2 refresh token (optional)
    token_expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, provider), -- One account per provider per user
    UNIQUE(provider, provider_user_id) -- One provider account per user ID
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_oauth2_accounts_user_id ON oauth2_accounts(user_id);
CREATE INDEX IF NOT EXISTS idx_oauth2_accounts_provider ON oauth2_accounts(provider);
CREATE INDEX IF NOT EXISTS idx_oauth2_accounts_provider_user_id ON oauth2_accounts(provider, provider_user_id);
CREATE INDEX IF NOT EXISTS idx_oauth2_accounts_email ON oauth2_accounts(email);

-- Create oauth2_settings table for OAuth2 provider configurations
CREATE TABLE IF NOT EXISTS oauth2_settings (
    id SERIAL PRIMARY KEY,
    provider VARCHAR(50) NOT NULL UNIQUE, -- 'google', 'facebook', etc.
    client_id VARCHAR(255) NOT NULL,
    client_secret TEXT NOT NULL, -- Encrypted
    redirect_url VARCHAR(500) NOT NULL,
    auth_url VARCHAR(500),
    token_url VARCHAR(500),
    user_info_url VARCHAR(500),
    scopes TEXT, -- JSON array of scopes
    is_enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create index for oauth2_settings
CREATE INDEX IF NOT EXISTS idx_oauth2_settings_provider ON oauth2_settings(provider);
CREATE INDEX IF NOT EXISTS idx_oauth2_settings_enabled ON oauth2_settings(is_enabled);

-- Insert default OAuth2 provider configurations
INSERT INTO oauth2_settings (provider, client_id, client_secret, redirect_url, auth_url, token_url, user_info_url, scopes, is_enabled) VALUES
('google', 
 'your-google-client-id', 
 'your-encrypted-google-client-secret', 
 'http://localhost:5173/auth/callback/google',
 'https://accounts.google.com/o/oauth2/auth',
 'https://oauth2.googleapis.com/token',
 'https://www.googleapis.com/oauth2/v2/userinfo',
 '["https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"]',
 false) -- Disabled by default until configured
ON CONFLICT (provider) DO UPDATE SET
    redirect_url = EXCLUDED.redirect_url,
    auth_url = EXCLUDED.auth_url,
    token_url = EXCLUDED.token_url,
    user_info_url = EXCLUDED.user_info_url,
    scopes = EXCLUDED.scopes,
    updated_at = CURRENT_TIMESTAMP;

-- Add OAuth2 settings to email_settings table for easier management
INSERT INTO email_settings (setting_key, setting_value, is_encrypted) VALUES
('google_oauth_enabled', 'false', false),
('google_oauth_client_id', '', false),
('google_oauth_client_secret', '', true),
('google_oauth_redirect_url', 'http://localhost:5173/auth/callback/google', false),
('oauth2_state_ttl_minutes', '10', false),
('oauth2_auto_link_accounts', 'true', false),
('oauth2_auto_verify_email', 'true', false)
ON CONFLICT (setting_key) DO UPDATE SET 
    setting_value = EXCLUDED.setting_value,
    updated_at = CURRENT_TIMESTAMP;

-- Create trigger for updated_at on oauth2_accounts
CREATE OR REPLACE FUNCTION update_oauth2_accounts_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS update_oauth2_accounts_updated_at ON oauth2_accounts;
CREATE TRIGGER update_oauth2_accounts_updated_at 
    BEFORE UPDATE ON oauth2_accounts 
    FOR EACH ROW EXECUTE FUNCTION update_oauth2_accounts_updated_at();

-- Create trigger for updated_at on oauth2_settings
DROP TRIGGER IF EXISTS update_oauth2_settings_updated_at ON oauth2_settings;
CREATE TRIGGER update_oauth2_settings_updated_at 
    BEFORE UPDATE ON oauth2_settings 
    FOR EACH ROW EXECUTE FUNCTION update_oauth2_accounts_updated_at();

-- Add comments for documentation
COMMENT ON TABLE oauth2_accounts IS 'Links OAuth2 provider accounts to local user accounts';
COMMENT ON COLUMN oauth2_accounts.user_id IS 'Foreign key to users table';
COMMENT ON COLUMN oauth2_accounts.provider IS 'OAuth2 provider name (google, facebook, etc.)';
COMMENT ON COLUMN oauth2_accounts.provider_user_id IS 'User ID from the OAuth2 provider';
COMMENT ON COLUMN oauth2_accounts.access_token IS 'Encrypted OAuth2 access token (optional storage)';
COMMENT ON COLUMN oauth2_accounts.refresh_token IS 'Encrypted OAuth2 refresh token (optional storage)';

COMMENT ON TABLE oauth2_settings IS 'Configuration for OAuth2 providers';
COMMENT ON COLUMN oauth2_settings.client_id IS 'OAuth2 client ID from provider';
COMMENT ON COLUMN oauth2_settings.client_secret IS 'Encrypted OAuth2 client secret';
COMMENT ON COLUMN oauth2_settings.scopes IS 'JSON array of OAuth2 scopes to request';
