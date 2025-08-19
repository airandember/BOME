-- Migration: Create article_exclusions table
-- This table stores words that should be excluded from tag generation

CREATE TABLE IF NOT EXISTS article_exclusions (
    id SERIAL PRIMARY KEY,
    word VARCHAR(100) NOT NULL UNIQUE,
    excluded BOOLEAN NOT NULL DEFAULT true,
    subsite_id INTEGER REFERENCES subsites(id),
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_article_exclusions_word ON article_exclusions(word);
CREATE INDEX IF NOT EXISTS idx_article_exclusions_subsite ON article_exclusions(subsite_id);

-- Insert default common words that should be excluded
INSERT INTO article_exclusions (word, excluded, subsite_id) VALUES
    ('and', true, 1), ('the', true, 1), ('of', true, 1), ('for', true, 1), ('a', true, 1), ('an', true, 1),
    ('in', true, 1), ('on', true, 1), ('at', true, 1), ('to', true, 1), ('with', true, 1), ('by', true, 1),
    ('from', true, 1), ('into', true, 1), ('during', true, 1), ('including', true, 1), ('until', true, 1),
    ('against', true, 1), ('among', true, 1), ('throughout', true, 1), ('despite', true, 1),
    ('towards', true, 1), ('upon', true, 1), ('concerning', true, 1), ('excepting', true, 1),
    ('excluding', true, 1), ('following', true, 1), ('inside', true, 1), ('outside', true, 1),
    ('over', true, 1), ('past', true, 1), ('since', true, 1), ('under', true, 1), ('within', true, 1),
    ('without', true, 1), ('about', true, 1), ('above', true, 1), ('across', true, 1), ('after', true, 1),
    ('along', true, 1), ('around', true, 1), ('before', true, 1), ('behind', true, 1), ('below', true, 1),
    ('beneath', true, 1), ('beside', true, 1), ('between', true, 1), ('beyond', true, 1),
    ('down', true, 1), ('except', true, 1), ('near', true, 1), ('off', true, 1), ('onto', true, 1),
    ('out', true, 1), ('through', true, 1), ('toward', true, 1), ('underneath', true, 1),
    ('up', true, 1), ('1080p', true, 1), ('720p', true, 1), ('480p', true, 1), ('360p', true, 1), ('240p', true, 1), ('144p', true, 1),
    ('are', true, 1), ('how', true, 1), ('is', true, 1), ('29fps', true, 1), ('your', true, 1), ('why', true, 1), ('what', true, 1),
    ('when', true, 1), ('where', true, 1), ('who', true, 1), ('which', true, 1), ('that', true, 1), ('this', true, 1), ('these', true, 1),
    ('they', true, 1), ('them', true, 1), ('their', true, 1), ('they''re', true, 1), ('they''ve', true, 1), ('they''ll', true, 1)
ON CONFLICT (word) DO NOTHING;

-- Add trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_article_exclusions_updated_at 
    BEFORE UPDATE ON article_exclusions 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();
