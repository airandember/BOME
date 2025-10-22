-- =====================================================
-- OPTIONAL: Reset master_video_list ID sequence
-- =====================================================
-- ⚠️ WARNING: Only run this if you understand the implications!
-- ⚠️ This should ONLY be done if:
--    1. Your master_video_list table is EMPTY or nearly empty
--    2. You want clean sequential IDs starting from 1
--    3. You have NO foreign key references to old IDs
-- =====================================================

-- Step 1: Check how many videos you currently have
SELECT 
    COUNT(*) as total_videos,
    MIN(id) as min_id,
    MAX(id) as max_id
FROM master_video_list;

-- Step 2: If you're sure you want to reset (and table is empty/small):
-- Uncomment the following lines:

/*
-- Get the max ID (or use 1 if table is empty)
SELECT SETVAL('master_video_list_id_seq', 
    COALESCE((SELECT MAX(id) FROM master_video_list), 0) + 1, 
    false
);

-- Or force reset to 1 (ONLY if table is empty!)
-- SELECT SETVAL('master_video_list_id_seq', 1, false);
*/

-- Step 3: Verify the sequence
SELECT 
    last_value as current_sequence_value,
    is_called
FROM master_video_list_id_seq;

-- =====================================================
-- BETTER ALTERNATIVE: Just use your existing IDs!
-- =====================================================
-- Instead of resetting, just query for real IDs:

SELECT 
    id, 
    title, 
    bunny_video_id,
    status as vid_status,
    views
FROM master_video_list 
ORDER BY id ASC 
LIMIT 20;

-- Use those IDs in your test data and examples!
-- Example: If first ID is 11041, use that in video_presenters

