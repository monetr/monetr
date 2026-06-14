package validators

import (
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

// MarshalErrorTree prepares an error returned from a [validation.Validate]
// chain or a [validators] helper for JSON serialization. It walks the error
// tree, stripping pkg/errors wraps at every level (each per-spec errors.Wrapf
// in datasources/table, plus the outermost Mapping wrap) so nested
// [validation.Errors] and [validation.OneOfError] values are reached by the caller's
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
	case validation.OneOfError:
		// A union from [validation.OneOf] / [validation.MatchOneOfStruct]. A single
		// alternative flattens to its inner errors, multiple alternatives become a
		// {"oneOf": [...]} envelope so the client can see each shape it could match.
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
// preserves the nested structure (including [validation.OneOfError] variants).
func FlattenValidationError(err error) error {
	switch err := errors.Cause(err).(type) {
	case validation.Errors:
		errs := make(Problems)
		for key, problem := range err {
			errs[key] = []error{FlattenValidationError(problem)}
		}
		return errs
	case validation.OneOfError:
		// A union is a set of mutually exclusive alternatives, it does not flatten into
		// a single map. Return it as is so its own "must match one of: (...) or (...)"
		// rendering is used rather than letting the generic Unwrap case below merge the
		// variants into one map.
		return err
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
