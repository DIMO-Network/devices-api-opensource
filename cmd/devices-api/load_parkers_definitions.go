package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

const parkersSource = "parkers"

type ManufacturersResponse struct {
	Manufacturers []struct {
		Name   string `json:"name"`
		Key    string `json:"key"`
		Ranges []struct {
			Name string `json:"name"`
			Key  string `json:"key"`
			URL  string `json:"url"`
		} `json:"ranges"`
	} `json:"manufacturers"`
}

const baseURL = "https://www.parkers.co.uk"
const minYear = 2000

type IntSet struct {
	elements map[int]struct{}
}

func NewIntSet() *IntSet {
	return &IntSet{elements: make(map[int]struct{})}
}

func (s *IntSet) Add(i int) {
	s.elements[i] = struct{}{}
}

func (s *IntSet) Contains(i int) bool {
	_, ok := s.elements[i]
	return ok
}

func (s *IntSet) Slice() []int {
	out := make([]int, 0, len(s.elements))
	for i := range s.elements {
		out = append(out, i)
	}
	return out
}

func (s *IntSet) Len() int {
	return len(s.elements)
}

func get(url string, processBody func(io.Reader) error) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("status code %d", resp.StatusCode)
	}
	return processBody(resp.Body)
}

// // Needed for version years
// var monthYearRegexp = regexp.MustCompile(`^(?:January|February|March|April|May|June|July|August|September|October|November|December) (\d{4})`)

var yearsRegexp = regexp.MustCompile(`(\d{4}) (?:onwards|- (\d{4}))\) Specifications$`)

func loadParkersDeviceDefinitions(ctx context.Context, logger *zerolog.Logger, settings *config.Settings, pdb database.DbStore) error {
	manufacturersURL := baseURL + "/api/cars/quick-find/specs/"
	manufacturersBody := new(ManufacturersResponse)
	if err := get(manufacturersURL, makeDecoder(manufacturersBody)); err != nil {
		return fmt.Errorf("failed to retrieve manufacturers: %v", err)
	}

	db := pdb.DBS().Writer

	for _, manufacturer := range manufacturersBody.Manufacturers {
		dbMake, err := models.DeviceMakes(models.DeviceMakeWhere.Name.EQ(manufacturer.Name)).One(ctx, db)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("error retrieving existing make with name %q: %w", manufacturer.Name, err)
			} else {
				dbMake = &models.DeviceMake{
					ID:   ksuid.New().String(),
					Name: manufacturer.Name,
				}
			}
		}

		externalIDs := make(map[string]string)
		if dbMake.ExternalIds.Valid {
			if err := json.Unmarshal(dbMake.ExternalIds.JSON, &externalIDs); err != nil {
				logger.Warn().Err(err).Msgf("Failed to load existing external IDs from make %s, overwriting", dbMake.ID)
			}
		}

		externalIDs[parkersSource] = manufacturer.Key

		externalIDsBytes, err := json.Marshal(externalIDs)
		if err != nil {
			return fmt.Errorf("failed to serialize external IDs")
		}

		dbMake.ExternalIds = null.JSONFrom(externalIDsBytes)
		if err := dbMake.Upsert(ctx, db, true, []string{models.DeviceMakeColumns.ID}, boil.Infer(), boil.Infer()); err != nil {
			return fmt.Errorf("failed upserting make %s", manufacturer.Name)
		}

		for _, mfrRange := range manufacturer.Ranges {
			rangeURL := baseURL + mfrRange.URL

			var rangeDoc *goquery.Document
			if err := get(rangeURL, makeDoc(&rangeDoc)); err != nil {
				logger.Err(err).Msgf("Failed to retrieve range page %s, skipping", mfrRange.URL)
				continue
			}

			years := NewIntSet()

			rangeDoc.Find("a.panel__primary-link").Each(func(i int, s *goquery.Selection) {
				match := yearsRegexp.FindStringSubmatch(s.Text())
				if match == nil {
					logger.Err(err).Msgf("Unexpected model text %q on %s", s.Text(), mfrRange.URL)
					return
				}

				startYear, err := strconv.Atoi(match[1])
				if err != nil {
					// This simply should not happen after the regex matches.
					logger.Err(err).Msgf("Failed to parse year string %q into int", match[1])
					return
				}

				// Trying to not use anything before 2000.
				if startYear < minYear {
					startYear = minYear
				}

				endYear := time.Now().Year()
				if match[2] != "" {
					var err error
					endYear, err = strconv.Atoi(match[2])
					if err != nil {
						// This simply should not happen after the regex matches.
						logger.Err(err).Msgf("Failed to parse year string %q into int", match[2])
						return
					}
				}

				for year := startYear; year <= endYear; year++ {
					years.Add(year)
				}
			})

			for _, year := range years.Slice() {
				dd, err := models.DeviceDefinitions(
					models.DeviceDefinitionWhere.DeviceMakeID.EQ(dbMake.ID),
					models.DeviceDefinitionWhere.Model.EQ(mfrRange.Name),
					models.DeviceDefinitionWhere.Year.EQ(int16(year)),
				).One(ctx, db)
				if err != nil {
					if !errors.Is(err, sql.ErrNoRows) {
						return fmt.Errorf("failed retrieving existing device definition: %w", err)
					}
					dd = &models.DeviceDefinition{
						ID:           ksuid.New().String(),
						DeviceMakeID: dbMake.ID,
						Model:        mfrRange.Name,
						Year:         int16(year),
					}
				}
				if !dd.Source.Valid {
					dd.Source.SetValid(parkersSource)
					dd.ExternalID.SetValid(manufacturer.Key + "/" + mfrRange.Key)
				}
				dd.Verified = true

				if err := dd.Upsert(ctx, db, true, []string{models.DeviceDefinitionColumns.ID}, boil.Infer(), boil.Infer()); err != nil {
					return fmt.Errorf("failed to upsert device definition %s: %w", dd.ID, err)
				}
			}
		}
	}

	return nil
}

func makeDecoder(out interface{}) func(io.Reader) error {
	return func(body io.Reader) error {
		return json.NewDecoder(body).Decode(out)
	}
}

func makeDoc(out **goquery.Document) func(io.Reader) error {
	return func(body io.Reader) error {
		var err error
		*out, err = goquery.NewDocumentFromReader(body)
		return err
	}
}
