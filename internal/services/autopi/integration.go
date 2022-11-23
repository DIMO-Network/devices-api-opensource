package autopi

import (
	"context"
	"fmt"
	"math/big"

	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/ericlagergren/decimal"
	"github.com/pkg/errors"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"
)

type Integration struct {
	db     func() *database.DBReaderWriter
	defs   services.DeviceDefinitionService
	ap     services.AutoPiAPIService
	apTask services.AutoPiTaskService
	apReg  services.IngestRegistrar
}

func NewIntegration(
	db func() *database.DBReaderWriter,
	defs services.DeviceDefinitionService,
	ap services.AutoPiAPIService,
	apTask services.AutoPiTaskService,
	apReg services.IngestRegistrar,
) *Integration {
	return &Integration{db: db, defs: defs, ap: ap, apTask: apTask, apReg: apReg}
}

func intToDec(x *big.Int) types.NullDecimal {
	return types.NewNullDecimal(new(decimal.Big).SetBigMantScale(x, 0))
}

func powertrainToTemplate(pt *services.PowertrainType, integ *ddgrpc.Integration) int32 {
	out := integ.AutoPiDefaultTemplateId
	if pt != nil {
		switch *pt {
		case services.ICE:
			out = integ.AutoPiPowertrainTemplate.ICE
		case services.HEV:
			out = integ.AutoPiPowertrainTemplate.HEV
		case services.PHEV:
			out = integ.AutoPiPowertrainTemplate.PHEV
		case services.BEV:
			out = integ.AutoPiPowertrainTemplate.BEV
		}
	}
	return out
}

