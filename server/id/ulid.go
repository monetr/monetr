package id

import (
	"crypto/rand"
	"io"
	"strings"
	"sync"

	"github.com/oklog/ulid/v2"
)

var (
	entropy     io.Reader
	entropyOnce sync.Once
)

func cryptoEntropy() io.Reader {
	entropyOnce.Do(func() {
		entropy = &ulid.LockedMonotonicReader{
			MonotonicReader: ulid.Monotonic(rand.Reader, 0),
		}
	})
	return entropy
}

func New() string {
	return strings.ToLower(ulid.MustNew(ulid.Now(), cryptoEntropy()).String())
}
