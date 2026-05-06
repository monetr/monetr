package table

import (
	"context"

	"github.com/monetr/monetr/server/validators"
	"github.com/monetr/validation"
	"github.com/pkg/errors"
)

// PostedSpec determines whether the current transaction is in a posted or
// pending status. It does this by making a case-insensitive comparison against
// the field specified. When the field matches the posted value provided it is
// considered posted, any other value is considered pending. When this struct is
// nil on the [Mapping] then all transactions are considered posted.
type PostedSpec struct {
	Fields []FieldRef `json:"fields"`
	// Posted is the string we look for in the field specified to know if the
	// transaction is posted. Any other value is considered pending.
	Posted string `json:"posted,omitempty"`
}

func (s *PostedSpec) Validate(ctx context.Context) error {
	return errors.Wrapf(
		validation.ValidateStructWithContext(
			ctx,
			s,
			validation.Field(
				&s.Fields,
				validation.Each(
					validators.By(func(ctx context.Context, field FieldRef) error {
						return field.Validate(ctx)
					}),
				),
				validation.Length(1, 1),
				validators.Unique[FieldRef](),
				validation.Required,
			),
			validation.Field(
				&s.Posted,
				validators.PrintableUnicode,
				validation.Length(1, 100),
			),
		),
		"failed to validate %T",
		s,
	)
}
