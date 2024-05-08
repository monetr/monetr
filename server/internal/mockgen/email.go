// Code generated by MockGen. DO NOT EDIT.
// Source: email.go
//
// Generated by this command:
//
//	mockgen -source=email.go -package=mockgen -destination=../internal/mockgen/email.go EmailCommunication
//

// Package mockgen is a generated GoMock package.
package mockgen

import (
	context "context"
	reflect "reflect"

	communication "github.com/monetr/monetr/server/communication"
	gomock "go.uber.org/mock/gomock"
)

// MockEmailCommunication is a mock of EmailCommunication interface.
type MockEmailCommunication struct {
	ctrl     *gomock.Controller
	recorder *MockEmailCommunicationMockRecorder
}

// MockEmailCommunicationMockRecorder is the mock recorder for MockEmailCommunication.
type MockEmailCommunicationMockRecorder struct {
	mock *MockEmailCommunication
}

// NewMockEmailCommunication creates a new mock instance.
func NewMockEmailCommunication(ctrl *gomock.Controller) *MockEmailCommunication {
	mock := &MockEmailCommunication{ctrl: ctrl}
	mock.recorder = &MockEmailCommunicationMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockEmailCommunication) EXPECT() *MockEmailCommunicationMockRecorder {
	return m.recorder
}

// SendPasswordChanged mocks base method.
func (m *MockEmailCommunication) SendPasswordChanged(ctx context.Context, params communication.PasswordChangedParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendPasswordChanged", ctx, params)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendPasswordChanged indicates an expected call of SendPasswordChanged.
func (mr *MockEmailCommunicationMockRecorder) SendPasswordChanged(ctx, params any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendPasswordChanged", reflect.TypeOf((*MockEmailCommunication)(nil).SendPasswordChanged), ctx, params)
}

// SendPasswordReset mocks base method.
func (m *MockEmailCommunication) SendPasswordReset(ctx context.Context, params communication.PasswordResetParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendPasswordReset", ctx, params)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendPasswordReset indicates an expected call of SendPasswordReset.
func (mr *MockEmailCommunicationMockRecorder) SendPasswordReset(ctx, params any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendPasswordReset", reflect.TypeOf((*MockEmailCommunication)(nil).SendPasswordReset), ctx, params)
}

// SendVerification mocks base method.
func (m *MockEmailCommunication) SendVerification(ctx context.Context, params communication.VerifyEmailParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendVerification", ctx, params)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendVerification indicates an expected call of SendVerification.
func (mr *MockEmailCommunicationMockRecorder) SendVerification(ctx, params any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendVerification", reflect.TypeOf((*MockEmailCommunication)(nil).SendVerification), ctx, params)
}
