package main

import (
	"fmt"

	"bome-backend/internal/services"
)

func main() {
	fmt.Println("🧪 Testing Smart Tagging Service")
	fmt.Println("=================================")

	// Create smart tagging service
	taggingService := services.NewSmartTaggingService()

	// Test cases
	testTitles := []string{
		"John Smith - The Ancient Ruins of Mesoamerica and Archaeological Discoveries",
		"Dr. Sarah Johnson - DNA Evidence and Genetic Research in Ancient Populations",
		"Professor Michael Brown - Linguistic Analysis of Mayan Hieroglyphs and Scripts",
		"Archaeological Survey of Ancient Temples and Religious Sites",
		"Virtual Tour of Historical Evidence and Cultural Artifacts",
		"Interview with Expert on Historical Documentation and Records",
		"Lecture Series on Cultural Studies and Anthropological Research",
		"Presentation of Geographic Features and Regional Analysis",
	}

	fmt.Println("\n📝 Testing Title Processing:")
	fmt.Println("-----------------------------")

	for i, title := range testTitles {
		fmt.Printf("\n%d. Original Title: %s\n", i+1, title)

		result := taggingService.GenerateTagsFromTitle(title)

		fmt.Printf("   Name: %s\n", result.Name)
		fmt.Printf("   Tags: %v\n", result.Tags)
		fmt.Printf("   Processed: %s\n", result.ProcessedTitle)

		// Test categorization
		fmt.Printf("   Categories: ")
		for j, tag := range result.Tags {
			if j > 0 {
				fmt.Print(", ")
			}
			category := taggingService.CategorizeTag(tag)
			fmt.Printf("%s(%s)", tag, category)
		}
		fmt.Println()
	}

	fmt.Println("\n🏷️ Available Tag Categories:")
	fmt.Println("-----------------------------")
	categories := taggingService.GetTagCategories()
	for i, category := range categories {
		fmt.Printf("%d. %s\n", i+1, category)
	}

	fmt.Println("\n✅ Smart Tagging Test Complete!")
	fmt.Println("\nThis demonstrates how the system:")
	fmt.Println("1. Extracts names from titles (before '-')")
	fmt.Println("2. Removes common articles and prepositions")
	fmt.Println("3. Generates meaningful tags from remaining words")
	fmt.Println("4. Automatically categorizes tags by subject area")
	fmt.Println("5. Maintains the original title intact")
}
