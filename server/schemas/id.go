package schemas

import (
	"context"

	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/validators"
	"github.com/monetr/validation"
	"github.com/monetr/validation/is"
	"github.com/pkg/errors"
)

func ValidID[T models.Identifiable]() validation.Rule {
	prefix := (*new(T)).IdentityPrefix()
	return validation.AllOf(
		validation.IsString,
		validators.By(func(_ context.Context, value *any) error {
			if value == nil {
				return errors.Errorf("id does not match format %s_...", prefix)
			}
			switch value := (*value).(type) {
			case string:
				_, err := models.ParseID[T](value)
				if err != nil {
					return errors.Errorf("id does not match format %s_...", prefix)
				}

				return nil
			default:
				return errors.Errorf("id does not match format %s_...", prefix)
			}
		}),
		is.PrintableASCII,
		validation.Length(28, 32).Error("id should be between 28 and 32 characters"),
	)
}
