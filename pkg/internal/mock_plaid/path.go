package mock_plaid

import (
	"fmt"
	"github.com/plaid/plaid-go/plaid"
	"github.com/stretchr/testify/require"
	"testing"
)

func Path(t *testing.T, relative string) string {
	require.NotEmpty(t, relative, "relative url cannot be empty")
	return fmt.Sprintf("%s%s", plaid.Development, relative)
}
