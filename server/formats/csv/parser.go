package csv

import (
	"encoding/csv"
	"io"

	"github.com/monetr/monetr/server/formats"
	"github.com/pkg/errors"
)

type CSVParser struct {
	mapping        formats.FieldIndex
	firstRowHeader bool
	reader         *csv.Reader
}

func NewCSVParser(
	mapping formats.FieldIndex,
	firstRowHeader bool,
	reader io.Reader,
) *CSVParser {
	return &CSVParser{
		mapping:        mapping,
		firstRowHeader: firstRowHeader,
		reader:         csv.NewReader(reader),
	}
}

func (c *CSVParser) GetNextRow() (formats.Row, error) {
	baseRow, err := c.reader.Read()
	if err != nil {
		return nil, err
	}

	if len(c.mapping) > len(baseRow) {
		return nil, errors.Errorf(
			"col number mismatch, expected %d column(s); found %d",
			len(c.mapping), len(baseRow),
		)
	}

	row := make(formats.Row)

	for index, field := range c.mapping {
		if field == formats.FieldIgnore {
			continue
		}
		row[field] = baseRow[index]
	}

	return row, nil
}
