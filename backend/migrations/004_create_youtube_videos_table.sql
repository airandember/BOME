-- Migration: Create youtube_videos table
-- Description: Table to store YouTube videos from PubSubHubbub webhooks

CREATE TABLE youtube_videos (
    id VARCHAR(255) PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    published_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    thumbnail_url TEXT,
    video_url TEXT NOT NULL,
    embed_url TEXT NOT NULL,
    duration VARCHAR(50),
    view_count BIGINT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_published_at (published_at),
    INDEX idx_created_at (created_at)
);

-- Insert sample YouTube videos for testing
INSERT INTO youtube_videos (
    id, title, description, published_at, updated_at, 
    thumbnail_url, video_url, embed_url, duration, view_count
) VALUES 
(
    'sample_video_1',
    'Book of Mormon Archaeological Evidence - Introduction',
    'An introduction to archaeological evidence supporting the Book of Mormon narrative.',
    '2024-01-15 10:00:00',
    '2024-01-15 10:00:00',
    'https://img.youtube.com/vi/sample_video_1/maxresdefault.jpg',
    'https://www.youtube.com/watch?v=sample_video_1',
    'https://www.youtube.com/embed/sample_video_1',
    '15:30',
    12500
),
(
    'sample_video_2',
    'Ancient Civilizations and Book of Mormon Geography',
    'Exploring the geographical connections between ancient civilizations and Book of Mormon locations.',
    '2024-01-10 14:30:00',
    '2024-01-10 14:30:00',
    'https://img.youtube.com/vi/sample_video_2/maxresdefault.jpg',
    'https://www.youtube.com/watch?v=sample_video_2',
    'https://www.youtube.com/embed/sample_video_2',
    '22:45',
    8750
),
(
    'sample_video_3',
    'DNA Evidence and Book of Mormon Peoples',
    'Examining DNA research and its implications for Book of Mormon populations.',
    '2024-01-05 09:15:00',
    '2024-01-05 09:15:00',
    'https://img.youtube.com/vi/sample_video_3/maxresdefault.jpg',
    'https://www.youtube.com/watch?v=sample_video_3',
    'https://www.youtube.com/embed/sample_video_3',
    '18:20',
    15200
); 