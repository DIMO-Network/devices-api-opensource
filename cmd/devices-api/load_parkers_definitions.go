package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
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

type Manufacturer struct {
	Name   string `json:"name"`
	Key    string `json:"key"`
	Ranges []struct {
		Name string `json:"name"`
		Key  string `json:"key"`
		URL  string `json:"url"`
	} `json:"ranges"`
}
type ManufacturersResponse struct {
	Manufacturers []Manufacturer `json:"manufacturers"`
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

var httpClient = http.Client{
	Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   20,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	},
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

// Needed for version years
var monthYearRegexp = regexp.MustCompile(`^(?:January|February|March|April|May|June|July|August|September|October|November|December) (\d{4})`)

var yearsRegexp = regexp.MustCompile(`(\d{4}) (?:onwards|- (\d{4}))\) Specifications$`)

func loadParkersDeviceDefinitions(ctx context.Context, logger *zerolog.Logger, settings *config.Settings, pdb database.DbStore) error {
	var numRanges uint64
	var numRangesProcessed uint64

	logger.Info().Msg("Loading device definitions from Parkers")
	manufacturersURL := baseURL + "/api/cars/quick-find/specs/"
	manufacturersBody := new(ManufacturersResponse)
	if err := get(manufacturersURL, makeDecoder(manufacturersBody)); err != nil {
		return fmt.Errorf("failed to retrieve manufacturers: %v", err)
	}

	db := pdb.DBS().Writer

	var wg sync.WaitGroup

	for _, manufacturer := range manufacturersBody.Manufacturers {
		wg.Add(1)
		go func(manufacturer Manufacturer) {
			atomic.AddUint64(&numRanges, uint64(len(manufacturer.Ranges)))
			dbMake, err := models.DeviceMakes(models.DeviceMakeWhere.Name.EQ(manufacturer.Name)).One(ctx, db)
			if err != nil {
				if !errors.Is(err, sql.ErrNoRows) {
					logger.Err(err).Msgf("Failed searching for make with name %q", manufacturer.Name)
					return
				} else {
					dbMake = &models.DeviceMake{
						ID:   ksuid.New().String(),
						Name: manufacturer.Name,
					}
					logger.Debug().Msgf("Creating make %s", manufacturer.Name)
				}
			} else {
				logger.Debug().Msgf("Found make %s", manufacturer.Name)
			}

			externalIDs := make(map[string]string)
			if dbMake.ExternalIds.Valid {
				if err := json.Unmarshal(dbMake.ExternalIds.JSON, &externalIDs); err != nil {
					logger.Warn().Err(err).Msgf("Failed to load existing external IDs from make %s, overwriting", dbMake.ID)
					externalIDs = make(map[string]string)
				}
			}

			externalIDs[parkersSource] = manufacturer.Key

			externalIDsBytes, err := json.Marshal(externalIDs)
			if err != nil {
				logger.Err(err).Msgf("Failed to serialize externalID map: %w", err)
			}

			dbMake.ExternalIds = null.JSONFrom(externalIDsBytes)
			if err := dbMake.Upsert(ctx, db, true, []string{models.DeviceMakeColumns.ID}, boil.Infer(), boil.Infer()); err != nil {
				logger.Err(err).Msgf("Failed upserting make %s", manufacturer.Name)
				return
			}

			for _, mfrRange := range manufacturer.Ranges {
				rangeURL := baseURL + mfrRange.URL

				var rangeDoc *goquery.Document
				if err := get(rangeURL, makeDoc(&rangeDoc)); err != nil {
					logger.Err(err).Msgf("Failed to retrieve range page %s, skipping", mfrRange.URL)
					continue
				}

				years := NewIntSet()

				modelLinks := make([]string, 0)

				rangeDoc.Find("a.panel__primary-link").Each(func(i int, s *goquery.Selection) {
					modelLink, modelLinkExists := s.Attr("href")
					if !modelLinkExists {
						logger.Warn().Msgf("No link for model %s, odd", s.Text())
					}

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

					if modelLinkExists && endYear >= minYear {
						modelLinks = append(modelLinks, modelLink)
					}

					for year := startYear; year <= endYear; year++ {
						years.Add(year)
					}
				})

				dds := make(map[int]*models.DeviceDefinition)

				for _, year := range years.Slice() {
					dd, err := models.DeviceDefinitions(
						models.DeviceDefinitionWhere.DeviceMakeID.EQ(dbMake.ID),
						models.DeviceDefinitionWhere.Model.EQ(mfrRange.Name),
						models.DeviceDefinitionWhere.Year.EQ(int16(year)),
					).One(ctx, db)
					if err != nil {
						if !errors.Is(err, sql.ErrNoRows) {
							logger.Err(err).Msgf("Failed searching for existing device definition")
							return
						}
						dd = &models.DeviceDefinition{
							ID:           ksuid.New().String(),
							DeviceMakeID: dbMake.ID,
							Model:        mfrRange.Name,
							Year:         int16(year),
						}
						logger.Debug().Msgf("Creating device definition for %s %s %d", manufacturer.Name, mfrRange.Name, year)
					}
					if !dd.Source.Valid {
						dd.Source.SetValid(parkersSource)
						dd.ExternalID.SetValid(manufacturer.Key + "/" + mfrRange.Key)
					}
					dd.Verified = true

					if err := dd.Upsert(ctx, db, true, []string{models.DeviceDefinitionColumns.ID}, boil.Infer(), boil.Infer()); err != nil {
						logger.Err(err).Msgf("Failed upserting device definition")
						return
					}

					dds[year] = dd
				}

				for _, modelLink := range modelLinks {
					var modelDoc *goquery.Document
					if err := get(baseURL+modelLink, makeDoc(&modelDoc)); err != nil {
						logger.Err(err).Msgf("Failed to retrieve model page %s, skipping", mfrRange.URL)
						continue
					}

					modelDoc.Find("select.trim-equipment-list__filter").First().Find("option").Each(func(i int, s *goquery.Selection) {
						val, exists := s.Attr("value")
						if !exists {
							logger.Warn().Msgf("Trim selection index %d has no value attribute", i)
						}
						if val == "placeholder" {
							return
						}
						trimName := s.Text()
						versionSelector := fmt.Sprintf(`ul[data-derivative-id^="%s-engine_"]`, val)
						modelDoc.Find(versionSelector).Find("li").Each(func(i int, s *goquery.Selection) {
							versionName := s.Text()
							versionID, exists := s.Attr("value")
							if !exists {
								logger.Warn().Msgf("Version name has no value attribute")
								return
							}
							versionLinkSelector := fmt.Sprintf(`div[data-derivative-link-id="%s"]`, versionID)
							link, exists := modelDoc.Find(versionLinkSelector).Find("a").First().Attr("href")
							if !exists {
								logger.Warn().Msgf("Version has no associated link")
								return
							}

							// Sometimes they don't URL-encode "#1" in names.
							safeLink := strings.Replace(link, "#", "%23", -1)

							var versionDoc *goquery.Document
							if err := get(baseURL+safeLink, makeDoc(&versionDoc)); err != nil {
								logger.Warn().Msgf("Couldn't retrieve version doc")
								return
							}

							from := strings.TrimSpace(versionDoc.Find("span.specs-detail-page__available-dates__from").First().Text())

							match := monthYearRegexp.FindStringSubmatch(from)
							if match == nil {
								logger.Warn().Err(err).Msgf("From date not in the expected format")
								return
							}
							startYear, err := strconv.Atoi(match[1])
							if err != nil {
								logger.Warn().Err(err).Msgf("From date not in the expected format")
								return
							}

							if startYear < minYear {
								startYear = minYear
							}

							to := strings.TrimSpace(versionDoc.Find("span.specs-detail-page__available-dates__to").First().Text())
							endYear := 2022
							if to != "Now" {
								match := monthYearRegexp.FindStringSubmatch(to)
								if match == nil {
									logger.Warn().Err(err).Msgf("To date not in the expected format")
									return
								}
								endYear, _ = strconv.Atoi(match[1])
							}

							for year := startYear; year <= endYear; year++ {
								dd, ok := dds[year]
								if !ok {
									logger.Warn().Err(err).Msgf("Version year not in the computed year list for the range")
									continue
								}
								ds, err := models.DeviceStyles(
									models.DeviceStyleWhere.DeviceDefinitionID.EQ(dd.ID),
									models.DeviceStyleWhere.Name.EQ(versionName),
									models.DeviceStyleWhere.SubModel.EQ(trimName),
								).One(ctx, db)
								if err != nil {
									if !errors.Is(err, sql.ErrNoRows) {
										logger.Warn().Err(err).Msgf("Failed to look up styles")
										return
									}
									ds = &models.DeviceStyle{
										ID:                 ksuid.New().String(),
										DeviceDefinitionID: dd.ID,
										Name:               versionName,
										SubModel:           trimName,
									}
								}
								if ds.Source == "" {
									ds.Source = parkersSource
									ds.ExternalStyleID = versionID
								}
								if err := ds.Upsert(ctx, db, true, []string{models.DeviceStyleColumns.ID}, boil.Infer(), boil.Infer()); err != nil {
									logger.Err(err).Msgf("Failed to upsert styles")
									return
								}
							}
						})

					})
				}

				atomic.AddUint64(&numRangesProcessed, 1)
			}
			wg.Done()
		}(manufacturer)
	}

	done := make(chan struct{})

	go func() {
		tick := time.NewTicker(10 * time.Second)
		for {
			select {
			case <-tick.C:
				logger.Info().Msgf("Processed %d/%d makes", numRangesProcessed, numRanges)
			case <-done:
				tick.Stop()
				return
			}
		}
	}()

	wg.Wait()
	done <- struct{}{}

	logger.Info().Msg("Finished syncing with Parkers")

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
