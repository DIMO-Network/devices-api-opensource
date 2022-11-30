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
	"github.com/ethereum/go-ethereum/common"
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
// @Param       tokenId path int true "token id"
// @Produce     json
// @Success     200 {object} controllers.NFTMetadataResp
// @Failure     404
// @Router      /vehicle/{tokenId} [get]
func (nc *NFTController) GetNFTMetadata(c *fiber.Ctx) error {
	tis := c.Params("tokenID")
	ti, ok := new(big.Int).SetString(tis, 10)
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Couldn't parse token id %q.", tis))
	}

	tid := types.NewNullDecimal(new(decimal.Big).SetBigMantScale(ti, 0))

	var maybeName null.String
	var deviceDefinitionID string

	nft, err := models.VehicleNFTS(
		models.VehicleNFTWhere.TokenID.EQ(tid),
		qm.Load(models.VehicleNFTRels.UserDevice),
	).One(c.Context(), nc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "NFT not found.")
		}
		nc.log.Err(err).Msg("Database error retrieving NFT metadata.")
		return opaqueInternalError
	}

	if nft.R.UserDevice == nil {
		return fiber.NewError(fiber.StatusNotFound, "NFT not found.")
	}

	maybeName = nft.R.UserDevice.Name
	deviceDefinitionID = nft.R.UserDevice.DeviceDefinitionID

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
		Image:       fmt.Sprintf("%s/v1/vehicle/%s/image", nc.Settings.DeploymentBaseURL, ti),
		Attributes: []NFTAttribute{
			{TraitType: "Make", Value: def.Make.Name},
			{TraitType: "Model", Value: def.Type.Model},
			{TraitType: "Year", Value: strconv.Itoa(int(def.Type.Year))},
		},
	})
}

type NFTMetadataResp struct {
	Name        string         `json:"name,omitempty"`
	Description string         `json:"description,omitempty"`
	Image       string         `json:"image,omitempty"`
	Attributes  []NFTAttribute `json:"attributes"`
}

type NFTAttribute struct {
	TraitType string `json:"trait_type"`
	Value     string `json:"value"`
}

// GetNFTImage godoc
// @Description Returns the image for the given vehicle NFT.
// @Tags        nfts
// @Param       tokenId     path  int  true  "token id"
// @Param       transparent query bool false "whether to remove the image background"
// @Produce     png
// @Success     200
// @Failure     404
// @Router      /vehicle/{tokenId}/image [get]
func (nc *NFTController) GetNFTImage(c *fiber.Ctx) error {
	tis := c.Params("tokenID")
	ti, ok := new(big.Int).SetString(tis, 10)
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Couldn't parse token id %q.", tis))
	}

	var transparent bool
	if c.Query("transparent") == "true" {
		transparent = true
	}

	tid := types.NewNullDecimal(new(decimal.Big).SetBigMantScale(ti, 0))

	var imageName string

	nft, err := models.VehicleNFTS(
		models.VehicleNFTWhere.TokenID.EQ(tid),
		qm.Load(models.VehicleNFTRels.UserDevice),
	).One(c.Context(), nc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "NFT not found.")
		}
		nc.log.Err(err).Msg("Database error retrieving NFT metadata.")
		return opaqueInternalError
	}

	if nft.R.UserDevice == nil {
		return fiber.NewError(fiber.StatusNotFound, "NFT not found.")
	}

	imageName = nft.MintRequestID
	suffix := ".png"

	if transparent {
		suffix = "_transparent.png"
	}

	s3o, err := nc.s3.GetObject(c.Context(), &s3.GetObjectInput{
		Bucket: aws.String(nc.Settings.NFTS3Bucket),
		Key:    aws.String(imageName + suffix),
	})
	if err != nil {
		nc.log.Err(err).Msg("Failure communicating with S3.")
		return opaqueInternalError
	}

	c.Set("Content-Type", "image/png")
	return c.SendStream(s3o.Body)
}

// GetAftermarketDeviceNFTMetadata godoc
// @Description Retrieves NFT metadata for a given aftermarket device.
// @Tags        nfts
// @Param       tokenId path int true "token id"
// @Produce     json
// @Success     200 {object} controllers.NFTMetadataResp
// @Failure     404
// @Router      /aftermarket/device/{tokenId} [get]
func (nc *NFTController) GetAftermarketDeviceNFTMetadata(c *fiber.Ctx) error {
	tidStr := c.Params("tokenID")

	tid, ok := new(big.Int).SetString(tidStr, 10)
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse token id.")
	}

	unit, err := models.AutopiUnits(
		models.AutopiUnitWhere.TokenID.EQ(types.NewNullDecimal(new(decimal.Big).SetBigMantScale(tid, 0))),
	).One(c.Context(), nc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "No device with that id.")
		}
		return err
	}

	return c.JSON(NFTMetadataResp{
		Attributes: []NFTAttribute{
			{TraitType: "Ethereum Address", Value: common.BytesToAddress(unit.EthereumAddress.Bytes).String()},
			{TraitType: "Serial Number", Value: unit.AutopiUnitID},
		},
	})
}

// GetManufacturerNFTMetadata godoc
// @Description Retrieves NFT metadata for a given manufacturer.
// @Tags        nfts
// @Param       tokenId path int true "token id"
// @Produce     json
// @Success     200 {object} controllers.NFTMetadataResp
// @Failure     404
// @Router      /manufacturer/{tokenId} [get]
func (nc *NFTController) GetManufacturerNFTMetadata(c *fiber.Ctx) error {
	tidStr := c.Params("tokenID")

	tid, ok := new(big.Int).SetString(tidStr, 10)
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse token id.")
	}

	dm, err := nc.deviceDefSvc.GetMakeByTokenID(c.Context(), tid)
	if err != nil {
		return api.GrpcErrorToFiber(err, "Couldn't retrieve manufacturer")
	}

	return c.JSON(NFTMetadataResp{
		Name:       dm.Name,
		Attributes: []NFTAttribute{},
	})
}
