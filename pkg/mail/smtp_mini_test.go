//go:build mini

package mail

import (
	"context"
	"fmt"
	"testing"

	"github.com/monetr/monetr/pkg/config"
	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestSmtpCommunication_Send(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		mail := NewSMTPCommunication(testutils.GetLog(t), config.SMTPClient{
			Username: "restapi",
			Password: "mailpassword",
			Host:     "mail.default.svc.cluster.local",
			Port:     1025,
		})

		err := mail.Send(context.Background(), SendEmailRequest{
			From:    fmt.Sprintf("no-reply@%s", "monetr.mini"),
			To:      "test@monetr.mini",
			Subject: "Verify your email address",
			IsHTML:  true,
			Content: "<html><body><h1>Hello World!</h1></body></html>",
		})
		assert.NoError(t, err, "should succeed")
	})
}
