package schemas

import (
	"context"
	"encoding/json"
	"io"

	"github.com/monetr/monetr/server/merge"
	"github.com/monetr/validation"
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
