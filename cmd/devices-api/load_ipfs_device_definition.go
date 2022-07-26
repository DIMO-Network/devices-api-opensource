package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/gosimple/slug"
	shell "github.com/ipfs/go-ipfs-api"
	files "github.com/ipfs/go-ipfs-files"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// load makes, models and device definitions. This only needs to be run once.
func loadSyncIPFSDeviceDefinition(ctx context.Context, logger *zerolog.Logger, settings *config.Settings, pdb database.DbStore) error {
	// get all devices from DB.

	sh := shell.NewShell(settings.IPFSNodeEndpoint)

	all, err := models.DeviceDefinitions(models.DeviceDefinitionWhere.Verified.EQ(true),
		qm.Load(models.DeviceDefinitionRels.DeviceStyles),
		qm.Load(models.DeviceDefinitionRels.DeviceMake)).All(ctx, pdb.DBS().Reader)
	if err != nil {
		return err
	}
	logger.Info().Msgf("found %d device definitions verified", len(all))
	if len(all) == 0 {
		return errors.New("0 items found to index, stopping")
	}

	makes, err := models.DeviceMakes().All(ctx, pdb.DBS().Reader)
	if err != nil {
		return err
	}

	basePath := "/makes"

	logger.Info().Msgf("Get %s directory", basePath)
	_, err = sh.FileList(basePath)

	if err != nil && !strings.Contains(err.Error(), "invalid path") {
		logger.Info().Msgf(err.Error())
		return err
	}

	logger.Info().Msgf("Creating %s directory", basePath)
	err = sh.FilesMkdir(ctx, basePath)

	if err != nil && !strings.Contains(err.Error(), "file already exists") {
		logger.Info().Msgf("error creating %s directory %s", basePath, err.Error())
		return err
	}

	logger.Info().Msgf("Creation of makes folders")

	for _, v := range makes {
		// create make path
		path := fmt.Sprintf("%s/%s", basePath, slug.Make(v.Name))

		logger.Info().Msgf("Creating make directory => %s", path)

		err := sh.FilesMkdir(ctx, path)
		if err != nil && !strings.Contains(err.Error(), "file already exists") {
			logger.Info().Msgf("error creating make %s directory %s", basePath, err.Error())
			return err
		}

		tsdBin, _ := json.Marshal(v)
		reader := bytes.NewReader(tsdBin)

		// create index.json
		fr := files.NewReaderFile(reader)
		slf := files.NewSliceDirectory([]files.DirEntry{files.FileEntry("", fr)})
		fileReader := files.NewMultiFileReader(slf, true)

		indexFilePath := fmt.Sprintf("%s/index.json", path)

		logger.Info().Msgf(indexFilePath)

		rb := sh.Request("files/write", indexFilePath)
		rb.Option("create", "true")

		err = rb.Body(fileReader).Exec(ctx, nil)
		if err != nil {
			return err
		}

	}

	logger.Info().Msgf("Creation of models folders")

	for _, definition := range all {
		path := fmt.Sprintf("%s/%s/%s",
			basePath,
			slug.Make(definition.R.DeviceMake.Name),
			slug.Make(definition.Model))

		logger.Info().Msgf("Creating model directory => %s", path)

		err := sh.FilesMkdir(ctx, path)
		if err != nil && !strings.Contains(err.Error(), "file already exists") {
			logger.Info().Msgf("error creating model %s directory %s", basePath, err.Error())
			return err
		}

		tsdBin, _ := json.Marshal(definition)
		reader := bytes.NewReader(tsdBin)

		// create index.json
		fr := files.NewReaderFile(reader)
		slf := files.NewSliceDirectory([]files.DirEntry{files.FileEntry("", fr)})
		fileReader := files.NewMultiFileReader(slf, true)

		indexFilePath := fmt.Sprintf("%s/index.json", path)

		logger.Info().Msgf(indexFilePath)

		rb := sh.Request("files/write", indexFilePath)
		rb.Option("create", "true")

		err = rb.Body(fileReader).Exec(ctx, nil)
		if err != nil {
			return err
		}
	}

	logger.Info().Msgf("Creation of models/years folders")

	for _, definition := range all {
		path := fmt.Sprintf("%s/%s/%s/%d",
			basePath,
			slug.Make(definition.R.DeviceMake.Name),
			slug.Make(definition.Model),
			definition.Year)

		logger.Info().Msgf("Creating model/year directory => %s", path)

		err := sh.FilesMkdir(ctx, path)
		if err != nil && !strings.Contains(err.Error(), "file already exists") {
			logger.Info().Msgf("error creating model/year %s directory %s", basePath, err.Error())
			return err
		}

		tsdBin, _ := json.Marshal(definition)
		reader := bytes.NewReader(tsdBin)

		// create index.json
		fr := files.NewReaderFile(reader)
		slf := files.NewSliceDirectory([]files.DirEntry{files.FileEntry("", fr)})
		fileReader := files.NewMultiFileReader(slf, true)

		indexFilePath := fmt.Sprintf("%s/index.json", path)

		logger.Info().Msgf(indexFilePath)

		rb := sh.Request("files/write", indexFilePath)
		rb.Option("create", "true")

		err = rb.Body(fileReader).Exec(ctx, nil)
		if err != nil {
			return err
		}
	}

	logger.Info().Msgf("Done !!")

	return nil
}
