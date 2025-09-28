package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Jethzabell/go-dctrack-client/dctrack"
)

func main() {
	// Create DCTrack client with default settings
	client, err := dctrack.New(
		"https://dctrack.company.com/api/v2",
		"your-username",
		"your-password",
	)
	if err != nil {
		log.Fatalf("Failed to create DCTrack client: %v", err)
	}

	ctx := context.Background()

	// Example 1: Get all installed assets
	fmt.Println("=== Getting All Installed Assets ===")
	allItems, err := client.GetItems(ctx)
	if err != nil {
		log.Fatalf("Failed to get items: %v", err)
	}
	fmt.Printf("Found %d total assets\n", len(allItems))

	// Example 2: Filter by location
	fmt.Println("\n=== Getting Assets in RDU2 ===")
	rdu2Items, err := client.GetItemsWithParams(ctx, dctrack.ByLocation("RDU2"))
	if err != nil {
		log.Fatalf("Failed to get RDU2 items: %v", err)
	}
	fmt.Printf("Found %d assets in RDU2\n", len(rdu2Items))

	// Example 3: Search for specific assets
	fmt.Println("\n=== Searching for Dell PowerEdge Servers ===")
	dellServers, err := client.SearchItems(ctx, "PowerEdge")
	if err != nil {
		log.Fatalf("Failed to search items: %v", err)
	}
	fmt.Printf("Found %d Dell PowerEdge servers\n", len(dellServers))

	// Example 4: Get specific asset by ID
	if len(allItems) > 0 {
		firstItemID := allItems[0].ID
		fmt.Printf("\n=== Getting Asset Details for ID: %s ===\n", firstItemID)

		item, err := client.GetItemByID(ctx, firstItemID)
		if err != nil {
			log.Printf("Failed to get item by ID: %v", err)
		} else {
			fmt.Printf("Asset: %s\n", item.Name)
			fmt.Printf("Location: %s\n", item.Location)
			fmt.Printf("Make: %s\n", item.Make)
			fmt.Printf("Model: %s\n", item.Model)
			fmt.Printf("Power: %.2f W\n", item.OriginalPower)
		}
	}

	// Example 5: Simple analytics - power summary by location
	fmt.Println("\n=== Power Summary Analysis ===")
	locationPower := make(map[string]float64)
	locationCount := make(map[string]int)

	for _, item := range allItems {
		if item.OriginalPower > 0 {
			locationPower[item.Location] += item.OriginalPower
			locationCount[item.Location]++
		}
	}

	fmt.Println("Top 5 locations by power consumption:")
	// Simple display of top locations
	count := 0
	for location, power := range locationPower {
		if count < 5 {
			assets := locationCount[location]
			avgPower := power / float64(assets)
			fmt.Printf("  %s: %.2f kW total, %d assets, %.2f W avg\n",
				location, power/1000, assets, avgPower)
			count++
		}
	}
}
