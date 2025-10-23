package dctrack

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
	}{
		{
			name: "valid config",
			config: Config{
				URL:      "https://dctrack.example.com/api/v2",
				Username: "testuser",
				Password: "testpass",
			},
			expectError: false,
		},
		{
			name: "missing URL",
			config: Config{
				Username: "testuser",
				Password: "testpass",
			},
			expectError: true,
		},
		{
			name: "missing username",
			config: Config{
				URL:      "https://dctrack.example.com/api/v2",
				Password: "testpass",
			},
			expectError: true,
		},
		{
			name: "missing password",
			config: Config{
				URL:      "https://dctrack.example.com/api/v2",
				Username: "testuser",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestConfigDefaults(t *testing.T) {
	config := Config{
		URL:      "https://test.com/api/v2",
		Username: "user",
		Password: "pass",
		// Leave other fields as zero values
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Check that defaults were applied
	if client.config.PageSize != 1000 {
		t.Errorf("Expected default PageSize 1000, got %d", client.config.PageSize)
	}
	if client.config.MaxRetries != 3 {
		t.Errorf("Expected default MaxRetries 3, got %d", client.config.MaxRetries)
	}
	if client.config.RetryDelay != time.Second {
		t.Errorf("Expected default RetryDelay 1s, got %v", client.config.RetryDelay)
	}
}

func TestConfigFromFile(t *testing.T) {
	// Create temporary config file
	tempDir, err := os.MkdirTemp("", "dctrack-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configContent := `
url: "https://dctrack.test.com/api/v2"
username: "testuser"
password: "testpass"
page_size: 500
max_retries: 5
retry_delay: "2s"
verify_ssl: false
request_all_fields: true
`

	configFile := filepath.Join(tempDir, "config.yaml")
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Test would require YAML parsing capability
	// This is a placeholder for configuration file loading
	t.Log("Config file created for testing:", configFile)
}

func TestEnvironmentVariableConfig(t *testing.T) {
	// Save original environment
	originalURL := os.Getenv("DCTRACK_URL")
	originalUsername := os.Getenv("DCTRACK_USERNAME")
	originalPassword := os.Getenv("DCTRACK_PASSWORD")

	defer func() {
		// Restore original environment
		os.Setenv("DCTRACK_URL", originalURL)
		os.Setenv("DCTRACK_USERNAME", originalUsername)
		os.Setenv("DCTRACK_PASSWORD", originalPassword)
	}()

	// Set test environment variables
	os.Setenv("DCTRACK_URL", "https://test.dctrack.com/api/v2")
	os.Setenv("DCTRACK_USERNAME", "testuser")
	os.Setenv("DCTRACK_PASSWORD", "testpass")

	// This would test environment variable loading
	// Implementation depends on adding YAML/env config loading
	t.Log("Environment variables set for testing")
}

func TestPasswordFileHandling(t *testing.T) {
	// Create temporary password file
	tempDir, err := os.MkdirTemp("", "dctrack-password-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	passwordFile := filepath.Join(tempDir, "password.txt")
	password := "secret-password-123"

	err = os.WriteFile(passwordFile, []byte(password), 0600)
	if err != nil {
		t.Fatalf("Failed to write password file: %v", err)
	}

	// Test password file reading (would require implementation)
	content, err := os.ReadFile(passwordFile)
	if err != nil {
		t.Fatalf("Failed to read password file: %v", err)
	}

	if string(content) != password {
		t.Errorf("Expected password %s, got %s", password, string(content))
	}
}
