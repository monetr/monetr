package myownsanity

import (
	"crypto/sha256"
	"encoding/hex"
)

func SHA256(input []byte) string {
	sha := sha256.Sum256(input)
	return hex.EncodeToString(sha[:])
}
