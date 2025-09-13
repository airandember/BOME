-- Fix email domain settings to use bookofmormonevidence.org instead of bome.test
-- Run this directly in your production database

-- Check current settings
SELECT setting_key, setting_value FROM email_settings 
WHERE setting_key IN ('smtp_from_email', 'smtp_from_name', 'mailgun_from_email', 'mailgun_from_name');

-- Update SMTP settings (used by Resend)
INSERT INTO email_settings (setting_key, setting_value, is_encrypted, created_at, updated_at)
VALUES ('smtp_from_email', 'noreply@bookofmormonevidence.org', false, NOW(), NOW())
ON CONFLICT (setting_key) 
DO UPDATE SET 
    setting_value = 'noreply@bookofmormonevidence.org',
    updated_at = NOW();

INSERT INTO email_settings (setting_key, setting_value, is_encrypted, created_at, updated_at)
VALUES ('smtp_from_name', 'Book of Mormon Evidence', false, NOW(), NOW())
ON CONFLICT (setting_key) 
DO UPDATE SET 
    setting_value = 'Book of Mormon Evidence',
    updated_at = NOW();

-- Update Mailgun settings
INSERT INTO email_settings (setting_key, setting_value, is_encrypted, created_at, updated_at)
VALUES ('mailgun_from_email', 'noreply@bookofmormonevidence.org', false, NOW(), NOW())
ON CONFLICT (setting_key) 
DO UPDATE SET 
    setting_value = 'noreply@bookofmormonevidence.org',
    updated_at = NOW();

INSERT INTO email_settings (setting_key, setting_value, is_encrypted, created_at, updated_at)
VALUES ('mailgun_from_name', 'Book of Mormon Evidence', false, NOW(), NOW())
ON CONFLICT (setting_key) 
DO UPDATE SET 
    setting_value = 'Book of Mormon Evidence',
    updated_at = NOW();

-- Fix the typo in support_email (boook -> book)
UPDATE email_settings 
SET setting_value = 'jake@bookofmormonevidence.org', updated_at = NOW()
WHERE setting_key = 'support_email' AND setting_value = 'jake@boookofmormonevidence.org';

-- Verify the changes
SELECT setting_key, setting_value FROM email_settings 
WHERE setting_key IN ('smtp_from_email', 'smtp_from_name', 'mailgun_from_email', 'mailgun_from_name', 'support_email')
ORDER BY setting_key;
