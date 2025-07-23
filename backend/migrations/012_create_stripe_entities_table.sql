-- Migration: Create Stripe Entities Table
-- Description: Creates abstraction table for Stripe entities to enhance security
-- Version: 012
-- Date: 2024-12-19

-- Create stripe_entities table
CREATE TABLE IF NOT EXISTS stripe_entities (
    id SERIAL PRIMARY KEY,
    entity_type VARCHAR(50) NOT NULL CHECK (entity_type IN ('subscription', 'customer', 'payment_intent', 'price', 'product', 'invoice', 'refund')),
    entity_id VARCHAR(255) NOT NULL, -- Stripe ID
    local_id INTEGER NOT NULL, -- Your internal ID
    metadata JSONB, -- Flexible metadata storage
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for entity_type and entity_id lookups
CREATE INDEX IF NOT EXISTS idx_stripe_entities_entity_type ON stripe_entities(entity_type);
CREATE INDEX IF NOT EXISTS idx_stripe_entities_entity_id ON stripe_entities(entity_id);
CREATE INDEX IF NOT EXISTS idx_stripe_entities_local_id ON stripe_entities(local_id);
CREATE INDEX IF NOT EXISTS idx_stripe_entities_type_and_id ON stripe_entities(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_stripe_entities_type_and_local_id ON stripe_entities(entity_type, local_id);
CREATE INDEX IF NOT EXISTS idx_stripe_entities_created_at ON stripe_entities(created_at);

-- Create unique constraint to prevent duplicate mappings
CREATE UNIQUE INDEX IF NOT EXISTS idx_stripe_entities_unique_mapping ON stripe_entities(entity_type, entity_id);

-- Add comments for documentation
COMMENT ON TABLE stripe_entities IS 'Abstraction layer for Stripe entities to enhance security';
COMMENT ON COLUMN stripe_entities.id IS 'Primary key for stripe entity mapping';
COMMENT ON COLUMN stripe_entities.entity_type IS 'Type of Stripe entity (subscription, customer, payment_intent, etc.)';
COMMENT ON COLUMN stripe_entities.entity_id IS 'Stripe ID for the entity';
COMMENT ON COLUMN stripe_entities.local_id IS 'Internal ID in our system';
COMMENT ON COLUMN stripe_entities.metadata IS 'JSON metadata for additional information';
COMMENT ON COLUMN stripe_entities.created_at IS 'When the mapping was created';
COMMENT ON COLUMN stripe_entities.updated_at IS 'When the mapping was last updated';

-- Create a function to automatically update the updated_at timestamp
CREATE OR REPLACE FUNCTION update_stripe_entities_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to automatically update updated_at
CREATE TRIGGER trigger_update_stripe_entities_updated_at
    BEFORE UPDATE ON stripe_entities
    FOR EACH ROW
    EXECUTE FUNCTION update_stripe_entities_updated_at(); 