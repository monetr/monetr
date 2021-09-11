package mock_plaid

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlaidHeaders(t *testing.T) {
	assert.NotPanics(t, func() {
		headers := PlaidHeaders(t, nil, nil, http.StatusOK)
		assert.NotEmpty(t, headers, "headers should not be empty")
		assert.Contains(t, headers, "X-Request-Id", "should contain the X-Request-Id header")
	}, "this method must not panic if we pass a nil request and response")
}
