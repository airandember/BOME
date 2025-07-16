package database

// Remove all duplicate video functions from admin.go
// Keep only admin-specific functions, remove:
// - CreateVideo
// - GetVideoByID
// - GetVideoByBunnyID
// - GetVideos
// - UpdateVideoStatus
// - UpdateVideoViews
// - IncrementViewCount
// - GetVideoCategories
// - SearchVideos
// - UpdateVideo
// - DeleteVideo
// - ScheduleVideo
// - GetScheduledVideos
// - UnscheduleVideo

// These functions already exist in video.go, so remove them from admin.go
