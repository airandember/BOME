package main

import (
	"bome-backend/internal/services"
	"fmt"
)

func main() {
	// Test the date parsing with the format the frontend sends
	testDate := "2025-07-31"

	parsed, err := services.ParseFlexibleDate(testDate)
	if err != nil {
		fmt.Printf("Error parsing date: %v\n", err)
		return
	}

	if parsed != nil {
		fmt.Printf("Successfully parsed date: %s -> %v\n", testDate, parsed)

		// Test the database formatting
		formatted := services.FormatDateForDatabase(parsed, false) // start date
		fmt.Printf("Formatted for database (start): %v, Valid: %v\n", formatted.Time, formatted.Valid)

		formattedEnd := services.FormatDateForDatabase(parsed, true) // end date
		fmt.Printf("Formatted for database (end): %v, Valid: %v\n", formattedEnd.Time, formattedEnd.Valid)
	} else {
		fmt.Printf("Date was nil (empty string)\n")
	}
}
