package main

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/DIMO-INC/devices-api/internal/config"
	"github.com/DIMO-INC/devices-api/internal/database"
	"github.com/DIMO-INC/devices-api/models"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

//go:embed mmy_definitions.csv
var mmyDefinitions string

func loadMMYCSVData(ctx context.Context, logger zerolog.Logger, settings *config.Settings, pdb database.DbStore) {
	// check db ready
	time.Sleep(time.Second * 3)

	csvReader := csv.NewReader(strings.NewReader(mmyDefinitions))
	for {
		rec, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Fatal().Err(err)
		}
		// do something with read line
		fmt.Printf("%+v\n", rec)
		yrInt, err := strconv.Atoi(rec[2])
		if err != nil {
			logger.Info().Err(err).Msg("can't parse year: " + rec[2])
			continue
		}
		dd, err := models.DeviceDefinitions(
			qm.Where("make = ?", strings.ToUpper(rec[0])),
			qm.And("model = ?", strings.ToUpper(rec[1])),
			qm.And("year = ?", yrInt)).One(ctx, pdb.DBS().Writer)
		if err != nil && err != sql.ErrNoRows {
			logger.Fatal().Err(err).Msg("can't read existing definition")
		}
		if dd != nil {
			fmt.Printf(" ignoring, already exists: %s", dd.ID)
			continue
		}

		dd = &models.DeviceDefinition{
			ID:       ksuid.New().String(),
			Make:     strings.ToUpper(rec[0]),
			Model:    strings.ToUpper(rec[1]),
			Year:     int16(yrInt),
			Source:   null.StringFrom("csv import"),
			Verified: true,
		}
		// insert
		err = dd.Insert(ctx, pdb.DBS().Writer, boil.Infer())
		if err != nil {
			logger.Fatal().Err(err).Msg("can't insert new definition")
		}
	}
}
