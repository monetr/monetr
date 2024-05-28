package qfx

import (
	"encoding/xml"
	"io"
	"time"

	"github.com/elliotcourant/gofx"
	"github.com/pkg/errors"
)

func Parse(reader io.Reader) (*gofx.OFX, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read OFX buffer")
	}

	tokens := Tokenize(string(data))
	xmlData := ConvertQFXToXML(tokens)

	var ofx gofx.OFX
	if err := xml.Unmarshal(xmlData, &ofx); err != nil {
		return nil, errors.Wrap(err, "failed to parse OFX")
	}

	return &ofx, nil
}

func ParseDate(input string, timezone *time.Location) (time.Time, error) {
	result, err := time.ParseInLocation("20060102150405.000", input, timezone)
	return result, errors.Wrapf(err, "failed to parse QFX timestamp [%s]", input)
}
