package dctrack

import (
	"fmt"
	"strconv"
	"time"
)

// DCTrackResponse represents the response structure from DCTrack API
type DCTrackResponse struct {
	Records []map[string]interface{} `json:"records"`
}

// AssetEfficiency represents efficiency metrics for an asset
type AssetEfficiency struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Location        string  `json:"location"`
	Cabinet         string  `json:"cabinet"`
	Make            string  `json:"make"`
	Model           string  `json:"model"`
	PowerEfficiency float64 `json:"power_efficiency"`
	SpaceEfficiency float64 `json:"space_efficiency"`
	CostEfficiency  float64 `json:"cost_efficiency"`
	OverallScore    float64 `json:"overall_score"`
}

// mapDCTrackRecord converts a raw DCTrack API record to a DCTrackItem
func (c *Client) mapDCTrackRecord(record map[string]interface{}) (DCTrackItem, error) {
	item := DCTrackItem{}

	// Helper function to safely get string values
	getString := func(key string) string {
		if val, ok := record[key]; ok && val != nil {
			return fmt.Sprintf("%v", val)
		}
		return ""
	}

	// Helper function to safely get float values
	getFloat := func(key string) float64 {
		if val, ok := record[key]; ok && val != nil {
			if str := fmt.Sprintf("%v", val); str != "" && str != "null" {
				if f, err := strconv.ParseFloat(str, 64); err == nil {
					return f
				}
			}
		}
		return 0
	}

	// Helper function to safely get int values
	getInt := func(key string) int {
		if val, ok := record[key]; ok && val != nil {
			if str := fmt.Sprintf("%v", val); str != "" && str != "null" {
				if i, err := strconv.Atoi(str); err == nil {
					return i
				}
			}
		}
		return 0
	}

	// Helper function to safely get bool values
	getBool := func(key string) bool {
		if val, ok := record[key]; ok && val != nil {
			if str := fmt.Sprintf("%v", val); str == "true" || str == "1" {
				return true
			}
		}
		return false
	}

	// Helper function to safely get time values
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

	// Map core identification fields
	item.ID = getString("id")
	item.Name = getString("tiName")
	item.ItemClass = getString("tiClass")
	item.Type = getString("tiClass") // DCTrack uses tiClass for type
	item.Status = getString("cmbStatus")
	item.Location = getString("cmbLocation")
	item.Cabinet = getString("cmbCabinet")
	item.Rack = getString("cmbCabinet") // Often the same in DCTrack
	item.Position = getString("cmbPosition")
	item.Height = getInt("tiRUs")

	// Map basic asset information
	item.Make = getString("cmbMake")
	item.Model = getString("cmbModel")
	item.SerialNumber = getString("tiSerialNumber")
	item.OriginalPower = getFloat("tiItemOriginalPower")
	item.PowerConsumption = getFloat("tiEffectivePower")
	item.SystemAdminTeam = getString("cmbSystemAdminTeam")
	item.PrimaryContact = getString("tiCustomField_Primary Contact")
	item.InstallDate = getTime("installationDate")
	item.LastServiceDate = getTime("lastServiceDate")

	// Map ti_ fields (DCTrack-specific)
	item.TiSubclass = getString("tiSubclass")
	item.TiSerialNumber = getString("tiSerialNumber")
	item.TiAssetTag = getString("tiAssetTag")
	item.TiFormFactor = getString("tiFormFactor")
	item.TiMounting = getString("tiMounting")
	item.TiWidth = getFloat("tiWidth")
	item.TiDepth = getFloat("tiDepth")
	item.TiWeight = getFloat("tiWeight")
	item.TiRuHeight = getString("tiRUs")
	item.TiPotentialPower = getFloat("tiPotentialPower")
	item.TiEffectivePower = getFloat("tiEffectivePower")
	item.TiPowerCapacity = getFloat("tiPowerCapacity")
	item.TiPsRedundancy = getString("tiPSRedundancy")
	item.TiPurchasePrice = getFloat("tiPurchasePrice")
	item.TiContractAmount = getFloat("tiContractAmount")

	// Map infrastructure details
	item.CmbPlantBay = getString("cmbPlantBay")
	item.CmbCabinetID = getString("cmbCabinetId")
	item.CmbRowPosition = getString("cmbRowPosition")
	item.CmbRowLabel = getString("cmbRowLabel")
	item.TiFloorNodeCode = getString("tiFloorNodeCode")
	item.TiFloorName = getString("tiFloorName")
	item.TiRoomNodeCode = getString("tiRoomNodeCode")
	item.TiRoomName = getString("tiRoomName")

	// Map integration status
	item.TiIntegrationStatus = getString("tiIntegrationStatus")
	item.TiVmwareIntegrationStatus = getString("tiVMwareIntegrationStatus")
	item.TiCmdbIntegrationStatus = getString("tiCmdbIntegrationStatus")

	// Map budget and planning
	item.TiItemBudgetStatus = getString("tiItemBudgetStatus")
	item.ChkItemAutoPowerBudget = getBool("chkItemAutoPowerBudget")
	item.ChkDerateAmps = getBool("chkDerateAmps")

	// Map network configuration
	item.IPAddresses = getString("ipAddresses")
	item.IPAddressPortName = getString("ipAddressPortName")
	item.TifreeDataPortCount = getInt("tifreeDataPortCount")
	item.TifreePowerPortCount = getInt("tifreePowerPortCount")
	item.TiPxUsername = getString("tiPXUsername")
	item.TiSnmpWriteCommString = getString("tiSnmpWriteCommString")

	// Map technical specifications
	item.TiUsers = getInt("tiUsers")
	item.TiRam = getInt("tiRAM")
	item.TiProcesses = getInt("tiProcesses")
	item.TiCpuQuantity = getInt("tiCpuQuantity")
	item.TiCpuType = getString("tiCpuType")
	item.TiPartNumber = getString("tiPartNumber")

	// Map important dates
	item.InstallationDate = getTime("installationDate")
	item.ContractEndDate = getTime("contractEndDate")
	item.PurchaseDate = getTime("purchaseDate")
	item.TiPlannedDecommDate = getTime("tiPlannedDecommDate")

	// Map custom fields
	if val := getString("tiCustomField_Primary Contact"); val != "" {
		item.TiCustomFieldPrimaryContact = &val
	}
	if val := getString("tiCustomField_Contact Team Name"); val != "" {
		item.TiCustomFieldContactTeamName = &val
	}
	if val := getString("tiCustomField_Audit Remarks"); val != "" {
		item.TiCustomFieldAuditRemarks = &val
	}
	if val := getString("tiCustomField_Audit By"); val != "" {
		item.TiCustomFieldAuditBy = &val
	}
	if val := getString("tiCustomField_Asset Status"); val != "" {
		item.TiCustomFieldAssetStatus = &val
	}
	if val := getString("tiCustomField_PNT/IT"); val != "" {
		item.TiCustomFieldPntIt = &val
	}

	item.TiCustomFieldAuditDate = getTime("tiCustomField_Audit Date")
	item.TiCustomFieldWarrantyExpirationDate = getTime("tiCustomField_Warranty Expiration Date")

	// Map administrative data
	item.CmbSystemAdminTeam = getString("cmbSystemAdminTeam")
	item.CmbSystemAdmin = getString("cmbSystemAdmin")
	item.CmbCustomer = getString("cmbCustomer")
	item.TiPoNumber = getString("tiPONumber")
	item.TiNotes = getString("tiNotes")

	// Map system timestamps
	if lastUpdated := getTime("lastUpdatedOn"); lastUpdated != nil {
		item.LastUpdatedAt = *lastUpdated
	} else {
		item.LastUpdatedAt = time.Now()
	}
	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()

	return item, nil
}
