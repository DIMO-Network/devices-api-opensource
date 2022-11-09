package services

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/shared"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

//go:embed test_edmunds_photo_api.json
var testEdmundsPhotosJSON string

func Test_findPhotoByShotType(t *testing.T) {
	photosFull := photosResponse{}
	err := json.Unmarshal([]byte(testEdmundsPhotosJSON), &photosFull)
	assert.NoError(t, err)

	found := findPhotoByShotType(&photosFull, "FQ")

	assert.NotNilf(t, found, "expected to find images")
	assert.Len(t, found, 18)

	notFound := findPhotoByShotType(&photosFull, "ABCDEF")
	assert.Len(t, notFound, 0)
}

func TestEdmundsService_findImageURLByMaxWidth(t *testing.T) {
	photosFull := photosResponse{}
	err := json.Unmarshal([]byte(testEdmundsPhotosJSON), &photosFull)
	assert.NoError(t, err)
	es := NewEdmundsService("", test.Logger())
	imageURL := es.findImageURLByMaxWidth(photosFull.Photos[0].Sources, 900)

	assert.Equal(t, edmundsBaseMediaURL+photosFull.Photos[0].Sources[4].Link.Href, *imageURL)

	var emptySources []photosResponseSource
	noImage := es.findImageURLByMaxWidth(emptySources, 900)
	assert.Nil(t, noImage)
}

func TestEdmundsService_GetDefaultImageForMMY_happyPath(t *testing.T) {
	es := NewEdmundsService("", test.Logger())
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	const m = "honda"
	const model = "civic"
	const year = 2020
	url := fmt.Sprintf("%s/api/media/v2/%s/%s/%d/photos?format=json&pageSize=50", edmundsAPIURL, m, model, year)
	httpmock.RegisterResponder(http.MethodGet, url, httpmock.NewStringResponder(200, testEdmundsPhotosJSON))

	imageURL, err := es.GetDefaultImageForMMY(m, model, year)
	assert.NoError(t, err)
	assert.Equal(t, edmundsBaseMediaURL+"/honda/civic/2016/oem/2016_honda_civic_coupe_touring_fq_oem_17_815.jpg", *imageURL)
}

func TestEdmundsService_retriesWillFail(t *testing.T) {
	es := NewEdmundsService("", test.Logger())
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	url := fmt.Sprintf("%s/api/vehicle/v2/makes", edmundsAPIURL)
	httpmock.RegisterResponder(http.MethodGet, url, httpmock.NewStringResponder(409, "error: too many requests"))

	response, err := es.getAllMakes()

	assert.Error(t, err, "expected error")
	assert.ErrorIs(t, err, err.(shared.HTTPResponseError), "received a non 200 response from edmunds. status code: 409")
	assert.Nil(t, response)
}
