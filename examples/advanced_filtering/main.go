package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Jethzabell/go-dctrack-client/dctrack"
)

func main() {
	// Create DCTrack client with custom configuration
	client, err := dctrack.New(
		"https://dctrack.company.com/api/v2",
		"your-username",
		"your-password",
		dctrack.WithPageSize(500),   // Smaller pages for faster response
		dctrack.WithMaxRetries(5),   // More retries for reliability
		dctrack.WithLimitedFields(), // Only essential fields for performance
	)
	if err != nil {
		log.Fatalf("Failed to create DCTrack client: %v", err)
	}

	ctx := context.Background()

	// Example 1: Complex filtering with parameter struct
	fmt.Println("=== Advanced Filtering with Parameters ===")
	params := dctrack.ItemsParams{
		Location:  "RDU2",
		Status:    "Installed",
		ItemClass: "Device",
		Make:      "Dell",
		PageSize:  50,
	}

	dellDevices, err := client.GetItemsWithParams(ctx, params)
	if err != nil {
		log.Fatalf("Failed to get Dell devices: %v", err)
	}
	fmt.Printf("Found %d Dell devices in RDU2\n", len(dellDevices))

	// Example 2: Using the fluent filter builder
	fmt.Println("\n=== Using Filter Builder (Fluent Interface) ===")
	filters := dctrack.NewFilterBuilder().
		Location("RDU2").
		Status("Installed").
		Make("HPE").
		ItemClass("Device").
		WithPagination(1, 100).
		Build()

	hpeDevices, err := client.GetItemsWithParams(ctx, filters)
	if err != nil {
		log.Fatalf("Failed to get HPE devices: %v", err)
	}
	fmt.Printf("Found %d HPE devices in RDU2\n", len(hpeDevices))

	// Example 3: Multiple vendor comparison
	fmt.Println("\n=== Vendor Analysis ===")
	vendors := []string{"Dell", "HPE", "Cisco", "IBM"}

	for _, vendor := range vendors {
		vendorItems, err := client.GetItemsWithParams(ctx, dctrack.ByVendor(vendor))
		if err != nil {
			log.Printf("Failed to get %s items: %v", vendor, err)
			continue
		}

		// Calculate power statistics
		totalPower := 0.0
		powerAssets := 0
		for _, item := range vendorItems {
			if item.OriginalPower > 0 {
				totalPower += item.OriginalPower
				powerAssets++
			}
		}

		avgPower := 0.0
		if powerAssets > 0 {
			avgPower = totalPower / float64(powerAssets)
		}

		fmt.Printf("  %s: %d assets, %.2f kW total, %.2f W average\n",
			vendor, len(vendorItems), totalPower/1000, avgPower)
	}

	// Example 4: Location-based analysis
	fmt.Println("\n=== Location Analysis ===")
	locations := []string{"RDU2", "RDU3", "BOS2", "TLV2", "PNQ2"}

	for _, location := range locations {
		locationItems, err := client.GetItemsWithParams(ctx, dctrack.ByLocation(location))
		if err != nil {
			log.Printf("Failed to get %s items: %v", location, err)
			continue
		}

		// Asset type breakdown
		deviceCount := 0
		networkCount := 0
		for _, item := range locationItems {
			if item.ItemClass == "Device" {
				deviceCount++
			} else if item.ItemClass == "Network" {
				networkCount++
			}
		}

		fmt.Printf("  %s: %d total (%d devices, %d network)\n",
			location, len(locationItems), deviceCount, networkCount)
	}

	// Example 5: Advanced search patterns
	fmt.Println("\n=== Advanced Search Patterns ===")

	// Search for specific server models
	r730Servers, err := client.SearchItems(ctx, "PowerEdge R730")
	if err != nil {
		log.Printf("Failed to search R730 servers: %v", err)
	} else {
		fmt.Printf("Found %d PowerEdge R730 servers\n", len(r730Servers))
	}

	// Search for network equipment
	switches, err := client.SearchItems(ctx, "switch")
	if err != nil {
		log.Printf("Failed to search switches: %v", err)
	} else {
		fmt.Printf("Found %d switches\n", len(switches))
	}

	// Example 6: Error handling and edge cases
	fmt.Println("\n=== Error Handling Examples ===")

	// Try to get non-existent item
	_, err = client.GetItemByID(ctx, "non-existent-id")
	if err != nil {
		fmt.Printf("Expected error for non-existent ID: %v\n", err)
	}

	// Try empty search
	emptyResults, err := client.SearchItems(ctx, "")
	if err != nil {
		fmt.Printf("Search with empty query: %v\n", err)
	} else {
		fmt.Printf("Empty search returned %d items\n", len(emptyResults))
	}

	fmt.Println("\n=== Advanced Filtering Complete ===")
}
