package schema

import (
	"context"
	"encoding/json"
	"io"
	"time"

	"github.com/Oudwins/zog"
	"github.com/Oudwins/zog/pkgs/internals"
	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/pkg/errors"
)

type ParseMetadata struct {
	Clock    clock.Clock
	Timezone *time.Location
}

func Parse[T any](
	ctx context.Context,
	schema zog.ComplexZogSchema,
	base *T,
	reader io.Reader,
	metadata ParseMetadata,
) (T, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	myownsanity.ASSERT_NOTNIL(base, "base cannot be nil for parsing!")

	var m map[string]any
	if err := json.NewDecoder(reader).Decode(&m); err != nil {
		return *base, errors.Wrapf(err, "failed to parse data for %T schema", *base)
	}

	dp := internals.NewMapDataProvider(m, myownsanity.Pointer("json"))
	result := *base

	issues := schema.Parse(
		dp,
		&result,
		zog.WithCtxValue("clock", metadata.Clock),
		zog.WithCtxValue("timezone", metadata.Timezone),
		WithContext(span.Context()),
	)
	if len(issues) > 0 {
		return result, errors.WithStack(NewIssueError(issues))
	}

	return result, nil
}
