package dctrack

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
			name:        "valid parameters",
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

			// Test that config is properly set
			if client.config.URL != tt.url {
				t.Errorf("Expected URL %s, got %s", tt.url, client.config.URL)
			}
			if client.config.Username != tt.username {
				t.Errorf("Expected Username %s, got %s", tt.username, client.config.Username)
			}
			if client.config.Password != tt.password {
				t.Errorf("Expected Password %s, got %s", tt.password, client.config.Password)
			}
		})
	}
}

func TestNewWithOptions(t *testing.T) {
	client, err := New(
		"https://test.com/api/v2",
		"user",
		"pass",
		WithPageSize(500),
		WithMaxRetries(5),
		WithRetryDelay(2*time.Second),
		WithInsecureSSL(),
		WithLimitedFields(),
	)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if client.config.PageSize != 500 {
		t.Errorf("Expected PageSize 500, got %d", client.config.PageSize)
	}
	if client.config.MaxRetries != 5 {
		t.Errorf("Expected MaxRetries 5, got %d", client.config.MaxRetries)
	}
	if client.config.RetryDelay != 2*time.Second {
		t.Errorf("Expected RetryDelay 2s, got %v", client.config.RetryDelay)
	}
	if client.config.VerifySSL != false {
		t.Errorf("Expected VerifySSL false, got %t", client.config.VerifySSL)
	}
	if client.config.RequestAllFields != false {
		t.Errorf("Expected RequestAllFields false (limited fields), got %t", client.config.RequestAllFields)
	}
}

func TestClientLogin(t *testing.T) {
	// Create mock DCTrack server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/authentication/login" {
			// Check basic auth
			username, password, ok := r.BasicAuth()
			if !ok || username != "testuser" || password != "testpass" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			w.Header().Set("Authorization", "Bearer test-token-12345")
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client, err := New(server.URL+"/api/v2", "testuser", "testpass")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	// Note: SetLogger method doesn't exist in this package, logger is set during NewClient

	// Test login
	ctx := context.Background()
	err = client.login(ctx)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	// Verify token was set
	if client.token != "test-token-12345" {
		t.Errorf("Expected token 'test-token-12345', got '%s'", client.token)
	}
}

func TestClientLoginFailure(t *testing.T) {
	// Create mock server that returns 401
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client, err := New(server.URL+"/api/v2", "baduser", "badpass")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	// Note: SetLogger method doesn't exist in this package, logger is set during NewClient

	// Test login failure
	ctx := context.Background()
	err = client.login(ctx)
	if err == nil {
		t.Errorf("Expected login to fail but it succeeded")
	}
}

func TestNewClientWithLogger(t *testing.T) {
	// Test creating client with custom logger
	logger := zap.NewNop()
	config := DefaultConfig("https://test.com/api/v2", "user", "pass")
	client := NewClient(config, logger)

	if client == nil {
		t.Errorf("Expected client but got nil")
	}

	// Test creating client with nil logger (should use default)
	client2 := NewClient(config, nil)
	if client2 == nil {
		t.Errorf("Expected client but got nil with nil logger")
	}
}

func TestClientCreation(t *testing.T) {
	client, err := New("https://test.com/api/v2", "user", "pass")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	if client == nil {
		t.Errorf("Expected client but got nil")
	}
}

func TestBuildStandardFieldsPayload(t *testing.T) {
	client, err := New("https://test.com/api/v2", "user", "pass")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	t.Run("standard fields payload", func(t *testing.T) {
		payload := client.buildStandardFieldsPayload()

		selectedColumns, ok := payload["selectedColumns"]
		if !ok {
			t.Errorf("Expected selectedColumns in payload")
		}

		columns, ok := selectedColumns.([]map[string]string)
		if !ok {
			t.Errorf("Expected selectedColumns to be []map[string]string")
		}

		// Should have essential fields
		if len(columns) == 0 {
			t.Errorf("Expected at least some essential fields")
		}
	})
}

