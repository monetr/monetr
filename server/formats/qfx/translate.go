package qfx

import "fmt"

func ConvertToXML(token Token) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="no"?>%s`, token.XML())
}
