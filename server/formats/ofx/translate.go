package ofx

const xmlHeader = `<?xml version="1.0" encoding="UTF-8" standalone="no"?>`

func ConvertOFXToXML(token Token) []byte {
	return []byte(xmlHeader + token.XML())
}