func TestMapDCTrackRecord(t *testing.T) {
	client, err := New("https://test.com/api/v2", "user", "pass")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test record mapping
	record := map[string]interface{}{
		"id":                            "test-001",
		"tiName":                        "Test Server",
		"cmbLocation":                   "TEST-DC",
		"cmbStatus":                     "Installed",
		"tiClass":                       "Device",
		"cmbMake":                       "Dell",
		"cmbModel":                      "PowerEdge R730",
		"tiItemOriginalPower":           "500.5",
		"tiSerialNumber":                "SN12345",
		"lastUpdatedOn":                 "2024-01-01 12:00:00-05",
		"cmbSystemAdminTeam":            "Infrastructure",
		"tiCustomField_Primary Contact": "admin@test.com",
	}

	item, err := client.mapDCTrackRecord(record)
	if err != nil {
		t.Fatalf("Failed to map record: %v", err)
	}

	// Verify mapping
	if item.ID != "test-001" {
		t.Errorf("Expected ID 'test-001', got '%s'", item.ID)
	}
	if item.Name != "Test Server" {
		t.Errorf("Expected Name 'Test Server', got '%s'", item.Name)
	}
	if item.Location != "TEST-DC" {
		t.Errorf("Expected Location 'TEST-DC', got '%s'", item.Location)
	}
	if item.Status != "Installed" {
		t.Errorf("Expected Status 'Installed', got '%s'", item.Status)
	}
	if item.ItemClass != "Device" {
		t.Errorf("Expected ItemClass 'Device', got '%s'", item.ItemClass)
	}
	if item.Make != "Dell" {
		t.Errorf("Expected Make 'Dell', got '%s'", item.Make)
	}
	if item.Model != "PowerEdge R730" {
		t.Errorf("Expected Model 'PowerEdge R730', got '%s'", item.Model)
	}
	if item.OriginalPower != 500.5 {
		t.Errorf("Expected OriginalPower 500.5, got %f", item.OriginalPower)
	}
	if item.SerialNumber != "SN12345" {
		t.Errorf("Expected SerialNumber 'SN12345', got '%s'", item.SerialNumber)
	}
	if item.SystemAdminTeam != "Infrastructure" {
		t.Errorf("Expected SystemAdminTeam 'Infrastructure', got '%s'", item.SystemAdminTeam)
	}
	if item.PrimaryContact != "admin@test.com" {
		t.Errorf("Expected PrimaryContact 'admin@test.com', got '%s'", item.PrimaryContact)
	}
}

func TestMapDCTrackRecordWithNilValues(t *testing.T) {
	client, err := New("https://test.com/api/v2", "user", "pass")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test record with nil/missing values
	record := map[string]interface{}{
		"id":                  "test-002",
		"tiName":              nil,
		"cmbLocation":         "",
		"tiItemOriginalPower": nil,
		"invalidNumber":       "not-a-number",
	}

	item, err := client.mapDCTrackRecord(record)
	if err != nil {
		t.Fatalf("Failed to map record with nil values: %v", err)
	}

	// Verify nil handling
	if item.ID != "test-002" {
		t.Errorf("Expected ID 'test-002', got '%s'", item.ID)
	}
	if item.Name != "" {
		t.Errorf("Expected empty Name for nil value, got '%s'", item.Name)
	}
	if item.Location != "" {
		t.Errorf("Expected empty Location for empty string, got '%s'", item.Location)
	}
	if item.OriginalPower != 0 {
		t.Errorf("Expected OriginalPower 0 for nil value, got %f", item.OriginalPower)
	}
}

// Benchmark tests
func BenchmarkMapDCTrackRecord(b *testing.B) {
	client, err := New("https://test.com/api/v2", "user", "pass")
	if err != nil {
		b.Fatalf("Failed to create client: %v", err)
	}

	record := map[string]interface{}{
		"id":                  "test-001",
		"tiName":              "Test Server",
		"cmbLocation":         "TEST-DC",
		"cmbStatus":           "Installed",
		"tiClass":             "Device",
		"cmbMake":             "Dell",
		"cmbModel":            "PowerEdge R730",
		"tiItemOriginalPower": "500.5",
		"tiSerialNumber":      "SN12345",
		"lastUpdatedOn":       "2024-01-01 12:00:00-05",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.mapDCTrackRecord(record)
		if err != nil {
			b.Fatalf("Mapping failed: %v", err)
		}
	}
}

func BenchmarkBuildStandardFieldsPayload(b *testing.B) {
	client, err := New("https://test.com/api/v2", "user", "pass")
	if err != nil {
		b.Fatalf("Failed to create client: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.buildStandardFieldsPayload()
	}
}
