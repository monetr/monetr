package validators

import (
	"math"
	"regexp"

	locale "github.com/elliotcourant/go-lclocale"
	"github.com/monetr/validation"
	"github.com/monetr/validation/is"
)

type OptionalOrRequire = bool

var (
	Require  OptionalOrRequire = true
	Optional OptionalOrRequire = false
)

func Name(required OptionalOrRequire) *validation.KeyRules[string] {
	return validation.Key(
		"name",
		validation.Required.When(required).Error("Name is required"),
		is.PrintableUnicode,
		validation.Length(1, 300).Error("Name must be between 1 and 300 characters"),
	).Required(required)
}

// Description is a shorthand for description fields on object.
// Deprecated: Use a custom rule instead.
func Description() *validation.KeyRules[string] {
	return validation.Key(
		"description",
		is.PrintableUnicode,
		validation.Length(1, 300).Error("Description must be between 1 and 300 characters"),
	).Required(Optional)
}

func Mask() *validation.KeyRules[string] {
	return validation.Key(
		"mask",
		validation.Match(regexp.MustCompile(`\d{4}`)).Error("Mask must be a 4 digit string"),
	).Required(Optional)
}

func CurrencyCode(required OptionalOrRequire) *validation.KeyRules[string] {
	return validation.Key(
		"currency",
		validation.Required.When(required).Error("Currency is required"),
		validation.In(
			locale.GetInstalledCurrencies()...,
		).Error("Currency must be one supported by the server"),
	).Required(required)
}

func LimitBalance(name string) *validation.KeyRules[string] {
	return validation.Key(
		name,
		validation.Min(float64(0)).Error("Limit balance cannot be negative"),
	).Required(Optional)
}

func Balance(name string) *validation.KeyRules[string] {
	return validation.Key(
		name,
		validation.Max(math.MaxFloat64),
	).Required(Optional)
}
