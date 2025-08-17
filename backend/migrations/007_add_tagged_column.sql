-- Migration: Add tagged column to master_video_list
-- Description: Adds tagged boolean column and creates video tags tables for smart tagging system

-- Add tagged column to master_video_list if it doesn't exist
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'master_video_list' AND column_name = 'tagged'
    ) THEN
        ALTER TABLE master_video_list ADD COLUMN tagged BOOLEAN DEFAULT FALSE;
    END IF;
END $$;

-- Create video_tags table if it doesn't exist
CREATE TABLE IF NOT EXISTS video_tags (
    id SERIAL PRIMARY KEY,
    word VARCHAR(100) NOT NULL UNIQUE,
    frequency INTEGER DEFAULT 1,
    category_id INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create tag_categories table if it doesn't exist
CREATE TABLE IF NOT EXISTS tag_categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    color VARCHAR(7) DEFAULT '#6b7280',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for video tags
CREATE INDEX IF NOT EXISTS idx_video_tags_word ON video_tags(word);
CREATE INDEX IF NOT EXISTS idx_video_tags_frequency ON video_tags(frequency DESC);
CREATE INDEX IF NOT EXISTS idx_video_tags_category ON video_tags(category_id);

-- Create indexes for tag categories
CREATE INDEX IF NOT EXISTS idx_tag_categories_name ON tag_categories(name);

-- Create index for tagged column
CREATE INDEX IF NOT EXISTS idx_master_video_tagged ON master_video_list(tagged);

-- Insert default tag categories if they don't exist
INSERT INTO tag_categories (name, description, color) VALUES
    ('Archaeology', 'Archaeological terms and concepts', '#8b5cf6'),
    ('Geography', 'Geographic locations and features', '#06b6d4'),
    ('DNA Research', 'Genetic and DNA-related terms', '#10b981'),
    ('Linguistics', 'Language and linguistic terms', '#f59e0b'),
    ('Historical Evidence', 'Historical documentation and evidence', '#ef4444'),
    ('Cultural Studies', 'Cultural and anthropological terms', '#ec4899'),
    ('Religious Studies', 'Religious and theological terms', '#6366f1'),
    ('Documentary', 'Documentary and media terms', '#84cc16'),
    ('Lecture', 'Educational and lecture terms', '#f97316'),
    ('Interview', 'Interview and discussion terms', '#06b6d4'),
    ('Presentation', 'Presentation and presentation terms', '#8b5cf6'),
    ('Virtual Tour', 'Tour and exploration terms', '#10b981')
ON CONFLICT (name) DO NOTHING;

-- Update existing videos to have tagged = false
UPDATE master_video_list SET tagged = false WHERE tagged IS NULL;
