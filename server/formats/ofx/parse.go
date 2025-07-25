package ofx

import (
	"bytes"
	"encoding/xml"
	"io"
	"regexp"
	"time"

	"github.com/elliotcourant/gofx"
	"github.com/pkg/errors"
	"golang.org/x/text/encoding/ianaindex"
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
	xmlBytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read all of the OFX data from the reader")
	}

	ofx, err := parseInner(bytes.NewReader(xmlBytes))
	if err != nil {
		tokens, err := Tokenize(xmlBytes)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse")
		}
		xmlBytes = ConvertOFXToXML(tokens)
		ofx, err = parseInner(bytes.NewReader(xmlBytes))
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse OFX")
		}
	}

	return ofx, nil
}

func parseInner(reader io.Reader) (*gofx.OFX, error) {
	var ofx gofx.OFX
	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		enc, err := ianaindex.IANA.Encoding(charset)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to decode charset %s", charset)
		}

		return enc.NewDecoder().Reader(input), nil
	}
	if err := decoder.Decode(&ofx); err != nil {
		return nil, errors.Wrap(err, "failed to decode OFX file as XML")
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
