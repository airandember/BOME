package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	// Test the actual API endpoint
	url := "http://localhost:8080/api/v1/admin/streaming/stripe/database/customers?limit=5&includeSubscriptions=true"

	// Create HTTP client
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}

	// Add headers (we'll need to get a real token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer YOUR_TOKEN_HERE") // This will fail, but let's see the response structure

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Response Body:\n%s\n", string(body))

	// Try to parse JSON to see structure
	if resp.StatusCode == 200 {
		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err == nil {
			fmt.Printf("\nParsed JSON Structure:\n")
			printJSON(result, 0)
		}
	}
}

func printJSON(data interface{}, indent int) {
	prefix := ""
	for i := 0; i < indent; i++ {
		prefix += "  "
	}

	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			fmt.Printf("%s%s: ", prefix, key)
			if nestedMap, ok := value.(map[string]interface{}); ok {
				fmt.Printf("\n")
				printJSON(nestedMap, indent+1)
			} else if nestedArray, ok := value.([]interface{}); ok {
				fmt.Printf("[%d items]\n", len(nestedArray))
				if len(nestedArray) > 0 {
					printJSON(nestedArray[0], indent+1)
				}
			} else {
				fmt.Printf("%v\n", value)
			}
		}
	case []interface{}:
		for i, item := range v {
			fmt.Printf("%s[%d]:\n", prefix, i)
			printJSON(item, indent+1)
		}
	default:
		fmt.Printf("%v\n", v)
	}
}
