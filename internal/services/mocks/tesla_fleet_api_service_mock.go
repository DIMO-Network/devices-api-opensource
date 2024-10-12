// Code generated by MockGen. DO NOT EDIT.
// Source: tesla_fleet_api_service.go
//
// Generated by this command:
//
//	mockgen -source tesla_fleet_api_service.go -destination mocks/tesla_fleet_api_service_mock.go
//
// Package mock_services is a generated GoMock package.
package mock_services

import (
	context "context"
	reflect "reflect"

	services "github.com/DIMO-Network/devices-api/internal/services"
	gomock "go.uber.org/mock/gomock"
)

// MockTeslaFleetAPIService is a mock of TeslaFleetAPIService interface.
type MockTeslaFleetAPIService struct {
	ctrl     *gomock.Controller
	recorder *MockTeslaFleetAPIServiceMockRecorder
}

// MockTeslaFleetAPIServiceMockRecorder is the mock recorder for MockTeslaFleetAPIService.
type MockTeslaFleetAPIServiceMockRecorder struct {
	mock *MockTeslaFleetAPIService
}

// NewMockTeslaFleetAPIService creates a new mock instance.
func NewMockTeslaFleetAPIService(ctrl *gomock.Controller) *MockTeslaFleetAPIService {
	mock := &MockTeslaFleetAPIService{ctrl: ctrl}
	mock.recorder = &MockTeslaFleetAPIServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTeslaFleetAPIService) EXPECT() *MockTeslaFleetAPIServiceMockRecorder {
	return m.recorder
}

// CompleteTeslaAuthCodeExchange mocks base method.
func (m *MockTeslaFleetAPIService) CompleteTeslaAuthCodeExchange(ctx context.Context, authCode, redirectURI string) (*services.TeslaAuthCodeResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CompleteTeslaAuthCodeExchange", ctx, authCode, redirectURI)
	ret0, _ := ret[0].(*services.TeslaAuthCodeResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CompleteTeslaAuthCodeExchange indicates an expected call of CompleteTeslaAuthCodeExchange.
func (mr *MockTeslaFleetAPIServiceMockRecorder) CompleteTeslaAuthCodeExchange(ctx, authCode, redirectURI any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CompleteTeslaAuthCodeExchange", reflect.TypeOf((*MockTeslaFleetAPIService)(nil).CompleteTeslaAuthCodeExchange), ctx, authCode, redirectURI)
}

// GetAvailableCommands mocks base method.
func (m *MockTeslaFleetAPIService) GetAvailableCommands(token string) (*services.UserDeviceAPIIntegrationsMetadataCommands, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAvailableCommands", token)
	ret0, _ := ret[0].(*services.UserDeviceAPIIntegrationsMetadataCommands)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAvailableCommands indicates an expected call of GetAvailableCommands.
func (mr *MockTeslaFleetAPIServiceMockRecorder) GetAvailableCommands(token any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAvailableCommands", reflect.TypeOf((*MockTeslaFleetAPIService)(nil).GetAvailableCommands), token)
}

// GetTelemetrySubscriptionStatus mocks base method.
func (m *MockTeslaFleetAPIService) GetTelemetrySubscriptionStatus(ctx context.Context, token string, tokenID int) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTelemetrySubscriptionStatus", ctx, token, tokenID)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetTelemetrySubscriptionStatus indicates an expected call of GetTelemetrySubscriptionStatus.
func (mr *MockTeslaFleetAPIServiceMockRecorder) GetTelemetrySubscriptionStatus(ctx, token, tokenID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTelemetrySubscriptionStatus", reflect.TypeOf((*MockTeslaFleetAPIService)(nil).GetTelemetrySubscriptionStatus), ctx, token, tokenID)
}

// GetVehicle mocks base method.
func (m *MockTeslaFleetAPIService) GetVehicle(ctx context.Context, token string, vehicleID int) (*services.TeslaVehicle, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetVehicle", ctx, token, vehicleID)
	ret0, _ := ret[0].(*services.TeslaVehicle)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetVehicle indicates an expected call of GetVehicle.
func (mr *MockTeslaFleetAPIServiceMockRecorder) GetVehicle(ctx, token, vehicleID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetVehicle", reflect.TypeOf((*MockTeslaFleetAPIService)(nil).GetVehicle), ctx, token, vehicleID)
}

// GetVehicles mocks base method.
func (m *MockTeslaFleetAPIService) GetVehicles(ctx context.Context, token string) ([]services.TeslaVehicle, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetVehicles", ctx, token)
	ret0, _ := ret[0].([]services.TeslaVehicle)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetVehicles indicates an expected call of GetVehicles.
func (mr *MockTeslaFleetAPIServiceMockRecorder) GetVehicles(ctx, token any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetVehicles", reflect.TypeOf((*MockTeslaFleetAPIService)(nil).GetVehicles), ctx, token)
}

// SubscribeForTelemetryData mocks base method.
func (m *MockTeslaFleetAPIService) SubscribeForTelemetryData(ctx context.Context, token, vin string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubscribeForTelemetryData", ctx, token, vin)
	ret0, _ := ret[0].(error)
	return ret0
}

// SubscribeForTelemetryData indicates an expected call of SubscribeForTelemetryData.
func (mr *MockTeslaFleetAPIServiceMockRecorder) SubscribeForTelemetryData(ctx, token, vin any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubscribeForTelemetryData", reflect.TypeOf((*MockTeslaFleetAPIService)(nil).SubscribeForTelemetryData), ctx, token, vin)
}

// VirtualKeyConnectionStatus mocks base method.
func (m *MockTeslaFleetAPIService) VirtualKeyConnectionStatus(ctx context.Context, token, vin string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "VirtualKeyConnectionStatus", ctx, token, vin)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// VirtualKeyConnectionStatus indicates an expected call of VirtualKeyConnectionStatus.
func (mr *MockTeslaFleetAPIServiceMockRecorder) VirtualKeyConnectionStatus(ctx, token, vin any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "VirtualKeyConnectionStatus", reflect.TypeOf((*MockTeslaFleetAPIService)(nil).VirtualKeyConnectionStatus), ctx, token, vin)
}

// WakeUpVehicle mocks base method.
func (m *MockTeslaFleetAPIService) WakeUpVehicle(ctx context.Context, token string, vehicleID int) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WakeUpVehicle", ctx, token, vehicleID)
	ret0, _ := ret[0].(error)
	return ret0
}

// WakeUpVehicle indicates an expected call of WakeUpVehicle.
func (mr *MockTeslaFleetAPIServiceMockRecorder) WakeUpVehicle(ctx, token, vehicleID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WakeUpVehicle", reflect.TypeOf((*MockTeslaFleetAPIService)(nil).WakeUpVehicle), ctx, token, vehicleID)
}
