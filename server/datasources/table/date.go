package table

import (
	"context"
	"regexp"
	"strings"

	"github.com/monetr/monetr/server/validators"
	"github.com/monetr/validation"
	"github.com/pkg/errors"
)

type DateSpec struct {
	Fields []FieldRef `json:"fields"`
	// Format specified the date format to use for parsing dates from the CSV
	// file. These are stored in a format like `YYYY-MM-DD` or `MM/DD/YYYY` or
	// something of the sort, and then converted to Go's date format when the file
	// is actually processed.
	Format string `json:"format"`
}

func (s *DateSpec) Validate(ctx context.Context) error {
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
				&s.Format,
				// We want to be able to convert a human readable date format such as
				// `YYYY-MM-DD` into golangs date format such as `2006-01-02`.
				// Validate that the year format, YY or YYYY can only happen once in the
				// string.
				validation.Match(regexp.MustCompile(`^[^Y]*(?:YYYY|YY){1}[^Y]*$`)).
					Error("Date format does not include the year"),
				// Validate that they month format, M or MM can only happen once in the
				// string.
				validation.Match(regexp.MustCompile(`^[^M]*(?:MM|M){1}[^M]*$`)).
					Error("Date format does not include the month"),
				// Validate that the day format, D or DD can only happen once in the
				// string.
				validation.Match(regexp.MustCompile(`^[^D]*(?:DD|D){1}[^D]*$`)).
					Error("Date format does not include the day of the month"),
				// Validate that only these characters are allowed in the string at all.
				validation.Match(regexp.MustCompile(`^(Y|M|D|\.| |-|/)+$`)).
					Error("Date format contains invalid characters"),
				validation.Length(6, 10),
				validation.Required,
			),
		),
		"failed to validate %T",
		s,
	)
}

func (s *DateSpec) GetTimeFormat() string {
	replacements := [][2]string{
		{
			"YYYY", "2006",
		},
		{
			"YY", "06",
		},
		{
			"MM", "01",
		},
		{
			"M", "1",
		},
		{
			"DD", "02",
		},
		{
			"D", "2",
		},
	}
	format := s.Format
	for _, replacement := range replacements {
		format = strings.Replace(format, replacement[0], replacement[1], 1)
	}
	return format
}
