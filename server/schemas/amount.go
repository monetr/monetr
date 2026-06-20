package schemas

import (
	"context"
	"encoding/json"

	"github.com/monetr/monetr/server/validators"
	"github.com/monetr/validation"
	"github.com/monetr/validation/is"
	"github.com/pkg/errors"
)

func Amount() validation.Rule {
	return validation.AllOf(
		is.Integer,
		validators.By[any](func(_ context.Context, value *any) error {
			if value == nil {
				return errors.New("Amount is not a valid integer")
			}
			switch value := (*value).(type) {
			case json.Number:
				if jint, err := value.Int64(); err != nil {
					return errors.New("Amount is not a valid integer")
				} else if jint == 0 {
					return errors.New("Amount cannot be zero")
				}

				return nil
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
				if value == 0 {
					return errors.New("Amount cannot be zero")
				}
				return nil
			default:
				return errors.New("Amount is not a valid integer")
			}
		}),
	)
}

func PositiveAmount(prefix string) validation.Rule {
	return validation.AllOf(
		is.Integer,
		validators.By[any](func(_ context.Context, value *any) error {
			if value == nil {
				return errors.Errorf("%s is not a valid integer", prefix)
			}
			switch value := (*value).(type) {
			case json.Number:
				if jint, err := value.Int64(); err != nil {
					return errors.Errorf("%s is not a valid integer", prefix)
				} else if jint <= 0 {
					return errors.Errorf("%s must be greater than zero", prefix)
				}

				return nil
			default:
				return errors.Errorf("%s is not a valid integer", prefix)
			}
		}),
	)
}
