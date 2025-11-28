-- Migration 063: Add vid_status column to master_video_list
-- Description: Adds a boolean column to track video status (active/inactive)

-- Add the vid_status column with a default value of true
ALTER TABLE master_video_list 
ADD COLUMN IF NOT EXISTS vid_status BOOLEAN DEFAULT true;

-- Backfill existing records to have vid_status = true
UPDATE master_video_list
SET vid_status = true
WHERE vid_status IS NULL;

-- Add an index for faster filtering by vid_status
CREATE INDEX IF NOT EXISTS idx_master_video_vid_status ON master_video_list(vid_status);

-- Add comment to document the column
COMMENT ON COLUMN master_video_list.vid_status IS 'Indicates if the video is active (true) or inactive (false). Defaults to true.';

