-- Migration: Create OAuth2 states table for persistent state storage
-- This replaces in-memory state storage to fix production OAuth2 issues

CREATE TABLE IF NOT EXISTS oauth2_states (
    id SERIAL PRIMARY KEY,
    state VARCHAR(255) UNIQUE NOT NULL,
    provider VARCHAR(50) NOT NULL,
    return_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    state_data JSONB,
    
    -- Indexes for performance
    INDEX idx_oauth2_states_state (state),
    INDEX idx_oauth2_states_expires_at (expires_at)
);

-- Add comment
COMMENT ON TABLE oauth2_states IS 'Stores OAuth2 state parameters for CSRF protection and production persistence';
COMMENT ON COLUMN oauth2_states.state IS 'Unique state parameter for OAuth2 flow';
COMMENT ON COLUMN oauth2_states.provider IS 'OAuth2 provider name (google, etc.)';
COMMENT ON COLUMN oauth2_states.return_url IS 'URL to redirect to after OAuth2 completion';
COMMENT ON COLUMN oauth2_states.expires_at IS 'When this state expires (typically 10 minutes)';
COMMENT ON COLUMN oauth2_states.state_data IS 'Additional state data as JSON';
