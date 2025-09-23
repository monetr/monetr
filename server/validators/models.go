package validators

import (
	"math"
	"regexp"

	locale "github.com/elliotcourant/go-lclocale"
	"github.com/monetr/validation"
)

func Name(required bool) *validation.KeyRules {
	return validation.Key(
		"name",
		validation.Required.When(required).Error("Name is required"),
		validation.Length(1, 300).Error("Name must be between 1 and 300 characters"),
	).Optional()
}

func Mask() *validation.KeyRules {
	return validation.Key(
		"mask",
		validation.Match(regexp.MustCompile(`\d{4}`)).Error("Mask must be a 4 digit string"),
	).Optional()
}

func CurrencyCode(required bool) *validation.KeyRules {
	return validation.Key(
		"currency",
		validation.Required.When(required).Error("Currency is required"),
		validation.In(
			locale.GetInstalledCurrencies()...,
		).Error("Currency must be one supported by the server"),
	).Optional()
}

func LimitBalance(name string) *validation.KeyRules {
	return validation.Key(
		name,
		validation.Min(float64(0)).Error("Limit balance cannot be negative"),
	).Optional()
}

func Balance(name string) *validation.KeyRules {
	return validation.Key(
		name,
		validation.Max(math.MaxFloat64),
	).Optional()
}
