-- Migration 006: Standardize Roles System
-- This migration updates the role system to use standardized role names and adds role hierarchy

-- Create roles table if it doesn't exist
CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    role_id VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    category VARCHAR(50) NOT NULL,
    level INTEGER NOT NULL DEFAULT 1,
    permissions JSONB DEFAULT '[]',
    is_system_role BOOLEAN DEFAULT FALSE,
    color VARCHAR(7) DEFAULT '#6b7280',
    icon VARCHAR(50),
    subsystem_access JSONB DEFAULT '[]',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create user_roles junction table for many-to-many relationship
CREATE TABLE IF NOT EXISTS user_roles (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    role_id VARCHAR(100) REFERENCES roles(role_id) ON DELETE CASCADE,
    assigned_by INTEGER REFERENCES users(id),
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    UNIQUE(user_id, role_id)
);

-- Create role_permissions junction table
CREATE TABLE IF NOT EXISTS role_permissions (
    id SERIAL PRIMARY KEY,
    role_id VARCHAR(100) REFERENCES roles(role_id) ON DELETE CASCADE,
    permission_id VARCHAR(100) NOT NULL,
    granted_by INTEGER REFERENCES users(id),
    granted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(role_id, permission_id)
);

-- Insert standardized roles
INSERT INTO roles (role_id, name, slug, description, category, level, permissions, is_system_role, color, icon, subsystem_access) VALUES
-- System Administration (Level 10-9)
('super_admin', 'Super Administrator', 'super-administrator', 'Full system access and role management capabilities', 'system', 10, '[]', true, '#dc2626', 'crown', '["hub", "articles", "youtube", "streaming", "events"]'),
('system_admin', 'System Administrator', 'system-administrator', 'Technical system management without role changes', 'system', 9, '[]', true, '#7c3aed', 'server', '["hub", "articles", "youtube", "streaming", "events"]'),

-- Content Management (Level 8-6)
('content_manager', 'Content Manager', 'content-manager', 'Overall content strategy and oversight', 'content', 8, '[]', true, '#059669', 'document-text', '["articles", "youtube", "streaming"]'),
('content_editor', 'Content Editor', 'content-editor', 'Review, approve, edit, and publish content', 'content', 7, '[]', false, '#059669', 'pencil', '["articles", "youtube", "streaming"]'),
('content_creator', 'Content Creator', 'content-creator', 'Create and edit content with limited publishing', 'content', 6, '[]', false, '#059669', 'plus-circle', '["articles", "youtube", "streaming"]'),

-- Subsystem-Specific Roles (Level 7)
('articles_manager', 'Articles Manager', 'articles-manager', 'Full articles subsystem management', 'subsystem', 7, '[]', false, '#1e40af', 'document', '["articles"]'),
('youtube_manager', 'YouTube Manager', 'youtube-manager', 'YouTube system management', 'subsystem', 7, '[]', false, '#dc2626', 'play', '["youtube"]'),
('streaming_manager', 'Video Streaming Manager', 'streaming-manager', 'Bunny.net streaming platform management', 'subsystem', 7, '[]', false, '#7c3aed', 'video', '["streaming"]'),
('events_manager', 'Events Manager', 'events-manager', 'Events system management', 'subsystem', 7, '[]', false, '#2563eb', 'calendar', '["events"]'),

-- Marketing & Advertising (Level 7-4)
('advertisement_manager', 'Advertisement Manager', 'advertisement-manager', 'Full advertisement system oversight', 'marketing', 7, '[]', false, '#f59e0b', 'presentation-chart-line', '["hub"]'),
('marketing_specialist', 'Marketing Specialist', 'marketing-specialist', 'Campaign creation and advertiser relations', 'marketing', 4, '[]', false, '#f59e0b', 'megaphone', '["hub"]'),

-- User Management (Level 7-5)
('user_manager', 'User Account Manager', 'user-manager', 'User management and support operations', 'user_management', 7, '[]', false, '#2563eb', 'users', '["hub"]'),
('support_specialist', 'Support Specialist', 'support-specialist', 'User support and basic account management', 'user_management', 5, '[]', false, '#2563eb', 'life-buoy', '["hub"]'),

-- Analytics & Financial (Level 7)
('analytics_manager', 'Analytics Manager', 'analytics-manager', 'Data analysis and reporting across all systems', 'analytics', 7, '[]', false, '#059669', 'chart-bar', '["hub", "articles", "youtube", "streaming", "events"]'),
('financial_admin', 'Financial Administrator', 'financial-administrator', 'Revenue, billing, and financial reporting', 'financial', 7, '[]', false, '#059669', 'credit-card', '["hub"]'),

