package api

import (
	"context"
	"database/sql"
	"errors"

	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/models"
	pb "github.com/DIMO-Network/shared/api/devices"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		Id:     dbDevice.ID,
		UserId: dbDevice.UserID,
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
			Id:     device.ID,
			UserId: device.UserID,
		}
	}

	list := &pb.ListUserDevicesForUserResponse{UserDevices: devOut}
	return list, nil
}
