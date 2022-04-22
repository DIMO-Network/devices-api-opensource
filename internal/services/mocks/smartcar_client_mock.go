// Code generated by MockGen. DO NOT EDIT.
// Source: smartcar_client.go

// Package mock_services is a generated GoMock package.
package mock_services

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	smartcar "github.com/smartcar/go-sdk"
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
func (mr *MockSmartcarClientMockRecorder) ExchangeCode(ctx, code, redirectURI interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ExchangeCode", reflect.TypeOf((*MockSmartcarClient)(nil).ExchangeCode), ctx, code, redirectURI)
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
func (mr *MockSmartcarClientMockRecorder) GetEndpoints(ctx, accessToken, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEndpoints", reflect.TypeOf((*MockSmartcarClient)(nil).GetEndpoints), ctx, accessToken, id)
}

// GetExternalId mocks base method.
func (m *MockSmartcarClient) GetExternalId(ctx context.Context, accessToken string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetExternalId", ctx, accessToken)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetExternalId indicates an expected call of GetExternalId.
func (mr *MockSmartcarClientMockRecorder) GetExternalId(ctx, accessToken interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetExternalId", reflect.TypeOf((*MockSmartcarClient)(nil).GetExternalId), ctx, accessToken)
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
func (mr *MockSmartcarClientMockRecorder) GetVIN(ctx, accessToken, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetVIN", reflect.TypeOf((*MockSmartcarClient)(nil).GetVIN), ctx, accessToken, id)
}