-- Technical & Security (Level 6-5)
('security_admin', 'Security Administrator', 'security-administrator', 'Security monitoring and incident response', 'security', 6, '[]', false, '#dc2626', 'shield', '["hub", "articles", "youtube", "streaming", "events"]'),
('technical_specialist', 'Technical Specialist', 'technical-specialist', 'Technical support and maintenance', 'technical', 5, '[]', false, '#7c3aed', 'wrench', '["hub", "articles", "youtube", "streaming", "events"]'),

-- Academic & Research (Level 6-5)
('academic_reviewer', 'Academic Reviewer', 'academic-reviewer', 'Review scholarly content for accuracy and quality', 'academic', 6, '[]', false, '#7c2d12', 'academic-cap', '["articles"]'),
('research_coordinator', 'Research Coordinator', 'research-coordinator', 'Coordinate academic research and citations', 'academic', 5, '[]', false, '#7c2d12', 'book-open', '["articles"]'),

-- Base User Roles (Level 3-1)
('advertiser', 'Advertiser', 'advertiser', 'Create and manage advertising campaigns', 'base', 3, '[]', false, '#f59e0b', 'megaphone', '["hub"]'),
('user', 'User', 'user', 'Basic platform access', 'base', 1, '[]', true, '#6b7280', 'user', '["hub", "articles", "youtube", "streaming", "events"]')
ON CONFLICT (role_id) DO NOTHING;

-- Insert standardized permissions
INSERT INTO role_permissions (role_id, permission_id) VALUES
-- Super Administrator gets all permissions
('super_admin', 'system:full_access'),

-- System Administrator permissions
('system_admin', 'system:read'),
('system_admin', 'system:update'),
('system_admin', 'system:manage'),
('system_admin', 'security:read'),
('system_admin', 'security:manage'),
('system_admin', 'technical:read'),
('system_admin', 'technical:manage'),
('system_admin', 'analytics:read'),
('system_admin', 'analytics:export'),

-- Content Manager permissions
('content_manager', 'content:create'),
('content_manager', 'content:read'),
('content_manager', 'content:update'),
('content_manager', 'content:delete'),
('content_manager', 'content:publish'),
('content_manager', 'content:moderate'),
('content_manager', 'videos:create'),
('content_manager', 'videos:read'),
('content_manager', 'videos:update'),
('content_manager', 'videos:delete'),
('content_manager', 'videos:manage'),
('content_manager', 'articles:create'),
('content_manager', 'articles:read'),
('content_manager', 'articles:update'),
('content_manager', 'articles:delete'),
('content_manager', 'articles:publish'),
('content_manager', 'analytics:read'),
('content_manager', 'analytics:export'),

-- Content Editor permissions
('content_editor', 'content:read'),
('content_editor', 'content:update'),
('content_editor', 'content:publish'),
('content_editor', 'videos:read'),
('content_editor', 'videos:update'),
('content_editor', 'articles:read'),
('content_editor', 'articles:update'),
('content_editor', 'articles:publish'),
('content_editor', 'analytics:read'),

-- Content Creator permissions
('content_creator', 'content:create'),
('content_creator', 'content:read'),
('content_creator', 'content:update'),
('content_creator', 'videos:create'),
('content_creator', 'videos:read'),
('content_creator', 'videos:update'),
('content_creator', 'articles:create'),
('content_creator', 'articles:read'),
('content_creator', 'articles:update'),

-- Articles Manager permissions
('articles_manager', 'articles:create'),
('articles_manager', 'articles:read'),
('articles_manager', 'articles:update'),
('articles_manager', 'articles:delete'),
('articles_manager', 'articles:publish'),
('articles_manager', 'articles:manage'),
('articles_manager', 'content:read'),
('articles_manager', 'analytics:read'),

-- YouTube Manager permissions
('youtube_manager', 'videos:create'),
('youtube_manager', 'videos:read'),
('youtube_manager', 'videos:update'),
('youtube_manager', 'videos:delete'),
('youtube_manager', 'videos:manage'),
('youtube_manager', 'content:read'),
('youtube_manager', 'analytics:read'),

