package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	dctrack "github.com/Jethzabell/go-dctrack-client"
	"go.uber.org/zap"
)

const (
	defaultPageSize = 100
	defaultTimeout  = 30 * time.Second
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Get configuration from environment variables
	config := getConfigFromEnv()
	if err := validateConfig(config); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Create logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// Create DCTrack client
	client, err := dctrack.NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create DCTrack client: %v", err)
	}
	client.SetLogger(logger)
	defer client.Close()

	// Parse command line arguments
	searchQuery := os.Args[1]
	command := "search"
	if len(os.Args) > 2 {
		command = os.Args[1]
		searchQuery = os.Args[2]
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	switch command {
	case "search":
		handleSearch(ctx, client, searchQuery)
	case "list":
		handleList(ctx, client, searchQuery)
	case "item":
		handleGetItem(ctx, client, searchQuery)
	case "power":
		handlePowerAnalysis(ctx, client, searchQuery)
	default:
		handleSearch(ctx, client, command) // Treat as search query
	}
}

func getConfigFromEnv() dctrack.Config {
	config := dctrack.Config{
		URL:              getEnvOrDefault("DCTRACK_URL", ""),
		Username:         getEnvOrDefault("DCTRACK_USERNAME", ""),
		Password:         getEnvOrDefault("DCTRACK_PASSWORD", ""),
		PageSize:         getEnvIntOrDefault("DCTRACK_PAGE_SIZE", defaultPageSize),
		MaxRetries:       getEnvIntOrDefault("DCTRACK_MAX_RETRIES", 3),
		RetryDelay:       time.Duration(getEnvIntOrDefault("DCTRACK_RETRY_DELAY", 1)) * time.Second,
		VerifySSL:        getEnvBoolOrDefault("DCTRACK_VERIFY_SSL", true),
		RequestAllFields: getEnvBoolOrDefault("DCTRACK_ALL_FIELDS", false),
	}

	return config
}

func validateConfig(config dctrack.Config) error {
	if config.URL == "" {
		return fmt.Errorf("DCTRACK_URL environment variable is required")
	}
	if config.Username == "" {
		return fmt.Errorf("DCTRACK_USERNAME environment variable is required")
	}
	if config.Password == "" {
		return fmt.Errorf("DCTRACK_PASSWORD environment variable is required")
	}
	return nil
}

func handleSearch(ctx context.Context, client *dctrack.Client, query string) {
	fmt.Printf("Searching DCTrack for: %s\n", query)
	fmt.Println("=====================================")

	items, err := client.SearchItems(ctx, query)
	if err != nil {
		log.Fatalf("Search failed: %v", err)
	}

	if len(items) == 0 {
		fmt.Println("No items found")
		return
	}

	fmt.Printf("Found %d items:\n\n", len(items))
	for i, item := range items {
		if i >= 10 { // Limit output
			fmt.Printf("... and %d more items\n", len(items)-10)
			break
		}
		printItemSummary(item)
	}
}

func handleList(ctx context.Context, client *dctrack.Client, location string) {
	fmt.Printf("Listing DCTrack items in location: %s\n", location)
	fmt.Println("==========================================")

	params := dctrack.ByLocation(location)
	params.PageSize = defaultPageSize

	items, err := client.GetItemsWithParams(ctx, params)
	if err != nil {
		log.Fatalf("List failed: %v", err)
	}

	if len(items) == 0 {
		fmt.Printf("No items found in location: %s\n", location)
		return
	}

	fmt.Printf("Found %d items in %s:\n\n", len(items), location)

	// Group by vendor
	vendorCounts := make(map[string]int)
	totalPower := 0.0

	for _, item := range items {
		if item.Make != "" {
			vendorCounts[item.Make]++
		}
		totalPower += item.OriginalPower
	}

	fmt.Println("Vendor Distribution:")
	for vendor, count := range vendorCounts {
		fmt.Printf("  %s: %d assets\n", vendor, count)
	}

	fmt.Printf("\nTotal Power: %.2f kW\n", totalPower/1000)
	fmt.Printf("Average Power per Asset: %.2f W\n", totalPower/float64(len(items)))
}

func handleGetItem(ctx context.Context, client *dctrack.Client, itemID string) {
	fmt.Printf("Getting DCTrack item: %s\n", itemID)
	fmt.Println("===========================")

	item, err := client.GetItemByID(ctx, itemID)
	if err != nil {
		log.Fatalf("Failed to get item: %v", err)
	}

	printItemDetails(*item)
}

