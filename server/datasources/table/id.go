package table

import (
	"context"

	"github.com/monetr/monetr/server/validators"
	"github.com/monetr/validation"
	"github.com/pkg/errors"
)

// IDSpecKind represents the method of uniquly identifying transactions in the
// provided CSV file.
type IDSpecKind string

const (
	// IDSpecKindNative means that a single field within the file represents the
	// unique identifier, and it can be trusted entirely for de-duplication.
	IDSpecKindNative IDSpecKind = "native"
	// IDSpecKindHashed means that a combination of multiple fields within the
	// file represent the unique identifier.
	IDSpecKindHashed IDSpecKind = "hashed"
)

// IDSpec is used to derive the unique identifier for a given transaction row.
// This generates a hash of the fields provided (including derived fields) in
// order to determine a unique identifier. This ultimately is fed into the
// [models.Transaction.UploadIdentifier] field. The [IDSpec.Kind] field
// indicates whether an existing field should be used plainly, or whether all of
// the specified fields (or derivatives) should be hashed.
type IDSpec struct {
	Kind   IDSpecKind `json:"kind"`
	Fields []FieldRef `json:"fields"`
}

func (s *IDSpec) Validate(ctx context.Context) error {
	return errors.Wrapf(
		validation.ValidateStructWithContext(
			ctx,
			s,
			validation.Field(
				&s.Kind,
				validators.In(
					IDSpecKindNative,
					IDSpecKindHashed,
				),
				validation.Required,
			),
			validation.Field(
				&s.Fields,
				validation.Each(
					validators.By(func(ctx context.Context, field FieldRef) error {
						return field.Validate(ctx)
					}),
				),
				// Reasonable upper bounds on number of fields.
				validation.Length(1, 20),
				validators.Unique[FieldRef](),
				validation.Required,
			),
		),
		"failed to validate %T",
		s,
	)
}
