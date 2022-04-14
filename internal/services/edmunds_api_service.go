package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/DIMO-Network/shared"
	"github.com/ahmetb/go-linq/v3"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type EdmundsService interface {
	GetDefaultImageForMMY(make, model string, year int) (*string, error)
	GetFlattenedVehicles() (*[]FlatMMYDefinition, error)
	findImageURLByMaxWidth(sources []photosResponseSource, maxWidth int) *string
	getAllMakes() (*makesResponse, error)
}

type edmundsService struct {
	baseMediaURL string
	log          *zerolog.Logger
	httpClient   shared.HTTPClientWrapper
}

const (
	edmundsAPIURL       = "https://www.edmunds.com/gateway"
	edmundsBaseMediaURL = "https://media.ed.edmunds-media.com" // we don't actually call this, just use it to generate image urls.
)

// NewEdmundsService if torProxyURL is empty, will not go via tor
func NewEdmundsService(torProxyURL string, logger *zerolog.Logger) EdmundsService {
	h := map[string]string{
		"Host":                 "www.edmunds.com",
		"x-client-action-name": "edmunds-ios-anypage",
		"Accept-Language":      "en-us",
		"x-artifact-id":        "edmunds-ios",
		"Accept":               "application/json",
		"User-Agent":           "Edmunds/790 CFNetwork/1312 Darwin/21.0.0",
		"Referer":              "https://www.edmunds.com"}
	hcw, _ := shared.NewHTTPClientWrapper(edmundsAPIURL, torProxyURL, 10*time.Second, h, true) // ok to ignore err since only used for tor check

	return &edmundsService{log: logger, httpClient: hcw, baseMediaURL: edmundsBaseMediaURL}
}

var ErrVehicleNotFound = errors.New("vehicle not found in Edmunds")

func (e *edmundsService) getAllMakes() (*makesResponse, error) {
	res, err := e.httpClient.ExecuteRequest("/api/vehicle/v2/makes", "GET", nil)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received a non 200 response from edmunds. status code: %d", res.StatusCode)
	}
	defer res.Body.Close() //nolint

	items := makesResponse{}
	err = json.NewDecoder(res.Body).Decode(&items)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body from edmunds api")
	}

	return &items, nil
}

func (e *edmundsService) getModelsForMake(makeNiceName string) (*modelsResponse, error) {
	res, err := e.httpClient.ExecuteRequest(fmt.Sprintf("/api/vehicle/v2/%s/models", makeNiceName), "GET", nil)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received a non 200 response from edmunds. status code: %d", res.StatusCode)
	}
	defer res.Body.Close() //nolint

	items := modelsResponse{}
	err = json.NewDecoder(res.Body).Decode(&items)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body from edmunds api")
	}

	return &items, nil
}

func (e *edmundsService) GetFlattenedVehicles() (*[]FlatMMYDefinition, error) {
	makes, err := e.getAllMakes()
	if err != nil {
		return nil, err
	}
	e.log.Info().Msgf("found makes: %d", makes.MakesCount)

	var flattened []FlatMMYDefinition
	for _, mk := range makes.Makes {
		models, err := e.getModelsForMake(mk.NiceName) // this could time out
		if err != nil {
			// log and skip this one
			e.log.Err(err).Msgf("requests to get models failed for make: %s. ignoring and continuing.", mk.NiceName)
			continue
		}
		e.log.Info().Msgf("found models %d for: %s", models.ModelsCount, mk.NiceName)
		for _, model := range models.Models {
			for _, year := range model.Years {
				var subModels []string
				var styles []FlatMMYDefinitionStyle

				for _, style := range year.Styles {
					// dedupe submodel (trim)
					exists := linq.From(subModels).AnyWith(func(sm interface{}) bool {
						return sm.(string) == style.Trim
					})
					if !exists {
						subModels = append(subModels, style.Trim)
					}
					// dedupe styles (name + trim)
					exists = linq.From(styles).AnyWith(func(s interface{}) bool {
						return s.(FlatMMYDefinitionStyle).Name == style.Name &&
							s.(FlatMMYDefinitionStyle).Trim == style.Trim
					})
					if !exists {
						styles = append(styles, FlatMMYDefinitionStyle{
							StyleID: style.ID,
							Name:    style.Name,
							Trim:    style.Trim,
						})
					}
				}
				definition := FlatMMYDefinition{
					Make:        mk.Name,
					ModelYearID: year.ID,
					Model:       model.Name,
					Year:        year.Year,
					SubModels:   subModels,
					Styles:      styles,
				}
				flattened = append(flattened, definition)
			}
		}
	}
	return &flattened, nil
}