-- Streaming Manager permissions
('streaming_manager', 'videos:create'),
('streaming_manager', 'videos:read'),
('streaming_manager', 'videos:update'),
('streaming_manager', 'videos:delete'),
('streaming_manager', 'videos:manage'),
('streaming_manager', 'content:read'),
('streaming_manager', 'analytics:read'),

-- Events Manager permissions
('events_manager', 'events:create'),
('events_manager', 'events:read'),
('events_manager', 'events:update'),
('events_manager', 'events:delete'),
('events_manager', 'events:manage'),
('events_manager', 'users:read'),
('events_manager', 'analytics:read'),

-- Advertisement Manager permissions
('advertisement_manager', 'advertisements:create'),
('advertisement_manager', 'advertisements:read'),
('advertisement_manager', 'advertisements:update'),
('advertisement_manager', 'advertisements:delete'),
('advertisement_manager', 'advertisements:manage'),
('advertisement_manager', 'advertisements:approve'),
('advertisement_manager', 'analytics:read'),
('advertisement_manager', 'financial:read'),

-- Marketing Specialist permissions
('marketing_specialist', 'advertisements:create'),
('marketing_specialist', 'advertisements:read'),
('marketing_specialist', 'advertisements:update'),
('marketing_specialist', 'analytics:read'),

-- User Manager permissions
('user_manager', 'users:create'),
('user_manager', 'users:read'),
('user_manager', 'users:update'),
('user_manager', 'users:delete'),
('user_manager', 'users:manage'),
('user_manager', 'analytics:read'),

-- Support Specialist permissions
('support_specialist', 'users:read'),
('support_specialist', 'users:update'),
('support_specialist', 'technical:support'),

-- Analytics Manager permissions
('analytics_manager', 'analytics:read'),
('analytics_manager', 'analytics:export'),
('analytics_manager', 'analytics:manage'),

-- Financial Administrator permissions
('financial_admin', 'financial:read'),
('financial_admin', 'financial:manage'),
('financial_admin', 'financial:refund'),
('financial_admin', 'analytics:read'),

-- Security Administrator permissions
('security_admin', 'security:read'),
('security_admin', 'security:manage'),
('security_admin', 'security:incident'),

-- Technical Specialist permissions
('technical_specialist', 'technical:read'),
('technical_specialist', 'technical:support'),

-- Academic Reviewer permissions
('academic_reviewer', 'academic:review'),
('academic_reviewer', 'articles:read'),
('academic_reviewer', 'articles:update'),

-- Research Coordinator permissions
('research_coordinator', 'academic:coordinate'),
('research_coordinator', 'articles:read'),
('research_coordinator', 'articles:update'),

-- Advertiser permissions
('advertiser', 'advertisements:create'),
('advertiser', 'advertisements:read'),
('advertiser', 'advertisements:update'),

-- User permissions
('user', 'content:read')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Update existing users to have the 'user' role by default
INSERT INTO user_roles (user_id, role_id)
SELECT id, 'user' FROM users 
WHERE id NOT IN (SELECT user_id FROM user_roles WHERE role_id = 'user')
ON CONFLICT (user_id, role_id) DO NOTHING;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_roles_category ON roles(category);
CREATE INDEX IF NOT EXISTS idx_roles_level ON roles(level);
CREATE INDEX IF NOT EXISTS idx_roles_system_role ON roles(is_system_role);
CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role_id ON user_roles(role_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_permission_id ON role_permissions(permission_id);

-- Add role_id column to users table if it doesn't exist (for backward compatibility)
ALTER TABLE users ADD COLUMN IF NOT EXISTS role_id VARCHAR(100) DEFAULT 'user';
ALTER TABLE users ADD CONSTRAINT fk_users_role_id FOREIGN KEY (role_id) REFERENCES roles(role_id);

-- Update existing users to have proper role_id
UPDATE users SET role_id = 'user' WHERE role_id IS NULL OR role_id = '';

-- Create a view for easy role queries
CREATE OR REPLACE VIEW user_roles_view AS
SELECT 
    u.id as user_id,
    u.email,
    u.first_name,
    u.last_name,
    u.role_id as primary_role,
    r.name as role_name,
    r.level as role_level,
    r.category as role_category,
    r.permissions as role_permissions,
    r.subsystem_access,
    u.created_at,
    u.updated_at
FROM users u
LEFT JOIN roles r ON u.role_id = r.role_id;

-- Log the migration
INSERT INTO migrations (name) VALUES ('006_standardize_roles') ON CONFLICT (name) DO NOTHING; 