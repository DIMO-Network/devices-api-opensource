package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type IEdmundsService interface {
	GetDefaultImageForMMY(make, model string, year int) (*string, error)
}

type EdmundsService struct {
	baseAPIURL   string
	torProxyURL  string
	baseMediaURL string
}

func NewEdmundsService(torProxyURL string) *EdmundsService {
	return &EdmundsService{torProxyURL: torProxyURL, baseAPIURL: "https://www.edmunds.com/gateway/api", baseMediaURL: "https://media.ed.edmunds-media.com"}
}

var ErrVehicleNotFound = errors.New("vehicle not found")

func (e EdmundsService) getAllPhotosForMMY(make, model, year string) (*photosResponse, error) {
	make = strings.ReplaceAll(make, " ", "_")
	model = strings.ReplaceAll(model, " ", "_")
	listURL := fmt.Sprintf("%s/media/v2/%s/%s/%s/photos?format=json", e.baseAPIURL, strings.ToLower(make), strings.ToLower(model), year)

	req, err := http.NewRequest("GET", listURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Host", "www.edmunds.com")
	req.Header.Set("x-client-action-name", "edmunds-ios-anypage")
	req.Header.Set("Accept-Language", "en-us")
	req.Header.Set("x-artifact-id", "edmunds-ios")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Edmunds/790 CFNetwork/1312 Darwin/21.0.0")
	req.Header.Set("Referer", "https://www.edmunds.com")

	var client *http.Client
	if e.torProxyURL != "" {
		proxyURL, err := url.Parse(e.torProxyURL)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse Tor proxy URL")
		}
		client = &http.Client{
			Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)},
		}
	} else {
		client = http.DefaultClient
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusBadRequest {
			return nil, ErrVehicleNotFound
		}
		return nil, fmt.Errorf("received a non 200 response from edmunds photos api. status code: %d", res.StatusCode)
	}
	photos := photosResponse{}
	err = json.NewDecoder(res.Body).Decode(&photos)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body from edmunds photos api")
	}

	return &photos, nil
}

// GetDefaultImageForMMY call edmunds photos api and finds the first Frontal image, returning it in size 600
func (e EdmundsService) GetDefaultImageForMMY(make, model string, year int) (*string, error) {
	photos, err := e.getAllPhotosForMMY(make, model, strconv.Itoa(year))
	if err != nil {
		return nil, err
	}
	for _, photo := range photos.Photos {
		// find the first Front image https://developer.edmunds.com/api-documentation/media/photos/v2/
		if len(photo.ShotTypeAbbreviation) > 0 && photo.ShotTypeAbbreviation[:1] == "F" {
			return e.buildImageURL(photo.Sources), nil
		}
	}
	// if no F photo found, just return the first Exterior image
	for _, photo := range photos.Photos {
		// find the first exterior photo
		if photo.Category == "EXTERIOR" {
			return e.buildImageURL(photo.Sources), nil
		}
	}

	return nil, nil
}

func (e *EdmundsService) buildImageURL(sources []photosResponseSource) *string {
	if len(sources) > 3 {
		img := e.baseMediaURL + sources[3].Link.Href
		return &img
	}
	return nil
}

type photosResponse struct {
	Photos []struct {
		Title                string                 `json:"title"`
		Category             string                 `json:"category,omitempty"`
		Tags                 []string               `json:"tags"`
		Provider             string                 `json:"provider"`
		Sources              []photosResponseSource `json:"sources"`
		Makes                []string               `json:"makes"`
		Models               []string               `json:"models"`
		Years                []string               `json:"years"`
		Color                string                 `json:"color"`
		Submodels            []string               `json:"submodels"`
		Trims                []string               `json:"trims"`
		ModelYearID          int                    `json:"modelYearId"`
		ShotTypeAbbreviation string                 `json:"shotTypeAbbreviation"`
		StyleIds             []string               `json:"styleIds"`
		ExactStyleIds        []string               `json:"exactStyleIds"`
	} `json:"photos"`
	PhotosCount int `json:"photosCount"`
	Links       []struct {
		Rel  string `json:"rel"`
		Href string `json:"href"`
	} `json:"links"`
}

type photosResponseSource struct {
	Link struct {
		Rel  string `json:"rel"`
		Href string `json:"href"`
	} `json:"link"`
	Extension string `json:"extension"`
	Size      struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"size"`
}
