package testutils

import (
	"crypto/rand"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func MustGenerateRandomString(t *testing.T, n int) string {
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		require.NoError(t, err, "must generate random int for creating string")
		ret[i] = letters[num.Int64()]
	}

	return string(ret)
}
