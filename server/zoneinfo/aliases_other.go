//go:build !unix

package zoneinfo

import (
	"context"

	"github.com/sirupsen/logrus"
)

func LoadAliasesFromHost(ctx context.Context, log *logrus.Entry) {
	// no-op
}
