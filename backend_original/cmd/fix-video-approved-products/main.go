package main

import (
	"fmt"
	"log"

	"bome-backend/internal/config"
	"bome-backend/internal/database"
)

func main() {
	fmt.Println("🎯 BOME Video Access Fix - Updating Product Video Approval")
	fmt.Println("========================================================")

	// Load configuration
	cfg := config.New()

	// Connect to database
	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Check current status
	fmt.Println("\n📊 Current Video Approved Products:")
	checkCurrentStatus(db)

	// Ask for confirmation
	fmt.Print("\n❓ Do you want to update current subscription products to enable video access? (y/N): ")
	var response string
	fmt.Scanln(&response)

	if response != "y" && response != "Y" {
		fmt.Println("❌ Operation cancelled.")
		return
	}

	// Update the current subscription products
	fmt.Println("\n🔧 Updating video_approved for current subscription products...")

	// Products that should have video access (current active subscription products)
	productsToUpdate := []struct {
		ID   int
		Name string
	}{
		{993, "Monthly Plan"},
		{999, "Annual Plan"},
		{13, "premium yearly"},
		{14, "Premium Monthly"},
		{15, "Premium Annual"},
		{16, "Premium Yearly"},
		{17, "Premium Semi-Annual"},
	}

	for _, product := range productsToUpdate {
		err := updateProductVideoApproved(db, product.ID, product.Name)
		if err != nil {
			log.Printf("❌ Failed to update product %d (%s): %v", product.ID, product.Name, err)
		} else {
			fmt.Printf("✅ Updated product %d: %s\n", product.ID, product.Name)
		}
	}

	// Check final status
	fmt.Println("\n📊 Updated Video Approved Products:")
	checkCurrentStatus(db)

	fmt.Println("\n🎉 Video access fix completed!")
	fmt.Println("Users with active subscriptions should now have video access.")
}

func checkCurrentStatus(db *database.DB) {
	query := `
		SELECT id, name, video_approved, active
		FROM stripe_products 
		WHERE active = true 
		AND (
			name ILIKE '%plan%' 
			OR name ILIKE '%premium%'
			OR name ILIKE '%basic%'
		)
		ORDER BY id DESC`

	rows, err := db.DB.Query(query)
	if err != nil {
		log.Printf("Error checking current status: %v", err)
		return
	}
	defer rows.Close()

	fmt.Printf("%-5s %-30s %-15s %-8s\n", "ID", "Name", "Video Approved", "Active")
	fmt.Println("----------------------------------------------------------------")

	for rows.Next() {
		var id int
		var name string
		var videoApproved, active bool

		err := rows.Scan(&id, &name, &videoApproved, &active)
		if err != nil {
			continue
		}

		approvedIcon := "❌"
		if videoApproved {
			approvedIcon = "✅"
		}

		activeIcon := "❌"
		if active {
			activeIcon = "✅"
		}

		fmt.Printf("%-5d %-30s %-15s %-8s\n", id, name, approvedIcon, activeIcon)
	}
}

func updateProductVideoApproved(db *database.DB, productID int, productName string) error {
	query := `
		UPDATE stripe_products 
		SET video_approved = true 
		WHERE id = $1`

	result, err := db.DB.Exec(query, productID)
	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no rows updated - product ID %d may not exist", productID)
	}

	return nil
}
