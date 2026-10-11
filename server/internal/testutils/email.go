package testutils

import (
	"fmt"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
)

func GetUniqueEmail(_ *testing.T) string {
	return fmt.Sprintf("%s@monetr.mini", strings.ReplaceAll(gofakeit.UUID(), "-", ""))
}
