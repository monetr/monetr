package schema

import (
	"context"
	"io"
	"time"

	"github.com/Oudwins/zog"
	"github.com/benbjohnson/clock"
)

type ParseContext struct {
	context.Context
	Timezone *time.Location
	Clock    *clock.Clock
}

func Parse[T any](
	schema *zog.StructSchema,
	reader io.Reader,
	dest *T,

) (T, map[string][]string) {
	return *new(T), nil
}
