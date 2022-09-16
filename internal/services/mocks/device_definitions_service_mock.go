// Code generated by MockGen. DO NOT EDIT.
// Source: device_definitions_service.go

// Package mock_services is a generated GoMock package.
package mock_services

import (
	context "context"
	reflect "reflect"

	grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	models "github.com/DIMO-Network/devices-api/models"
	gomock "github.com/golang/mock/gomock"
	boil "github.com/volatiletech/sqlboiler/v4/boil"
)

// MockDeviceDefinitionService is a mock of DeviceDefinitionService interface.
type MockDeviceDefinitionService struct {
	ctrl     *gomock.Controller
	recorder *MockDeviceDefinitionServiceMockRecorder
}

// MockDeviceDefinitionServiceMockRecorder is the mock recorder for MockDeviceDefinitionService.
type MockDeviceDefinitionServiceMockRecorder struct {
	mock *MockDeviceDefinitionService
}

// NewMockDeviceDefinitionService creates a new mock instance.
func NewMockDeviceDefinitionService(ctrl *gomock.Controller) *MockDeviceDefinitionService {
	mock := &MockDeviceDefinitionService{ctrl: ctrl}
	mock.recorder = &MockDeviceDefinitionServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDeviceDefinitionService) EXPECT() *MockDeviceDefinitionServiceMockRecorder {
	return m.recorder
}

// CheckAndSetImage mocks base method.
func (m *MockDeviceDefinitionService) CheckAndSetImage(ctx context.Context, dd *grpc.GetDeviceDefinitionItemResponse, overwrite bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckAndSetImage", ctx, dd, overwrite)
	ret0, _ := ret[0].(error)
	return ret0
}

// CheckAndSetImage indicates an expected call of CheckAndSetImage.
func (mr *MockDeviceDefinitionServiceMockRecorder) CheckAndSetImage(ctx, dd, overwrite interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckAndSetImage", reflect.TypeOf((*MockDeviceDefinitionService)(nil).CheckAndSetImage), ctx, dd, overwrite)
}

// FindDeviceDefinitionByMMY mocks base method.
func (m *MockDeviceDefinitionService) FindDeviceDefinitionByMMY(ctx context.Context, mk, model string, year int) (*grpc.GetDeviceDefinitionItemResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindDeviceDefinitionByMMY", ctx, mk, model, year)
	ret0, _ := ret[0].(*grpc.GetDeviceDefinitionItemResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindDeviceDefinitionByMMY indicates an expected call of FindDeviceDefinitionByMMY.
func (mr *MockDeviceDefinitionServiceMockRecorder) FindDeviceDefinitionByMMY(ctx, mk, model, year interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindDeviceDefinitionByMMY", reflect.TypeOf((*MockDeviceDefinitionService)(nil).FindDeviceDefinitionByMMY), ctx, mk, model, year)
}

// GetDeviceDefinitionsByIDs mocks base method.
func (m *MockDeviceDefinitionService) GetDeviceDefinitionsByIDs(ctx context.Context, ids []string) ([]*grpc.GetDeviceDefinitionItemResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDeviceDefinitionsByIDs", ctx, ids)
	ret0, _ := ret[0].([]*grpc.GetDeviceDefinitionItemResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDeviceDefinitionsByIDs indicates an expected call of GetDeviceDefinitionsByIDs.
func (mr *MockDeviceDefinitionServiceMockRecorder) GetDeviceDefinitionsByIDs(ctx, ids interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDeviceDefinitionsByIDs", reflect.TypeOf((*MockDeviceDefinitionService)(nil).GetDeviceDefinitionsByIDs), ctx, ids)
}

// GetOrCreateMake mocks base method.
func (m *MockDeviceDefinitionService) GetOrCreateMake(ctx context.Context, tx boil.ContextExecutor, makeName string) (*models.DeviceMake, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOrCreateMake", ctx, tx, makeName)
	ret0, _ := ret[0].(*models.DeviceMake)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOrCreateMake indicates an expected call of GetOrCreateMake.
func (mr *MockDeviceDefinitionServiceMockRecorder) GetOrCreateMake(ctx, tx, makeName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOrCreateMake", reflect.TypeOf((*MockDeviceDefinitionService)(nil).GetOrCreateMake), ctx, tx, makeName)
}

// PullBlackbookData mocks base method.
func (m *MockDeviceDefinitionService) PullBlackbookData(ctx context.Context, userDeviceID, deviceDefinitionID, vin string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PullBlackbookData", ctx, userDeviceID, deviceDefinitionID, vin)
	ret0, _ := ret[0].(error)
	return ret0
}

// PullBlackbookData indicates an expected call of PullBlackbookData.
func (mr *MockDeviceDefinitionServiceMockRecorder) PullBlackbookData(ctx, userDeviceID, deviceDefinitionID, vin interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PullBlackbookData", reflect.TypeOf((*MockDeviceDefinitionService)(nil).PullBlackbookData), ctx, userDeviceID, deviceDefinitionID, vin)
}

// PullDrivlyData mocks base method.
func (m *MockDeviceDefinitionService) PullDrivlyData(ctx context.Context, userDeviceID, deviceDefinitionID, vin string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PullDrivlyData", ctx, userDeviceID, deviceDefinitionID, vin)
	ret0, _ := ret[0].(error)
	return ret0
}

// PullDrivlyData indicates an expected call of PullDrivlyData.
func (mr *MockDeviceDefinitionServiceMockRecorder) PullDrivlyData(ctx, userDeviceID, deviceDefinitionID, vin interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PullDrivlyData", reflect.TypeOf((*MockDeviceDefinitionService)(nil).PullDrivlyData), ctx, userDeviceID, deviceDefinitionID, vin)
}

// UpdateDeviceDefinitionFromNHTSA mocks base method.
func (m *MockDeviceDefinitionService) UpdateDeviceDefinitionFromNHTSA(ctx context.Context, deviceDefinitionID, vin string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateDeviceDefinitionFromNHTSA", ctx, deviceDefinitionID, vin)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateDeviceDefinitionFromNHTSA indicates an expected call of UpdateDeviceDefinitionFromNHTSA.
func (mr *MockDeviceDefinitionServiceMockRecorder) UpdateDeviceDefinitionFromNHTSA(ctx, deviceDefinitionID, vin interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateDeviceDefinitionFromNHTSA", reflect.TypeOf((*MockDeviceDefinitionService)(nil).UpdateDeviceDefinitionFromNHTSA), ctx, deviceDefinitionID, vin)
}
