package services

import (
	"context"
	"database/sql"
	"fmt"

	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

//go:generate mockgen -source device_definition_integration_service.go -destination mocks/device_definition_integration_service_mock.go

type DeviceDefinitionIntegrationService interface {
	GetAutoPiIntegration(ctx context.Context) (*ddgrpc.GetIntegrationItemResponse, error)
	AppendAutoPiCompatibility(ctx context.Context, dcs []DeviceCompatibility, deviceDefinitionID string) ([]DeviceCompatibility, error)
	FindUserDeviceAutoPiIntegration(ctx context.Context, exec boil.ContextExecutor, userDeviceID, userID string) (*models.UserDeviceAPIIntegration, *UserDeviceAPIIntegrationsMetadata, error)
	CreateDeviceDefinitionIntegration(ctx context.Context, integrationID string, deviceDefinitionID string, region string) (*ddgrpc.GetIntegrationItemResponse, error)
	GetDeviceDefinitionIntegration(ctx context.Context, deviceDefinitionID string) ([]*ddgrpc.GetDeviceDefinitionIntegrationItemResponse, error)
}

type deviceDefinitionIntegrationService struct {
	dbs                 func() *database.DBReaderWriter
	definitionsGRPCAddr string
}

func NewDeviceDefinitionIntegrationService(DBS func() *database.DBReaderWriter, settings *config.Settings) DeviceDefinitionIntegrationService {
	return &deviceDefinitionIntegrationService{
		dbs:                 DBS,
		definitionsGRPCAddr: settings.DefinitionsGRPCAddr,
	}
}

// GetAutoPiIntegration calls integrations api via GRPC to get the definition. idea for testing: http://www.inanzzz.com/index.php/post/w9qr/unit-testing-golang-grpc-client-and-server-application-with-bufconn-package
func (d *deviceDefinitionIntegrationService) GetAutoPiIntegration(ctx context.Context) (*ddgrpc.GetIntegrationItemResponse, error) {
	const (
		autoPiType  = models.IntegrationTypeHardware
		autoPiStyle = models.IntegrationStyleAddon
	)

	definitionsClient, conn, err := d.getDeviceDefsIntGrpcClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	definitions, err := definitionsClient.GetIntegrations(ctx, &ddgrpc.EmptyRequest{})
	if err != nil {
		return nil, err
	}

	integrationResponse := definitions.GetIntegrations()

	integration := &ddgrpc.GetIntegrationItemResponse{}

	for _, integrationItem := range integrationResponse {
		if integrationItem.Vendor == AutoPiVendor && integrationItem.Style == autoPiStyle && integrationItem.Type == autoPiType {
			return integration, nil
		}
	}

	return nil, errors.New("Autopi integration not found")
}

// AppendAutoPiCompatibility adds autopi compatibility for AmericasRegion and EuropeRegion regions
func (d *deviceDefinitionIntegrationService) AppendAutoPiCompatibility(ctx context.Context, dcs []DeviceCompatibility, deviceDefinitionID string) ([]DeviceCompatibility, error) {
	integration, err := d.GetAutoPiIntegration(ctx)
	if err != nil {
		return nil, err
	}

	for _, dc := range dcs {
		if dc.ID == integration.Id {
			return dcs, nil
		}
	}

	// insert into db
	di, err := d.CreateDeviceDefinitionIntegration(ctx, integration.Id, deviceDefinitionID, AmericasRegion.String())
	if err != nil {
		return nil, err
	}

	if di == nil {
		return nil, errors.New("Failed to create device integration")
	}

	di, err = d.CreateDeviceDefinitionIntegration(ctx, integration.Id, deviceDefinitionID, EuropeRegion.String())
	if err != nil {
		return nil, err
	}

	if di == nil {
		return nil, errors.New("Failed to create device integration")
	}

	// prepare return object for api
	dcs = append(dcs, DeviceCompatibility{
		ID:           integration.Id,
		Type:         integration.Type,
		Style:        integration.Style,
		Vendor:       integration.Vendor,
		Region:       AmericasRegion.String(),
		Capabilities: nil,
	})
	dcs = append(dcs, DeviceCompatibility{
		ID:           integration.Id,
		Type:         integration.Type,
		Style:        integration.Style,
		Vendor:       integration.Vendor,
		Region:       EuropeRegion.String(),
		Capabilities: nil,
	})

	return dcs, nil
}

// FindUserDeviceAutoPiIntegration gets the user_device_api_integration record and unmarshalled metadata, returns fiber error where makes sense
func (d *deviceDefinitionIntegrationService) FindUserDeviceAutoPiIntegration(ctx context.Context, exec boil.ContextExecutor, userDeviceID, userID string) (*models.UserDeviceAPIIntegration, *UserDeviceAPIIntegrationsMetadata, error) {
	autoPiInteg, err := d.GetAutoPiIntegration(ctx)
	if err != nil {
		return nil, nil, err
	}
	ud, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
		models.UserDeviceWhere.UserID.EQ(userID),
		qm.Load(models.UserDeviceRels.UserDeviceAPIIntegrations),
	).One(ctx, exec)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("could not find device with id %s for user %s", userDeviceID, userID))
		}
		return nil, nil, errors.Wrap(err, "Unexpected database error searching for user device")
	}
	udai := new(models.UserDeviceAPIIntegration)
	for _, apiInteg := range ud.R.UserDeviceAPIIntegrations {
		if apiInteg.IntegrationID == autoPiInteg.Id {
			udai = apiInteg
		}
	}
	if !(udai != nil && udai.ExternalID.Valid) {
		return nil, nil, fiber.NewError(fiber.StatusBadRequest, "user does not have an autopi integration registered for userDeviceId: "+userDeviceID)
	}
	// get metadata for a little later
	md := new(UserDeviceAPIIntegrationsMetadata)
	err = udai.Metadata.Unmarshal(md)
	if err != nil {
		return nil, nil, errors.Wrap(err, "metadata for user device api integrations in wrong format for unmarshal")
	}
	return udai, md, nil
}