func (i *Integration) Pair(ctx context.Context, autoPiTokenID, vehicleTokenID *big.Int) error {
	tx, err := i.db().Writer.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	integ, err := i.defs.GetIntegrationByVendor(ctx, "AutoPi")
	if err != nil {
		return err
	}

	var autoPiUnitID string

	autoPiModel, err := models.AutopiUnits(
		models.AutopiUnitWhere.TokenID.EQ(intToDec(autoPiTokenID)),
	).One(ctx, tx)
	if err != nil {
		return err
	}

	autoPiUnitID = autoPiModel.AutopiUnitID

	autoPi, err := i.ap.GetDeviceByUnitID(autoPiUnitID)
	if err != nil {
		return err
	}

	nft, err := models.VehicleNFTS(
		models.VehicleNFTWhere.TokenID.EQ(intToDec(vehicleTokenID)),
		qm.Load(models.VehicleNFTRels.UserDevice),
	).One(ctx, tx)
	if err != nil {
		return err
	}

	if nft.R.UserDevice == nil {
		return errors.New("vehicle deleted")
	}

	ud := nft.R.UserDevice

	// There are some old units that were paired off-chain but not on.
	already, err := models.UserDeviceAPIIntegrationExists(ctx, tx, ud.ID, integ.Id)
	if err != nil || already {
		return err
	}

	def, err := i.defs.GetDeviceDefinitionByID(ctx, ud.DeviceDefinitionID)
	if err != nil {
		return err
	}

	if integ.AutoPiDefaultTemplateId == 0 {
		return fmt.Errorf("integration lacks a default template")
	}

	templateID := int(integ.AutoPiDefaultTemplateId)

	if integ.AutoPiPowertrainTemplate != nil {
		udMd := services.UserDeviceMetadata{}
		err = ud.Metadata.Unmarshal(&udMd)
		if err != nil {
			return err
		}
		templateID = int(powertrainToTemplate(udMd.PowertrainType, integ))
	}

	udMetadata := services.UserDeviceAPIIntegrationsMetadata{
		AutoPiUnitID:          &autoPi.UnitID,
		AutoPiIMEI:            &autoPi.IMEI,
		AutoPiTemplateApplied: &templateID,
	}

	apiInt := models.UserDeviceAPIIntegration{
		UserDeviceID:  ud.ID,
		IntegrationID: integ.Id,
		ExternalID:    null.StringFrom(autoPi.ID),
		Status:        models.UserDeviceAPIIntegrationStatusPending,
		AutopiUnitID:  null.StringFrom(autoPi.UnitID),
	}

	err = apiInt.Metadata.Marshal(udMetadata)
	if err != nil {
		return err
	}

	if err = apiInt.Insert(ctx, tx, boil.Infer()); err != nil {
		return err
	}

	substatus := constants.QueriedDeviceOk
	// update integration record as failed if errors after this
	defer func() {
		if err != nil {
			apiInt.Status = models.UserDeviceAPIIntegrationStatusFailed
			msg := err.Error()
			udMetadata.AutoPiRegistrationError = &msg
			ss := substatus.String()
			udMetadata.AutoPiSubStatus = &ss
			_ = apiInt.Metadata.Marshal(udMetadata)
			_, _ = apiInt.Update(ctx, tx,
				boil.Whitelist(models.UserDeviceAPIIntegrationColumns.Status, models.UserDeviceAPIIntegrationColumns.UpdatedAt))
			err = tx.Commit()
		}
	}()
	// update the profile on autopi
	profile := services.PatchVehicleProfile{
		Year: int(def.Type.Year),
	}
	if !ud.VinIdentifier.IsZero() {
		profile.Vin = ud.VinIdentifier.String
	}
	if !ud.Name.IsZero() {
		profile.CallName = ud.Name.String
	}
	err = i.ap.PatchVehicleProfile(autoPi.Vehicle.ID, profile)
	if err != nil {
		return errors.Wrap(err, "failed to patch autopi vehicle profile")
	}

	substatus = constants.PatchedVehicleProfile
	// update autopi to unassociate from current base template
	if autoPi.Template > 0 {
		err = i.ap.UnassociateDeviceTemplate(autoPi.ID, autoPi.Template)
		if err != nil {
			return errors.Wrapf(err, "failed to unassociate template %d", autoPi.Template)
		}
	}
	// set our template on the autoPiDevice
	err = i.ap.AssociateDeviceToTemplate(autoPi.ID, templateID)
	if err != nil {
		return errors.Wrapf(err, "failed to associate autoPiDevice %s to template %d", autoPi.ID, templateID)
	}
	substatus = constants.AssociatedDeviceToTemplate
	// apply for next reboot
	err = i.ap.ApplyTemplate(autoPi.ID, templateID)
	if err != nil {
		return errors.Wrapf(err, "failed to apply autoPiDevice %s with template %d", autoPi.ID, templateID)
	}

	substatus = constants.AppliedTemplate
	// send sync command in case autoPiDevice is on at this moment (should be during initial setup)
	_, err = i.ap.CommandSyncDevice(ctx, autoPi.UnitID, autoPi.ID, ud.ID)
	if err != nil {
		return errors.Wrapf(err, "failed to sync changes to autoPiDevice %s", autoPi.ID)
	}

	substatus = constants.PendingTemplateConfirm
	ss := substatus.String()
	udMetadata.AutoPiSubStatus = &ss
	err = apiInt.Metadata.Marshal(udMetadata)
	if err != nil {
		return errors.Wrap(err, "failed to marshall user device integration metadata")
	}

	_, err = apiInt.Update(ctx, tx, boil.Whitelist(models.UserDeviceAPIIntegrationColumns.Metadata,
		models.UserDeviceAPIIntegrationColumns.UpdatedAt))
	if err != nil {
		return errors.Wrap(err, "failed to update integration status to Pending")
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit new autopi integration")
	}

	// send kafka message to autopi ingest registrar. Note we're using the UnitID for the data stream join.
	err = i.apReg.Register(autoPi.UnitID, ud.ID, integ.Id)
	if err != nil {
		return err
	}

	_, err = i.apTask.StartQueryAndUpdateVIN(autoPi.ID, autoPi.UnitID, ud.ID)

	return nil
}
