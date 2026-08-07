package resources

import "strings"

// azureRegionContinents maps Azure region names to their continent ("location" in
// stagehand's terms). This mirrors the `location` values for the `azure` provider
// under the `HOSTED` clusterType in apps/stagehand/pkg/assets/provider-specs.yaml,
// which is what backs apps/core's ValidateEnableReplicationTimeControl (used to
// validate that a Replication Time Control-enabled DR cluster's region and dr_region
// are on the same continent). Keep this in sync if Azure regions are added/changed
// there.
var azureRegionContinents = map[string]string{
	"eastus2":            "americas",
	"eastus":             "americas",
	"northeurope":        "europe",
	"australiaeast":      "asiapacific",
	"westeurope":         "europe",
	"westus2":            "americas",
	"japaneast":          "asiapacific",
	"canadacentral":      "americas",
	"uksouth":            "europe",
	"brazilsouth":        "americas",
	"centralindia":       "asiapacific",
	"francecentral":      "europe",
	"centralus":          "americas",
	"westus3":            "americas",
	"southcentralus":     "americas",
	"uaenorth":           "middleeast",
	"germanywestcentral": "europe",
	"southafricanorth":   "africa",
}

// azureRegionContinent returns the continent for an Azure region name, matched
// case-insensitively (mirroring stagehand.GetRegionLocations), and whether the
// region was recognized.
func azureRegionContinent(region string) (string, bool) {
	for name, continent := range azureRegionContinents {
		if strings.EqualFold(name, region) {
			return continent, true
		}
	}
	return "", false
}
