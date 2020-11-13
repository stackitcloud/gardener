// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/gardener/gardener/pkg/client/kubernetes/clientmap (interfaces: ClientMap)

// Package clientmap is a generated GoMock package.
package clientmap

import (
	context "context"
	reflect "reflect"

	kubernetes "github.com/gardener/gardener/pkg/client/kubernetes"
	clientmap "github.com/gardener/gardener/pkg/client/kubernetes/clientmap"
	gomock "github.com/golang/mock/gomock"
)

// MockClientMap is a mock of ClientMap interface.
type MockClientMap struct {
	ctrl     *gomock.Controller
	recorder *MockClientMapMockRecorder
}

// MockClientMapMockRecorder is the mock recorder for MockClientMap.
type MockClientMapMockRecorder struct {
	mock *MockClientMap
}

// NewMockClientMap creates a new mock instance.
func NewMockClientMap(ctrl *gomock.Controller) *MockClientMap {
	mock := &MockClientMap{ctrl: ctrl}
	mock.recorder = &MockClientMapMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClientMap) EXPECT() *MockClientMapMockRecorder {
	return m.recorder
}

// GetClient mocks base method.
func (m *MockClientMap) GetClient(arg0 context.Context, arg1 clientmap.ClientSetKey) (kubernetes.Interface, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetClient", arg0, arg1)
	ret0, _ := ret[0].(kubernetes.Interface)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetClient indicates an expected call of GetClient.
func (mr *MockClientMapMockRecorder) GetClient(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetClient", reflect.TypeOf((*MockClientMap)(nil).GetClient), arg0, arg1)
}

// InvalidateClient mocks base method.
func (m *MockClientMap) InvalidateClient(arg0 clientmap.ClientSetKey) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InvalidateClient", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// InvalidateClient indicates an expected call of InvalidateClient.
func (mr *MockClientMapMockRecorder) InvalidateClient(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InvalidateClient", reflect.TypeOf((*MockClientMap)(nil).InvalidateClient), arg0)
}

// Start mocks base method.
func (m *MockClientMap) Start(arg0 <-chan struct{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Start", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Start indicates an expected call of Start.
func (mr *MockClientMapMockRecorder) Start(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockClientMap)(nil).Start), arg0)
}
