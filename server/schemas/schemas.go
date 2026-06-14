package schemas

import (
	"context"
	"encoding/json"
	"io"
	"regexp"

	locale "github.com/elliotcourant/go-lclocale"
	"github.com/monetr/monetr/server/merge"
	"github.com/monetr/validation"
	"github.com/monetr/validation/is"
	"github.com/pkg/errors"
)

func Parse[T any](
	ctx context.Context,
	reader io.Reader,
	baseData *T,
	schema validation.Rule,
) (*T, error) {
	rawData := map[string]any{}
	decoder := json.NewDecoder(reader)
	decoder.UseNumber()
	if err := decoder.Decode(&rawData); err != nil {
		return nil, errors.WithStack(err)
	}

	if err := validation.ValidateWithContext(
		ctx,
		&rawData,
		schema,
	); err != nil {
		return nil, err
	}

	var output T
	if baseData != nil {
		output = *baseData
	}
	if err := merge.Merge(
		&output, rawData, merge.ErrorOnUnknownField,
	); err != nil {
		return nil, errors.Wrap(err, "failed to merge patched data")
	}

	return &output, nil
}

type OptionalOrRequire = bool

var (
	Require  OptionalOrRequire = true
	Optional OptionalOrRequire = false
)

func Name(required OptionalOrRequire) *validation.KeyRules[string] {
	return validation.Key(
		"name",
		validation.Required.When(required).Error("Name is required"),
		validation.IsString,
		is.PrintableUnicode,
		validation.Length(1, 300).Error("Name must be between 1 and 300 characters"),
	).Required(required)
}

func Mask() validation.Rule {
	return validation.AllOf(
		validation.IsString,
		validation.Length(4, 4).Error("Mask must be exactly 4 digits"),
		validation.Match(regexp.MustCompile(`\d{4}`)).Error("Mask must be a 4 digit string"),
	)
}

func CurrencyCode() validation.Rule {
	return validation.AllOf(
		validation.IsString,
		validation.Length(3, 3).Error("Currency must be exactly 3 characters long"),
		is.Alpha.Error("Currency must be alphabetical characters only"),
		is.UpperCase.Error("Currency must be all upper case"),
		validation.In(
			locale.GetInstalledCurrencies()...,
		).Error("Currency must be one supported by the server"),
	)
}
