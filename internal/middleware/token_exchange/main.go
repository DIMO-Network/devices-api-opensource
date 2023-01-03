package tokenexchange

import (
	"database/sql"
	"errors"
	"fmt"
	"math/big"

	"github.com/DIMO-Network/devices-api/internal/api"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/ericlagergren/decimal"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"
	"golang.org/x/exp/slices"
)

type PrivilegeHandler struct {
	Log *zerolog.Logger
	DBS func() *database.DBReaderWriter
}

func New(cfg PrivilegeHandler) PrivilegeHandler {
	return PrivilegeHandler{
		Log: cfg.Log,
		DBS: cfg.DBS,
	}
}

func (p *PrivilegeHandler) HasTokenPrivilege(privilegeID int64) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return p.checkPrivilege(c, privilegeID)
	}
}

func (p *PrivilegeHandler) checkPrivilege(c *fiber.Ctx, privilegeID int64) error {
	claims, err := api.GetVehicleTokenClaims(c)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	tkID := c.Params("tokenID")
	if tkID != claims.VehicleTokenID {
		p.Log.Debug().Str("VehicleTokenID In Request", tkID).
			Str("VehicleTokenID in bearer token", claims.VehicleTokenID).
			Msg("Invalid vehicle token")
		return fiber.NewError(fiber.StatusUnauthorized, "Unauthorized! Wrong vehicle token provided")
	}

	ti, ok := new(big.Int).SetString(tkID, 10)
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Couldn't parse token id %q.", tkID))
	}

	tid := types.NewNullDecimal(new(decimal.Big).SetBigMantScale(ti, 0))

	privilegeFound := slices.Contains(claims.Privileges, privilegeID)
	if !privilegeFound {
		p.Log.Debug().Interface("Privilege In Request", claims.Privileges).Msg("Invalid privilege requested")
		return fiber.NewError(fiber.StatusUnauthorized, "Unauthorized! Token does not contain privilege 1.")
	}

	// Verify vehicle exists for tokenID
	nft, err := models.VehicleNFTS(
		models.VehicleNFTWhere.TokenID.EQ(tid),
		qm.Load(models.VehicleNFTRels.UserDevice),
	).One(c.Context(), p.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "Vehicle not found!")
		}
		p.Log.Err(err).Msg("Database error retrieving Vehicle metadata.")
		return fiber.NewError(fiber.StatusInternalServerError, "Internal error.")
	}

	if nft.R.UserDevice == nil {
		return fiber.NewError(fiber.StatusNotFound, "Vehicle not found!")
	}

	// Verify privilege is correct
	c.Locals("vehicleTokenClaims", api.VehicleTokenClaims{
		VehicleTokenID: claims.VehicleTokenID,
		UserEthAddress: claims.UserEthAddress,
		Privileges:     claims.Privileges,
	})

	return c.Next()
}
