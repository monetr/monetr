package csv

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"io"

	"github.com/pkg/errors"
)

const peekSize = 1000

var testDelimeters = []rune{
	',',
	'|',
	'\t',
}

// PeekHeader reads the first [peekSize] of the provided reader, making a copy
// of it such that an intact reader can be returned in order to facilitate file
// uploads. If the first peek of the data contains valid CSV data then the
// headers of the file are returned with a buffer to upload the file. If the
// file is not valid then an error is returned indicating that it cannot be
// parsed.
func PeekHeader(
	reader io.Reader,
) ([]string, io.Reader, error) {
	buffer := bufio.NewReaderSize(reader, peekSize)
	preview, err := buffer.Peek(peekSize)
	if err != nil && err != io.EOF {
		return nil, nil, errors.WithStack(err)
	}

	for _, delimeter := range testDelimeters {
		csvReader := csv.NewReader(bytes.NewReader(preview))
		csvReader.Comma = delimeter
		csvReader.TrimLeadingSpace = true

		header, err := csvReader.Read()
		if err != nil {
			// TODO Log the error?
			continue
		}

		// We must have at least:
		// - Date
		// - Amount
		// - Description
		if len(header) >= 3 {
			return header, buffer, nil
		}
	}

	return nil, nil, errors.New("failed to determine CSV headers")
}
