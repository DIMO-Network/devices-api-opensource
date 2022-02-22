package services

import (
	"context"

	"github.com/DIMO-Network/devices-api/internal/config"
	smartcar "github.com/smartcar/go-sdk"
)

//go:generate mockgen -source smartcar_client.go -destination mocks/smartcar_client_mock.go

type SmartcarClient interface {
	ExchangeCode(ctx context.Context, code, redirectURI string) (*smartcar.Token, error)
}

type smartcarClient struct {
	settings       *config.Settings
	officialClient smartcar.Client
}

func NewSmartcarClient(settings *config.Settings) SmartcarClient {
	return &smartcarClient{
		settings:       settings,
		officialClient: smartcar.NewClient(),
	}
}

var smartcarScopes = []string{
	"read_engine_oil",
	"read_battery",
	"read_charge",
	"control_charge",
	"read_fuel",
	"read_location",
	"read_odometer",
	"read_tires",
	"read_vehicle_info",
	"read_vin",
}

func (s *smartcarClient) ExchangeCode(ctx context.Context, code, redirectURI string) (*smartcar.Token, error) {
	client := smartcar.NewClient()
	params := &smartcar.AuthParams{
		ClientID:     s.settings.SmartcarClientID,
		ClientSecret: s.settings.SmartcarClientSecret,
		RedirectURI:  redirectURI,
		Scope:        smartcarScopes,
	}
	return client.NewAuth(params).ExchangeCode(ctx, &smartcar.ExchangeCodeParams{Code: code})
}
