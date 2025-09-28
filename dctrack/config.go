package dctrack

import (
	"fmt"
	"time"
)

// Config holds DCTrack API configuration
type Config struct {
	// Connection settings
	URL      string `json:"url"`      // DCTrack API base URL (e.g., "https://dctrack.company.com/api/v2")
	Username string `json:"username"` // DCTrack username
	Password string `json:"password"` // DCTrack password

	// API settings
	PageSize   int           `json:"page_size"`   // Default page size for API requests (default: 1000)
	MaxRetries int           `json:"max_retries"` // Maximum number of retry attempts (default: 3)
	RetryDelay time.Duration `json:"retry_delay"` // Delay between retry attempts (default: 1s)

	// Security settings
	VerifySSL        bool `json:"verify_ssl"`         // Whether to verify SSL certificates (default: true)
	RequestAllFields bool `json:"request_all_fields"` // Whether to request all available fields (default: true)
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig(url, username, password string) Config {
	return Config{
		URL:              url,
		Username:         username,
		Password:         password,
		PageSize:         1000,
		MaxRetries:       3,
		RetryDelay:       time.Second,
		VerifySSL:        true,
		RequestAllFields: true,
	}
}

// Option provides functional options for client configuration
type Option func(*Config)

// WithPageSize sets the default page size for API requests
func WithPageSize(size int) Option {
	return func(c *Config) {
		c.PageSize = size
	}
}

// WithMaxRetries sets the maximum number of retry attempts
func WithMaxRetries(retries int) Option {
	return func(c *Config) {
		c.MaxRetries = retries
	}
}

// WithRetryDelay sets the delay between retry attempts
func WithRetryDelay(delay time.Duration) Option {
	return func(c *Config) {
		c.RetryDelay = delay
	}
}

// WithInsecureSSL disables SSL certificate verification
func WithInsecureSSL() Option {
	return func(c *Config) {
		c.VerifySSL = false
	}
}

// WithLimitedFields requests only essential fields for better performance
func WithLimitedFields() Option {
	return func(c *Config) {
		c.RequestAllFields = false
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.URL == "" {
		return fmt.Errorf("URL is required")
	}
	if c.Username == "" {
		return fmt.Errorf("Username is required")
	}
	if c.Password == "" {
		return fmt.Errorf("Password is required")
	}
	if c.PageSize <= 0 {
		c.PageSize = 1000
	}
	if c.MaxRetries < 0 {
		c.MaxRetries = 3
	}
	if c.RetryDelay <= 0 {
		c.RetryDelay = time.Second
	}
	return nil
}
