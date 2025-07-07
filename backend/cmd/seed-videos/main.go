package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	// Database connection
	db, err := sql.Open("postgres", "postgres://bome_user:your_secure_password_here@localhost:5432/bome_streaming?sslmode=disable")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test the connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	fmt.Println("Connected to database successfully")

	// Create test videos
	testVideos := []struct {
		title        string
		description  string
		bunnyVideoID string
		thumbnailURL string
		category     string
		duration     int
		fileSize     int64
		tags         []string
		createdBy    int
	}{
		{
			title:        "Archaeological Evidence for the Book of Mormon",
			description:  "Exploring recent archaeological discoveries that support Book of Mormon narratives, including ancient civilizations, metallurgy, and cultural practices found in the Americas.",
			bunnyVideoID: "test-video-1",
			thumbnailURL: "https://example.com/thumb1.jpg",
			category:     "Archaeology",
			duration:     942, // 15:42
			fileSize:     145600000,
			tags:         []string{"archaeology", "evidence", "ancient-america", "civilizations"},
			createdBy:    1,
		},
		{
			title:        "DNA and the Book of Mormon: Scientific Perspectives",
			description:  "A comprehensive look at DNA evidence and its relationship to Book of Mormon populations, examining recent genetic studies and their implications.",
			bunnyVideoID: "test-video-2",
			thumbnailURL: "https://example.com/thumb2.jpg",
			category:     "Science",
			duration:     1335, // 22:15
			fileSize:     298400000,
			tags:         []string{"dna", "science", "genetics", "populations"},
			createdBy:    1,
		},
		{
			title:        "Mesoamerican Connections to Book of Mormon Geography",
			description:  "Examining cultural and geographical connections between Mesoamerica and the Book of Mormon, including recent discoveries and scholarly research.",
			bunnyVideoID: "test-video-3",
			thumbnailURL: "https://example.com/thumb3.jpg",
			category:     "Geography",
			duration:     1113, // 18:33
			fileSize:     245800000,
			tags:         []string{"mesoamerica", "geography", "culture", "maya"},
			createdBy:    1,
		},
		{
			title:        "Linguistic Analysis of Book of Mormon Names",
			description:  "Scholarly analysis of Hebrew and Egyptian linguistic patterns in Book of Mormon names and their ancient Near Eastern connections.",
			bunnyVideoID: "test-video-4",
			thumbnailURL: "https://example.com/thumb4.jpg",
			category:     "Linguistics",
			duration:     1518, // 25:18
			fileSize:     312600000,
			tags:         []string{"linguistics", "hebrew", "names", "ancient-languages"},
			createdBy:    1,
		},
		{
			title:        "Metallurgy in Ancient America: Book of Mormon Evidence",
			description:  "Evidence of advanced metallurgy in pre-Columbian America and its relationship to Book of Mormon descriptions of metalworking.",
			bunnyVideoID: "test-video-5",
			thumbnailURL: "https://example.com/thumb5.jpg",
			category:     "Archaeology",
			duration:     1267, // 21:07
			fileSize:     267400000,
			tags:         []string{"metallurgy", "ancient-technology", "archaeology", "metals"},
			createdBy:    1,
		},
	}

	// Insert videos
	for i, video := range testVideos {
		// Convert tags array to string for database storage
		tagsStr := "{" + fmt.Sprintf("%q", video.tags[0])
		for _, tag := range video.tags[1:] {
			tagsStr += "," + fmt.Sprintf("%q", tag)
		}
		tagsStr += "}"

		query := `
			INSERT INTO videos (
				title, description, bunny_video_id, thumbnail_url, 
				duration, file_size, status, category, tags, 
				view_count, like_count, created_by, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
			ON CONFLICT (bunny_video_id) DO NOTHING
		`

		now := time.Now()
		_, err := db.Exec(query,
			video.title,
			video.description,
			video.bunnyVideoID,
			video.thumbnailURL,
			video.duration,
			video.fileSize,
			"ready", // status
			video.category,
			tagsStr,
			0, // view_count
			0, // like_count
			video.createdBy,
			now,
			now,
		)

		if err != nil {
			log.Printf("Failed to insert video %d: %v", i+1, err)
		} else {
			fmt.Printf("Inserted video: %s\n", video.title)
		}
	}

	fmt.Println("Database seeding completed!")
}
