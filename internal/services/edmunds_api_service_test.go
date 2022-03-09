package services

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/rs/zerolog"
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
	es := &EdmundsService{
		baseMediaURL: "https://something.com/gateway",
	}
	imageURL := es.findImageURLByMaxWidth(photosFull.Photos[0].Sources, 900)

	assert.Equal(t, "https://something.com/gateway"+photosFull.Photos[0].Sources[4].Link.Href, *imageURL)

	var emptySources []photosResponseSource
	noImage := es.findImageURLByMaxWidth(emptySources, 900)
	assert.Nil(t, noImage)
}

func TestEdmundsService_GetDefaultImageForMMY_happyPath(t *testing.T) {
	es := &EdmundsService{
		baseMediaURL: "https://something.com/gateway",
		baseAPIURL:   "https://test-api.com",
	}
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	const m = "honda"
	const model = "civic"
	const year = 2020
	url := fmt.Sprintf("%s/api/media/v2/%s/%s/%d/photos?format=json&pageSize=50", es.baseAPIURL, m, model, year)
	httpmock.RegisterResponder(http.MethodGet, url, httpmock.NewStringResponder(200, testEdmundsPhotosJSON))

	imageURL, err := es.GetDefaultImageForMMY(m, model, year)
	assert.NoError(t, err)
	assert.Equal(t, "https://something.com/gateway/honda/civic/2016/oem/2016_honda_civic_coupe_touring_fq_oem_17_815.jpg", *imageURL)
}

func TestEdmundsService_buildAndExecuteRequest(t *testing.T) {
	log := zerolog.New(os.Stdout)
	es := &EdmundsService{
		baseMediaURL: "https://something.com/gateway",
		baseAPIURL:   "https://test-api.com",
		log:          &log,
	}
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	url := fmt.Sprintf("%s/testing", es.baseAPIURL)
	httpmock.RegisterResponder(http.MethodGet, url, httpmock.NewStringResponder(409, "error: too many requests"))

	response, err := es.buildAndExecuteRequest(url)

	assert.Error(t, err, "expected error")
	assert.Contains(t, err.Error(), "all retries failed")
	assert.Nil(t, response)
}
