package dctrack

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Version information
const (
	Version   = "v1.0.0"
	UserAgent = "go-dctrack-client/v1.0.0"
)

// Config holds DCTrack API configuration
type Config struct {
	// Connection settings
	URL      string `yaml:"url"`      // DCTrack API base URL (e.g., "https://dctrack.company.com/api/v2")
	Username string `yaml:"username"` // DCTrack username
	Password string `yaml:"password"` // DCTrack password

	// API settings
	PageSize   int           `yaml:"page_size"`   // Default page size for API requests (default: 1000)
	MaxRetries int           `yaml:"max_retries"` // Maximum number of retry attempts (default: 3)
	RetryDelay time.Duration `yaml:"retry_delay"` // Delay between retry attempts (default: 1s)

	// Security settings
	VerifySSL        bool `yaml:"verify_ssl"`         // Whether to verify SSL certificates (default: true)
	RequestAllFields bool `yaml:"request_all_fields"` // Whether to request all available fields (default: true)
}

// DCTrackItem represents a hardware item from DCTrack API
type DCTrackItem struct {
	// Core identification
	ID        string `json:"id"`
	Name      string `json:"name"`
	ItemClass string `json:"item_class"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	Location  string `json:"location"`
	Cabinet   string `json:"cabinet"`
	Rack      string `json:"rack"`
	Position  string `json:"position"`
	Height    int    `json:"height"`

	// Basic asset information
	Make             string     `json:"make"`
	Model            string     `json:"model"`
	SerialNumber     string     `json:"serial_number"`
	OriginalPower    float64    `json:"original_power"`
	PowerConsumption float64    `json:"power_consumption"`
	SystemAdminTeam  string     `json:"system_admin_team"`
	PrimaryContact   string     `json:"primary_contact"`
	InstallDate      *time.Time `json:"install_date"`
	LastServiceDate  *time.Time `json:"last_service_date"`

	// DCTrack-specific fields (ti_ prefix)
	TiSubclass       string  `json:"ti_subclass"`
	TiSerialNumber   string  `json:"ti_serial_number"`
	TiAssetTag       string  `json:"ti_asset_tag"`
	TiFormFactor     string  `json:"ti_form_factor"`
	TiMounting       string  `json:"ti_mounting"`
	TiWidth          float64 `json:"ti_width"`
	TiDepth          float64 `json:"ti_depth"`
	TiWeight         float64 `json:"ti_weight"`
	TiRuHeight       string  `json:"ti_ru_height"`
	TiPotentialPower float64 `json:"ti_potential_power"`
	TiEffectivePower float64 `json:"ti_effective_power"`
	TiPowerCapacity  float64 `json:"ti_power_capacity"`
	TiPsRedundancy   string  `json:"ti_ps_redundancy"`
	TiPurchasePrice  float64 `json:"ti_purchase_price"`
	TiContractAmount float64 `json:"ti_contract_amount"`

	// Infrastructure Details
	CmbPlantBay     string `json:"cmb_plant_bay"`
	CmbCabinetID    string `json:"cmb_cabinet_id"`
	CmbRowPosition  string `json:"cmb_row_position"`
	CmbRowLabel     string `json:"cmb_row_label"`
	TiFloorNodeCode string `json:"ti_floor_node_code"`
	TiFloorName     string `json:"ti_floor_name"`
	TiRoomNodeCode  string `json:"ti_room_node_code"`
	TiRoomName      string `json:"ti_room_name"`

	// Integration Status
	TiIntegrationStatus       string `json:"ti_integration_status"`
	TiVmwareIntegrationStatus string `json:"ti_vmware_integration_status"`
	TiCmdbIntegrationStatus   string `json:"ti_cmdb_integration_status"`

	// Budget and Planning
	TiItemBudgetStatus     string `json:"ti_item_budget_status"`
	ChkItemAutoPowerBudget bool   `json:"chk_item_auto_power_budget"`
	ChkDerateAmps          bool   `json:"chk_derate_amps"`

	// Network Configuration
	IPAddresses           string `json:"ip_addresses"`
	IPAddressPortName     string `json:"ip_address_port_name"`
	TifreeDataPortCount   int    `json:"tifree_data_port_count"`
	TifreePowerPortCount  int    `json:"tifree_power_port_count"`
	TiPxUsername          string `json:"ti_px_username"`
	TiSnmpWriteCommString string `json:"ti_snmp_write_comm_string"`

	// Technical Specifications
	TiUsers       int    `json:"ti_users"`
	TiRam         int    `json:"ti_ram"`
	TiProcesses   int    `json:"ti_processes"`
	TiCpuQuantity int    `json:"ti_cpu_quantity"`
	TiCpuType     string `json:"ti_cpu_type"`
	TiPartNumber  string `json:"ti_part_number"`

	// Important Dates
	InstallationDate    *time.Time `json:"installation_date"`
	ContractEndDate     *time.Time `json:"contract_end_date"`
	PurchaseDate        *time.Time `json:"purchase_date"`
	TiPlannedDecommDate *time.Time `json:"ti_planned_decomm_date"`

	// Custom Fields
	TiCustomFieldPrimaryContact         *string    `json:"ti_custom_field_primary_contact"`
	TiCustomFieldContactTeamName        *string    `json:"ti_custom_field_contact_team_name"`
	TiCustomFieldAuditRemarks           *string    `json:"ti_custom_field_audit_remarks"`
	TiCustomFieldAuditDate              *time.Time `json:"ti_custom_field_audit_date"`
	TiCustomFieldAuditBy                *string    `json:"ti_custom_field_audit_by"`
	TiCustomFieldWarrantyExpirationDate *time.Time `json:"ti_custom_field_warranty_expiration_date"`
	TiCustomFieldAssetStatus            *string    `json:"ti_custom_field_asset_status"`
	TiCustomFieldPntIt                  *string    `json:"ti_custom_field_pnt_it"`

	// Administrative Data
	CmbSystemAdminTeam string `json:"cmb_system_admin_team"`
	CmbSystemAdmin     string `json:"cmb_system_admin"`
	CmbCustomer        string `json:"cmb_customer"`
	TiPoNumber         string `json:"ti_po_number"`
	TiNotes            string `json:"ti_notes"`

	// System timestamps
	LastUpdatedAt time.Time `json:"last_updated_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ItemsParams represents parameters for DCTrack /items API endpoint
type ItemsParams struct {
	// Pagination
	PageNumber int `url:"pageNumber,omitempty"` // Page number to retrieve (default: 1)
	PageSize   int `url:"pageSize,omitempty"`   // Number of items per page (default: 1000)

	// Basic filters
	Location   string `url:"location,omitempty"`   // Filter by location (partial match)
	Status     string `url:"status,omitempty"`     // Filter by status (e.g., "Installed", "Planned")
	ItemClass  string `url:"itemClass,omitempty"`  // Filter by item class (e.g., "Device", "Network")
	Make       string `url:"make,omitempty"`       // Filter by manufacturer
	Model      string `url:"model,omitempty"`      // Filter by model
	SearchText string `url:"searchText,omitempty"` // General text search across fields
}

// PowerSummary represents power consumption analytics from DCTrack
type PowerSummary struct {
	Location     string  `json:"location"`
	TotalPower   float64 `json:"total_power"`
	AveragePower float64 `json:"average_power"`
	AssetCount   int     `json:"asset_count"`
	PowerDensity float64 `json:"power_density"`
	MaxPower     float64 `json:"max_power"`
	MinPower     float64 `json:"min_power"`
}

// Client provides DCTrack API access
type Client struct {
	config     Config
	httpClient *http.Client
	logger     *zap.Logger
	token      string // Store the JWT token after login
}

// NewClient creates a new DCTrack API client
func NewClient(config Config) (*Client, error) {
	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Set defaults
	if config.PageSize <= 0 {
		config.PageSize = 1000
	}
	if config.MaxRetries <= 0 {
		config.MaxRetries = 3
	}
	if config.RetryDelay <= 0 {
		config.RetryDelay = time.Second
	}

	// Create transport with TLS config
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !config.VerifySSL,
		},
	}

	// Create HTTP client with transport and timeout
	httpClient := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}

	return &Client{
		config:     config,
		httpClient: httpClient,
		logger:     zap.NewNop(), // Default to no-op logger
	}, nil
}

