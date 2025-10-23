package dctrack

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"
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
			client, err := New(tt.url, tt.username, tt.password)

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

func TestFilterBuilder(t *testing.T) {
	filters := NewFilterBuilder().
		Location("RDU2").
		Status("Installed").
		Make("Dell").
		SearchText("PowerEdge").
		WithPagination(2, 50).
		Build()

	expected := ItemsParams{
		PageNumber: 2,
		PageSize:   50,
		Location:   "RDU2",
		Status:     "Installed",
		Make:       "Dell",
		SearchText: "PowerEdge",
	}

	if filters.PageNumber != expected.PageNumber {
		t.Errorf("Expected PageNumber %d, got %d", expected.PageNumber, filters.PageNumber)
	}
	if filters.Location != expected.Location {
		t.Errorf("Expected Location %s, got %s", expected.Location, filters.Location)
	}
	if filters.Status != expected.Status {
		t.Errorf("Expected Status %s, got %s", expected.Status, filters.Status)
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
		params := InstalledOnly()
		if params.Status != "Installed" {
			t.Errorf("Expected Status 'Installed', got %s", params.Status)
		}
	})

	t.Run("ByLocation", func(t *testing.T) {
		params := ByLocation("RDU2")
		if params.Location != "RDU2" {
			t.Errorf("Expected Location 'RDU2', got %s", params.Location)
		}
		if params.Status != "Installed" {
			t.Errorf("Expected Status 'Installed', got %s", params.Status)
		}
	})

	t.Run("ByVendor", func(t *testing.T) {
		params := ByVendor("Dell")
		if params.Make != "Dell" {
			t.Errorf("Expected Make 'Dell', got %s", params.Make)
		}
		if params.Status != "Installed" {
			t.Errorf("Expected Status 'Installed', got %s", params.Status)
		}
	})
}

func TestConfig(t *testing.T) {
	t.Run("ValidConfig", func(t *testing.T) {
		config := Config{
			URL:      "https://test.com/api/v2",
			Username: "user",
			Password: "pass",
		}

		client, err := NewClient(config)
		if err != nil {
			t.Errorf("Unexpected error with valid config: %v", err)
		}
		if client == nil {
			t.Errorf("Expected client but got nil")
		}
	})

	t.Run("InvalidConfig", func(t *testing.T) {
		config := Config{} // Empty config

		client, err := NewClient(config)
		if err == nil {
			t.Errorf("Expected error with invalid config")
		}
		if client != nil {
			t.Errorf("Expected nil client with invalid config")
		}
	})
}

// Mock server integration test
func TestClientIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create mock DCTrack server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/authentication/login":
			// Provide token both as a header and in the JSON body to support either client behavior
			token := "mock-token-12345"
			w.Header().Set("Authorization", "Bearer "+token)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"token":"` + token + `", "access_token":"` + token + `"}`))
		case "/api/v2/quicksearch/items":
			// Verify request method
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}

			// Verify auth header more permissively (accept Bearer <token> or similar)
			auth := r.Header.Get("Authorization")
			if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// Verify content type more permissively
			if ct := r.Header.Get("Content-Type"); !strings.Contains(ct, "application/json") {
				w.WriteHeader(http.StatusUnsupportedMediaType)
				return
			}

			// Parse and verify payload
			var payload map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			response := `{
				"records": [
					{
						"id": "test-001",
						"name": "Test Server 1",
						"location": "TEST-DC",
						"status": "Installed",
						"item_class": "Device",
						"make": "Dell",
						"model": "PowerEdge R730",
						"original_power": "500",
						"last_updated_at": "2024-01-01 12:00:00-05"
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
	config := Config{
		URL:       server.URL + "/api/v2",
		Username:  "testuser",
		Password:  "testpass",
		PageSize:  1000,
		VerifySSL: false,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	client.SetLogger(zap.NewNop())

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

func TestSearchItems(t *testing.T) {
	// Test search validation
	client := &Client{
		config: Config{
			URL:      "https://test.com/api/v2",
			Username: "test",
			Password: "test",
		},
	}

	// Test empty search query
	_, err := client.SearchItems(context.Background(), "")
	if err == nil {
		t.Errorf("Expected error for empty search query")
	}
}

func TestClientClose(t *testing.T) {
	client, err := New("https://test.com/api/v2", "user", "pass")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	err = client.Close()
	if err != nil {
		t.Errorf("Close() should not return error: %v", err)
	}
}

func BenchmarkFilterBuilder(b *testing.B) {
	b.Run("filter_builder", func(b *testing.B) {
		_ = NewFilterBuilder().
			Location("RDU2").
			Status("Installed").
			Make("Dell").
			Model("PowerEdge").
			SearchText("R730").
			WithPagination(1, 100).
			Build()
	})
}

func BenchmarkConfigValidation(b *testing.B) {
	config := Config{
		URL:      "https://test.com/api/v2",
		Username: "user",
		Password: "pass",
	}

	b.Run("validate_config", func(b *testing.B) {
		_ = validateConfig(config)
	})
}
