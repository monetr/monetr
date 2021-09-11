package mock_mail

import (
	"context"
	"github.com/monetr/rest-api/pkg/mail"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMockMailCommunication_Send(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		mockMail := NewMockMail()

		assert.Len(t, mockMail.Sent, 0, "should have 0 sent")

		err := mockMail.Send(context.Background(), mail.SendEmailRequest{
			From:    "test@test.com",
			To:      "example@example.com",
			Subject: "Test",
			Content: "Test",
			IsHTML:  false,
		})

		assert.NoError(t, err, "should not fail")
		assert.Len(t, mockMail.Sent, 1, "should have added request to list")
	})

	t.Run("none sent", func(t *testing.T) {
		mockMail := NewMockMail()

		mockMail.ShouldFail = true
		assert.Len(t, mockMail.Sent, 0, "should have 0 sent")

		err := mockMail.Send(context.Background(), mail.SendEmailRequest{
			From:    "test@test.com",
			To:      "example@example.com",
			Subject: "Test",
			Content: "Test",
			IsHTML:  false,
		})

		assert.Error(t, err)
		assert.Len(t, mockMail.Sent, 0, "should have 0 sent")
	})
}
