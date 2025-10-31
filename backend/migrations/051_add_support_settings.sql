-- Migration 051: Add Support Settings to public_settings Table
-- Purpose: Add support contact settings (email, phone, URL, hours, message)
-- Date: 2025-10-30

-- Insert default support settings into existing public_settings table
INSERT INTO public_settings (key, value, created_at, updated_at) VALUES
('support_email', NULL, NOW(), NOW()),
('support_phone', NULL, NOW(), NOW()),
('support_url', NULL, NOW(), NOW()),
('support_hours', NULL, NOW(), NOW()),
('support_message', 'Please contact our support team for assistance.', NOW(), NOW())
ON CONFLICT (key) DO NOTHING;

-- Add comment
COMMENT ON TABLE public_settings IS 'Application-wide public configuration settings (including support contact info)';

