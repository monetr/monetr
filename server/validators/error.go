package validators

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/monetr/validation"
	"github.com/pkg/errors"
)

type Problems map[string][]error

func (p Problems) Error() string {
	if len(p) == 0 {
		return ""
	}

	keys := make([]string, len(p))
	i := 0
	for key := range p {
		keys[i] = key
		i++
	}
	sort.Strings(keys)

	var s strings.Builder
	for i, key := range keys {
		if i > 0 {
			s.WriteString("; ")
		}
		_, _ = fmt.Fprintf(&s, "%v: ", key)
		for x, item := range p[key] {
			if x > 0 {
				s.WriteString(", ")
			}
			_, _ = fmt.Fprintf(&s, "%v", item.Error())
		}
	}
	s.WriteString(".")
	return s.String()
}

// OneOfStruct validates structPtr against several alternative schemas. The
// first schema that fully validates wins and OneOfStruct returns nil. If none
// match, OneOfStruct returns a [OneOfError] where each element is one schema's
// [validation.Errors] in input order. Variants are unlabeled; the per-rule
// failures inside each variant are expected to be self-discriminating (the
// kind-style fields will pass on exactly one variant when the input is
// internally consistent).
//
// If a custom rule returns a non-validation error (anything that's not a
// [validation.Errors]), OneOfStruct surfaces it directly so callers can
// distinguish unexpected errors from schema-mismatch failures.
func OneOfStruct[T any](
	ctx context.Context,
	structPtr *T,
	schemas ...[]*validation.FieldRules,
) error {
	failures := make(OneOfError, 0, len(schemas))
	for _, schema := range schemas {
		err := validation.ValidateStructWithContext(
			ctx,
			structPtr,
			schema...,
		)
		if err == nil {
			return nil
		}
		verrs, ok := err.(validation.Errors)
		if !ok {
			return err
		}
		failures = append(failures, verrs)
	}
	return failures
}

// MarshalErrorTree prepares an error returned from a [validation.Validate]
// chain or a [validators] helper for JSON serialization. It walks the error
// tree, stripping pkg/errors wraps at every level (each per-spec errors.Wrapf
// in datasources/table, plus the outermost Mapping wrap) so nested
// [validation.Errors] and [OneOfError] values are reached by the caller's
// encoding/json. Returns a JSON-marshalable value tree built from maps, slices,
// and strings.
//
// A walker (rather than relying on [json.Marshaler] delegation in
// [validation.Errors.MarshalJSON]) is required because that delegation only
// fires for direct [json.Marshaler] values; pkg/errors wraps in between would
// otherwise fall back to err.Error() and collapse a sub-tree into a flat
// string.
func MarshalErrorTree(err error) any {
	if err == nil {
		return nil
	}
	return walkError(err)
}

func walkError(err error) any {
	if err == nil {
		return nil
	}
	err = errors.Cause(err)
	switch e := err.(type) {
	case validation.Errors:
		out := make(map[string]any, len(e))
		for k, v := range e {
			out[k] = walkError(v)
		}
		return out
	case OneOfError:
		if len(e) == 1 {
			return walkError(e[0])
		}
		variants := make([]any, len(e))
		for i, v := range e {
			variants[i] = walkError(v)
		}
		return map[string]any{"oneOf": variants}
	default:
		return err.Error()
	}
}

// FlattenValidationError walks a validation error and produces a flat
// [Problems] map suitable for human-readable logs and string assertions. It is
// not the JSON marshaller; for API responses use [MarshalErrorTree] which
// preserves the nested structure (including [OneOfError] variants).
func FlattenValidationError(err error) error {
	switch err := errors.Cause(err).(type) {
	case validation.Errors:
		errs := make(Problems)
		for key, problem := range err {
			errs[key] = []error{FlattenValidationError(problem)}
		}
		return errs
	case interface {
		error
		Unwrap() []error
	}:
		// If we get nested merged errors then join them.
		errs := make(Problems)
		for _, item := range err.Unwrap() {
			switch item := errors.Cause(item).(type) {
			case validation.Errors:
				for key, problem := range item {
					errs[key] = append(errs[key], FlattenValidationError(problem))
				}
			default:
				return err
			}
		}
		return errs
	default:
		return err
	}
}
