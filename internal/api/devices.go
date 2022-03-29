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
		Id:                 dbDevice.ID,
		UserId:             dbDevice.UserID,
		DeviceDefinitionId: dbDevice.DeviceDefinitionID,
		VinIdentifier:      dbDevice.VinIdentifier.Ptr(),
	}
	return pbDevice, nil
}
