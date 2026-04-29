package validators

import (
	"context"

	"github.com/monetr/validation"
	"github.com/pkg/errors"
)

type inlineRule[T any] struct {
	f func(ctx context.Context, value T) error
}

// ValidateWithContext implements [validation.Rule].
func (i *inlineRule[T]) Validate(value any) error {
	return i.f(context.Background(), value.(T))
}

// ValidateWithContext implements [validation.RuleWithContext].
func (i *inlineRule[T]) ValidateWithContext(ctx context.Context, value any) error {
	return i.f(ctx, value.(T))
}

func By[T any](callback func(ctx context.Context, value T) error) validation.Rule {
	return &inlineRule[T]{
		f: callback,
	}
}

func Unique[T comparable]() validation.Rule {
	return By(func(ctx context.Context, fields []T) error {
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
