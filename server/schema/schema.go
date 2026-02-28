package schema

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/merge"
	"github.com/pkg/errors"
)

var (
	resolveOptions = &jsonschema.ResolveOptions{
		BaseURI:          "https://monetr.app/schemas/",
		Loader:           nil,
		ValidateDefaults: true,
	}
)

func Parse[T any](
	ctx context.Context,
	existing *T,
	input map[string]any,
	schema *jsonschema.Resolved,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	if err := schema.Validate(input); err != nil {
		return errors.WithStack(err)
	}

	if err := merge.Merge(
		existing, input, merge.ErrorOnUnknownField,
	); err != nil {
		return errors.Wrap(err, "failed to merge patched data")
	}

	return nil
}

func ParseInto[T any](
	ctx context.Context,
	reader io.Reader,
	schema *jsonschema.Resolved,
) (*T, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	duplicate := bytes.NewBuffer(nil)
	{ // Decode the reader into a map and then validate it
		teeReader := io.TeeReader(reader, duplicate)
		raw := map[string]any{}
		decoder := json.NewDecoder(teeReader)
		if err := decoder.Decode(&raw); err != nil {
			return nil, errors.WithStack(err)
		}

		if err := schema.Validate(raw); err != nil {
			return nil, errors.WithStack(err)
		}
	}

	result := new(T)
	if err := json.NewDecoder(duplicate).Decode(result); err != nil {
		return nil, errors.WithStack(err)
	}

	return result, nil
}

func ParseReaderInto[T any](
	ctx context.Context,
	reader io.Reader,
	schema *jsonschema.Resolved,
) (*T, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	rawData := map[string]any{}
	decoder := json.NewDecoder(reader)
	decoder.UseNumber()
	if err := decoder.Decode(&rawData); err != nil {
		return nil, errors.WithStack(err)
	}

	result := new(T)
	if err := Parse(span.Context(), result, rawData, schema); err != nil {
		return nil, err
	}

	return result, nil
}
