package ofx

import (
	"encoding/xml"
	"testing"

	"github.com/elliotcourant/gofx"
	"github.com/stretchr/testify/assert"
)

func TestConvertOFXToXML(t *testing.T) {
	t.Run("nfcu", func(t *testing.T) {
		data := GetFixtures(t, "sample-nfcu.qfx")
		token, err := Tokenize(data)
		assert.NoError(t, err)

		xmlString := ConvertOFXToXML(token)
		assert.NotEmpty(t, xmlString, "must produce an xml string")
	})

	t.Run("nfcu wrapped", func(t *testing.T) {
		data := GetFixtures(t, "sample-nfcu-wrapped.qfx")
		token, err := Tokenize(data)
		assert.NoError(t, err)

		xmlString := ConvertOFXToXML(token)
		assert.NotEmpty(t, xmlString, "must produce an xml string")
	})

	t.Run("us bank", func(t *testing.T) {
		data := GetFixtures(t, "sample-usbank.qfx")
		token, err := Tokenize(data)
		assert.NoError(t, err)

		xmlString := ConvertOFXToXML(token)
		assert.NotEmpty(t, xmlString, "must produce an xml string")
	})
}

func TestValidXMLOutput(t *testing.T) {
	t.Run("nfcu", func(t *testing.T) {
		data := GetFixtures(t, "sample-nfcu.qfx")
		token, err := Tokenize(data)
		assert.NoError(t, err)

		convertedToXml := ConvertOFXToXML(token)
		assert.NotEmpty(t, convertedToXml, "must produce an xml string")

		var ofx gofx.OFX
		assert.NoError(t, xml.Unmarshal(convertedToXml, &ofx), "should unmarshal an error")
		assert.NotNil(t, ofx.SIGNONMSGSRSV1, "sign on message response must not be nil")
		assert.NotNil(t, ofx.BANKMSGSRSV1, "bank message response must not be nil")
	})

	t.Run("nfcu wrapped", func(t *testing.T) {
		data := GetFixtures(t, "sample-nfcu-wrapped.qfx")
		token, err := Tokenize(data)
		assert.NoError(t, err)

		convertedToXml := ConvertOFXToXML(token)
		assert.NotEmpty(t, convertedToXml, "must produce an xml string")

		var ofx gofx.OFX
		assert.NoError(t, xml.Unmarshal(convertedToXml, &ofx), "should unmarshal an error")
		assert.NotNil(t, ofx.SIGNONMSGSRSV1, "sign on message response must not be nil")
		assert.NotNil(t, ofx.BANKMSGSRSV1, "bank message response must not be nil")
	})

	t.Run("us bank", func(t *testing.T) {
		data := GetFixtures(t, "sample-usbank.qfx")
		token, err := Tokenize(data)
		assert.NoError(t, err)

		convertedToXml := ConvertOFXToXML(token)
		assert.NotEmpty(t, convertedToXml, "must produce an xml string")

		var ofx gofx.OFX
		assert.NoError(t, xml.Unmarshal(convertedToXml, &ofx), "should unmarshal an error")
		assert.NotNil(t, ofx.SIGNONMSGSRSV1, "sign on message response must not be nil")
		assert.NotNil(t, ofx.CREDITCARDMSGSRSV1, "credit card message response must not be nil")
	})
}