// New creates a new DCTrack client with simple parameters
func New(url, username, password string) (*Client, error) {
	config := Config{
		URL:              url,
		Username:         username,
		Password:         password,
		PageSize:         1000,
		MaxRetries:       3,
		RetryDelay:       time.Second,
		VerifySSL:        true,
		RequestAllFields: true,
	}

	return NewClient(config)
}

// SetLogger sets a custom logger for the client
func (c *Client) SetLogger(logger *zap.Logger) {
	if logger != nil {
		c.logger = logger
	}
}

// GetItems retrieves all items from DCTrack API
func (c *Client) GetItems(ctx context.Context) ([]DCTrackItem, error) {
	return c.GetItemsWithParams(ctx, ItemsParams{})
}

// GetItemsWithParams retrieves items with specific parameters
func (c *Client) GetItemsWithParams(ctx context.Context, params ItemsParams) ([]DCTrackItem, error) {
	// Login first to get the JWT token
	if err := c.login(ctx); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
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

		payload := c.buildFieldsPayload()

		items, err := c.fetchPage(ctx, url, payload)
		if err != nil {
			return nil, err
		}

		if len(items) == 0 {
			break
		}

		allItems = append(allItems, items...)

		c.logger.Debug("Fetched DCTrack page",
			zap.Int("page", page),
			zap.Int("items_count", len(items)),
			zap.Int("total_items", len(allItems)))

		// If we got fewer items than requested, we're done
		if len(items) < pageSize {
			break
		}

		// If single page requested, stop after first page
		if params.PageNumber > 0 {
			break
		}

		page++
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
	if query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	return c.GetItemsWithParams(ctx, ItemsParams{
		SearchText: query,
	})
}

