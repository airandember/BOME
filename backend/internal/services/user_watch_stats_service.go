package services

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"bome-backend/internal/database"
)

// UserWatchStatsService handles user watching statistics and achievements
type UserWatchStatsService struct {
	db *database.DB
}

// NewUserWatchStatsService creates a new user watch stats service
func NewUserWatchStatsService(db *database.DB) *UserWatchStatsService {
	return &UserWatchStatsService{db: db}
}

// UserWatchStats represents comprehensive user watching statistics
type UserWatchStats struct {
	UserID                int             `json:"user_id"`
	TotalWatchTimeMinutes int             `json:"total_watch_time_minutes"`
	TotalWatchTimeHours   float64         `json:"total_watch_time_hours"`
	TotalVideosWatched    int             `json:"total_videos_watched"`
	VideosCompleted       int             `json:"videos_completed"`
	CompletionRate        float64         `json:"completion_rate"`
	CurrentStreak         int             `json:"current_streak"`
	LongestStreak         int             `json:"longest_streak"`
	TotalDaysActive       int             `json:"total_days_active"`
	AverageSessionMinutes float64         `json:"average_session_minutes"`
	FavoriteCategories    []CategoryStats `json:"favorite_categories"`
	RecentActivity        []DailyActivity `json:"recent_activity"`
	Achievements          []Achievement   `json:"achievements"`
	MemberSince           time.Time       `json:"member_since"`
	LastWatchedAt         *time.Time      `json:"last_watched_at,omitempty"`
}

// CategoryStats represents watching stats for a category
type CategoryStats struct {
	CategoryID       int     `json:"category_id"`
	CategoryName     string  `json:"category_name"`
	VideosWatched    int     `json:"videos_watched"`
	WatchTimeMinutes int     `json:"watch_time_minutes"`
	Percentage       float64 `json:"percentage"`
}

// DailyActivity represents activity for a specific day
type DailyActivity struct {
	Date             string `json:"date"`
	VideosWatched    int    `json:"videos_watched"`
	WatchTimeMinutes int    `json:"watch_time_minutes"`
	VideosCompleted  int    `json:"videos_completed"`
}

// Achievement represents a user achievement/badge
type Achievement struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	UnlockedAt  time.Time `json:"unlocked_at"`
	Progress    float64   `json:"progress"`
	IsUnlocked  bool      `json:"is_unlocked"`
}

// WatchSession represents a watching session
type WatchSession struct {
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	DurationMinutes int       `json:"duration_minutes"`
	VideosWatched   int       `json:"videos_watched"`
	VideosCompleted int       `json:"videos_completed"`
}

// TopVideo represents a user's most watched video
type TopVideo struct {
	VideoID           int       `json:"video_id"`
	Title             string    `json:"title"`
	ThumbnailURL      string    `json:"thumbnail_url"`
	WatchCount        int       `json:"watch_count"`
	TotalWatchMinutes int       `json:"total_watch_minutes"`
	CompletionRate    float64   `json:"completion_rate"`
	LastWatchedAt     time.Time `json:"last_watched_at"`
}

// ===========================
// USER STATISTICS
// ===========================

// GetUserWatchStats retrieves comprehensive watch statistics for a user
func (s *UserWatchStatsService) GetUserWatchStats(userID int) (*UserWatchStats, error) {
	stats := &UserWatchStats{
		UserID: userID,
	}

	// Get basic watch stats
	err := s.db.DB.QueryRow(`
		SELECT 
			COALESCE(SUM(wh.total_watch_time), 0) as total_seconds,
			COUNT(DISTINCT wh.video_id) as videos_watched,
			COUNT(DISTINCT CASE WHEN wh.completed = true THEN wh.video_id END) as videos_completed,
			COUNT(DISTINCT DATE(wh.last_watched_at)) as days_active,
			MAX(wh.last_watched_at) as last_watched
		FROM watch_history wh
		WHERE wh.user_id = $1
	`, userID).Scan(
		&stats.TotalWatchTimeMinutes,
		&stats.TotalVideosWatched,
		&stats.VideosCompleted,
		&stats.TotalDaysActive,
		&stats.LastWatchedAt,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get basic stats: %v", err)
	}

	// Convert seconds to minutes and hours
	stats.TotalWatchTimeHours = float64(stats.TotalWatchTimeMinutes) / 3600.0
	stats.TotalWatchTimeMinutes = stats.TotalWatchTimeMinutes / 60

	// Calculate completion rate
	if stats.TotalVideosWatched > 0 {
		stats.CompletionRate = float64(stats.VideosCompleted) / float64(stats.TotalVideosWatched) * 100
	}

	// Calculate average session length
	var totalSessions int
	err = s.db.DB.QueryRow(`
		SELECT COUNT(DISTINCT DATE(last_watched_at))
		FROM watch_history
		WHERE user_id = $1
	`, userID).Scan(&totalSessions)
	if err == nil && totalSessions > 0 {
		stats.AverageSessionMinutes = float64(stats.TotalWatchTimeMinutes) / float64(totalSessions)
	}

	// Get member since date
	err = s.db.DB.QueryRow(`
		SELECT created_at FROM users WHERE id = $1
	`, userID).Scan(&stats.MemberSince)
	if err != nil {
		stats.MemberSince = time.Now() // Fallback
	}

	// Get streaks
	stats.CurrentStreak, stats.LongestStreak = s.calculateStreaks(userID)

	// Get favorite categories
	stats.FavoriteCategories, _ = s.getFavoriteCategories(userID)

	// Get recent activity (last 30 days)
	stats.RecentActivity, _ = s.getRecentActivity(userID, 30)

	// Get achievements
	stats.Achievements = s.calculateAchievements(stats)

	return stats, nil
}

