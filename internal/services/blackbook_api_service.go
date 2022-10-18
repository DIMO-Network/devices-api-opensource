package services

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/shared"
	"github.com/pkg/errors"
)

/*
Blackbook Used Car API Documentation ***REQUIRES LOGIN***
https://developer.blackbookcloud.com/Documentation?product=Used%20Car%20Web%20API
*/

//go:generate mockgen -source blackbook_api_service.go -destination mocks/blackbook_api_service_mock.go
type BlackbookAPIService interface {
	GetVINInfo(vin string, reqData *ValuationRequestData) ([]byte, error)
	// VIN, State
	GetUniversalVINInfo(vin, state string) ([]byte, error)
	GetBatch(vin, state string) ([]byte, error)
	// UVC, State
	GetLikeVehicles(uvc, state string) ([]byte, error)
	GetUVCInfo(uvc, state string) ([]byte, error)
	// UVC
	GetColors(uvc string) ([]byte, error)
	// YMMSS, State
	GetDrilldown(year int, make, model, series, style, state string) ([]byte, error)
	// Plate, State
	GetPlate(plate, state string) ([]byte, error)

	// Sumary
	GetSummaryByVIN(vin, state string) (*BlackbookVINSummary, error)
	GetSummaryByPlate(plate, state string) (*BlackbookVINSummary, error)
}

type blackbookAPIService struct {
	settings      *config.Settings
	httpClientVIN shared.HTTPClientWrapper
	dbs           func() *database.DBReaderWriter
}

func NewBlackbookAPIService(settings *config.Settings, dbs func() *database.DBReaderWriter) BlackbookAPIService {
	if settings.BlackbookAPIURL == "" || settings.BlackbookAPIUser == "" || settings.BlackbookAPIPassword == "" {
		log.Fatal("Blackbook configuration not set")
	}
	h := map[string]string{"Authorization": "Basic " + basicAuth(settings.BlackbookAPIUser, settings.BlackbookAPIPassword)}
	hcwv, err := shared.NewHTTPClientWrapper(settings.BlackbookAPIURL, "", 10*time.Second, h, true)
	if err != nil {
		log.Fatal(err)
	}

	return &blackbookAPIService{
		settings:      settings,
		httpClientVIN: hcwv,
		dbs:           dbs,
	}
}

