package crumbs

import (
	"context"
	"testing"

	"github.com/getsentry/sentry-go"
	"github.com/stretchr/testify/assert"
)

type structTest struct {

}

func (s *structTest) Foo(ctx context.Context) *sentry.Span {
	span := StartFnTrace(ctx)
	defer span.Finish()
	return span
}

func TestStartFnTrace(t *testing.T) {
	span := (&structTest{}).Foo(context.Background())
	assert.NotEmpty(t, span.Description)
}
