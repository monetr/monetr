package validators

import (
	"context"
	"fmt"
	"reflect"

	"github.com/monetr/validation"
)

type EqRule interface {
	validation.Rule
	validation.RuleWithContext
}

type eqRule[T any] struct {
	value T
}

// Validate implements [EqRule].
func (e *eqRule[T]) Validate(value any) error {
	return e.ValidateWithContext(context.Background(), value)
}

// ValidateWithContext implements [EqRule].
func (e *eqRule[T]) ValidateWithContext(ctx context.Context, value any) error {
	if reflect.DeepEqual(value, e.value) {
		return nil
	}

	return validation.NewError("validation_eq_invalid", fmt.Sprintf("must equal \"%v\"", e.value))
}

func Eq[T any](value T) EqRule {
	return &eqRule[T]{
		value,
	}
}
