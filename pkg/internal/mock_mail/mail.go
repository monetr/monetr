package mock_mail

import (
	"context"
	"github.com/monetr/rest-api/pkg/mail"
	"github.com/pkg/errors"
)

var (
	_ mail.Communication = &MockMailCommunication{}
)

type MockMailCommunication struct {
	ShouldFail bool
	Sent       []mail.SendEmailRequest
}

func NewMockMail() *MockMailCommunication {
	return &MockMailCommunication{
		ShouldFail: false,
		Sent:       make([]mail.SendEmailRequest, 0),
	}
}

func (m *MockMailCommunication) Send(ctx context.Context, request mail.SendEmailRequest) error {
	if m.ShouldFail {
		return errors.Errorf("cannot send email")
	}

	m.Sent = append(m.Sent, request)

	return nil
}
