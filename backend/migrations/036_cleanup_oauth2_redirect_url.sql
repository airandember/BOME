-- Migration 036: Clean up OAuth2 redirect URL from database
-- We now use OAUTH2_REDIRECT_URL environment variable instead

-- Remove the conflicting redirect URL setting from email_settings
DELETE FROM email_settings WHERE setting_key = 'google_oauth_redirect_url';

-- Add a comment to track this change
INSERT INTO email_settings (setting_key, setting_value, is_encrypted) VALUES
('oauth2_config_note', 'OAuth2 redirect URLs now managed via OAUTH2_REDIRECT_URL environment variable', false)
ON CONFLICT (setting_key) DO UPDATE SET 
    setting_value = EXCLUDED.setting_value,
    updated_at = CURRENT_TIMESTAMP;
