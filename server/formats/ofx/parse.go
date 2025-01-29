package ofx

import (
	"encoding/xml"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/elliotcourant/gofx"
	"github.com/pkg/errors"
)

var (
	ofxDateRegex = regexp.MustCompile(`^(?<timestamp>\d{14})(?<extra>.\d{3})?`)
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

	// Typically we would see the `.000` suffix for the timestamp, but some files
	// might not have this from some institutions.
	if strings.Contains(input, ".") {
		result, err := time.ParseInLocation("20060102150405.000", input, timezone)
		return result, errors.Wrapf(err, "failed to parse OFX timestamp [%s]", input)
	}

	// If they don't have that suffix then try to parse the timestamp without it.
	// If it is still bad then that means the file isn't OFX somehow, or there is
	// a new timestamp format that we haven't seen yet.
	// See: https://github.com/monetr/monetr/issues/2362
	result, err := time.ParseInLocation("20060102150405", input, timezone)
	return result, errors.Wrapf(err, "failed to parse OFX timestamp [%s]", input)
}
