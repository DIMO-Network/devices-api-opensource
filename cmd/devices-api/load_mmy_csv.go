package main

import (
	"context"
	_ "embed"
	"encoding/csv"
	"fmt"
	"github.com/DIMO-INC/devices-api/internal/config"
	"github.com/DIMO-INC/devices-api/internal/database"
	"github.com/DIMO-INC/devices-api/models"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"io"
	"strconv"
	"strings"
)

//go:embed mmy_definitions.csv
var mmyDefinitions string

func loadMMYCSVData(ctx context.Context, logger zerolog.Logger, settings *config.Settings, pdb database.DbStore) {
	// csv loader
	// for each
	// check for duplicate in DB, insert, if duplicate log
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
			logger.Fatal().Err(err).Msg("can't parse year")
		}
		dd, err := models.DeviceDefinitions(
			qm.Where("make = ?", strings.ToUpper(rec[0])),
			qm.And("model = ?", strings.ToUpper(rec[1])),
			qm.And("year = ?", yrInt)).One(ctx, pdb.DBS().Writer)
		if err != nil {
			logger.Fatal().Err(err).Msg("can't read existing definition")
		}
		if dd != nil {
			continue
		}
		dd = &models.DeviceDefinition{
			ID:       ksuid.New().String(),
			Make:     rec[0],
			Model:    rec[1],
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
