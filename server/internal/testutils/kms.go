package testutils

import (
	"testing"

	"github.com/monetr/monetr/server/secrets"
)

func GetKMS(t *testing.T) secrets.KeyManagement {
	return secrets.NewPlaintextKMS()
}
