// Code generated by MockGen. DO NOT EDIT.
// Source: key_management.go
//
// Generated by this command:
//
//	mockgen -source=key_management.go -package=mockgen -destination=../internal/mockgen/key_management.go KeyManagement
//

// Package mockgen is a generated GoMock package.
package mockgen

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockKeyManagement is a mock of KeyManagement interface.
type MockKeyManagement struct {
	ctrl     *gomock.Controller
	recorder *MockKeyManagementMockRecorder
	isgomock struct{}
}

// MockKeyManagementMockRecorder is the mock recorder for MockKeyManagement.
type MockKeyManagementMockRecorder struct {
	mock *MockKeyManagement
}

// NewMockKeyManagement creates a new mock instance.
func NewMockKeyManagement(ctrl *gomock.Controller) *MockKeyManagement {
	mock := &MockKeyManagement{ctrl: ctrl}
	mock.recorder = &MockKeyManagementMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockKeyManagement) EXPECT() *MockKeyManagementMockRecorder {
	return m.recorder
}

// Decrypt mocks base method.
func (m *MockKeyManagement) Decrypt(ctx context.Context, keyID, version *string, input string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Decrypt", ctx, keyID, version, input)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Decrypt indicates an expected call of Decrypt.
func (mr *MockKeyManagementMockRecorder) Decrypt(ctx, keyID, version, input any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Decrypt", reflect.TypeOf((*MockKeyManagement)(nil).Decrypt), ctx, keyID, version, input)
}

// Encrypt mocks base method.
func (m *MockKeyManagement) Encrypt(ctx context.Context, input string) (*string, *string, string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Encrypt", ctx, input)
	ret0, _ := ret[0].(*string)
	ret1, _ := ret[1].(*string)
	ret2, _ := ret[2].(string)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// Encrypt indicates an expected call of Encrypt.
func (mr *MockKeyManagementMockRecorder) Encrypt(ctx, input any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Encrypt", reflect.TypeOf((*MockKeyManagement)(nil).Encrypt), ctx, input)
}
