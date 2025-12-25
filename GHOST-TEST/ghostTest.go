package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Known ghost product IDs from our system
var ghostProductIDs = []string{
	"prod_HEmcX1PE8TO2CO", // Deleted
	"prod_FvNAeI348dup9w", // Deleted
	"prod_HF5YzcBH5Rwr0d", // Deleted
	"prod_GVV5efccnh13h9", // Deleted
	"prod_FvNAJgnw48hwpZ", // Deleted
	"prod_KG0YNos3k94WAS", // Active
	"prod_TVXu8WrKNoTC4A", // Archived
}

// StripeError represents a Stripe API error response
type StripeError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

// StripeProduct represents a Stripe product response
type StripeProduct struct {
	ID          string `json:"id"`
	Object      string `json:"object"`
	Active      bool   `json:"active"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Created     int64  `json:"created"`
	Deleted     bool   `json:"deleted,omitempty"`
}

func main() {
	fmt.Println("=" + strings.Repeat("=", 79))
	fmt.Println("🔍 STRIPE GHOST PRODUCT TEST")
	fmt.Println("=" + strings.Repeat("=", 79))
	fmt.Println()

	// Get API key from environment or prompt
	apiKey := os.Getenv("STRIPE_SECRET_KEY")
	if apiKey == "" {
		fmt.Print("Enter your Stripe Secret Key (sk_live_... or sk_test_...): ")
		fmt.Scanln(&apiKey)
	}

	if apiKey == "" {
		fmt.Println("❌ No API key provided. Exiting.")
		os.Exit(1)
	}

	// Mask key for display
	maskedKey := apiKey[:12] + "..." + apiKey[len(apiKey)-4:]
	fmt.Printf("🔑 Using API Key: %s\n\n", maskedKey)

	// Test each ghost product
	results := make(map[string]string)

	for i, productID := range ghostProductIDs {
		fmt.Printf("[%d/%d] Testing: %s\n", i+1, len(ghostProductIDs), productID)
		fmt.Println(strings.Repeat("-", 60))

		status, details := testProduct(apiKey, productID)
		results[productID] = status

		fmt.Println(details)
		fmt.Println()

		// Rate limiting - be nice to Stripe API
		time.Sleep(500 * time.Millisecond)
	}

	// Summary
	fmt.Println("=" + strings.Repeat("=", 79))
	fmt.Println("📊 SUMMARY")
	fmt.Println("=" + strings.Repeat("=", 79))

	archived := 0
	deleted := 0
	active := 0
	errors := 0

	for productID, status := range results {
		emoji := "❓"
		switch status {
		case "ARCHIVED":
			emoji = "📦"
			archived++
		case "DELETED":
			emoji = "👻"
			deleted++
		case "ACTIVE":
			emoji = "✅"
			active++
		case "ERROR":
			emoji = "❌"
			errors++
		}
		fmt.Printf("%s %s: %s\n", emoji, productID, status)
	}

	fmt.Println()
	fmt.Printf("📊 Results: %d Archived, %d Deleted (Ghost), %d Active, %d Errors\n",
		archived, deleted, active, errors)

	if deleted > 0 {
		fmt.Println()
		fmt.Println("⚠️  GHOST PRODUCTS FOUND!")
		fmt.Println("   These products return 'resource_missing' from Stripe API.")
		fmt.Println("   Send this output to Stripe support as evidence.")
	}

	// Write results to file for sending to Stripe
	writeResultsToFile(results, apiKey)
}

func testProduct(apiKey, productID string) (string, string) {
	url := fmt.Sprintf("https://api.stripe.com/v1/products/%s", productID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "ERROR", fmt.Sprintf("   ❌ Failed to create request: %v", err)
	}

	req.SetBasicAuth(apiKey, "")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "ERROR", fmt.Sprintf("   ❌ Request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "ERROR", fmt.Sprintf("   ❌ Failed to read response: %v", err)
	}

	// Pretty print JSON
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, body, "   ", "  "); err != nil {
		prettyJSON.Write(body) // fallback to raw if indent fails
	}

	// Build output
	var output strings.Builder
	output.WriteString(fmt.Sprintf("   HTTP Status: %d\n", resp.StatusCode))
	output.WriteString(fmt.Sprintf("   Request-ID: %s\n", resp.Header.Get("Request-Id")))
	output.WriteString(fmt.Sprintf("   \n   JSON Response:\n   %s\n", prettyJSON.String()))

	if resp.StatusCode == 200 {
		// Product exists - check if active or archived
		var product StripeProduct
		if err := json.Unmarshal(body, &product); err != nil {
			output.WriteString(fmt.Sprintf("   ❌ Failed to parse response: %v\n", err))
			return "ERROR", output.String()
		}

		if product.Active {
			output.WriteString(fmt.Sprintf("\n   ✅ STATUS: ACTIVE - Name: %s\n", product.Name))
			return "ACTIVE", output.String()
		} else {
			output.WriteString(fmt.Sprintf("\n   📦 STATUS: ARCHIVED - Name: %s\n", product.Name))
			output.WriteString(fmt.Sprintf("   Created: %s\n", time.Unix(product.Created, 0).Format("2006-01-02")))
			return "ARCHIVED", output.String()
		}
	} else if resp.StatusCode == 404 {
		// Product not found - this is a GHOST!
		output.WriteString(fmt.Sprintf("\n   👻 STATUS: GHOST DETECTED! (resource_missing)\n"))
		return "DELETED", output.String()
	} else {
		output.WriteString(fmt.Sprintf("\n   ⚠️  STATUS: Unexpected status code %d\n", resp.StatusCode))
		return "ERROR", output.String()
	}
}

func writeResultsToFile(results map[string]string, apiKey string) {
	filename := fmt.Sprintf("ghost_test_results_%s.txt", time.Now().Format("2006-01-02_150405"))

	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("⚠️  Could not write results file: %v\n", err)
		return
	}
	defer file.Close()

	file.WriteString("STRIPE GHOST PRODUCT TEST RESULTS\n")
	file.WriteString("=================================\n")
	file.WriteString(fmt.Sprintf("Date: %s\n", time.Now().Format("2006-01-02 15:04:05 MST")))
	file.WriteString(fmt.Sprintf("API Key: %s...%s\n\n", apiKey[:12], apiKey[len(apiKey)-4:]))

	file.WriteString("RESULTS:\n")
	for productID, status := range results {
		file.WriteString(fmt.Sprintf("  %s: %s\n", productID, status))
	}

	file.WriteString("\n\nCURL COMMANDS TO REPRODUCE:\n")
	file.WriteString("============================\n")
	for _, productID := range ghostProductIDs {
		file.WriteString(fmt.Sprintf("\ncurl https://api.stripe.com/v1/products/%s \\\n", productID))
		file.WriteString(fmt.Sprintf("  -u \"%s:\"\n", apiKey[:12]+"...REDACTED"))
	}

	fmt.Printf("\n📄 Results saved to: %s\n", filename)
}
