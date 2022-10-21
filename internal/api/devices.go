package api

import (
	"context"
	"database/sql"
	"errors"

	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/models"
	pb "github.com/DIMO-Network/shared/api/devices"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func NewUserDeviceService(dbs func() *database.DBReaderWriter, logger *zerolog.Logger) pb.UserDeviceServiceServer {
	return &userDeviceService{dbs: dbs, logger: logger}
}

type userDeviceService struct {
	pb.UnimplementedUserDeviceServiceServer
	dbs    func() *database.DBReaderWriter
	logger *zerolog.Logger
}

func (s *userDeviceService) GetUserDevice(ctx context.Context, req *pb.GetUserDeviceRequest) (*pb.UserDevice, error) {
	dbDevice, err := models.FindUserDevice(ctx, s.dbs().Reader, req.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "No device with that ID found.")
		}
		s.logger.Err(err).Str("userDeviceId", req.Id).Msg("Database failure retrieving device.")
		return nil, status.Error(codes.Internal, "Internal error.")
	}

	pbDevice := &pb.UserDevice{
		Id:        dbDevice.ID,
		UserId:    dbDevice.UserID,
		TokenId:   s.toUint64(dbDevice.TokenID),
		OptedInAt: nullTimeToPB(dbDevice.OptedInAt),
	}

	return pbDevice, nil
}

func (s *userDeviceService) ListUserDevicesForUser(ctx context.Context, req *pb.ListUserDevicesForUserRequest) (*pb.ListUserDevicesForUserResponse, error) {
	devices, err := models.UserDevices(models.UserDeviceWhere.UserID.EQ(req.UserId)).All(ctx, s.dbs().Reader)
	if err != nil {
		s.logger.Err(err).Str("userId", req.UserId).Msg("Database failure retrieving user's devices.")
		return nil, status.Error(codes.Internal, "Internal error.")
	}

	devOut := make([]*pb.UserDevice, len(devices))
	for i := 0; i < len(devices); i++ {
		device := devices[i]

		devOut[i] = &pb.UserDevice{
			Id:        device.ID,
			UserId:    device.UserID,
			TokenId:   s.toUint64(device.TokenID),
			OptedInAt: nullTimeToPB(device.OptedInAt),
		}
	}

	list := &pb.ListUserDevicesForUserResponse{UserDevices: devOut}
	return list, nil
}

// toUint64 takes a nullable decimal and returns nil if there is no value, or
// a reference to the uint64 value of the decimal otherwise. If the value does not
// fit then we return nil and log.
func (s *userDeviceService) toUint64(dec types.NullDecimal) *uint64 {
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

func nullTimeToPB(t null.Time) *timestamppb.Timestamp {
	if !t.Valid {
		return nil
	}

	return timestamppb.New(t.Time)
}
