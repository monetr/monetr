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
	return By(func(ctx context.Context, fields *[]T) error {
		if fields == nil {
			return nil
		}
		seen := make(map[T]struct{}, len(*fields))
		for i, f := range *fields {
			if _, dup := seen[f]; dup {
				return errors.Errorf("fields[%d] is a duplicate of an earlier entry", i)
			}
			seen[f] = struct{}{}
		}
		return nil
	})
}
