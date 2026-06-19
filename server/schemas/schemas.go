package schemas

import (
	"context"
	"encoding/json"
	"io"
	"regexp"
	"strings"

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

	cleanStrings(rawData)

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

// cleanStrings is a recursive function that takes the json body as a map and
// trims any string fields specified on it. This is so that the validation code
// doesnt get tripped up by any dumb whitespace provided.
func cleanStrings(input map[string]any) {
	for key, value := range input {
		switch value := value.(type) {
		case string:
			input[key] = strings.TrimSpace(value)
		case map[string]any:
			cleanStrings(value)
			// TODO Handle arrays?
		}
	}
}

type OptionalOrRequire = bool

var (
	Require  OptionalOrRequire = true
	Optional OptionalOrRequire = false
)

// NameOld is a shorthand.
//
// Deprecated: Use [Name] instead!
func NameOld(required OptionalOrRequire) *validation.KeyRules[string] {
	return validation.Key(
		"name",
		validation.Required.When(required).Error("Name is required"),
		is.String,
		is.PrintableUnicode,
		validation.Length(1, 300).Error("Name must be between 1 and 300 characters"),
	).Required(required)
}

func Name() validation.AllOfRule {
	return validation.AllOf(
		// All names cannot be set to an empty string or nil, so they are soft
		// required if the key is present!
		is.String,
		is.PrintableUnicode,
		validation.Length(1, 300).Error("Name must be between 1 and 300 characters"),
		validation.Required.Error("Name is required"),
	)
}

func TextField() validation.AllOfRule {
	return validation.AllOf(
		is.String,
		is.PrintableUnicode,
		validation.Length(1, 300).Error("Must be between 1 and 300 characters"),
	)
}

func Mask() validation.AllOfRule {
	return validation.AllOf(
		is.String,
		validation.Length(4, 4).Error("Mask must be exactly 4 digits"),
		validation.Match(regexp.MustCompile(`\d{4}`)).Error("Mask must be a 4 digit string"),
	)
}

func CurrencyCode() validation.Rule {
	return validation.AllOf(
		is.String,
		validation.Length(3, 3).Error("Currency must be exactly 3 characters long"),
		is.Alpha.Error("Currency must be alphabetical characters only"),
		is.UpperCase.Error("Currency must be all upper case"),
		validation.In(
			locale.GetInstalledCurrencies()...,
		).Error("Currency must be one supported by the server"),
	)
}