// Close cleans up the client resources
func (c *Client) Close() error {
	// Close any resources if needed
	return nil
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
	req.Header.Set("User-Agent", UserAgent)

	c.logger.Debug("Making DCTrack login request",
		zap.String("method", "POST"),
		zap.String("url", loginURL),
		zap.String("username", c.config.Username))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
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
			zap.Int("attempt", attempt))

		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payloadBytes))
		if err != nil {
			return nil, fmt.Errorf("error creating request: %w", err)
		}

		// Add headers
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.token)
		req.Header.Set("User-Agent", UserAgent)

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

		// Read response body first for debugging
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response body: %w", err)
		}

		c.logger.Debug("DCTrack raw response received",
			zap.Int("response_size", len(bodyBytes)),
			zap.String("response_preview", string(bodyBytes)[:minInt(200, len(bodyBytes))]))

		// Parse response with detailed logging
		var dctrackResp DCTrackResponse
		if err := json.Unmarshal(bodyBytes, &dctrackResp); err != nil {
			c.logger.Error("Failed to decode DCTrack response",
				zap.Error(err),
				zap.String("response_preview", string(bodyBytes)[:minInt(500, len(bodyBytes))]))
			return nil, fmt.Errorf("error decoding response: %w", err)
		}

		c.logger.Debug("DCTrack API response parsed",
			zap.Int("total_rows", dctrackResp.TotalRows),
			zap.Int("page_number", dctrackResp.PageNumber),
			zap.Int("page_size", dctrackResp.PageSize),
			zap.Int("items_in_response", len(dctrackResp.SearchResults.Items)))

		var items []DCTrackItem
		for i, record := range dctrackResp.SearchResults.Items {
			item, err := c.mapDCTrackRecord(record)
			if err != nil {
				c.logger.Warn("Error mapping DCTrack record",
					zap.Int("record_index", i),
					zap.Error(err))
				continue
			}
			items = append(items, item)
		}

		c.logger.Debug("DCTrack items mapped successfully",
			zap.Int("raw_items_count", len(dctrackResp.SearchResults.Items)),
			zap.Int("mapped_items_count", len(items)))

		return items, nil
	}

	return nil, fmt.Errorf("all retry attempts failed, last error: %w", lastErr)
}

// DCTrackResponse represents the actual response structure from DCTrack API
type DCTrackResponse struct {
	TotalRows     int `json:"totalRows"`
	PageNumber    int `json:"pageNumber"`
	PageSize      int `json:"pageSize"`
	SearchResults struct {
		Items []map[string]interface{} `json:"items"`
	} `json:"searchResults"`
}

// Helper functions

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func validateConfig(config Config) error {
	if config.URL == "" {
		return fmt.Errorf("URL is required")
	}
	if config.Username == "" {
		return fmt.Errorf("Username is required")
	}
	if config.Password == "" {
		return fmt.Errorf("Password is required")
	}
	return nil
}

func (c *Client) buildFieldsPayload() map[string]interface{} {
	// Use minimal payload that works (confirmed by testing)
	// Complex field selection was causing issues
	return map[string]interface{}{}
}

