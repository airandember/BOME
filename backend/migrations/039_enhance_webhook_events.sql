-- Migration: Enhance Webhook Events Table
-- Description: Adds detailed columns for storing Stripe webhook data and user identification

-- Add new columns for enhanced webhook logging
ALTER TABLE webhook_events 
ADD COLUMN IF NOT EXISTS stripe_event_id VARCHAR(255),  -- evt_xxx from Stripe
ADD COLUMN IF NOT EXISTS stripe_object_id VARCHAR(255), -- sub_xxx, cus_xxx, in_xxx, etc.
ADD COLUMN IF NOT EXISTS stripe_object_type VARCHAR(50), -- subscription, customer, invoice, etc.
ADD COLUMN IF NOT EXISTS user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
ADD COLUMN IF NOT EXISTS user_email VARCHAR(255),
ADD COLUMN IF NOT EXISTS customer_id VARCHAR(255), -- Stripe customer ID (cus_xxx)
ADD COLUMN IF NOT EXISTS subscription_id VARCHAR(255), -- Stripe subscription ID (sub_xxx)
ADD COLUMN IF NOT EXISTS invoice_id VARCHAR(255), -- Stripe invoice ID (in_xxx)
ADD COLUMN IF NOT EXISTS amount_cents INTEGER, -- Amount in cents (for invoice/payment events)
ADD COLUMN IF NOT EXISTS currency VARCHAR(3), -- USD, EUR, etc.
ADD COLUMN IF NOT EXISTS subscription_status VARCHAR(50), -- active, canceled, unpaid, etc.
ADD COLUMN IF NOT EXISTS payment_status VARCHAR(50), -- paid, unpaid, pending, etc.
ADD COLUMN IF NOT EXISTS event_data JSONB, -- Full webhook payload for detailed inspection
ADD COLUMN IF NOT EXISTS api_version VARCHAR(50), -- Stripe API version
ADD COLUMN IF NOT EXISTS livemode BOOLEAN DEFAULT true, -- true for production, false for test
ADD COLUMN IF NOT EXISTS description TEXT; -- Human-readable event description

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_webhook_events_stripe_event_id ON webhook_events(stripe_event_id);
CREATE INDEX IF NOT EXISTS idx_webhook_events_stripe_object_id ON webhook_events(stripe_object_id);
CREATE INDEX IF NOT EXISTS idx_webhook_events_user_id ON webhook_events(user_id);
CREATE INDEX IF NOT EXISTS idx_webhook_events_user_email ON webhook_events(user_email);
CREATE INDEX IF NOT EXISTS idx_webhook_events_customer_id ON webhook_events(customer_id);
CREATE INDEX IF NOT EXISTS idx_webhook_events_subscription_id ON webhook_events(subscription_id);
CREATE INDEX IF NOT EXISTS idx_webhook_events_invoice_id ON webhook_events(invoice_id);
CREATE INDEX IF NOT EXISTS idx_webhook_events_livemode ON webhook_events(livemode);

-- Add comment to describe the table
COMMENT ON TABLE webhook_events IS 'Enhanced webhook event logging with detailed Stripe data and user tracking';
COMMENT ON COLUMN webhook_events.stripe_event_id IS 'Unique Stripe event ID (evt_xxx)';
COMMENT ON COLUMN webhook_events.stripe_object_id IS 'ID of the Stripe object affected (sub_xxx, cus_xxx, etc.)';
COMMENT ON COLUMN webhook_events.stripe_object_type IS 'Type of Stripe object (subscription, customer, invoice, etc.)';
COMMENT ON COLUMN webhook_events.user_id IS 'BOME user ID associated with the webhook event';
COMMENT ON COLUMN webhook_events.user_email IS 'User email (from Stripe or BOME user)';
COMMENT ON COLUMN webhook_events.event_data IS 'Full webhook payload as JSON for detailed inspection';
COMMENT ON COLUMN webhook_events.description IS 'Human-readable description (e.g., "user@example.com subscription changed to active")';

