-- Migration 006: Add audit logging system
-- This migration creates comprehensive audit logging for security events

-- Create audit_logs table
CREATE TABLE IF NOT EXISTS audit_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    user_email VARCHAR(255),
    action VARCHAR(100) NOT NULL,
    resource VARCHAR(100) NOT NULL,
    resource_id VARCHAR(255),
    ip_address INET NOT NULL,
    user_agent TEXT,
    status VARCHAR(20) NOT NULL CHECK (status IN ('success', 'failed', 'warning')),
    details TEXT,
    metadata JSONB DEFAULT '{}',
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')) DEFAULT 'low',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_status ON audit_logs(status);
CREATE INDEX IF NOT EXISTS idx_audit_logs_severity ON audit_logs(severity);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_audit_logs_ip_address ON audit_logs(ip_address);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_email ON audit_logs(user_email);

-- Create composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_action ON audit_logs(user_id, action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action_status ON audit_logs(action, status);
CREATE INDEX IF NOT EXISTS idx_audit_logs_severity_created ON audit_logs(severity, created_at);

-- Create function to automatically set user_email from user_id
CREATE OR REPLACE FUNCTION set_audit_user_email()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.user_id IS NOT NULL AND NEW.user_email IS NULL THEN
        SELECT email INTO NEW.user_email FROM users WHERE id = NEW.user_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to automatically set user_email
CREATE TRIGGER trigger_set_audit_user_email
    BEFORE INSERT ON audit_logs
    FOR EACH ROW
    EXECUTE FUNCTION set_audit_user_email();

-- Create function to cleanup old audit logs
CREATE OR REPLACE FUNCTION cleanup_old_audit_logs(days_to_keep INTEGER DEFAULT 90)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM audit_logs 
    WHERE created_at < NOW() - INTERVAL '1 day' * days_to_keep;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Create a scheduled job to cleanup old audit logs (if using pg_cron)
-- SELECT cron.schedule('cleanup-old-audit-logs', '0 2 * * 0', 'SELECT cleanup_old_audit_logs(90);');

-- Add comments for documentation
COMMENT ON TABLE audit_logs IS 'Comprehensive audit logging for security events and user activities';
COMMENT ON COLUMN audit_logs.user_id IS 'User ID (nullable for anonymous actions)';
COMMENT ON COLUMN audit_logs.user_email IS 'User email (auto-populated from user_id)';
COMMENT ON COLUMN audit_logs.action IS 'Action performed (login, logout, create, update, delete, etc.)';
COMMENT ON COLUMN audit_logs.resource IS 'Resource affected (user, video, subscription, etc.)';
COMMENT ON COLUMN audit_logs.resource_id IS 'Specific resource identifier';
COMMENT ON COLUMN audit_logs.ip_address IS 'IP address of the request';
COMMENT ON COLUMN audit_logs.user_agent IS 'User agent string';
COMMENT ON COLUMN audit_logs.status IS 'Action status: success, failed, warning';
COMMENT ON COLUMN audit_logs.details IS 'Additional details about the action';
COMMENT ON COLUMN audit_logs.metadata IS 'JSON metadata for additional context';
COMMENT ON COLUMN audit_logs.severity IS 'Security severity: low, medium, high, critical';
COMMENT ON COLUMN audit_logs.created_at IS 'Timestamp when the audit log was created';

-- Create view for security metrics
CREATE OR REPLACE VIEW security_metrics AS
SELECT 
    COUNT(*) as total_events,
    COUNT(*) FILTER (WHERE status = 'failed') as failed_events,
    COUNT(*) FILTER (WHERE severity IN ('high', 'critical')) as high_severity_events,
    COUNT(*) FILTER (WHERE action = 'login' AND status = 'failed') as failed_logins,
    COUNT(*) FILTER (WHERE action = 'login' AND status = 'success') as successful_logins,
    COUNT(DISTINCT user_id) FILTER (WHERE user_id IS NOT NULL) as unique_users,
    COUNT(DISTINCT ip_address) as unique_ips
FROM audit_logs 
WHERE created_at > NOW() - INTERVAL '24 hours'; 