func (bs *blackbookAPIService) GetVINInfo(vin string, reqData *ValuationRequestData) ([]byte, error) {
	params := url.Values{}
	if reqData.Mileage != nil {
		params.Add("mileage", strconv.Itoa(int(*reqData.Mileage)))
	}
	if reqData.ZipCode != nil {
		params.Add("state", *reqData.ZipCode)
	}
	res, err := bs.executeAPI(bs.httpClientVIN, fmt.Sprintf("/UsedVehicle/VIN/%s?%s", vin, params.Encode()), nil)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (bs *blackbookAPIService) GetUniversalVINInfo(vin, state string) ([]byte, error) {
	res, err := bs.executeAPI(bs.httpClientVIN, fmt.Sprintf("/Universal/VIN/%s?markets=UU&all_matches=true&state=%s", vin, state), nil)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (bs *blackbookAPIService) GetLikeVehicles(uvc, state string) ([]byte, error) {
	res, err := bs.executeAPI(bs.httpClientVIN, fmt.Sprintf("/UsedVehicle/Like/%s?state=%s", uvc, state), nil)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (bs *blackbookAPIService) GetBatch(vin, state string) ([]byte, error) {
	payload, err := json.Marshal([]batchVehicle{
		{State: state, VIN: vin},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to marshal []batchVehicle")
	}
	res, err := bs.executeAPI(bs.httpClientVIN, "/Batch", payload)

	if err != nil {
		return nil, err
	}

	return res, nil
}

type batchVehicle struct {
	State string `json:"state"`
	VIN   string `json:"vin"`
}

func (bs *blackbookAPIService) GetDrilldown(year int, make, model, series, style, state string) ([]byte, error) {
	params := url.Values{
		"model":  {model},
		"series": {series},
		"style":  {style},
		"state":  {state},
	}
	res, err := bs.executeAPI(bs.httpClientVIN, fmt.Sprintf("/UsedVehicle/%d/%s?%s", year, url.QueryEscape(make), params.Encode()), nil)

	if err != nil {
		return nil, err
	}

	return res, nil
}

// Used Vehicle By License Plate
func (bs *blackbookAPIService) GetPlate(plate, state string) ([]byte, error) {
	res, err := bs.executeAPI(bs.httpClientVIN, fmt.Sprintf("/UsedVehicle/plate/%s/%s", url.QueryEscape(plate), url.QueryEscape(state)), nil)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (bs *blackbookAPIService) GetUVCInfo(uvc, state string) ([]byte, error) {
	res, err := bs.executeAPI(bs.httpClientVIN, fmt.Sprintf("/UsedVehicle/UVC/%s?state=%s", uvc, state), nil)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (bs *blackbookAPIService) GetColors(uvc string) ([]byte, error) {
	res, err := bs.executeAPI(bs.httpClientVIN, fmt.Sprintf("/Colors/%s", uvc), nil)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (bs *blackbookAPIService) GetSummaryByVIN(vin, state string) (*BlackbookVINSummary, error) {

	// Used Vehicle By VIN
	vinRes, err := bs.GetVINInfo(vin, nil)
	if err != nil {
		return nil, err
	}

	// Extract details from vinRes
	type VINResp struct {
		UsedVehicles struct {
			UsedVehiclesList []struct {
				UVC    string `json:"uvc"`
				Year   int    `json:"model_year"`
				Make   string `json:"make"`
				Model  string `json:"model"`
				Series string `json:"series"`
				Style  string `json:"style"`
			} `json:"used_vehicle_list"`
		} `json:"used_vehicles"`
	}
	vr := &VINResp{}
	err = json.Unmarshal(vinRes, vr)
	if err != nil {
		return nil, err
	}
	vd := vr.UsedVehicles.UsedVehiclesList[0]

	// Like Vehicles
	likeVehRes, err := bs.GetLikeVehicles(vd.UVC, state)
	if err != nil {
		return nil, err
	}

	// Universal Lookup By VIN
	uVINRes, err := bs.GetUniversalVINInfo(vin, state)
	if err != nil {
		return nil, err
	}

	// Used Vehicle Batch
	batchRes, err := bs.GetBatch(vin, state)
	if err != nil {
		return nil, err
	}

	// Used Vehicle By Drilldown
	drilldownRes, err := bs.GetDrilldown(vd.Year, vd.Make, vd.Model, vd.Series, vd.Style, state)
	if err != nil {
		return nil, err
	}

	// Used Vehicle By UVC
	uvcRes, err := bs.GetUVCInfo(vd.UVC, state)
	if err != nil {
		return nil, err
	}

	// Vehicle Colors
	colorsRes, err := bs.GetColors(vd.UVC) // doesn't use state
	if err != nil {
		return nil, err
	}

	// Collect the results and return it
	result := &BlackbookVINSummary{
		VIN:          vinRes,
		UniversalVIN: uVINRes,
		LikeVehicles: likeVehRes,
		Batch:        batchRes,
		Drilldown:    drilldownRes,
		UVC:          uvcRes,
		Colors:       colorsRes,
	}
	return result, nil
}

func (bs *blackbookAPIService) GetSummaryByPlate(plate, state string) (*BlackbookVINSummary, error) {
	result := new(BlackbookVINSummary)

	// Used Vehicle By License Plate
	plateRes, err := bs.GetPlate(plate, state)
	if err != nil {
		return nil, err
	}

	result.Plate = plateRes

	return result, nil
}

type BlackbookVINSummary struct {
	LikeVehicles []byte
	UniversalVIN []byte
	Batch        []byte
	Drilldown    []byte
	Plate        []byte
	UVC          []byte
	VIN          []byte
	Colors       []byte
}

// todo add tests to this
func (bs *blackbookAPIService) executeAPI(httpClient shared.HTTPClientWrapper, path string, payload []byte) (body []byte, err error) {
	var res *http.Response
	// If a payload is present, make it a POST request
	if payload != nil {
		res, err = httpClient.ExecuteRequest(path, "POST", payload)
	} else {
		res, err = httpClient.ExecuteRequest(path, "GET", nil)
	}
	if res == nil {
		if err != nil {
			return nil, errors.Wrapf(err, "error calling Blackbook api => %s", path)
		}
		return nil, fmt.Errorf("received error with no response when calling to %s", path)
	}

	if err != nil {
		switch res.StatusCode {
		case 401: // Expected when Blackbook API subscription is expired or disabled
			return nil, errors.Wrapf(err, "error 401 'Unauthorized' while calling Blackbook api => %s", path)
		case 404: // Not expected at all
			return nil, errors.Wrapf(err, "error 404 'Not Found' while calling Blackbook api => %s", path)
		}
		return nil, errors.Wrapf(err, "error calling Blackbook api => %s", path)
	}
	defer func() {
		err = res.Body.Close()
		if err != nil {
			err = errors.Wrapf(err, "error closing response body => %s", path)
		}
	}()

	// return the response body
	body, err = io.ReadAll(res.Body)
	return body, err
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

/*
---------------------------------------------------------- REFERENCE ----------------------------------------------------------
	Endpoint Name                    Endpoint Path                         Notes
-------------------------------------------------------------------------------------------------------------------------------
	"Drilldown Information"          "/Drilldown/{vehicleclass}/{year}"    Narrowed by YMMS (Year/Make/Model/Series)
	"Drilldown Information w/CPI"    "/CPIUsedDrilldown/{year}"            Narrowed by YMMS (Year/Make/Model/Series)
	"Expire Token"                   "/Token/Expire"                       N/A
	"Get Token"                      "/Token/Get"                          N/A
	"GraphQL"                        "/GraphQL"                            N/A
	"Like Vehicles"                  "/UsedVehicle/Like/{uvc}"             By UVC
	"Publish Dates"                  "/PublishDate/get"                    Error "This account does not have permission to access historical data."
	"Standard Equipment"             "/StdEquip/{uvc}"                     Error "Access Denied", By UVC
	"State"                          "/State"                              N/A
	"Universal Lookup By VIN"        "/Universal/VIN/{vin}"                By VIN
	"Used Vehicle Batch"             "/Batch"                              By VIN/UVC, JSON payload, use POST method
	"Used Vehicle by Chrome ID"      "/UsedVehicle/Chrome/{chromeid}"      Error "You don't have permission to retrieve by chrome Id - Contact Sales for more information"
	"Used Vehicle By Drilldown"      "/UsedVehicle/{year}/{make}"          By YMMSS (Year/Make/Model/Series/Style)
	"Used Vehicle By License Plate"  "/UsedVehicle/plate/{plate}/{state}"  By Plate and State (State = State Name, State Abbreviation, or Zip Code)
	"Used Vehicle By UVC"            "/UsedVehicle/UVC/{uvc}"              By UVC
	"Used Vehicle By VIN"            "/UsedVehicle/VIN/{vin}"              By VIN
	"Vehicle Colors"                 "/Colors/{uvc}"                       By UVC
	"Vehicle Lookup Autocomplete"    "/Autocomplete"                       By search text
	"Vehicle Photo"                  "/Photos/{uvc}"                       Error "Access Denied"
	"Vehicle Specs PDF File"         "/PDFSpecs/{uvc}"                     Error "Access Denied"
	"Your IP address"                "/IPAddress"                          N/A
-------------------------------------------------------------------------------------------------------------------------------
*/
