// Code generated by MockGen. DO NOT EDIT.
// Source: internal/services/synthetic_device_instance_service.go

// Package mock_services is a generated GoMock package.
package mock_services

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockSyntheticWalletInstanceService is a mock of SyntheticWalletInstanceService interface.
type MockSyntheticWalletInstanceService struct {
	ctrl     *gomock.Controller
	recorder *MockSyntheticWalletInstanceServiceMockRecorder
}

// MockSyntheticWalletInstanceServiceMockRecorder is the mock recorder for MockSyntheticWalletInstanceService.
type MockSyntheticWalletInstanceServiceMockRecorder struct {
	mock *MockSyntheticWalletInstanceService
}

// NewMockSyntheticWalletInstanceService creates a new mock instance.
func NewMockSyntheticWalletInstanceService(ctrl *gomock.Controller) *MockSyntheticWalletInstanceService {
	mock := &MockSyntheticWalletInstanceService{ctrl: ctrl}
	mock.recorder = &MockSyntheticWalletInstanceServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSyntheticWalletInstanceService) EXPECT() *MockSyntheticWalletInstanceServiceMockRecorder {
	return m.recorder
}

// GetAddress mocks base method.
func (m *MockSyntheticWalletInstanceService) GetAddress(ctx context.Context, childNumber uint32) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAddress", ctx, childNumber)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAddress indicates an expected call of GetAddress.
func (mr *MockSyntheticWalletInstanceServiceMockRecorder) GetAddress(ctx, childNumber interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAddress", reflect.TypeOf((*MockSyntheticWalletInstanceService)(nil).GetAddress), ctx, childNumber)
}

// SignHash mocks base method.
func (m *MockSyntheticWalletInstanceService) SignHash(ctx context.Context, childNumber uint32, hash []byte) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SignHash", ctx, childNumber, hash)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SignHash indicates an expected call of SignHash.
func (mr *MockSyntheticWalletInstanceServiceMockRecorder) SignHash(ctx, childNumber, hash interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SignHash", reflect.TypeOf((*MockSyntheticWalletInstanceService)(nil).SignHash), ctx, childNumber, hash)
}