func (c *Client) mapDCTrackRecord(record map[string]interface{}) (DCTrackItem, error) {
	item := DCTrackItem{}

	// Helper functions for safe type conversion
	getString := func(key string) string {
		if val, ok := record[key]; ok && val != nil {
			return fmt.Sprintf("%v", val)
		}
		return ""
	}

	getFloat := func(key string) float64 {
		if val, ok := record[key]; ok && val != nil {
			if str := fmt.Sprintf("%v", val); str != "" && str != "null" {
				var f float64
				if _, err := fmt.Sscanf(str, "%f", &f); err == nil {
					return f
				}
			}
		}
		return 0
	}

	getInt := func(key string) int {
		if val, ok := record[key]; ok && val != nil {
			if str := fmt.Sprintf("%v", val); str != "" && str != "null" {
				var i int
				if _, err := fmt.Sscanf(str, "%d", &i); err == nil {
					return i
				}
			}
		}
		return 0
	}

	getTime := func(key string) *time.Time {
		if val, ok := record[key]; ok && val != nil {
			if str := fmt.Sprintf("%v", val); str != "" && str != "null" {
				// Try multiple time formats
				formats := []string{
					"2006-01-02 15:04:05-07",
					"2006-01-02T15:04:05Z",
					"2006-01-02",
					time.RFC3339,
				}
				for _, format := range formats {
					if t, err := time.Parse(format, str); err == nil {
						return &t
					}
				}
			}
		}
		return nil
	}

	// Map core fields (fixed to match actual DCTrack field names)
	item.ID = getString("id")
	item.Name = getString("tiName")
	item.ItemClass = getString("tiClass")
	item.Type = getString("tiClass")
	item.Status = getString("cmbStatus")
	item.Location = getString("cmbLocation")
	item.Cabinet = getString("cmbCabinet")
	item.Rack = getString("cmbCabinet")       // DCTrack uses cabinet for both
	item.Position = getString("cmbUPosition") // Fixed: was cmbPosition, should be cmbUPosition
	item.Height = getInt("tiRUs")
	item.Make = getString("cmbMake")
	item.Model = getString("cmbModel")
	item.SerialNumber = getString("tiSerialNumber")
	item.OriginalPower = getFloat("tiItemOriginalPower")
	item.PowerConsumption = getFloat("tiEffectivePower")
	item.SystemAdminTeam = getString("cmbSystemAdminTeam")
	item.PrimaryContact = getString("tiCustomField_Primary Contact")
	item.InstallDate = getTime("installationDate")
	item.LastServiceDate = getTime("lastServiceDate")

	// Set timestamps
	if lastUpdated := getTime("lastUpdatedOn"); lastUpdated != nil {
		item.LastUpdatedAt = *lastUpdated
	} else {
		item.LastUpdatedAt = time.Now()
	}
	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()

	// Validate required fields for database NOT NULL constraints
	if item.ID == "" {
		return item, fmt.Errorf("missing required field: id")
	}
	if item.Name == "" {
		return item, fmt.Errorf("missing required field: tiName")
	}
	if item.ItemClass == "" {
		return item, fmt.Errorf("missing required field: tiClass")
	}
	if item.Status == "" {
		return item, fmt.Errorf("missing required field: cmbStatus")
	}
	if item.Location == "" {
		return item, fmt.Errorf("missing required field: cmbLocation")
	}

	return item, nil
}

// FilterBuilder provides a fluent interface for building DCTrack API filters
type FilterBuilder struct {
	params ItemsParams
}

// NewFilterBuilder creates a new filter builder
func NewFilterBuilder() *FilterBuilder {
	return &FilterBuilder{
		params: ItemsParams{},
	}
}

// Location sets the location filter
func (f *FilterBuilder) Location(location string) *FilterBuilder {
	f.params.Location = location
	return f
}

// Status sets the status filter
func (f *FilterBuilder) Status(status string) *FilterBuilder {
	f.params.Status = status
	return f
}

// Make sets the manufacturer filter
func (f *FilterBuilder) Make(make string) *FilterBuilder {
	f.params.Make = make
	return f
}

// Model sets the model filter
func (f *FilterBuilder) Model(model string) *FilterBuilder {
	f.params.Model = model
	return f
}

// SearchText sets the general search text
func (f *FilterBuilder) SearchText(text string) *FilterBuilder {
	f.params.SearchText = text
	return f
}

// WithPagination sets pagination parameters
func (f *FilterBuilder) WithPagination(page, size int) *FilterBuilder {
	f.params.PageNumber = page
	f.params.PageSize = size
	return f
}

// Build returns the constructed parameters
func (f *FilterBuilder) Build() ItemsParams {
	return f.params
}

// Common filter presets

// InstalledOnly returns parameters for installed items only
func InstalledOnly() ItemsParams {
	return ItemsParams{Status: "Installed"}
}

// ByLocation returns parameters filtered by location
func ByLocation(location string) ItemsParams {
	return ItemsParams{
		Location: location,
		Status:   "Installed",
	}
}

// ByVendor returns parameters filtered by manufacturer
func ByVendor(make string) ItemsParams {
	return ItemsParams{
		Make:   make,
		Status: "Installed",
	}
}
