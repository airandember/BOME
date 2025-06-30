/*
 * BOME - Book of Mormon Evidence Hub
 * Copyright © 2024 BOME Development Team. All Rights Reserved.
 *
 * PROPRIETARY AND CONFIDENTIAL
 * This software is the exclusive property of the copyright holder.
 * Unauthorized use, reproduction, or distribution is strictly prohibited.
 * See LICENSE file for full terms and conditions.
 */

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

	// Initialize database (PostgreSQL)
	var db *database.DB
	var err error

	log.Println("Attempting to initialize PostgreSQL database...")
	db, err = database.New(cfg)
	if err != nil {
		log.Printf("Failed to connect to PostgreSQL database: %v", err)
		log.Println("Continuing without database...")
		db = nil
	} else {
		log.Println("Database connection successful")
		defer db.Close()
		// Run migrations
		if err := db.RunMigrations(); err != nil {
			log.Printf("Failed to run migrations: %v", err)
		} else {
			log.Println("Database migrations completed successfully")
		}
	}

	// Skip Redis for development (not needed for basic functionality)
	var redis *database.Redis
	log.Println("Skipping Redis for development...")
	redis = nil

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
	router.Use(middleware.RecoveryWithLogging())
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.CORS())
	router.Use(middleware.RateLimiting())

	// Setup routes
	log.Println("Setting up routes...")
	routes.SetupRoutes(router, cfg, db, redis, bunnyService, stripeService, spacesService, emailService)
	log.Println("Routes setup completed successfully")

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
