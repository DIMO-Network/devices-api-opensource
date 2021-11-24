package services

//go:generate mockgen -source nhtsa_api_service.go -destination mocks/nhtsa_api_service_mock.go

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
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
	req, err := http.NewRequest(http.MethodGet, url, nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, errors.New("received a non 200 response from nhtsa api")
	}
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body from nhtsa api")
	}
	decodedVin := NHTSADecodeVINResponse{}
	_ = json.Unmarshal(resBody, &decodedVin)

	return &decodedVin, nil
}

type NHTSADecodeVINResponse struct {
	Count          int    `json:"Count"`
	Message        string `json:"Message"`
	SearchCriteria string `json:"SearchCriteria"`
	Results        []struct {
		Value      *string `json:"Value"`
		ValueId    *string `json:"ValueId"`
		Variable   string  `json:"Variable"`
		VariableId int     `json:"VariableId"`
	} `json:"Results"`
}

func (n *NHTSADecodeVINResponse) LookupValue(variableName string) string {
	for _, result := range n.Results {
		if result.Variable == variableName {
			if result.Value != nil {
				return *result.Value
			}
			return ""
		}
	}
	return ""
}
