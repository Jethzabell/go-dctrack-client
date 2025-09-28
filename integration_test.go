package dctrack

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
)

// TestRealDCTrackIntegration tests against a real DCTrack instance
// This test is skipped unless DCTRACK_INTEGRATION_TEST=true is set
func TestRealDCTrackIntegration(t *testing.T) {
	if os.Getenv("DCTRACK_INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test. Set DCTRACK_INTEGRATION_TEST=true to run.")
	}

	// Get configuration from environment
	url := os.Getenv("DCTRACK_URL")
	username := os.Getenv("DCTRACK_USERNAME")
	password := os.Getenv("DCTRACK_PASSWORD")

	if url == "" || username == "" || password == "" {
		t.Skip("Missing required environment variables: DCTRACK_URL, DCTRACK_USERNAME, DCTRACK_PASSWORD")
	}

	// Create client with test configuration
	config := Config{
		URL:              url,
		Username:         username,
		Password:         password,
		PageSize:         10, // Small page size for testing
		MaxRetries:       3,
		RetryDelay:       time.Second,
		VerifySSL:        false, // Often false for internal DCTrack instances
		RequestAllFields: false, // Faster for testing
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Set up logger for detailed output
	logger, _ := zap.NewDevelopment()
	client.SetLogger(logger)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	t.Run("GetItems", func(t *testing.T) {
		items, err := client.GetItems(ctx)
		if err != nil {
			t.Fatalf("Failed to get items: %v", err)
		}

		if len(items) == 0 {
			t.Skip("No items returned from DCTrack (empty database?)")
		}

		t.Logf("Successfully retrieved %d items", len(items))

		// Validate first item structure
		item := items[0]
		if item.ID == "" {
			t.Errorf("First item has empty ID")
		}
		if item.Name == "" {
			t.Errorf("First item has empty Name")
		}

		t.Logf("First item: ID=%s, Name=%s, Location=%s", item.ID, item.Name, item.Location)
	})

	t.Run("SearchItems", func(t *testing.T) {
		// Test search functionality
		items, err := client.SearchItems(ctx, "Server")
		if err != nil {
			t.Fatalf("Failed to search items: %v", err)
		}

		t.Logf("Search returned %d items", len(items))

		// Validate search results contain the search term (case insensitive)
		for _, item := range items {
			found := false
			searchTerm := "server"
			if containsIgnoreCase(item.Name, searchTerm) ||
				containsIgnoreCase(item.Model, searchTerm) ||
				containsIgnoreCase(item.Type, searchTerm) {
				found = true
			}
			if !found && len(items) > 0 {
				t.Logf("Warning: Item %s may not match search term 'Server'", item.Name)
			}
		}
	})

	t.Run("GetItemsWithParams", func(t *testing.T) {
		// Test filtered queries
		params := ItemsParams{
			Status:   "Installed",
			PageSize: 5, // Small page for testing
		}

		items, err := client.GetItemsWithParams(ctx, params)
		if err != nil {
			t.Fatalf("Failed to get items with params: %v", err)
		}

		t.Logf("Filtered query returned %d items", len(items))

		// Verify all items have "Installed" status
		for _, item := range items {
			if item.Status != "Installed" && item.Status != "" {
				t.Errorf("Expected status 'Installed' or empty, got '%s'", item.Status)
			}
		}
	})

	t.Run("GetItemByID", func(t *testing.T) {
		// First get some items to get a valid ID
		items, err := client.GetItemsWithParams(ctx, ItemsParams{PageSize: 1})
		if err != nil {
			t.Fatalf("Failed to get items for ID test: %v", err)
		}

		if len(items) == 0 {
			t.Skip("No items available for GetItemByID test")
		}

		itemID := items[0].ID
		if itemID == "" {
			t.Skip("First item has no ID for GetItemByID test")
		}

		// Test getting item by ID
		item, err := client.GetItemByID(ctx, itemID)
		if err != nil {
			t.Fatalf("Failed to get item by ID %s: %v", itemID, err)
		}

		if item.ID != itemID {
			t.Errorf("Expected ID %s, got %s", itemID, item.ID)
		}

		t.Logf("Successfully retrieved item by ID: %s", item.Name)
	})

	t.Run("FilterBuilder", func(t *testing.T) {
		// Test fluent filter builder
		filters := NewFilterBuilder().
			Status("Installed").
			WithPagination(1, 3).
			Build()

		items, err := client.GetItemsWithParams(ctx, filters)
		if err != nil {
			t.Fatalf("Failed to get items with filter builder: %v", err)
		}

		t.Logf("Filter builder returned %d items", len(items))

		// Should return at most 3 items due to pagination
		if len(items) > 3 {
			t.Errorf("Expected at most 3 items due to page size, got %d", len(items))
		}
	})

	t.Run("PresetFilters", func(t *testing.T) {
		// Test preset filter functions
		installedItems, err := client.GetItemsWithParams(ctx, InstalledOnly())
		if err != nil {
			t.Fatalf("Failed to get installed items: %v", err)
		}

		t.Logf("InstalledOnly filter returned %d items", len(installedItems))

		// All items should be installed
		for _, item := range installedItems {
			if item.Status != "Installed" && item.Status != "" {
				t.Errorf("Expected installed item, got status '%s'", item.Status)
			}
		}
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		// Test error handling with invalid item ID
		_, err := client.GetItemByID(ctx, "non-existent-id-12345")
		if err == nil {
			t.Errorf("Expected error for non-existent item ID")
		}

		// Test error handling with empty search
		_, err = client.SearchItems(ctx, "")
		if err == nil {
			t.Errorf("Expected error for empty search query")
		}
	})

	t.Run("PowerAnalysis", func(t *testing.T) {
		// Test power-related data analysis
		items, err := client.GetItemsWithParams(ctx, ItemsParams{PageSize: 10})
		if err != nil {
			t.Fatalf("Failed to get items for power analysis: %v", err)
		}

		if len(items) == 0 {
			t.Skip("No items for power analysis")
		}

		totalPower := 0.0
		powerAssets := 0
		for _, item := range items {
			if item.OriginalPower > 0 {
				totalPower += item.OriginalPower
				powerAssets++
			}
		}

		t.Logf("Power analysis: %d assets with power data, %.2f kW total",
			powerAssets, totalPower/1000)

		if powerAssets > 0 {
			avgPower := totalPower / float64(powerAssets)
			t.Logf("Average power per asset: %.2f W", avgPower)
		}
	})

	t.Run("PaginationTest", func(t *testing.T) {
		// Test pagination functionality
		page1, err := client.GetItemsWithParams(ctx, ItemsParams{
			PageNumber: 1,
			PageSize:   2,
		})
		if err != nil {
			t.Fatalf("Failed to get page 1: %v", err)
		}

		page2, err := client.GetItemsWithParams(ctx, ItemsParams{
			PageNumber: 2,
			PageSize:   2,
		})
		if err != nil {
			t.Fatalf("Failed to get page 2: %v", err)
		}

		t.Logf("Page 1: %d items, Page 2: %d items", len(page1), len(page2))

		// Pages should not contain the same items (if there are enough items)
		if len(page1) > 0 && len(page2) > 0 {
			for _, item1 := range page1 {
				for _, item2 := range page2 {
					if item1.ID == item2.ID && item1.ID != "" {
						t.Errorf("Found duplicate item %s in both pages", item1.ID)
					}
				}
			}
		}
	})

	t.Run("PerformanceTest", func(t *testing.T) {
		// Test performance with larger page sizes
		start := time.Now()

		items, err := client.GetItemsWithParams(ctx, ItemsParams{
			PageSize: 100, // Larger page size
		})
		if err != nil {
			t.Fatalf("Performance test failed: %v", err)
		}

		duration := time.Since(start)
		t.Logf("Retrieved %d items in %v (%.2f items/second)",
			len(items), duration, float64(len(items))/duration.Seconds())

		// Performance should be reasonable (this is just a sanity check)
		if duration > 30*time.Second {
			t.Logf("Warning: Query took longer than expected: %v", duration)
		}
	})
}

// Helper function for case-insensitive string matching
func containsIgnoreCase(s, substr string) bool {
	s = strings.ToLower(s)
	substr = strings.ToLower(substr)
	return strings.Contains(s, substr)
}
