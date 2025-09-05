-- Fix OAuth2 credentials in database
-- Replace the API key with proper OAuth2 Client ID and Secret

-- Update Google OAuth2 Client ID (replace YOUR_ACTUAL_CLIENT_ID with the real one from Google Console)
UPDATE email_settings 
SET setting_value = 'YOUR_ACTUAL_CLIENT_ID_HERE.apps.googleusercontent.com'
WHERE setting_key = 'google_oauth_client_id';

-- Update Google OAuth2 Client Secret (replace YOUR_ACTUAL_CLIENT_SECRET with the real one from Google Console)
UPDATE email_settings 
SET setting_value = 'YOUR_ACTUAL_CLIENT_SECRET_HERE'
WHERE setting_key = 'google_oauth_client_secret';

-- Verify the settings
SELECT setting_key, setting_value FROM email_settings 
WHERE setting_key IN ('google_oauth_client_id', 'google_oauth_client_secret');