func (e *edmundsService) getAllPhotosForMMY(make, model, year string, overridePath *string) (*photosResponse, error) {
	make = strings.ReplaceAll(make, " ", "_")
	model = strings.ReplaceAll(model, " ", "_")
	var photosURL string
	if overridePath != nil {
		photosURL = *overridePath
	} else {
		photosURL = fmt.Sprintf("/api/media/v2/%s/%s/%s/photos?format=json&pageSize=50", strings.ToLower(make), strings.ToLower(model), year)
	}
	res, err := e.httpClient.ExecuteRequest(photosURL, "GET", nil)
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
func (e *edmundsService) GetDefaultImageForMMY(make, model string, year int) (*string, error) {
	const maxWidth = 900
	const shotType = "FQ"
	photos, err := e.getAllPhotosForMMY(make, model, strconv.Itoa(year), nil)
	if err != nil {
		return nil, err
	}
	best := findPhotoByShotType(photos, shotType)
	cont := len(best) == 0
	// keep looping to find desired image
	for cont {
		relNextFound := false
		for _, link := range photos.Links {
			if link.Rel == "next" {
				relNextFound = true
				photos, err = e.getAllPhotosForMMY(make, model, strconv.Itoa(year), &link.Href)
				if err != nil {
					return nil, err
				}
				best = findPhotoByShotType(photos, shotType)
				break
			}
		}
		cont = len(best) == 0 && relNextFound
	}
	if len(best) == 0 {
		if len(photos.Photos) > 0 {
			best = photos.Photos[0].Sources
		} else {
			return nil, nil
		}
	}

	return e.findImageURLByMaxWidth(best, maxWidth), nil
}

// findPhotoByShotType looks for shot types of shotType, usually want FQ
func findPhotoByShotType(photos *photosResponse, shotType string) []photosResponseSource {
	// find the first Front image https://developer.edmunds.com/api-documentation/media/photos/v2/
	for _, photo := range photos.Photos {
		if strings.ToUpper(photo.ShotTypeAbbreviation) == shotType {
			return photo.Sources
		}
	}
	return nil
}

func (e *edmundsService) findImageURLByMaxWidth(sources []photosResponseSource, maxWidth int) *string {
	source := linq.From(sources).WhereT(func(p photosResponseSource) bool {
		return maxWidth >= p.Size.Width
	}).OrderByDescendingT(func(p photosResponseSource) int {
		return p.Size.Width
	}).First()

	if source != nil {
		img := e.baseMediaURL + source.(photosResponseSource).Link.Href
		return &img
	}
	return nil
}

// FlatMMYDefinition represents edmunds properties for a vehicle
type FlatMMYDefinition struct {
	Make  string
	Model string
	Year  int
	// edmunds response: models.years.id
	ModelYearID int
	// SubModels are edmunds Trims
	SubModels []string
	Styles    []FlatMMYDefinitionStyle
}

// FlatMMYDefinitionStyle edmunds style level properties
type FlatMMYDefinitionStyle struct {
	// Edmunds StyleId
	StyleID int
	Name    string
	Trim    string
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

type makesResponse struct {
	Makes []struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		NiceName string `json:"niceName"`
		Models   []struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			NiceName string `json:"niceName"`
			Years    []struct {
				ID   int `json:"id"`
				Year int `json:"year"`
			} `json:"years"`
		} `json:"models"`
	} `json:"makes"`
	MakesCount int `json:"makesCount"`
}

type modelsResponse struct {
	Models []struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		NiceName string `json:"niceName"`
		Years    []struct {
			ID     int `json:"id"`
			Year   int `json:"year"`
			Styles []struct {
				ID       int    `json:"id"`
				Name     string `json:"name"`
				Submodel struct {
					Body      string `json:"body"`
					ModelName string `json:"modelName"`
					NiceName  string `json:"niceName"`
					Fuel      string `json:"fuel,omitempty"`
					Tuner     string `json:"tuner,omitempty"`
				} `json:"submodel"`
				Trim string `json:"trim"`
			} `json:"styles"`
		} `json:"years"`
	} `json:"models"`
	ModelsCount int `json:"modelsCount"`
}
