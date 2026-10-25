package mock_simplefin

import (
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
)

// MockClaimSetupToken registers the claim URL and returns the provided access
// URL as plain text, which is how SimpleFIN responds to a claim.
func MockClaimSetupToken(t *testing.T, claimUrl string, accessUrl string) {
	require.NotEmpty(t, claimUrl, "claim url cannot be empty")
	httpmock.RegisterResponder(
		http.MethodPost,
		claimUrl,
		httpmock.NewStringResponder(http.StatusOK, accessUrl),
	)
}

// MockClaimSetupTokenAlreadyClaimed registers the claim URL but responds the
// way SimpleFIN does when the token has already been used.
func MockClaimSetupTokenAlreadyClaimed(t *testing.T, claimUrl string) {
	require.NotEmpty(t, claimUrl, "claim url cannot be empty")
	httpmock.RegisterResponder(
		http.MethodPost,
		claimUrl,
		httpmock.NewStringResponder(http.StatusForbidden, "Forbidden"),
	)
}
