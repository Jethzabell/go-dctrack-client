package dctrack

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Client provides DCTrack API access
type Client struct {
	config     Config
	httpClient *http.Client
	logger     *zap.Logger
	token      string // Store the JWT token after login
}

// NewClient creates a new DCTrack API client
func NewClient(cfg Config, logger *zap.Logger) *Client {
	// Use default logger if none provided
	if logger == nil {
		logger = zap.NewNop()
	}

	// Create transport with TLS config
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !cfg.VerifySSL,
		},
	}

	// Create HTTP client with transport and timeout
	httpClient := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}

	return &Client{
		config:     cfg,
		httpClient: httpClient,
		logger:     logger,
	}
}

// New creates a new DCTrack client with functional options
func New(url, username, password string, opts ...Option) (*Client, error) {
	config := DefaultConfig(url, username, password)

	// Apply options
	for _, opt := range opts {
		opt(&config)
	}

	// Validate config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return NewClient(config, zap.NewNop()), nil
}

// login authenticates with DCTrack API and stores the JWT token
func (c *Client) login(ctx context.Context) error {
	loginURL := fmt.Sprintf("%s/authentication/login", c.config.URL)

	req, err := http.NewRequestWithContext(ctx, "POST", loginURL, nil)
	if err != nil {
		return fmt.Errorf("error creating login request: %w", err)
	}

	// Use Basic Auth for login
	req.SetBasicAuth(c.config.Username, c.config.Password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	c.logger.Debug("Making DCTrack login request",
		zap.String("method", "POST"),
		zap.String("url", loginURL),
		zap.String("username", c.config.Username))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error making login request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed with status %d", resp.StatusCode)
	}

	// Extract Bearer token from Authorization header
	authHeader := resp.Header.Get("Authorization")
	if authHeader == "" {
		return fmt.Errorf("no Authorization header in login response")
	}

	// Extract token from "Bearer <token>" format
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		c.token = authHeader[7:]
		c.logger.Debug("Successfully obtained DCTrack token",
			zap.Int("token_length", len(c.token)))
		return nil
	}

	return fmt.Errorf("invalid Authorization header format")
}

// GetItems retrieves all items from DCTrack API
func (c *Client) GetItems(ctx context.Context) ([]DCTrackItem, error) {
	return c.GetItemsWithParams(ctx, ItemsParams{
		PageSize: c.config.PageSize,
	})
}

// GetItemsWithParams retrieves items with specific parameters
func (c *Client) GetItemsWithParams(ctx context.Context, params ItemsParams) ([]DCTrackItem, error) {
	// Login first to get the JWT token
	if err := c.login(ctx); err != nil {
		return nil, fmt.Errorf("error logging in to DCTrack: %w", err)
	}

	var allItems []DCTrackItem
	page := 1
	if params.PageNumber > 0 {
		page = params.PageNumber
	}

	pageSize := c.config.PageSize
	if params.PageSize > 0 {
		pageSize = params.PageSize
	}

	for {
		url := fmt.Sprintf("%s/quicksearch/items?pageNumber=%d&pageSize=%d",
			c.config.URL, page, pageSize)

		// Add optional query parameters
		if params.Location != "" {
			url += "&location=" + params.Location
		}
		if params.Status != "" {
			url += "&status=" + params.Status
		}
		if params.SearchText != "" {
			url += "&searchText=" + params.SearchText
		}

		payload := c.buildStandardFieldsPayload()

		items, err := c.fetchPage(ctx, url, payload)
		if err != nil {
			return nil, err
		}

		if len(items) == 0 {
			break
		}

		allItems = append(allItems, items...)

		c.logger.Info("Fetched DCTrack page",
			zap.Int("page", page),
			zap.Int("items_count", len(items)),
			zap.Int("total_items", len(allItems)))

		// If we got fewer items than requested, we're done
		if len(items) < pageSize {
			break
		}

		page++

		// If single page requested, stop after first page
		if params.PageNumber > 0 {
			break
		}
	}

	c.logger.Info("Fetched items from DCTrack", zap.Int("count", len(allItems)))
	return allItems, nil
}

// GetItemByID retrieves a specific item by ID
func (c *Client) GetItemByID(ctx context.Context, id string) (*DCTrackItem, error) {
	items, err := c.GetItemsWithParams(ctx, ItemsParams{
		SearchText: id,
		PageSize:   1,
	})
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		if item.ID == id {
			return &item, nil
		}
	}

	return nil, fmt.Errorf("item with ID %s not found", id)
}

// SearchItems performs a text search for items
func (c *Client) SearchItems(ctx context.Context, query string) ([]DCTrackItem, error) {
	return c.GetItemsWithParams(ctx, ItemsParams{
		SearchText: query,
		PageSize:   c.config.PageSize,
	})
}

