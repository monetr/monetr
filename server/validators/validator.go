package validators

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/merge"
	"github.com/monetr/validation"
	"github.com/pkg/errors"
)

var (
	_ error          = OneOfError{}
	_ json.Marshaler = OneOfError{}
)

// OneOfError is returned by [OneOf] and [OneOfStruct] when none of the
// alternative schemas validate. Each entry is the [validation.Errors] produced
// by one schema attempt, in the order the schemas were provided. Variants are
// unlabeled on purpose: the per-rule errors inside each entry (in particular
// the kind-discriminator field's failure or pass) tell the caller which
// variant was being attempted.
type OneOfError []validation.Errors

// Error implements [error].
func (o OneOfError) Error() string {
	return fmt.Sprintf(
		"input must be considered valid by: %s",
		strings.Join(myownsanity.Map(o, func(err validation.Errors) string {
			return err.Error()
		}), " or "),
	)
}

// MarshalJSON serializes a [OneOfError] for API responses. When there is a
// single alternative (the common case for an endpoint that's only ever
// validated against one schema, even though it goes through OneOf), the inner
// [validation.Errors] is emitted directly so callers see a flat
// {"field": "message"} object. When there are multiple alternatives (the
// discriminated-union case), the output is {"oneOf": [<errs>, ...]}; nested
// unions work because [validation.Errors.MarshalJSON] delegates to any nested
// [json.Marshaler] values.
func (o OneOfError) MarshalJSON() ([]byte, error) {
	if len(o) == 1 {
		return json.Marshal(o[0])
	}
	return json.Marshal(map[string]any{
		"oneOf": []validation.Errors(o),
	})
}

// OneOf takes multiple map rules and combines them into an "or" type schema. At
// least one of the map rules provided must be considered valid. The first valid
// map rule for the provided input is the one that is parsed, merged and
// returned. If none of the rules are valid then an error is returned. If more
// than one rule is valid, only the first valid rule is considered. Schemas
// should be written such that they are mutually exclusive.
// The input data is parsed over the existing data provided, but the existing
// data is copied first such that a copy is returned with the valid and parsed
// data returned. If the input is not valid then a simple copy of the existing
// data is returned along with an error.
func OneOf[T any](
	ctx context.Context,
	existing *T,
	input map[string]any,
	schemas ...validation.MapRule,
) (T, error) {
	if existing == nil {
		existing = new(T)
	}
	errs := make(OneOfError, len(schemas))
	for i, schema := range schemas {
		err := validation.ValidateWithContext(
			ctx,
			&input,
			schema,
		)
		switch err := err.(type) {
		case validation.Errors:
			errs[i] = err
		case nil:
			// On the first schema that is considered valid we return the parsed
			// object from it.
			output := *existing
			if err := merge.Merge(
				&output, input, merge.ErrorOnUnknownField,
			); err != nil {
				return *existing, errors.Wrap(err, "failed to merge patched data")
			}
			return output, nil
		default:
			return *existing, errors.Wrap(err, "failed to validate schema")
		}
	}

	return *existing, errors.WithStack(errs)
}