// calculateStreaks calculates current and longest viewing streaks
func (s *UserWatchStatsService) calculateStreaks(userID int) (int, int) {
	// Get all unique days user watched videos
	rows, err := s.db.DB.Query(`
		SELECT DISTINCT DATE(last_watched_at) as watch_date
		FROM watch_history
		WHERE user_id = $1
		ORDER BY watch_date DESC
	`, userID)
	if err != nil {
		return 0, 0
	}
	defer rows.Close()

	var dates []time.Time
	for rows.Next() {
		var date time.Time
		if err := rows.Scan(&date); err == nil {
			dates = append(dates, date)
		}
	}

	if len(dates) == 0 {
		return 0, 0
	}

	currentStreak := 0
	longestStreak := 0
	streakCount := 1

	today := time.Now().Truncate(24 * time.Hour)
	yesterday := today.AddDate(0, 0, -1)

	// Check if current streak is active (watched today or yesterday)
	if dates[0].Equal(today) || dates[0].Equal(yesterday) {
		currentStreak = 1

		// Count consecutive days
		for i := 1; i < len(dates); i++ {
			expectedDate := dates[i-1].AddDate(0, 0, -1)
			if dates[i].Equal(expectedDate) {
				currentStreak++
			} else {
				break
			}
		}
	}

	// Calculate longest streak
	streakCount = 1
	for i := 1; i < len(dates); i++ {
		expectedDate := dates[i-1].AddDate(0, 0, -1)
		if dates[i].Equal(expectedDate) {
			streakCount++
			if streakCount > longestStreak {
				longestStreak = streakCount
			}
		} else {
			streakCount = 1
		}
	}

	if longestStreak < currentStreak {
		longestStreak = currentStreak
	}

	return currentStreak, longestStreak
}