// CreateDeviceDefinitionIntegration calls device definitions integration api via GRPC to get the definition. idea for testing: http://www.inanzzz.com/index.php/post/w9qr/unit-testing-golang-grpc-client-and-server-application-with-bufconn-package
func (d *deviceDefinitionIntegrationService) CreateDeviceDefinitionIntegration(ctx context.Context, integrationID string, deviceDefinitionID string, region string) (*ddgrpc.GetIntegrationItemResponse, error) {
	definitionsClient, conn, err := d.getDeviceDefsIntGrpcClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	di, err := definitionsClient.CreateDeviceIntegration(ctx, &ddgrpc.CreateDeviceIntegrationRequest{
		DeviceDefinitionId: deviceDefinitionID,
		IntegrationId:      integrationID,
		Region:             region,
	})
	if err != nil {
		return nil, err
	}

	deviceIntegrations, err := d.GetDeviceDefinitionIntegration(ctx, di.Id)
	if err != nil {
		return nil, err
	}

	for _, item := range deviceIntegrations {
		if item.Id == integrationID {
			return &ddgrpc.GetIntegrationItemResponse{
				Id:     item.Id,
				Vendor: item.Vendor,
				Style:  item.Style,
				Type:   item.Type,
			}, nil
		}
	}

	return nil, nil
}

// GetDeviceDefinitionIntegration calls device definitions integrations api via GRPC to get the definition. idea for testing: http://www.inanzzz.com/index.php/post/w9qr/unit-testing-golang-grpc-client-and-server-application-with-bufconn-package
func (d *deviceDefinitionIntegrationService) GetDeviceDefinitionIntegration(ctx context.Context, deviceDefinitionID string) ([]*ddgrpc.GetDeviceDefinitionIntegrationItemResponse, error) {
	definitionsClient, conn, err := d.getDeviceDefsIntGrpcClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	definitions, err := definitionsClient.GetDeviceDefinitionIntegration(ctx, &ddgrpc.GetDeviceDefinitionIntegrationRequest{
		Id: deviceDefinitionID,
	})
	if err != nil {
		return nil, err
	}

	return definitions.GetIntegrations(), nil
}

// getDeviceDefsIntGrpcClient instanties new connection with client to dd service. You must defer conn.close from returned connection
func (d *deviceDefinitionIntegrationService) getDeviceDefsIntGrpcClient() (ddgrpc.DeviceDefinitionServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(d.definitionsGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, conn, err
	}
	definitionsClient := ddgrpc.NewDeviceDefinitionServiceClient(conn)
	return definitionsClient, conn, nil
}
