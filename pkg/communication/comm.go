package communication

import (
	"github.com/monetrapp/rest-api/pkg/config"
	"github.com/pkg/errors"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type CommHelper struct {
	configuration config.SendGrid
}

func (c *CommHelper) SendMessage(m *mail.SGMailV3) error {
	request := sendgrid.GetRequest(c.configuration.APIKey, "/v3/mail/send", "https://api.sendgrid.com")
	request.Method = "POST"
	var Body = mail.GetRequestBody(m)
	request.Body = Body
	_, err := sendgrid.API(request)
	if err != nil {
		return errors.Wrap(err, "failed to send email via sendgrid")
	}

	return nil
}
