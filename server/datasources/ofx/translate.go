package ofx

import "bytes"

const xmlHeader = `<?xml version="1.0" encoding="UTF-8" standalone="no"?>`

func ConvertOFXToXML(token Token) []byte {
	buf := bytes.NewBufferString(xmlHeader)
	token.writeXML(buf)
	return buf.Bytes()
}