// getFavoriteCategories gets user's top categories by watch time
func (s *UserWatchStatsService) getFavoriteCategories(userID int) ([]CategoryStats, error) {
	rows, err := s.db.DB.Query(`
		SELECT 
			vc.category_id,
			c.name as category_name,
			COUNT(DISTINCT wh.video_id) as videos_watched,
			SUM(wh.total_watch_time) / 60 as watch_minutes
		FROM watch_history wh
		JOIN video_categories vc ON vc.video_id = wh.video_id
		JOIN categories c ON c.id = vc.category_id
		WHERE wh.user_id = $1
		GROUP BY vc.category_id, c.name
		ORDER BY watch_minutes DESC
		LIMIT 5
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []CategoryStats
	var totalMinutes int

	// First pass: collect data
	for rows.Next() {
		cat := CategoryStats{}
		err := rows.Scan(&cat.CategoryID, &cat.CategoryName, &cat.VideosWatched, &cat.WatchTimeMinutes)
		if err != nil {
			continue
		}
		totalMinutes += cat.WatchTimeMinutes
		categories = append(categories, cat)
	}

	// Second pass: calculate percentages
	for i := range categories {
		if totalMinutes > 0 {
			categories[i].Percentage = float64(categories[i].WatchTimeMinutes) / float64(totalMinutes) * 100
		}
	}

	return categories, nil
}

// getRecentActivity gets daily activity for the last N days
func (s *UserWatchStatsService) getRecentActivity(userID int, days int) ([]DailyActivity, error) {
	rows, err := s.db.DB.Query(`
		SELECT 
			DATE(last_watched_at) as activity_date,
			COUNT(DISTINCT video_id) as videos_watched,
			SUM(total_watch_time) / 60 as watch_minutes,
			COUNT(CASE WHEN completed = true THEN 1 END) as videos_completed
		FROM watch_history
		WHERE user_id = $1 
		  AND last_watched_at >= CURRENT_DATE - INTERVAL '1 day' * $2
		GROUP BY DATE(last_watched_at)
		ORDER BY activity_date DESC
	`, userID, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []DailyActivity
	for rows.Next() {
		var activity DailyActivity
		var date time.Time
		err := rows.Scan(&date, &activity.VideosWatched, &activity.WatchTimeMinutes, &activity.VideosCompleted)
		if err != nil {
			continue
		}
		activity.Date = date.Format("2006-01-02")
		activities = append(activities, activity)
	}

	return activities, nil
}

// calculateAchievements determines which achievements a user has unlocked
func (s *UserWatchStatsService) calculateAchievements(stats *UserWatchStats) []Achievement {
	achievements := []Achievement{
		// Watch Time Achievements
		{
			ID:          "first_video",
			Name:        "First Steps",
			Description: "Watch your first video",
			Icon:        "🎬",
			Progress:    minFloat64(float64(stats.TotalVideosWatched)/1.0*100, 100.0),
			IsUnlocked:  stats.TotalVideosWatched >= 1,
		},
		{
			ID:          "10_videos",
			Name:        "Getting Started",
			Description: "Watch 10 videos",
			Icon:        "📺",
			Progress:    minFloat64(float64(stats.TotalVideosWatched)/10.0*100, 100.0),
			IsUnlocked:  stats.TotalVideosWatched >= 10,
		},
		{
			ID:          "50_videos",
			Name:        "Video Enthusiast",
			Description: "Watch 50 videos",
			Icon:        "🎥",
			Progress:    minFloat64(float64(stats.TotalVideosWatched)/50.0*100, 100.0),
			IsUnlocked:  stats.TotalVideosWatched >= 50,
		},
		{
			ID:          "100_videos",
			Name:        "Century Club",
			Description: "Watch 100 videos",
			Icon:        "💯",
			Progress:    minFloat64(float64(stats.TotalVideosWatched)/100.0*100, 100.0),
			IsUnlocked:  stats.TotalVideosWatched >= 100,
		},

		// Watch Hours Achievements
		{
			ID:          "1_hour",
			Name:        "The First Hour",
			Description: "Watch 1 hour of content",
			Icon:        "⏱️",
			Progress:    minFloat64(stats.TotalWatchTimeHours/1.0*100, 100.0),
			IsUnlocked:  stats.TotalWatchTimeHours >= 1,
		},
		{
			ID:          "10_hours",
			Name:        "Dedicated Viewer",
			Description: "Watch 10 hours of content",
			Icon:        "⏰",
			Progress:    minFloat64(stats.TotalWatchTimeHours/10.0*100, 100.0),
			IsUnlocked:  stats.TotalWatchTimeHours >= 10,
		},
		{
			ID:          "50_hours",
			Name:        "Binge Watcher",
			Description: "Watch 50 hours of content",
			Icon:        "📻",
			Progress:    minFloat64(stats.TotalWatchTimeHours/50.0*100, 100.0),
			IsUnlocked:  stats.TotalWatchTimeHours >= 50,
		},
		{
			ID:          "100_hours",
			Name:        "Content Connoisseur",
			Description: "Watch 100 hours of content",
			Icon:        "🎭",
			Progress:    minFloat64(stats.TotalWatchTimeHours/100.0*100, 100.0),
			IsUnlocked:  stats.TotalWatchTimeHours >= 100,
		},

		// Streak Achievements
		{
			ID:          "3_day_streak",
			Name:        "On a Roll",
			Description: "Watch videos 3 days in a row",
			Icon:        "🔥",
			Progress:    minFloat64(float64(stats.CurrentStreak)/3.0*100, 100.0),
			IsUnlocked:  stats.LongestStreak >= 3,
		},
		{
			ID:          "7_day_streak",
			Name:        "Week Warrior",
			Description: "Watch videos 7 days in a row",
			Icon:        "⚡",
			Progress:    minFloat64(float64(stats.CurrentStreak)/7.0*100, 100.0),
			IsUnlocked:  stats.LongestStreak >= 7,
		},
		{
			ID:          "30_day_streak",
			Name:        "Monthly Master",
			Description: "Watch videos 30 days in a row",
			Icon:        "🏆",
			Progress:    minFloat64(float64(stats.CurrentStreak)/30.0*100, 100.0),
			IsUnlocked:  stats.LongestStreak >= 30,
		},

		// Completion Achievements
		{
			ID:          "first_completion",
			Name:        "Finisher",
			Description: "Complete your first video",
			Icon:        "✅",
			Progress:    minFloat64(float64(stats.VideosCompleted)/1.0*100, 100.0),
			IsUnlocked:  stats.VideosCompleted >= 1,
		},
		{
			ID:          "10_completions",
			Name:        "Completionist",
			Description: "Complete 10 videos",
			Icon:        "✔️",
			Progress:    minFloat64(float64(stats.VideosCompleted)/10.0*100, 100.0),
			IsUnlocked:  stats.VideosCompleted >= 10,
		},
		{
			ID:          "high_completion",
			Name:        "Committed Viewer",
			Description: "Maintain 80% completion rate (min 10 videos)",
			Icon:        "🎯",
			Progress:    minFloat64(stats.CompletionRate/80.0*100, 100.0),
			IsUnlocked:  stats.CompletionRate >= 80 && stats.TotalVideosWatched >= 10,
		},
	}

	// Set unlocked timestamps (for now, use current time)
	now := time.Now()
	for i := range achievements {
		if achievements[i].IsUnlocked {
			achievements[i].UnlockedAt = now
		}
	}

	return achievements
}

// GetTopVideos gets user's most watched videos
func (s *UserWatchStatsService) GetTopVideos(userID int, limit int) ([]TopVideo, error) {
	if limit <= 0 {
		limit = 10
	}

	rows, err := s.db.DB.Query(`
		SELECT 
			v.id,
		v.title,
		v.thumbnail_url,
		wh.view_count as watch_count,
		wh.total_watch_time / 60 as total_minutes,
		CASE WHEN wh.completed THEN 100.0 ELSE 0.0 END as completion_rate,
		wh.last_watched_at as last_watched
	FROM watch_history wh
	JOIN videos v ON v.id = wh.video_id
	WHERE wh.user_id = $1
	ORDER BY watch_count DESC, total_minutes DESC
	LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []TopVideo
	for rows.Next() {
		video := TopVideo{}
		err := rows.Scan(
			&video.VideoID, &video.Title, &video.ThumbnailURL,
			&video.WatchCount, &video.TotalWatchMinutes, &video.CompletionRate,
			&video.LastWatchedAt,
		)
		if err != nil {
			log.Printf("Error scanning top video: %v", err)
			continue
		}
		videos = append(videos, video)
	}

	return videos, nil
}

// GetWatchingSessions gets user's recent watching sessions
func (s *UserWatchStatsService) GetWatchingSessions(userID int, limit int) ([]WatchSession, error) {
	if limit <= 0 {
		limit = 10
	}

	// Group views into sessions (views within 30 minutes of each other)
	rows, err := s.db.DB.Query(`
		WITH session_groups AS (
			SELECT 
				created_at,
				watch_duration,
			completed,
			CASE 
				WHEN last_watched_at - LAG(last_watched_at) OVER (ORDER BY last_watched_at) > INTERVAL '30 minutes'
				THEN 1 
				ELSE 0 
			END as new_session
		FROM watch_history
		WHERE user_id = $1
		ORDER BY last_watched_at
		),
		sessions AS (
			SELECT 
				*,
				SUM(new_session) OVER (ORDER BY last_watched_at) as session_id
			FROM session_groups
		)
		SELECT 
			MIN(created_at) as start_time,
			MAX(created_at) as end_time,
			COUNT(*) as videos_watched,
			COUNT(CASE WHEN completed THEN 1 END) as videos_completed,
			SUM(watch_duration) / 60 as duration_minutes
		FROM sessions
		GROUP BY session_id
		ORDER BY start_time DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []WatchSession
	for rows.Next() {
		session := WatchSession{}
		err := rows.Scan(
			&session.StartTime, &session.EndTime,
			&session.VideosWatched, &session.VideosCompleted,
			&session.DurationMinutes,
		)
		if err != nil {
			log.Printf("Error scanning session: %v", err)
			continue
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// Helper function for minFloat64
func minFloat64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
