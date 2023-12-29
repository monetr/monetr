package csv

import (
	"encoding/csv"
	"io"

	"github.com/monetr/monetr/server/formats"
)

type FieldIndex []formats.Field

type Row map[formats.Field]string

type CSVParser struct {
	mapping        FieldIndex
	firstRowHeader bool
	reader         *csv.Reader
}

func NewCSVParser(mapping FieldIndex, firstRowHeader bool, reader io.Reader) *CSVParser {
	return &CSVParser{
		mapping:        mapping,
		firstRowHeader: firstRowHeader,
		reader:         csv.NewReader(reader),
	}
}

func (c *CSVParser) GetNextRow() (Row, error) {
	baseRow, err := c.reader.Read()
	if err != nil {
		return nil, err
	}

}