func handlePowerAnalysis(ctx context.Context, client *dctrack.Client, location string) {
	fmt.Printf("Power analysis for location: %s\n", location)
	fmt.Println("==================================")

	params := dctrack.ByLocation(location)
	items, err := client.GetItemsWithParams(ctx, params)
	if err != nil {
		log.Fatalf("Power analysis failed: %v", err)
	}

	if len(items) == 0 {
		fmt.Printf("No items found in location: %s\n", location)
		return
	}

	// Calculate power statistics
	totalPower := 0.0
	powerAssets := 0
	maxPower := 0.0
	minPower := float64(^uint(0) >> 1) // Max float64

	for _, item := range items {
		if item.OriginalPower > 0 {
			totalPower += item.OriginalPower
			powerAssets++
			if item.OriginalPower > maxPower {
				maxPower = item.OriginalPower
			}
			if item.OriginalPower < minPower {
				minPower = item.OriginalPower
			}
		}
	}

	avgPower := 0.0
	if powerAssets > 0 {
		avgPower = totalPower / float64(powerAssets)
	}

	fmt.Printf("Total Assets: %d\n", len(items))
	fmt.Printf("Assets with Power Data: %d\n", powerAssets)
	fmt.Printf("Total Power: %.2f kW\n", totalPower/1000)
	fmt.Printf("Average Power: %.2f W\n", avgPower)
	fmt.Printf("Maximum Power: %.2f W\n", maxPower)
	if minPower != float64(^uint(0)>>1) {
		fmt.Printf("Minimum Power: %.2f W\n", minPower)
	}
	fmt.Printf("Power Density: %.2f W/asset\n", totalPower/float64(len(items)))
}

func printItemSummary(item dctrack.DCTrackItem) {
	fmt.Printf("ID: %s\n", item.ID)
	fmt.Printf("Name: %s\n", item.Name)
	fmt.Printf("Location: %s\n", item.Location)
	fmt.Printf("Make/Model: %s %s\n", item.Make, item.Model)
	if item.OriginalPower > 0 {
		fmt.Printf("Power: %.0f W\n", item.OriginalPower)
	}
	fmt.Println()
}

func printItemDetails(item dctrack.DCTrackItem) {
	fmt.Printf("ID: %s\n", item.ID)
	fmt.Printf("Name: %s\n", item.Name)
	fmt.Printf("Status: %s\n", item.Status)
	fmt.Printf("Type/Class: %s / %s\n", item.Type, item.ItemClass)
	fmt.Printf("Location: %s\n", item.Location)
	fmt.Printf("Cabinet: %s\n", item.Cabinet)
	fmt.Printf("Position: %s (Height: %d RU)\n", item.Position, item.Height)
	fmt.Printf("Make/Model: %s %s\n", item.Make, item.Model)
	fmt.Printf("Serial Number: %s\n", item.SerialNumber)

	if item.OriginalPower > 0 {
		fmt.Printf("Power: %.0f W\n", item.OriginalPower)
	}
	if item.TiAssetTag != "" {
		fmt.Printf("Asset Tag: %s\n", item.TiAssetTag)
	}
	if item.PrimaryContact != "" {
		fmt.Printf("Primary Contact: %s\n", item.PrimaryContact)
	}
	if item.SystemAdminTeam != "" {
		fmt.Printf("Admin Team: %s\n", item.SystemAdminTeam)
	}
	if item.InstallDate != nil {
		fmt.Printf("Install Date: %s\n", item.InstallDate.Format("2006-01-02"))
	}
	if item.ContractEndDate != nil {
		fmt.Printf("Contract End: %s\n", item.ContractEndDate.Format("2006-01-02"))
	}
}

func printUsage() {
	fmt.Println("dctrackcheck - DCTrack API testing tool")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  dctrackcheck <search-query>                 # Search for items")
	fmt.Println("  dctrackcheck search <query>                 # Search for items")
	fmt.Println("  dctrackcheck list <location>                # List items by location")
	fmt.Println("  dctrackcheck item <item-id>                 # Get specific item details")
	fmt.Println("  dctrackcheck power <location>               # Power analysis for location")
	fmt.Println("")
	fmt.Println("Environment Variables:")
	fmt.Println("  DCTRACK_URL        DCTrack API URL (required)")
	fmt.Println("  DCTRACK_USERNAME   DCTrack username (required)")
	fmt.Println("  DCTRACK_PASSWORD   DCTrack password (required)")
	fmt.Println("  DCTRACK_PAGE_SIZE  Page size for API requests (default: 100)")
	fmt.Println("  DCTRACK_VERIFY_SSL Verify SSL certificates (default: true)")
	fmt.Println("  DCTRACK_ALL_FIELDS Request all fields (default: false)")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  dctrackcheck PowerEdge                     # Search for PowerEdge servers")
	fmt.Println("  dctrackcheck list RDU2                     # List all items in RDU2")
	fmt.Println("  dctrackcheck item 12345                    # Get details for item 12345")
	fmt.Println("  dctrackcheck power RDU2                    # Power analysis for RDU2")
}

// Helper functions
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

func getEnvBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}
