package validators_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/monetr/monetr/server/validators"
	"github.com/monetr/validation"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOneOfError_MarshalJSON(t *testing.T) {
	t.Run("single schema unwraps to a flat validation.Errors object", func(t *testing.T) {
		// When OneOf or OneOfStruct is used with a single schema (the common case
		// for an endpoint that only ever validates against one shape), the JSON
		// shape should be a plain {field: message} object so existing API
		// consumers see a flat error map. This keeps the wire format compatible
		// with non-union endpoints.
		oe := validators.OneOfError{
			validation.Errors{
				"name": errors.New("Name must be between 1 and 300 characters"),
			},
		}
		raw, err := json.Marshal(oe)
		require.NoError(t, err)

		var decoded map[string]string
		require.NoError(t, json.Unmarshal(raw, &decoded))
		assert.Equal(t, map[string]string{
			"name": "Name must be between 1 and 300 characters",
		}, decoded)
	})

	t.Run("multiple schemas wrap in oneOf array", func(t *testing.T) {
		oe := validators.OneOfError{
			validation.Errors{
				"name": errors.New("cannot be blank"),
			},
			validation.Errors{
				"derivedKind": errors.New("cannot be blank"),
			},
		}
		raw, err := json.Marshal(oe)
		require.NoError(t, err)

		var decoded struct {
			OneOf []map[string]string `json:"oneOf"`
		}
		require.NoError(t, json.Unmarshal(raw, &decoded))
		assert.Equal(t, []map[string]string{
			{"name": "cannot be blank"},
			{"derivedKind": "cannot be blank"},
		}, decoded.OneOf)
	})

	t.Run("nested inside a validation.Errors recurses", func(t *testing.T) {
		// This is the load-bearing case: a OneOfError sitting as the value of a
		// field inside a parent validation.Errors must serialize structurally
		// when the parent is marshalled, not collapse to a string.
		// validation.Errors.MarshalJSON delegates to nested json.Marshaler
		// values, which is what makes this work.
		parent := validation.Errors{
			"memo": validators.OneOfError{
				validation.Errors{
					"name":        errors.New("must be a valid value"),
					"derivedKind": errors.New("must be blank"),
				},
				validation.Errors{
					"name":        errors.New("must be blank"),
					"derivedKind": errors.New("cannot be blank"),
				},
			},
		}
		raw, err := json.Marshal(parent)
		require.NoError(t, err)

		var decoded map[string]struct {
			OneOf []map[string]string `json:"oneOf"`
		}
		require.NoError(t, json.Unmarshal(raw, &decoded))
		require.Contains(t, decoded, "memo")
		assert.Equal(t, []map[string]string{
			{"name": "must be a valid value", "derivedKind": "must be blank"},
			{"name": "must be blank", "derivedKind": "cannot be blank"},
		}, decoded["memo"].OneOf)
	})
}

type sampleStruct struct {
	Kind   string
	Fields []string
}

func TestOneOfStruct_ReturnsOneOfError(t *testing.T) {
	// OneOfStruct must return a typed OneOfError when no alternative validates,
	// not an errors.Join wrapper. Returning the typed error is what lets the
	// caller's type switch and the JSON marshaller find the union structure.
	s := sampleStruct{Kind: "neither"}
	err := validators.OneOfStruct(
		context.Background(),
		&s,
		[]*validation.FieldRules{
			validation.Field(&s.Kind, validation.Required, validation.In("alpha")),
		},
		[]*validation.FieldRules{
			validation.Field(&s.Kind, validation.Required, validation.In("beta")),
		},
	)
	require.Error(t, err)
	oe, ok := err.(validators.OneOfError)
	require.True(t, ok, "expected OneOfError, got %T", err)
	assert.Len(t, oe, 2, "one entry per attempted schema")
}

func TestOneOfStruct_NilOnFirstSuccess(t *testing.T) {
	// Short-circuit on first valid schema: subsequent schemas are not consulted.
	s := sampleStruct{Kind: "alpha"}
	err := validators.OneOfStruct(
		context.Background(),
		&s,
		[]*validation.FieldRules{
			validation.Field(&s.Kind, validation.Required, validation.In("alpha")),
		},
		[]*validation.FieldRules{
			validation.Field(&s.Kind, validation.Required, validation.In("beta")),
		},
	)
	assert.NoError(t, err)
}

func TestMarshalErrorTree(t *testing.T) {
	t.Run("nil returns nil", func(t *testing.T) {
		assert.Nil(t, validators.MarshalErrorTree(nil))
	})

	t.Run("strips pkg/errors wraps and produces a JSON-marshalable map", func(t *testing.T) {
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
		oe := validators.OneOfError{validation.Errors{"name": errors.New("blank")}}
		got := validators.MarshalErrorTree(oe)
		assert.Equal(t, map[string]any{"name": "blank"}, got)
	})

	t.Run("OneOfError with multiple variants becomes oneOf array", func(t *testing.T) {
		oe := validators.OneOfError{
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
