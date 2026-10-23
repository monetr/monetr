package mock_plaid

import (
	"fmt"
	"testing"

	plaid "github.com/plaid/plaid-go/v49"
	"github.com/stretchr/testify/require"
)

func Path(t *testing.T, relative string) string {
	require.NotEmpty(t, relative, "relative url cannot be empty")
	return fmt.Sprintf("%s%s", plaid.Environments.Sandbox, relative)
}
