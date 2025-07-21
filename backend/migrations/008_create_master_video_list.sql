-- Migration: Create master video list table
-- Description: Master list for video metadata with Bunny.net synchronization

CREATE TABLE master_video_list (
    id SERIAL PRIMARY KEY,
    bunny_video_id VARCHAR(255) UNIQUE NOT NULL,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    category VARCHAR(100),
    tags JSONB DEFAULT '[]',
    duration INTEGER DEFAULT 0,
    file_size BIGINT DEFAULT 0,
    resolution VARCHAR(50),
    framerate DECIMAL(5,2),
    thumbnail_url TEXT,
    video_url TEXT,
    iframe_src TEXT,
    playback_url TEXT,
    status VARCHAR(50) DEFAULT 'processing',
    views INTEGER DEFAULT 0,
    likes INTEGER DEFAULT 0,
    is_public BOOLEAN DEFAULT true,
    encode_progress INTEGER DEFAULT 0,
    available_resolutions JSONB DEFAULT '[]',
    collection_id VARCHAR(255),
    average_watch_time INTEGER DEFAULT 0,
    total_watch_time BIGINT DEFAULT 0,
    
    -- Sync tracking fields
    last_bunny_sync TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_master_update TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    sync_status VARCHAR(50) DEFAULT 'synced', -- synced, needs_attention, conflict
    sync_notes TEXT,
    
    -- Metadata tracking
    metadata_version INTEGER DEFAULT 1,
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX idx_master_video_bunny_id ON master_video_list(bunny_video_id);
CREATE INDEX idx_master_video_status ON master_video_list(status);
CREATE INDEX idx_master_video_category ON master_video_list(category);
CREATE INDEX idx_master_video_sync_status ON master_video_list(sync_status);
CREATE INDEX idx_master_video_created_at ON master_video_list(created_at);
CREATE INDEX idx_master_video_views ON master_video_list(views DESC);
CREATE INDEX idx_master_video_collection ON master_video_list(collection_id);

-- Create sync conflicts table for tracking discrepancies
CREATE TABLE video_sync_conflicts (
    id SERIAL PRIMARY KEY,
    master_video_id INTEGER REFERENCES master_video_list(id) ON DELETE CASCADE,
    bunny_video_id VARCHAR(255) NOT NULL,
    conflict_type VARCHAR(50) NOT NULL, -- field_mismatch, missing_field, status_mismatch
    field_name VARCHAR(100),
    master_value TEXT,
    bunny_value TEXT,
    proposed_action VARCHAR(50) NOT NULL, -- update_master, update_bunny, update_both, manual_review
    admin_notes TEXT,
    resolved BOOLEAN DEFAULT false,
    resolved_by INTEGER REFERENCES users(id),
    resolved_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for sync conflicts
CREATE INDEX idx_sync_conflicts_master_id ON video_sync_conflicts(master_video_id);
CREATE INDEX idx_sync_conflicts_bunny_id ON video_sync_conflicts(bunny_video_id);
CREATE INDEX idx_sync_conflicts_resolved ON video_sync_conflicts(resolved);
CREATE INDEX idx_sync_conflicts_type ON video_sync_conflicts(conflict_type);

-- Create sync audit log table
CREATE TABLE video_sync_audit_log (
    id SERIAL PRIMARY KEY,
    master_video_id INTEGER REFERENCES master_video_list(id) ON DELETE CASCADE,
    bunny_video_id VARCHAR(255) NOT NULL,
    sync_action VARCHAR(50) NOT NULL, -- sync_from_bunny, sync_to_bunny, conflict_resolved, manual_update
    sync_result VARCHAR(50) NOT NULL, -- success, failed, partial, conflict
    changes_made JSONB DEFAULT '{}',
    error_message TEXT,
    performed_by INTEGER REFERENCES users(id),
    performed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for audit log
CREATE INDEX idx_sync_audit_master_id ON video_sync_audit_log(master_video_id);
CREATE INDEX idx_sync_audit_bunny_id ON video_sync_audit_log(bunny_video_id);
CREATE INDEX idx_sync_audit_action ON video_sync_audit_log(sync_action);
CREATE INDEX idx_sync_audit_performed_at ON video_sync_audit_log(performed_at);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_master_video_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for updated_at
CREATE TRIGGER trigger_master_video_updated_at
    BEFORE UPDATE ON master_video_list
    FOR EACH ROW
    EXECUTE FUNCTION update_master_video_updated_at();

-- Create function to log sync conflicts
CREATE OR REPLACE FUNCTION log_sync_conflict(
    p_master_video_id INTEGER,
    p_bunny_video_id VARCHAR(255),
    p_conflict_type VARCHAR(50),
    p_field_name VARCHAR(100),
    p_master_value TEXT,
    p_bunny_value TEXT,
    p_proposed_action VARCHAR(50)
) RETURNS INTEGER AS $$
DECLARE
    conflict_id INTEGER;
BEGIN
    INSERT INTO video_sync_conflicts (
        master_video_id, bunny_video_id, conflict_type, field_name,
        master_value, bunny_value, proposed_action
    ) VALUES (
        p_master_video_id, p_bunny_video_id, p_conflict_type, p_field_name,
        p_master_value, p_bunny_value, p_proposed_action
    ) RETURNING id INTO conflict_id;
    
    RETURN conflict_id;
END;
$$ LANGUAGE plpgsql;

-- Create function to log sync audit
CREATE OR REPLACE FUNCTION log_sync_audit(
    p_master_video_id INTEGER,
    p_bunny_video_id VARCHAR(255),
    p_sync_action VARCHAR(50),
    p_sync_result VARCHAR(50),
    p_changes_made JSONB,
    p_error_message TEXT,
    p_performed_by INTEGER
) RETURNS INTEGER AS $$
DECLARE
    audit_id INTEGER;
BEGIN
    INSERT INTO video_sync_audit_log (
        master_video_id, bunny_video_id, sync_action, sync_result,
        changes_made, error_message, performed_by
    ) VALUES (
        p_master_video_id, p_bunny_video_id, p_sync_action, p_sync_result,
        p_changes_made, p_error_message, p_performed_by
    ) RETURNING id INTO audit_id;
    
    RETURN audit_id;
END;
$$ LANGUAGE plpgsql;

-- Insert sample data for testing
INSERT INTO master_video_list (
    bunny_video_id, title, description, category, tags, duration, file_size,
    resolution, thumbnail_url, status, views, is_public, collection_id
) VALUES 
(
    'sample-master-1',
    'Book of Mormon Archaeological Evidence - Introduction',
    'An introduction to archaeological evidence supporting the Book of Mormon narrative.',
    'Archaeology',
    '["archaeology", "evidence", "introduction", "book-of-mormon"]',
    930,
    145600000,
    '1920x1080',
    'https://example.com/thumb1.jpg',
    'ready',
    12500,
    true,
    'archaeology-collection'
),
(
    'sample-master-2',
    'Ancient Civilizations and Book of Mormon Geography',
    'Exploring the geographical connections between ancient civilizations and Book of Mormon locations.',
    'Geography',
    '["geography", "ancient-civilizations", "locations", "book-of-mormon"]',
    1365,
    298400000,
    '1920x1080',
    'https://example.com/thumb2.jpg',
    'ready',
    8750,
    true,
    'geography-collection'
),
(
    'sample-master-3',
    'DNA Evidence and Book of Mormon Peoples',
    'Examining DNA research and its implications for Book of Mormon populations.',
    'Science',
    '["dna", "science", "genetics", "populations", "book-of-mormon"]',
    1100,
    245800000,
    '1920x1080',
    'https://example.com/thumb3.jpg',
    'ready',
    15200,
    true,
    'science-collection'
);

-- Create view for easy access to sync status
CREATE VIEW video_sync_status AS
SELECT 
    mvl.id,
    mvl.bunny_video_id,
    mvl.title,
    mvl.status as master_status,
    mvl.sync_status,
    mvl.last_bunny_sync,
    mvl.last_master_update,
    COUNT(vsc.id) as pending_conflicts,
    COUNT(vsc.id) FILTER (WHERE vsc.resolved = false) as unresolved_conflicts
FROM master_video_list mvl
LEFT JOIN video_sync_conflicts vsc ON mvl.id = vsc.master_video_id
GROUP BY mvl.id, mvl.bunny_video_id, mvl.title, mvl.status, mvl.sync_status, 
         mvl.last_bunny_sync, mvl.last_master_update;

-- Create view for admin dashboard
CREATE VIEW admin_video_dashboard AS
SELECT 
    mvl.id,
    mvl.bunny_video_id,
    mvl.title,
    mvl.category,
    mvl.status,
    mvl.views,
    mvl.duration,
    mvl.file_size,
    mvl.sync_status,
    mvl.last_bunny_sync,
    mvl.last_master_update,
    COUNT(vsc.id) FILTER (WHERE vsc.resolved = false) as needs_attention,
    CASE 
        WHEN mvl.sync_status = 'needs_attention' THEN '⚠️ Needs Review'
        WHEN mvl.sync_status = 'conflict' THEN '🚨 Conflict'
        WHEN mvl.sync_status = 'synced' THEN '✅ Synced'
        ELSE '❓ Unknown'
    END as sync_status_display
FROM master_video_list mvl
LEFT JOIN video_sync_conflicts vsc ON mvl.id = vsc.master_video_id
GROUP BY mvl.id, mvl.bunny_video_id, mvl.title, mvl.category, mvl.status, 
         mvl.views, mvl.duration, mvl.file_size, mvl.sync_status, 
         mvl.last_bunny_sync, mvl.last_master_update; 