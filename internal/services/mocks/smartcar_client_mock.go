// Code generated by MockGen. DO NOT EDIT.
// Source: smartcar_client.go
//
// Generated by this command:
//
//	mockgen -source smartcar_client.go -destination mocks/smartcar_client_mock.go
//
// Package mock_services is a generated GoMock package.
package mock_services

import (
	context "context"
	reflect "reflect"

	services "github.com/DIMO-Network/devices-api/internal/services"
	smartcar "github.com/smartcar/go-sdk"
	gomock "go.uber.org/mock/gomock"
)

// MockSmartcarClient is a mock of SmartcarClient interface.
type MockSmartcarClient struct {
	ctrl     *gomock.Controller
	recorder *MockSmartcarClientMockRecorder
}

// MockSmartcarClientMockRecorder is the mock recorder for MockSmartcarClient.
type MockSmartcarClientMockRecorder struct {
	mock *MockSmartcarClient
}

// NewMockSmartcarClient creates a new mock instance.
func NewMockSmartcarClient(ctrl *gomock.Controller) *MockSmartcarClient {
	mock := &MockSmartcarClient{ctrl: ctrl}
	mock.recorder = &MockSmartcarClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSmartcarClient) EXPECT() *MockSmartcarClientMockRecorder {
	return m.recorder
}

// ExchangeCode mocks base method.
func (m *MockSmartcarClient) ExchangeCode(ctx context.Context, code, redirectURI string) (*smartcar.Token, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ExchangeCode", ctx, code, redirectURI)
	ret0, _ := ret[0].(*smartcar.Token)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ExchangeCode indicates an expected call of ExchangeCode.
func (mr *MockSmartcarClientMockRecorder) ExchangeCode(ctx, code, redirectURI any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ExchangeCode", reflect.TypeOf((*MockSmartcarClient)(nil).ExchangeCode), ctx, code, redirectURI)
}

// GetAvailableCommands mocks base method.
func (m *MockSmartcarClient) GetAvailableCommands() *services.UserDeviceAPIIntegrationsMetadataCommands {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAvailableCommands")
	ret0, _ := ret[0].(*services.UserDeviceAPIIntegrationsMetadataCommands)
	return ret0
}

// GetAvailableCommands indicates an expected call of GetAvailableCommands.
func (mr *MockSmartcarClientMockRecorder) GetAvailableCommands() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAvailableCommands", reflect.TypeOf((*MockSmartcarClient)(nil).GetAvailableCommands))
}

// GetEndpoints mocks base method.
func (m *MockSmartcarClient) GetEndpoints(ctx context.Context, accessToken, id string) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetEndpoints", ctx, accessToken, id)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetEndpoints indicates an expected call of GetEndpoints.
func (mr *MockSmartcarClientMockRecorder) GetEndpoints(ctx, accessToken, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEndpoints", reflect.TypeOf((*MockSmartcarClient)(nil).GetEndpoints), ctx, accessToken, id)
}

// GetExternalID mocks base method.
func (m *MockSmartcarClient) GetExternalID(ctx context.Context, accessToken string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetExternalID", ctx, accessToken)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetExternalID indicates an expected call of GetExternalID.
func (mr *MockSmartcarClientMockRecorder) GetExternalID(ctx, accessToken any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetExternalID", reflect.TypeOf((*MockSmartcarClient)(nil).GetExternalID), ctx, accessToken)
}

// GetInfo mocks base method.
func (m *MockSmartcarClient) GetInfo(ctx context.Context, accessToken, id string) (*smartcar.Info, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetInfo", ctx, accessToken, id)
	ret0, _ := ret[0].(*smartcar.Info)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetInfo indicates an expected call of GetInfo.
func (mr *MockSmartcarClientMockRecorder) GetInfo(ctx, accessToken, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetInfo", reflect.TypeOf((*MockSmartcarClient)(nil).GetInfo), ctx, accessToken, id)
}

// GetUserID mocks base method.
func (m *MockSmartcarClient) GetUserID(ctx context.Context, accessToken string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserID", ctx, accessToken)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserID indicates an expected call of GetUserID.
func (mr *MockSmartcarClientMockRecorder) GetUserID(ctx, accessToken any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserID", reflect.TypeOf((*MockSmartcarClient)(nil).GetUserID), ctx, accessToken)
}

// GetVIN mocks base method.
func (m *MockSmartcarClient) GetVIN(ctx context.Context, accessToken, id string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetVIN", ctx, accessToken, id)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetVIN indicates an expected call of GetVIN.
func (mr *MockSmartcarClientMockRecorder) GetVIN(ctx, accessToken, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetVIN", reflect.TypeOf((*MockSmartcarClient)(nil).GetVIN), ctx, accessToken, id)
}

// HasDoorControl mocks base method.
func (m *MockSmartcarClient) HasDoorControl(ctx context.Context, accessToken, id string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HasDoorControl", ctx, accessToken, id)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// HasDoorControl indicates an expected call of HasDoorControl.
func (mr *MockSmartcarClientMockRecorder) HasDoorControl(ctx, accessToken, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HasDoorControl", reflect.TypeOf((*MockSmartcarClient)(nil).HasDoorControl), ctx, accessToken, id)
}
