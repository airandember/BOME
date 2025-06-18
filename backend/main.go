package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bome-backend/internal/config"
	"bome-backend/internal/database"
	"bome-backend/internal/middleware"
	"bome-backend/internal/routes"
	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize configuration
	cfg := config.New()

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize database (try PostgreSQL, fallback to development mode)
	var db *database.DB
	var err error

	if cfg.DBHost != "" && cfg.DBPassword != "" {
		db, err = database.New(cfg)
		if err != nil {
			log.Printf("Failed to connect to PostgreSQL: %v", err)
			log.Println("Continuing in development mode without database...")
			db = nil
		} else {
			defer db.Close()
		}
	} else {
		log.Println("PostgreSQL not configured, continuing in development mode without database...")
		db = nil
	}

	// Initialize Redis (skip if not configured for development)
	var redis *database.Redis
	if cfg.RedisHost != "" {
		redis, err = database.NewRedis(cfg)
		if err != nil {
			log.Printf("Failed to connect to Redis: %v", err)
			log.Println("Continuing without Redis for development...")
		} else {
			defer redis.Close()
		}
	} else {
		log.Println("Redis not configured, continuing without Redis for development...")
	}

	// Initialize services
	bunnyService := services.NewBunnyService()
	stripeService := services.NewStripeService()
	spacesService, err := services.NewSpacesService()
	if err != nil {
		log.Printf("Failed to initialize Spaces service: %v", err)
		log.Println("Continuing without Spaces service for development...")
		spacesService = nil
	}
	emailService := services.NewEmailService()

	// Create Gin router
	router := gin.New()

	// Add middleware
	router.Use(middleware.Logger())
	router.Use(middleware.CORS(cfg))
	router.Use(middleware.Recovery())
	router.Use(middleware.RateLimit(cfg))

	// Setup routes
	routes.SetupRoutes(router, cfg, db, redis, bunnyService, stripeService, spacesService, emailService)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.ServerPort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give outstanding requests a deadline for completion
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
