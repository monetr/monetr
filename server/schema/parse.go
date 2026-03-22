package schema

import (
	"io"

	"github.com/Oudwins/zog"
	z "github.com/Oudwins/zog"
	"github.com/Oudwins/zog/parsers/zjson"
)

func Parse[T any](
	schema zog.ComplexZogSchema,
	reader io.Reader,
	existing T,
	options ...z.ExecOption,
) (*T, zog.ZogIssueList) {
	errs := schema.Parse(
		zjson.Decode(reader),
		&existing,
		options...,
	)
	return &existing, errs
}
