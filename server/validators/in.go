package validators

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/validation"
)

type InRule interface {
	validation.Rule
	validation.RuleWithContext
}

type inRule[T any] struct {
	values []T
}

// Validate implements [InRule].
func (i *inRule[T]) Validate(value any) error {
	return i.ValidateWithContext(context.Background(), value)
}

// ValidateWithContext implements [InRule]. Empty/nil values pass through so
// that a sibling [validation.Required] rule controls the "missing" path with
// "cannot be blank"; this rule speaks up only when the caller supplied a value
// that wasn't in the allowed set.
func (i *inRule[T]) ValidateWithContext(ctx context.Context, value any) error {
	value, isNil := validation.Indirect(value)
	if isNil || validation.IsEmpty(value) {
		return nil
	}
	for _, v := range i.values {
		if reflect.DeepEqual(value, v) {
			return nil
		}
	}

	return validation.NewError(
		"validation_in_invalid",
		fmt.Sprintf("must be one of: [%s]", strings.Join(
			myownsanity.Map(i.values, func(v T) string {
				return fmt.Sprintf("%q", any(v))
			}),
			", ",
		)),
	)
}

// In returns a rule that succeeds when the value matches one of the supplied
// alternatives. The failure message lists the alternatives as
// `must be one of: ["a", "b", ...]`, which is more actionable than
// [validation.In]'s default `must be a valid value`. Empty values are passed
// through; pair with [validation.Required] when the field is also required.
func In[T any](values ...T) InRule {
	return &inRule[T]{values: values}
}