// fetchPage fetches a single page of data from DCTrack API
func (c *Client) fetchPage(ctx context.Context, url string, payload interface{}) ([]DCTrackItem, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error marshaling payload: %w", err)
	}

	// Retry logic
	var lastErr error
	for attempt := 1; attempt <= c.config.MaxRetries; attempt++ {
		c.logger.Debug("Making DCTrack request",
			zap.String("method", "POST"),
			zap.String("url", url),
			zap.String("username", c.config.Username),
			zap.Int("attempt", attempt))

		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payloadBytes))
		if err != nil {
			return nil, fmt.Errorf("error creating request: %w", err)
		}

		// Add headers
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.token)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			if attempt < c.config.MaxRetries {
				c.logger.Warn("Request failed, retrying",
					zap.Error(err),
					zap.Int("attempt", attempt),
					zap.Duration("retry_delay", c.config.RetryDelay))
				time.Sleep(c.config.RetryDelay)
				continue
			}
			break
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("request failed with status %d", resp.StatusCode)
			if attempt < c.config.MaxRetries {
				time.Sleep(c.config.RetryDelay)
				continue
			}
			break
		}

		// Parse response
		var dctrackResp DCTrackResponse
		if err := json.NewDecoder(resp.Body).Decode(&dctrackResp); err != nil {
			return nil, fmt.Errorf("error decoding response: %w", err)
		}

		var items []DCTrackItem
		for _, record := range dctrackResp.Records {
			item, err := c.mapDCTrackRecord(record)
			if err != nil {
				c.logger.Warn("Error mapping DCTrack record", zap.Error(err))
				continue
			}
			items = append(items, item)
		}

		return items, nil
	}

	return nil, fmt.Errorf("all retry attempts failed, last error: %w", lastErr)
}

// buildStandardFieldsPayload creates the standard field selection payload
func (c *Client) buildStandardFieldsPayload() map[string]interface{} {
	if !c.config.RequestAllFields {
		// Minimal field set for better performance
		return map[string]interface{}{
			"selectedColumns": []map[string]string{
				{"name": "id"},
				{"name": "tiName"},
				{"name": "cmbLocation"},
				{"name": "cmbStatus"},
				{"name": "tiClass"},
				{"name": "cmbMake"},
				{"name": "cmbModel"},
				{"name": "tiItemOriginalPower"},
				{"name": "lastUpdatedOn"},
			},
		}
	}

	// Full field set (comprehensive data)
	return map[string]interface{}{
		"selectedColumns": []map[string]string{
			// Core fields
			{"name": "id"},
			{"name": "cmbLocation"},
			{"name": "tiClass"},
			{"name": "cmbStatus"},
			{"name": "tiName"},
			{"name": "cmbMake"},
			{"name": "cmbModel"},
			{"name": "cmbCabinet"},
			{"name": "tiSerialNumber"},
			{"name": "lastUpdatedOn"},
			{"name": "tiItemOriginalPower"},
			{"name": "cmbSystemAdminTeam"},
			{"name": "tiCustomField_Primary Contact"},

			// Asset information
			{"name": "tiSubclass"},
			{"name": "tiAssetTag"},
			{"name": "tiFormFactor"},
			{"name": "tiMounting"},
			{"name": "tiWidth"},
			{"name": "tiDepth"},
			{"name": "tiWeight"},
			{"name": "tiRUs"},

			// Power data
			{"name": "tiPotentialPower"},
			{"name": "tiEffectivePower"},
			{"name": "tiPowerCapacity"},
			{"name": "tiPSRedundancy"},
			{"name": "tiPurchasePrice"},
			{"name": "tiContractAmount"},

			// Infrastructure
			{"name": "cmbPlantBay"},
			{"name": "cmbCabinetId"},
			{"name": "cmbRowPosition"},
			{"name": "cmbRowLabel"},
			{"name": "tiFloorNodeCode"},
			{"name": "tiFloorName"},
			{"name": "tiRoomNodeCode"},
			{"name": "tiRoomName"},

			// Integration status
			{"name": "tiIntegrationStatus"},
			{"name": "tiVMwareIntegrationStatus"},
			{"name": "tiCmdbIntegrationStatus"},
			{"name": "tiItemBudgetStatus"},
			{"name": "chkItemAutoPowerBudget"},
			{"name": "chkDerateAmps"},

			// Network config
			{"name": "ipAddresses"},
			{"name": "ipAddressPortName"},
			{"name": "tifreeDataPortCount"},
			{"name": "tifreePowerPortCount"},
			{"name": "tiPXUsername"},
			{"name": "tiSnmpWriteCommString"},

			// Technical specs
			{"name": "tiUsers"},
			{"name": "tiRAM"},
			{"name": "tiProcesses"},
			{"name": "tiCpuQuantity"},
			{"name": "tiCpuType"},
			{"name": "tiPartNumber"},

			// Dates
			{"name": "installationDate"},
			{"name": "contractEndDate"},
			{"name": "purchaseDate"},
			{"name": "tiPlannedDecommDate"},

			// Custom fields
			{"name": "tiCustomField_Contact Team Name"},
			{"name": "tiCustomField_Audit Remarks"},
			{"name": "tiCustomField_Audit Date"},
			{"name": "tiCustomField_Audit By"},
			{"name": "tiCustomField_Warranty Expiration Date"},
			{"name": "tiCustomField_Asset Status"},
			{"name": "tiCustomField_PNT/IT"},

			// Admin data
			{"name": "cmbSystemAdminTeam"},
			{"name": "cmbSystemAdmin"},
			{"name": "cmbCustomer"},
			{"name": "tiPONumber"},
			{"name": "tiNotes"},
		},
	}
}
