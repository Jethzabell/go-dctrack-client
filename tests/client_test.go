package dctrack_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Jethzabell/go-dctrack-client/dctrack"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		username    string
		password    string
		expectedErr bool
	}{
		{
			name:        "valid config",
			url:         "https://dctrack.example.com/api/v2",
			username:    "testuser",
			password:    "testpass",
			expectedErr: false,
		},
		{
			name:        "empty URL",
			url:         "",
			username:    "testuser",
			password:    "testpass",
			expectedErr: true,
		},
		{
			name:        "empty username",
			url:         "https://dctrack.example.com/api/v2",
			username:    "",
			password:    "testpass",
			expectedErr: true,
		},
		{
			name:        "empty password",
			url:         "https://dctrack.example.com/api/v2",
			username:    "testuser",
			password:    "",
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := dctrack.New(tt.url, tt.username, tt.password)

			if tt.expectedErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if client == nil {
				t.Errorf("Expected client but got nil")
			}
		})
	}
}

func TestClientWithOptions(t *testing.T) {
	client, err := dctrack.New(
		"https://dctrack.example.com/api/v2",
		"testuser",
		"testpass",
		dctrack.WithPageSize(500),
		dctrack.WithMaxRetries(5),
		dctrack.WithRetryDelay(2*time.Second),
		dctrack.WithInsecureSSL(),
		dctrack.WithLimitedFields(),
	)

	if err != nil {
		t.Fatalf("Failed to create client with options: %v", err)
	}

	if client == nil {
		t.Fatalf("Expected client but got nil")
	}
}

func TestFilterBuilder(t *testing.T) {
	filters := dctrack.NewFilterBuilder().
		Location("RDU2").
		Status("Installed").
		Make("Dell").
		ItemClass("Device").
		SearchText("PowerEdge").
		WithPagination(2, 50).
		Build()

	expected := dctrack.ItemsParams{
		PageNumber: 2,
		PageSize:   50,
		Location:   "RDU2",
		Status:     "Installed",
		ItemClass:  "Device",
		Make:       "Dell",
		SearchText: "PowerEdge",
	}

	if filters.PageNumber != expected.PageNumber {
		t.Errorf("Expected PageNumber %d, got %d", expected.PageNumber, filters.PageNumber)
	}
	if filters.PageSize != expected.PageSize {
		t.Errorf("Expected PageSize %d, got %d", expected.PageSize, filters.PageSize)
	}
	if filters.Location != expected.Location {
		t.Errorf("Expected Location %s, got %s", expected.Location, filters.Location)
	}
	if filters.Status != expected.Status {
		t.Errorf("Expected Status %s, got %s", expected.Status, filters.Status)
	}
	if filters.ItemClass != expected.ItemClass {
		t.Errorf("Expected ItemClass %s, got %s", expected.ItemClass, filters.ItemClass)
	}
	if filters.Make != expected.Make {
		t.Errorf("Expected Make %s, got %s", expected.Make, filters.Make)
	}
	if filters.SearchText != expected.SearchText {
		t.Errorf("Expected SearchText %s, got %s", expected.SearchText, filters.SearchText)
	}
}

func TestPresetFilters(t *testing.T) {
	t.Run("InstalledOnly", func(t *testing.T) {
		params := dctrack.InstalledOnly()
		if params.Status != "Installed" {
			t.Errorf("Expected Status 'Installed', got %s", params.Status)
		}
	})

	t.Run("ByLocation", func(t *testing.T) {
		params := dctrack.ByLocation("RDU2")
		if params.Location != "RDU2" {
			t.Errorf("Expected Location 'RDU2', got %s", params.Location)
		}
		if params.Status != "Installed" {
			t.Errorf("Expected Status 'Installed', got %s", params.Status)
		}
	})

	t.Run("ByVendor", func(t *testing.T) {
		params := dctrack.ByVendor("Dell")
		if params.Make != "Dell" {
			t.Errorf("Expected Make 'Dell', got %s", params.Make)
		}
		if params.Status != "Installed" {
			t.Errorf("Expected Status 'Installed', got %s", params.Status)
		}
	})
}

// Mock server tests (integration test example)
func TestClientIntegration(t *testing.T) {
	// Create mock DCTrack server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/authentication/login":
			w.Header().Set("Authorization", "Bearer mock-token-12345")
			w.WriteHeader(http.StatusOK)
		case "/api/v2/quicksearch/items":
			w.Header().Set("Content-Type", "application/json")
			response := `{
				"records": [
					{
						"id": "test-001",
						"tiName": "Test Server 1",
						"cmbLocation": "TEST-DC",
						"cmbStatus": "Installed",
						"tiClass": "Device",
						"cmbMake": "Dell",
						"cmbModel": "PowerEdge R730",
						"tiItemOriginalPower": "500"
					}
				]
			}`
			w.Write([]byte(response))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create client pointing to mock server
	client, err := dctrack.New(
		server.URL+"/api/v2",
		"testuser",
		"testpass",
		dctrack.WithPageSize(1000),
		dctrack.WithInsecureSSL(),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	// Note: SetLogger method doesn't exist, logger is set during client creation

	// Test getting items
	items, err := client.GetItems(context.Background())
	if err != nil {
		t.Fatalf("Failed to get items from mock server: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(items))
	}

	item := items[0]
	if item.ID != "test-001" {
		t.Errorf("Expected ID 'test-001', got %s", item.ID)
	}
	if item.Name != "Test Server 1" {
		t.Errorf("Expected Name 'Test Server 1', got %s", item.Name)
	}
	if item.Make != "Dell" {
		t.Errorf("Expected Make 'Dell', got %s", item.Make)
	}
	if item.OriginalPower != 500 {
		t.Errorf("Expected OriginalPower 500, got %f", item.OriginalPower)
	}
}

func TestConfig(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		config := dctrack.DefaultConfig("https://test.com", "user", "pass")

		if config.PageSize != 1000 {
			t.Errorf("Expected PageSize 1000, got %d", config.PageSize)
		}
		if config.MaxRetries != 3 {
			t.Errorf("Expected MaxRetries 3, got %d", config.MaxRetries)
		}
		if config.RetryDelay != time.Second {
			t.Errorf("Expected RetryDelay 1s, got %v", config.RetryDelay)
		}
		if !config.VerifySSL {
			t.Errorf("Expected VerifySSL true, got false")
		}
	})

	t.Run("ConfigValidation", func(t *testing.T) {
		config := dctrack.Config{}
		err := config.Validate()
		if err == nil {
			t.Errorf("Expected validation error for empty config")
		}
	})
}
