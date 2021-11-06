package ctxkeys

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestLogrusFieldsFromContext(t *testing.T) {
	t.Run("field does not already exist", func(t *testing.T) {
		inputFields := logrus.Fields{
			"linkId": uint64(1234),
		}

		ctx := context.WithValue(context.Background(), AccountID, uint64(5678))

		fieldsToBeAdded := LogrusFieldsFromContext(ctx, inputFields)
		assert.Len(t, fieldsToBeAdded, 1, "there should be one field to be added")
		assert.EqualValues(t, logrus.Fields{
			"accountId": uint64(5678),
		}, fieldsToBeAdded, "fields to be added should have the account Id we stored on the context")
	})

	t.Run("field should not be overwritten", func(t *testing.T) {
		inputFields := logrus.Fields{
			"accountId": uint64(7654),
			"linkId":    uint64(1234),
		}

		// Note that the accountId we are passing here is different from the one on the inputFields. This allows us to
		// properly assert that we are not replacing the value
		ctx := context.WithValue(context.Background(), AccountID, uint64(5678))
		ctx = context.WithValue(context.Background(), UserID, uint64(9876))

		fieldsToBeAdded := LogrusFieldsFromContext(ctx, inputFields)
		assert.Len(t, fieldsToBeAdded, 1, "there should be one field to be added")
		assert.EqualValues(t, logrus.Fields{
			"userId": uint64(9876),
		}, fieldsToBeAdded, "only the userId field should be present on the resulting fields")
	})
}
