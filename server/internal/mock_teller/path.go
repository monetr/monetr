package mock_teller

import (
	"fmt"
	"testing"

	"github.com/monetr/monetr/server/teller"
	"github.com/stretchr/testify/require"
)

func Path(t *testing.T, relative string) string {
	require.NotEmpty(t, relative, "relative url cannot be empty")
	return fmt.Sprintf("https://%s%s", teller.APIHostname, relative)
}
