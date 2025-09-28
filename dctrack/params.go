package dctrack

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

// SearchParams represents parameters for DCTrack /quicksearch API
type SearchParams struct {
	Query         string   `url:"query"`                   // Search query text
	PageNumber    int      `url:"pageNumber,omitempty"`    // Page number (default: 1)
	PageSize      int      `url:"pageSize,omitempty"`      // Items per page (default: 1000)
	IncludeFields []string `url:"includeFields,omitempty"` // Specific fields to include
}

// PowerParams represents parameters for power-related queries
type PowerParams struct {
	Location  string  `url:"location,omitempty"`  // Filter by location
	Cabinet   string  `url:"cabinet,omitempty"`   // Filter by cabinet
	MinPower  float64 `url:"minPower,omitempty"`  // Minimum power threshold
	MaxPower  float64 `url:"maxPower,omitempty"`  // Maximum power threshold
	PowerType string  `url:"powerType,omitempty"` // Power type (original, consumption, potential)
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

// ItemClass sets the item class filter
func (f *FilterBuilder) ItemClass(itemClass string) *FilterBuilder {
	f.params.ItemClass = itemClass
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

// PowerAssets returns parameters for assets with power data
func PowerAssets() ItemsParams {
	return ItemsParams{
		Status: "Installed",
		// Will be filtered by client to include only items with power > 0
	}
}
