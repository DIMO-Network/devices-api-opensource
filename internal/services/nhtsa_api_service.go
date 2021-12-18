package services

//go:generate mockgen -source nhtsa_api_service.go -destination mocks/nhtsa_api_service_mock.go

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

type INHTSAService interface {
	DecodeVIN(vin string) (*NHTSADecodeVINResponse, error)
}

type NHTSAService struct {
	baseURL string
}

func NewNHTSAService() INHTSAService {
	return &NHTSAService{
		baseURL: "https://vpic.nhtsa.dot.gov/api/",
	}
}

func (ns *NHTSAService) DecodeVIN(vin string) (*NHTSADecodeVINResponse, error) {
	url := fmt.Sprintf("%s/vehicles/decodevinextended/%s?format=json", ns.baseURL, vin)
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received a non 200 response from nhtsa api. status code: %d", res.StatusCode)
	}

	decodedVin := NHTSADecodeVINResponse{}
	err = json.NewDecoder(res.Body).Decode(&decodedVin)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body from nhtsa api")
	}

	for _, r := range decodedVin.Results {
		if r.Variable == "Error Code" {
			if r.Value != "0" && r.Value != "" {
				return nil, fmt.Errorf("nhtsa api responded with error code %s", r.Value)
			}
			break
		}
	}

	return &decodedVin, nil
}

type NHTSADecodeVINResponse struct {
	Count          int    `json:"Count"`
	Message        string `json:"Message"`
	SearchCriteria string `json:"SearchCriteria"`
	Results        []struct {
		Value      string `json:"Value"`
		ValueID    string `json:"ValueId"`
		Variable   string `json:"Variable"`
		VariableID int    `json:"VariableId"`
	} `json:"Results"`
}

// LookupValue looks up value in nhtsa object, and uppercase the resulting value
func (n *NHTSADecodeVINResponse) LookupValue(variableName string) string {
	for _, result := range n.Results {
		if result.Variable == variableName {
			return strings.ToUpper(result.Value)
		}
	}
	return ""
}
