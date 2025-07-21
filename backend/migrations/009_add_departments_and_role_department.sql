-- Migration 009: Add Departments and Role-Department Relationships
-- This migration creates a departments table and links roles to departments

-- Create departments table
CREATE TABLE IF NOT EXISTS departments (
    id INTEGER PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    icon VARCHAR(10) NOT NULL,
    color VARCHAR(7) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert the departments data
INSERT INTO departments (id, name, icon, color, description) VALUES
(100, 'Secure_Tech', '🛡️', '#3E6313', 'Security and technical infrastructure'),
(200, 'System_Insight', '📈', '#D6BD1A', 'System analytics and insights'),
(300, 'Finance', '💰', '#059669', 'Financial operations and billing'),
(400, 'User_Admin', '👥', '#5FB7E0', 'User management and administration'),
(500, 'Content_Strat', '🎨', '#2563EB', 'Content strategy and planning'),
(600, 'Content_Management', '📚', '#7C3AED', 'Content creation and management'),
(700, 'Marketing', '📢', '#f59e0b', 'Marketing and advertising'),
(800, 'Academia', '🎓', '#7c2d12', 'Academic and research operations'),
(900, 'Base', '🫂', '#6b7280', 'Base user operations'),
(1000, 'Core_Admin', '🔧', '#dc2626', 'Core system administration')
ON CONFLICT (id) DO NOTHING;

-- Add department_id column to roles table
ALTER TABLE roles ADD COLUMN IF NOT EXISTS department_id INTEGER REFERENCES departments(id);

-- Update existing roles with appropriate department assignments
UPDATE roles SET department_id = 1000 WHERE role_id IN ('super_admin', 'system_admin'); -- Core_Admin
UPDATE roles SET department_id = 600 WHERE role_id IN ('content_manager', 'content_editor', 'content_creator'); -- Content_Management
UPDATE roles SET department_id = 500 WHERE role_id IN ('articles_manager', 'youtube_manager', 'streaming_manager'); -- Content_Strat
UPDATE roles SET department_id = 400 WHERE role_id IN ('user_manager', 'support_specialist'); -- User_Admin
UPDATE roles SET department_id = 200 WHERE role_id IN ('analytics_manager'); -- System_Insight
UPDATE roles SET department_id = 300 WHERE role_id IN ('financial_admin'); -- Finance
UPDATE roles SET department_id = 700 WHERE role_id IN ('advertisement_manager', 'marketing_specialist', 'advertiser'); -- Marketing
UPDATE roles SET department_id = 800 WHERE role_id IN ('academic_reviewer', 'research_coordinator'); -- Academia
UPDATE roles SET department_id = 100 WHERE role_id IN ('security_admin', 'technical_specialist'); -- Secure_Tech
UPDATE roles SET department_id = 900 WHERE role_id IN ('user'); -- Base

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_roles_department_id ON roles(department_id);
CREATE INDEX IF NOT EXISTS idx_departments_name ON departments(name);

-- Create a view for roles with department information
CREATE OR REPLACE VIEW roles_with_departments AS
SELECT 
    r.id,
    r.role_id,
    r.name as role_name,
    r.slug,
    r.description as role_description,
    r.category,
    r.level,
    r.permissions,
    r.is_system_role,
    r.color as role_color,
    r.icon as role_icon,
    r.subsystem_access,
    r.created_at,
    r.updated_at,
    d.id as department_id,
    d.name as department_name,
    d.icon as department_icon,
    d.color as department_color,
    d.description as department_description
FROM roles r
LEFT JOIN departments d ON r.department_id = d.id;

-- Log the migration
INSERT INTO migrations (name) VALUES ('009_add_departments_and_role_department') ON CONFLICT (name) DO NOTHING; 