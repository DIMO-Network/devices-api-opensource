package controllers

import (
	"database/sql"
	"fmt"
	"math/big"
	"strconv"

	"github.com/DIMO-Network/devices-api/internal/api"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ericlagergren/decimal"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"
)

type NFTController struct {
	Settings     *config.Settings
	DBS          func() *database.DBReaderWriter
	s3           *s3.Client
	log          *zerolog.Logger
	deviceDefSvc services.DeviceDefinitionService
}

// NewNFTController constructor
func NewNFTController(settings *config.Settings, dbs func() *database.DBReaderWriter, logger *zerolog.Logger, s3 *s3.Client,
	deviceDefSvc services.DeviceDefinitionService) NFTController {
	return NFTController{
		Settings:     settings,
		DBS:          dbs,
		log:          logger,
		s3:           s3,
		deviceDefSvc: deviceDefSvc,
	}
}

// GetNFTMetadata godoc
// @Description retrieves NFT metadata for a given tokenID
// @Tags        nfts
// @Param       tokenID path int true "NFT token ID"
// @Produce     json
// @Success     200 {object} controllers.NFTMetadataResp
// @Failure     404
// @Router      /nfts/{tokenID} [get]
func (nc *NFTController) GetNFTMetadata(c *fiber.Ctx) error {
	tis := c.Params("tokenID")
	ti, ok := new(big.Int).SetString(tis, 10)
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Couldn't parse token id %q.", tis))
	}

	tid := types.NewNullDecimal(new(decimal.Big).SetBigMantScale(ti, 0))

	var maybeName null.String
	var deviceDefinitionID string

	if nc.Settings.Environment != "prod" {
		ud, err := models.UserDevices(models.UserDeviceWhere.TokenID.EQ(tid)).One(c.Context(), nc.DBS().Reader)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fiber.NewError(fiber.StatusNotFound, "NFT not found.")
			}
			nc.log.Err(err).Msg("Database error retrieving NFT metadata.")
			return opaqueInternalError
		}
		maybeName = ud.Name
		deviceDefinitionID = ud.DeviceDefinitionID
	} else {
		mr, err := models.MintRequests(
			models.MintRequestWhere.TokenID.EQ(tid),
			qm.Load(models.MintRequestRels.UserDevice),
		).One(c.Context(), nc.DBS().Writer)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fiber.NewError(fiber.StatusNotFound, "NFT not found.")
			}
			nc.log.Err(err).Msg("Database error retrieving NFT metadata.")
			return opaqueInternalError
		}
		maybeName = mr.R.UserDevice.Name
		deviceDefinitionID = mr.R.UserDevice.DeviceDefinitionID
	}

	def, err := nc.deviceDefSvc.GetDeviceDefinitionByID(c.Context(), deviceDefinitionID)
	if err != nil {
		return api.GrpcErrorToFiber(err, "failed to get device definition")
	}

	description := fmt.Sprintf("%s %s %d", def.Make.Name, def.Type.Model, def.Type.Year)

	var name string
	if maybeName.Valid {
		name = maybeName.String
	} else {
		name = description
	}

	return c.JSON(NFTMetadataResp{
		Name:        name,
		Description: description,
		Image:       fmt.Sprintf("%s/v1/nfts/%s/image", nc.Settings.DeploymentBaseURL, ti),
		Attributes: []NFTAttribute{
			{TraitType: "Make", Value: def.Make.Name},
			{TraitType: "Model", Value: def.Type.Model},
			{TraitType: "Year", Value: strconv.Itoa(int(def.Type.Year))},
		},
	})
}

type NFTMetadataResp struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Image       string         `json:"image"`
	Attributes  []NFTAttribute `json:"attributes"`
}

type NFTAttribute struct {
	TraitType string `json:"trait_type"`
	Value     string `json:"value"`
}

// GetNFTImage godoc
// @Description retrieves NFT metadata for a given tokenID
// @Tags        nfts
// @Param       tokenID     path  int  true  "NFT token ID"
// @Param       transparent query bool false "If true, remove the background in the PNG. Defaults to false."
// @Produce     png
// @Router      /nfts/:tokenID/image [get]
func (nc *NFTController) GetNFTImage(c *fiber.Ctx) error {
	tis := c.Params("tokenID")
	ti, ok := new(big.Int).SetString(tis, 10)
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Couldn't parse token id %q.", tis))
	}

	tid := types.NewNullDecimal(new(decimal.Big).SetBigMantScale(ti, 0))

	var imageName string

	if nc.Settings.Environment != "prod" {
		ud, err := models.UserDevices(models.UserDeviceWhere.TokenID.EQ(tid)).One(c.Context(), nc.DBS().Reader)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fiber.NewError(fiber.StatusNotFound, "NFT not found.")
			}
			nc.log.Err(err).Msg("Database error retrieving NFT metadata.")
			return opaqueInternalError
		}
		imageName = ud.ID
	} else {
		mr, err := models.MintRequests(
			models.MintRequestWhere.TokenID.EQ(tid),
			qm.Load(models.MintRequestRels.UserDevice),
		).One(c.Context(), nc.DBS().Writer)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fiber.NewError(fiber.StatusNotFound, "NFT not found.")
			}
			nc.log.Err(err).Msg("Database error retrieving NFT metadata.")
			return opaqueInternalError
		}
		imageName = mr.ID
	}

	s3o, err := nc.s3.GetObject(c.Context(), &s3.GetObjectInput{
		Bucket: aws.String(nc.Settings.NFTS3Bucket),
		Key:    aws.String(imageName + ".png"),
	})
	if err != nil {
		nc.log.Err(err).Msg("Failure communicating with S3.")
		return opaqueInternalError
	}

	c.Set("Content-Type", "image/png")
	return c.SendStream(s3o.Body)
}
