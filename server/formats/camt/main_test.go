package camt_test

import (
	"testing"

	"github.com/monetr/monetr/server/internal/fixtures"
)

func GetFixtures(t *testing.T, name string) []byte {
	return fixtures.LoadFile(t, name)
}
