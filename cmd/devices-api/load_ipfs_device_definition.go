package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

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

	err = sh.FilesRm(ctx, "/makes", true)

	if err != nil {
		return err
	}

	err = sh.FilesMkdir(ctx, "/makes")

	if err != nil {
		return err
	}

	for _, v := range makes {
		// create path
		path := fmt.Sprintf("/makes/%s", slug.Make(v.Name))
		err := sh.FilesMkdir(ctx, path)
		if err != nil {
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

	for _, definition := range all {
		path := fmt.Sprintf("%s/%s/%s",
			"/makes",
			slug.Make(definition.R.DeviceMake.Name),
			slug.Make(definition.Model))
		logger.Info().Msgf("model path: %s", path)
		err = sh.FilesRm(ctx, path, true)
		if err != nil {
			return err
		}

		err := sh.FilesMkdir(ctx, path)
		if err != nil {
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

	for _, definition := range all {
		path := fmt.Sprintf("%s/%s/%s/%d",
			"/makes",
			slug.Make(definition.R.DeviceMake.Name),
			slug.Make(definition.Model),
			definition.Year)
		logger.Info().Msgf("%s", path)
		err = sh.FilesRm(ctx, path, true)
		if err != nil {
			return err
		}

		err := sh.FilesMkdir(ctx, path)
		if err != nil {
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

	return nil
}
