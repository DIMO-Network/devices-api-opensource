// Code generated by MockGen. DO NOT EDIT.
// Source: user_device_service.go
//
// Generated by this command:
//
//	mockgen -source user_device_service.go -destination mocks/user_device_service_mock.go -package mock_services
//

// Package mock_services is a generated GoMock package.
package mock_services

import (
	context "context"
	reflect "reflect"

	grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	models "github.com/DIMO-Network/devices-api/models"
	gomock "go.uber.org/mock/gomock"
)

// MockUserDeviceService is a mock of UserDeviceService interface.
type MockUserDeviceService struct {
	ctrl     *gomock.Controller
	recorder *MockUserDeviceServiceMockRecorder
}

// MockUserDeviceServiceMockRecorder is the mock recorder for MockUserDeviceService.
type MockUserDeviceServiceMockRecorder struct {
	mock *MockUserDeviceService
}

// NewMockUserDeviceService creates a new mock instance.
func NewMockUserDeviceService(ctrl *gomock.Controller) *MockUserDeviceService {
	mock := &MockUserDeviceService{ctrl: ctrl}
	mock.recorder = &MockUserDeviceServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUserDeviceService) EXPECT() *MockUserDeviceServiceMockRecorder {
	return m.recorder
}

// CreateUserDevice mocks base method.
func (m *MockUserDeviceService) CreateUserDevice(ctx context.Context, definitionID, styleID, countryCode, userID string, vin, canProtocol *string, vinConfirmed bool) (*models.UserDevice, *grpc.GetDeviceDefinitionItemResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateUserDevice", ctx, definitionID, styleID, countryCode, userID, vin, canProtocol, vinConfirmed)
	ret0, _ := ret[0].(*models.UserDevice)
	ret1, _ := ret[1].(*grpc.GetDeviceDefinitionItemResponse)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// CreateUserDevice indicates an expected call of CreateUserDevice.
func (mr *MockUserDeviceServiceMockRecorder) CreateUserDevice(ctx, definitionID, styleID, countryCode, userID, vin, canProtocol, vinConfirmed any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateUserDevice", reflect.TypeOf((*MockUserDeviceService)(nil).CreateUserDevice), ctx, definitionID, styleID, countryCode, userID, vin, canProtocol, vinConfirmed)
}

// CreateUserDeviceByOwner mocks base method.
func (m *MockUserDeviceService) CreateUserDeviceByOwner(ctx context.Context, definitionID, styleID, countryCode, vin string, ownerAddress []byte) (*models.UserDevice, *grpc.GetDeviceDefinitionItemResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateUserDeviceByOwner", ctx, definitionID, styleID, countryCode, vin, ownerAddress)
	ret0, _ := ret[0].(*models.UserDevice)
	ret1, _ := ret[1].(*grpc.GetDeviceDefinitionItemResponse)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// CreateUserDeviceByOwner indicates an expected call of CreateUserDeviceByOwner.
func (mr *MockUserDeviceServiceMockRecorder) CreateUserDeviceByOwner(ctx, definitionID, styleID, countryCode, vin, ownerAddress any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateUserDeviceByOwner", reflect.TypeOf((*MockUserDeviceService)(nil).CreateUserDeviceByOwner), ctx, definitionID, styleID, countryCode, vin, ownerAddress)
}
