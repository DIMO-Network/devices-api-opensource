package api

import (
	"context"

	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/models"
	pb "github.com/DIMO-Network/shared/api/devices"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func NewIntegrationService(dbs func() *database.DBReaderWriter) pb.IntegrationServiceServer {
	return &integrationsService{dbs: dbs}
}

type integrationsService struct {
	pb.UnimplementedIntegrationServiceServer
	dbs    func() *database.DBReaderWriter
	logger *zerolog.Logger
}

func (s *integrationsService) ListIntegrations(ctx context.Context, _ *emptypb.Empty) (*pb.ListIntegrationsResponse, error) {
	integs, err := models.Integrations().All(ctx, s.dbs().Reader)
	if err != nil {
		s.logger.Err(err).Msg("Database failure retrieving integrations.")
		return nil, status.Error(codes.Internal, "Internal error.")
	}
	list := make([]*pb.Integration, len(integs))
	for i, integ := range integs {
		list[i] = &pb.Integration{
			Id:     integ.ID,
			Vendor: integ.Vendor,
		}
	}
	out := &pb.ListIntegrationsResponse{Integrations: list}
	return out, nil
}
