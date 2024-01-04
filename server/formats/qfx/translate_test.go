package qfx

import (
	"encoding/xml"
	"testing"

	"github.com/elliotcourant/gofx"
	"github.com/stretchr/testify/assert"
)

func TestConvertToXML(t *testing.T) {
	t.Run("nfcu", func(t *testing.T) {
		data := GetFixtures(t, "sample-nfcu.qfx")
		token := Tokenize(string(data))

		xmlString := ConvertToXML(token)
		assert.NotEmpty(t, xmlString, "must produce an xml string")
	})
}

func TestValidXMLOutput(t *testing.T) {
	t.Run("nfcu", func(t *testing.T) {
		data := GetFixtures(t, "sample-nfcu.qfx")
		token := Tokenize(string(data))

		convertedToXml := ConvertToXML(token)
		assert.NotEmpty(t, convertedToXml, "must produce an xml string")

		var ofx gofx.OFX
		assert.NoError(t, xml.Unmarshal(convertedToXml, &ofx), "should unmarshal an error")
		assert.NotNil(t, ofx.SIGNONMSGSRSV1, "sign on message response must not be nil")
		assert.NotNil(t, ofx.BANKMSGSRSV1, "bank message response must not be nil")
	})

	t.Run("us bank", func(t *testing.T) {
		data := GetFixtures(t, "sample-usbank.qfx")
		token := Tokenize(string(data))

		convertedToXml := ConvertToXML(token)
		assert.NotEmpty(t, convertedToXml, "must produce an xml string")

		var ofx gofx.OFX
		assert.NoError(t, xml.Unmarshal(convertedToXml, &ofx), "should unmarshal an error")
		assert.NotNil(t, ofx.SIGNONMSGSRSV1, "sign on message response must not be nil")
		assert.NotNil(t, ofx.CREDITCARDMSGSRSV1, "credit card message response must not be nil")
	})
}
