package mock_stripe

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

type StripeError struct {
	Error struct {
		Code    string `json:"code"`
		DocURL  string `json:"doc_url"`
		Message string `json:"message"`
		Param   string `json:"param"`
		Type    string `json:"type"`
	} `json:"error"`
}

func NewResourceMissingError(t *testing.T, id, object string) StripeError {
	require.NotEmpty(t, object, "object cannot be empty")
	return StripeError{
		Error: struct {
			Code    string `json:"code"`
			DocURL  string `json:"doc_url"`
			Message string `json:"message"`
			Param   string `json:"param"`
			Type    string `json:"type"`
		}{
			Code:    "resource_missing",
			DocURL:  "https://stripe.com/docs/error-codes/resource-missing",
			Message: fmt.Sprintf("No such %s: '%s'", object, id),
			Param:   object,
			Type:    "invalid_request_error",
		},
	}
}

func NewInternalServerError(t *testing.T, object string) StripeError {
	require.NotEmpty(t, object, "object cannot be empty")
	return StripeError{
		Error: struct {
			Code    string `json:"code"`
			DocURL  string `json:"doc_url"`
			Message string `json:"message"`
			Param   string `json:"param"`
			Type    string `json:"type"`
		}{
			Code:    "url_invalid",
			DocURL:  "https://stripe.com/docs/error-codes#url-invalid",
			Message: fmt.Sprintf("No such %s", object),
			Param:   object,
			Type:    "api_error",
		},
	}
}
