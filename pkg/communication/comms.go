package communication

import (
	"context"
	"github.com/monetr/rest-api/pkg/models"
)

type VerifyEmailParams struct {
	Login models.Login
}

type UserCommunication interface {
	SendVerificationEmail(ctx context.Context, params VerifyEmailParams) error
}
