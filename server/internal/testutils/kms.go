package testutils

import (
	"testing"

	"github.com/monetr/monetr/server/secrets"
)

func GetKMS(_ *testing.T) secrets.KeyManagement {
	return secrets.NewPlaintextKMS()
}
