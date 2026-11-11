package schemas_test

import (
	"strings"
	"testing"

	"github.com/monetr/monetr/server/schemas"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	t.Run("null body", func(t *testing.T) {
		// A null body must not skip validation, the login schema requires an email
		// and password so this would otherwise come back as an empty request.
		result, err := schemas.Parse[schemas.LoginRequest](
			t.Context(),
			strings.NewReader(`null`),
			nil,
			schemas.LoginSchema,
		)
		assert.Error(t, err, "a null body must be rejected")
		assert.ErrorIs(t, errors.Cause(err), schemas.ErrBodyMustBeObject, "must be rejected because the body is not an object")
		assert.Nil(t, result, "must not return a result for a null body")
	})
}
