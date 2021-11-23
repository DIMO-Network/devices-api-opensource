package services

import (
	"fmt"
	"github.com/pkg/errors"
	"net/http"
)

type NHTSAService struct {
	BaseURL string
}

type INHTSAService interface {
	DecodeVIN(vin string) error
}

func NewNHTSAService() NHTSAService {
	return NHTSAService{
		BaseURL: "https://vpic.nhtsa.dot.gov/api/",
	}
}

func (ns *NHTSAService) DecodeVIN(vin string) error {
	url := fmt.Sprintf("%s/vehicles/decodevinextended/%s?format=json", ns.BaseURL, vin)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return errors.New("received a non 200 response")
	}
	// todo: parse response body

	return nil
}
