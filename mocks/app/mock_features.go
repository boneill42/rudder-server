// Code generated by MockGen. DO NOT EDIT.
// Source: app/features.go

// Package mocks is a generated GoMock package.
package mocks

import (
	gomock "github.com/golang/mock/gomock"
	jobsdb "github.com/rudderlabs/rudder-server/jobsdb"
	types "github.com/rudderlabs/rudder-server/utils/types"
	reflect "reflect"
)

// MockMigratorFeature is a mock of MigratorFeature interface
type MockMigratorFeature struct {
	ctrl     *gomock.Controller
	recorder *MockMigratorFeatureMockRecorder
}

// MockMigratorFeatureMockRecorder is the mock recorder for MockMigratorFeature
type MockMigratorFeatureMockRecorder struct {
	mock *MockMigratorFeature
}

// NewMockMigratorFeature creates a new mock instance
func NewMockMigratorFeature(ctrl *gomock.Controller) *MockMigratorFeature {
	mock := &MockMigratorFeature{ctrl: ctrl}
	mock.recorder = &MockMigratorFeatureMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockMigratorFeature) EXPECT() *MockMigratorFeatureMockRecorder {
	return m.recorder
}

// Setup mocks base method
func (m *MockMigratorFeature) Setup(arg0, arg1, arg2 *jobsdb.HandleT, arg3, arg4 func()) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Setup", arg0, arg1, arg2, arg3, arg4)
}

// Setup indicates an expected call of Setup
func (mr *MockMigratorFeatureMockRecorder) Setup(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Setup", reflect.TypeOf((*MockMigratorFeature)(nil).Setup), arg0, arg1, arg2, arg3, arg4)
}

// MockWebhookFeature is a mock of WebhookFeature interface
type MockWebhookFeature struct {
	ctrl     *gomock.Controller
	recorder *MockWebhookFeatureMockRecorder
}

// MockWebhookFeatureMockRecorder is the mock recorder for MockWebhookFeature
type MockWebhookFeatureMockRecorder struct {
	mock *MockWebhookFeature
}

// NewMockWebhookFeature creates a new mock instance
func NewMockWebhookFeature(ctrl *gomock.Controller) *MockWebhookFeature {
	mock := &MockWebhookFeature{ctrl: ctrl}
	mock.recorder = &MockWebhookFeatureMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockWebhookFeature) EXPECT() *MockWebhookFeatureMockRecorder {
	return m.recorder
}

// Setup mocks base method
func (m *MockWebhookFeature) Setup(arg0 types.GatewayWebhookI) types.WebHookI {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Setup", arg0)
	ret0, _ := ret[0].(types.WebHookI)
	return ret0
}

// Setup indicates an expected call of Setup
func (mr *MockWebhookFeatureMockRecorder) Setup(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Setup", reflect.TypeOf((*MockWebhookFeature)(nil).Setup), arg0)
}
