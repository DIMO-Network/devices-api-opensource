package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type Query struct {
	OperationName string                 `json:"operationName"`
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables"`
}

type ADACClient struct {
	HTTPClient *http.Client
}

func (c *ADACClient) Query(q *Query, body interface{}) error {
	reqBody, err := json.Marshal(q)
	if err != nil {
		return err
	}
	sum := md5.Sum(reqBody)
	hash := hex.EncodeToString(sum[:])

	req, err := http.NewRequest("POST", "https://www.adac.de/bff", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GraphQL-Query-Hash", hash)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(body); err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code %d", resp.StatusCode)
	}

	return nil
}

type BrandsResponse struct {
	Data struct {
		CarSearchBrandFacets []Facet `json:"carSearchBrandFacets"`
	} `json:"data"`
}

type Facet struct {
	FacetID int    `json:"facetId"`
	Label   string `json:"label"`
}

type ADACRangesResponse struct {
	Data struct {
		CarSearchRangeFacets []Facet `json:"carSearchRangeFacets"`
	} `json:"data"`
}

type SearchResponse struct {
	Data struct {
		CarSearch struct {
			Items []struct {
				Cars []struct {
					BrandSlug      string `json:"brandSlug"`
					RangeSlug      string `json:"rangeSlug"`
					GenerationSlug string `json:"generationSlug"`
					ID             string `json:"id"`
				}
			} `json:"items"`
			NextPage *string `json:"nextPage"`
		} `json:"carSearch"`
	} `json:"data"`
}

type CarPageResponse struct {
	Data struct {
		Page struct {
			Result struct {
				CarPage struct {
					TechnicalData []struct {
						Name string `json:"name"`
						Data []struct {
							Name  string `json:"name"`
							Value string `json:"value"`
						} `json:"data"`
					} `json:"technicalData"`
				} `json:"carPage"`
			} `json:"result"`
		} `json:"page"`
	} `json:"data"`
}

type StyleCollector struct {
	SubModel, Style    string
	StartYear, EndYear int
}

func parseMonth(s string) (time.Month, int, error) {
	if len(s) != 5 {
		return 0, 0, fmt.Errorf("string is length %d, not 5", len(s))
	}
	m, err := strconv.Atoi(s[:2])
	if err != nil {
		return 0, 0, fmt.Errorf("failed parsing month string %q: %w", s[:2], err)
	}
	if m < 1 || m > 12 {
		return 0, 0, fmt.Errorf("invalid month number %d", m)
	}
	y, err := strconv.Atoi(s[3:])
	if err != nil {
		return 0, 0, fmt.Errorf("failed parsing month string %q: %w", s[:2], err)
	}

	if y < 30 {
		return time.Month(m), 2000 + y, nil
	}
	return time.Month(m), 1900 + y, nil
}

func (c *StyleCollector) Collect(name, value string) error {
	switch name {
	case "Modell":
		c.SubModel = html.UnescapeString(value)
	case "Typ":
		c.Style = html.UnescapeString(value)
	case "Modellstart":
		m, y, err := parseMonth(value)
		if err != nil {
			return fmt.Errorf("failed to parse start month: %w", err)
		}
		if m <= time.June {
			y++
		}
		if y < minYear {
			y = minYear
		}
		c.StartYear = y
	case "Modellende":
		if value == "" {
			c.EndYear = modelYear(time.Now())
		} else {
			m, y, err := parseMonth(value)
			if err != nil {
				return fmt.Errorf("failed to parse end month: %w", err)
			}
			if m <= time.June {
				y++
			}
			c.EndYear = y
		}
	}
	return nil
}

func (c *StyleCollector) Validate() error {
	if c.SubModel == "" {
		return errors.New("no sub-model (Modell) found")
	}
	if c.StartYear == 0 {
		return errors.New("no manufacturing start (Modellstart) found")
	}
	if c.EndYear == 0 {
		return errors.New("no manufacturing end (Modellende) found")
	}
	return nil
}

func modelYear(t time.Time) int {
	if t.Month() <= time.June {
		return t.Year()
	}
	return t.Year() + 1
}

const adacSource = "adac"

