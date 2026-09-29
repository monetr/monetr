package validators

import (
	"context"

	"github.com/monetr/validation"
	"github.com/pkg/errors"
)

type inlineRule[T any] struct {
	f func(ctx context.Context, value *T) error
}

// ValidateWithContext implements [validation.Rule].
func (i *inlineRule[T]) Validate(value any) error {
	return i.ValidateWithContext(context.Background(), value)
}

// ValidateWithContext implements [validation.RuleWithContext].
func (i *inlineRule[T]) ValidateWithContext(ctx context.Context, value any) error {
	switch v := value.(type) {
	case *T:
		// Pointer struct fields arrive already as a *T (the validation library
		// hands us the field value verbatim). Pass it through; it may be nil.
		return i.f(ctx, v)
	case T:
		// Value fields arrive as a T; wrap it so the callback always sees a *T.
		return i.f(ctx, &v)
	default:
		return i.f(ctx, nil)
	}
}

func By[T any](callback func(ctx context.Context, value *T) error) validation.Rule {
	return &inlineRule[T]{
		f: callback,
	}
}

func Unique[T comparable]() validation.Rule {
	return By(func(_ context.Context, value *any) error {
		if value == nil {
			return nil
		}
		var fields []T
		switch v := (*value).(type) {
		case []T:
			fields = v
		case *[]T:
			if v == nil {
				return nil
			}
			fields = *v
		case []any:
			// Raw JSON arrays decode as []any, assert each element to T so schemas
			// validating decoded request bodies still get checked.
			fields = make([]T, 0, len(v))
			for i, raw := range v {
				item, ok := raw.(T)
				if !ok {
					return errors.Errorf("fields[%d] is not a valid value", i)
				}
				fields = append(fields, item)
			}
		default:
			return nil
		}
		seen := make(map[T]struct{}, len(fields))
		for i, f := range fields {
			if _, dup := seen[f]; dup {
				return errors.Errorf("fields[%d] is a duplicate of an earlier entry", i)
			}
			seen[f] = struct{}{}
		}
		return nil
	})
}
