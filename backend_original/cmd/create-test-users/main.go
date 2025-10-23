package main

import (
	"database/sql"
	"log"
	"os"

	"bome-backend/internal/database"
	"bome-backend/internal/services"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// TestUser represents a test user with role information
type TestUser struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
	RoleID    string
	RoleName  string
	Level     int
	Category  string
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Database connection parameters
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbName := getEnv("DB_NAME", "bome_db")
	dbUser := getEnv("DB_USER", "bome_admin")
	dbPassword := getEnv("DB_PASSWORD", "AdminBOME")

	// Connect to PostgreSQL database
	connStr := "host=" + dbHost + " port=" + dbPort + " user=" + dbUser + " password=" + dbPassword + " dbname=" + dbName + " sslmode=disable"

	postgresDB, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer postgresDB.Close()

	// Test connection
	if err := postgresDB.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("✅ Connected to PostgreSQL database successfully")

	// Create database wrapper
	db := &database.DB{DB: postgresDB}

	// Define test users with standardized roles
	testUsers := []TestUser{
		// System Administration (Level 10-9)
		{
			Email:     "super.admin@bome.test",
			Password:  "SuperAdmin123!",
			FirstName: "Super",
			LastName:  "Administrator",
			RoleID:    "super_admin",
			RoleName:  "Super Administrator",
			Level:     10,
			Category:  "system",
		},
		{
			Email:     "system.admin@bome.test",
			Password:  "SystemAdmin123!",
			FirstName: "System",
			LastName:  "Administrator",
			RoleID:    "system_admin",
			RoleName:  "System Administrator",
			Level:     9,
			Category:  "system",
		},

		// Content Management (Level 8-6)
		{
			Email:     "content.manager@bome.test",
			Password:  "ContentManager123!",
			FirstName: "Sarah",
			LastName:  "Johnson",
			RoleID:    "content_manager",
			RoleName:  "Content Manager",
			Level:     8,
			Category:  "content",
		},
		{
			Email:     "content.editor@bome.test",
			Password:  "ContentEditor123!",
			FirstName: "Michael",
			LastName:  "Chen",
			RoleID:    "content_editor",
			RoleName:  "Content Editor",
			Level:     7,
			Category:  "content",
		},
		{
			Email:     "content.creator@bome.test",
			Password:  "ContentCreator123!",
			FirstName: "David",
			LastName:  "Thompson",
			RoleID:    "content_creator",
			RoleName:  "Content Creator",
			Level:     6,
			Category:  "content",
		},

		// Subsystem-Specific Roles (Level 7)
		{
			Email:     "articles.manager@bome.test",
			Password:  "ArticlesManager123!",
			FirstName: "Emily",
			LastName:  "Rodriguez",
			RoleID:    "articles_manager",
			RoleName:  "Articles Manager",
			Level:     7,
			Category:  "subsystem",
		},
		{
			Email:     "youtube.manager@bome.test",
			Password:  "YouTubeManager123!",
			FirstName: "Alex",
			LastName:  "Kim",
			RoleID:    "youtube_manager",
			RoleName:  "YouTube Manager",
			Level:     7,
			Category:  "subsystem",
		},
		{
			Email:     "streaming.manager@bome.test",
			Password:  "StreamingManager123!",
			FirstName: "Jessica",
			LastName:  "Wang",
			RoleID:    "streaming_manager",
			RoleName:  "Video Streaming Manager",
			Level:     7,
			Category:  "subsystem",
		},
		{
			Email:     "events.manager@bome.test",
			Password:  "EventsManager123!",
			FirstName: "Robert",
			LastName:  "Davis",
			RoleID:    "events_manager",
			RoleName:  "Events Manager",
			Level:     7,
			Category:  "subsystem",
		},

		// Marketing & Advertising (Level 7-4)
		{
			Email:     "advertisement.manager@bome.test",
			Password:  "AdManager123!",
			FirstName: "Lisa",
			LastName:  "Brown",
			RoleID:    "advertisement_manager",
			RoleName:  "Advertisement Manager",
			Level:     7,
			Category:  "marketing",
		},
		{
			Email:     "marketing.specialist@bome.test",
			Password:  "MarketingSpecialist123!",
			FirstName: "Tom",
			LastName:  "Wilson",
			RoleID:    "marketing_specialist",
			RoleName:  "Marketing Specialist",
			Level:     4,
			Category:  "marketing",
		},

		// User Management (Level 7-5)
		{
			Email:     "user.manager@bome.test",
			Password:  "UserManager123!",
			FirstName: "Rachel",
			LastName:  "Green",
			RoleID:    "user_manager",
			RoleName:  "User Account Manager",
			Level:     7,
			Category:  "user_management",
		},
		{
			Email:     "support.specialist@bome.test",
			Password:  "SupportSpecialist123!",
			FirstName: "Kevin",
			LastName:  "Martinez",
			RoleID:    "support_specialist",
			RoleName:  "Support Specialist",
			Level:     5,
			Category:  "user_management",
		},

		// Analytics & Financial (Level 7)
		{
			Email:     "analytics.manager@bome.test",
			Password:  "AnalyticsManager123!",
			FirstName: "Amanda",
			LastName:  "Taylor",
			RoleID:    "analytics_manager",
			RoleName:  "Analytics Manager",
			Level:     7,
			Category:  "analytics",
		},
		{
			Email:     "financial.admin@bome.test",
			Password:  "FinancialAdmin123!",
			FirstName: "Daniel",
			LastName:  "Anderson",
			RoleID:    "financial_admin",
			RoleName:  "Financial Administrator",
			Level:     7,
			Category:  "financial",
		},

		// Technical & Security (Level 6-5)
		{
			Email:     "security.admin@bome.test",
			Password:  "SecurityAdmin123!",
			FirstName: "Sophie",
			LastName:  "Clark",
			RoleID:    "security_admin",
			RoleName:  "Security Administrator",
			Level:     6,
			Category:  "security",
		},
		{
			Email:     "technical.specialist@bome.test",
			Password:  "TechnicalSpecialist123!",
			FirstName: "Chris",
			LastName:  "Lee",
			RoleID:    "technical_specialist",
			RoleName:  "Technical Specialist",
			Level:     5,
			Category:  "technical",
		},

		// Academic & Research (Level 6-5)
		{
			Email:     "academic.reviewer@bome.test",
			Password:  "AcademicReviewer123!",
			FirstName: "Dr. Rebecca",
			LastName:  "Williams",
			RoleID:    "academic_reviewer",
			RoleName:  "Academic Reviewer",
			Level:     6,
			Category:  "academic",
		},
		{
			Email:     "research.coordinator@bome.test",
			Password:  "ResearchCoordinator123!",
			FirstName: "Dr. James",
			LastName:  "Miller",
			RoleID:    "research_coordinator",
			RoleName:  "Research Coordinator",
			Level:     5,
			Category:  "academic",
		},

		// Base User Roles (Level 3-1)
		{
			Email:     "advertiser@bome.test",
			Password:  "Advertiser123!",
			FirstName: "Maria",
			LastName:  "Garcia",
			RoleID:    "advertiser",
			RoleName:  "Advertiser",
			Level:     3,
			Category:  "base",
		},
		{
			Email:     "user@bome.test",
			Password:  "User123!",
			FirstName: "John",
			LastName:  "Doe",
			RoleID:    "user",
			RoleName:  "User",
			Level:     1,
			Category:  "base",
		},
	}

	log.Println("🚀 Creating test users with standardized roles...")
	log.Println("")

	createdUsers := 0
	skippedUsers := 0

	for _, testUser := range testUsers {
		// Check if user already exists
		exists, err := db.CheckUserExists(testUser.Email)
		if err != nil {
			log.Printf("❌ Error checking if user exists %s: %v", testUser.Email, err)
			continue
		}

		if exists {
			log.Printf("⏭️  User %s already exists, skipping...", testUser.Email)
			skippedUsers++
			continue
		}

		// Hash password
		passwordHash, err := services.HashPassword(testUser.Password)
		if err != nil {
			log.Printf("❌ Failed to hash password for %s: %v", testUser.Email, err)
			continue
		}

		// Create user
		user, err := db.CreateUser(testUser.Email, passwordHash, testUser.FirstName, testUser.LastName, testUser.RoleID)
		if err != nil {
			log.Printf("❌ Failed to create user %s: %v", testUser.Email, err)
			continue
		}

		// Assign role to user
		_, err = db.Exec(
			`INSERT INTO user_roles (user_id, role_id, assigned_at) VALUES ($1, $2, NOW())`,
			user.ID, testUser.RoleID,
		)
		if err != nil {
			log.Printf("❌ Failed to assign role to user %s: %v", testUser.Email, err)
			continue
		}

		// Mark email as verified
		err = db.SetUserEmailVerified(user.ID)
		if err != nil {
			log.Printf("⚠️  Failed to verify email for %s: %v", testUser.Email, err)
		}

		log.Printf("✅ Created user: %s %s (%s)", testUser.FirstName, testUser.LastName, testUser.Email)
		log.Printf("   🔑 Password: %s", testUser.Password)
		log.Printf("   🎭 Role: %s (Level %d, %s)", testUser.RoleName, testUser.Level, testUser.Category)
		log.Printf("   🆔 User ID: %d", user.ID)
		log.Println("")

		createdUsers++
	}

	log.Println("🎉 Test user creation completed!")
	log.Printf("📊 Summary:")
	log.Printf("   ✅ Created: %d users", createdUsers)
	log.Printf("   ⏭️  Skipped: %d users (already exist)", skippedUsers)
	log.Printf("   📝 Total: %d test users", len(testUsers))
	log.Println("")
	log.Println("🌐 You can now test the different dashboards:")
	log.Println("   Frontend: http://localhost:5173")
	log.Println("   Backend API: http://localhost:8080")
	log.Println("")
	log.Println("📋 Test Scenarios:")
	log.Println("   1. Login with super.admin@bome.test (Level 10) - Full access")
	log.Println("   2. Login with content.manager@bome.test (Level 8) - Content management")
	log.Println("   3. Login with advertisement.manager@bome.test (Level 7) - Ad management")
	log.Println("   4. Login with user@bome.test (Level 1) - Basic access")
	log.Println("")
	log.Println("⚠️  IMPORTANT: These are test credentials only!")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
