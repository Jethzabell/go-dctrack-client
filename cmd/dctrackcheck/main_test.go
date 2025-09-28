package main

import (
	"os"
	"testing"
	"time"

	dctrack "github.com/Jethzabell/go-dctrack-client"
)

// TestMain provides setup and teardown for the test suite
func TestMain(m *testing.M) {
	// Setup: Save original environment variables
	originalEnv := saveEnvironment()

	// Run tests
	code := m.Run()

	// Teardown: Restore original environment
	restoreEnvironment(originalEnv)

	os.Exit(code)
}

func TestGetConfigFromEnv(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected dctrack.Config
	}{
		{
			name: "default values",
			envVars: map[string]string{
				"DCTRACK_URL":      "https://test.com/api/v2",
				"DCTRACK_USERNAME": "testuser",
				"DCTRACK_PASSWORD": "testpass",
			},
			expected: dctrack.Config{
				URL:              "https://test.com/api/v2",
				Username:         "testuser",
				Password:         "testpass",
				PageSize:         defaultPageSize,
				MaxRetries:       3,
				RetryDelay:       time.Second,
				VerifySSL:        true,
				RequestAllFields: false,
			},
		},
		{
			name: "custom values",
			envVars: map[string]string{
				"DCTRACK_URL":         "https://custom.com/api/v2",
				"DCTRACK_USERNAME":    "customuser",
				"DCTRACK_PASSWORD":    "custompass",
				"DCTRACK_PAGE_SIZE":   "200",
				"DCTRACK_MAX_RETRIES": "5",
				"DCTRACK_RETRY_DELAY": "2",
				"DCTRACK_VERIFY_SSL":  "false",
				"DCTRACK_ALL_FIELDS":  "true",
			},
			expected: dctrack.Config{
				URL:              "https://custom.com/api/v2",
				Username:         "customuser",
				Password:         "custompass",
				PageSize:         200,
				MaxRetries:       5,
				RetryDelay:       2 * time.Second,
				VerifySSL:        false,
				RequestAllFields: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			config := getConfigFromEnv()

			// Validate configuration
			if config.URL != tt.expected.URL {
				t.Errorf("Expected URL %s, got %s", tt.expected.URL, config.URL)
			}
			if config.Username != tt.expected.Username {
				t.Errorf("Expected Username %s, got %s", tt.expected.Username, config.Username)
			}
			if config.Password != tt.expected.Password {
				t.Errorf("Expected Password %s, got %s", tt.expected.Password, config.Password)
			}
			if config.PageSize != tt.expected.PageSize {
				t.Errorf("Expected PageSize %d, got %d", tt.expected.PageSize, config.PageSize)
			}
			if config.MaxRetries != tt.expected.MaxRetries {
				t.Errorf("Expected MaxRetries %d, got %d", tt.expected.MaxRetries, config.MaxRetries)
			}
			if config.RetryDelay != tt.expected.RetryDelay {
				t.Errorf("Expected RetryDelay %v, got %v", tt.expected.RetryDelay, config.RetryDelay)
			}
			if config.VerifySSL != tt.expected.VerifySSL {
				t.Errorf("Expected VerifySSL %t, got %t", tt.expected.VerifySSL, config.VerifySSL)
			}
			if config.RequestAllFields != tt.expected.RequestAllFields {
				t.Errorf("Expected RequestAllFields %t, got %t", tt.expected.RequestAllFields, config.RequestAllFields)
			}

			// Clean up environment variables
			for key := range tt.envVars {
				os.Unsetenv(key)
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      dctrack.Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: dctrack.Config{
				URL:      "https://test.com/api/v2",
				Username: "testuser",
				Password: "testpass",
			},
			expectError: false,
		},
		{
			name: "missing URL",
			config: dctrack.Config{
				Username: "testuser",
				Password: "testpass",
			},
			expectError: true,
			errorMsg:    "DCTRACK_URL environment variable is required",
		},
		{
			name: "missing username",
			config: dctrack.Config{
				URL:      "https://test.com/api/v2",
				Password: "testpass",
			},
			expectError: true,
			errorMsg:    "DCTRACK_USERNAME environment variable is required",
		},
		{
			name: "missing password",
			config: dctrack.Config{
				URL:      "https://test.com/api/v2",
				Username: "testuser",
			},
			expectError: true,
			errorMsg:    "DCTRACK_PASSWORD environment variable is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestHelperFunctions(t *testing.T) {
	t.Run("getEnvOrDefault", func(t *testing.T) {
		// Test with existing environment variable
		os.Setenv("TEST_VAR", "test_value")
		result := getEnvOrDefault("TEST_VAR", "default")
		if result != "test_value" {
			t.Errorf("Expected 'test_value', got '%s'", result)
		}
		os.Unsetenv("TEST_VAR")

		// Test with non-existing environment variable
		result = getEnvOrDefault("NON_EXISTENT", "default")
		if result != "default" {
			t.Errorf("Expected 'default', got '%s'", result)
		}
	})

	t.Run("getEnvIntOrDefault", func(t *testing.T) {
		// Test with valid integer
		os.Setenv("TEST_INT", "42")
		result := getEnvIntOrDefault("TEST_INT", 10)
		if result != 42 {
			t.Errorf("Expected 42, got %d", result)
		}
		os.Unsetenv("TEST_INT")

		// Test with invalid integer
		os.Setenv("TEST_INT", "invalid")
		result = getEnvIntOrDefault("TEST_INT", 10)
		if result != 10 {
			t.Errorf("Expected default 10, got %d", result)
		}
		os.Unsetenv("TEST_INT")

		// Test with non-existing variable
		result = getEnvIntOrDefault("NON_EXISTENT", 20)
		if result != 20 {
			t.Errorf("Expected default 20, got %d", result)
		}
	})

	t.Run("getEnvBoolOrDefault", func(t *testing.T) {
		// Test with valid boolean
		os.Setenv("TEST_BOOL", "true")
		result := getEnvBoolOrDefault("TEST_BOOL", false)
		if result != true {
			t.Errorf("Expected true, got %t", result)
		}
		os.Unsetenv("TEST_BOOL")

		// Test with invalid boolean
		os.Setenv("TEST_BOOL", "invalid")
		result = getEnvBoolOrDefault("TEST_BOOL", false)
		if result != false {
			t.Errorf("Expected default false, got %t", result)
		}
		os.Unsetenv("TEST_BOOL")

		// Test with non-existing variable
		result = getEnvBoolOrDefault("NON_EXISTENT", true)
		if result != true {
			t.Errorf("Expected default true, got %t", result)
		}
	})
}

func TestPrintFunctions(t *testing.T) {
	// Test item creation for print functions
	testItem := dctrack.DCTrackItem{
		ID:              "test-001",
		Name:            "Test Server",
		Status:          "Installed",
		Type:            "Server",
		ItemClass:       "Device",
		Location:        "TEST-DC",
		Cabinet:         "CAB-01",
		Position:        "U10",
		Height:          2,
		Make:            "Dell",
		Model:           "PowerEdge R730",
		SerialNumber:    "SN12345",
		OriginalPower:   500.0,
		TiAssetTag:      "AT-001",
		PrimaryContact:  "admin@test.com",
		SystemAdminTeam: "Infrastructure",
	}

	// Test printItemSummary (no return value, just ensure no panic)
	t.Run("printItemSummary", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("printItemSummary panicked: %v", r)
			}
		}()
		printItemSummary(testItem)
	})

	// Test printItemDetails (no return value, just ensure no panic)
	t.Run("printItemDetails", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("printItemDetails panicked: %v", r)
			}
		}()
		printItemDetails(testItem)
	})

	// Test printUsage (no return value, just ensure no panic)
	t.Run("printUsage", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("printUsage panicked: %v", r)
			}
		}()
		printUsage()
	})
}

// Helper functions for TestMain
func saveEnvironment() map[string]string {
	envVars := []string{
		"DCTRACK_URL", "DCTRACK_USERNAME", "DCTRACK_PASSWORD",
		"DCTRACK_PAGE_SIZE", "DCTRACK_MAX_RETRIES", "DCTRACK_RETRY_DELAY",
		"DCTRACK_VERIFY_SSL", "DCTRACK_ALL_FIELDS",
	}

	saved := make(map[string]string)
	for _, key := range envVars {
		saved[key] = os.Getenv(key)
	}
	return saved
}

func restoreEnvironment(saved map[string]string) {
	for key, value := range saved {
		if value == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, value)
		}
	}
}
