-- Migration: Fix video_tags table constraints
-- Date: 2025-08-17

-- Drop existing table if it exists (this will remove any data)
DROP TABLE IF EXISTS video_tags CASCADE;

-- Recreate video_tags table with proper constraints
CREATE TABLE video_tags (
    id SERIAL PRIMARY KEY,
    word VARCHAR(100) NOT NULL UNIQUE,
    frequency INTEGER DEFAULT 1,
    category_id INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_video_tags_word ON video_tags(word);
CREATE INDEX IF NOT EXISTS idx_video_tags_frequency ON video_tags(frequency DESC);
CREATE INDEX IF NOT EXISTS idx_video_tags_category ON video_tags(category_id);

-- Add foreign key constraint if tag_categories table exists
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'tag_categories') THEN
        ALTER TABLE video_tags ADD CONSTRAINT fk_video_tags_category 
        FOREIGN KEY (category_id) REFERENCES tag_categories(id) ON DELETE SET NULL;
    END IF;
END $$;

-- Insert some sample tags for testing (optional)
INSERT INTO video_tags (word, frequency) VALUES 
    ('test', 1),
    ('sample', 1)
ON CONFLICT (word) DO NOTHING;
