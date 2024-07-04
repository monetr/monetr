package mock_plaid

import (
	"fmt"
	"testing"

	"github.com/plaid/plaid-go/v26/plaid"
	"github.com/stretchr/testify/require"
)

func Path(t *testing.T, relative string) string {
	require.NotEmpty(t, relative, "relative url cannot be empty")
	return fmt.Sprintf("%s%s", plaid.Sandbox, relative)
}
