package api

import (
	"context"
	"database/sql"
	"errors"

	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/models"
	pb "github.com/DIMO-Network/shared/api/devices"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func NewAftermarketDeviceService(dbs func() *database.DBReaderWriter, logger *zerolog.Logger) pb.AftermarketDeviceServiceServer {
	return &aftermarketDeviceService{dbs: dbs, logger: logger}
}

type aftermarketDeviceService struct {
	pb.UnimplementedAftermarketDeviceServiceServer
	dbs    func() *database.DBReaderWriter
	logger *zerolog.Logger
}

func (s *aftermarketDeviceService) GetDeviceBySerial(ctx context.Context, req *pb.GetDeviceBySerialRequest) (*pb.AftermarketDevice, error) {
	unit, err := models.AutopiUnits(
		models.AutopiUnitWhere.AutopiUnitID.EQ(req.Serial),
	).One(ctx, s.dbs().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "No device with that id found.")
		}
		return nil, status.Error(codes.Internal, "Internal error.")
	}

	out := pb.AftermarketDevice{
		Serial:         req.Serial,
		VehicleTokenId: s.toUint64(unit.VehicleTokenID),
	}

	return &out, nil
}

// toUint64 takes a nullable decimal and returns nil if there is no value, or
// a reference to the uint64 value of the decimal otherwise. If the value does not
// fit then we return nil and log.
func (s *aftermarketDeviceService) toUint64(dec types.NullDecimal) *uint64 {
	if dec.IsZero() {
		return nil
	}

	ui, ok := dec.Uint64()
	if !ok {
		s.logger.Error().Str("decimal", dec.String()).Msg("Value too large for uint64.")
		return nil
	}

	return &ui
}
