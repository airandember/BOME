-- Migration 005: Add user sessions management
-- This migration adds session tracking capabilities

-- Add session management fields to users table
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS last_logout TIMESTAMP,
ADD COLUMN IF NOT EXISTS max_sessions INTEGER DEFAULT 5;

-- Create user_sessions table
CREATE TABLE IF NOT EXISTS user_sessions (
    id SERIAL PRIMARY KEY,
    session_id VARCHAR(255) UNIQUE NOT NULL,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_id VARCHAR(255) NOT NULL,
    device_info TEXT,
    ip_address INET,
    user_agent TEXT,
    last_activity TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    
    -- Indexes for performance
    INDEX idx_user_sessions_user_id (user_id),
    INDEX idx_user_sessions_token_id (token_id),
    INDEX idx_user_sessions_active (is_active),
    INDEX idx_user_sessions_expires (expires_at),
    INDEX idx_user_sessions_activity (last_activity)
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_token_id ON user_sessions(token_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_active ON user_sessions(is_active);
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires ON user_sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_user_sessions_activity ON user_sessions(last_activity);

-- Add constraint to ensure session_id is unique
ALTER TABLE user_sessions ADD CONSTRAINT unique_session_id UNIQUE (session_id);

-- Add constraint to ensure token_id is unique per user
ALTER TABLE user_sessions ADD CONSTRAINT unique_user_token UNIQUE (user_id, token_id);

-- Create function to cleanup expired sessions
CREATE OR REPLACE FUNCTION cleanup_expired_sessions()
RETURNS void AS $$
BEGIN
    DELETE FROM user_sessions WHERE expires_at < NOW();
END;
$$ LANGUAGE plpgsql;

-- Create a scheduled job to cleanup expired sessions (if using pg_cron)
-- SELECT cron.schedule('cleanup-expired-sessions', '0 */6 * * *', 'SELECT cleanup_expired_sessions();');

-- Update existing users to have default max_sessions
UPDATE users SET max_sessions = 5 WHERE max_sessions IS NULL;

-- Add comment to document the table
COMMENT ON TABLE user_sessions IS 'Stores active user sessions for device tracking and security';
COMMENT ON COLUMN user_sessions.session_id IS 'Unique session identifier';
COMMENT ON COLUMN user_sessions.token_id IS 'JWT token ID for blacklisting';
COMMENT ON COLUMN user_sessions.device_info IS 'Device fingerprint information';
COMMENT ON COLUMN user_sessions.ip_address IS 'IP address of the session';
COMMENT ON COLUMN user_sessions.user_agent IS 'User agent string';
COMMENT ON COLUMN user_sessions.last_activity IS 'Last activity timestamp';
COMMENT ON COLUMN user_sessions.is_active IS 'Whether the session is currently active';
COMMENT ON COLUMN user_sessions.expires_at IS 'Session expiration timestamp'; 