// Code generated by MockGen. DO NOT EDIT.
// Source: ingest_registrar.go

// Package mock_services is a generated GoMock package.
package mock_services

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockIngestRegistrar is a mock of IngestRegistrar interface.
type MockIngestRegistrar struct {
	ctrl     *gomock.Controller
	recorder *MockIngestRegistrarMockRecorder
}

// MockIngestRegistrarMockRecorder is the mock recorder for MockIngestRegistrar.
type MockIngestRegistrarMockRecorder struct {
	mock *MockIngestRegistrar
}

// NewMockIngestRegistrar creates a new mock instance.
func NewMockIngestRegistrar(ctrl *gomock.Controller) *MockIngestRegistrar {
	mock := &MockIngestRegistrar{ctrl: ctrl}
	mock.recorder = &MockIngestRegistrarMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIngestRegistrar) EXPECT() *MockIngestRegistrarMockRecorder {
	return m.recorder
}

// Deregister mocks base method.
func (m *MockIngestRegistrar) Deregister(externalID, userDeviceID, integrationID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Deregister", externalID, userDeviceID, integrationID)
	ret0, _ := ret[0].(error)
	return ret0
}

// Deregister indicates an expected call of Deregister.
func (mr *MockIngestRegistrarMockRecorder) Deregister(externalID, userDeviceID, integrationID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Deregister", reflect.TypeOf((*MockIngestRegistrar)(nil).Deregister), externalID, userDeviceID, integrationID)
}

// Register mocks base method.
func (m *MockIngestRegistrar) Register(externalID, userDeviceID, integrationID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Register", externalID, userDeviceID, integrationID)
	ret0, _ := ret[0].(error)
	return ret0
}

// Register indicates an expected call of Register.
func (mr *MockIngestRegistrarMockRecorder) Register(externalID, userDeviceID, integrationID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Register", reflect.TypeOf((*MockIngestRegistrar)(nil).Register), externalID, userDeviceID, integrationID)
}
