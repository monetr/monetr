package validators_test

import (
	"testing"

	"github.com/monetr/monetr/server/validators"
	"github.com/monetr/validation"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

// sampleStruct is just a stand-in type so the "failed to validate %T" wrap below
// has something to format.
type sampleStruct struct{}

func TestMarshalErrorTree(t *testing.T) {
	t.Run("nil returns nil", func(t *testing.T) {
		assert.Nil(t, validators.MarshalErrorTree(nil))
	})

	t.Run("strips pkg.errors wraps and produces a JSON-marshalable map", func(t *testing.T) {
		inner := validation.Errors{"name": errors.New("cannot be blank")}
		wrapped := errors.Wrapf(inner, "failed to validate %T", &sampleStruct{})
		got := validators.MarshalErrorTree(wrapped)
		assert.Equal(t, map[string]any{"name": "cannot be blank"}, got)
	})

	t.Run("walks nested validation.Errors with per-spec wraps", func(t *testing.T) {
		// This mirrors the production case: the outer Mapping.Validate wraps a
		// validation.Errors whose values are themselves wrapped (per-spec
		// IDSpec / DateSpec wraps). The walker has to unwrap at every level,
		// not just the outermost, so the inner field-level errors are reachable.
		innerSpec := validation.Errors{"kind": errors.New("cannot be blank")}
		wrappedInner := errors.Wrapf(innerSpec, "failed to validate inner")
		outer := validation.Errors{"id": wrappedInner}
		wrappedOuter := errors.Wrapf(outer, "failed to validate outer")

		got := validators.MarshalErrorTree(wrappedOuter)
		assert.Equal(t, map[string]any{
			"id": map[string]any{
				"kind": "cannot be blank",
			},
		}, got)
	})

	t.Run("OneOfError with one variant flattens to its inner errors", func(t *testing.T) {
		oe := validation.OneOfError{validation.Errors{"name": errors.New("blank")}}
		got := validators.MarshalErrorTree(oe)
		assert.Equal(t, map[string]any{"name": "blank"}, got)
	})

	t.Run("OneOfError with multiple variants becomes oneOf array", func(t *testing.T) {
		oe := validation.OneOfError{
			validation.Errors{"name": errors.New("must be a valid value")},
			validation.Errors{"derivedKind": errors.New("cannot be blank")},
		}
		got := validators.MarshalErrorTree(oe)
		assert.Equal(t, map[string]any{
			"oneOf": []any{
				map[string]any{"name": "must be a valid value"},
				map[string]any{"derivedKind": "cannot be blank"},
			},
		}, got)
	})

	t.Run("non-validation error becomes its string", func(t *testing.T) {
		got := validators.MarshalErrorTree(errors.New("something else"))
		assert.Equal(t, "something else", got)
	})
}
