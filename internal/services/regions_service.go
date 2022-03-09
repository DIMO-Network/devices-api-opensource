package services

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tidwall/gjson"
)

//go:embed countries_regions.json
var countriesJSON string

// FindCountry finds country by 3 digit country code and returns info including region. returns nil if nothing found
func FindCountry(countryCode string) *CountryInfo {
	countryCode = strings.ToUpper(countryCode)
	cj := gjson.Get(countriesJSON, fmt.Sprintf("#(alpha_3==%q)", countryCode))
	if !cj.Exists() {
		return nil
	}
	c := &CountryInfo{}

	err := json.Unmarshal([]byte(cj.String()), c)
	if err != nil {
		return nil
	}
	return c
}

// CountryInfo represent a country with it's region and sub-region
type CountryInfo struct {
	Name          string `json:"name"`
	Alpha3        string `json:"alpha_3"`
	Region        string `json:"region"`
	SubRegion     string `json:"sub_region"`
	RegionCode    int    `json:"region_code"`
	SubRegionCode int    `json:"sub_region_code"`
}
