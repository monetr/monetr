//go:build !unix

package zoneinfo

import (
	"context"
	"log/slog"
)

func LoadAliasesFromHost(ctx context.Context, log *slog.Logger) {
	// no-op
}
