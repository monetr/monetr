package ofx

import (
	"encoding/xml"
	"io"
	"regexp"
	"time"

	"github.com/elliotcourant/gofx"
	"github.com/pkg/errors"
)

var (
	ofxDateRegex   = regexp.MustCompile(`^(?<timestamp>(?:\d{14}|\d{8})(?:\.\d{3})?)`)
	ofxDateFormats = []string{
		"20060102150405.000",
		"20060102150405",
		"20060102",
	}
)

func Parse(reader io.Reader) (*gofx.OFX, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read OFX buffer")
	}

	var ofx gofx.OFX
	if originalErr := xml.Unmarshal(data, &ofx); originalErr != nil {
		tokens, err := Tokenize(string(data))
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse")
		}
		xmlData := ConvertOFXToXML(tokens)
		data = xmlData

		if err := xml.Unmarshal(data, &ofx); err != nil {
			return nil, errors.Wrap(err, "failed to parse OFX")
		}
	}

	return &ofx, nil
}

func ParseDate(input string, timezone *time.Location) (time.Time, error) {
	matches := ofxDateRegex.FindAllString(input, -1)
	if len(matches) != 1 {
		return time.Time{}, errors.Errorf("failed to parse OFX timestamp [%s], found %d matching patterns", input, len(matches))
	}

	// We know that matches has exactly one item. Overwrite our input variable
	// with our "cleaned" version.
	input = matches[0]

	// Attempt to parse the input in all of the known formats that monetr can
	// handle.
	for _, format := range ofxDateFormats {
		result, err := time.ParseInLocation(format, input, timezone)
		if err != nil {
			continue
		}

		return result, nil
	}

	// If none of the formats are valid then return an error, if you are a user
	// reading this code and you see this error please open a bug issue on github
	// so we can add support for more date formats.
	return time.Time{}, errors.Errorf("failed to parse OFX timestamp [%s], unknown format", input)
}
