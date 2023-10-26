package testutils

import (
	"fmt"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
)

func GetUniqueEmail(t *testing.T) string {
	return fmt.Sprintf("%s@monetr.mini", strings.ReplaceAll(gofakeit.UUID(), "-", ""))
}
