-- BOME Database Initialization Script
-- This script runs automatically when the PostgreSQL container starts

-- Create extensions if they don't exist
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Set timezone
SET timezone = 'UTC';

-- Create custom types if they don't exist
DO $$ BEGIN
    CREATE TYPE user_role AS ENUM (
        'user',
        'subscriber',
        'content_editor',
        'events_manager',
        'admin',
        'super_admin',
        'system_admin'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE subscription_status AS ENUM (
        'active',
        'inactive',
        'suspended',
        'cancelled',
        'pending'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE video_status AS ENUM (
        'draft',
        'published',
        'archived',
        'processing',
        'error'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Log initialization
DO $$
BEGIN
    RAISE NOTICE 'BOME database initialized successfully at %', NOW();
END $$;
