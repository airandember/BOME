package main

import (
	"bome-backend/internal/config"
	"bome-backend/internal/database"
	"log"
)

func main() {
	// Load configuration
	cfg := config.New()

	// Connect to database
	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("🔍 Checking analytics tables...")

	// Check if analytics_events table exists
	var exists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'analytics_events'
		)
	`).Scan(&exists)
	if err != nil {
		log.Fatalf("Failed to check analytics_events table: %v", err)
	}

	if !exists {
		log.Println("❌ analytics_events table does not exist. Creating...")

		// Create analytics_events table
		_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS analytics_events (
				id SERIAL PRIMARY KEY,
				event_type VARCHAR(100) NOT NULL,
				user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
				session_id VARCHAR(255),
				subsite VARCHAR(50) DEFAULT 'streaming',
				event_data TEXT,
				ip_address INET,
				user_agent TEXT,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`)
		if err != nil {
			log.Fatalf("Failed to create analytics_events table: %v", err)
		}
		log.Println("✅ analytics_events table created")
	} else {
		log.Println("✅ analytics_events table exists")
	}

	// Check if user_metrics table exists
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'user_metrics'
		)
	`).Scan(&exists)
	if err != nil {
		log.Fatalf("Failed to check user_metrics table: %v", err)
	}

	if !exists {
		log.Println("❌ user_metrics table does not exist. Creating...")

		// Create user_metrics table
		_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS user_metrics (
				id SERIAL PRIMARY KEY,
				user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
				date DATE NOT NULL,
				session_count INTEGER DEFAULT 0,
				session_duration INTEGER DEFAULT 0,
				page_views INTEGER DEFAULT 0,
				video_views INTEGER DEFAULT 0,
				video_watch_time INTEGER DEFAULT 0,
				likes_given INTEGER DEFAULT 0,
				comments_made INTEGER DEFAULT 0,
				shares_made INTEGER DEFAULT 0,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				UNIQUE(user_id, date)
			)
		`)
		if err != nil {
			log.Fatalf("Failed to create user_metrics table: %v", err)
		}
		log.Println("✅ user_metrics table created")
	} else {
		log.Println("✅ user_metrics table exists")
	}

	// Check if video_metrics table exists
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'video_metrics'
		)
	`).Scan(&exists)
	if err != nil {
		log.Fatalf("Failed to check video_metrics table: %v", err)
	}

	if !exists {
		log.Println("❌ video_metrics table does not exist. Creating...")

		// Create video_metrics table
		_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS video_metrics (
				id SERIAL PRIMARY KEY,
				video_id INTEGER REFERENCES videos(id) ON DELETE CASCADE,
				date DATE NOT NULL,
				views INTEGER DEFAULT 0,
				unique_views INTEGER DEFAULT 0,
				watch_time INTEGER DEFAULT 0,
				completion_rate DECIMAL(5,2) DEFAULT 0.00,
				likes INTEGER DEFAULT 0,
				comments INTEGER DEFAULT 0,
				shares INTEGER DEFAULT 0,
				bounce_rate DECIMAL(5,2) DEFAULT 0.00,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				UNIQUE(video_id, date)
			)
		`)
		if err != nil {
			log.Fatalf("Failed to create video_metrics table: %v", err)
		}
		log.Println("✅ video_metrics table created")
	} else {
		log.Println("✅ video_metrics table exists")
	}

	// Check if system_metrics table exists
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'system_metrics'
		)
	`).Scan(&exists)
	if err != nil {
		log.Fatalf("Failed to check system_metrics table: %v", err)
	}

	if !exists {
		log.Println("❌ system_metrics table does not exist. Creating...")

		// Create system_metrics table
		_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS system_metrics (
				id SERIAL PRIMARY KEY,
				timestamp TIMESTAMP NOT NULL,
				cpu_usage DECIMAL(5,2) DEFAULT 0.00,
				memory_usage DECIMAL(5,2) DEFAULT 0.00,
				disk_usage DECIMAL(5,2) DEFAULT 0.00,
				network_in BIGINT DEFAULT 0,
				network_out BIGINT DEFAULT 0,
				active_sessions INTEGER DEFAULT 0,
				error_rate DECIMAL(5,4) DEFAULT 0.0000,
				response_time INTEGER DEFAULT 0,
				database_size BIGINT DEFAULT 0,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`)
		if err != nil {
			log.Fatalf("Failed to create system_metrics table: %v", err)
		}
		log.Println("✅ system_metrics table created")
	} else {
		log.Println("✅ system_metrics table exists")
	}

	// Check if webhook_events table exists
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'webhook_events'
		)
	`).Scan(&exists)
	if err != nil {
		log.Fatalf("Failed to check webhook_events table: %v", err)
	}

	if !exists {
		log.Println("❌ webhook_events table does not exist. Creating...")

		// Create webhook_events table
		_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS webhook_events (
				id SERIAL PRIMARY KEY,
				event_type VARCHAR(100) NOT NULL,
				subsite VARCHAR(50) DEFAULT 'streaming',
				endpoint VARCHAR(500) NOT NULL,
				status VARCHAR(20) NOT NULL CHECK (status IN ('success', 'failed', 'pending')),
				response_time INTEGER DEFAULT 0,
				payload_size INTEGER DEFAULT 0,
				status_code INTEGER,
				error_message TEXT,
				retry_count INTEGER DEFAULT 0,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`)
		if err != nil {
			log.Fatalf("Failed to create webhook_events table: %v", err)
		}
		log.Println("✅ webhook_events table created")
	} else {
		log.Println("✅ webhook_events table exists")
	}

	// Check if alerts table exists
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'alerts'
		)
	`).Scan(&exists)
	if err != nil {
		log.Fatalf("Failed to check alerts table: %v", err)
	}

	if !exists {
		log.Println("❌ alerts table does not exist. Creating...")

		// Create alerts table
		_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS alerts (
				id SERIAL PRIMARY KEY,
				severity VARCHAR(20) NOT NULL CHECK (severity IN ('info', 'warning', 'critical')),
				title VARCHAR(255) NOT NULL,
				message TEXT NOT NULL,
				subsite VARCHAR(50),
				acknowledged BOOLEAN DEFAULT FALSE,
				acknowledged_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
				acknowledged_at TIMESTAMP,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`)
		if err != nil {
			log.Fatalf("Failed to create alerts table: %v", err)
		}
		log.Println("✅ alerts table created")
	} else {
		log.Println("✅ alerts table exists")
	}

	// Check if cross_subsite_stats table exists
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'cross_subsite_stats'
		)
	`).Scan(&exists)
	if err != nil {
		log.Fatalf("Failed to check cross_subsite_stats table: %v", err)
	}

	if !exists {
		log.Println("❌ cross_subsite_stats table does not exist. Creating...")

		// Create cross_subsite_stats table
		_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS cross_subsite_stats (
				id SERIAL PRIMARY KEY,
				date DATE NOT NULL,
				subsite VARCHAR(50) NOT NULL,
				users INTEGER DEFAULT 0,
				content INTEGER DEFAULT 0,
				views INTEGER DEFAULT 0,
				revenue DECIMAL(10,2) DEFAULT 0.00,
				engagement_rate DECIMAL(5,4) DEFAULT 0.0000,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				UNIQUE(date, subsite)
			)
		`)
		if err != nil {
			log.Fatalf("Failed to create cross_subsite_stats table: %v", err)
		}
		log.Println("✅ cross_subsite_stats table created")
	} else {
		log.Println("✅ cross_subsite_stats table exists")
	}

	log.Println("🎉 Analytics tables check completed!")
}
