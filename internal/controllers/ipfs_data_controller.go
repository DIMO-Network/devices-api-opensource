package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/gofiber/fiber/v2"
	"github.com/gosimple/slug"
	shell "github.com/ipfs/go-ipfs-api"
	files "github.com/ipfs/go-ipfs-files"
)

type IPFSDataController struct {
	settings *config.Settings
	DBS      func() *database.DBReaderWriter
	sh       *shell.Shell
}

// NewDocumentsController constructor
func NewIPFSDataController(
	settings *config.Settings,
	dbs func() *database.DBReaderWriter,
	sh *shell.Shell) IPFSDataController {
	return IPFSDataController{settings: settings, DBS: dbs, sh: sh}
}

func (idc *IPFSDataController) PostMakes(c *fiber.Ctx) error {
	reg := &DataMakeRequest{}
	if err := c.BodyParser(reg); err != nil {
		return errorResponseHandler(c, err, fiber.StatusBadRequest)
	}

	// create path
	friendlyURL := slug.Make(reg.Name)
	path := fmt.Sprintf("/makes/%s", friendlyURL)

	err := idc.sh.FilesMkdir(c.Context(), path)
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	tsdBin, _ := json.Marshal(reg)
	reader := bytes.NewReader(tsdBin)

	// create index.json
	fr := files.NewReaderFile(reader)
	slf := files.NewSliceDirectory([]files.DirEntry{files.FileEntry("", fr)})
	fileReader := files.NewMultiFileReader(slf, true)

	indexFilePath := fmt.Sprintf("/%s/index.json", path)
	rb := idc.sh.Request("files/write", indexFilePath)
	rb.Option("create", "true")

	err = rb.Body(fileReader).Exec(c.Context(), nil)
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	dataStat, err := idc.sh.FilesStat(c.Context(), path)
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	return c.JSON(DataMakeResponse{ID: dataStat.Hash, Name: reg.Name})
}

func (idc *IPFSDataController) GetMakeByName(c *fiber.Ctx) error {
	make := c.Params("make")

	path := fmt.Sprintf("/makes/%s", make)
	pathIndex := fmt.Sprintf("%s/index.json", path)

	pathStat, err := idc.sh.FilesStat(c.Context(), path)
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	fileIndexStat, err := idc.sh.FilesStat(c.Context(), pathIndex)
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	contentIndex, err := idc.sh.Cat(fileIndexStat.Hash)
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	bufx := new(bytes.Buffer)
	_, err = bufx.ReadFrom(contentIndex)
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}
	newStrx := bufx.String()

	response := &DataMakeResponse{}
	_ = json.Unmarshal([]byte(newStrx), &response)
	return c.JSON(DataMakeResponse{ID: pathStat.Hash, Name: response.Name})
}

func (idc *IPFSDataController) PostModels(c *fiber.Ctx) error {
	make := c.Params("make")
	reg := &DataMakeRequest{}
	if err := c.BodyParser(reg); err != nil {
		// Return status 400 and error message.
		return errorResponseHandler(c, err, fiber.StatusBadRequest)
	}

	// create path
	friendlyURL := slug.Make(reg.Name)
	path := fmt.Sprintf("/makes/%s/%s", make, friendlyURL)

	err := idc.sh.FilesMkdir(c.Context(), path)
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	tsdBin, _ := json.Marshal(reg)
	reader := bytes.NewReader(tsdBin)

	// create index.json
	fr := files.NewReaderFile(reader)
	slf := files.NewSliceDirectory([]files.DirEntry{files.FileEntry("", fr)})
	fileReader := files.NewMultiFileReader(slf, true)

	indexFilePath := fmt.Sprintf("/%s/index.json", path)
	rb := idc.sh.Request("files/write", indexFilePath)
	rb.Option("create", "true")

	err = rb.Body(fileReader).Exec(c.Context(), nil)
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	dataStat, err := idc.sh.FilesStat(c.Context(), path)
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	return c.JSON(DataModelResponse{ID: dataStat.Hash, Name: reg.Name})
}

func (idc *IPFSDataController) GetModelByName(c *fiber.Ctx) error {
	make := c.Params("make")
	model := c.Params("model")

	path := fmt.Sprintf("/makes/%s/%s", make, model)
	pathIndex := fmt.Sprintf("%s/index.json", path)

	pathStat, err := idc.sh.FilesStat(c.Context(), path)
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	fileIndexStat, err := idc.sh.FilesStat(c.Context(), pathIndex)
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	contentIndex, err := idc.sh.Cat(fileIndexStat.Hash)
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	bufx := new(bytes.Buffer)
	_, err = bufx.ReadFrom(contentIndex)
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	newStrx := bufx.String()

	response := &DataModelResponse{}
	_ = json.Unmarshal([]byte(newStrx), &response)
	return c.JSON(DataModelResponse{ID: pathStat.Hash, Name: response.Name})
}

func (idc *IPFSDataController) PostDeviceDenifition(c *fiber.Ctx) error {
	make := c.Params("make")
	model := c.Params("model")
	year := c.Params("year")

	reg := &DataDeviceDefinitionRequest{}
	if err := c.BodyParser(reg); err != nil {
		return errorResponseHandler(c, err, fiber.StatusBadRequest)
	}

	// create path year
	path := fmt.Sprintf("/makes/%s/%s/%s", make, model, year)

	err := idc.sh.FilesMkdir(c.Context(), path)
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	// create path device definition
	friendlyURL := slug.Make(reg.Name)
	pathDefinition := fmt.Sprintf("%s/%s", path, friendlyURL)

	err = idc.sh.FilesMkdir(c.Context(), pathDefinition)
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	tsdBin, _ := json.Marshal(reg)
	reader := bytes.NewReader(tsdBin)

	// create index.json
	fr := files.NewReaderFile(reader)
	slf := files.NewSliceDirectory([]files.DirEntry{files.FileEntry("", fr)})
	fileReader := files.NewMultiFileReader(slf, true)

	indexFilePath := fmt.Sprintf("/%s/index.json", pathDefinition)
	rb := idc.sh.Request("files/write", indexFilePath)
	rb.Option("create", "true")

	err = rb.Body(fileReader).Exec(c.Context(), nil)
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	dataStat, err := idc.sh.FilesStat(c.Context(), pathDefinition)
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	return c.JSON(DataModelResponse{ID: dataStat.Hash, Name: reg.Name})
}

type DataMakeRequest struct {
	Name string `json:"name"`
}

type DataMakeResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type DataModelRequest struct {
	Name string `json:"name"`
}

type DataModelResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type DataDeviceDefinitionRequest struct {
	Name string `json:"name"`
}
