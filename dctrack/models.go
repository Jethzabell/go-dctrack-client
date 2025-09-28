package dctrack

import "time"

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

	// Core Asset Information (ti_ fields from DCTrack)
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

// AnalyticsSummary represents site-level analytics from DCTrack
type AnalyticsSummary struct {
	Location             string  `json:"location"`
	TotalAssets          int     `json:"total_assets"`
	TotalPowerKW         float64 `json:"total_power_kw"`
	TotalRUs             int     `json:"total_rus"`
	TotalPurchaseValue   float64 `json:"total_purchase_value"`
	TotalWeight          float64 `json:"total_weight"`
	AveragePowerPerAsset float64 `json:"average_power_per_asset"`
	PowerDensityPerRU    float64 `json:"power_density_per_ru"`
}

// RefreshPlan represents asset refresh planning data
type RefreshPlan struct {
	TotalAssetsTracked int               `json:"total_assets_tracked"`
	ByContractEndYear  map[string]int    `json:"by_contract_end_year"`
	ByInstallYear      map[string]int    `json:"by_install_year"`
	UpcomingRefreshes  []UpcomingRefresh `json:"upcoming_refreshes"`
}

// UpcomingRefresh represents assets due for refresh
type UpcomingRefresh struct {
	Year       string   `json:"year"`
	AssetCount int      `json:"asset_count"`
	TotalValue float64  `json:"total_value"`
	Locations  []string `json:"locations"`
}

// ContactAudit represents contact audit results
type ContactAudit struct {
	TotalAssets        int `json:"total_assets"`
	MissingPrimary     int `json:"missing_primary_contact"`
	MissingExecSponsor int `json:"missing_exec_sponsor"`
	MissingWarranty    int `json:"missing_warranty_expiration"`
}

// VendorDistribution represents vendor/make distribution
type VendorDistribution struct {
	Make       string  `json:"make"`
	AssetCount int     `json:"asset_count"`
	Percentage float64 `json:"percentage"`
}

// CabinetUtilization represents cabinet space utilization
type CabinetUtilization struct {
	Location           string  `json:"location"`
	Cabinet            string  `json:"cabinet"`
	UsedRU             int     `json:"used_ru"`
	TotalRU            int     `json:"total_ru"`
	UtilizationPercent float64 `json:"utilization_percent"`
	AssetCount         int     `json:"asset_count"`
}

// EfficiencyMetrics represents efficiency analysis
type EfficiencyMetrics struct {
	TotalAssets          int               `json:"total_assets"`
	AvgPowerPerRU        float64           `json:"avg_power_per_ru"`
	AvgCostPerRU         float64           `json:"avg_cost_per_ru"`
	PowerEfficiencyScore float64           `json:"power_efficiency_score"`
	SpaceEfficiencyScore float64           `json:"space_efficiency_score"`
	CostEfficiencyScore  float64           `json:"cost_efficiency_score"`
	TopPerformers        []AssetEfficiency `json:"top_performers"`
}