func loadADACDeviceDefinitions(ctx context.Context, logger *zerolog.Logger, settings *config.Settings, pdb database.DbStore) error {
	client := &ADACClient{
		HTTPClient: &http.Client{},
	}
	db := pdb.DBS().Writer
	brandsQuery := Query{
		OperationName: "CarSearchBrandFacets",
		Query: `
			query CarSearchBrandFacets {
				carSearchBrandFacets {
					facetId
					label
				}
		  	}`,
		Variables: map[string]interface{}{},
	}
	brandsResponse := new(BrandsResponse)
	if err := client.Query(&brandsQuery, brandsResponse); err != nil {
		return err
	}
	for _, brand := range brandsResponse.Data.CarSearchBrandFacets {
		if brand.Label != "Opel" {
			// For now.
			continue
		}

		var (
			numStyles, numStylesProcessed uint64
		)
		var wg sync.WaitGroup

		dbMake, err := models.DeviceMakes(models.DeviceMakeWhere.Name.EQ(brand.Label)).One(ctx, db)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("failed searching for make with name %q", brand.Label)
			}
			dbMake = &models.DeviceMake{
				ID:   ksuid.New().String(),
				Name: brand.Label,
			}
			logger.Info().Msgf("Creating make %s", brand.Label)
		} else {
			logger.Info().Msgf("Found make %s", brand.Label)
		}

		externalIDs := make(map[string]string)
		if dbMake.ExternalIds.Valid {
			if err := json.Unmarshal(dbMake.ExternalIds.JSON, &externalIDs); err != nil {
				logger.Warn().Err(err).Msgf("Failed to load existing external IDs from make %s, overwriting", dbMake.ID)
				externalIDs = make(map[string]string)
			}
		}

		externalIDs[adacSource] = strconv.Itoa(brand.FacetID)

		externalIDsBytes, err := json.Marshal(externalIDs)
		if err != nil {
			logger.Err(err).Msgf("Failed to serialize externalID map: %v", err)
		}

		dbMake.ExternalIds = null.JSONFrom(externalIDsBytes)
		if err := dbMake.Upsert(ctx, db, true, []string{models.DeviceMakeColumns.ID}, boil.Infer(), boil.Infer()); err != nil {
			return fmt.Errorf("failed upserting make %s", brand.Label)
		}

		rangesQuery := Query{
			OperationName: "CarSearchRangeFacets",
			Query: `
				query CarSearchRangeFacets($brandId: String!) {
					carSearchRangeFacets(brandId: $brandId) {
						facetId
						label
					}
				}`,
			Variables: map[string]interface{}{"brandId": strconv.Itoa(brand.FacetID)},
		}
		rangesResponse := new(ADACRangesResponse)
		if err := client.Query(&rangesQuery, rangesResponse); err != nil {
			return err
		}
		for _, carRange := range rangesResponse.Data.CarSearchRangeFacets {
			wg.Add(1)
			go func(carRange Facet) {
				defer func() { wg.Done() }()
				ddCache := make(map[int]*models.DeviceDefinition)
				getOrCreateDeviceDefinition := func(year int) (*models.DeviceDefinition, error) {
					dd, ok := ddCache[year]
					if ok {
						return dd, nil
					}

					dd, err := models.DeviceDefinitions(
						models.DeviceDefinitionWhere.DeviceMakeID.EQ(dbMake.ID),
						models.DeviceDefinitionWhere.Model.EQ(carRange.Label),
						models.DeviceDefinitionWhere.Year.EQ(int16(year)),
					).One(ctx, db)
					if err != nil {
						if !errors.Is(err, sql.ErrNoRows) {
							return nil, err
						}

						dd = &models.DeviceDefinition{
							ID:           ksuid.New().String(),
							DeviceMakeID: dbMake.ID,
							Model:        carRange.Label,
							Year:         int16(year),
						}
					}

					if !dd.Source.Valid {
						dd.Source.SetValid(adacSource)
						dd.ExternalID.SetValid(strconv.Itoa(carRange.FacetID))
					}
					dd.Verified = true

					if err := dd.Upsert(ctx, db, true, []string{models.DeviceDefinitionColumns.ID}, boil.Infer(), boil.Infer()); err != nil {
						logger.Err(err).Msgf("Failed upserting device definition")
						return nil, err
					}

					ddCache[year] = dd
					return dd, nil
				}

				var nextPage *string
				for {
					searchQuery := Query{
						OperationName: "CarSearch",
						Query: `
					query CarSearch($filters: CarSearchFilters!, $page: ID) {
						carSearch(filters: $filters, page: $page) {
							items {
								cars {
									brandSlug
									rangeSlug
									generationSlug
									id
								}
							}
							nextPage
						}
					}`,
						Variables: map[string]interface{}{
							"filters": map[string]interface{}{
								"brandId": strconv.Itoa(brand.FacetID),
								"rangeId": strconv.Itoa(carRange.FacetID),
							},
							"nextPage": nextPage,
						},
					}
					searchResponse := new(SearchResponse)
					if err := client.Query(&searchQuery, searchResponse); err != nil {
						logger.Warn().Err(err).Msgf("Failed running query")
						return
					}
					for _, variant := range searchResponse.Data.CarSearch.Items {
						for _, car := range variant.Cars {
							func() {
								atomic.AddUint64(&numStyles, 1)
								defer func() { atomic.AddUint64(&numStylesProcessed, 1) }()
								path := fmt.Sprintf("/rund-ums-fahrzeug/autokatalog/marken-modelle/%s/%s/%s/%s/", car.BrandSlug, car.RangeSlug, car.GenerationSlug, car.ID)
								pageQuery := Query{
									OperationName: "ResolvePage",
									Query: `
										query ResolvePage($path: String!) {
											page(path: $path) {
												result {
													... on CarPage {
														carPage {
															technicalData {
																name
																data {
																	name
																	value
																}
															}
														}
													}
												}
											}
										}`,
									Variables: map[string]interface{}{"path": path},
								}
								pageResponse := new(CarPageResponse)
								if err := client.Query(&pageQuery, pageResponse); err != nil {
									logger.Warn().Err(err).Msgf("Failed to retrieve page")
									return
								}

								coll := StyleCollector{}
								for _, section := range pageResponse.Data.Page.Result.CarPage.TechnicalData {
									if section.Name == "Allgemein" {
										for _, datum := range section.Data {
											if err := coll.Collect(datum.Name, datum.Value); err != nil {
												logger.Warn().Err(err).Msgf("Failed to collect style attributes")
												return
											}
										}
									}
								}

								if err := coll.Validate(); err != nil {
									return
								}

								for year := coll.StartYear; year < coll.EndYear; year++ {
									dd, err := getOrCreateDeviceDefinition(year)
									if err != nil {
										logger.Warn().Err(err).Msgf("Failed on device definition lookup")
										return
									}
									ds, err := models.DeviceStyles(
										models.DeviceStyleWhere.DeviceDefinitionID.EQ(dd.ID),
										models.DeviceStyleWhere.Name.EQ(coll.Style),
										models.DeviceStyleWhere.SubModel.EQ(coll.SubModel),
									).One(ctx, db)
									if err != nil {
										if !errors.Is(err, sql.ErrNoRows) {
											logger.Warn().Err(err).Msgf("Failed to look up styles")
											return
										}
										ds = &models.DeviceStyle{
											ID:                 ksuid.New().String(),
											DeviceDefinitionID: dd.ID,
											Name:               coll.Style,
											SubModel:           coll.SubModel,
										}
									}
									if ds.Source == "" {
										ds.Source = adacSource
										ds.ExternalStyleID = car.ID
									}
									if err := ds.Upsert(ctx, db, true, []string{models.DeviceStyleColumns.ID}, boil.Infer(), boil.Infer()); err != nil {
										logger.Err(err).Msgf("Failed to upsert styles")
										return
									}
								}
							}()
						}
					}
					nextPage = searchResponse.Data.CarSearch.NextPage
					if nextPage == nil {
						break
					}
				}
			}(carRange)
		}

		done := make(chan struct{})

		go func() {
			tick := time.NewTicker(10 * time.Second)
			for {
				select {
				case <-tick.C:
					logger.Info().Msgf("Processed %d/%d styles", numStylesProcessed, numStyles)
				case <-done:
					tick.Stop()
					return
				}
			}
		}()

		wg.Wait()
		done <- struct{}{}
	}
	return nil
}
